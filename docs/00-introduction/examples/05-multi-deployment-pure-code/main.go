package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/docs/00-introduction/examples/05-multi-deployment-pure-code/repository"
	"github.com/primadi/lokstra/docs/00-introduction/examples/05-multi-deployment-pure-code/service"
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

	// 2. Register service definitions (replaces YAML service-definitions section)
	lokstra_registry.RegisterLazyService("user-repository", "user-repository-factory", nil)
	lokstra_registry.RegisterLazyService("order-repository", "order-repository-factory", nil)

	lokstra_registry.RegisterLazyService("user-service", "user-service-factory", map[string]any{
		"depends-on": []string{"user-repository"},
	})

	lokstra_registry.RegisterLazyService("order-service", "order-service-factory", map[string]any{
		"depends-on": []string{"order-repository", "user-service"},
	})

	// 3. Register deployments (replaces YAML deployments section)
	// Monolith: All services in one process
	err := lokstra_registry.RegisterDeployment("monolith", &lokstra_registry.DeploymentConfig{
		Servers: map[string]*lokstra_registry.ServerConfig{
			"api-server": {
				BaseURL:           "http://localhost",
				Addr:              ":3003",
				PublishedServices: []string{"user-service", "order-service"},
			},
		},
	})
	if err != nil {
		log.Fatal("❌ Failed to register monolith deployment:", err)
	}

	// Microservice: Each service in its own process
	err = lokstra_registry.RegisterDeployment("microservice", &lokstra_registry.DeploymentConfig{
		Servers: map[string]*lokstra_registry.ServerConfig{
			"user-server": {
				BaseURL:           "http://localhost",
				Addr:              ":3004",
				PublishedServices: []string{"user-service"},
			},
			"order-server": {
				BaseURL:           "http://localhost",
				Addr:              ":3005",
				PublishedServices: []string{"order-service"},
			},
		},
	})
	if err != nil {
		log.Fatal("❌ Failed to register microservice deployment:", err)
	}

	// 4. Run server (no more YAML needed!)
	if err := lokstra_registry.RunServer(*server, 30*time.Second); err != nil {
		log.Fatal("❌ Failed to run server:", err)
	}
}
