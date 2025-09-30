package middleware

import (
	"net/http"
	"slices"

	"github.com/primadi/lokstra/core/request"
)

func Cors(allowOrigins []string) request.HandlerFunc {
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
