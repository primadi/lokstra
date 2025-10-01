package main

import (
	"fmt"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Simple middleware
func loggingMiddleware(c *request.Context) error {
	fmt.Printf("[LOG] %s %s\n", c.R.Method, c.R.URL.Path)
	return nil
}

func authMiddleware(c *request.Context) error {
	fmt.Printf("[AUTH] Checking authorization\n")
	return nil
}

func main() {
	fmt.Println("🔥 Minimal Config Demo - No Router Configuration Needed!")
	fmt.Println("========================================================")

	// Setup service factories (still needed)
	lokstra_registry.RegisterServiceFactory("database", func(config map[string]any) any {
		fmt.Printf("📊 Connected to database: %s\n", config["name"])
		return "db-connection"
	})

	// Setup router with middleware in code (no config needed)
	fmt.Println("\n1. Setting up router with middleware in code...")
	apiRouter := router.New("api")

	// Add middleware directly in code
	apiRouter.Use(loggingMiddleware)

	// Register routes with middleware
	apiRouter.GET("/health", func(c *request.Context) error {
		return c.Resp.WithStatus(200).Json(map[string]string{"status": "ok"})
	}, route.WithNameOption("health"))

	apiRouter.GET("/users", func(c *request.Context) error {
		return c.Resp.WithStatus(200).Json([]string{"user1", "user2"})
	}, authMiddleware, route.WithNameOption("users")) // Add auth middleware directly

	apiRouter.GET("/admin", func(c *request.Context) error {
		return c.Resp.WithStatus(200).Json(map[string]string{"admin": "panel"})
	}, authMiddleware, route.WithNameOption("admin"))

	lokstra_registry.RegisterRouter("api", apiRouter)
	fmt.Println("✅ Router registered with middleware in code")

	// Setup server
	fmt.Println("\n2. Setting up server...")
	app1 := app.New("web-app", ":8080")
	srv := server.New("web-server", app1)
	lokstra_registry.RegisterServer("web-server", srv)
	fmt.Println("✅ Server registered")

	// Minimal config - ONLY services and servers, NO routers!
	fmt.Println("\n3. Loading minimal configuration...")
	cfg := config.Config{
		Services: []config.Service{
			{
				Name: "main-db",
				Type: "database",
				Config: map[string]interface{}{
					"name": "my_app_db",
				},
			},
		},
		// NO routers configuration - using middleware from code!
		Servers: []config.Server{
			{
				Name:     "web-server",
				Services: []string{"main-db"},
				Apps: []config.App{
					{Name: "web-app", Addr: ":8080", Routers: []string{"api"}},
				},
			},
		},
	}

	// Apply minimal config
	if err := config.ApplyAllConfig(&cfg, "web-server"); err != nil {
		fmt.Printf("❌ Config failed: %v\n", err)
		return
	}

	fmt.Println("✅ Minimal config applied successfully")
	fmt.Println("\n🎉 Server ready!")
	fmt.Println("\n📋 Key Points:")
	fmt.Println("• Router middleware defined in code (no YAML needed)")
	fmt.Println("• Only services and servers in config")
	fmt.Println("• Much simpler configuration")
	fmt.Println("• All middleware behavior controlled by code")
}
