package lokstra_registry

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/router/autogen"
	"github.com/primadi/lokstra/core/router/convention"
	"github.com/primadi/lokstra/core/service"
)

// NewRouterFromService creates a router from a service instance using metadata.
// This is a convenience helper for the Service as Router pattern.
//
// The metadata (typically from XXXRemote struct) provides:
//   - Resource name (singular/plural)
//   - Convention (REST, RPC, etc.)
//   - Route overrides (optional)
//
// Example:
//
//	userSvc := NewUserService()
//	userMeta := NewUserServiceRemote()
//	router := lokstra_registry.NewRouterFromService(userSvc, userMeta, nil)
//
//	app := lokstra.NewApp("api", ":3000", router)
//
// The router will have auto-generated endpoints based on service methods:
//   - GET    /users       → List()
//   - GET    /users/{id}  → GetByID()
//   - POST   /users       → Create()
//   - PUT    /users/{id}  → Update()
//   - DELETE /users/{id}  → Delete()
func NewRouterFromService(
	serviceInstance any,
	metadata service.RemoteServiceMeta,
	override *autogen.RouteOverride,
) lokstra.Router {
	resource, plural := metadata.GetResourceName()
	conventionName := metadata.GetConventionName()

	rule := autogen.ConversionRule{
		Convention:     convention.ConventionType(conventionName),
		Resource:       resource,
		ResourcePlural: plural,
	}

	var routeOverride autogen.RouteOverride
	if override != nil {
		routeOverride = *override
	} else {
		// Use default override from metadata
		routeOverride = metadata.GetRouteOverride()
	}

	return autogen.NewFromService(serviceInstance, rule, routeOverride)
}

// NewRouterFromServiceType creates a router from a service instance using metadata
// from RegisterServiceType options.
//
// This is even simpler than NewRouterFromService - no need for Remote struct!
//
// Example:
//
//	// Register with metadata
//	lokstra_registry.RegisterServiceType(
//	    "user-service",
//	    NewUserService,
//	    nil,
//	    deploy.WithResource("user", "users"),
//	    deploy.WithConvention("rest"),
//	)
//
//	// Create router from metadata
//	userSvc := NewUserService()
//	router := lokstra_registry.NewRouterFromServiceType("user-service", userSvc)
//
// The router will have auto-generated endpoints based on service methods.
func NewRouterFromServiceType(
	serviceType string,
	serviceInstance any,
) lokstra.Router {
	// Get metadata from global registry
	metadata := deploy.Global().GetServiceMetadata(serviceType)
	if metadata == nil {
		panic(fmt.Sprintf("service type '%s' not found or has no metadata", serviceType))
	}

	// Build conversion rule from metadata
	rule := autogen.ConversionRule{
		Convention:     convention.ConventionType(metadata.Convention),
		Resource:       metadata.Resource,
		ResourcePlural: metadata.ResourcePlural,
	}

	// Build route override from metadata
	override := autogen.RouteOverride{
		PathPrefix: metadata.PathPrefix,
		Hidden:     metadata.HiddenMethods,
		Custom:     make(map[string]autogen.Route),
	}

	// Convert route overrides from metadata
	for methodName, path := range metadata.RouteOverrides {
		override.Custom[methodName] = autogen.Route{
			Path: path,
		}
	}

	return autogen.NewFromService(serviceInstance, rule, override)
}
