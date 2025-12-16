package main

import (
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Register service type factories
	registerServiceTypes()

	// Get config path
	// configPath := filepath.Join("docs", "00-introduction", "examples", "full-framework", "06-inline-definitions-example", "config.yaml")

	if _, err := loader.LoadConfig(); err != nil {
		logger.LogPanic("❌ Failed to load config:", err)
	}

	if err := lokstra_registry.RunConfiguredServer(); err != nil {
		logger.LogPanic("❌ Failed to run server:", err)
	}

	// Load config and run server
	// if err := lokstra_registry.LoadAndBuild([]string{"config.yaml"}); err != nil {
	// 	log.Fatalf("Failed to load config: %v", err)
	// }

	// // Run the development server (demonstrates inline definitions)
	// if err := lokstra_registry.RunServer("development.dev-server", 30*time.Second); err != nil {
	// 	log.Fatalf("Server error: %v", err)
	// }
}
