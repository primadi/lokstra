package cors

import (
	"net/http"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/iface"
)

const NAME = "cors"
const WHITELIST_KEY = "whitelist"

type CorsMiddleware struct{}

// Description implements registration.Module.
func (c *CorsMiddleware) Description() string {
	return "CORS middleware for handling cross-origin requests"
}

// Register implements registration.Module.
func (c *CorsMiddleware) Register(regCtx iface.RegistrationContext) error {
	return regCtx.RegisterMiddlewareFactoryWithPriority(NAME, factory, 30)
}

// Name implements registration.Module.
func (c *CorsMiddleware) Name() string {
	return NAME
}

func factory(config any) lokstra.MiddlewareFunc {
	whitelist := []string{"*"}
	var vParam any

	switch cfg := config.(type) {
	case map[string]any:
		if v, ok := cfg[WHITELIST_KEY]; ok {
			vParam = v
		}
	case []any:
		if len(cfg) > 0 {
			vParam = cfg[0]
		}
	}

	if rawList, ok := vParam.([]any); ok {
		var result []string
		for _, item := range rawList {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		if len(result) > 0 {
			whitelist = result
		}
	} else if list, ok := vParam.([]string); ok {
		whitelist = list
	}

	return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			origin := ctx.GetHeader("Origin")
			if origin != "" && matchOrigin(whitelist, origin) {
				ctx.WithHeader("Access-Control-Allow-Origin", origin).
					WithHeader("Access-Control-Allow-Credentials", "true")
			}

			if ctx.Request.Method == "OPTIONS" {
				if reqHeader := ctx.GetHeader("Access-Control-Request-Headers"); reqHeader != "" {
					ctx.WithHeader("Access-Control-Allow-Headers", reqHeader)
				}
				if reqMethod := ctx.GetHeader("Access-Control-Request-Method"); reqMethod != "" {
					ctx.WithHeader("Access-Control-Allow-Methods", reqMethod)
				}
				ctx.SetStatusCode(http.StatusNoContent)
				return nil
			}
			return next(ctx)
		}
	}
}

var _ lokstra.Module = (*CorsMiddleware)(nil)

func GetModule() lokstra.Module {
	return &CorsMiddleware{}
}
