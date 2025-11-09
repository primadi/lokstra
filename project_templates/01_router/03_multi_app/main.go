package main

import (
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/project_templates/01_router/03_multi_app/adminapp"
	"github.com/primadi/lokstra/project_templates/01_router/03_multi_app/mainapp"
)

func main() {
	// Create main API app on port 3000
	mainApp := mainapp.CreateApp()

	// Create admin API app on port 3001
	adminApp := adminapp.CreateApp()

	// Create a server to orchestrate multiple apps
	server := lokstra.NewServer("demo-server", mainApp, adminApp)

	// Print startup information for all apps
	server.PrintStartInfo()

	// Run the server - starts all apps with graceful shutdown (30 second timeout)
	// Press Ctrl+C to gracefully shutdown all apps
	log.Println("Starting multi-app server...")
	log.Println("Main API:  http://localhost:3000")
	log.Println("Admin API: http://localhost:3001")
	log.Println("Press Ctrl+C to gracefully shutdown")

	if err := server.Run(30 * time.Second); err != nil {
		log.Fatal("Failed to start the server:", err)
	}

	log.Println("All applications stopped gracefully")
}
