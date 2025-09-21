package midware

import (
	"net/http"

	"github.com/primadi/lokstra/core/request"
)

// WrapStdMiddleware Wrap Standard middleware (http-aware) to Lokstra middleware
func WrapStdMiddleware(mw func(http.Handler) http.Handler) Func {
	return func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// re-use the same ctx
				ctx.Writer = w
				ctx.Request = r
				_ = next(ctx)
			})
			mw(h).ServeHTTP(ctx.Writer, ctx.Request)
			return nil
		}
	}
}

// AsStdMiddleware converts Lokstra middleware to standard middleware (http-aware)
func AsStdMiddleware(mw Func) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, deferFunc := request.NewContext(nil, w, r)
			defer deferFunc()
			_ = mw(func(c *request.Context) error {
				next.ServeHTTP(c.Writer, c.Request)
				return nil
			})(ctx)
		})
	}
}
