package recovery

import (
	"runtime/debug"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/iface"
)

const NAME = "recovery"

// Config holds the configuration for recovery middleware
type Config struct {
	EnableStackTrace bool `json:"enable_stack_trace" yaml:"enable_stack_trace"`
}

type RecoveryMiddleware struct{}

// Description implements registration.Module.
func (r *RecoveryMiddleware) Description() string {
	return "Recover from panic and return 500 error response. Should be the outermost middleware."
}

// Register implements registration.Module.
func (r *RecoveryMiddleware) Register(regCtx iface.RegistrationContext) error {
	regCtx.RegisterMiddlewareFactoryWithPriority(NAME, factory, 10)

	return nil
}

// Name implements registration.Module.
func (r *RecoveryMiddleware) Name() string {
	return NAME
}

func factory(config any) lokstra.MiddlewareFunc {
	// Parse configuration
	cfg := &Config{
		EnableStackTrace: true, // Default to true for backward compatibility
	}

	if config != nil {
		switch v := config.(type) {
		case map[string]any:
			if val, ok := v["enable_stack_trace"]; ok {
				if b, ok := val.(bool); ok {
					cfg.EnableStackTrace = b
				}
			}
		case *Config:
			cfg = v
		case Config:
			cfg = &v
		}
	}

	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			defer func() {
				if err := recover(); err != nil {
					_ = ctx.ErrorInternal("Internal Server Error")

					logFields := lokstra.LogFields{
						"error": err,
					}

					// Include stack trace only if enabled
					if cfg.EnableStackTrace {
						logFields["stack"] = string(debug.Stack())
					}

					lokstra.Logger.WithFields(logFields).
						Errorf("Recovered from panic in middleware")
				}
			}()

			return next(ctx)
		}
	}
}

var _ lokstra.Module = (*RecoveryMiddleware)(nil)

// return RecoveryMiddleware with name "lokstra.recovery"
func GetModule() lokstra.Module {
	return &RecoveryMiddleware{}
}
