package main

import (
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/old_registry"
)

func main() {
	// Load configuration
	log.Println("=======================================")
	log.Println("🚀 Starting Lokstra Reverse Proxy Test")
	log.Println("=======================================")

	// Register service factories (mock implementations)
	registerServiceFactories()

	// Load and parse config file
	cfg := config.New()
	if err := config.LoadConfigFile("config.yaml", cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	old_registry.RegisterConfig(cfg, "demo-api-gateway")

	// Print server info
	old_registry.PrintServerStartInfo()

	// Run server with graceful shutdown
	if err := old_registry.RunServer(5 * time.Second); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
	fmt.Println("Server ended")
}

func registerServiceFactories() {
	// Register User Service (for App1)
	old_registry.RegisterServiceType("user_service", UserServiceFactory)

	// Register Product Service (for App2)
	old_registry.RegisterServiceType("product_service", ProductServiceFactory)
}
