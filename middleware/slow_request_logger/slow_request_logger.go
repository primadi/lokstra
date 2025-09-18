package slow_request_logger

import (
	"time"

	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

const NAME = "slow_request_logger"
const THRESHOLDKEY = "threshold"
const DEFAULT_THRESHOLD = 1200 * time.Millisecond // Default threshold for slow requests

type Config struct {
	IncludeRequestBody  bool          `json:"include_request_body" yaml:"include_request_body"`
	IncludeResponseBody bool          `json:"include_response_body" yaml:"include_response_body"`
	Threshold           time.Duration `json:"threshold" yaml:"threshold"`
}

type SlowRequestLogger struct{}

// Description implements registration.Module.
func (r *SlowRequestLogger) Description() string {
	return "Logs slow requests and their metadata."
}

var logger serviceapi.Logger

// Register implements registration.Module.
func (r *SlowRequestLogger) Register(regCtx registration.Context) error {
	if svc, err := regCtx.GetService("logger"); err == nil {
		logger = svc.(serviceapi.Logger)
	}
	return regCtx.RegisterMiddlewareFactoryWithPriority(NAME, factory, 20)
}

// Name implements registration.Module.
func (r *SlowRequestLogger) Name() string {
	return NAME
}

func factory(config any) midware.Func {
	cfg := &Config{
		IncludeRequestBody:  false,
		IncludeResponseBody: false,
		Threshold:           DEFAULT_THRESHOLD,
	}

	switch c := config.(type) {
	case map[string]any:
		if thany, ok := c[THRESHOLDKEY]; ok {
			if th, ok := thany.(string); ok {
				if dur, err := time.ParseDuration(th); err == nil {
					cfg.Threshold = dur
				}
			}
		}
		if val, ok := c["include_request_body"]; ok {
			if b, ok := val.(bool); ok {
				cfg.IncludeRequestBody = b
			}
		}
		if val, ok := c["include_response_body"]; ok {
			if b, ok := val.(bool); ok {
				cfg.IncludeResponseBody = b
			}
		}
	case string:
		if dur, err := time.ParseDuration(c); err == nil {
			cfg.Threshold = dur
		}
	case *Config:
		cfg = c
	case Config:
		cfg = &c
	}

	return func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			startTime := time.Now()
			defer func() {
				duration := time.Since(startTime)

				if duration >= cfg.Threshold && logger != nil {
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

					// Include response body if configured
					if cfg.IncludeResponseBody {
						responseBody, err := ctx.GetRawResponseBody()
						if err == nil && len(responseBody) > 0 {
							var jsonBody any
							if err := json.Unmarshal(responseBody, &jsonBody); err == nil {
								logFields["response_body"] = jsonBody
							} else {
								bodyStr := string(responseBody)
								if len(bodyStr) > 1000 {
									bodyStr = bodyStr[:1000] + "... (truncated)"
								}
								logFields["response_body"] = bodyStr
							}
						}
					}
					logFields["duration"] = duration.String()
					logFields["status"] = ctx.Response.StatusCode

					logger.WithFields(logFields).Infof("Slow request detected")
				}
			}()

			// Call the next handler in the chain
			return next(ctx)
		}
	}
}

var _ registration.Module = (*SlowRequestLogger)(nil)

// return SlowRequestLogger with name "lokstra.slow_request_logger"
func GetModule() registration.Module {
	return &SlowRequestLogger{}
}

// Preferred way to get slow request logger middleware execution
func GetMidware(cfg *Config) *midware.Execution {
	return &midware.Execution{
		Name:         NAME,
		Config:       cfg,
		MiddlewareFn: factory(cfg),
		Priority:     20,
	}
}
