package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Parse command line flags
	server := flag.String("server", "monolith.api-server",
		"Server to run (monolith.api-server or microservice.user-server, microservice.user-server, or microservice.order-server)")
	flag.Parse()

	fmt.Println("")
	fmt.Println("╔═════════════════════════════════════════════╗")
	fmt.Println("║   LOKSTRA MULTI-DEPLOYMENT DEMO             ║")
	fmt.Println("╚═════════════════════════════════════════════╝")
	fmt.Println("")

	// 1. Register service types
	registerServiceTypes()

	// 2, Register middleware types
	registerMiddlewareTypes()

	// 2. Load config (loads ALL deployments into Global registry)
	if err := lokstra_registry.LoadAndBuild([]string{"config.yaml"}); err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	// 3. Run server
	if err := lokstra_registry.RunServer(*server, 30*time.Second); err != nil {
		log.Fatal("❌ Failed to run server:", err)
	}
}
