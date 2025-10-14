package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra/cmd_draft/examples/25-single-binary-deployment/services"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Parse command-line flags
	configFile := flag.String("config", "config-monolith-v2.yaml", "Path to configuration file")
	serverName := flag.String("server", "", "Server name to run (empty = first server in config)")
	flag.Parse()

	log.Printf("🚀 Starting with config: %s", *configFile)
	if *serverName != "" {
		log.Printf("📍 Server filter: %s", *serverName)
	}

	// ==============================================================================
	// STEP 1: Register Service Factories (LOCAL + REMOTE)
	// ==============================================================================
	log.Println("📋 Step 1: Registering service factories...")
	services.RegisterAuthService()
	services.RegisterCartService()
	services.RegisterInvoiceService()
	services.RegisterOrderService()
	services.RegisterPaymentService()
	services.RegisterUserService()
	log.Println("   ✅ All service factories registered (local + remote)")

	// ==============================================================================
	// STEP 2: Load Configuration from YAML (routers auto-generated from services)
	// ==============================================================================
	log.Println("📋 Step 2: Loading configuration from YAML...")

	cfg := config.New()
	if err := config.LoadConfigFile(*configFile, cfg); err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	log.Println("   ✅ Configuration loaded successfully")
	log.Println("   ✅ Routers will be auto-generated from services using convention system")

	// ==============================================================================
	// STEP 3: Register server and run
	// ==============================================================================
	// Register config after setting server name
	lokstra_registry.RegisterConfig(cfg, *serverName)

	// Print server info
	lokstra_registry.PrintServerStartInfo()

	// Run server with graceful shutdown
	if err := lokstra_registry.RunServer(5 * time.Second); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}

	fmt.Println("👋 Goodbye!")
}
