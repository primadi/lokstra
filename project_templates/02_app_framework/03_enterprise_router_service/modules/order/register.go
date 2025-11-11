package order

import (
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/project_templates/02_app_framework/03_enterprise_router_service/modules/order/application"
	"github.com/primadi/lokstra/project_templates/02_app_framework/03_enterprise_router_service/modules/order/infrastructure/repository"
)

// Register registers all order module service types
// This is owned by the module and defines intrinsic routing behavior
func Register() {
	// Register order repository (infrastructure - local only)
	lokstra_registry.RegisterServiceType("order-repository-factory",
		repository.OrderRepositoryFactory, nil)

	lokstra_registry.RegisterLazyService("order-repository", "order-repository-factory", nil)

	application.Register()
}
