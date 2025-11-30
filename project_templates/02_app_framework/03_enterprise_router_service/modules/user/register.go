package user

import (
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/project_templates/02_app_framework/03_enterprise_router_service/modules/user/application"
	"github.com/primadi/lokstra/project_templates/02_app_framework/03_enterprise_router_service/modules/user/infrastructure/repository"
)

// Register registers all user module service types
// This is owned by the module and defines intrinsic routing behavior
func Register() {
	// Register user repository (infrastructure - local only)
	lokstra_registry.RegisterServiceType("user-repository-factory",
		repository.UserRepositoryFactory)

	lokstra_registry.RegisterLazyService("user-repository", "user-repository-factory", nil)

	application.Register()
}
