// Package bootstrap owns process composition and lifecycle wiring.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ai-forum/backend/internal/admin"
	aichat "ai-forum/backend/internal/ai/chat"
	"ai-forum/backend/internal/ai/decision"
	"ai-forum/backend/internal/ai/followup"
	"ai-forum/backend/internal/ai/modelclient"
	"ai-forum/backend/internal/ai/reply"
	"ai-forum/backend/internal/ai/tagging"
	"ai-forum/backend/internal/auth"
	"ai-forum/backend/internal/cache"
	"ai-forum/backend/internal/config"
	"ai-forum/backend/internal/database"
	comment "ai-forum/backend/internal/forum/comment"
	favorite "ai-forum/backend/internal/forum/favorite"
	like "ai-forum/backend/internal/forum/like"
	post "ai-forum/backend/internal/forum/post"
	forumtag "ai-forum/backend/internal/forum/tag"
	"ai-forum/backend/internal/internalapi"
	"ai-forum/backend/internal/logger"
	"ai-forum/backend/internal/moderation"
	"ai-forum/backend/internal/mq"
	"ai-forum/backend/internal/notification"
	"ai-forum/backend/internal/outbox"
	"ai-forum/backend/internal/rbac"
	"ai-forum/backend/internal/router"
	"ai-forum/backend/internal/search"
	"ai-forum/backend/internal/sse"
	"ai-forum/backend/internal/task"
	"ai-forum/backend/internal/user"
)

// Process is the common lifecycle contract used by the three backend binaries.
type Process interface {
	Start(context.Context) error
	Stop(context.Context) error
}

// RunProcess starts p, waits for SIGINT/SIGTERM, then shuts down p and app.
func RunProcess(ctx context.Context, app *App, p Process, timeout time.Duration) error {
	runCtx, stopSignals := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	if err := p.Start(runCtx); err != nil {
		return err
	}
	<-runCtx.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := p.Stop(stopCtx)
	if closeErr := app.Close(stopCtx); closeErr != nil {
		err = errors.Join(err, closeErr)
	}
	return err
}

// App owns shared infrastructure for the three processes.
type App struct {
	Cfg         *config.Config
	Log         *logger.Logger
	DB          *sqlx.DB
	Redis       *redis.Client
	RabbitMQ    *mq.Connection
	ES          *es.Client
	AsynqClient *asynq.Client
	AsynqServer *asynq.Server
	Scheduler   *asynq.Scheduler
}

// NewApp builds shared dependencies once for process constructors.
func NewApp(cfg *config.Config) (*App, error) {
	log, err := logger.New(cfg.Log)
	if err != nil {
		return nil, err
	}
	db, err := database.NewMySQL(cfg.MySQL)
	if err != nil {
		return nil, err
	}
	redisClient, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	rabbit, err := mq.NewRabbitMQ(cfg.RabbitMQ)
	if err != nil {
		_ = redisClient.Close()
		_ = db.Close()
		return nil, err
	}
	esClient, err := search.NewES(cfg.Elasticsearch)
	if err != nil {
		_ = rabbit.Close()
		_ = redisClient.Close()
		_ = db.Close()
		return nil, err
	}
	return &App{
		Cfg:         cfg,
		Log:         log,
		DB:          db,
		Redis:       redisClient,
		RabbitMQ:    rabbit,
		ES:          esClient,
		AsynqClient: task.NewAsynqClient(cfg.Redis),
		AsynqServer: task.NewAsynqServer(cfg.Redis),
		Scheduler:   task.NewScheduler(cfg.Redis),
	}, nil
}

