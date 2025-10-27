package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/config"
)

// HealthCheckHandler returns health status
func HealthCheckHandler() map[string]string {
	return map[string]string{"status": "ok"}
}

// InfoHandler returns environment information
func InfoHandler() map[string]interface{} {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	return map[string]interface{}{
		"environment": env,
		"app_name":    os.Getenv("APP_NAME"),
		"app_port":    os.Getenv("APP_PORT"),
	}
}

func main() {
	// Get environment (defaults to "dev")
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	fmt.Printf("🔧 Starting application in %s environment\n\n", env)

	// Create router and register routes
	r := lokstra.NewRouter("api")
	r.GET("/health", HealthCheckHandler)
	r.GET("/info", InfoHandler)

	// Load configuration
	cfg := config.New()

	// Load base configuration
	if err := config.LoadConfigFile("config/base.yaml", cfg); err != nil {
		log.Fatalf("Failed to load base config: %v", err)
	}

	// Load environment-specific configuration
	envConfigFile := fmt.Sprintf("config/%s.yaml", env)
	if err := config.LoadConfigFile(envConfigFile, cfg); err != nil {
		log.Fatalf("Failed to load %s config: %v", env, err)
	}

	// Determine port based on environment
	port := ":8080"
	if env == "dev" {
		port = ":3000"
	}

	// Create app
	app := lokstra.NewApp("main-app", port, r)

	fmt.Printf("🚀 Server starting on http://localhost%s\n", port)
	fmt.Println("📖 Try:")
	fmt.Printf("   curl http://localhost%s/health\n", port)
	fmt.Printf("   curl http://localhost%s/info\n", port)
	fmt.Println("\n💡 To change environment:")
	fmt.Println("   APP_ENV=prod go run main.go")

	// Start server with graceful shutdown
	app.Run(30 * time.Second)
}
