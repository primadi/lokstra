package main

import (
	"fmt"

	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	fmt.Println("🚀 Starting Lokstra E-commerce Demo with Auto Router Discovery")

	// Register routers in code - no need for routers: section in config.yaml
	setupRouters()

	cfg := config.New()
	configFiles := []string{"config.yaml"}

	// load multiple config files, later files override earlier ones
	for _, file := range configFiles {
		if err := config.LoadConfigFile(file, cfg); err == nil {
			fmt.Printf("📄 Loaded config: %s\n", file)
		} else {
			fmt.Printf("❌ Failed to load config: %s - %v\n", file, err)
			return
		}
	}

	lokstra_registry.RegisterConfig(cfg)

	serverName := lokstra_registry.GetConfig("server-name", "monolith-single-port-server")
	lokstra_registry.SetCurrentServerName(serverName)

	lokstra_registry.PrintServerStartInfo()
	lokstra_registry.StartServer()
}
