package loader

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/primadi/lokstra/core/service"
)

// flattenConfigs flattens nested config maps using dot notation
// Example: {"db": {"host": "localhost"}} => {"db.host": "localhost"}
func flattenConfigs(configs map[string]any) map[string]any {
	result := make(map[string]any)
	flattenConfigsRecursive("", configs, result)
	return result
}

func flattenConfigsRecursive(prefix string, configs map[string]any, result map[string]any) {
	for key, value := range configs {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		// If value is a map, recurse
		if nestedMap, ok := value.(map[string]any); ok {
			flattenConfigsRecursive(fullKey, nestedMap, result)
		} else {
			// Store leaf value
			result[fullKey] = value
		}
	}
}

// normalizeServerDefinitions converts server-level helper fields to a new app
// This allows shorthand syntax: addr/routers/published-services at server level
// for the common case of 1 server = 1 app
//
// Smart merging behavior:
//   - If helper has addr: Create new app and prepend to Apps array
//   - If helper has NO addr but has routers/published-services: Merge into first existing app
//   - If no existing apps: Create new app (even without addr - will fail validation later)
func normalizeServerDefinitions(config *schema.DeployConfig) {
	for _, depDef := range config.Deployments {
		for _, serverDef := range depDef.Servers {
			// Check if helper fields are used
			hasHelperFields := serverDef.HelperAddr != "" ||
				len(serverDef.HelperRouters) > 0 ||
				len(serverDef.HelperPublishedServices) > 0

			if !hasHelperFields {
				continue // No helper fields, skip
			}

			// Case 1: Helper has addr - create new app and prepend
			if serverDef.HelperAddr != "" {
				newApp := &schema.AppDefMap{
					Addr:              serverDef.HelperAddr,
					Routers:           serverDef.HelperRouters,
					PublishedServices: serverDef.HelperPublishedServices,
				}
				// PREPEND new app to Apps array (so it becomes first)
				serverDef.Apps = append([]*schema.AppDefMap{newApp}, serverDef.Apps...)
			} else if len(serverDef.Apps) > 0 {
				// Case 2: Helper has NO addr but has routers/published-services
				// Merge into first existing app
				firstApp := serverDef.Apps[0]

				// Merge routers (append, no duplicates)
				firstApp.Routers = mergeStringSlices(firstApp.Routers, serverDef.HelperRouters)

				// Merge published-services (append, no duplicates)
				firstApp.PublishedServices = mergeStringSlices(firstApp.PublishedServices, serverDef.HelperPublishedServices)
			} else {
				// Case 3: Helper has NO addr and NO existing apps
				// Create new app anyway (will fail validation if addr is required)
				newApp := &schema.AppDefMap{
					Addr:              serverDef.HelperAddr, // Empty
					Routers:           serverDef.HelperRouters,
					PublishedServices: serverDef.HelperPublishedServices,
				}
				serverDef.Apps = append(serverDef.Apps, newApp)
			}

			// Clear helper fields
			serverDef.HelperAddr = ""
			serverDef.HelperRouters = nil
			serverDef.HelperPublishedServices = nil
		}
	}
}

// mergeStringSlices merges two string slices without duplicates
func mergeStringSlices(a, b []string) []string {
	if len(b) == 0 {
		return a
	}
	if len(a) == 0 {
		return b
	}

	// Create a set from first slice
	seen := make(map[string]bool, len(a))
	for _, item := range a {
		seen[item] = true
	}

	// Append items from second slice that are not in first
	result := make([]string, len(a))
	copy(result, a)
	for _, item := range b {
		if !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}

	return result
}

