package cors

import (
	"net/http"
	"strings"

	"github.com/primadi/lokstra"
)

const NAME = "lokstra.cors"
const WHITELIST_KEY = "whitelist"

type CorsMiddleware struct{}

// Name implements iface.MiddlewareModule.
func (c *CorsMiddleware) Name() string {
	return NAME
}

// Meta implements iface.MiddlewareModule.
func (c *CorsMiddleware) Meta() *lokstra.MiddlewareMeta {
	return &lokstra.MiddlewareMeta{
		Priority:    30,
		Description: "CORS middleware for handling cross-origin requests",
		Tags:        []string{"cors", "middleware"},
	}
}

// Factory implements iface.MiddlewareModule.
func (c *CorsMiddleware) Factory(config any) lokstra.MiddlewareFunc {
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

var _ lokstra.MiddlewareModule = (*CorsMiddleware)(nil)

func matchOrigin(whitelist []string, origin string) bool {
	for _, allowed := range whitelist {
		allowed = strings.TrimSpace(allowed)
		if allowed == "*" {
			return true
		}
		if suffix, ok := strings.CutPrefix(allowed, "*"); ok {
			if strings.HasSuffix(origin, suffix) {
				return true
			}
		}
		if prefix, ok := strings.CutSuffix(allowed, "*"); ok {
			if strings.HasPrefix(origin, prefix) {
				return true
			}
		}
		if origin == allowed {
			return true
		}
	}

	return false
}

// return CorsMiddleware with name "lokstra.cors"
func GetModule() lokstra.MiddlewareModule {
	return &CorsMiddleware{}
}
