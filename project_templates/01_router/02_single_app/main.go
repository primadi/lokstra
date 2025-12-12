package main

import (
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/logger"
)

func main() {
	// Create routers
	apiRouter := setupAPIRouter()
	healthRouter := setupHealthRouter()

	// Create a single app with multiple routers
	// Health check router is added first, so it takes priority
	app := lokstra.NewApp("demo-app", ":3000", healthRouter, apiRouter)

	// Print application startup information
	app.PrintStartInfo()

	// Run the app with graceful shutdown (30 second timeout)
	// This handles SIGINT/SIGTERM signals automatically
	logger.LogInfo("Starting application...")
	logger.LogInfo("Press Ctrl+C to gracefully shutdown")
	if err := app.Run(30 * time.Second); err != nil {
		logger.LogPanic("Failed to start the app: %v", err)
	}

	logger.LogInfo("Application stopped gracefully")
}
