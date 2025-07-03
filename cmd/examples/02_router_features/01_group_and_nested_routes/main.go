package main

import (
	"fmt"
	"lokstra"
)

func main() {
	// Register middleware used in route groups
	registerMiddlewares()

	// Create the root router
	router := lokstra.NewRouter()

	// Create a route group with prefix /api and apply auth middleware
	apiGroup := router.Group("/api", lokstra.NamedMiddleware("auth"))

	// Create a nested group under /api/user
	userGroup := apiGroup.Group("/user")

	// Register GET /api/user/profile
	userGroup.GET("/profile", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "User profile data",
		})
	})

	// Register POST /api/user/update
	userGroup.POST("/update", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "User updated successfully",
		})
	})

	// Create another group under /api/admin with extra middleware
	adminGroup := apiGroup.Group("/admin", lokstra.NamedMiddleware("admin_only"))

	// Register GET /api/admin/dashboard
	adminGroup.GET("/dashboard", func(ctx *lokstra.Context) error {
		return ctx.Ok("Welcome to the admin dashboard")
	})

	// Create an application and mount the router to it
	app := lokstra.NewApp("my-app", 8080)
	app.Mount(router)

	// Create and start the server
	server := lokstra.NewServer("server")
	server.AddApp(app)
	server.Start()
}

// registerMiddlewares registers the named middlewares used in this example.
func registerMiddlewares() {
	// Simulate an auth middleware
	lokstra.RegisterMiddlewareFunc("auth", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			fmt.Println("[Middleware] Auth check passed")
			return next(ctx)
		}
	})

	// Simulate an admin-only middleware
	lokstra.RegisterMiddlewareFunc("admin_only", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			fmt.Println("[Middleware] Admin check passed")
			return next(ctx)
		}
	})
}
