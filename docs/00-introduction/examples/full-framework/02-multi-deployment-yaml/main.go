package main

import (
	"fmt"

	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Parse command line flags
	// server := flag.String("server", "monolith.api-server",
	// 	"Server to run (monolith.api-server or microservice.user-server, microservice.user-server, or microservice.order-server)")
	// flag.Parse()

	fmt.Println("")
	fmt.Println("╔═════════════════════════════════════════════╗")
	fmt.Println("║   LOKSTRA MULTI-DEPLOYMENT DEMO             ║")
	fmt.Println("╚═════════════════════════════════════════════╝")
	fmt.Println("")

	// 1. Register service types
	registerServiceTypes()

	// 2. Register middleware types
	registerMiddlewareTypes()

	// 3. RunServerFromConfig
	lokstra_registry.RunServerFromConfig()
}
