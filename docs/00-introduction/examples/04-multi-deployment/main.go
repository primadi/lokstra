package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/docs/00-introduction/examples/04-multi-deployment/repository"
	"github.com/primadi/lokstra/docs/00-introduction/examples/04-multi-deployment/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Parse command line flags
	server := flag.String("server", "monolith.api-server", "Server to run (monolith.api-server or microservice.user-server, microservice.user-server, or microservice.order-server)")
	flag.Parse()

	fmt.Println("")
	fmt.Println("╔═════════════════════════════════════════════╗")
	fmt.Println("║   LOKSTRA MULTI-DEPLOYMENT DEMO             ║")
	fmt.Println("╚═════════════════════════════════════════════╝")
	fmt.Println("")

	// 1. Register service factories
	// Clean Architecture: Separate layers for contract, model, service, repository

	// Register repositories (infrastructure layer)
	lokstra_registry.RegisterServiceType("user-repository-factory",
		repository.NewUserRepositoryMemory, nil)

	lokstra_registry.RegisterServiceType("order-repository-factory",
		repository.NewOrderRepositoryMemory, nil)

	// Register services (application layer)
	// Metadata provided via RegisterServiceType options (not in factory structs)
	lokstra_registry.RegisterServiceType("user-service-factory",
		service.UserServiceFactory,
		service.UserServiceRemoteFactory,
		deploy.WithResource("user", "users"),
		deploy.WithConvention("rest"),
	)

	lokstra_registry.RegisterServiceType("order-service-factory",
		service.OrderServiceFactory,
		service.OrderServiceRemoteFactory,
		deploy.WithResource("order", "orders"),
		deploy.WithConvention("rest"),
		deploy.WithRouteOverride("GetByUserID", "/users/{user_id}/orders"),
	)

	// Replace YAML service-definitions with code
	// Equivalent to:
	//   user-service:
	//     type: user-service-factory
	//     depends-on:
	//       - user-repository
	lokstra_registry.RegisterLazyService(
		"user-service",         // service name
		"user-service-factory", // service type
		map[string]any{
			"depends-on": []string{"user-repository"}, // dep key -> service name
		},
	)

	// 2. Load config (loads ALL deployments into Global registry)
	if err := lokstra_registry.LoadAndBuild([]string{"config.yaml"}); err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	// 3. Run server
	if err := lokstra_registry.RunServer(*server, 30*time.Second); err != nil {
		log.Fatal("❌ Failed to run server:", err)
	}
}
