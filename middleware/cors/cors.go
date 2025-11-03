package cors

import (
	"net/http"
	"slices"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

const CORS_TYPE = "cors"
const PARAMS_ALLOW_ORIGINS = "allow_origins"

// CORS middleware to handle CORS requests
// allowOrigins can be a list of allowed origins or ["*"] to allow all
func Middleware(allowOrigins ...string) request.HandlerFunc {
	AllOrigins := slices.Contains(allowOrigins, "*")
	return request.HandlerFunc(func(c *request.Context) error {
		origin := c.R.Header.Get("Origin")
		// only set CORS headers if Origin header is present
		if origin != "" {
			// if not allowing all origins, check if origin is in the allowed list
			if !AllOrigins && !slices.Contains(allowOrigins, origin) {
				c.W.WriteHeader(http.StatusForbidden)
				return nil
			}

			// Set CORS headers
			c.W.Header().Set("Access-Control-Allow-Origin", origin)
			c.W.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if c.R.Method == http.MethodOptions {
				if reqHeaders := c.R.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
					c.W.Header().Set("Access-Control-Allow-Headers", reqHeaders)
				}
				// Sets commonly used methods
				c.W.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				c.W.WriteHeader(http.StatusNoContent)
			}
		}
		return c.Next()
	})
}

func MiddlewareFactory(params map[string]any) request.HandlerFunc {
	if params == nil {
		return Middleware("*")
	}

	allowOrigins := utils.GetValueFromMap(params, PARAMS_ALLOW_ORIGINS, []string{})
	return Middleware(allowOrigins...)
}

func Register() {
	lokstra_registry.RegisterMiddlewareFactory(CORS_TYPE, MiddlewareFactory,
		lokstra_registry.AllowOverride(true))
}
