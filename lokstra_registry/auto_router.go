package lokstra_registry

import (
	"fmt"
	"strings"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/router/autogen"
	"github.com/primadi/lokstra/core/router/convention"
	"github.com/primadi/lokstra/core/service"
)

// BuildRouterFromDefinition creates a router instance from a RouterDef in the global registry
// This is used for auto-generated routers from published-services
//
// Metadata Resolution Priority (3 sources with fallback):
//  1. YAML config (router-overrides)          ← HIGHEST - Runtime override per deployment
//  2. XXXRemote struct (RemoteServiceMeta)    ← MEDIUM - Service-level defaults in code
//  3. RegisterServiceType options             ← LOWEST - Framework-level defaults
//
// This allows flexibility:
//   - Simple case: Just metadata in XXXRemote struct (recommended)
//   - Override per deployment: Add YAML config
//   - Framework defaults: Use RegisterServiceType options
//
// Resolution order:
//  1. Try to instantiate remote service factory and check for RemoteServiceMeta
//  2. Fall back to RouterDef from config (convention, resource, overrides)
//  3. YAML overrides are MERGED on top (highest priority)
func BuildRouterFromDefinition(routerName string) (router.Router, error) {
	// Get router definition from global registry
	routerDef := deploy.Global().GetRouterDef(routerName)
	if routerDef == nil {
		return nil, fmt.Errorf("router definition '%s' not found in global registry", routerName)
	}

	// Get service instance from global registry (already registered as lazy in SetCurrentServer)
	svc, ok := deploy.Global().GetServiceAny(routerDef.Service)
	if !ok {
		return nil, fmt.Errorf("service '%s' not found for router '%s'", routerDef.Service, routerName)
	}

	// Build conversion rule and override
	var rule autogen.ConversionRule
	var override autogen.RouteOverride
	var metadataFound bool

	// Strategy 1: Check metadata from RegisterServiceType options (from service type)
	serviceDef := deploy.Global().GetServiceDef(routerDef.Service)
	if serviceDef != nil {
		metadata := deploy.Global().GetServiceMetadata(serviceDef.Type)
		if metadata != nil && metadata.Resource != "" {
			// Use metadata from RegisterServiceType
			rule = autogen.ConversionRule{
				Convention:     convention.ConventionType(metadata.Convention),
				Resource:       metadata.Resource,
				ResourcePlural: metadata.ResourcePlural,
			}
			// TODO: Convert metadata.RouteOverrides to autogen.RouteOverride
			override = autogen.RouteOverride{
				PathPrefix: metadata.PathPrefix,
				Hidden:     metadata.HiddenMethods,
			}

			// Config can still override
			if routerDef.Convention != "" {
				rule.Convention = convention.ConventionType(routerDef.Convention)
			}
			if routerDef.Resource != "" {
				rule.Resource = routerDef.Resource
			}
			if routerDef.ResourcePlural != "" {
				rule.ResourcePlural = routerDef.ResourcePlural
			}

			metadataFound = true
		}
	}

	// Strategy 2: Try to get metadata from service instance (ServiceMeta)
	// This works for both local and remote services!
	if !metadataFound {
		serviceMeta := tryGetServiceMeta(svc)
		if serviceMeta != nil {
			// Service provides metadata - use it!
			resource, plural := serviceMeta.GetResourceName()
			conventionName := serviceMeta.GetConventionName()
			serviceOverride := serviceMeta.GetRouteOverride()

			rule = autogen.ConversionRule{
				Convention:     convention.ConventionType(conventionName),
				Resource:       resource,
				ResourcePlural: plural,
			}
			override = serviceOverride

			// Config can still override service metadata if explicitly set
			if routerDef.Convention != "" {
				rule.Convention = convention.ConventionType(routerDef.Convention)
			}
			if routerDef.Resource != "" {
				rule.Resource = routerDef.Resource
			}
			if routerDef.ResourcePlural != "" {
				rule.ResourcePlural = routerDef.ResourcePlural
			}

			metadataFound = true
		}
	}

	// Strategy 3: Fall back to config-based metadata or auto-generate from service name
	if !metadataFound {
		// Auto-generate resource name from service name if not in config
		resource := routerDef.Resource
		resourcePlural := routerDef.ResourcePlural
		conventionType := convention.REST // Default convention

		if resource == "" {
			// Auto-generate from service name: "order-service" -> "order"
			resource = strings.TrimSuffix(routerDef.Service, "-service")
		}
		if resourcePlural == "" {
			// Simple pluralization
			resourcePlural = resource + "s"
		}
		if routerDef.Convention != "" {
			conventionType = convention.ConventionType(routerDef.Convention)
		}

		rule = autogen.ConversionRule{
			Convention:     conventionType,
			Resource:       resource,
			ResourcePlural: resourcePlural,
		}
		override = autogen.RouteOverride{}
	}

	// If overrides are specified in config, apply them
	if routerDef.Overrides != "" {
		overrideDef := deploy.Global().GetRouterOverride(routerDef.Overrides)
		if overrideDef == nil {
			return nil, fmt.Errorf("router override '%s' not found", routerDef.Overrides)
		}

		// Merge config override with service override
		if overrideDef.PathPrefix != "" {
			override.PathPrefix = overrideDef.PathPrefix
		}
		if len(overrideDef.Hidden) > 0 {
			override.Hidden = append(override.Hidden, overrideDef.Hidden...)
		}

		// Convert custom routes from config schema to autogen.Route
		if len(overrideDef.Custom) > 0 {
			if override.Custom == nil {
				override.Custom = make(map[string]autogen.Route)
			}
			for _, customRoute := range overrideDef.Custom {
				override.Custom[customRoute.Name] = autogen.Route{
					Method: customRoute.Method,
					Path:   customRoute.Path,
				}
			}
		}

		// TODO: Convert middlewares
	}

	// Create router from service using autogen
	r := autogen.NewFromService(svc, rule, override)

	return r, nil
}

// tryGetServiceMeta attempts to get ServiceMeta from a service instance
// This works for both local and remote services that implement ServiceMeta
func tryGetServiceMeta(svc any) service.ServiceMeta {
	// Check if service implements ServiceMeta interface
	if metaSvc, ok := svc.(service.ServiceMeta); ok {
		return metaSvc
	}
	return nil
}

// tryGetRemoteServiceMeta attempts to instantiate remote service and get metadata
// This creates a temporary remote service instance just to read metadata
// DEPRECATED: Use tryGetServiceMeta instead (works for both local and remote)
func tryGetRemoteServiceMeta(serviceName string) (result service.RemoteServiceMeta) {
	// Get service definition to find its type
	serviceDef := deploy.Global().GetServiceDef(serviceName)
	if serviceDef == nil {
		return nil
	}

	// Try to get remote service factory from registry
	remoteFactory := GetServiceFactory(serviceDef.Type, false)
	if remoteFactory == nil {
		return nil
	}

	// Create temporary remote service instance
	// We need to handle the case where factory expects proxy.Service
	// but we only want to read metadata. Recover from panics.
	defer func() {
		if r := recover(); r != nil {
			// Factory panicked (probably CastProxyService), return nil
			result = nil
		}
	}()

	// Try to create remote service instance
	// Most remote services expect config["remote"] to be a proxy.Service
	// but we don't have one. Pass nil and let factory handle it.
	remoteSvc := remoteFactory(map[string]any{}, map[string]any{
		"remote": nil, // Factory might panic on CastProxyService
	})

	// Check if it implements RemoteServiceMeta
	if metaSvc, ok := remoteSvc.(service.RemoteServiceMeta); ok {
		return metaSvc
	}

	return nil
}
