package post

import (
	"context"
	"database/sql"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"ai-forum/backend/internal/cache"
)

func TestComputeHotScoreUsesSpecFormula(t *testing.T) {
	got := ComputeHotScore(HotCounters{Likes: 10, Comments: 5, AIReplies: 2, Views: 30}, 6)
	want := (10*2 + 5*3 + 2*2 + 30*0.1) / math.Pow(6+2, 1.2)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("score = %.12f, want %.12f", got, want)
	}
}

func TestRedisHotStoreInteractionRecreatesEmptyRedis(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 10})
	ctx := context.Background()
	require.NoError(t, client.Ping(ctx).Err())
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.FlushDB(ctx).Err())

	store := NewRedisHotStore(client, func(context.Context, int64) (PostSnapshot, error) {
		return PostSnapshot{ID: 42, CreatedAt: time.Now().Add(-2 * time.Hour)}, nil
	})

	require.NoError(t, store.RecordInteraction(ctx, 42, HotCounterLike, 1))

	likeCount, err := client.Get(ctx, cache.HotPostCounterKey(42, cache.HotCounterLike)).Int()
	require.NoError(t, err)
	if likeCount != 1 {
		t.Fatalf("like count = %d, want 1", likeCount)
	}
	dirty, err := client.SIsMember(ctx, cache.DirtyHotPostsSetKey, "42").Result()
	require.NoError(t, err)
	if !dirty {
		t.Fatal("post was not added to dirty hot set")
	}
	score, err := client.ZScore(ctx, cache.HotPostsZSetKey, "42").Result()
	require.NoError(t, err)
	if score <= 0 {
		t.Fatalf("zset score = %f, want positive", score)
	}
}

func TestRefreshHotScoresSnapshotsBatchAndLeavesRemainder(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 10})
	ctx := context.Background()
	require.NoError(t, client.Ping(ctx).Err())
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.FlushDB(ctx).Err())
	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	db := &hotScoreDB{snapshots: map[int64]PostSnapshot{}}
	for i := int64(1); i <= 201; i++ {
		db.snapshots[i] = PostSnapshot{ID: i, CreatedAt: now.Add(-time.Hour)}
		require.NoError(t, client.SAdd(ctx, cache.DirtyHotPostsSetKey, i).Err())
		require.NoError(t, client.Set(ctx, cache.HotPostCounterKey(i, cache.HotCounterLike), 2, 0).Err())
	}
	store := NewRedisHotStore(client, db.LoadPostSnapshot)
	store.now = func() time.Time { return now }

	n, err := store.RefreshHotScores(ctx, db, 200)
	require.NoError(t, err)

	if n != 200 || len(db.updates) != 200 {
		t.Fatalf("snapshots = %d updates=%d, want 200", n, len(db.updates))
	}
	remaining, err := client.SCard(ctx, cache.DirtyHotPostsSetKey).Result()
	require.NoError(t, err)
	if remaining != 1 {
		t.Fatalf("dirty remaining = %d, want 1", remaining)
	}
	if strings.Contains(db.query, "UPDATE posts SET hot_score") {
		t.Fatalf("query must update counters with hot_score snapshot, not hot_score-only hot path: %q", db.query)
	}
}

func TestRebuildHotScoresFromMySQLRestoresZSet(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 10})
	ctx := context.Background()
	require.NoError(t, client.Ping(ctx).Err())
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.FlushDB(ctx).Err())
	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	db := &hotScoreDB{recent: []PostSnapshot{
		{ID: 1, LikeCount: 3, CreatedAt: now.Add(-time.Hour)},
		{ID: 2, CommentCount: 4, CreatedAt: now.Add(-2 * time.Hour)},
	}}
	store := NewRedisHotStore(client, db.LoadPostSnapshot)
	store.now = func() time.Time { return now }

	n, err := store.RebuildHotScores(ctx, db)
	require.NoError(t, err)

	if n != 2 {
		t.Fatalf("rebuilt = %d, want 2", n)
	}
	got, err := client.ZRevRange(ctx, cache.HotPostsZSetKey, 0, -1).Result()
	require.NoError(t, err)
	if len(got) != 2 {
		t.Fatalf("zset = %#v, want 2 posts", got)
	}
}

type hotScoreDB struct {
	query     string
	snapshots map[int64]PostSnapshot
	recent    []PostSnapshot
	updates   []PostSnapshot
}

func (d *hotScoreDB) LoadPostSnapshot(_ context.Context, postID int64) (PostSnapshot, error) {
	return d.snapshots[postID], nil
}

func (d *hotScoreDB) ExecContext(_ context.Context, query string, args ...interface{}) (sql.Result, error) {
	d.query = query
	if len(args) >= 6 {
		d.updates = append(d.updates, PostSnapshot{
			ViewCount:    args[0].(int64),
			LikeCount:    args[1].(int64),
			CommentCount: args[2].(int64),
			AIReplyCount: args[3].(int64),
			ID:           args[5].(int64),
		})
	}
	return hotResult(1), nil
}

func (d *hotScoreDB) SelectContext(_ context.Context, dest interface{}, query string, _ ...interface{}) error {
	d.query = query
	*(dest.(*[]PostSnapshot)) = d.recent
	return nil
}

func (d *hotScoreDB) GetContext(context.Context, interface{}, string, ...interface{}) error {
	return nil
}

type hotResult int64

func (r hotResult) LastInsertId() (int64, error) { return 0, nil }
func (r hotResult) RowsAffected() (int64, error) { return int64(r), nil }
