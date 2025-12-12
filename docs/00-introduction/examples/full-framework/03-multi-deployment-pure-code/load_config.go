package main

import (
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/lokstra_registry"
)

func loadconfigFromCode() {
	// ---------------------------------------------------------------
	// 1. Register service definitions
	// (replaces YAML service-definitions section)

	// service-definitions:
	//   user-repository:
	//      type: user-repository-factory
	//    config:
	//      dsn: "user:password@tcp(localhost:3306)/users_db"
	lokstra_registry.RegisterLazyService("user-repository",
		"user-repository-factory", map[string]any{
			"dsn": "user:password@tcp(localhost:3306)/users_db",
		})

	// service-definitions:
	//   order-repository:
	//      type: order-repository-factory
	//    config:
	//      dsn: "user:password@tcp(localhost:3306)/orders_db"
	lokstra_registry.RegisterLazyService("order-repository",
		"order-repository-factory", map[string]any{
			"dsn": "user:password@tcp(localhost:3306)/orders_db",
		})

	// service-definitions:
	//   user-service:
	//      type: user-service-factory
	//      depends-on: [user-repository]
	lokstra_registry.RegisterLazyService("user-service",
		"user-service-factory", map[string]any{
			"depends-on": []string{"user-repository"},
		})

	// service-definitions:
	//   order-service:
	//      type: order-service-factory
	//      depends-on: [order-repository, user-service]
	lokstra_registry.RegisterLazyService("order-service",
		"order-service-factory", map[string]any{
			"depends-on": []string{"order-repository", "user-service"},
		})

	// ---------------------------------------------------------------
	// 2. Register deployments
	// (replaces YAML deployments section)

	// Monolith: All services in one process

	// deployments:
	//   monolith:
	//     servers:
	//       api-server:
	//         base-url: "http://localhost"
	//         addr: ":3003"
	//         published-services: [user-service, order-service]
	err := lokstra_registry.RegisterDeployment("monolith",
		&lokstra_registry.DeploymentConfig{
			Servers: map[string]*lokstra_registry.ServerConfig{
				"api-server": {
					BaseURL:           "http://localhost",
					Addr:              ":3003",
					PublishedServices: []string{"user-service", "order-service"},
				},
			},
		})
	if err != nil {
		logger.LogInfo("❌ Failed to register monolith deployment:", err)
	}

	// Microservice: Each service in its own process

	// deployments:
	//   microservice:
	//     servers:
	//       user-server:
	//         base-url: "http://localhost"
	//         addr: ":3004"
	//         published-services: [user-service]
	//       order-server:
	//         base-url: "http://localhost"
	//         addr: ":3005"
	//         published-services: [order-service]
	err = lokstra_registry.RegisterDeployment("microservice",
		&lokstra_registry.DeploymentConfig{
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
		logger.LogPanic("❌ Failed to register microservice deployment:", err)
	}
}
