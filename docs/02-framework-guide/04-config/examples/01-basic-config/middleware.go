package main

import (
	"log"

	"github.com/primadi/lokstra/core/request"
)

// RequestLoggerMiddleware logs incoming requests
func RequestLoggerMiddleware(config map[string]any) request.HandlerFunc {
	return func(ctx *request.Context) error {
		// Before request
		log.Printf("→ %s %s", ctx.R.Method, ctx.R.URL.Path)

		// Process request
		err := ctx.Next()

		// After request
		if err != nil {
			log.Printf("← ERROR: %v", err)
		} else {
			log.Printf("← %d", ctx.Resp.RespStatusCode)
		}

		return err
	}
}
