package main

import (
	"log"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/recovery"
)

func main() {
	// Auto-generate code from @RouterService annotations
	lokstra.Bootstrap()

	// STEP 1: Load Config
	if err := lokstra_registry.LoadConfig("config.yaml"); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// STEP 2: Register Service Types
	registerServiceTypes()

	// STEP 3: Register Middleware Types
	registerMiddlewareTypes()

	// STEP 4: Initialize and Run Server
	if err := lokstra_registry.InitAndRunServer(); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}

func registerServiceTypes() {
	// Register user repository factory
	lokstra_registry.RegisterServiceType(
		"user-repository-factory",
		NewUserRepository,
	)

	// Register as lazy service
	lokstra_registry.RegisterLazyService(
		"user-repository",
		"user-repository-factory",
		nil, // no dependencies
	)
}

func registerMiddlewareTypes() {
	// Register built-in recovery middleware
	recovery.Register()

	// Register request logger middleware
	lokstra_registry.RegisterMiddlewareFactory("request-logger", RequestLoggerMiddleware)
}
