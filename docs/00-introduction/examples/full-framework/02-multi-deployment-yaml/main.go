package main

import (
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Parse command line flags
	// server := flag.String("server", "monolith.api-server",
	// 	"Server to run (monolith.api-server or microservice.user-server, microservice.user-server, or microservice.order-server)")
	// flag.Parse()

	logger.LogInfo("")
	logger.LogInfo("╔═════════════════════════════════════════════╗")
	logger.LogInfo("║   LOKSTRA MULTI-DEPLOYMENT DEMO             ║")
	logger.LogInfo("╚═════════════════════════════════════════════╝")
	logger.LogInfo("")
	// 1. Register service types
	registerServiceTypes()

	// 2. Register middleware types
	registerMiddlewareTypes()

	// 3. RunServerFromConfig
	if err := lokstra_registry.LoadConfig(); err != nil {
		logger.LogPanic("❌ Failed to load config:", err)
	}

	if err := lokstra_registry.RunConfiguredServer(); err != nil {
		logger.LogPanic("❌ Failed to run server:", err)
	}
}
