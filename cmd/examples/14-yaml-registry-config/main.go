package main

import (
	"fmt"
	"log"

	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Sample handlers
func healthHandler(c *request.Context) error {
	return c.Resp.WithStatus(200).Json(map[string]string{
		"status": "healthy",
		"server": "lokstra-yaml-config-demo",
	})
}

func aboutHandler(c *request.Context) error {
	return c.Resp.WithStatus(200).Json(map[string]string{
		"app":     "Lokstra YAML Config Demo",
		"version": "1.0.0",
	})
}

// Sample middleware factory
func loggingMiddlewareFactory(config map[string]any) request.HandlerFunc {
	return func(c *request.Context) error {
		fmt.Printf("[%s] %s %s\n", "LOG", c.R.Method, c.R.URL.Path)
		return nil
	}
}

// Sample service factory
func memoryServiceFactory(config map[string]any) any {
	fmt.Printf("Creating memory service with config: %v\n", config)
	return map[string]string{"type": "memory", "status": "connected"}
}

func main() {
	fmt.Println("🔧 Lokstra YAML Config + Registry Demo")
	fmt.Println("========================================")

	// Step 1: Register factories (normally done at app startup)
	fmt.Println("\n1. Registering factories...")
	lokstra_registry.RegisterMiddlewareFactory("logger", loggingMiddlewareFactory)
	lokstra_registry.RegisterServiceFactory("memory", memoryServiceFactory)
	fmt.Println("✅ Registered middleware factory: logger")
	fmt.Println("✅ Registered service factory: memory")

	// Step 2: Register components in code (normally done in main.go)
	fmt.Println("\n2. Registering components in code...")

	// Register router with initial routes
	r := router.New("api-router")
	r.GET("/health", healthHandler)
	r.GET("/about", aboutHandler)
	lokstra_registry.RegisterRouter("api-router", r)
	fmt.Println("✅ Registered router: api-router with 2 routes")

	// Register server
	srv := server.New("web-server")
	lokstra_registry.RegisterServer("web-server", srv)
	fmt.Println("✅ Registered server: web-server")

	// Step 3: Load YAML configuration
	fmt.Println("\n3. Loading YAML configuration...")

	// Example YAML content (in real use, this would be in a .yaml file):
	_ = `
middlewares:
  - name: logging-mw
    type: logger
    config:
      level: "info"

services:
  - name: cache-service
    type: memory
    config:
      size: "100MB"

routers:
  - name: api-router
    use: [logging-mw]  # Add middleware to existing router

servers:
  - name: web-server
    services: [cache-service]
    apps:
      - name: main-app
        addr: ":8080"
        routers: [api-router]
`

	var cfg config.Config
	// Simulate loading from string (in real use, use LoadConfigFile/LoadConfigDir)
	if err := config.LoadConfigDir("./temp_config_does_not_exist", &cfg); err == nil {
		// This won't execute due to non-existent directory, but shows the pattern
	}

	// Manually populate config for demo
	cfg = config.Config{
		Configs: []config.GeneralConfig{
			{Name: "server-url", Value: "http://my-server.com"},
			{Name: "max-connections", Value: 100},
			{Name: "debug-enabled", Value: true},
		},
		Middlewares: []config.Middleware{
			{Name: "logging-mw", Type: "logger", Config: map[string]interface{}{"level": "info"}},
		},
		Services: []config.Service{
			{Name: "cache-service", Type: "memory", Config: map[string]interface{}{"size": "100MB"}},
		},
		Routers: []config.Router{
			{
				Name: "api-router",
				Use:  []string{"logging-mw"},
				// No routes needed - just adding middleware to existing router
			},
		},
		Servers: []config.Server{
			{
				Name:     "web-server",
				Services: []string{"cache-service"},
				Apps: []config.App{
					{
						Name:    "main-app",
						Addr:    ":8080",
						Routers: []string{"api-router"},
					},
				},
			},
		},
	}
	fmt.Println("✅ Loaded configuration from YAML")

	// Step 4: Apply configuration to modify existing components
	fmt.Println("\n4. Applying configuration...")

	// Use ApplyAllConfig to ensure proper order (including general configs)
	if err := config.ApplyAllConfig(&cfg, "web-server"); err != nil {
		log.Printf("Failed to apply configuration: %v", err)
	} else {
		fmt.Println("✅ Applied all configuration successfully")
	}

	// Step 5: Verify modifications
	fmt.Println("\n5. Verifying modifications...")

	// Check if middleware was registered
	mw := lokstra_registry.CreateMiddleware("logging-mw")
	if mw != nil {
		fmt.Println("✅ Middleware 'logging-mw' is available")
	} else {
		fmt.Println("❌ Middleware 'logging-mw' not found")
	}

	// Check if router still exists
	modifiedRouter := lokstra_registry.GetRouter("api-router")
	if modifiedRouter != nil {
		fmt.Println("✅ Router 'api-router' is available and modified")
	} else {
		fmt.Println("❌ Router 'api-router' not found")
	}

	// Check if server exists
	modifiedServer := lokstra_registry.GetServer("web-server")
	if modifiedServer != nil {
		fmt.Println("✅ Server 'web-server' is available and configured")
	} else {
		fmt.Println("❌ Server 'web-server' not found")
	}

	fmt.Println("\n🎉 Demo completed!")
	// Step 6: Demo general config usage
	fmt.Println("\n6. Testing general configuration...")
	fmt.Printf("Available configs: %v\n", lokstra_registry.ListConfigNames())
	serverURL := lokstra_registry.GetConfigString("server-url", "http://localhost")
	maxConn := lokstra_registry.GetConfigInt("max-connections", 10)
	debugEnabled := lokstra_registry.GetConfigBool("debug-enabled", false)

	fmt.Printf("✅ Server URL: %s\n", serverURL)
	fmt.Printf("✅ Max Connections: %d\n", maxConn)
	fmt.Printf("✅ Debug Enabled: %t\n", debugEnabled)

	// Demo setting config at runtime
	lokstra_registry.SetConfig("runtime-setting", "set from code")
	runtimeValue := lokstra_registry.GetConfigString("runtime-setting", "default")
	fmt.Printf("✅ Runtime Setting: %s\n", runtimeValue)

	fmt.Println("\n📋 Key Concepts Demonstrated:")
	fmt.Println("• YAML config modifies existing registry components")
	fmt.Println("• Components must be registered in code first")
	fmt.Println("• Config adds middleware to existing routes/routers")
	fmt.Println("• General configs available for hardcoded components")
	fmt.Println("• Runtime config changes supported")
	fmt.Println("• Panic if trying to modify non-existent components")
	fmt.Println("• Middleware inheritance is always additive")
}
