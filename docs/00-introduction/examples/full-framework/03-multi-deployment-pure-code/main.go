package main

import (
	"flag"
	"time"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Parse command line flags
	server := flag.String("server", "monolith.api-server",
		"Server to run (monolith.api-server or microservice.user-server, microservice.user-server, or microservice.order-server)")
	flag.Parse()

	logger.LogInfo("")
	logger.LogInfo("╔═════════════════════════════════════════════╗")
	logger.LogInfo("║   LOKSTRA MULTI-DEPLOYMENT DEMO             ║")
	logger.LogInfo("╚═════════════════════════════════════════════╝")
	logger.LogInfo("")
	// 1. Register service types
	registerServiceTypes()

	// 2, Register middleware types
	registerMiddlewareTypes()

	// 2. Load config (loads ALL deployments into Global registry)
	// (no more YAML needed!)
	loadconfigFromCode()

	// 3. Run server (no more YAML needed!)
	if err := lokstra_registry.RunServer(*server, 30*time.Second); err != nil {
		logger.LogPanic("❌ Failed to run server:", err)
	}
}
