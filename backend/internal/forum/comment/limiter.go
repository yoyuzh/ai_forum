package comment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type MemoryMentionLimiter struct {
	limit  int
	window time.Duration
	now    func() time.Time
	mu     sync.Mutex
	hits   map[int64][]time.Time
}

func NewMemoryMentionLimiter(limit int, window time.Duration, now func() time.Time) *MemoryMentionLimiter {
	return &MemoryMentionLimiter{limit: limit, window: window, now: now, hits: map[int64][]time.Time{}}
}

func (l *MemoryMentionLimiter) AllowMentions(_ context.Context, userID int64, count int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	cutoff := now.Add(-l.window)
	var kept []time.Time
	for _, hit := range l.hits[userID] {
		if hit.After(cutoff) {
			kept = append(kept, hit)
		}
	}
	if len(kept)+count > l.limit {
		l.hits[userID] = kept
		return ErrMentionRateLimited
	}
	for i := 0; i < count; i++ {
		kept = append(kept, now)
	}
	l.hits[userID] = kept
	return nil
}

type RedisMentionLimiter struct {
	client *redis.Client
	limit  int
	ttl    time.Duration
	now    func() time.Time
}

func NewRedisMentionLimiter(client *redis.Client, limit int, ttl time.Duration) *RedisMentionLimiter {
	return &RedisMentionLimiter{client: client, limit: limit, ttl: ttl, now: time.Now}
}

func (l *RedisMentionLimiter) AllowMentions(ctx context.Context, userID int64, count int) error {
	key := fmt.Sprintf("comment:ai_mentions:%d", userID)
	now := l.now().UnixMilli()
	windowStart := now - l.ttl.Milliseconds()
	script := redis.NewScript(`
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
local current = redis.call('ZCARD', KEYS[1])
if current + tonumber(ARGV[3]) > tonumber(ARGV[4]) then
  return 0
end
for i = 1, tonumber(ARGV[3]) do
  redis.call('ZADD', KEYS[1], ARGV[2], ARGV[2] .. ':' .. (current + i))
end
redis.call('PEXPIRE', KEYS[1], ARGV[5])
return 1
`)
	ok, err := script.Run(ctx, l.client, []string{key}, windowStart, now, count, l.limit, l.ttl.Milliseconds()).Int()
	if err != nil {
		return err
	}
	if ok != 1 {
		return ErrMentionRateLimited
	}
	return nil
}
