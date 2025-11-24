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

// GetUsersHandler returns sample users
func GetUsersHandler() map[string]any {
	return map[string]any{
		"users": []map[string]any{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
		},
		"count": 2,
	}
}

// GetConfigHandler returns configuration info
func GetConfigHandler() map[string]any {
	return map[string]any{
		"message":     "Configuration loaded successfully",
		"api_version": "v1",
		"base_path":   "/api",
		"features": map[string]bool{
			"auth":    true,
			"logging": true,
			"metrics": false,
		},
	}
}

func main() {
	// Create router and register routes
	r := lokstra.NewRouter("api")
	r.GET("/api/health", HealthCheckHandler)
	r.GET("/api/v1/users", GetUsersHandler)
	r.GET("/api/v1/config", GetConfigHandler)

	// Load configuration with CFG references
	cfg := config.New()
	if err := config.LoadConfigFile("config.yaml", cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create app
	app := lokstra.NewApp("api-app", ":8080", r)

	fmt.Println("🚀 Server starting on http://localhost:8080")
	fmt.Println("\n📖 Try:")
	fmt.Println("   curl http://localhost:8080/api/health")
	fmt.Println("   curl http://localhost:8080/api/v1/users")
	fmt.Println("   curl http://localhost:8080/api/v1/config")
	fmt.Println("\n💡 This example demonstrates:")
	fmt.Println("   - CFG references: ${@CFG:path.to.value}")
	fmt.Println("   - Shared config values in 'configs' section")
	fmt.Println("   - DRY principle in configuration")

	// Start server with graceful shutdown
	if err := app.Run(30 * time.Second); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