// NewAPIServer wires the api-server HTTP process.
func (a *App) NewAPIServer() Process {
	hub := sse.NewHub()
	internal := internalapi.NewHandler(a.Cfg.InternalAPI, hub, a.Log)
	userRepo := user.NewSQLRepository(a.DB)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)
	tokens := auth.NewTokenManager(a.Cfg.JWT.Secret, time.Duration(a.Cfg.JWT.ExpireHours)*time.Hour)
	authHandler := auth.NewHandler(userSvc, tokens)
	runTx := func(ctx context.Context, fn func(database.DBTX) error) error {
		return database.RunInTx(ctx, a.DB, func(tx *sqlx.Tx) error { return fn(tx) })
	}
	var hot *post.RedisHotStore
	if a.Redis != nil {
		hot = post.NewRedisHotStore(a.Redis, func(ctx context.Context, postID int64) (post.PostSnapshot, error) {
			return post.LoadPostSnapshot(ctx, a.DB, postID)
		})
	}
	var postOpts []post.Option
	var commentOpts []comment.Option
	if hot != nil {
		postOpts = append(postOpts, post.WithHotTracker(hot))
		commentOpts = append(commentOpts, comment.WithHotTracker(hot))
	}
	postSvc := post.NewService(post.NewSQLRepository(), outbox.Append, postOpts...)
	postHandler := post.NewHandler(postSvc, runTx)
	var likeOpts []like.Option
	if hot != nil {
		likeOpts = append(likeOpts, like.WithHotTracker(hot))
	}
	likeSvc := like.NewService(like.NewSQLRepository(), outbox.Append, likeOpts...)
	likeHandler := like.NewHandler(likeSvc, runTx)
	favoriteSvc := favorite.NewService(favorite.NewSQLRepository(), outbox.Append)
	favoriteHandler := favorite.NewHandler(favoriteSvc, runTx)
	if a.Redis != nil {
		commentOpts = append(commentOpts, comment.WithMentionLimiter(comment.NewRedisMentionLimiter(a.Redis, 5, time.Minute)))
	} else {
		commentOpts = append(commentOpts, comment.WithMentionLimiter(comment.NewMemoryMentionLimiter(5, time.Minute, time.Now)))
	}
	var retryAIReplies http.Handler
	if a.AsynqClient != nil {
		enqueuer := task.NewAsynqEnqueuer(a.AsynqClient)
		generateEnqueuer := task.NewGenerateAIReplyEnqueuer(enqueuer)
		commentOpts = append(commentOpts,
			comment.WithGenerateEnqueuer(generateEnqueuer),
			comment.WithFollowupEnqueuer(task.NewJudgeAIFollowupEnqueuer(enqueuer)),
		)
		retryAIReplies = http.HandlerFunc(reply.NewRetryHandler(a.DB, generateEnqueuer).RetryPostFailedReplies)
	}
	commentSvc := comment.NewService(comment.NewSQLRepository(), outbox.Append, commentOpts...)
	commentHandler := comment.NewHandler(commentSvc, runTx)
	tagHandler := forumtag.NewHandler(forumtag.NewSQLRepository(), runTx)
	notificationHTTPHandler := notification.NewHTTPHandler(a.DB)
	adminHandler := admin.NewHandler(admin.NewSQLStore(a.DB), mustAdminAuthorizer())
	chatHandler := a.newChatHandler()
	searchStore := search.NewESIndexStore(a.ES)
	searchQueryHandler := search.NewQueryHandler(a.DB, searchStore)
	routes, err := businessRoutes(businessRouteDeps{
		tokens:                      tokens,
		register:                    http.HandlerFunc(userHandler.Register),
		login:                       http.HandlerFunc(authHandler.Login),
		profile:                     http.HandlerFunc(userHandler.Profile),
		updateProfile:               http.HandlerFunc(userHandler.UpdateProfile),
		profileStats:                http.HandlerFunc(userHandler.Stats),
		hotTags:                     http.HandlerFunc(tagHandler.ListHot),
		listPosts:                   http.HandlerFunc(postHandler.List),
		getPost:                     http.HandlerFunc(postHandler.Get),
		createPost:                  http.HandlerFunc(postHandler.Create),
		updatePost:                  http.HandlerFunc(postHandler.UpdateOwn),
		deletePost:                  http.HandlerFunc(postHandler.Delete),
		listComments:                http.HandlerFunc(commentHandler.List),
		createComment:               http.HandlerFunc(commentHandler.Create),
		likePost:                    http.HandlerFunc(likeHandler.Like),
		unlikePost:                  http.HandlerFunc(likeHandler.Unlike),
		favoritePost:                http.HandlerFunc(favoriteHandler.Favorite),
		unfavoritePost:              http.HandlerFunc(favoriteHandler.Unfavorite),
		listNotifications:           http.HandlerFunc(notificationHTTPHandler.List),
		unreadNotifications:         http.HandlerFunc(notificationHTTPHandler.UnreadCount),
		markNotificationRead:        http.HandlerFunc(notificationHTTPHandler.MarkRead),
		markAllNotificationsRead:    http.HandlerFunc(notificationHTTPHandler.MarkAllRead),
		postEvents:                  sse.NewEventsHandler(hub),
		aiStatus:                    sse.NewStatusHandler(sse.NewSQLStatusStore(a.DB)),
		retryAIReplies:              retryAIReplies,
		listAgentChatConversations:  http.HandlerFunc(chatHandler.List),
		getAgentChatConversation:    http.HandlerFunc(chatHandler.Get),
		streamAgentChatMessage:      http.HandlerFunc(chatHandler.Stream),
		deleteAgentChatConversation: http.HandlerFunc(chatHandler.Delete),
		retryAgentChatMessage:       http.HandlerFunc(chatHandler.Retry),
		searchPosts:                 http.HandlerFunc(searchQueryHandler.SearchPosts),
		updatePostStatus:            http.HandlerFunc(postHandler.UpdateStatus),
		admin:                       adminHandler,
	})
	if err != nil {
		return NewErrorProcess(err)
	}
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", a.Cfg.Server.Port),
		Handler:           router.NewWithBusinessRoutes(a.readinessChecks(), internal, routes),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return NewHTTPProcess("api-server", srv, a.Log)
}

