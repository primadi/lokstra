package main

import (
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/config"
)

// HealthCheckHandler returns health status
func HealthCheckHandler() map[string]string {
	return map[string]string{"status": "ok"}
}

// VersionHandler returns version info
func VersionHandler() map[string]string {
	return map[string]string{"version": "1.0.0"}
}

func main() {
	// Create router and register routes
	r := lokstra.NewRouter("api")
	r.GET("/health", HealthCheckHandler)
	r.GET("/version", VersionHandler)

	// Load configuration
	cfg := config.New()
	if err := config.LoadConfigFile("config.yaml", cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create app from config
	app := lokstra.NewApp("api-app", ":8080", r)

	fmt.Println("🚀 Server starting on http://localhost:8080")
	fmt.Println("📖 Try:")
	fmt.Println("   curl http://localhost:8080/health")
	fmt.Println("   curl http://localhost:8080/version")

	// Start server with graceful shutdown
	app.Run(30 * time.Second)
}
