package main

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	lokstra.Bootstrap()

	// Load multiple config files in order
	// Later files override earlier ones
	// Base config + environment-specific config
	lokstra_registry.RunServerFromConfig(
		"config/base.yaml",
		"config/dev.yaml", // or production.yaml for prod
	)
}