type businessRouteDeps struct {
	tokens                      *auth.TokenManager
	register                    http.Handler
	login                       http.Handler
	profile                     http.Handler
	updateProfile               http.Handler
	profileStats                http.Handler
	hotTags                     http.Handler
	listPosts                   http.Handler
	getPost                     http.Handler
	createPost                  http.Handler
	updatePost                  http.Handler
	deletePost                  http.Handler
	listComments                http.Handler
	createComment               http.Handler
	likePost                    http.Handler
	unlikePost                  http.Handler
	favoritePost                http.Handler
	unfavoritePost              http.Handler
	listNotifications           http.Handler
	unreadNotifications         http.Handler
	markNotificationRead        http.Handler
	markAllNotificationsRead    http.Handler
	postEvents                  http.Handler
	aiStatus                    http.Handler
	retryAIReplies              http.Handler
	listAgentChatConversations  http.Handler
	getAgentChatConversation    http.Handler
	streamAgentChatMessage      http.Handler
	deleteAgentChatConversation http.Handler
	retryAgentChatMessage       http.Handler
	searchPosts                 http.Handler
	updatePostStatus            http.Handler
	admin                       *admin.Handler
}

func businessRoutes(deps businessRouteDeps) (router.BusinessRoutes, error) {
	authz, err := rbac.NewAuthorizer(rbac.DefaultModelPath())
	if err != nil {
		return router.BusinessRoutes{}, err
	}
	if err := authz.SeedAdminPolicies(); err != nil {
		return router.BusinessRoutes{}, err
	}
	routes := router.BusinessRoutes{
		Register:                    deps.register,
		Login:                       deps.login,
		Profile:                     deps.tokens.Middleware(deps.profile),
		UpdateProfile:               deps.tokens.Middleware(deps.updateProfile),
		ProfileStats:                deps.tokens.Middleware(deps.profileStats),
		HotTags:                     deps.hotTags,
		ListPosts:                   deps.listPosts,
		GetPost:                     deps.getPost,
		CreatePost:                  deps.tokens.Middleware(deps.createPost),
		UpdatePost:                  deps.tokens.Middleware(deps.updatePost),
		DeletePost:                  deps.tokens.Middleware(authz.RequireSubject("post", "delete-any", deps.deletePost)),
		ListComments:                deps.listComments,
		CreateComment:               deps.tokens.Middleware(deps.createComment),
		LikePost:                    deps.tokens.Middleware(deps.likePost),
		UnlikePost:                  deps.tokens.Middleware(deps.unlikePost),
		FavoritePost:                deps.tokens.Middleware(deps.favoritePost),
		UnfavoritePost:              deps.tokens.Middleware(deps.unfavoritePost),
		ListNotifications:           deps.tokens.Middleware(deps.listNotifications),
		UnreadNotifications:         deps.tokens.Middleware(deps.unreadNotifications),
		MarkNotificationRead:        deps.tokens.Middleware(deps.markNotificationRead),
		MarkAllNotificationsRead:    deps.tokens.Middleware(deps.markAllNotificationsRead),
		PostEvents:                  deps.postEvents,
		AIStatus:                    deps.aiStatus,
		RetryAIReplies:              deps.retryAIReplies,
		ListAgentChatConversations:  deps.tokens.Middleware(deps.listAgentChatConversations),
		GetAgentChatConversation:    deps.tokens.Middleware(deps.getAgentChatConversation),
		StreamAgentChatMessage:      deps.tokens.Middleware(deps.streamAgentChatMessage),
		DeleteAgentChatConversation: deps.tokens.Middleware(deps.deleteAgentChatConversation),
		RetryAgentChatMessage:       deps.tokens.Middleware(deps.retryAgentChatMessage),
		SearchPosts:                 deps.searchPosts,
		AdminUpdatePostStatus:       deps.tokens.Middleware(authz.RequireSubject("post", "delete-any", deps.updatePostStatus)),
	}
	if deps.admin != nil {
		routes.ListAgents = http.HandlerFunc(deps.admin.ListPublicAgents)
		routes.ListAITasks = http.HandlerFunc(deps.admin.ListPublicTasks)
		routes.ListDecisionLogs = http.HandlerFunc(deps.admin.ListPublicDecisionLogs)
		routes.ListPostDecisionLogs = http.HandlerFunc(deps.admin.ListPostDecisionLogs)
		routes.ListPostAITasks = http.HandlerFunc(deps.admin.ListPostTasks)
		routes.ListAIActivities = http.HandlerFunc(deps.admin.ListActivities)
		routes.AdminDashboardStats = deps.tokens.Middleware(http.HandlerFunc(deps.admin.DashboardStats))
		routes.AdminDashboardTrend = deps.tokens.Middleware(http.HandlerFunc(deps.admin.WeeklyTrend))
		routes.AdminDashboardBreakdown = deps.tokens.Middleware(http.HandlerFunc(deps.admin.TaskStatusBreakdown))
		routes.AdminDashboardServices = deps.tokens.Middleware(http.HandlerFunc(deps.admin.Services))
		routes.AdminDashboardRecentPosts = deps.tokens.Middleware(http.HandlerFunc(deps.admin.RecentPosts))
		routes.AdminDashboardRecentTasks = deps.tokens.Middleware(http.HandlerFunc(deps.admin.RecentTasks))
		routes.AdminDashboardDecisions = deps.tokens.Middleware(http.HandlerFunc(deps.admin.DecisionTimeline))
		routes.AdminPermissions = deps.tokens.Middleware(http.HandlerFunc(deps.admin.Permissions))
		routes.AdminListUsers = deps.tokens.Middleware(http.HandlerFunc(deps.admin.ListUsers))
		routes.AdminListPosts = deps.tokens.Middleware(http.HandlerFunc(deps.admin.ListPosts))
		routes.AdminListComments = deps.tokens.Middleware(http.HandlerFunc(deps.admin.ListComments))
		routes.AdminListAgents = deps.tokens.Middleware(http.HandlerFunc(deps.admin.ListAgents))
		routes.AdminUpdateAgent = deps.tokens.Middleware(http.HandlerFunc(deps.admin.UpdateAgent))
		routes.AdminListTasks = deps.tokens.Middleware(http.HandlerFunc(deps.admin.ListTasks))
		routes.AdminRetryTask = deps.tokens.Middleware(http.HandlerFunc(deps.admin.RetryTask))
		routes.AdminTerminateTask = deps.tokens.Middleware(http.HandlerFunc(deps.admin.TerminateTask))
		routes.AdminMarkTaskProcessed = deps.tokens.Middleware(http.HandlerFunc(deps.admin.MarkTaskProcessed))
		routes.AdminListDecisionLogs = deps.tokens.Middleware(http.HandlerFunc(deps.admin.ListDecisionLogs))
		routes.AdminListTags = deps.tokens.Middleware(http.HandlerFunc(deps.admin.ListTags))
		routes.AdminListPreferences = deps.tokens.Middleware(http.HandlerFunc(deps.admin.ListPreferences))
	}
	return routes, nil
}

