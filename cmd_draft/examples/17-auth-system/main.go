package main

import (
	"fmt"

	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/old_registry"
)

func main() {
	fmt.Println("🚀 Starting Lokstra Auth System Demo")
	fmt.Println("=====================================")

	// Register all services and middleware
	registerServices()
	registerMiddleware()

	// Setup routers
	setupRouters()

	// Load configuration
	cfg := config.New()
	configFile := "config.yaml"

	if err := config.LoadConfigFile(configFile, cfg); err != nil {
		fmt.Printf("❌ Failed to load config: %s - %v\n", configFile, err)
		return
	}

	fmt.Printf("📄 Loaded config: %s\n", configFile)

	old_registry.RegisterConfig(cfg, "")

	serverName := old_registry.GetConfig("server-name", "auth-server")
	old_registry.SetCurrentServerName(serverName)

	old_registry.PrintServerStartInfo()
	if err := old_registry.StartServer(); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
	}
}
