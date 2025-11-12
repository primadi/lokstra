package lokstra_registry

import (
	"fmt"
	"strings"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/router/autogen"
	"github.com/primadi/lokstra/core/router/convention"
)

// logDebug is a helper to print debug logs only when enabled
func logDebug(format string, args ...any) {
	if deploy.GetLogLevel() >= deploy.LogLevelDebug {
		fmt.Printf("🐛 "+format+"\n", args...)
	}
}

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

	// Strategy 1: Try to get service type from YAML config first (to know which metadata to load)
	serviceDef := deploy.Global().GetDeferredServiceDef(serviceName)

	var serviceType string
	if serviceDef != nil {
		serviceType = serviceDef.Type
		logDebug("[auto_router] Found serviceDef for '%s': type=%s", serviceName, serviceType)
	} else {
		// Fallback: Try to infer type from service name
		// Convention: service name "user-service" → type "user-service-factory"
		serviceType = serviceName + "-factory"
		logDebug("[auto_router] No serviceDef for '%s', trying type=%s", serviceName, serviceType)
	}

	// Strategy 2: Get metadata from RegisterServiceType (CODE - baseline)
	metadata := deploy.Global().GetServiceMetadata(serviceType)
	if metadata != nil {
		logDebug("[auto_router] Found metadata for type '%s': Resource=%s, RouteOverrides=%d",
			serviceType, metadata.Resource, len(metadata.RouteOverrides))
		metadataFound = true

		// Infer resource name from service name if not provided
		resource := metadata.Resource
		if resource == "" {
			// Auto-generate from service name: "order-service" -> "order"
			resource = strings.TrimSuffix(serviceName, "-service")
			logDebug("[auto_router] Inferred resource from service name '%s': %s", serviceName, resource)
		}

		resourcePlural := metadata.ResourcePlural
		if resourcePlural == "" {
			// Simple pluralization
			resourcePlural = resource + "s"
			logDebug("[auto_router] Inferred resourcePlural: %s", resourcePlural)
		}

		conventionType := metadata.Convention
		if conventionType == "" {
			conventionType = "rest" // Default convention
		}

		// Use metadata from RegisterServiceType as baseline
		rule = autogen.ConversionRule{
			Convention:     convention.ConventionType(conventionType),
			Resource:       resource,
			ResourcePlural: resourcePlural,
		}

		// Convert RouteOverrides map to autogen.RouteOverride
		override = autogen.RouteOverride{
			PathPrefix: metadata.PathPrefix,
			Hidden:     metadata.HiddenMethods,
		}

		if len(metadata.RouteOverrides) > 0 {
			override.Custom = make(map[string]autogen.Route)
			for methodName, routeMetadata := range metadata.RouteOverrides {
				// Convert middleware names to []any
				var routeMiddlewares []any
				for _, mwName := range routeMetadata.Middlewares {
					routeMiddlewares = append(routeMiddlewares, mwName)
				}

				override.Custom[methodName] = autogen.Route{
					Method:      routeMetadata.Method,
					Path:        routeMetadata.Path,
					Middlewares: routeMiddlewares,
				}
				logDebug("[metadata CODE] Custom route '%s': method=%s, path=%s, middlewares=%d",
					methodName, routeMetadata.Method, routeMetadata.Path, len(routeMetadata.Middlewares))
			}
		}

		// Collect middleware names from RegisterServiceType metadata
		var middlewareNames []string
		if len(metadata.MiddlewareNames) > 0 {
			middlewareNames = append(middlewareNames, metadata.MiddlewareNames...)
		}

		// Create middleware instances
		if len(middlewareNames) > 0 {
			override.Middlewares = make([]any, 0, len(middlewareNames))
			for _, mwName := range middlewareNames {
				mw := deploy.Global().CreateMiddleware(mwName)
				if mw != nil {
					override.Middlewares = append(override.Middlewares, mw)
				} else {
					fmt.Printf("⚠️  Warning: Middleware '%s' not found for service '%s' (skipping)\n", mwName, serviceName)
				}
			}
		}
	}

	// Strategy 3: Override with YAML config (if exists and different from current)
	if routerDef.Convention != "" && string(rule.Convention) != routerDef.Convention {
		rule.Convention = convention.ConventionType(routerDef.Convention)
		logDebug("[YAML override] Convention: %s", routerDef.Convention)
	}
	if routerDef.Resource != "" && rule.Resource != routerDef.Resource {
		rule.Resource = routerDef.Resource
		logDebug("[YAML override] Resource: %s", routerDef.Resource)
	}
	if routerDef.ResourcePlural != "" && rule.ResourcePlural != routerDef.ResourcePlural {
		rule.ResourcePlural = routerDef.ResourcePlural
		logDebug("[YAML override] ResourcePlural: %s", routerDef.ResourcePlural)
	}

	// Strategy 4: Fallback to auto-generate from service name (if no metadata found)
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

	// Apply YAML PathPrefix override
	if routerDef.PathPrefix != "" {
		override.PathPrefix = routerDef.PathPrefix
		logDebug("[YAML override] PathPrefix: %s", routerDef.PathPrefix)
	}
	if len(routerDef.Hidden) > 0 {
		override.Hidden = append(override.Hidden, routerDef.Hidden...)
	}

	// Apply YAML custom routes override (replaces or adds to metadata custom routes)
	if len(routerDef.Custom) > 0 {
		logDebug("[YAML override] Custom routes count: %d", len(routerDef.Custom))
		if override.Custom == nil {
			override.Custom = make(map[string]autogen.Route)
		}
		for _, customRoute := range routerDef.Custom {
			override.Custom[customRoute.Name] = autogen.Route{
				Method: customRoute.Method,
				Path:   customRoute.Path,
			}
			logDebug("[YAML override] Custom route '%s': method=%s, path=%s",
				customRoute.Name, customRoute.Method, customRoute.Path)
		}
	}

	logDebug("[final] override.PathPrefix=%s, override.Custom count=%d",
		override.PathPrefix, len(override.Custom))

	// NOTE: Middlewares are applied in deployment.go after router creation
	// Router-level: via router.ApplyMiddlewares()
	// Route-level: via router.UpdateRoute()

	// Create router from service using autogen
	r := autogen.NewFromService(svc, rule, override)

	return r, nil
}
