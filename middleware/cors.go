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
		if origin != "" {
			if !AllOrigins && !slices.Contains(allowOrigins, origin) {
				c.W.WriteHeader(http.StatusForbidden)
				return nil
			}
			c.W.Header().Set("Access-Control-Allow-Origin", origin)
			c.W.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if c.R.Method == http.MethodOptions {
			if reqHeaders := c.R.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
				c.W.Header().Set("Access-Control-Allow-Headers", reqHeaders)
			}
			c.W.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.W.WriteHeader(http.StatusNoContent)
		}
		return c.Next()
	})
}
