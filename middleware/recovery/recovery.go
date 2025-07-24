package recovery

import (
	"runtime/debug"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/registration"
)

const NAME = "recovery"

type RecoveryMiddleware struct{}

// Description implements registration.Module.
func (r *RecoveryMiddleware) Description() string {
	return "Recover from panic and return 500 error response. Should be the outermost middleware."
}

// Register implements registration.Module.
func (r *RecoveryMiddleware) Register(regCtx registration.Context) error {
	regCtx.RegisterMiddlewareFactoryWithPriority(NAME, factory, 10)

	return nil
}

// Name implements registration.Module.
func (r *RecoveryMiddleware) Name() string {
	return NAME
}

func factory(_ any) lokstra.MiddlewareFunc {
	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			defer func() {
				if err := recover(); err != nil {
					_ = ctx.ErrorInternal("Internal Server Error")
					lokstra.Logger.WithField("error", err).
						WithField("stack", string(debug.Stack())).
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
