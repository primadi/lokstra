package lokstra_registry

import (
	"sync"

	"github.com/primadi/lokstra/api_client"
	"github.com/primadi/lokstra/common/utils"
)

var serviceRegistry sync.Map

type lazyServiceConfig struct {
	serviceName string // Service instance name (e.g., "user-service")
	serviceType string // Service type/factory name (e.g., "user_service")
	config      map[string]any
}

var lazyServiceConfigRegistry sync.Map

// Registers a service instance with a given name.
// If the same service name already exists,
// and the RegisterOption allowOverride is not set to true, it will panic.
func RegisterService(svcName string, svcInstance any, opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	if !options.allowOverride {
		if _, exists := serviceRegistry.Load(svcName); exists {
			panic("service " + svcName + " already registered")
		}
	}

	serviceRegistry.Store(svcName, svcInstance)
}

// Registers a lazy service configuration with a given name.
// The actual service instance will be created when first requested via LazyGetService.
// If the same lazy service name already exists,
// and the RegisterOption allowOverride is not set to true, it will panic.
func RegisterLazyService(svcName string, svcType string,
	config map[string]any, opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	// Check both registries with proper locking
	_, serviceExists := serviceRegistry.Load(svcName)

	if !options.allowOverride {
		if _, exists := lazyServiceConfigRegistry.Load(svcName); exists {
			panic("lazy service " + svcName + " already registered")
		}
		if serviceExists {
			panic("service " + svcName + " already registered")
		}
	}
	lazyServiceConfigRegistry.Store(svcName, lazyServiceConfig{
		serviceName: svcName,
		serviceType: svcType,
		config:      config,
	})
}

// Tries to resolve a service from the registry.
func GetService[T any](svcName string) T {
	if v, ok := TryGetService[T](svcName); ok {
		return v
	}
	var zero T
	return zero
}

func MustGetService[T any](svcName string) T {
	if v, ok := TryGetService[T](svcName); ok {
		return v
	}
	panic("service " + svcName + " not found or type mismatch")
}

// Tries to resolve a service from the registry
func TryGetService[T any](svcName string) (T, bool) {
	// lookup in registry (read lock)
	svc, ok := serviceRegistry.Load(svcName)
	var zero T

	if ok {
		if typed, ok := svc.(T); ok {
			return typed, true
		}
		// type mismatch
		return zero, false
	}

	// not found, check if lazy config exists
	lazyCfgAny, lazyExists := lazyServiceConfigRegistry.Load(svcName)

	if lazyExists {
		lazyCfg := lazyCfgAny.(lazyServiceConfig)
		if factory := GetServiceFactory(lazyCfg.serviceType, lazyCfg.serviceName); factory != nil {
			if svc := factory(lazyCfg.config); svc != nil {
				// Double-check if another goroutine already created it
				if existing, loaded := serviceRegistry.LoadOrStore(svcName, svc); loaded {
					// Another goroutine created it first, use that
					if typed, ok := existing.(T); ok {
						return typed, true
					}
					return zero, false
				}

				// We successfully stored it
				if typed, ok := svc.(T); ok {
					return typed, true
				}
				return zero, false
			}
		}
	}

	// not found
	return zero, false
}

// Create a new service using registered factory, register it, and return it.
// If factory not found or creation failed, return zero value of T.
// It will panic if the created service type does not match T or if
// a service with the same name already exists, unless
// the RegisterOption allowOverride is set to true.
func NewService[T any](svcName, svcType string, config map[string]any,
	opts ...RegisterOption) T {
	if factory := GetServiceFactory(svcType, svcName); factory != nil {
		if svc := factory(config); svc != nil {
			RegisterService(svcName, svc, opts...)
			return svc.(T)
		}
	}
	var zero T
	return zero
}

// ==============================================================================
// Remote Service Helper
// ==============================================================================

// GetRemoteService creates an api_client.RemoteService from configuration map.
// This simplifies remote service factory functions by handling router resolution
// and path-prefix extraction automatically.
//
// Configuration keys:
//   - "router": Router name for client lookup (required)
//   - "path-prefix": API path prefix (optional, defaults to "/")
//   - "convention": Convention for path generation (optional, defaults to "rest")
//   - "resource-name": Resource name singular (optional)
//   - "plural-resource-name": Resource name plural (optional)
//   - "routes": Array of route overrides with {name, method, path} (optional)
//
// Example usage in service factories:
//
//	func CreateAuthServiceRemote(cfg map[string]any) any {
//	    return &authServiceRemote{
//	        client: lokstra_registry.GetRemoteService(cfg),
//	    }
//	}
//
// This replaces the manual pattern:
//
//	routerName := utils.GetValueFromMap(cfg, "router", "service-name")
//	pathPrefix := utils.GetValueFromMap(cfg, "path-prefix", "/path")
//	clientRouter := lokstra_registry.GetClientRouter(routerName)
//	client := api_client.NewRemoteService(clientRouter, pathPrefix)
func GetRemoteService(cfg map[string]any) *api_client.RemoteService {
	routerName := utils.GetValueFromMap(cfg, "router", "")
	if routerName == "" {
		panic("GetRemoteService: 'router' field is required in config")
	}

	pathPrefix := utils.GetValueFromMap(cfg, "path-prefix", "/")
	convention := utils.GetValueFromMap(cfg, "convention", "rest")
	resourceName := utils.GetValueFromMap(cfg, "resource-name", "")
	pluralResourceName := utils.GetValueFromMap(cfg, "plural-resource-name", "")

	// Resolve router using existing GetClientRouter
	clientRouter := GetClientRouter(routerName)

	// Create RemoteService with basic config
	remoteService := api_client.NewRemoteService(clientRouter, pathPrefix)

	// Apply convention and resource names if provided
	if convention != "" {
		remoteService.WithConvention(convention)
	}
	if resourceName != "" {
		remoteService.WithResourceName(resourceName)
	}
	if pluralResourceName != "" {
		remoteService.WithPluralResourceName(pluralResourceName)
	}

	// Apply route overrides if provided
	if routesRaw, ok := cfg["routes"]; ok {
		if routes, ok := routesRaw.([]any); ok {
			for _, routeRaw := range routes {
				if routeMap, ok := routeRaw.(map[string]any); ok {
					name := utils.GetValueFromMap(routeMap, "name", "")
					method := utils.GetValueFromMap(routeMap, "method", "")
					path := utils.GetValueFromMap(routeMap, "path", "")

					if name != "" && path != "" {
						remoteService.WithRouteOverride(name, path)
					}
					if name != "" && method != "" {
						remoteService.WithMethodOverride(name, method)
					}
				}
			}
		}
	}

	return remoteService
}
