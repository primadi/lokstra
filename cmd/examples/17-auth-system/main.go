package main

import (
	"fmt"

	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/lokstra_registry"
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

	lokstra_registry.RegisterConfig(cfg)

	serverName := lokstra_registry.GetConfig("server-name", "auth-server")
	lokstra_registry.SetCurrentServerName(serverName)

	lokstra_registry.PrintServerStartInfo()
	lokstra_registry.StartServer()
}
