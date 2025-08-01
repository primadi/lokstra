package slow_request_logger

import (
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/iface"
)

const NAME = "slow_request_logger"
const THRESHOLDKEY = "threshold"
const DEFAULT_THRESHOLD = 500 * time.Millisecond // Default threshold for slow requests

type SlowRequestLogger struct{}

// Description implements registration.Module.
func (r *SlowRequestLogger) Description() string {
	return "Logs slow requests and their metadata."
}

// Register implements registration.Module.
func (r *SlowRequestLogger) Register(regCtx iface.RegistrationContext) error {
	return regCtx.RegisterMiddlewareFactoryWithPriority(NAME, factory, 20)
}

// Name implements registration.Module.
func (r *SlowRequestLogger) Name() string {
	return NAME
}

func factory(config any) lokstra.MiddlewareFunc {
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

var _ lokstra.Module = (*SlowRequestLogger)(nil)

// return SlowRequestLogger with name "lokstra.slow_request_logger"
func GetModule() lokstra.Module {
	return &SlowRequestLogger{}
}
