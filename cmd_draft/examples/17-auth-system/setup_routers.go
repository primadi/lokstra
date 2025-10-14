package main

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/accesscontrol"
	"github.com/primadi/lokstra/middleware/jwtauth"
)

// setupRouters creates and registers all routers for the auth system demo
func setupRouters() {
	setupAuthRouter()
	setupUserRouter()
	setupAdminRouter()
	setupPublicRouter()
}

// setupAuthRouter creates the authentication router
// Handles login, register, refresh token, logout
func setupAuthRouter() {
	router := lokstra.NewRouter("auth-api")

	// JWT middleware instance for protected routes
	jwtMiddleware := jwtauth.MiddlewareFactory(map[string]any{
		"validator_service_name": "auth_validator",
	})

	// Public routes - no authentication required
	router.POST("/auth/register", registerHandler, route.WithNameOption("auth-register"))
	router.POST("/auth/login", loginHandler, route.WithNameOption("auth-login"))
	router.POST("/auth/refresh", refreshTokenHandler, route.WithNameOption("auth-refresh"))

	// Protected routes - requires authentication (middleware applied at handler level)
	router.POST("/auth/logout", logoutHandler, route.WithNameOption("auth-logout"), jwtMiddleware)
	router.GET("/auth/me", getCurrentUserHandler, route.WithNameOption("auth-me"), jwtMiddleware)

	// OTP endpoints
	router.POST("/auth/otp/generate", generateOTPHandler, route.WithNameOption("auth-otp-generate"))
	router.POST("/auth/otp/verify", verifyOTPHandler, route.WithNameOption("auth-otp-verify"))

	lokstra_registry.RegisterRouter("auth-api", router)
}

// setupUserRouter creates the user router
// Handles user profile and user-specific operations
func setupUserRouter() {
	router := lokstra.NewRouter("user-api")

	// Apply JWT middleware at router level - all routes require authentication
	jwtMiddleware := jwtauth.MiddlewareFactory(map[string]any{
		"validator_service_name": "auth_validator",
	})
	router.Use(jwtMiddleware)

	router.GET("/users/profile", getUserProfileHandler, route.WithNameOption("user-profile"))
	router.PUT("/users/profile", updateUserProfileHandler, route.WithNameOption("user-profile-update"))
	router.POST("/users/change-password", changePasswordHandler, route.WithNameOption("user-change-password"))

	// User's own resources
	router.GET("/users/orders", getUserOrdersHandler, route.WithNameOption("user-orders"))
	router.POST("/users/orders", createUserOrderHandler, route.WithNameOption("user-order-create"))

	lokstra_registry.RegisterRouter("user-api", router)
}

// setupAdminRouter creates the admin router
// Handles admin operations - requires admin role
func setupAdminRouter() {
	router := lokstra.NewRouter("admin-api")

	// Apply JWT middleware at router level
	jwtMiddleware := jwtauth.MiddlewareFactory(map[string]any{
		"validator_service_name": "auth_validator",
	})
	router.Use(jwtMiddleware)

	// All routes require admin role
	adminMiddleware := accesscontrol.RequireAdmin()

	router.GET("/admin/users", listAllUsersHandler, route.WithNameOption("admin-users-list"), adminMiddleware)
	router.GET("/admin/users/{id}", getUserByIDHandler, route.WithNameOption("admin-user-detail"), adminMiddleware)
	router.PUT("/admin/users/{id}/activate", activateUserHandler, route.WithNameOption("admin-user-activate"), adminMiddleware)
	router.PUT("/admin/users/{id}/deactivate", deactivateUserHandler, route.WithNameOption("admin-user-deactivate"), adminMiddleware)
	router.DELETE("/admin/users/{id}", deleteUserHandler, route.WithNameOption("admin-user-delete"), adminMiddleware)

	// System stats - requires admin or manager
	router.GET("/admin/stats", getSystemStatsHandler, route.WithNameOption("admin-stats"), accesscontrol.RequireAdminOrManager())

	lokstra_registry.RegisterRouter("admin-api", router)
}

// setupPublicRouter creates the public router
// Handles public endpoints that don't require authentication
func setupPublicRouter() {
	router := lokstra.NewRouter("public-api")

	router.GET("/health", healthCheckHandler, route.WithNameOption("health-check"))
	router.GET("/info", getSystemInfoHandler, route.WithNameOption("system-info"))

	lokstra_registry.RegisterRouter("public-api", router)
}
