package main

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
)

// NewHealthRouter creates a manual health check router
// This demonstrates how to create routers manually without @EndpointService annotation
func NewHealthRouter() router.Router {
	r := router.New("health-router")

	// Simple health check endpoint
	r.GET("/health", func(ctx *request.Context) error {
		return ctx.Api.Ok(map[string]any{
			"status":  "healthy",
			"service": "lokstra-enterprise-template",
		})
	})

	// Readiness check endpoint
	r.GET("/ready", func(ctx *request.Context) error {
		return ctx.Api.Ok(map[string]any{
			"status": "ready",
		})
	})

	return r
}
