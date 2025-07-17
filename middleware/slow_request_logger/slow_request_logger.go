package slow_request_logger

import (
	"time"

	"github.com/primadi/lokstra"
)

const NAME = "lokstra.slow_request_logger"
const THRESHOLDKEY = "threshold"
const DEFAULT_THRESHOLD = 500 * time.Millisecond // Default threshold for slow requests

type SlowRequestLogger struct{}

// Name implements iface.MiddlewareModule.
func (r *SlowRequestLogger) Name() string {
	return NAME
}

// Meta implements iface.MiddlewareModule.
func (r *SlowRequestLogger) Meta() *lokstra.MiddlewareMeta {
	return &lokstra.MiddlewareMeta{
		Priority:    20,
		Description: "logs slow requests and their metadata.",
		Tags:        []string{"logging", "request"},
	}
}

// Factory implements iface.MiddlewareModule.
func (r *SlowRequestLogger) Factory(config any) lokstra.MiddlewareFunc {
	dur_th := DEFAULT_THRESHOLD
	switch cfg := config.(type) {
	case map[string]any:
		if thany, ok := cfg[THRESHOLDKEY]; ok {
			if th, ok := thany.(string); ok {
				if dur, err := time.ParseDuration(th); err == nil {
					dur_th = dur
				}
			}
		}
	case map[string]string:
		if threshold, ok := cfg[THRESHOLDKEY]; ok {
			if dur, err := time.ParseDuration(threshold); err == nil {
				dur_th = dur
			}
		}
	case string:
		if dur, err := time.ParseDuration(cfg); err == nil {
			dur_th = dur
		}
	}

	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {

			startTime := time.Now()
			defer func() {
				duration := time.Since(startTime)
				if duration >= dur_th {
					lokstra.Logger.WithFields(lokstra.LogFields{
						"method":   ctx.Request.Method,
						"path":     ctx.Request.URL.Path,
						"duration": duration.String(),
						"status":   ctx.Response.StatusCode}).
						Infof("Slow request detected")
				}
			}()

			// Call the next handler in the chain
			return next(ctx)
		}
	}
}

var _ lokstra.MiddlewareModule = (*SlowRequestLogger)(nil)

// return SlowRequestLogger with name "lokstra.slow_request_logger"
func GetModule() lokstra.MiddlewareModule {
	return &SlowRequestLogger{}
}
