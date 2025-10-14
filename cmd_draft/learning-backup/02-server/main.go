package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
)

// This example demonstrates Servers that can run multiple Apps.
// A Server is the top-level concept that manages Apps and their listeners.
//
// Key Concepts:
// 1. Multiple Apps can run on different addresses
// 2. Apps with the same address are automatically merged
// 3. Server provides centralized management and graceful shutdown
//
// Run: go run main.go
// Test: curl http://localhost:8080/api/users
//       curl http://localhost:8081/admin/stats

func main() {
	// Step 1: Create routers for different domains
	usersRouter := createUsersRouter()
	productsRouter := createProductsRouter()
	adminRouter := createAdminRouter()

	// IMPORTANT: Router Reusability
	// A single router instance can be used in multiple apps!
	// Here, healthRouter is used in both publicApp and adminApp
	healthRouter := createHealthRouter()

	// Step 2: Create multiple Apps
	// Each App can listen on a different address

	// Public API - port 8080
	publicApp := lokstra.NewApp("public-api", ":8080", usersRouter, productsRouter, healthRouter)

	// Admin API - port 8081 (separate port for security)
	// Notice: We reuse the same healthRouter instance here!
	// This is a powerful pattern - one router definition, multiple apps
	adminApp := lokstra.NewApp("admin-api", ":8081", adminRouter, healthRouter)

	// Internal API - also port 8080 (will be merged with publicApp!)
	internalApp := lokstra.NewApp("internal-api", ":8080", createInternalRouter())

	// Step 3: Create a Server and add all Apps
	// Apps with the same address will be automatically merged
	server := lokstra.NewServer("my-server")
	server.AddApp(publicApp)
	server.AddApp(adminApp)
	server.AddApp(internalApp) // This will merge with publicApp (same :8080)

	// Step 4: Print server information
	fmt.Println("🖥️  Server Configuration:")
	fmt.Println("  Server Name: my-server")
	fmt.Println("  Total Apps: 3 (2 actual listeners due to merging)")
	fmt.Println("")
	fmt.Println("📡 Listeners:")
	fmt.Println("  Port 8080: public-api + internal-api (MERGED)")
	fmt.Println("    - Users Router")
	fmt.Println("    - Products Router")
	fmt.Println("    - Health Router")
	fmt.Println("    - Internal Router")
	fmt.Println("")
	fmt.Println("  Port 8081: admin-api")
	fmt.Println("    - Admin Router")
	fmt.Println("    - Health Router")
	fmt.Println("")

	// Step 5: Print all routes
	fmt.Println("📋 Public API Routes (port 8080):")
	usersRouter.PrintRoutes()
	productsRouter.PrintRoutes()
	healthRouter.PrintRoutes()
	createInternalRouter().PrintRoutes()

	fmt.Println("\n📋 Admin API Routes (port 8081):")
	adminRouter.PrintRoutes()
	healthRouter.PrintRoutes()

	// Step 6: Start the server
	fmt.Println("\n🚀 Server starting...")
	fmt.Println("Public API: http://localhost:8080")
	fmt.Println("  - GET http://localhost:8080/api/users")
	fmt.Println("  - GET http://localhost:8080/api/products")
	fmt.Println("  - GET http://localhost:8080/health")
	fmt.Println("  - GET http://localhost:8080/internal/config")
	fmt.Println("")
	fmt.Println("Admin API: http://localhost:8081")
	fmt.Println("  - GET http://localhost:8081/admin/stats")
	fmt.Println("  - GET http://localhost:8081/health")
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop")

	// Run server with 30 second graceful shutdown timeout
	server.Run(30 * time.Second)
}

func createUsersRouter() lokstra.Router {
	router := lokstra.NewRouter("users-router")
	router.GET("/api/users", func(c *lokstra.RequestContext) error {
		return c.Api.Ok([]map[string]any{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
		})
	})
	router.GET("/api/users/:id", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "0")
		return c.Api.Ok(map[string]any{"id": id, "name": "User " + id})
	})
	return router
}

func createProductsRouter() lokstra.Router {
	router := lokstra.NewRouter("products-router")
	router.GET("/api/products", func(c *lokstra.RequestContext) error {
		return c.Api.Ok([]map[string]any{
			{"id": 1, "name": "Laptop", "price": 1000},
			{"id": 2, "name": "Mouse", "price": 25},
		})
	})
	return router
}

func createAdminRouter() lokstra.Router {
	router := lokstra.NewRouter("admin-router")
	router.GET("/admin/stats", func(c *lokstra.RequestContext) error {
		return c.Api.Ok(map[string]any{
			"total_users":    150,
			"total_products": 500,
			"active_orders":  25,
		})
	})
	router.GET("/admin/users", func(c *lokstra.RequestContext) error {
		return c.Api.Ok([]map[string]any{
			{"id": 1, "name": "Alice", "role": "admin"},
			{"id": 2, "name": "Bob", "role": "user"},
		})
	})
	return router
}

func createHealthRouter() lokstra.Router {
	router := lokstra.NewRouter("health-router")
	router.GET("/health", func(c *lokstra.RequestContext) error {
		return c.Api.Ok(map[string]any{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})
	return router
}

func createInternalRouter() lokstra.Router {
	router := lokstra.NewRouter("internal-router")
	router.GET("/internal/config", func(c *lokstra.RequestContext) error {
		return c.Api.Ok(map[string]any{
			"app_name": "my-server",
			"version":  "1.0.0",
		})
	})
	return router
}
