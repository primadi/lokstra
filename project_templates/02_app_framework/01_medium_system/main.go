package main

import (
	"fmt"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	fmt.Println("")
	fmt.Println("╔═════════════════════════════════════════════╗")
	fmt.Println("║   LOKSTRA MEDIUM SYSTEM TEMPLATE            ║")
	fmt.Println("║   Domain-Driven Modular Architecture        ║")
	fmt.Println("╚═════════════════════════════════════════════╝")
	fmt.Println("")

	// Set log level from environment variable LOKSTRA_LOG_LEVEL
	// Supported values: silent, error, warn, info, debug
	// Default: info
	deploy.SetLogLevelFromEnv()

	// Or set manually:
	// deploy.SetLogLevel(deploy.LogLevelDebug)  // Show all debug logs
	// deploy.SetLogLevel(deploy.LogLevelInfo)   // Default
	// deploy.SetLogLevel(deploy.LogLevelSilent) // No logs

	// 1. Register service types
	registerServiceTypes()

	// 2. Register middleware types
	registerMiddlewareTypes()

	// 3. Run server from config
	lokstra_registry.RunServerFromConfig()
}
