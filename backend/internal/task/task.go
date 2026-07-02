// Package task provides Asynq client and server construction bound to the
// shared Redis broker. The enqueuer (NewAsynqClient) and worker
// (NewAsynqServer) both dial the same Redis instance from config.Redis; P2
// only proves construction and a trivial round-trip. Task type constants,
// per-kind concurrency, and handler registration are deferred to P5/P6.
package task

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"

	"ai-forum/backend/internal/config"
	"ai-forum/backend/internal/database"
	"ai-forum/backend/internal/event"
)

// p2Concurrency is the small worker concurrency used by P2 smoke tests. Real
// per-kind concurrency is configured in P5/P6 via config.WorkerConfig.
const p2Concurrency = 2

const (
	TagPost                = "tag_post"
	DecideAIReply          = "decide_ai_reply"
	GenerateAIReply        = "generate_ai_reply"
	JudgeAIFollowup        = "judge_ai_followup"
	ModerateAIReply        = "moderate_ai_reply"
	SyncSearchIndex        = "sync_search_index"
	SendNotification       = "send_notification"
	RefreshHotScore        = "refresh_hot_score"
	CleanupProcessedEvents = "cleanup_processed_events"
)

// CronContractInfo records owner metadata for P5 ownership tests.
type CronContractInfo struct {
	HandlerOwner string
}

type Handlers struct {
	TagPost       func(context.Context, int64) error
	DecideAIReply func(context.Context, int64) error
}

type TagPostPayload struct {
	PostID int64 `json:"post_id"`
}

type DecideAIReplyPayload struct {
	PostID int64 `json:"post_id"`
}

type GenerateAIReplyPayload struct {
	PostID    int64 `json:"post_id"`
	AIAgentID int64 `json:"ai_agent_id"`
}

type GenerateAIReplyEnqueuer struct {
	enqueuer Enqueuer
}

func NewGenerateAIReplyEnqueuer(enqueuer Enqueuer) *GenerateAIReplyEnqueuer {
	return &GenerateAIReplyEnqueuer{enqueuer: enqueuer}
}

func (e *GenerateAIReplyEnqueuer) EnqueueGenerateAIReply(ctx context.Context, postID, agentID int64) error {
	payload := GenerateAIReplyPayload{PostID: postID, AIAgentID: agentID}
	if enqueuer, ok := e.enqueuer.(optionEnqueuer); ok {
		err := enqueuer.EnqueueWithOptions(ctx, GenerateAIReply, payload, asynq.TaskID(generateAIReplyTaskID(postID, agentID)))
		if errors.Is(err, asynq.ErrTaskIDConflict) {
			return nil
		}
		return err
	}
	return e.enqueuer.Enqueue(ctx, GenerateAIReply, payload)
}

var cronContracts = map[string]CronContractInfo{
	CleanupProcessedEvents: {HandlerOwner: "P5 task cleanup_processed_events"},
}

// Types returns all §7.2 Asynq task type constants in stable order.
func Types() []string {
	return []string{
		TagPost,
		DecideAIReply,
		GenerateAIReply,
		JudgeAIFollowup,
		ModerateAIReply,
		SyncSearchIndex,
		SendNotification,
		RefreshHotScore,
		CleanupProcessedEvents,
	}
}

// CronTypes returns all §9.3 periodic task types owned by this package.
func CronTypes() []string {
	return []string{CleanupProcessedEvents}
}

// CronContract returns owner metadata for a periodic task.
func CronContract(taskType string) (CronContractInfo, bool) {
	c, ok := cronContracts[taskType]
	return c, ok
}

// NewAsynqClient returns an Asynq enqueuer connected to the Redis broker
// described by cfg. Both the client and the server returned by NewAsynqServer
// target the same Redis instance (design D3).
func NewAsynqClient(cfg config.RedisConfig) *asynq.Client {
	return asynq.NewClient(redisClientOpt(cfg))
}

