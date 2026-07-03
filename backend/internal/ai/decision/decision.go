// Package decision computes AI reply willingness and selection.
package decision

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"

	aiagent "ai-forum/backend/internal/ai/agent"
	forumtag "ai-forum/backend/internal/forum/tag"
)

const (
	DecisionReply    = "REPLY"
	DecisionIgnore   = "IGNORE"
	DecisionFallback = "FALLBACK"
)

// TagScore is an agent preference weight for a post tag.
type TagScore struct {
	Type   string
	Weight float64
}

// ScoreInput contains formula inputs for §11.2.
type ScoreInput struct {
	Tags             []TagScore
	ActivityScore    float64
	RiskPenalty      float64
	FrequencyPenalty float64
}

// AgentScore is a computed score plus selection metadata.
type AgentScore struct {
	AgentID   int64
	Score     float64
	Threshold float64
	Fallback  bool
	Decision  string
}

type Agent struct {
	ID             int64
	ReplyThreshold float64
	ActivityLevel  float64
	AllowAutoReply bool
	Fallback       bool
	Preferences    []Preference
}

type Preference struct {
	TagType string
	TagName string
	Weight  float64
}

type PostTag struct {
	Type string
	Name string
}

type Log struct {
	PostID           int64
	AgentID          int64
	TriggerType      string
	WillingnessScore float64
	ThresholdValue   float64
	Decision         string
	Reason           string
	HitTags          []PostTag
}

type AgentReader interface {
	ListEnabledAgents(context.Context) ([]Agent, error)
}

type TagReader interface {
	ListPostTags(context.Context, int64) ([]PostTag, error)
}

type DecisionLogger interface {
	WriteDecisionLog(context.Context, Log) error
}

type ReplyEnqueuer interface {
	EnqueueAutoGenerateAIReply(ctx context.Context, postID, agentID int64) error
}

type Handler struct {
	agents   AgentReader
	tags     TagReader
	logs     DecisionLogger
	enqueuer ReplyEnqueuer
}

type SQLHandler struct {
	db       *sqlx.DB
	agents   *aiagent.SQLRepository
	tags     *forumtag.SQLRepository
	enqueuer ReplyEnqueuer
}

func NewHandler(agents AgentReader, tags TagReader, logs DecisionLogger, enqueuer ReplyEnqueuer) *Handler {
	return &Handler{agents: agents, tags: tags, logs: logs, enqueuer: enqueuer}
}

func NewSQLHandler(db *sqlx.DB, enqueuer ReplyEnqueuer) *SQLHandler {
	return &SQLHandler{
		db:       db,
		agents:   aiagent.NewSQLRepository(db),
		tags:     forumtag.NewSQLRepository(),
		enqueuer: enqueuer,
	}
}

func (h *Handler) HandleDecideAIReply(ctx context.Context, postID int64) error {
	agents, err := h.agents.ListEnabledAgents(ctx)
	if err != nil {
		return err
	}
	postTags, err := h.tags.ListPostTags(ctx, postID)
	if err != nil {
		return err
	}
	scores := make([]AgentScore, 0, len(agents))
	logs := make([]Log, 0, len(agents))
	for _, agent := range agents {
		if !agent.AllowAutoReply {
			continue
		}
		hitTags, tagScores := matchTags(agent.Preferences, postTags)
		score := WillingnessScore(ScoreInput{Tags: tagScores, ActivityScore: agent.ActivityLevel})
		scores = append(scores, AgentScore{AgentID: agent.ID, Score: score, Threshold: agent.ReplyThreshold, Fallback: agent.Fallback})
		logs = append(logs, Log{
			PostID:           postID,
			AgentID:          agent.ID,
			TriggerType:      "AUTO",
			WillingnessScore: score,
			ThresholdValue:   agent.ReplyThreshold,
			Decision:         DecisionIgnore,
			Reason:           "below threshold",
			HitTags:          hitTags,
		})
	}
	selected := SelectReplies(scores)
	selectedByID := map[int64]string{}
	for _, agent := range selected {
		selectedByID[agent.AgentID] = agent.Decision
	}
	for _, log := range logs {
		if decision, ok := selectedByID[log.AgentID]; ok {
			log.Decision = decision
			log.Reason = "selected"
			if err := h.enqueuer.EnqueueAutoGenerateAIReply(ctx, postID, log.AgentID); err != nil {
				return err
			}
		}
		if err := h.logs.WriteDecisionLog(ctx, log); err != nil {
			return err
		}
	}
	return nil
}

