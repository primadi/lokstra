package main

import (
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/old_registry"
)

func main() {
	// Register services
	// old_registry.RegisterServiceFactory("db", func() any {
	// 	return NewDatabase()
	// })
	// register service factory: dbFactory
	old_registry.RegisterServiceType("dbFactory", NewDatabase)
	// regiuster lazy service: db using dbFactory
	old_registry.RegisterLazyService("db", "dbFactory", nil)

	// old_registry.RegisterServiceFactory("users", func() any {
	// 	return &UserService{
	// 		DB: service.LazyLoad[*Database]("db"),
	// 	}
	// })
	// register service factory: usersFactory
	old_registry.RegisterServiceType("usersFactory", func() any {
		return &UserService{
			DB: service.LazyLoad[*Database]("db"),
		}
	})
	// register lazy service: users using usersFactory
	old_registry.RegisterLazyService("users", "usersFactory", nil)

	// Create router
	r := lokstra.NewRouter("api")

	// Route 1: Manual routes with custom handlers
	r.Group("/api/v1/users", func(api lokstra.Router) {
		api.GET("/", listUsersHandler)
		api.GET("/{id}", getUserHandler)
		api.POST("/", createUserHandler)
		api.PUT("/{id}", updateUserHandler)
		api.DELETE("/{id}", deleteUserHandler)
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

	// Create app
	app := lokstra.NewApp("crud-api", ":3002", r)

	app.PrintStartInfo()
	app.Run(30 * time.Second)
}
