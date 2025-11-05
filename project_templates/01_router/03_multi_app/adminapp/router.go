package adminapp

import (
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/slow_request_logger"
	"github.com/primadi/lokstra/project_templates/01_router/03_multi_app/mainapp"
	"github.com/primadi/lokstra/project_templates/01_router/03_multi_app/shared"
)

// SetupAdminAPIRouter configures the admin API router
func SetupAdminAPIRouter() lokstra.Router {
	r := lokstra.NewRouter("admin_api_router")

	// Apply middleware
	r.Use(recovery.Middleware(recovery.DefaultConfig()))
	r.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
		Threshold: 100 * time.Millisecond,
	}))
	r.Use(shared.CustomLoggingMiddleware("ADMIN"))

	// In production, add authentication middleware here
	// r.Use(adminAuthMiddleware())

	// Admin API routes
	setupAdminUsersRoutes(r)
	setupAdminRolesRoutes(r)
	setupAdminSystemRoutes(r)

	return r
}

// setupAdminUsersRoutes defines user management endpoints for admin
func setupAdminUsersRoutes(r lokstra.Router) {
	users := r.AddGroup("/admin/users")

	// Admin has full CRUD access
	users.GET("", HandleAdminGetAllUsers)
	users.GET("/:id", mainapp.HandleGetUser) // Reuse from mainapp
	users.POST("", mainapp.HandleCreateUser) // Reuse from mainapp
	users.PUT("/:id", mainapp.HandleUpdateUser)
	users.PATCH("/:id", mainapp.HandlePatchUser)
	users.DELETE("/:id", mainapp.HandleDeleteUser)

	// Admin-specific endpoints
	users.POST("/:id/suspend", HandleSuspendUser)
	users.POST("/:id/activate", HandleActivateUser)
}

// setupAdminRolesRoutes defines role management endpoints for admin
func setupAdminRolesRoutes(r lokstra.Router) {
	roles := r.AddGroup("/admin/roles")

	// Full CRUD for roles
	roles.GET("", mainapp.HandleGetRoles)
	roles.GET("/:id", mainapp.HandleGetRole)
	roles.POST("", mainapp.HandleCreateRole)
	roles.PUT("/:id", mainapp.HandleUpdateRole)
	roles.PATCH("/:id", mainapp.HandlePatchRole)
	roles.DELETE("/:id", mainapp.HandleDeleteRole)

	// Role assignment
	roles.POST("/:id/users/:userId", mainapp.HandleAssignRoleToUser)
	roles.DELETE("/:id/users/:userId", HandleRemoveRoleFromUser)
}

// setupAdminSystemRoutes defines system management endpoints
func setupAdminSystemRoutes(r lokstra.Router) {
	system := r.AddGroup("/admin/system")

	system.GET("/stats", HandleGetSystemStats)
	system.GET("/config", HandleGetSystemConfig)
	system.POST("/cache/clear", HandleClearCache)
}
