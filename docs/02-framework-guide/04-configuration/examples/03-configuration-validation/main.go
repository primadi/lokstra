package main

import (
	"fmt"
	"log"
	"os"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/config"
)

// ValidateConfig checks if required configs are present
func ValidateConfig(cfg *config.Config) error {
	// Check required configs
	required := []string{"app_name", "app_version", "db_host"}

	configMap := make(map[string]bool)
	for _, c := range cfg.Configs {
		configMap[c.Name] = true
	}

	for _, req := range required {
		if !configMap[req] {
			return fmt.Errorf("missing required config: %s", req)
		}
	}

	return nil
}

// Status handler
func StatusHandler() map[string]any {
	return map[string]any{
		"status":  "ok",
		"message": "All required configurations are valid",
	}
}

// Home handler
func HomeHandler() string {
	return `
	<html>
	<body>
		<h1>Configuration Validation Example</h1>
		<p>Validates required configuration before startup</p>
		<h2>Endpoints</h2>
		<ul>
			<li><a href="/status">Status</a> - Check validation status</li>
		</ul>
		<h2>Try Different Configs</h2>
		<pre>
# Valid config
go run main.go valid

# Invalid config (missing required fields)
go run main.go invalid
		</pre>
	</body>
	</html>
	`
}

func main() {
	// Load config
	configFile := "config-valid.yaml"
	if len(os.Args) > 1 && os.Args[1] == "invalid" {
		configFile = "config-invalid.yaml"
	}

	cfg := config.New()
	log.Printf("Loading configuration: %s", configFile)

	if err := config.LoadConfigFile(configFile, cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	log.Println("Validating configuration...")
	if err := ValidateConfig(cfg); err != nil {
		log.Fatalf("❌ Configuration validation failed: %v", err)
	}

	log.Println("✅ Configuration validation passed!")

	// Create router
	router := lokstra.NewRouter("main")
	router.GET("/", HomeHandler)
	router.GET("/status", StatusHandler)

	// Create and run app
	app := lokstra.NewApp("config-validation", ":3030", router)

	log.Println("Starting server on :3030")
	if err := app.Run(0); err != nil {
		log.Fatal(err)
	}
}
