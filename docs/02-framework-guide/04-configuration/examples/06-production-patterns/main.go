package main

import (
	"log"
	"os"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/config"
)

// Health check handler
func HealthHandler() map[string]any {
	return map[string]any{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"env":       os.Getenv("APP_ENV"),
	}
}

// Metrics handler
func MetricsHandler() map[string]any {
	return map[string]any{
		"requests_total": 1234,
		"errors_total":   5,
		"uptime_seconds": 3600,
	}
}

// Home handler
func HomeHandler() string {
	return `
	<html>
	<body>
		<h1>Production Configuration Patterns</h1>
		<p>Best practices for production deployments</p>
		<h2>Endpoints</h2>
		<ul>
			<li><a href="/health">Health</a> - Health check endpoint</li>
			<li><a href="/metrics">Metrics</a> - Application metrics</li>
		</ul>
		<h2>Production Patterns</h2>
		<ul>
			<li>✅ Health check endpoints</li>
			<li>✅ Metrics and monitoring</li>
			<li>✅ Graceful shutdown</li>
			<li>✅ Environment-based configuration</li>
			<li>✅ Structured logging</li>
			<li>✅ Configuration validation</li>
		</ul>
		<h2>Running</h2>
		<pre>
# Development
APP_ENV=development go run main.go

# Production
APP_ENV=production go run main.go production
		</pre>
	</body>
	</html>
	`
}

func main() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Load appropriate config
	configFile := "config-development.yaml"
	if len(os.Args) > 1 && os.Args[1] == "production" {
		configFile = "config-production.yaml"
	}

	cfg := config.New()
	log.Printf("Loading configuration: %s", configFile)

	if err := config.LoadConfigFile(configFile, cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting in %s mode", env)

	// Create router
	router := lokstra.NewRouter("main")
	router.GET("/", HomeHandler)
	router.GET("/health", HealthHandler)
	router.GET("/metrics", MetricsHandler)

	// Create and run app
	app := lokstra.NewApp("production-patterns", ":3060", router)

	log.Println("Starting server on :3060")
	log.Println("Ready to accept connections")

	// Production: use graceful shutdown
	shutdownTimeout := 30 * time.Second
	if env == "production" {
		shutdownTimeout = 60 * time.Second
	}

	if err := app.Run(shutdownTimeout); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stopped gracefully")
}
