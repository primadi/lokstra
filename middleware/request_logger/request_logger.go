package request_logger

import (
	"lokstra"
	"time"
)

const NAME = "lokstra.request_logger"

type RequestLogger struct{}

// Name implements iface.MiddlewareModule.
func (r *RequestLogger) Name() string {
	return NAME
}

// Meta implements iface.MiddlewareModule.
func (r *RequestLogger) Meta() *lokstra.MiddlewareMeta {
	return &lokstra.MiddlewareMeta{
		Priority:    20,
		Description: "Logs incoming requests and their metadata.",
		Tags:        []string{"logging", "request"},
	}
}

// Factory implements iface.MiddlewareModule.
func (r *RequestLogger) Factory(config any) lokstra.MiddlewareFunc {
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

var _ lokstra.MiddlewareModule = (*RequestLogger)(nil)

// return RequestLogger with name "lokstra.request_logger"
func GetModule() lokstra.MiddlewareModule {
	return &RequestLogger{}
}
