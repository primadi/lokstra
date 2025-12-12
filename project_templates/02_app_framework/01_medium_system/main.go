package main

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/logger"
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
	logger.SetLogLevelFromEnv()

	// Or set manually:
	// logger.SetLogLevel(logger.LogLevelDebug)  // Show all debug logs
	// logger.SetLogLevel(logger.LogLevelInfo)   // Default
	// logger.SetLogLevel(logger.LogLevelSilent) // No logs

	// 1. Register service types
	registerServiceTypes()

	// 2. Register middleware types
	registerMiddlewareTypes()

	// 3. Run server from config
	if err := lokstra.LoadConfig(); err != nil {
		logger.LogPanic("❌ Failed to load config:", err)
	}

	if err := lokstra.RunConfiguredServer(); err != nil {
		logger.LogPanic("❌ Failed to run server:", err)
	}
}