func mustAdminAuthorizer() *rbac.Authorizer {
	authz, err := rbac.NewAuthorizer(rbac.DefaultModelPath())
	if err != nil {
		panic(err)
	}
	if err := authz.SeedAdminPolicies(); err != nil {
		panic(err)
	}
	return authz
}

func (a *App) newChatHandler() *aichat.Handler {
	aiCfg := config.AIConfig{BaseURL: config.DefaultAIBaseURL, Model: "gpt-4o-mini"}
	if a.Cfg != nil {
		aiCfg = a.Cfg.AI
	}
	client := modelclient.NewObservedClient(
		modelclient.NewOpenAICompatibleClient(aiCfg.BaseURL, aiCfg.APIKey, aiCfg.Model, nil),
		a.Log,
		aiCfg.Model,
	)
	return aichat.NewHandler(aichat.NewService(aichat.NewSQLStore(a.DB), client))
}

// NewWorker wires the worker-service lifecycle harness and P6 task handlers.
func (a *App) NewWorker() Process {
	mux := asynq.NewServeMux()
	var consumers []workerRabbitConsumerSpec
	var ensureIndex func(context.Context) error
	if a.DB != nil {
		aiCfg := config.AIConfig{BaseURL: config.DefaultAIBaseURL, Model: "gpt-4o-mini"}
		if a.Cfg != nil {
			aiCfg = a.Cfg.AI
		}
		tagClient := modelclient.NewObservedClient(
			modelclient.NewOpenAICompatibleClient(aiCfg.BaseURL, aiCfg.APIKey, aiCfg.Model, nil),
			a.Log,
			aiCfg.Model,
		)
		handlers := task.Handlers{
			TagPost: tagging.NewSQLHandler(a.DB, tagging.NewModelTagger(tagClient, tagging.RuleTagger{})).HandleTagPost,
		}
		if a.AsynqClient != nil {
			enqueuer := task.NewAsynqEnqueuer(a.AsynqClient)
			decisionHandler := decision.NewSQLHandler(a.DB, task.NewGenerateAIReplyEnqueuer(enqueuer))
			decisionHandler.SetWillingnessScorer(decision.NewModelWillingnessScorer(tagClient))
			handlers.DecideAIReply = decisionHandler.HandleDecideAIReply
			consumers = workerRabbitConsumerSpecs(enqueuer, task.NewSQLProcessedStore(a.DB))
		}
		replyHandler := a.newReplyHandler()
		handlers.GenerateAIReply = func(ctx context.Context, payload task.GenerateAIReplyPayload) error {
			return replyHandler.HandleGenerateAIReply(ctx, reply.Task{PostID: payload.PostID, ParentCommentID: payload.ParentCommentID, AgentID: payload.AIAgentID, TriggerType: payload.TriggerType})
		}
		if a.AsynqClient != nil {
			followupHandler := a.newFollowupHandler()
			handlers.JudgeAIFollowup = followupHandler.HandleJudgeAIFollowup
		}
		if a.ES != nil {
			store := search.NewESIndexStore(a.ES)
			ensureIndex = store.EnsureIndex
			handlers.SyncSearchIndex = search.NewSyncHandler(a.DB, store).HandleSyncSearchIndex
		}
		handlers.SendNotification = notification.NewHandler(a.DB).HandleSendNotification
		if a.Redis != nil {
			hot := post.NewRedisHotStore(a.Redis, func(ctx context.Context, postID int64) (post.PostSnapshot, error) {
				return post.LoadPostSnapshot(ctx, a.DB, postID)
			})
			replyHandler.SetHotTracker(hot)
			handlers.RefreshHotScore = func(ctx context.Context) error {
				batchSize := 200
				if a.Cfg != nil && a.Cfg.HotScore.BatchSize > 0 {
					batchSize = a.Cfg.HotScore.BatchSize
				}
				_, err := hot.RefreshHotScores(ctx, a.DB, batchSize)
				return err
			}
		}
		task.RegisterHandlers(mux, a.DB, handlers)
	}
	return &WorkerProcess{name: "worker-service", log: a.Log, server: a.AsynqServer, scheduler: a.Scheduler, mux: mux, rabbit: a.RabbitMQ, consumers: consumers, ensureIndex: ensureIndex}
}

