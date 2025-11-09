package mainapp

import (
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/slow_request_logger"
	"github.com/primadi/lokstra/project_templates/01_router/03_multi_app/shared"
)

// SetupMainAPIRouter configures the main API router
func SetupMainAPIRouter() lokstra.Router {
	r := lokstra.NewRouter("main_api_router")

	// Apply middleware
	r.Use(recovery.Middleware(recovery.DefaultConfig()))
	r.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
		Threshold: 100 * time.Millisecond,
	}))
	r.Use(shared.CustomLoggingMiddleware("MAIN"))

	// Public API routes
	setupUsersRoutes(r)
	setupRolesRoutes(r)

	return r
}

// setupUsersRoutes defines user endpoints for main API
func setupUsersRoutes(r lokstra.Router) {
	users := r.AddGroup("/api/users")
	users.GET("", HandleGetUsers)
	users.GET("/:id", HandleGetUser)
	users.POST("", HandleCreateUser)
	users.PUT("/:id", HandleUpdateUser)
	users.PATCH("/:id", HandlePatchUser)
	users.DELETE("/:id", HandleDeleteUser)
}

// setupRolesRoutes defines role endpoints for main API
func setupRolesRoutes(r lokstra.Router) {
	roles := r.AddGroup("/api/roles")
	roles.GET("", HandleGetRoles)
	roles.GET("/:id", HandleGetRole)
	roles.POST("", HandleCreateRole)
	roles.PUT("/:id", HandleUpdateRole)
	roles.PATCH("/:id", HandlePatchRole)
	roles.DELETE("/:id", HandleDeleteRole)
	roles.POST("/:id/users/:userId", HandleAssignRoleToUser)
}
