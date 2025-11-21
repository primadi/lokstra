package main

import (
	"fmt"
	"log"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/lokstra_registry"
)

// NEW RECOMMENDED FLOW
// This flow separates config loading from service registration,
// allowing services to access config during registration.
func main() {
	lokstra.Bootstrap()

	fmt.Println("")
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║   LOKSTRA ENTERPRISE MODULAR TEMPLATE         ║")
	fmt.Println("║   Domain-Driven Design with Bounded Contexts  ║")
	fmt.Println("║   [Config First]                              ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Println("")

	deploy.SetLogLevelFromEnv()

	// ===== STEP 1: Load Config =====
	// Config is loaded first, making it available for service/middleware registration
	// This registers lazy load services and deployment structure from YAML
	if err := lokstra_registry.LoadConfigFromFolder("config"); err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	// ===== STEP 2: Register Service Types =====
	// At this point, config is already loaded and available
	// Service factories can now access config via lokstra_registry.GetConfig()
	registerServiceTypes()

	// ===== STEP 3: Register Middleware Types =====
	// Middleware factories can also access config if needed
	registerMiddlewareTypes()

	// ===== STEP 4: Initialize and Run Server =====
	// This will:
	// - Select server based on config (or auto-select first server)
	// - Read shutdown timeout from config
	// - Start the server
	if err := lokstra_registry.InitAndRunServer(); err != nil {
		log.Fatal("❌ Failed to run server:", err)
	}
}
