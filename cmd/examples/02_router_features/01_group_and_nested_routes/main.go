package main

import (
	"fmt"
	"lokstra"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	// Create the root router
	app1 := lokstra.NewApp(ctx, "my-app", ":8080")

	// Create a route group with prefix /api and apply auth middleware
	apiGroup := app1.Group("/api", "auth")

	// Create a nested group under /api/user
	userGroup := apiGroup.Group("/user")

	// Register GET /api/user/profile
	userGroup.GET("/profile", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "User profile data",
		})
	})

	// Register GET /api/user/list
	userGroup.GET("/list", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "Get user list",
		})
	})

	// Create another group under /api/admin with extra middleware
	adminGroup := apiGroup.Group("/admin", "admin_only")

	// Register GET /api/admin/dashboard
	adminGroup.GET("/dashboard", "dashboard")

	registerMiddlewares(ctx)

	// Create and start the server
	server := lokstra.NewServer(ctx, "server")
	server.AddApp(app1)
	server.Start()
}

// registerMiddlewares registers the named middlewares used in this example.
func registerMiddlewares(ctx lokstra.ComponentContext) {
	// Simulate an auth middleware
	ctx.RegisterMiddlewareFunc("auth", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			fmt.Println("[Middleware] Auth check passed")
			return next(ctx)
		}
	})

	// Simulate an admin-only middleware
	ctx.RegisterMiddlewareFunc("admin_only", func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
		return func(ctx *lokstra.Context) error {
			fmt.Println("[Middleware] Admin check passed")
			return next(ctx)
		}
	})

	ctx.RegisterHandler("dashboard", func(ctx *lokstra.Context) error {
		return ctx.Ok("Welcome to the admin dashboard")
	})
}
