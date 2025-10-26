package lokstra_registry

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/router/autogen"
	"github.com/primadi/lokstra/core/router/convention"
)

// NewRouterFromServiceType creates a router from a service type using metadata
// from RegisterServiceType options.
//
// This eliminates the need for Remote struct - all metadata comes from registration!
// The service instance is automatically created from the registered factory.
//
// Example:
//
//	// Register with metadata
//	lokstra_registry.RegisterServiceType(
//	    "user-service-factory",
//	    service.UserServiceFactory,
//	    nil,
//	    deploy.WithResource("user", "users"),
//	    deploy.WithConvention("rest"),
//	)
//
//	// Create router directly from service type (no manual instance creation!)
//	router := lokstra_registry.NewRouterFromServiceType("user-service-factory")
//
// The router will have auto-generated endpoints based on service methods.
func NewRouterFromServiceType(serviceType string) lokstra.Router {
	// Get metadata from service type
	metadata := deploy.Global().GetServiceMetadata(serviceType)
	if metadata == nil {
		panic(fmt.Sprintf("service type '%s' not found or has no metadata", serviceType))
	}

	// Get local factory to create service instance
	factory := deploy.Global().GetServiceFactory(serviceType, true) // true = local factory
	if factory == nil {
		panic(fmt.Sprintf("local factory for service type '%s' not found", serviceType))
	}

	// Create service instance (no deps, no config for router usage)
	serviceInstance := factory(nil, nil)

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
