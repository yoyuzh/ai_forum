package post

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"ai-forum/backend/internal/cache"
)

type HotCounter = cache.HotCounter

const (
	HotCounterView    = cache.HotCounterView
	HotCounterLike    = cache.HotCounterLike
	HotCounterComment = cache.HotCounterComment
	HotCounterAIReply = cache.HotCounterAIReply
)

type HotCounters struct {
	Views     int64
	Likes     int64
	Comments  int64
	AIReplies int64
}

type PostSnapshot struct {
	ID           int64     `db:"id"`
	ViewCount    int64     `db:"view_count"`
	LikeCount    int64     `db:"like_count"`
	CommentCount int64     `db:"comment_count"`
	AIReplyCount int64     `db:"ai_reply_count"`
	CreatedAt    time.Time `db:"created_at"`
}

type SnapshotLoader func(context.Context, int64) (PostSnapshot, error)

type HotScoreDB interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	SelectContext(context.Context, interface{}, string, ...interface{}) error
}

type RedisHotStore struct {
	client *redis.Client
	load   SnapshotLoader
	now    func() time.Time
}

func NewRedisHotStore(client *redis.Client, load SnapshotLoader) *RedisHotStore {
	return &RedisHotStore{client: client, load: load, now: time.Now}
}

func ComputeHotScore(c HotCounters, hoursSinceCreated float64) float64 {
	base := float64(c.Likes)*2 + float64(c.Comments)*3 + float64(c.AIReplies)*2 + float64(c.Views)*0.1
	return base / math.Pow(hoursSinceCreated+2, 1.2)
}

func (s *RedisHotStore) RecordInteraction(ctx context.Context, postID int64, counter HotCounter, delta int64) error {
	if postID <= 0 || delta == 0 {
		return fmt.Errorf("invalid hot interaction")
	}
	snapshot, err := s.load(ctx, postID)
	if err != nil {
		return err
	}
	keys := []string{
		cache.HotPostCounterKey(postID, cache.HotCounterView),
		cache.HotPostCounterKey(postID, cache.HotCounterLike),
		cache.HotPostCounterKey(postID, cache.HotCounterComment),
		cache.HotPostCounterKey(postID, cache.HotCounterAIReply),
	}
	values, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return err
	}
	pipe := s.client.TxPipeline()
	seeds := []int64{snapshot.ViewCount, snapshot.LikeCount, snapshot.CommentCount, snapshot.AIReplyCount}
	for i, value := range values {
		if value == nil {
			pipe.Set(ctx, keys[i], seeds[i], 0)
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	if _, err := s.client.IncrBy(ctx, cache.HotPostCounterKey(postID, counter), delta).Result(); err != nil {
		return err
	}
	counters, err := s.readCounters(ctx, postID)
	if err != nil {
		return err
	}
	hours := s.now().Sub(snapshot.CreatedAt).Hours()
	if hours < 0 {
		hours = 0
	}
	score := ComputeHotScore(counters, hours)
	postIDText := strconv.FormatInt(postID, 10)
	pipe = s.client.TxPipeline()
	pipe.Set(ctx, cache.HotPostScoreKey(postID), score, 0)
	pipe.ZAdd(ctx, cache.HotPostsZSetKey, redis.Z{Score: score, Member: postIDText})
	pipe.SAdd(ctx, cache.DirtyHotPostsSetKey, postIDText)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisHotStore) readCounters(ctx context.Context, postID int64) (HotCounters, error) {
	keys := []string{
		cache.HotPostCounterKey(postID, cache.HotCounterView),
		cache.HotPostCounterKey(postID, cache.HotCounterLike),
		cache.HotPostCounterKey(postID, cache.HotCounterComment),
		cache.HotPostCounterKey(postID, cache.HotCounterAIReply),
	}
	values, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return HotCounters{}, err
	}
	toInt := func(value any) (int64, error) {
		if value == nil {
			return 0, nil
		}
		return strconv.ParseInt(value.(string), 10, 64)
	}
	views, err := toInt(values[0])
	if err != nil {
		return HotCounters{}, err
	}
	likes, err := toInt(values[1])
	if err != nil {
		return HotCounters{}, err
	}
	comments, err := toInt(values[2])
	if err != nil {
		return HotCounters{}, err
	}
	aiReplies, err := toInt(values[3])
	if err != nil {
		return HotCounters{}, err
	}
	return HotCounters{Views: views, Likes: likes, Comments: comments, AIReplies: aiReplies}, nil
}

func (s *RedisHotStore) RefreshHotScores(ctx context.Context, db HotScoreDB, batchSize int) (int, error) {
	if batchSize <= 0 || batchSize > 200 {
		batchSize = 200
	}
	ids, err := s.client.SRandMemberN(ctx, cache.DirtyHotPostsSetKey, int64(batchSize)).Result()
	if err != nil {
		return 0, err
	}
	for _, idText := range ids {
		postID, err := strconv.ParseInt(idText, 10, 64)
		if err != nil {
			return 0, err
		}
		snapshot, err := s.load(ctx, postID)
		if err != nil {
			return 0, err
		}
		counters, err := s.readCounters(ctx, postID)
		if err != nil {
			return 0, err
		}
		hours := s.now().Sub(snapshot.CreatedAt).Hours()
		if hours < 0 {
			hours = 0
		}
		score := ComputeHotScore(counters, hours)
		if _, err := db.ExecContext(ctx, `
			UPDATE posts
			SET view_count = ?, like_count = ?, comment_count = ?, ai_reply_count = ?, hot_score = ?
			WHERE id = ?`,
			counters.Views, counters.Likes, counters.Comments, counters.AIReplies, score, postID); err != nil {
			return 0, err
		}
		if err := s.client.SRem(ctx, cache.DirtyHotPostsSetKey, idText).Err(); err != nil {
			return 0, err
		}
	}
	return len(ids), nil
}

func (s *RedisHotStore) RebuildHotScores(ctx context.Context, db HotScoreDB) (int, error) {
	var posts []PostSnapshot
	if err := db.SelectContext(ctx, &posts, `
		SELECT id, view_count, like_count, comment_count, ai_reply_count, created_at
		FROM posts
		WHERE status = 'NORMAL' AND deleted_at IS NULL AND created_at >= NOW() - INTERVAL 7 DAY`); err != nil {
		return 0, err
	}
	pipe := s.client.TxPipeline()
	for _, p := range posts {
		hours := s.now().Sub(p.CreatedAt).Hours()
		if hours < 0 {
			hours = 0
		}
		score := ComputeHotScore(HotCounters{
			Views:     p.ViewCount,
			Likes:     p.LikeCount,
			Comments:  p.CommentCount,
			AIReplies: p.AIReplyCount,
		}, hours)
		member := strconv.FormatInt(p.ID, 10)
		pipe.ZAdd(ctx, cache.HotPostsZSetKey, redis.Z{Score: score, Member: member})
		pipe.Set(ctx, cache.HotPostScoreKey(p.ID), score, 0)
		pipe.Set(ctx, cache.HotPostCounterKey(p.ID, cache.HotCounterView), p.ViewCount, 0)
		pipe.Set(ctx, cache.HotPostCounterKey(p.ID, cache.HotCounterLike), p.LikeCount, 0)
		pipe.Set(ctx, cache.HotPostCounterKey(p.ID, cache.HotCounterComment), p.CommentCount, 0)
		pipe.Set(ctx, cache.HotPostCounterKey(p.ID, cache.HotCounterAIReply), p.AIReplyCount, 0)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return len(posts), nil
}
