package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb    *redis.Client
	max    int
	window time.Duration
}

func NewRateLimiter(rdb *redis.Client, max int, window time.Duration) *RateLimiter {
	return &RateLimiter{rdb: rdb, max: max, window: window}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		key := fmt.Sprintf("ratelimit:%s", ip)

		count, err := rl.increment(r.Context(), key)
		if err != nil {
			// Redis down — fail open, don't block legitimate users
			next.ServeHTTP(w, r)
			return
		}

		if count > rl.max {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", rl.window.Seconds()))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limit exceeded"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) increment(ctx context.Context, key string) (int, error) {
	pipe := rl.rdb.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, rl.window)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return int(incr.Val()), nil
}

func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return r.RemoteAddr
}