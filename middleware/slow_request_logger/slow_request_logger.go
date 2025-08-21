package slow_request_logger

import (
	"time"

	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

const NAME = "slow_request_logger"
const THRESHOLDKEY = "threshold"
const DEFAULT_THRESHOLD = 500 * time.Millisecond // Default threshold for slow requests

type SlowRequestLogger struct{}

// Description implements registration.Module.
func (r *SlowRequestLogger) Description() string {
	return "Logs slow requests and their metadata."
}

var logger serviceapi.Logger

// Register implements registration.Module.
func (r *SlowRequestLogger) Register(regCtx iface.RegistrationContext) error {
	if svc, err := regCtx.GetService("logger.default"); err == nil {
		logger = svc.(serviceapi.Logger)
	}
	return regCtx.RegisterMiddlewareFactoryWithPriority(NAME, factory, 20)
}

// Name implements registration.Module.
func (r *SlowRequestLogger) Name() string {
	return NAME
}

func factory(config any) midware.Func {
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

	return func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {

			startTime := time.Now()
			defer func() {
				duration := time.Since(startTime)
				if duration >= dur_th {
					logger.WithFields(serviceapi.LogFields{
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

var _ iface.Module = (*SlowRequestLogger)(nil)

// return SlowRequestLogger with name "lokstra.slow_request_logger"
func GetModule() iface.Module {
	return &SlowRequestLogger{}
}
