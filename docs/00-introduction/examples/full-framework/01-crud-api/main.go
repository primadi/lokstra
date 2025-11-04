package main

import (
	"log"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	log.Println("🚀 Starting Simple CRUD API Example...")

	// 1. Register service factories
	lokstra_registry.RegisterServiceType("database-factory", DatabaseFactory, nil)
	lokstra_registry.RegisterServiceType("user-service-factory", UserServiceFactory, nil)

	// 2. Register router factory
	lokstra_registry.RegisterRouter("api", createAPIRouter())

	// 3. Run server from config file
	lokstra_registry.RunServerFromConfig()
}

// ========================================
// Router Setup
// ========================================

func createAPIRouter() lokstra.Router {
	// Get lazy service - will be loaded on first HTTP request
	handler := NewUserHandler(lokstra_registry.GetLazyService[*UserService]("user-service"))

	// Create router
	r := lokstra.NewRouter("api")

	// CRUD routes - Using Group for clean organization
	r.Group("/api/v1/users", func(api lokstra.Router) {
		api.GET("/", handler.listUsers)
		api.GET("/{id}", handler.getUser)
		api.POST("/", handler.createUser)
		api.PUT("/{id}", handler.updateUser)
		api.DELETE("/{id}", handler.deleteUser)
	})

	// Info endpoint
	r.GET("/", func() map[string]any {
		return map[string]any{
			"service": "User CRUD API",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"GET /api/v1/users":         "List all users",
				"GET /api/v1/users/{id}":    "Get user by ID",
				"POST /api/v1/users":        "Create user",
				"PUT /api/v1/users/{id}":    "Update user",
				"DELETE /api/v1/users/{id}": "Delete user",
			},
		}
	})

	return r
}
