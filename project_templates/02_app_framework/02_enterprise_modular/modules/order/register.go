package order

import (
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/order/application"
	"github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/order/infrastructure/repository"
)

// Register registers all order module service types
// This is owned by the module and defines intrinsic routing behavior
func Register() {
	// Register order repository (infrastructure - local only)
	lokstra_registry.RegisterServiceType("order-repository-factory",
		repository.OrderRepositoryFactory, nil)

	// Register order service (application - local and remote)
	lokstra_registry.RegisterServiceType("order-service-factory",
		application.OrderServiceFactory,
		OrderServiceRemoteFactory, // Remote factory for microservices
		deploy.WithRouter(&deploy.ServiceTypeRouter{
			PathPrefix:  "/api",
			Middlewares: []string{"recovery", "request-logger"},
			CustomRoutes: map[string]string{
				"UpdateStatus": "PUT /orders/{id}/status",
				"Cancel":       "POST /orders/{id}/cancel",
			},
		}),
	)

	lokstra_registry.RegisterLazyService("order-repository", "order-repository-factory", nil)
	lokstra_registry.RegisterLazyService("order-service", "order-service-factory", map[string]any{
		"depends-on": []string{"order-repository", "user-service"},
	})
}

// OrderServiceRemoteFactory creates a remote HTTP client for OrderService
func OrderServiceRemoteFactory(deps, config map[string]any) any {
	// Extract proxy.Service from config (provided by framework)
	proxyService, ok := config["remote"].(*proxy.Service)
	if !ok {
		panic("remote factory requires 'remote' (proxy.Service) in config")
	}

	// Return wrapper that implements OrderService interface
	return application.NewOrderServiceRemote(proxyService)
}
