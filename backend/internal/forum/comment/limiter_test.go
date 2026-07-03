package comment

import (
	"context"
	"testing"
	"time"
)

func TestMemoryMentionLimiterRejectsMoreThanFivePerMinute(t *testing.T) {
	limiter := NewMemoryMentionLimiter(5, time.Minute, func() time.Time { return time.Unix(100, 0) })

	if err := limiter.AllowMentions(context.Background(), 7, 5); err != nil {
		t.Fatal(err)
	}
	if err := limiter.AllowMentions(context.Background(), 7, 1); err != ErrMentionRateLimited {
		t.Fatalf("err = %v, want ErrMentionRateLimited", err)
	}
}
