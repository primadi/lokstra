package timeout

import (
	"context"
	"lokstra"
	"time"
)

const NAME = "lokstra.timeout"

type TimeoutMiddleware struct{}

func (t *TimeoutMiddleware) Name() string {
	return NAME
}

func (t *TimeoutMiddleware) Meta() *lokstra.MiddlewareMeta {
	return &lokstra.MiddlewareMeta{
		Priority:    15,
		Description: "Request timeout middleware to prevent long-running requests",
		Tags:        []string{"timeout", "performance", "reliability"},
	}
}

func (t *TimeoutMiddleware) Factory(config any) lokstra.MiddlewareFunc {
	configMap := make(map[string]any)
	if cfg, ok := config.(map[string]any); ok {
		configMap = cfg
	}

	timeoutSeconds := 30
	if ts, ok := configMap["timeout_seconds"].(int); ok {
		timeoutSeconds = ts
	}

	timeout := time.Duration(timeoutSeconds) * time.Second

	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			timeoutCtx, cancel := context.WithTimeout(ctx.Context(), timeout)
			defer cancel()

			ctx.SetContext(timeoutCtx)

			done := make(chan error, 1)
			go func() {
				done <- next(ctx)
			}()

			select {
			case err := <-done:
				return err
			case <-timeoutCtx.Done():
				return ctx.ErrorRequestTimeout("Request timeout")
			}
		}
	}
}

var _ lokstra.MiddlewareModule = (*TimeoutMiddleware)(nil)

func GetModule() lokstra.MiddlewareModule {
	return &TimeoutMiddleware{}
}
