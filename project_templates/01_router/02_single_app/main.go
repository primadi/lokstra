package main

import (
	"log"
	"time"

	"github.com/primadi/lokstra"
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
	log.Println("Starting application...")
	log.Println("Press Ctrl+C to gracefully shutdown")
	if err := app.Run(30 * time.Second); err != nil {
		log.Fatal("Failed to start the app:", err)
	}

	log.Println("Application stopped gracefully")
}