// NormalizeInlineDefinitionsForServer performs lazy normalization of inline definitions
// for a specific deployment and server. This is called just before running the server.
//
// Normalization strategy:
//   - Deployment-level inline: {deployment}.{name}
//   - Server-level inline: {deployment}.{server}.{name}
//
// This function ONLY updates the config structure (moves inline to global definitions).
// The actual registration happens in the normal flow via LoadAndBuild logic.
func NormalizeInlineDefinitionsForServer(
	config *schema.DeployConfig,
	deploymentName, serverName string,
) error {
	depDef, ok := config.Deployments[deploymentName]
	if !ok {
		return fmt.Errorf("deployment %s not found", deploymentName)
	}

	serverDef, ok := depDef.Servers[serverName]
	if !ok {
		return fmt.Errorf("server %s not found in deployment %s", serverName, deploymentName)
	}

	// Initialize global maps if nil
	if config.MiddlewareDefinitions == nil {
		config.MiddlewareDefinitions = make(map[string]*schema.MiddlewareDef)
	}
	if config.ServiceDefinitions == nil {
		config.ServiceDefinitions = make(map[string]*schema.ServiceDef)
	}
	if config.RouterDefinitions == nil {
		config.RouterDefinitions = make(map[string]*schema.RouterDef)
	}
	if config.ExternalServiceDefinitions == nil {
		config.ExternalServiceDefinitions = make(map[string]*schema.RemoteServiceSimple)
	}

	// Process deployment-level inline definitions
	// Move to global with normalized names
	for name, mwDef := range depDef.InlineMiddlewares {
		normalizedName := deploymentName + "." + name
		config.MiddlewareDefinitions[normalizedName] = mwDef
	}

	for name, svcDef := range depDef.InlineServices {
		normalizedName := deploymentName + "." + name
		config.ServiceDefinitions[normalizedName] = svcDef
	}

	for name, rtrDef := range depDef.InlineRouters {
		normalizedName := deploymentName + "." + name
		config.RouterDefinitions[normalizedName] = rtrDef
	}

	for name, extDef := range depDef.InlineExternalServices {
		normalizedName := deploymentName + "." + name
		config.ExternalServiceDefinitions[normalizedName] = extDef
	}

	// Process server-level inline definitions
	// Move to global with normalized names (server-level overrides deployment-level if same name)
	for name, mwDef := range serverDef.InlineMiddlewares {
		normalizedName := deploymentName + "." + serverName + "." + name
		config.MiddlewareDefinitions[normalizedName] = mwDef
	}

	for name, svcDef := range serverDef.InlineServices {
		normalizedName := deploymentName + "." + serverName + "." + name
		config.ServiceDefinitions[normalizedName] = svcDef
	}

	for name, rtrDef := range serverDef.InlineRouters {
		normalizedName := deploymentName + "." + serverName + "." + name
		config.RouterDefinitions[normalizedName] = rtrDef
	}

	for name, extDef := range serverDef.InlineExternalServices {
		normalizedName := deploymentName + "." + serverName + "." + name
		config.ExternalServiceDefinitions[normalizedName] = extDef
	}

	// Build renaming map for reference resolution BEFORE clearing inline definitions
	// Maps short names to normalized names for this deployment+server context
	renamings := make(map[string]string)

	// Add deployment-level renamings
	for name := range depDef.InlineMiddlewares {
		renamings[name] = deploymentName + "." + name
	}
	for name := range depDef.InlineServices {
		renamings[name] = deploymentName + "." + name
	}
	for name := range depDef.InlineRouters {
		renamings[name] = deploymentName + "." + name
	}
	for name := range depDef.InlineExternalServices {
		renamings[name] = deploymentName + "." + name
	}

	// Add server-level renamings (these override deployment-level if same name)
	for name := range serverDef.InlineMiddlewares {
		renamings[name] = deploymentName + "." + serverName + "." + name
	}
	for name := range serverDef.InlineServices {
		renamings[name] = deploymentName + "." + serverName + "." + name
	}
	for name := range serverDef.InlineRouters {
		renamings[name] = deploymentName + "." + serverName + "." + name
	}
	for name := range serverDef.InlineExternalServices {
		renamings[name] = deploymentName + "." + serverName + "." + name
	}

	// Clear inline definitions (they're now in global)
	depDef.InlineMiddlewares = nil
	depDef.InlineServices = nil
	depDef.InlineRouters = nil
	depDef.InlineExternalServices = nil

	serverDef.InlineMiddlewares = nil
	serverDef.InlineServices = nil
	serverDef.InlineRouters = nil
	serverDef.InlineExternalServices = nil

	// Update references in ALL service definitions (global + normalized inline)
	for _, svcDef := range config.ServiceDefinitions {
		// Update depends-on references
		for i, dep := range svcDef.DependsOn {
			// Parse "paramName:serviceName" or just "serviceName"
			parts := strings.SplitN(dep, ":", 2)
			if len(parts) == 2 {
				// Format: "paramName:serviceName"
				paramName := parts[0]
				serviceName := parts[1]
				if normalizedName, found := renamings[serviceName]; found {
					svcDef.DependsOn[i] = paramName + ":" + normalizedName
				}
			} else {
				// Format: "serviceName"
				serviceName := parts[0]
				if normalizedName, found := renamings[serviceName]; found {
					svcDef.DependsOn[i] = normalizedName
				}
			}
		}

		// Update middleware references in service.router (if exists)
		if svcDef.Router != nil {
			for i, mwName := range svcDef.Router.Middlewares {
				if normalizedName, found := renamings[mwName]; found {
					svcDef.Router.Middlewares[i] = normalizedName
				}
			}

			// Update middleware references in custom routes
			for _, customRoute := range svcDef.Router.Custom {
				for i, mwName := range customRoute.Middlewares {
					if normalizedName, found := renamings[mwName]; found {
						customRoute.Middlewares[i] = normalizedName
					}
				}
			}
		}
	}

	// Update references in ALL router definitions (global + normalized inline)
	for _, rtrDef := range config.RouterDefinitions {
		// Update middleware references
		for i, mwName := range rtrDef.Middlewares {
			if normalizedName, found := renamings[mwName]; found {
				rtrDef.Middlewares[i] = normalizedName
			}
		}

		// Update custom route middleware references
		for _, customRoute := range rtrDef.Custom {
			for i, mwName := range customRoute.Middlewares {
				if normalizedName, found := renamings[mwName]; found {
					customRoute.Middlewares[i] = normalizedName
				}
			}
		}
	}

	// Update references in ALL external service definitions (global + normalized inline)
	for _, extSvc := range config.ExternalServiceDefinitions {
		if extSvc.Router == nil {
			continue
		}

		// Update middleware references
		for i, mwName := range extSvc.Router.Middlewares {
			if normalizedName, found := renamings[mwName]; found {
				extSvc.Router.Middlewares[i] = normalizedName
			}
		}

		// Update custom route middleware references
		for _, customRoute := range extSvc.Router.Custom {
			for i, mwName := range customRoute.Middlewares {
				if normalizedName, found := renamings[mwName]; found {
					customRoute.Middlewares[i] = normalizedName
				}
			}
		}
	}

	// Update published-services references in apps
	// This is CRITICAL - apps reference services by name
	for _, appDef := range serverDef.Apps {
		for i, svcName := range appDef.PublishedServices {
			if normalizedName, found := renamings[svcName]; found {
				appDef.PublishedServices[i] = normalizedName
			}
		}

		// Also update router references (in case they reference inline routers)
		for i, rtrName := range appDef.Routers {
			if normalizedName, found := renamings[rtrName]; found {
				appDef.Routers[i] = normalizedName
			}
		}
	}

	// CRITICAL: Update server topology service names to normalized names
	// This is needed because RegisterDefinitionsForRuntime uses serverTopo.Services
	// to lookup services in config.ServiceDefinitions (which now has normalized names)
	compositeKey := deploymentName + "." + serverName
	serverTopo, ok := deploy.Global().GetServerTopology(compositeKey)
	if ok {
		// Update service names in topology
		for i, svcName := range serverTopo.Services {
			if normalizedName, found := renamings[svcName]; found {
				serverTopo.Services[i] = normalizedName
			}
		}

		// Update remote services map keys
		if len(serverTopo.RemoteServices) > 0 {
			newRemoteServices := make(map[string]string)
			for svcName, remoteURL := range serverTopo.RemoteServices {
				if normalizedName, found := renamings[svcName]; found {
					newRemoteServices[normalizedName] = remoteURL
				} else {
					newRemoteServices[svcName] = remoteURL
				}
			}
			serverTopo.RemoteServices = newRemoteServices
		}

		// CRITICAL: Update router names in app topologies
		// Routers auto-generated from published-services need to be updated to normalized names
		// Example: "user-service-router" -> "development.user-service-router"
		for _, appTopo := range serverTopo.Apps {
			for i, routerName := range appTopo.Routers {
				// Check if this router name was auto-generated from a service name
				// Format: "{service-name}-router"
				serviceName := strings.TrimSuffix(routerName, "-router")
				if normalizedServiceName, found := renamings[serviceName]; found {
					// Service was renamed, so router should be renamed too
					normalizedRouterName := normalizedServiceName + "-router"
					appTopo.Routers[i] = normalizedRouterName
				}
			}
		}
	}

	return nil
}

