package recovery

import (
	"fmt"
	"runtime/debug"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

const RECOVERY_TYPE = "recovery"
const PARAMS_ENABLE_STACK_TRACE = "enable_stack_trace"
const PARAMS_ENABLE_LOGGING = "enable_logging"

type Config struct {
	// EnableStackTrace includes stack trace in error response (for debugging)
	// Should be disabled in production
	EnableStackTrace bool

	// EnableLogging logs panic details to console
	EnableLogging bool

	// CustomHandler is a custom function to handle recovered panics
	// If nil, uses default error response
	CustomHandler func(c *request.Context, recovered any, stack []byte) error
}

func DefaultConfig() *Config {
	return &Config{
		EnableStackTrace: false, // Disabled by default for security
		EnableLogging:    true,
		CustomHandler:    nil,
	}
}

// middleware to recover from panics and return error response
func Middleware(cfg *Config) request.HandlerFunc {
	defConfig := DefaultConfig()
	if cfg == nil {
		cfg = defConfig
	} else {
		if cfg.CustomHandler == nil {
			cfg.CustomHandler = defConfig.CustomHandler
		}
	}

	return request.HandlerFunc(func(c *request.Context) error {
		defer func() {
			if r := recover(); r != nil {
				// Capture stack trace
				stack := debug.Stack()

				// Log panic if enabled
				if cfg.EnableLogging {
					logger.LogError("[PANIC RECOVERY] %v\n%s", r, stack)
				}

				// Use custom handler if provided
				if cfg.CustomHandler != nil {
					err := cfg.CustomHandler(c, r, stack)
					if err != nil {
						logger.LogError("[RECOVERY] Custom handler error: %v", err)
					}
					return
				}

				// Default error response - use InternalError which properly writes response
				message := fmt.Sprintf("Internal server error: %v", r)
				c.Api.InternalError(message)
			}
		}()

		// Continue to next handler
		return c.Next()
	})
}

func MiddlewareFactory(params map[string]any) request.HandlerFunc {
	defConfig := DefaultConfig()
	if params == nil {
		return Middleware(defConfig)
	}

	cfg := &Config{
		EnableStackTrace: utils.GetValueFromMap(params, PARAMS_ENABLE_STACK_TRACE, defConfig.EnableStackTrace),
		EnableLogging:    utils.GetValueFromMap(params, PARAMS_ENABLE_LOGGING, defConfig.EnableLogging),
		CustomHandler:    nil, // Cannot be set via params
	}
	return Middleware(cfg)
}

func Register() {
	lokstra_registry.RegisterMiddlewareFactory(RECOVERY_TYPE, MiddlewareFactory,
		lokstra_registry.AllowOverride(true))
}
