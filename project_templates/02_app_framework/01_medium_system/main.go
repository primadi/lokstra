package main

import (
	"fmt"

	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	fmt.Println("")
	fmt.Println("╔═════════════════════════════════════════════╗")
	fmt.Println("║   LOKSTRA MEDIUM SYSTEM TEMPLATE            ║")
	fmt.Println("║   Domain-Driven Modular Architecture        ║")
	fmt.Println("╚═════════════════════════════════════════════╝")
	fmt.Println("")

	// 1. Register service types
	registerServiceTypes()

	// 2. Register middleware types
	registerMiddlewareTypes()

	// 3. Run server from config
	lokstra_registry.RunServerFromConfig()
}
