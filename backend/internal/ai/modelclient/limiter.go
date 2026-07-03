package modelclient

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBucketLimiter struct {
	mu     sync.Mutex
	rps    float64
	burst  float64
	tokens float64
	last   time.Time
	now    func() time.Time
}

func NewTokenBucketLimiter(rps, burst int, now func() time.Time) *TokenBucketLimiter {
	if rps <= 0 {
		rps = 1
	}
	if burst <= 0 {
		burst = 1
	}
	if now == nil {
		now = time.Now
	}
	return &TokenBucketLimiter{rps: float64(rps), burst: float64(burst), tokens: float64(burst), last: now(), now: now}
}

func (l *TokenBucketLimiter) Allow(context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	elapsed := now.Sub(l.last).Seconds()
	l.last = now
	l.tokens = min(l.burst, l.tokens+elapsed*l.rps)
	if l.tokens < 1 {
		return false, nil
	}
	l.tokens--
	return true, nil
}

type RedisTokenBucketLimiter struct {
	client *redis.Client
	key    string
	rps    int
	burst  int
	now    func() time.Time
}

func NewRedisTokenBucketLimiter(client *redis.Client, key string, rps, burst int, now func() time.Time) *RedisTokenBucketLimiter {
	if rps <= 0 {
		rps = 1
	}
	if burst <= 0 {
		burst = 1
	}
	if now == nil {
		now = time.Now
	}
	return &RedisTokenBucketLimiter{client: client, key: key, rps: rps, burst: burst, now: now}
}

func (l *RedisTokenBucketLimiter) Allow(ctx context.Context) (bool, error) {
	script := redis.NewScript(`
local tokens_key = KEYS[1] .. ':tokens'
local ts_key = KEYS[1] .. ':ts'
local rps = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local tokens = tonumber(redis.call('GET', tokens_key) or burst)
local ts = tonumber(redis.call('GET', ts_key) or now)
tokens = math.min(burst, tokens + math.max(0, now - ts) * rps)
local allowed = 0
if tokens >= 1 then
  tokens = tokens - 1
  allowed = 1
end
redis.call('SET', tokens_key, tokens, 'EX', 60)
redis.call('SET', ts_key, now, 'EX', 60)
return allowed
`)
	got, err := script.Run(ctx, l.client, []string{l.key}, l.rps, l.burst, l.now().Unix()).Int()
	if err != nil {
		return false, err
	}
	return got == 1, nil
}
