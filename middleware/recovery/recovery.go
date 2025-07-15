package recovery

import (
	"lokstra"
	"runtime/debug"
)

const NAME = "lokstra.recovery"

type RecoveryMiddleware struct{}

// Name implements iface.MiddlewareModule.
func (r *RecoveryMiddleware) Name() string {
	return NAME
}

// Meta implements iface.MiddlewareModule.
func (r *RecoveryMiddleware) Meta() *lokstra.MiddlewareMeta {
	return &lokstra.MiddlewareMeta{
		Priority:    10,
		Description: "Recover from panic and return 500 error response. Should be the outermost middleware.",
		Tags:        []string{"recovery", "safety"},
	}
}

// Factory implements iface.MiddlewareModule.
func (r *RecoveryMiddleware) Factory(_ any) lokstra.MiddlewareFunc {
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

var _ lokstra.MiddlewareModule = (*RecoveryMiddleware)(nil)

// return RecoveryMiddleware with name "lokstra.recovery"
func GetModule() lokstra.MiddlewareModule {
	return &RecoveryMiddleware{}
}
