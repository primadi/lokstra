package request_logger

import (
	"time"

	"github.com/primadi/lokstra/common/json"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

const NAME = "request_logger"

// Config holds the configuration for request logger middleware
type Config struct {
	IncludeRequestBody  bool `json:"include_request_body" yaml:"include_request_body"`
	IncludeResponseBody bool `json:"include_response_body" yaml:"include_response_body"`
}

type RequestLogger struct{}

// Description implements registration.Module.
func (r *RequestLogger) Description() string {
	return "Logs incoming requests and their metadata with optional request/response body logging."
}

var logger serviceapi.Logger

// Register implements registration.Module.
func (r *RequestLogger) Register(regCtx registration.Context) error {
	if svc, err := regCtx.GetService("logger"); err == nil {
		logger = svc.(serviceapi.Logger)
	}
	return regCtx.RegisterMiddlewareFactoryWithPriority(NAME, factory, 20)
}

// Name implements registration.Module.
func (r *RequestLogger) Name() string {
	return NAME
}

func factory(config any) midware.Func {
	// Parse configuration
	cfg := &Config{
		IncludeRequestBody:  false,
		IncludeResponseBody: false,
	}

	if config != nil {
		switch v := config.(type) {
		case map[string]any:
			if val, ok := v["include_request_body"]; ok {
				if b, ok := val.(bool); ok {
					cfg.IncludeRequestBody = b
				}
			}
			if val, ok := v["include_response_body"]; ok {
				if b, ok := val.(bool); ok {
					cfg.IncludeResponseBody = b
				}
			}
		case *Config:
			cfg = v
		case Config:
			cfg = &v
		}
	}

	return func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			// Prepare base log fields
			logFields := serviceapi.LogFields{
				"method":     ctx.Request.Method,
				"path":       ctx.Request.URL.Path,
				"query":      ctx.Request.URL.RawQuery,
				"remote_ip":  ctx.Request.RemoteAddr,
				"user_agent": ctx.Request.Header.Get("User-Agent"),
			}

			// Include request body if configured
			if cfg.IncludeRequestBody {
				bodyBytes, err := ctx.GetRawRequestBody()
				if err == nil && len(bodyBytes) > 0 {
					// Try to parse as JSON for better logging
					var jsonBody any
					if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
						logFields["request_body"] = jsonBody
					} else {
						// If not JSON, log as string (truncate if too long)
						bodyStr := string(bodyBytes)
						if len(bodyStr) > 1000 {
							bodyStr = bodyStr[:1000] + "... (truncated)"
						}
						logFields["request_body"] = bodyStr
					}
				}
			}

			// Check if logger is available, if not skip logging
			if logger == nil {
				panic("Logger service not available for request_logger middleware")
			}

			loggerWithFields := logger.WithFields(logFields)
			loggerWithFields.Infof("Incoming request")

			startTime := time.Now()

			// Execute the request
			err := next(ctx)

			// Log completion
			duration := time.Since(startTime)
			completionFields := serviceapi.LogFields{
				"duration":    duration.String(),
				"duration_ms": duration.Milliseconds(),
				"status":      ctx.Response.StatusCode,
			}

			// Include response body if configured
			if cfg.IncludeResponseBody {
				responseBody, err := ctx.GetRawResponseBody()
				if err == nil && len(responseBody) > 0 {
					var jsonBody any
					if err := json.Unmarshal(responseBody, &jsonBody); err == nil {
						completionFields["response_body"] = jsonBody
					} else {
						bodyStr := string(responseBody)
						if len(bodyStr) > 1000 {
							bodyStr = bodyStr[:1000] + "... (truncated)"
						}
						completionFields["response_body"] = bodyStr
					}
				}
			}

			// Log appropriate level based on status code
			completionLogger := loggerWithFields.WithFields(completionFields)
			if ctx.Response.StatusCode >= 500 {
				completionLogger.Errorf("Request completed with server error")
			} else if ctx.Response.StatusCode >= 400 {
				completionLogger.Warnf("Request completed with client error")
			} else {
				completionLogger.Infof("Request completed successfully")
			}

			return err
		}
	}
}

var _ registration.Module = (*RequestLogger)(nil)

func GetModule() registration.Module {
	return &RequestLogger{}
}
