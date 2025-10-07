package main

import (
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/lokstra_registry"
)

// setupRouters registers all routers with lokstra_registry
// Routers are defined in CODE, not in YAML config
// The YAML config only references them by name in servers.apps.routers
func setupRouters() {
	// Create API router
	apiRouter := router.New("api-router")

	// Register routes with handlers

	// Health check endpoint - no middleware
	apiRouter.GET("/health", healthCheckHandler)

	// User endpoints - with named middleware from config
	// Middleware "cors", "request_logger", "recovery" are defined in config.yaml
	// They can be applied globally via router.Use() or per-route like this:
	apiRouter.GET("/users", listUsersHandler, "cors", "request_logger")
	apiRouter.POST("/users", createUserHandler, "cors", "request_logger")
	apiRouter.GET("/users/:id", getUserHandler, "cors", "request_logger")

	// Example: Using route options
	apiRouter.GET("/admin/stats", adminStatsHandler,
		"cors",
		"request_logger",
		route.WithDescriptionOption("Get admin statistics"),
	)

	// Register router with lokstra_registry
	// This makes it available to be referenced in config.yaml servers.apps.routers
	lokstra_registry.RegisterRouter("api-router", apiRouter)
}
