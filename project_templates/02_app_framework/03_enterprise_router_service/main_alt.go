package main

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/lokstra_registry"
)

func AltMain() {
	lokstra.Bootstrap()

	fmt.Println("")
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║   LOKSTRA ENTERPRISE MODULAR TEMPLATE         ║")
	fmt.Println("║   Domain-Driven Design with Bounded Contexts  ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Println("")

	deploy.SetLogLevelFromEnv()

	// 1. Register service types from all modules
	registerServiceTypes()

	// 2. Register middleware types
	registerMiddlewareTypes()

	// 3. Run server from config folder
	// Lokstra will automatically merge all YAML files in config/ folder
	lokstra_registry.RunServerFromConfigFolder("config")
}
