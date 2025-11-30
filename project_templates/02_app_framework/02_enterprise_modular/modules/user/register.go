package user

import (
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/user/application"
	"github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/user/infrastructure/repository"
)

// Register registers all user module service types
// This is owned by the module and defines intrinsic routing behavior
func Register() {
	// Register user repository (infrastructure - local only)
	lokstra_registry.RegisterServiceType("user-repository-factory",
		repository.UserRepositoryFactory)

	// Register user service (application - local and remote)
	lokstra_registry.RegisterRouterServiceType("user-service-factory",
		application.UserServiceFactory,
		UserServiceRemoteFactory, // Remote factory for microservices
		&deploy.ServiceTypeConfig{
			PathPrefix:  "/api",
			Middlewares: []string{"recovery", "request-logger"},
			RouteOverrides: map[string]deploy.RouteConfig{
				"Suspend":  {Method: "POST", Path: "/user/{id}/suspend"},
				"Activate": {Method: "POST", Path: "/user/{id}/activate"},
			},
		},
	)

	lokstra_registry.RegisterLazyService("user-repository", "user-repository-factory", nil)
	lokstra_registry.RegisterLazyService("user-service", "user-service-factory", map[string]any{
		"depends-on": []string{"user-repository"},
	})
}

// UserServiceRemoteFactory creates a remote HTTP client for UserService
func UserServiceRemoteFactory(deps, config map[string]any) any {
	// Extract proxy.Service from config (provided by framework)
	proxyService, ok := config["remote"].(*proxy.Service)
	if !ok {
		panic("remote factory requires 'remote' (proxy.Service) in config")
	}

	// Return wrapper that implements UserService interface
	return application.NewUserServiceRemote(proxyService)
}
