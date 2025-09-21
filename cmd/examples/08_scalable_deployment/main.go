package main

import (
	"log"
	"os"
	"time"

	"github.com/primadi/lokstra"
)

func main() {
	// Get config file from environment or use default
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.yaml"
	}

	// Initialize Lokstra with config file
	cfg, err := lokstra.LoadConfigFile(configFile)
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	regCtx := lokstra.NewGlobalRegistrationContext()
	svr, err := lokstra.NewServerFromConfig(regCtx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Lokstra: %v", err)
	}

	// Start the server
	log.Printf("Starting server with config: %s", configFile)
	if err := svr.StartWithGracefulShutdown(true, 30*time.Second); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