func (a *App) newReplyHandler() *reply.Handler {
	aiCfg := config.AIConfig{BaseURL: config.DefaultAIBaseURL, Model: "gpt-4o-mini", RequestPerSecond: 1, Burst: 1}
	if a.Cfg != nil {
		aiCfg = a.Cfg.AI
	}
	client := modelclient.NewObservedClient(
		modelclient.NewOpenAICompatibleClient(aiCfg.BaseURL, aiCfg.APIKey, aiCfg.Model, nil),
		a.Log,
		aiCfg.Model,
	)
	var limiter reply.Limiter = modelclient.NewTokenBucketLimiter(aiCfg.RequestPerSecond, aiCfg.Burst, time.Now)
	if a.Redis != nil {
		limiter = modelclient.NewRedisTokenBucketLimiter(a.Redis, "ai:reply:model", aiCfg.RequestPerSecond, aiCfg.Burst, time.Now)
	}
	handler := reply.NewSQLHandler(a.DB, client, moderation.NewRuleModerator(nil), limiter)
	if a.Cfg != nil {
		handler.SetNotifier(internalapi.NewClient(fmt.Sprintf("http://127.0.0.1:%d", a.Cfg.Server.Port), a.Cfg.InternalAPI.Token, nil))
	}
	return handler
}