func (h *SQLHandler) HandleDecideAIReply(ctx context.Context, postID int64) error {
	var existing int
	if err := h.db.GetContext(ctx, &existing, `SELECT COUNT(*) FROM decision_logs WHERE post_id = ? AND trigger_type = 'AUTO'`, postID); err != nil {
		return fmt.Errorf("count decision logs: %w", err)
	}
	if existing > 0 {
		return h.enqueueSelectedFromLogs(ctx, postID)
	}
	agents, err := h.agents.ListEnabledWithPreferences(ctx)
	if err != nil {
		return err
	}
	postTags, err := h.tags.List(ctx, h.db, postID)
	if err != nil {
		return err
	}
	logs, selected := buildDecisionLogs(postID, convertAgents(agents), convertTags(postTags))
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin decision transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	for _, log := range logs {
		hitTags, err := MarshalHitTags(log.HitTags)
		if err != nil {
			return fmt.Errorf("marshal decision hit tags: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO decision_logs (
				post_id, comment_id, ai_agent_id, trigger_type, willingness_score,
				threshold_value, decision, reason, hit_tags
			) VALUES (?, NULL, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				willingness_score = VALUES(willingness_score),
				threshold_value = VALUES(threshold_value),
				decision = VALUES(decision),
				reason = VALUES(reason),
				hit_tags = VALUES(hit_tags)`,
			log.PostID, log.AgentID, log.TriggerType, log.WillingnessScore,
			log.ThresholdValue, log.Decision, log.Reason, hitTags); err != nil {
			return fmt.Errorf("write decision log: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit decision logs: %w", err)
	}
	committed = true
	for _, agentID := range selected {
		if err := h.enqueuer.EnqueueAutoGenerateAIReply(ctx, postID, agentID); err != nil {
			return err
		}
	}
	return nil
}

func (h *SQLHandler) enqueueSelectedFromLogs(ctx context.Context, postID int64) error {
	var agentIDs []int64
	if err := h.db.SelectContext(ctx, &agentIDs, `
		SELECT ai_agent_id
		FROM decision_logs
		WHERE post_id = ? AND trigger_type = 'AUTO' AND decision IN (?, ?)
		ORDER BY id`, postID, DecisionReply, DecisionFallback); err != nil {
		return fmt.Errorf("list selected decision logs: %w", err)
	}
	for _, agentID := range agentIDs {
		if err := h.enqueuer.EnqueueAutoGenerateAIReply(ctx, postID, agentID); err != nil {
			return err
		}
	}
	return nil
}

// WillingnessScore computes the §11.2 weighted score.
func WillingnessScore(in ScoreInput) float64 {
	return tagTypeScore(in.Tags, "topic")*0.35 +
		tagTypeScore(in.Tags, "intent")*0.25 +
		tagTypeScore(in.Tags, "emotion")*0.15 +
		tagTypeScore(in.Tags, "debate")*0.15 +
		in.ActivityScore*0.10 -
		in.RiskPenalty -
		in.FrequencyPenalty
}

// SelectReplies applies threshold and fallback rules.
func SelectReplies(agents []AgentScore) []AgentScore {
	var selected []AgentScore
	for _, agent := range agents {
		if agent.Score >= agent.Threshold {
			agent.Decision = DecisionReply
			selected = append(selected, agent)
		}
	}
	if len(selected) > 0 {
		return selected
	}
	highest := highestScore(agents)
	if highest.Score >= 0.35 {
		highest.Decision = DecisionReply
		return []AgentScore{highest}
	}
	for _, agent := range agents {
		if agent.Fallback {
			agent.Decision = DecisionFallback
			return []AgentScore{agent}
		}
	}
	if len(agents) == 0 {
		return nil
	}
	highest.Decision = DecisionFallback
	return []AgentScore{highest}
}

func tagTypeScore(tags []TagScore, tagType string) float64 {
	var max, sum float64
	var count int
	for _, tag := range tags {
		if tag.Type != tagType {
			continue
		}
		if count == 0 || tag.Weight > max {
			max = tag.Weight
		}
		sum += tag.Weight
		count++
	}
	if count == 0 {
		return 0
	}
	return max*0.7 + (sum/float64(count))*0.3
}

func highestScore(agents []AgentScore) AgentScore {
	var highest AgentScore
	for i, agent := range agents {
		if i == 0 || agent.Score > highest.Score {
			highest = agent
		}
	}
	return highest
}

func matchTags(preferences []Preference, tags []PostTag) ([]PostTag, []TagScore) {
	var hits []PostTag
	var scores []TagScore
	for _, tag := range tags {
		for _, pref := range preferences {
			if pref.TagType == tag.Type && pref.TagName == tag.Name {
				hits = append(hits, tag)
				scores = append(scores, TagScore{Type: tag.Type, Weight: pref.Weight})
			}
		}
	}
	return hits, scores
}

func buildDecisionLogs(postID int64, agents []Agent, postTags []PostTag) ([]Log, []int64) {
	scores := make([]AgentScore, 0, len(agents))
	logs := make([]Log, 0, len(agents))
	for _, agent := range agents {
		if !agent.AllowAutoReply {
			continue
		}
		hitTags, tagScores := matchTags(agent.Preferences, postTags)
		score := WillingnessScore(ScoreInput{Tags: tagScores, ActivityScore: agent.ActivityLevel})
		scores = append(scores, AgentScore{AgentID: agent.ID, Score: score, Threshold: agent.ReplyThreshold, Fallback: agent.Fallback})
		logs = append(logs, Log{
			PostID:           postID,
			AgentID:          agent.ID,
			TriggerType:      "AUTO",
			WillingnessScore: score,
			ThresholdValue:   agent.ReplyThreshold,
			Decision:         DecisionIgnore,
			Reason:           "below threshold",
			HitTags:          hitTags,
		})
	}
	selected := SelectReplies(scores)
	selectedByID := map[int64]string{}
	selectedIDs := make([]int64, 0, len(selected))
	for _, agent := range selected {
		selectedByID[agent.AgentID] = agent.Decision
		selectedIDs = append(selectedIDs, agent.AgentID)
	}
	for i := range logs {
		if decision, ok := selectedByID[logs[i].AgentID]; ok {
			logs[i].Decision = decision
			logs[i].Reason = "selected"
		}
	}
	return logs, selectedIDs
}

func convertAgents(rows []aiagent.Agent) []Agent {
	agents := make([]Agent, 0, len(rows))
	for _, row := range rows {
		prefs := make([]Preference, 0, len(row.Preferences))
		for _, pref := range row.Preferences {
			prefs = append(prefs, Preference{TagType: pref.TagType, TagName: pref.TagName, Weight: pref.Weight})
		}
		agents = append(agents, Agent{
			ID:             row.ID,
			ReplyThreshold: row.ReplyThreshold,
			ActivityLevel:  row.ActivityLevel,
			AllowAutoReply: row.AllowAutoReply,
			Fallback:       row.Fallback,
			Preferences:    prefs,
		})
	}
	return agents
}

func convertTags(rows []forumtag.Tag) []PostTag {
	tags := make([]PostTag, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, PostTag{Type: row.Type, Name: row.Name})
	}
	return tags
}

func MarshalHitTags(tags []PostTag) ([]byte, error) {
	if tags == nil {
		tags = []PostTag{}
	}
	return json.Marshal(tags)
}
