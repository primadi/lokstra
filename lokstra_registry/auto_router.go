package lokstra_registry

import (
	"fmt"
	"strings"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/router/autogen"
	"github.com/primadi/lokstra/core/router/convention"
)

// BuildRouterFromDefinition creates a router instance from a RouterDef in the global registry
// This is used for auto-generated routers from published-services
//
// Metadata Resolution Priority (2 sources with fallback):
//  1. YAML config (router-overrides)          ← HIGHEST - Runtime override per deployment
//  2. RegisterServiceType options             ← MEDIUM - Framework-level defaults
//  3. Auto-generate from service name         ← LOWEST - Fallback
//
// This provides clear DX:
//   - Default: Metadata in RegisterServiceType (recommended)
//   - Override: Add YAML config per deployment
//   - Fallback: Auto-generate from service name
func BuildRouterFromDefinition(routerName string) (router.Router, error) {
	// Get router definition from global registry
	routerDef := deploy.Global().GetRouterDef(routerName)
	if routerDef == nil {
		return nil, fmt.Errorf("router definition '%s' not found in global registry", routerName)
	}

	// Derive service name from router name
	// Router name format: "{service-name}-router"
	// Example: "user-service-router" → "user-service"
	serviceName := strings.TrimSuffix(routerName, "-router")

	// Get service instance from global registry (already registered as lazy in SetCurrentServer)
	svc, ok := deploy.Global().GetServiceAny(serviceName)
	if !ok {
		return nil, fmt.Errorf("service '%s' not found for router '%s'", serviceName, routerName)
	}

	// Build conversion rule and override
	var rule autogen.ConversionRule
	var override autogen.RouteOverride
	var metadataFound bool

	// Strategy 1: Check metadata from RegisterServiceType options
	serviceDef := deploy.Global().GetServiceDef(serviceName)
	if serviceDef != nil {
		metadata := deploy.Global().GetServiceMetadata(serviceDef.Type)
		if metadata != nil && metadata.Resource != "" {
			// Use metadata from RegisterServiceType
			rule = autogen.ConversionRule{
				Convention:     convention.ConventionType(metadata.Convention),
				Resource:       metadata.Resource,
				ResourcePlural: metadata.ResourcePlural,
			}

			// Convert RouteOverrides map to autogen.RouteOverride
			override = autogen.RouteOverride{
				PathPrefix: metadata.PathPrefix,
				Hidden:     metadata.HiddenMethods,
			}
			if len(metadata.RouteOverrides) > 0 {
				override.Custom = make(map[string]autogen.Route)
				for methodName, routeMetadata := range metadata.RouteOverrides {
					// RouteMetadata already contains Method, Path, and Middlewares
					override.Custom[methodName] = autogen.Route{
						Method: routeMetadata.Method,
						Path:   routeMetadata.Path,
					}
				}
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

	// Strategy 2: Fall back to auto-generate from service name
	if !metadataFound {
		// Auto-generate resource name from service name if not in config
		resource := routerDef.Resource
		resourcePlural := routerDef.ResourcePlural
		conventionType := convention.REST // Default convention

		if resource == "" {
			// Auto-generate from service name: "order-service" -> "order"
			resource = strings.TrimSuffix(serviceName, "-service")
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

	// Apply inline overrides from YAML config (if any)
	if routerDef.PathPrefix != "" {
		override.PathPrefix = routerDef.PathPrefix
	}
	if len(routerDef.Hidden) > 0 {
		override.Hidden = append(override.Hidden, routerDef.Hidden...)
	}

	// Convert custom routes from config schema to autogen.Route
	if len(routerDef.Custom) > 0 {
		if override.Custom == nil {
			override.Custom = make(map[string]autogen.Route)
		}
		for _, customRoute := range routerDef.Custom {
			override.Custom[customRoute.Name] = autogen.Route{
				Method: customRoute.Method,
				Path:   customRoute.Path,
			}
		}
	}

	// NOTE: Middlewares are applied in deployment.go after router creation
	// Router-level: via router.ApplyMiddlewares()
	// Route-level: via router.UpdateRoute()

	// Create router from service using autogen
	r := autogen.NewFromService(svc, rule, override)

	return r, nil
}
