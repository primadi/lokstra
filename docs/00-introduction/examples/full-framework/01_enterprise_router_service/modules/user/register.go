package user

import (
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/01_enterprise_router_service/modules/user/application"
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/01_enterprise_router_service/modules/user/infrastructure/repository"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Register registers all user module service types
// This is owned by the module and defines intrinsic routing behavior
func Register() {
	// Register user repository (infrastructure - local only)
	lokstra_registry.RegisterServiceType("user-repository-factory",
		repository.UserRepositoryFactory, nil)

	lokstra_registry.RegisterLazyService("user-repository", "user-repository-factory", nil)

	application.Register()
}
