package request_logger

import (
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/registration"
)

const NAME = "request_logger"

type RequestLogger struct{}

// Description implements registration.Module.
func (r *RequestLogger) Description() string {
	return "Logs incoming requests and their metadata."
}

// Register implements registration.Module.
func (r *RequestLogger) Register(regCtx registration.Context) error {
	return regCtx.RegisterMiddlewareFactoryWithPriority(NAME, factory, 20)
}

// Name implements registration.Module.
func (r *RequestLogger) Name() string {
	return NAME
}

func factory(config any) lokstra.MiddlewareFunc {
	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			// Log the request details
			logger := lokstra.Logger.WithFields(lokstra.LogFields{
				"method": ctx.Request.Method,
				"path":   ctx.Request.URL.Path})

			logger.Infof("Incoming request")

			startTime := time.Now()
			defer func() {
				duration := time.Since(startTime)
				logger.WithFields(lokstra.LogFields{
					"duration": duration.String(),
					"status":   ctx.Response.StatusCode}).
					Infof("Request completed")
			}()

			// Call the next handler in the chain
			return next(ctx)
		}
	}
}

var _ lokstra.Module = (*RequestLogger)(nil)

func GetModule() lokstra.Module {
	return &RequestLogger{}
}