func (a *App) newFollowupHandler() *followup.Handler {
	aiCfg := config.AIConfig{BaseURL: config.DefaultAIBaseURL, Model: "gpt-4o-mini"}
	if a.Cfg != nil {
		aiCfg = a.Cfg.AI
	}
	client := modelclient.NewObservedClient(
		modelclient.NewOpenAICompatibleClient(aiCfg.BaseURL, aiCfg.APIKey, aiCfg.Model, nil),
		a.Log,
		aiCfg.Model,
	)
	return followup.NewHandler(
		followup.NewSQLRepository(a.DB),
		followup.NewModelClient(client),
		task.NewGenerateAIReplyEnqueuer(task.NewAsynqEnqueuer(a.AsynqClient)),
	)
}

// NewOutboxPublisher wires the outbox-publisher scan loop.
func (a *App) NewOutboxPublisher() Process {
	var publisher *outbox.Publisher
	if a.DB != nil && a.RabbitMQ != nil {
		publisher = outbox.NewPublisher(a.DB, mq.NewPublisher(a.RabbitMQ), outbox.Options{})
	}
	return &OutboxProcess{
		name:      "outbox-publisher",
		log:       a.Log,
		rabbit:    a.RabbitMQ,
		publisher: publisher,
		done:      make(chan error, 1),
	}
}

