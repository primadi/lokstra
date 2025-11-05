package main

import (
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/slow_request_logger"
)

// setupRouter creates and configures the Lokstra router
func setupRouter() lokstra.Router {
	// Create a new router instance
	r := lokstra.NewRouter("demo_router")

	// Apply global middleware
	// Recovery middleware catches panics and returns proper error responses
	r.Use(recovery.Middleware(recovery.DefaultConfig()))

	// Slow request logger helps identify performance bottlenecks
	// Logs requests that take longer than 100ms
	r.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
		Threshold: 100 * time.Millisecond,
	}))

	// Apply custom middleware as an example
	r.Use(customLoggingMiddleware(), customHeaderMiddleware())

	// Define route groups
	setupUsersRoutes(r)
	setupRolesRoutes(r)

	return r
}

// setupUsersRoutes defines all user-related endpoints
func setupUsersRoutes(r lokstra.Router) {
	// Create a route group for user operations
	users := r.AddGroup("/users")

	// List all users
	users.GET("", handleGetUsers)

	// Get a specific user by ID
	users.GET("/:id", handleGetUser)

	// Create a new user
	users.POST("", handleCreateUser)

	// Update an existing user (full update)
	users.PUT("/:id", handleUpdateUser)

	// Partially update a user
	users.PATCH("/:id", handlePatchUser)

	// Delete a user
	users.DELETE("/:id", handleDeleteUser)
}

// setupRolesRoutes defines all role-related endpoints
func setupRolesRoutes(r lokstra.Router) {
	// Create a route group for role operations
	roles := r.AddGroup("/roles")

	// List all roles
	roles.GET("", handleGetRoles)

	// Get a specific role by ID
	roles.GET("/:id", handleGetRole)

	// Create a new role
	roles.POST("", handleCreateRole)

	// Update an existing role (full update)
	roles.PUT("/:id", handleUpdateRole)

	// Partially update a role
	roles.PATCH("/:id", handlePatchRole)

	// Delete a role
	roles.DELETE("/:id", handleDeleteRole)

	// Assign a role to a user (nested resource example)
	roles.POST("/:id/users/:userId", handleAssignRoleToUser)
}
