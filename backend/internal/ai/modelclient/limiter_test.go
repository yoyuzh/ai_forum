package modelclient

import (
	"context"
	"testing"
	"time"
)

func TestMemoryLimiterAllowsBurstThenRefills(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := NewTokenBucketLimiter(1, 2, func() time.Time { return now })

	for i := 0; i < 2; i++ {
		ok, err := limiter.Allow(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("allow %d = false, want true", i)
		}
	}
	ok, err := limiter.Allow(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("third burst request allowed, want rate limited")
	}

	now = now.Add(time.Second)
	ok, err = limiter.Allow(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("request after refill denied")
	}
}
