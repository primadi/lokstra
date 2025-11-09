package shared

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra/core/request"
)

// CustomLoggingMiddleware logs requests with app identification
func CustomLoggingMiddleware(appName string) func(*request.Context) error {
	return func(c *request.Context) error {
		start := time.Now()

		// Process request
		err := c.Next()

		duration := time.Since(start)
		if err != nil {
			fmt.Printf("ERROR [%s] %s %s - %d - %v: %v\n",
				appName,
				c.R.Method,
				c.R.URL.Path,
				c.Resp.RespStatusCode,
				duration,
				err,
			)
		} else {
			// Log successful requests
			fmt.Printf("[%s] %s %s - %d - %v\n",
				appName,
				c.R.Method,
				c.R.URL.Path,
				c.Resp.RespStatusCode,
				duration,
			)
		}
		return err
	}
}
