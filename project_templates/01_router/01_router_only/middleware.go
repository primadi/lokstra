package main

import (
	"time"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/request"
)

// customLoggingMiddleware is an example of a custom middleware
// It logs basic request information including method, path, and processing time
func customLoggingMiddleware() request.HandlerFunc {
	return request.HandlerFunc(func(c *request.Context) error {
		// Record the start time
		startTime := time.Now()

		// Get request details
		method := c.R.Method
		path := c.R.URL.Path

		// Log the incoming request
		logger.LogInfo("[CUSTOM] Incoming request: %s %s", method, path)

		// Continue processing the request
		// Call Next() to pass the request to the next handler in the chain
		err := c.Next()

		// Calculate processing time
		duration := time.Since(startTime)

		// Log the completed request with timing
		logger.LogInfo("[CUSTOM] Completed request: %s %s - took %v", method, path, duration)

		return err
	})
}

// Example: Another custom middleware that could be used
// This demonstrates adding custom headers to all responses
func customHeaderMiddleware() request.HandlerFunc {
	return request.HandlerFunc(func(c *request.Context) error {
		// Add custom header before processing
		c.W.Header().Set("X-Powered-By", "Lokstra")
		c.W.Header().Set("X-API-Version", "1.0")

		// Continue to the next middleware/handler
		return c.Next()
	})
}
