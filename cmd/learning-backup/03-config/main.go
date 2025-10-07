package main

import (
	"fmt"

	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/cors"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
)

// This example demonstrates YAML configuration with Lokstra.
//
// Key concepts:
// 1. Register routers in code (not in YAML)
// 2. Register middleware factories in code
// 3. Load config from YAML (configs, services, middlewares, servers)
// 4. lokstra_registry wires everything automatically
//
// YAML Config Structure:
//   configs     → General key-value (lokstra_registry.GetConfig)
//   services    → Lazy services (lokstra_registry.GetService)
//   middlewares → Named middleware (router.Use("name") or in route options)
//   servers     → Server definitions with apps and routers
//
// Run: go run .

func main() {
	fmt.Println("🎯 Lokstra YAML Configuration Demo")
	fmt.Println("===================================")

	// ========================================
	// STEP 1: Register Middleware Factories
	// ========================================
	// Middleware factories must be registered BEFORE loading config
	// The YAML config references these by 'type' name

	lokstra_registry.RegisterMiddlewareFactory("cors", cors.MiddlewareFactory)
	lokstra_registry.RegisterMiddlewareFactory("request_logger", request_logger.MiddlewareFactory)
	lokstra_registry.RegisterMiddlewareFactory("recovery", recovery.MiddlewareFactory)

	fmt.Println("✅ Middleware factories registered: cors, request_logger, recovery")

	// ========================================
	// STEP 2: Register Routers in Code
	// ========================================
	// Routers are NOT defined in YAML, they're registered in code
	// Then referenced by name in servers.apps.routers

	setupRouters()

	fmt.Println("✅ Routers registered: api-router")

	// ========================================
	// STEP 3: Load Configuration from YAML
	// ========================================
	fmt.Println("\n📖 Loading configuration from config.yaml...")

	cfg := config.New()
	if err := config.LoadConfigFile("config.yaml", cfg); err != nil {
		panic(fmt.Sprintf("❌ Failed to load config: %v", err))
	}

	fmt.Printf("✅ Configuration loaded\n")
	fmt.Printf("   Configs: %d\n", len(cfg.Configs))
	fmt.Printf("   Services: %d\n", len(cfg.Services))
	fmt.Printf("   Middlewares: %d\n", len(cfg.Middlewares))
	fmt.Printf("   Servers: %d\n\n", len(cfg.Servers))

	// ========================================
	// STEP 4: Register Config with Registry
	// ========================================
	// This processes the config and registers:
	// - General configs (accessible via GetConfig)
	// - Services (lazy-loaded via GetService)
	// - Middlewares (accessible via CreateMiddleware or router.Use)
	// - Servers (with apps and routers)

	lokstra_registry.RegisterConfig(cfg)

	fmt.Println("✅ Configuration registered with lokstra_registry")

	// ========================================
	// STEP 5: Access General Configs
	// ========================================
	// General configs from 'configs:' section can be accessed

	appName := lokstra_registry.GetConfig("app-name", "unknown")
	appVersion := lokstra_registry.GetConfig("app-version", "0.0.0")
	environment := lokstra_registry.GetConfig("environment", "production")

	fmt.Println("\n� General Configs:")
	fmt.Printf("   App Name: %s\n", appName)
	fmt.Printf("   Version: %s\n", appVersion)
	fmt.Printf("   Environment: %s\n", environment)

	// ========================================
	// STEP 6: Set Current Server and Start
	// ========================================
	// Only ONE server runs at a time
	// For this example, we use "dev-server" from config.yaml

	serverName := "dev-server"
	lokstra_registry.SetCurrentServerName(serverName)

	fmt.Println("\n🚀 Starting Server...")
	fmt.Println("===================================")

	// Print server start info (endpoints, middleware, etc.)
	lokstra_registry.PrintServerStartInfo()

	// Start the server (blocks until shutdown)
	lokstra_registry.StartServer()
}
