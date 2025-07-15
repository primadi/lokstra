package ratelimit

import (
	"fmt"
	"lokstra"
	"sync"
	"time"
)

const NAME = "lokstra.ratelimit"

type RateLimitMiddleware struct{}

type rateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) isAllowed(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	requests := rl.requests[key]
	
	var validRequests []time.Time
	for _, req := range requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}

	if len(validRequests) >= rl.limit {
		rl.requests[key] = validRequests
		return false
	}

	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	return true
}

func (r *RateLimitMiddleware) Name() string {
	return NAME
}

func (r *RateLimitMiddleware) Meta() *lokstra.MiddlewareMeta {
	return &lokstra.MiddlewareMeta{
		Priority:    30,
		Description: "Rate limiting middleware to prevent abuse and ensure fair usage",
		Tags:        []string{"ratelimit", "security", "throttling"},
	}
}

func (r *RateLimitMiddleware) Factory(config any) lokstra.MiddlewareFunc {
	configMap := make(map[string]any)
	if cfg, ok := config.(map[string]any); ok {
		configMap = cfg
	}

	limit := 100
	if l, ok := configMap["limit"].(int); ok {
		limit = l
	}

	windowMinutes := 1
	if w, ok := configMap["window_minutes"].(int); ok {
		windowMinutes = w
	}

	keyFunc := "ip"
	if k, ok := configMap["key"].(string); ok {
		keyFunc = k
	}

	limiter := newRateLimiter(limit, time.Duration(windowMinutes)*time.Minute)

	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			var key string
			switch keyFunc {
			case "ip":
				key = ctx.ClientIP()
			case "user":
				if userID := ctx.Get("user_id"); userID != nil {
					key = fmt.Sprintf("user:%v", userID)
				} else {
					key = ctx.ClientIP()
				}
			default:
				key = ctx.ClientIP()
			}

			if !limiter.isAllowed(key) {
				return ctx.ErrorTooManyRequests("Rate limit exceeded")
			}

			return next(ctx)
		}
	}
}

var _ lokstra.MiddlewareModule = (*RateLimitMiddleware)(nil)

func GetModule() lokstra.MiddlewareModule {
	return &RateLimitMiddleware{}
}