// StoreDefinitionsToRegistry stores all definitions to the global registry WITHOUT runtime registration
// This is called during LoadAndBuild to prepare definitions for later lazy registration
// Runtime registration happens in RunCurrentServer after normalization
func StoreDefinitionsToRegistry(registry *deploy.GlobalRegistry, config *schema.DeployConfig) error {
	// Store configs as definitions (flattened for nested support)
	flatConfigs := flattenConfigs(config.Configs)
	for name, value := range flatConfigs {
		registry.DefineConfig(&schema.ConfigDef{
			Name:  name,
			Value: value,
		})
	}

	// Store middleware definitions to registry (no runtime registration yet)
	// Middlewares will be registered in RegisterDefinitionsForRuntime
	// For now, we don't need to store them - they're in config.MiddlewareDefinitions

	// Store service definitions to registry as deferred (no runtime registration yet)
	// This allows GetDeferredServiceDef to work during RegisterDefinitionsForRuntime
	for name, svc := range config.ServiceDefinitions {
		svc.Name = name
		// Convert ServiceDef to config map for registerDeferredService
		svcConfig := svc.Config
		if svcConfig == nil {
			svcConfig = make(map[string]any)
		}

		// Add depends-on to config (required by registerDeferredService)
		if len(svc.DependsOn) > 0 {
			svcConfig["depends-on"] = svc.DependsOn
		}

		// Call internal method via reflection or expose public method
		// For now, we'll use a workaround: store directly via RegisterLazyService with nil factory
		// Actually, let's use the fact that registry has serviceDefs field
		// But that's private... so we need to expose a public method
		//
		// WORKAROUND: Call RegisterLazyService with mode=Skip so it doesn't error on duplicates
		// But we DON'T want to register yet, just store definition
		//
		// BETTER: Don't store now - GetDeferredServiceDef will be called in RegisterDefinitionsForRuntime
		// But the problem is GetDeferredServiceDef reads from serviceDefs which is populated by registerDeferredService
		// And registerDeferredService is private!
		//
		// SOLUTION: We need to expose a public method in registry to store deferred service definitions
		// OR: We can skip storing here and directly use config.ServiceDefinitions in RegisterDefinitionsForRuntime
	}

	// Store router definitions to registry (deferred)
	for name, rtr := range config.RouterDefinitions {
		registry.DefineRouter(name, rtr)
	}

	// External service definitions don't need to be stored separately
	// They'll be processed in RegisterDefinitionsForRuntime

	return nil
}