// Close releases shared dependencies.
func (a *App) Close(ctx context.Context) error {
	var errs []error
	if a.AsynqClient != nil {
		if err := a.AsynqClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if a.Scheduler != nil {
		a.Scheduler.Shutdown()
	}
	if a.ES != nil {
		if err := a.ES.Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if a.RabbitMQ != nil {
		if err := a.RabbitMQ.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if a.DB != nil {
		if err := a.DB.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if a.Log != nil {
		if err := a.Log.Sync(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (a *App) readinessChecks() []router.Dependency {
	return []router.Dependency{
		{Name: "mysql", Check: func(ctx context.Context) error { return a.DB.PingContext(ctx) }},
		{Name: "redis", Check: func(ctx context.Context) error { return a.Redis.Ping(ctx).Err() }},
		{Name: "rabbitmq", Check: func(context.Context) error {
			ch, err := a.RabbitMQ.Channel()
			if err != nil {
				return err
			}
			return ch.Close()
		}},
		{Name: "elasticsearch", Check: func(ctx context.Context) error {
			res, err := a.ES.Ping(a.ES.Ping.WithContext(ctx))
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.IsError() {
				return fmt.Errorf("status %d", res.StatusCode)
			}
			return nil
		}},
	}
}

// IdleProcess is the P3 shutdown harness used until later phases attach real
// worker and outbox loops.
type IdleProcess struct {
	name    string
	log     *logger.Logger
	stopFn  func(context.Context) error
	stop    chan struct{}
	done    chan struct{}
	mu      sync.Mutex
	started bool
}

// HTTPProcess wraps http.Server with the common Process contract.
type HTTPProcess struct {
	name string
	srv  *http.Server
	log  *logger.Logger
	done chan error
}

type ErrorProcess struct {
	err error
}

type WorkerProcess struct {
	name          string
	log           *logger.Logger
	server        *asynq.Server
	scheduler     *asynq.Scheduler
	mux           *asynq.ServeMux
	rabbit        *mq.Connection
	consumers     []workerRabbitConsumerSpec
	ensureIndex   func(context.Context) error
	consumeCancel context.CancelFunc
	consumeDone   chan error
}

type OutboxProcess struct {
	name      string
	log       *logger.Logger
	rabbit    *mq.Connection
	publisher *outbox.Publisher
	done      chan error
}

type workerRabbitConsumerSpec struct {
	queue        string
	consumerName string
	handle       func(context.Context, []byte) error
}

func workerRabbitConsumerSpecs(enqueuer task.Enqueuer, processed task.ProcessedStore) []workerRabbitConsumerSpec {
	return []workerRabbitConsumerSpec{
		{
			queue:        mq.QueuePostTagging,
			consumerName: "worker.tag_post",
			handle: task.NewPostCreatedConsumer(
				enqueuer,
				task.WithProcessedStore(processed, "worker.tag_post"),
			).Handle,
		},
		{
			queue:        mq.QueueAIDecision,
			consumerName: "worker.decide_ai_reply",
			handle: task.NewPostTaggedConsumer(
				enqueuer,
				task.WithProcessedStore(processed, "worker.decide_ai_reply"),
			).Handle,
		},
		{
			queue:        mq.QueueSearchIndex,
			consumerName: "worker.sync_search_index",
			handle: task.NewSearchIndexConsumer(
				enqueuer,
				task.WithProcessedStore(processed, "worker.sync_search_index"),
			).Handle,
		},
		{
			queue:        mq.QueueNotification,
			consumerName: "worker.send_notification",
			handle: task.NewNotificationConsumer(
				enqueuer,
				task.WithProcessedStore(processed, "worker.send_notification"),
			).Handle,
		},
	}
}

func runRabbitConsumer(ctx context.Context, rabbit *mq.Connection, spec workerRabbitConsumerSpec) error {
	ch, err := rabbit.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	deliveries, err := ch.Consume(spec.queue, spec.consumerName, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume %s: %w", spec.queue, err)
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}
			if err := spec.handle(ctx, delivery.Body); err != nil {
				_ = delivery.Nack(false, true)
				continue
			}
			_ = delivery.Ack(false)
		}
	}
}

func NewErrorProcess(err error) *ErrorProcess {
	return &ErrorProcess{err: err}
}

func (p *ErrorProcess) Start(context.Context) error {
	return p.err
}

func (p *ErrorProcess) Stop(context.Context) error {
	return nil
}

func (p *WorkerProcess) Start(ctx context.Context) error {
	if p.log != nil {
		p.log.Info("process starting", zap.String("process", p.name))
	}
	if p.ensureIndex != nil {
		if err := p.ensureIndex(ctx); err != nil {
			return err
		}
	}
	if p.rabbit != nil && len(p.consumers) > 0 {
		if err := mq.DeclareTopology(p.rabbit); err != nil {
			return err
		}
		ctx, cancel := context.WithCancel(context.Background())
		p.consumeCancel = cancel
		p.consumeDone = make(chan error, len(p.consumers))
		for _, consumer := range p.consumers {
			go func(c workerRabbitConsumerSpec) {
				p.consumeDone <- runRabbitConsumer(ctx, p.rabbit, c)
			}(consumer)
		}
	}
	if p.scheduler != nil {
		if _, err := task.RegisterCleanupCron(p.scheduler); err != nil {
			return err
		}
		if _, err := task.RegisterRefreshHotScoreCron(p.scheduler); err != nil {
			return err
		}
		if err := p.scheduler.Start(); err != nil {
			return err
		}
	}
	if p.server != nil {
		return p.server.Start(p.mux)
	}
	return nil
}

func (p *WorkerProcess) Stop(context.Context) error {
	if p.consumeCancel != nil {
		p.consumeCancel()
	}
	if p.scheduler != nil {
		p.scheduler.Shutdown()
	}
	if p.server != nil {
		p.server.Shutdown()
	}
	var errs []error
	for range p.consumers {
		if p.consumeDone == nil {
			break
		}
		if err := <-p.consumeDone; err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, amqp.ErrClosed) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (p *OutboxProcess) Start(ctx context.Context) error {
	if p.log != nil {
		p.log.Info("process starting", zap.String("process", p.name))
	}
	if p.rabbit != nil {
		if err := mq.DeclareTopology(p.rabbit); err != nil {
			return err
		}
	}
	if p.publisher == nil {
		p.done <- nil
		return nil
	}
	go func() {
		p.done <- p.publisher.Start(ctx)
	}()
	return nil
}

func (p *OutboxProcess) Stop(ctx context.Context) error {
	if p.publisher == nil {
		select {
		case err := <-p.done:
			return err
		default:
			return nil
		}
	}
	return p.publisher.Stop(ctx)
}

// NewHTTPProcess returns a Process for an http.Server.
func NewHTTPProcess(name string, srv *http.Server, log *logger.Logger) *HTTPProcess {
	return &HTTPProcess{name: name, srv: srv, log: log, done: make(chan error, 1)}
}

// Start starts serving HTTP.
func (p *HTTPProcess) Start(context.Context) error {
	ln, err := net.Listen("tcp", p.srv.Addr)
	if err != nil {
		return fmt.Errorf("%s listen: %w", p.name, err)
	}
	if p.log != nil {
		p.log.Info("process starting", zap.String("process", p.name), zap.String("addr", p.srv.Addr))
	}
	go func() {
		err := p.srv.Serve(ln)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		p.done <- err
	}()
	return nil
}

// Stop stops accepting HTTP and waits for in-flight requests to drain.
func (p *HTTPProcess) Stop(ctx context.Context) error {
	if err := p.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s shutdown: %w", p.name, err)
	}
	select {
	case err := <-p.done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("%s wait: %w", p.name, ctx.Err())
	}
}

// NewIdleProcess returns a process that stays alive until Stop is called.
func NewIdleProcess(name string, stopFn func(context.Context) error) *IdleProcess {
	return NewLoggedIdleProcess(name, nil, stopFn)
}

// NewLoggedIdleProcess returns a logged idle process.
func NewLoggedIdleProcess(name string, log *logger.Logger, stopFn func(context.Context) error) *IdleProcess {
	return &IdleProcess{name: name, log: log, stopFn: stopFn, stop: make(chan struct{}), done: make(chan struct{})}
}

// Start starts the idle process loop.
func (p *IdleProcess) Start(context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.started {
		return fmt.Errorf("%s already started", p.name)
	}
	p.started = true
	if p.log != nil {
		p.log.Info("process starting", zap.String("process", p.name))
	}
	go func() {
		<-p.stop
		close(p.done)
	}()
	return nil
}

// Stop requests shutdown and waits for either cleanup or ctx timeout.
func (p *IdleProcess) Stop(ctx context.Context) error {
	p.mu.Lock()
	if !p.started {
		p.mu.Unlock()
		return nil
	}
	select {
	case <-p.stop:
	default:
		close(p.stop)
	}
	p.mu.Unlock()

	cleanupDone := make(chan error, 1)
	go func() {
		if p.stopFn != nil {
			cleanupDone <- p.stopFn(ctx)
			return
		}
		cleanupDone <- nil
	}()

	select {
	case <-p.done:
	case <-ctx.Done():
		return fmt.Errorf("%s shutdown: %w", p.name, ctx.Err())
	}

	select {
	case err := <-cleanupDone:
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return err
	case <-ctx.Done():
		if p.log != nil {
			p.log.Warn("abandoned work", zap.String("process", p.name), zap.Error(ctx.Err()))
		}
		return fmt.Errorf("%s cleanup: %w", p.name, ctx.Err())
	}
}
