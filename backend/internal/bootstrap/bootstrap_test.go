package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"

	"ai-forum/backend/internal/auth"
	"ai-forum/backend/internal/config"
	"ai-forum/backend/internal/logger"
	"ai-forum/backend/internal/router"
	"ai-forum/backend/internal/task"
)

func TestProcessStopReturnsAndDoesNotLeakGoroutines(t *testing.T) {
	before := runtime.NumGoroutine()
	const tolerance = 5
	p := NewIdleProcess("test", func(context.Context) error { return nil })

	if err := p.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	started := runtime.NumGoroutine()
	if started <= before {
		t.Fatalf("expected process goroutine to start, before=%d started=%d", before, started)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := p.Stop(ctx); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if after := runtime.NumGoroutine(); after <= before+tolerance {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("goroutines did not return within tolerance: before=%d after=%d", before, runtime.NumGoroutine())
}

func TestProcessStopHonorsTimeout(t *testing.T) {
	p := NewIdleProcess("test", func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	if err := p.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := p.Stop(ctx)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if time.Since(start) > 250*time.Millisecond {
		t.Fatalf("stop exceeded timeout: %s", time.Since(start))
	}
}

func TestIdleProcessLogsStartupAndAbandonedWork(t *testing.T) {
	var logs bytes.Buffer
	log := testProcessLogger(t, &logs)
	p := NewLoggedIdleProcess("worker-service", log, func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})

	if err := p.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := p.Stop(ctx); err == nil {
		t.Fatal("expected timeout error")
	}

	got := logs.String()
	for _, want := range []string{"process starting", "worker-service", "abandoned work"} {
		if !strings.Contains(got, want) {
			t.Fatalf("log missing %q: %s", want, got)
		}
	}
}

func TestWorkerAndOutboxProcessesStartAndStop(t *testing.T) {
	var logs bytes.Buffer
	app := &App{Log: testProcessLogger(t, &logs)}

	for _, p := range []Process{app.NewWorker(), app.NewOutboxPublisher()} {
		if err := p.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := p.Stop(ctx); err != nil {
			cancel()
			t.Fatal(err)
		}
		cancel()
	}

	got := logs.String()
	for _, want := range []string{"worker-service", "outbox-publisher"} {
		if !strings.Contains(got, want) {
			t.Fatalf("startup log missing %q: %s", want, got)
		}
	}
}

func TestNewWorkerRegistersP6TaskHandlersWhenDependenciesExist(t *testing.T) {
	db, err := sqlx.Open("mysql", "root:bad@tcp(127.0.0.1:1)/ai_forum?parseTime=true")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app := &App{DB: db, AsynqClient: asynq.NewClient(asynq.RedisClientOpt{Addr: "127.0.0.1:1"})}
	defer app.AsynqClient.Close()
	worker := app.NewWorker()
	p, ok := worker.(*WorkerProcess)
	if !ok {
		t.Fatalf("worker type = %T, want *WorkerProcess", worker)
	}

	tagPayload, err := json.Marshal(task.TagPostPayload{PostID: 42})
	if err != nil {
		t.Fatal(err)
	}
	err = p.mux.ProcessTask(context.Background(), asynq.NewTask(task.TagPost, tagPayload))
	if err == nil {
		t.Fatal("expected tag_post to fail against unavailable test DB")
	}
	if errors.Is(err, asynq.ErrHandlerNotFound) {
		t.Fatalf("tag_post was not registered: %v", err)
	}

	decisionPayload, err := json.Marshal(task.DecideAIReplyPayload{PostID: 42})
	if err != nil {
		t.Fatal(err)
	}
	err = p.mux.ProcessTask(context.Background(), asynq.NewTask(task.DecideAIReply, decisionPayload))
	if err == nil {
		t.Fatal("expected decide_ai_reply to fail against unavailable test DB")
	}
	if errors.Is(err, asynq.ErrHandlerNotFound) {
		t.Fatalf("decide_ai_reply was not registered: %v", err)
	}
}

func TestWorkerRabbitConsumerSpecsBindP6Queues(t *testing.T) {
	specs := workerRabbitConsumerSpecs(&task.AsynqEnqueuer{}, nil)

	if len(specs) != 2 {
		t.Fatalf("consumer specs = %d, want 2", len(specs))
	}
	if specs[0].queue != "q.post.tagging" || specs[0].consumerName != "worker.tag_post" {
		t.Fatalf("first spec = %#v, want q.post.tagging worker.tag_post", specs[0])
	}
	if specs[1].queue != "q.ai.decision" || specs[1].consumerName != "worker.decide_ai_reply" {
		t.Fatalf("second spec = %#v, want q.ai.decision worker.decide_ai_reply", specs[1])
	}
}

func TestAdminPostStatusRouteRequiresAdminRole(t *testing.T) {
	tokens := auth.NewTokenManager("secret", time.Hour)
	routes, err := businessRoutes(businessRouteDeps{
		tokens: tokens,
		register: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}),
		login: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		profile: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		listPosts: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("posts"))
		}),
		getPost: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("post"))
		}),
		createPost: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}),
		updatePost: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		deletePost: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		createComment: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}),
		updatePostStatus: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	h := router.NewWithBusinessRoutes(nil, nil, routes)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/posts", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "posts" {
		t.Fatalf("public list status/body = %d/%q", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/api/posts/42", strings.NewReader(`{"title":"t","content":"c"}`)))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous update status = %d, want 401", rec.Code)
	}

	userToken, err := tokens.Issue(auth.Subject{UserID: 1, Username: "alice", Role: "USER"})
	if err != nil {
		t.Fatal(err)
	}
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/posts/42", strings.NewReader(`{"title":"t","content":"c"}`))
	req.Header.Set("Authorization", "Bearer "+userToken)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("authenticated update status = %d, want 200", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/posts/42", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("USER delete status = %d, want 403", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/admin/posts/42/status", strings.NewReader(`{"status":"HIDDEN"}`))
	req.Header.Set("Authorization", "Bearer "+userToken)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("USER status = %d, want 403", rec.Code)
	}

	adminToken, err := tokens.Issue(auth.Subject{UserID: 2, Username: "admin", Role: "ADMIN"})
	if err != nil {
		t.Fatal(err)
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/admin/posts/42/status", strings.NewReader(`{"status":"HIDDEN"}`))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("ADMIN status = %d, want 204", rec.Code)
	}
}

func TestHTTPProcessStartReturnsListenError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	p := NewHTTPProcess("api-server", &http.Server{Addr: ln.Addr().String()}, nil)
	err = p.Start(context.Background())

	if err == nil {
		t.Fatal("expected listen error")
	}
	if !strings.Contains(err.Error(), "bind") && !strings.Contains(err.Error(), "address already in use") {
		t.Fatalf("expected bind/listen error, got %v", err)
	}
}

func testProcessLogger(t *testing.T, buf *bytes.Buffer) *logger.Logger {
	t.Helper()
	l, err := logger.NewWithWriter(config.LogConfig{Level: "info", Encoding: "json"}, buf)
	if err != nil {
		t.Fatal(err)
	}
	return l
}
