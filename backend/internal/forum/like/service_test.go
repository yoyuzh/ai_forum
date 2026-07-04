package like

import (
	"context"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"ai-forum/backend/internal/outbox"
)

func TestServiceLikeUnlikeCountAndOutbox(t *testing.T) {
	var tx DBTX
	repo := &recordingRepository{changed: true}
	hot := &recordingHotTracker{}
	var events []outbox.Event
	svc := NewService(repo, func(ctx context.Context, _ DBTX, event outbox.Event) error {
		events = append(events, event)
		return nil
	}, WithHotTracker(hot))

	if err := svc.Like(context.Background(), tx, 7, 42); err != nil {
		t.Fatal(err)
	}
	if err := svc.Unlike(context.Background(), tx, 7, 42); err != nil {
		t.Fatal(err)
	}

	count, err := svc.Count(context.Background(), tx, 42)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 || !repo.liked || !repo.unliked {
		t.Fatalf("count=%d liked=%v unliked=%v", count, repo.liked, repo.unliked)
	}
	if len(events) != 2 || events[0].EventType != "post.liked" || events[1].EventType != "post.unliked" {
		t.Fatalf("events = %#v", events)
	}
	if len(hot.deltas) != 2 || hot.deltas[0] != 1 || hot.deltas[1] != -1 {
		t.Fatalf("hot deltas = %#v, want +1/-1", hot.deltas)
	}

	repo.changed = false
	if err := svc.Like(context.Background(), tx, 7, 42); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("noop like appended outbox event: %#v", events)
	}
}

func TestServiceParallelLikesP99AndNoHotScoreMySQLWrite(t *testing.T) {
	repo := &parallelRepository{}
	hot := &parallelHotTracker{}
	svc := NewService(repo, func(context.Context, DBTX, outbox.Event) error { return nil }, WithHotTracker(hot))
	const n = 100
	latencies := make([]time.Duration, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			start := time.Now()
			if err := svc.Like(context.Background(), nil, int64(i+1), 42); err != nil {
				t.Errorf("like: %v", err)
			}
			latencies[i] = time.Since(start)
		}()
	}
	wg.Wait()
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p99 := latencies[int(float64(n)*0.99)-1]
	if p99 >= 50*time.Millisecond {
		t.Fatalf("p99 = %s, want < 50ms", p99)
	}
	if repo.hotScoreUpdates != 0 {
		t.Fatalf("hot score MySQL updates = %d, want 0", repo.hotScoreUpdates)
	}
	if hot.calls != n {
		t.Fatalf("hot tracker calls = %d, want %d", hot.calls, n)
	}
}

type recordingRepository struct {
	liked   bool
	unliked bool
	count   int
	changed bool
}

func (r *recordingRepository) Like(context.Context, DBTX, int64, int64) (bool, error) {
	r.liked = true
	if !r.changed {
		return false, nil
	}
	r.count++
	return true, nil
}

func (r *recordingRepository) Unlike(context.Context, DBTX, int64, int64) (bool, error) {
	r.unliked = true
	r.count--
	return true, nil
}

func (r *recordingRepository) Count(context.Context, DBTX, int64) (int, error) {
	return r.count, nil
}

type recordingHotTracker struct {
	deltas []int64
}

func (h *recordingHotTracker) RecordInteraction(_ context.Context, postID int64, counter HotCounter, delta int64) error {
	if postID != 42 || counter != HotCounterLike {
		return nil
	}
	h.deltas = append(h.deltas, delta)
	return nil
}

type parallelRepository struct {
	mu              sync.Mutex
	seen            map[int64]bool
	hotScoreUpdates int
}

func (r *parallelRepository) Like(_ context.Context, _ DBTX, userID, _ int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.seen == nil {
		r.seen = map[int64]bool{}
	}
	if r.seen[userID] {
		return false, nil
	}
	r.exec("INSERT INTO likes")
	r.seen[userID] = true
	return true, nil
}

func (r *parallelRepository) Unlike(context.Context, DBTX, int64, int64) (bool, error) {
	return true, nil
}

func (r *parallelRepository) Count(context.Context, DBTX, int64) (int, error) {
	return 0, nil
}

func (r *parallelRepository) exec(query string) {
	if strings.Contains(query, "UPDATE posts SET hot_score") {
		r.hotScoreUpdates++
	}
}

type parallelHotTracker struct {
	mu    sync.Mutex
	calls int
}

func (h *parallelHotTracker) RecordInteraction(context.Context, int64, HotCounter, int64) error {
	h.mu.Lock()
	h.calls++
	h.mu.Unlock()
	return nil
}
