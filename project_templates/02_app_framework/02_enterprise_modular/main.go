package main

import (
	"fmt"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	fmt.Println("")
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║   LOKSTRA ENTERPRISE MODULAR TEMPLATE         ║")
	fmt.Println("║   Domain-Driven Design with Bounded Contexts  ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Println("")

	logger.SetLogLevelFromEnv()

	// 1. Register service types from all modules
	registerServiceTypes()

	// 2. Register middleware types
	registerMiddlewareTypes()

	// 3. Run server from config folder
	// Lokstra will automatically merge all YAML files in config/ folder
	if _, err := loader.LoadConfig("config"); err != nil {
		logger.LogPanic("❌ Failed to load config:", err)
	}

	if err := lokstra_registry.RunConfiguredServer(); err != nil {
		logger.LogPanic("❌ Failed to run server:", err)
	}
}
