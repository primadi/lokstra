package main

import (
	"fmt"
	"log"
	"os"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/config"
)

// Status handler returns environment configuration
func StatusHandler() map[string]any {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	return map[string]any{
		"environment": env,
		"message":     fmt.Sprintf("Running in %s mode", env),
		"database":    os.Getenv("DB_HOST"),
		"log_level":   os.Getenv("LOG_LEVEL"),
	}
}

// Home handler
func HomeHandler() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	return fmt.Sprintf(`
	<html>
	<body>
		<h1>Environment Management Example</h1>
		<p>Current Environment: <strong>%s</strong></p>
		<h2>Endpoints</h2>
		<ul>
			<li><a href="/status">Status</a> - View environment configuration</li>
		</ul>
		<h2>Test Different Environments</h2>
		<pre>
# Development
APP_ENV=development go run main.go

# Staging  
APP_ENV=staging go run main.go

# Production
APP_ENV=production go run main.go
		</pre>
	</body>
	</html>
	`, env)
}

func main() {
	// Determine environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	cfg := config.New()

	// Load base configuration
	log.Println("Loading base configuration...")
	if err := config.LoadConfigFile("config-base.yaml", cfg); err != nil {
		log.Fatalf("Failed to load base config: %v", err)
	}

	// Load environment-specific configuration
	envConfigFile := fmt.Sprintf("config-%s.yaml", env)
	log.Printf("Loading environment configuration: %s", envConfigFile)
	if err := config.LoadConfigFile(envConfigFile, cfg); err != nil {
		log.Printf("Warning: No environment config found for %s, using base only", env)
	}

	// Set environment variables from config
	if len(cfg.Configs) > 0 {
		for _, c := range cfg.Configs {
			if c.Name == "db_host" {
				os.Setenv("DB_HOST", fmt.Sprint(c.Value))
			}
			if c.Name == "log_level" {
				os.Setenv("LOG_LEVEL", fmt.Sprint(c.Value))
			}
		}
	}

	// Create router
	router := lokstra.NewRouter("main")
	router.GET("/", HomeHandler)
	router.GET("/status", StatusHandler)

	// Create and run app
	app := lokstra.NewApp("env-mgmt", ":3020", router)

	log.Printf("Starting server in %s mode on :3020", env)
	if err := app.Run(0); err != nil {
		log.Fatal(err)
	}
}