// NewAsynqServer returns an Asynq worker server connected to the same Redis
// broker as NewAsynqClient, with a small P2 concurrency. Real per-kind
// concurrency is a P5/P6 concern.
func NewAsynqServer(cfg config.RedisConfig) *asynq.Server {
	return asynq.NewServer(redisClientOpt(cfg), asynq.Config{
		Concurrency: p2Concurrency,
	})
}

// RegisterHandlers registers task handlers available in the current phase.
func RegisterHandlers(mux *asynq.ServeMux, db database.DBTX, handlers ...Handlers) {
	mux.HandleFunc(CleanupProcessedEvents, func(ctx context.Context, _ *asynq.Task) error {
		return CleanupProcessedEventsRows(ctx, db)
	})
	var h Handlers
	if len(handlers) > 0 {
		h = handlers[0]
	}
	if h.TagPost != nil {
		mux.HandleFunc(TagPost, func(ctx context.Context, task *asynq.Task) error {
			return runDedupedTask(ctx, db, task, func() error {
				var payload TagPostPayload
				if err := json.Unmarshal(task.Payload(), &payload); err != nil {
					return fmt.Errorf("decode tag_post payload: %w", err)
				}
				return h.TagPost(ctx, payload.PostID)
			})
		})
	}
	if h.DecideAIReply != nil {
		mux.HandleFunc(DecideAIReply, func(ctx context.Context, task *asynq.Task) error {
			return runDedupedTask(ctx, db, task, func() error {
				var payload DecideAIReplyPayload
				if err := json.Unmarshal(task.Payload(), &payload); err != nil {
					return fmt.Errorf("decode decide_ai_reply payload: %w", err)
				}
				return h.DecideAIReply(ctx, payload.PostID)
			})
		})
	}
}

func runDedupedTask(ctx context.Context, db database.DBTX, task *asynq.Task, fn func() error) error {
	eventID := taskEventID(ctx, task)
	consumerName := "asynq." + task.Type()
	done, err := event.IsProcessed(ctx, db, eventID, consumerName)
	if err != nil {
		return err
	}
	if done {
		return nil
	}
	if err := fn(); err != nil {
		return err
	}
	return event.MarkProcessed(ctx, db, eventID, consumerName)
}

func taskEventID(ctx context.Context, task *asynq.Task) string {
	if id, ok := asynq.GetTaskID(ctx); ok && id != "" {
		return id
	}
	sum := sha256.Sum256(append([]byte(task.Type()+":"), task.Payload()...))
	return "task:" + hex.EncodeToString(sum[:])
}

func generateAIReplyTaskID(postID, agentID int64) string {
	return fmt.Sprintf("%s:%d:%d", GenerateAIReply, postID, agentID)
}

// NewScheduler returns an Asynq scheduler connected to the shared Redis broker.
func NewScheduler(cfg config.RedisConfig) *asynq.Scheduler {
	return asynq.NewScheduler(redisClientOpt(cfg), nil)
}

// RegisterCleanupCron schedules the daily processed_events cleanup task.
func RegisterCleanupCron(s *asynq.Scheduler) (string, error) {
	payload, err := json.Marshal(map[string]string{"task": CleanupProcessedEvents})
	if err != nil {
		return "", err
	}
	id, err := s.Register("@daily", asynq.NewTask(CleanupProcessedEvents, payload))
	if err != nil {
		return "", fmt.Errorf("register cleanup processed events cron: %w", err)
	}
	return id, nil
}

// CleanupProcessedEventsRows deletes old processed_events rows.
func CleanupProcessedEventsRows(ctx context.Context, db database.DBTX) error {
	_, err := db.ExecContext(ctx, `DELETE FROM processed_events WHERE processed_at < NOW() - INTERVAL 30 DAY`)
	if err != nil {
		return fmt.Errorf("cleanup processed events: %w", err)
	}
	return nil
}

// redisClientOpt maps the project config.RedisConfig to the Asynq
// RedisClientOpt used by both the client and the server so they share one
// broker definition.
func redisClientOpt(cfg config.RedisConfig) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
}