// collectAllServiceDependencies recursively collects all services and their dependencies
func collectAllServiceDependencies(config *schema.DeployConfig, publishedServices []string) []string {
	visited := make(map[string]bool)
	result := []string{}

	var collectDeps func(serviceName string)
	collectDeps = func(serviceName string) {
		if visited[serviceName] {
			return
		}
		visited[serviceName] = true

		// Get service definition
		svc, exists := config.ServiceDefinitions[serviceName]
		if !exists {
			return
		}

		// Add this service to result
		result = append(result, serviceName)

		// Recursively collect dependencies
		for _, depStr := range svc.DependsOn {
			// Parse "paramName:serviceName" or just "serviceName"
			parts := strings.SplitN(depStr, ":", 2)
			depServiceName := depStr
			if len(parts) == 2 {
				depServiceName = parts[1]
			}
			collectDeps(depServiceName)
		}
	}

	// Start with published services
	for _, serviceName := range publishedServices {
		collectDeps(serviceName)
	}

	return result
}

// RegisterDefinitionsForRuntime performs runtime registration of definitions
// This is called in RunCurrentServer AFTER normalization
// It registers middlewares, services (with remote/local logic), and auto-generates routers for published services
func RegisterDefinitionsForRuntime(registry *deploy.GlobalRegistry, config *schema.DeployConfig, deploymentName, serverName string, serverTopo *deploy.ServerTopology) error {
	// Get the current server's apps to know which services are published
	depDef, ok := config.Deployments[deploymentName]
	if !ok {
		return fmt.Errorf("deployment %s not found", deploymentName)
	}

	serverDef, ok := depDef.Servers[serverName]
	if !ok {
		return fmt.Errorf("server %s not found in deployment %s", serverName, deploymentName)
	}

	// Register middlewares
	for name, mw := range config.MiddlewareDefinitions {
		mw.Name = name
		registry.RegisterMiddlewareName(name, mw.Type, mw.Config)
	}

	// Auto-create service wrappers for external services with factory type
	externalServices := config.ExternalServiceDefinitions
	for name, extSvc := range externalServices {
		if extSvc.Type != "" {
			// Check if service definition already exists (manual override)
			if _, exists := config.ServiceDefinitions[name]; exists {
				continue
			}

			// Auto-create service definition
			autoServiceDef := &schema.ServiceDef{
				Name:      name,
				Type:      extSvc.Type,
				DependsOn: nil,
				Config:    make(map[string]any),
			}

			// Copy external service config
			if extSvc.Config != nil {
				maps.Copy(autoServiceDef.Config, extSvc.Config)
			}

			// Add URL to config
			autoServiceDef.Config["url"] = extSvc.URL

			// Copy router definition from external service to service definition
			if extSvc.Router != nil {
				autoServiceDef.Router = extSvc.Router

				// Also add to config for backward compatibility
				if extSvc.Router.Resource != "" {
					autoServiceDef.Config["resource"] = extSvc.Router.Resource
				}
				if extSvc.Router.ResourcePlural != "" {
					autoServiceDef.Config["resource_plural"] = extSvc.Router.ResourcePlural
				}
				if extSvc.Router.Convention != "" {
					autoServiceDef.Config["convention"] = extSvc.Router.Convention
				}
				if extSvc.Router.PathPrefix != "" {
					autoServiceDef.Config["path_prefix"] = extSvc.Router.PathPrefix
				}
				if len(extSvc.Router.Middlewares) > 0 {
					autoServiceDef.Config["middlewares"] = extSvc.Router.Middlewares
				}
				if len(extSvc.Router.Hidden) > 0 {
					autoServiceDef.Config["hidden"] = extSvc.Router.Hidden
				}
				if len(extSvc.Router.Custom) > 0 {
					autoServiceDef.Config["custom"] = extSvc.Router.Custom
				}
			}

			config.ServiceDefinitions[name] = autoServiceDef
		}
	}

	// Merge service definitions from registry (RegisterLazyService) into config
	// This allows services registered via code to be available in config.ServiceDefinitions
	registry.MergeRegistryServicesToConfig(config)

	// Collect all services needed for this server (published services + their dependencies)
	servicesToRegister := collectAllServiceDependencies(config, serverTopo.Services)

	// Register service definitions with remote/local logic
	// Iterate through all services (published + dependencies)
	for _, serviceName := range servicesToRegister {
		svc, exists := config.ServiceDefinitions[serviceName]
		if !exists {
			return fmt.Errorf("service %s in topology not found in service definitions", serviceName)
		}

		svc.Name = serviceName

		// Check if this is a remote service
		remoteURL, isRemote := serverTopo.RemoteServices[serviceName]

		if isRemote && remoteURL != "" {
			// Register REMOTE service
			registry.AutoRegisterRemoteService(serviceName, svc, remoteURL)
		} else {
			// Register LOCAL service with dependency resolution
			// Convert DependsOn to deps map
			deps := make(map[string]string)
			for _, depStr := range svc.DependsOn {
				// Parse "paramName:serviceName" or just "serviceName"
				parts := strings.SplitN(depStr, ":", 2)
				if len(parts) == 2 {
					deps[parts[0]] = parts[1]
				} else {
					deps[depStr] = depStr
				}
			}

			// Get service type factory (LOCAL)
			serviceType := svc.Type
			factory := registry.GetServiceFactory(serviceType, true) // true = local factory
			if factory == nil {
				return fmt.Errorf("service factory %s (local) not registered for service %s", serviceType, serviceName)
			}

			// Register as lazy service with wrapper factory
			// Use Skip mode to allow idempotent calls
			registry.RegisterLazyServiceWithDeps(serviceName, func(resolvedDeps, cfg map[string]any) any {
				// Factory expects lazy loaders (service.Cached), so wrap resolved deps
				lazyDeps := make(map[string]any)
				for key, depSvc := range resolvedDeps {
					depSvcCopy := depSvc // Capture for closure
					lazyDeps[key] = service.LazyLoadWith(func() any {
						return depSvcCopy
					})
				}

				// Call original factory
				return factory(lazyDeps, cfg)
			}, deps, svc.Config, deploy.WithRegistrationMode(deploy.LazyServiceSkip))
		}
	}

	// Collect published services from current server's apps
	publishedServicesMap := make(map[string]bool)
	for _, appDef := range serverDef.Apps {
		for _, serviceName := range appDef.PublishedServices {
			publishedServicesMap[serviceName] = true
		}
	}

	// Auto-generate router definitions for published services
	// Also update Apps.Routers to use normalized router names
	// Priority: service.router > router-definitions > metadata > auto-generate
	routerRenamings := make(map[string]string) // old router name -> new router name

	for serviceName := range publishedServicesMap {
		routerName := serviceName + "-router"

		// Get service definition from config (already normalized)
		serviceDef, exists := config.ServiceDefinitions[serviceName]
		if !exists {
			return fmt.Errorf("published service '%s' not found in service-definitions after normalization", serviceName)
		}

		metadata := registry.GetServiceMetadata(serviceDef.Type)

		var resourceName, resourcePlural, convention string
		var pathPrefix string
		var pathRewrites []schema.PathRewriteDef
		var middlewares []string
		var hidden []string
		var custom []schema.RouteDef

		// Priority 1: Check if service has embedded router definition
		if serviceDef.Router != nil {
			resourceName = serviceDef.Router.Resource
			resourcePlural = serviceDef.Router.ResourcePlural
			convention = serviceDef.Router.Convention
			pathPrefix = serviceDef.Router.PathPrefix
			pathRewrites = serviceDef.Router.PathRewrites
			middlewares = serviceDef.Router.Middlewares
			hidden = serviceDef.Router.Hidden
			custom = serviceDef.Router.Custom
		}

		// Priority 2: Check if router manually defined in router-definitions (override/standalone)
		if yamlRouter, exists := config.RouterDefinitions[routerName]; exists {
			// Override only if not set in service.router
			if resourceName == "" {
				resourceName = yamlRouter.Resource
			}
			if resourcePlural == "" {
				resourcePlural = yamlRouter.ResourcePlural
			}
			if convention == "" {
				convention = yamlRouter.Convention
			}
			if pathPrefix == "" {
				pathPrefix = yamlRouter.PathPrefix
			}
			if len(pathRewrites) == 0 {
				pathRewrites = yamlRouter.PathRewrites
			}
			if len(middlewares) == 0 {
				middlewares = yamlRouter.Middlewares
			}
			if len(hidden) == 0 {
				hidden = yamlRouter.Hidden
			}
			if len(custom) == 0 {
				custom = yamlRouter.Custom
			}
		}

		// Priority 3: Fallback to metadata from RegisterServiceType
		if resourceName == "" && metadata != nil && metadata.Resource != "" {
			resourceName = metadata.Resource
			resourcePlural = metadata.ResourcePlural
			convention = metadata.Convention
		}

		// Priority 4: Final fallback - auto-generate from service name
		if resourceName == "" {
			// Extract resource name from normalized service name
			// Examples:
			//   "development.user-service" -> "user"
			//   "development.dev-server.product-service" -> "product"
			//   "order-service" -> "order"
			parts := strings.Split(serviceName, ".")
			lastPart := parts[len(parts)-1]
			resourceName = strings.TrimSuffix(lastPart, "-service")
			resourcePlural = resourceName + "s"
			convention = "rest"
		}

		// Define auto-generated router
		autoRouter := &schema.RouterDef{
			Convention:     convention,
			Resource:       resourceName,
			ResourcePlural: resourcePlural,
			PathPrefix:     pathPrefix,
			PathRewrites:   pathRewrites,
			Middlewares:    middlewares,
			Hidden:         hidden,
			Custom:         custom,
		}

		// Store to config.RouterDefinitions so it's available for later lookup
		config.RouterDefinitions[routerName] = autoRouter

		// Also define in registry
		if registry.GetRouterDef(routerName) == nil {
			registry.DefineRouter(routerName, autoRouter)
		}

		// Track router renamings for Apps.Routers update
		// Old name: extract from service name without prefix
		// Example: "development.user-service" -> "user-service-router"
		serviceShortName := strings.Split(serviceName, ".")[len(strings.Split(serviceName, "."))-1]
		oldRouterName := serviceShortName + "-router"
		if oldRouterName != routerName {
			routerRenamings[oldRouterName] = routerName
		}
	}

	// Update Apps.Routers to use normalized router names
	for _, appDef := range serverDef.Apps {
		for i, routerName := range appDef.Routers {
			if normalizedName, found := routerRenamings[routerName]; found {
				appDef.Routers[i] = normalizedName
			}
		}
	}

	// Register standalone router definitions
	for routerName, routerDef := range config.RouterDefinitions {
		if registry.GetRouterDef(routerName) != nil {
			continue
		}
		registry.DefineRouter(routerName, routerDef)
	}

	return nil
}

