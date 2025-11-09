package main

import (
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/slow_request_logger"
)

// setupAPIRouter creates and configures the main API router
func setupAPIRouter() lokstra.Router {
	// Create a new router instance for API
	r := lokstra.NewRouter("api_router")

	// Apply global middleware
	r.Use(customHeaderMiddleware())
	r.Use(recovery.Middleware(recovery.DefaultConfig()))
	r.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
		Threshold: 100 * time.Millisecond,
	}))
	r.Use(customLoggingMiddleware())

	// Define route groups
	setupUsersRoutes(r)
	setupRolesRoutes(r)

	return r
}

// setupHealthRouter creates a dedicated health check router
// This demonstrates that one App can have multiple routers
func setupHealthRouter() lokstra.Router {
	// Create a separate router for health checks
	r := lokstra.NewRouter("health_router")

	// Health check endpoint - no middleware needed for performance
	r.GET("/health", handleHealth)
	r.GET("/ready", handleReady)

	return r
}

// setupUsersRoutes defines all user-related endpoints
func setupUsersRoutes(r lokstra.Router) {
	users := r.AddGroup("/users")
	users.GET("", handleGetUsers)
	users.GET("/:id", handleGetUser)
	users.POST("", handleCreateUser)
	users.PUT("/:id", handleUpdateUser)
	users.PATCH("/:id", handlePatchUser)
	users.DELETE("/:id", handleDeleteUser)
}

// setupRolesRoutes defines all role-related endpoints
func setupRolesRoutes(r lokstra.Router) {
	roles := r.AddGroup("/roles")
	roles.GET("", handleGetRoles)
	roles.GET("/:id", handleGetRole)
	roles.POST("", handleCreateRole)
	roles.PUT("/:id", handleUpdateRole)
	roles.PATCH("/:id", handlePatchRole)
	roles.DELETE("/:id", handleDeleteRole)
	roles.POST("/:id/users/:userId", handleAssignRoleToUser)
}
