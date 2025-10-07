package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
)

// This example demonstrates how to combine multiple routers into an App.
// An App is a container that runs multiple routers together on the same address.
//
// Key Concepts:
// 1. Create multiple routers for different domains (users, products, admin)
// 2. Create an App and pass all routers to it
// 3. Each router handles its own routes with its own path prefix
// 4. App coordinates request handling across all routers
//
// Run: go run main.go
// Test: curl http://localhost:8080/users
//       curl http://localhost:8080/products
//       curl http://localhost:8080/admin/stats

func main() {
	// Step 1: Create separate routers for different domains
	// Each router is independent and handles its own routes
	usersRouter := createUsersRouter()
	productsRouter := createProductsRouter()
	adminRouter := createAdminRouter()

	// Step 2: Print routes for each router
	fmt.Println("📋 Users Router Routes:")
	usersRouter.PrintRoutes()
	fmt.Println("\n📋 Products Router Routes:")
	productsRouter.PrintRoutes()
	fmt.Println("\n📋 Admin Router Routes:")
	adminRouter.PrintRoutes()

	// Step 3: Create an App with all routers
	// An App runs multiple routers on the same address
	// All routers are chained together and handle requests in order
	myApp := lokstra.NewApp("my-api-app", ":8080", usersRouter, productsRouter, adminRouter)

	// Step 4: Run the app with graceful shutdown
	fmt.Println("\n🚀 Server starting on http://localhost:8080")
	fmt.Println("Available endpoints:")
	fmt.Println("  Users:")
	fmt.Println("    - GET  http://localhost:8080/users")
	fmt.Println("    - GET  http://localhost:8080/users/123")
	fmt.Println("    - POST http://localhost:8080/users")
	fmt.Println("  Products:")
	fmt.Println("    - GET  http://localhost:8080/products")
	fmt.Println("    - GET  http://localhost:8080/products/456")
	fmt.Println("  Admin:")
	fmt.Println("    - GET  http://localhost:8080/admin/stats")
	fmt.Println("    - GET  http://localhost:8080/admin/health")
	fmt.Println("\nPress Ctrl+C to stop")

	// Run with 30 second graceful shutdown timeout
	myApp.Run(30 * time.Second)
}

func createUsersRouter() lokstra.Router {
	router := lokstra.NewRouter("users-router")

	// All routes in this router will match /users/*
	router.GET("/users", func(c *lokstra.RequestContext) error {
		return c.Api.Ok([]map[string]any{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
		})
	})

	router.GET("/users/:id", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "0")
		return c.Api.Ok(map[string]any{
			"id":   id,
			"name": "User " + id,
		})
	})

	router.POST("/users", func(c *lokstra.RequestContext) error {
		return c.Api.Created(map[string]any{
			"id":   3,
			"name": "New User",
		}, "User created")
	})

	return router
}

func createProductsRouter() lokstra.Router {
	router := lokstra.NewRouter("products-router")

	// All routes in this router will match /products/*
	router.GET("/products", func(c *lokstra.RequestContext) error {
		return c.Api.Ok([]map[string]any{
			{"id": 1, "name": "Laptop", "price": 1000},
			{"id": 2, "name": "Mouse", "price": 25},
		})
	})

	router.GET("/products/:id", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "0")
		return c.Api.Ok(map[string]any{
			"id":    id,
			"name":  "Product " + id,
			"price": 100,
		})
	})

	return router
}

func createAdminRouter() lokstra.Router {
	router := lokstra.NewRouter("admin-router")

	// All routes in this router will match /admin/*
	router.GET("/admin/stats", func(c *lokstra.RequestContext) error {
		return c.Api.Ok(map[string]any{
			"total_users":    150,
			"total_products": 500,
			"active_orders":  25,
		})
	})

	router.GET("/admin/health", func(c *lokstra.RequestContext) error {
		return c.Api.Ok(map[string]any{
			"status": "healthy",
			"uptime": "5h30m",
		})
	})

	return router
}