// LoadAndBuild loads config and builds ALL deployments into Global registry
// Returns error only - deployments are stored in deploy.Global()
func LoadAndBuild(configPaths []string) error {
	config, err := LoadConfig(configPaths...)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	registry := deploy.Global()

	// Store original config for inline definitions normalization
	registry.StoreDeployConfig(config)

	// Normalize server definitions (convert helper fields to apps)
	normalizeServerDefinitions(config)

	// Store definitions to registry (NO runtime registration, just store data)
	// Runtime registration will happen in RunCurrentServer
	if err := StoreDefinitionsToRegistry(registry, config); err != nil {
		return fmt.Errorf("failed to store definitions: %w", err)
	}

	// Build ALL deployments (2-Layer Architecture: YAML -> Topology only)
	for deploymentName, depDef := range config.Deployments {
		// Build service location registry (service-name → base-url)
		// This maps published services to their server URLs for remote service resolution
		serviceLocations := make(map[string]string)
		for _, serverDef := range depDef.Servers {
			for _, appDef := range serverDef.Apps {
				for _, serviceName := range appDef.PublishedServices {
					// Build full URL: base-url + addr
					// base-url should be protocol + host (e.g., "http://localhost")
					// addr should be port (e.g., ":4000")
					fullURL := serverDef.BaseURL + appDef.Addr
					serviceLocations[serviceName] = fullURL
				}
			}
		}

		// Create and store topology (NEW 2-Layer Architecture)
		deployTopo := &deploy.DeploymentTopology{
			Name:            deploymentName,
			ConfigOverrides: make(map[string]any),
			Servers:         make(map[string]*deploy.ServerTopology),
		}

		// Copy config overrides
		for key, value := range depDef.ConfigOverrides {
			deployTopo.ConfigOverrides[key] = value
		}

		// Build server topologies
		for serverName, serverDef := range depDef.Servers {
			serverTopo := &deploy.ServerTopology{
				Name:           serverName,
				DeploymentName: deploymentName,
				BaseURL:        serverDef.BaseURL,
				Services:       make([]string, 0),
				RemoteServices: make(map[string]string),
				Apps:           make([]*deploy.AppTopology, 0, len(serverDef.Apps)),
			}

			// Collect SERVER-LEVEL services (published services only)
			// Dependencies are auto-detected from service-definitions, not explicitly listed
			serviceMap := make(map[string]bool)

			// Add published services from all apps (these are local services on this server)
			for _, appDef := range serverDef.Apps {
				for _, svcName := range appDef.PublishedServices {
					serviceMap[svcName] = true
				}
			}

			// Convert to slice
			for svcName := range serviceMap {
				serverTopo.Services = append(serverTopo.Services, svcName)
			}

			// Build RemoteServices map (services published on OTHER servers in this deployment)
			for svcName, svcURL := range serviceLocations {
				// Skip if service is local to this server
				if serviceMap[svcName] {
					continue
				}
				// Add as remote service
				serverTopo.RemoteServices[svcName] = svcURL
			}

			// Add external services to RemoteServices map
			// External services are ALWAYS remote (never local)
			for extSvcName, extSvc := range config.ExternalServiceDefinitions {
				if extSvc.URL != "" {
					serverTopo.RemoteServices[extSvcName] = extSvc.URL
				}
			}

			// Build app topologies (only addr + routers, NO services)
			for _, appDef := range serverDef.Apps {
				appTopo := &deploy.AppTopology{
					Addr:    appDef.Addr,
					Routers: make([]string, 0, len(appDef.Routers)+len(appDef.PublishedServices)),
				}

				// Collect routers
				appTopo.Routers = append(appTopo.Routers, appDef.Routers...)
				// Auto-generated routers from published services
				for _, serviceName := range appDef.PublishedServices {
					routerName := serviceName + "-router"
					appTopo.Routers = append(appTopo.Routers, routerName)
				}

				serverTopo.Apps = append(serverTopo.Apps, appTopo)
			}

			deployTopo.Servers[serverName] = serverTopo
		}

		// Store topology in global registry
		registry.StoreDeploymentTopology(deployTopo)
	}

	// Resolve ALL config values throughout the entire configuration
	// This includes configs, deployment values, service configs, etc.
	if err := registry.ResolveConfigs(); err != nil {
		return fmt.Errorf("failed to resolve configs: %w", err)
	}

	return nil
}

// LoadAndBuildFromDir loads all YAML files from a directory and builds ALL deployments
func LoadAndBuildFromDir(dirPath string) error {
	// Scan directory for YAML files
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)
		if ext == ".yaml" || ext == ".yml" {
			paths = append(paths, filepath.Join(dirPath, name))
		}
	}

	if len(paths) == 0 {
		return fmt.Errorf("no YAML files found in directory: %s", dirPath)
	}

	// Delegate to LoadAndBuild
	return LoadAndBuild(paths)
}
