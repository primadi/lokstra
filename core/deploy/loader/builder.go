package loader

import (
	"fmt"
	"strings"
	"time"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/serviceapi"
)

// Case-insensitive map lookup helpers
// Try lowercase first, then fallback to original case for backward compatibility

func getServiceDef(defs map[string]*schema.ServiceDef, name string) (*schema.ServiceDef, bool) {
	// Try lowercase first
	if svc, ok := defs[strings.ToLower(name)]; ok {
		return svc, true
	}
	// Fallback to original case
	svc, ok := defs[name]
	return svc, ok
}

// func getMiddlewareDef(defs map[string]*schema.MiddlewareDef, name string) (*schema.MiddlewareDef, bool) {
// 	// Try lowercase first
// 	if mw, ok := defs[strings.ToLower(name)]; ok {
// 		return mw, true
// 	}
// 	// Fallback to original case
// 	mw, ok := defs[name]
// 	return mw, ok
// }

func getRouterDef(defs map[string]*schema.RouterDef, name string) (*schema.RouterDef, bool) {
	// Try lowercase first
	if rtr, ok := defs[strings.ToLower(name)]; ok {
		return rtr, true
	}
	// Fallback to original case
	rtr, ok := defs[name]
	return rtr, ok
}

func getDeploymentDef(config *schema.DeployConfig, name string) (*schema.DeploymentDefMap, bool) {
	// Try lowercase first
	if dep, ok := config.Deployments[strings.ToLower(name)]; ok {
		return dep, true
	}
	// Fallback to original case
	dep, ok := config.Deployments[name]
	return dep, ok
}

func getServerDef(deployment *schema.DeploymentDefMap, name string) (*schema.ServerDefMap, bool) {
	// Try lowercase first
	if srv, ok := deployment.Servers[strings.ToLower(name)]; ok {
		return srv, true
	}
	// Fallback to original case
	srv, ok := deployment.Servers[name]
	return srv, ok
}

// flattenAndStoreConfigs flattens configs and stores them to registry.resolvedConfigs
// This populates the config registry that GetConfig() reads from
func flattenAndStoreConfigs(registry *deploy.GlobalRegistry, configs map[string]any, prefix string) {
	for key, value := range configs {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		// If value is a map, recurse AND store the map itself
		if nestedMap, ok := value.(map[string]any); ok {
			// Store the map at this level (for GetConfig[map[string]any]("db"))
			registry.SetConfig(fullKey, nestedMap)
			// Recurse to flatten nested values
			flattenAndStoreConfigs(registry, nestedMap, fullKey)
		} else {
			// Store leaf value
			registry.SetConfig(fullKey, value)
		}
	}
}

// normalizeShorthandServers converts top-level 'servers' field to 'default' deployment
// This provides a shorthand syntax for single-deployment configs:
//
//	servers:
//	  api:
//	    base-url: http://localhost
//	    addr: ":8080"
//	    routers: [email-router]
//
// Becomes:
//
//	deployments:
//	  default:
//	    servers:
//	      api:
//	        base-url: http://localhost
//	        addr: ":8080"
//	        routers: [email-router]
//
// This makes current server "api" equivalent to "default.api"
func normalizeShorthandServers(config *schema.DeployConfig) {
	// Skip if no top-level servers defined
	if len(config.Servers) == 0 {
		return
	}

	// Initialize Deployments if needed
	if config.Deployments == nil {
		config.Deployments = make(map[string]*schema.DeploymentDefMap)
	}

	// Check if 'default' deployment already exists
	if _, exists := config.Deployments["default"]; exists {
		// Merge servers into existing 'default' deployment
		for serverName, serverDef := range config.Servers {
			config.Deployments["default"].Servers[serverName] = serverDef
		}
	} else {
		// Create new 'default' deployment with the servers
		config.Deployments["default"] = &schema.DeploymentDefMap{
			ConfigOverrides: make(map[string]any),
			Servers:         config.Servers,
		}
	}

	// Clear top-level servers (moved to deployment)
	config.Servers = nil
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
	// Case-insensitive deployment lookup
	depDef, ok := getDeploymentDef(config, deploymentName)
	if !ok {
		return fmt.Errorf("deployment %s not found", deploymentName)
	}

	// Case-insensitive server lookup
	serverDef, ok := getServerDef(depDef, serverName)
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

	// Process deployment-level inline definitions
	// Move to global with normalized names (lowercase for case-insensitive lookup)
	for name, mwDef := range depDef.InlineMiddlewares {
		normalizedName := strings.ToLower(deploymentName + "." + name)
		config.MiddlewareDefinitions[normalizedName] = mwDef
	}

	for name, svcDef := range depDef.InlineServices {
		normalizedName := strings.ToLower(deploymentName + "." + name)
		config.ServiceDefinitions[normalizedName] = svcDef
	}

	for name, rtrDef := range depDef.InlineRouters {
		normalizedName := strings.ToLower(deploymentName + "." + name)
		config.RouterDefinitions[normalizedName] = rtrDef
	}

	// Process server-level inline definitions
	// Move to global with normalized names (lowercase, server-level overrides deployment-level if same name)
	for name, mwDef := range serverDef.InlineMiddlewares {
		normalizedName := strings.ToLower(deploymentName + "." + serverName + "." + name)
		config.MiddlewareDefinitions[normalizedName] = mwDef
	}

	for name, svcDef := range serverDef.InlineServices {
		normalizedName := strings.ToLower(deploymentName + "." + serverName + "." + name)
		config.ServiceDefinitions[normalizedName] = svcDef
	}

	for name, rtrDef := range serverDef.InlineRouters {
		normalizedName := strings.ToLower(deploymentName + "." + serverName + "." + name)
		config.RouterDefinitions[normalizedName] = rtrDef
	}

	// Build renaming map for reference resolution BEFORE clearing inline definitions
	// Maps short names to normalized names for this deployment+server context
	renamings := make(map[string]string)

	// Add deployment-level renamings (lowercase)
	for name := range depDef.InlineMiddlewares {
		renamings[name] = strings.ToLower(deploymentName + "." + name)
	}
	for name := range depDef.InlineServices {
		renamings[name] = strings.ToLower(deploymentName + "." + name)
	}
	for name := range depDef.InlineRouters {
		renamings[name] = strings.ToLower(deploymentName + "." + name)
	}

	// Add server-level renamings (these override deployment-level if same name, lowercase)
	for name := range serverDef.InlineMiddlewares {
		renamings[name] = strings.ToLower(deploymentName + "." + serverName + "." + name)
	}
	for name := range serverDef.InlineServices {
		renamings[name] = strings.ToLower(deploymentName + "." + serverName + "." + name)
	}
	for name := range serverDef.InlineRouters {
		renamings[name] = strings.ToLower(deploymentName + "." + serverName + "." + name)
	}

	// Clear inline definitions (they're now in global)
	depDef.InlineMiddlewares = nil
	depDef.InlineServices = nil
	depDef.InlineRouters = nil

	serverDef.InlineMiddlewares = nil
	serverDef.InlineServices = nil
	serverDef.InlineRouters = nil

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
	// Flatten and store configs to resolvedConfigs
	// Configs are already resolved at YAML byte level by loader (2-step resolution)
	// Now we flatten nested maps to dot notation for easy access via GetConfig()
	flattenAndStoreConfigs(registry, config.Configs, "")

	// Store middleware definitions to registry (no runtime registration yet)
	// Middlewares will be registered in RegisterDefinitionsForRuntime
	// For now, we don't need to store them - they're in config.MiddlewareDefinitions

	// Store service definitions as unresolved lazy service entries
	// Factory will be resolved later in RegisterDefinitionsForRuntime
	for name, svc := range config.ServiceDefinitions {
		svc.Name = name

		// Create unresolved LazyServiceEntry (FactoryType not yet resolved to actual function)
		// This allows services to be registered before service type factories are available
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

		registry.RegisterLazyServiceUnresolved(name, svc.Type, deps, svc.Config)
	}

	// Store router definitions to registry (deferred)
	for name, rtr := range config.RouterDefinitions {
		registry.DefineRouter(name, rtr)
	}

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

		// Get service definition (case-insensitive)
		svc, exists := getServiceDef(config.ServiceDefinitions, serviceName)
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
	// Case-insensitive deployment lookup
	depDef, ok := getDeploymentDef(config, deploymentName)
	if !ok {
		return fmt.Errorf("deployment %s not found", deploymentName)
	}

	// Case-insensitive server lookup
	serverDef, ok := getServerDef(depDef, serverName)
	if !ok {
		return fmt.Errorf("server %s not found in deployment %s", serverName, deploymentName)
	}

	// Register middlewares
	for name, mw := range config.MiddlewareDefinitions {
		mw.Name = name
		registry.RegisterMiddlewareName(name, mw.Type, mw.Config)
	}

	// Register service definitions
	// Merge service definitions from registry (RegisterLazyService) into config
	// This allows services registered via code to be available in config.ServiceDefinitions
	registry.MergeRegistryServicesToConfig(config)

	// Collect all services needed for this server (published services + their dependencies)
	servicesToRegister := collectAllServiceDependencies(config, serverTopo.Services)

	// Register/Resolve service definitions with remote/local logic
	// Iterate through all services (published + dependencies)
	for _, serviceName := range servicesToRegister {
		// Check if service already registered with inline factory (e.g., from annotations)
		if existingEntry := registry.GetLazyServiceEntry(serviceName); existingEntry != nil && existingEntry.IsResolved() {
			// Service already has inline factory - skip config-based resolution
			logger.LogDebug("⏭️  Skipping config resolution for '%s': already resolved with inline factory", serviceName)
			continue
		}

		svc, exists := getServiceDef(config.ServiceDefinitions, serviceName)
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
			// Check if already stored as unresolved entry from StoreDefinitionsToRegistry
			if existingEntry := registry.GetLazyServiceEntry(serviceName); existingEntry != nil && !existingEntry.IsResolved() {
				// Resolve existing unresolved entry
				serviceType := svc.Type
				factory := registry.GetServiceFactory(serviceType, true) // true = local factory
				if factory == nil {
					return fmt.Errorf("service factory %s (local) not registered for service %s", serviceType, serviceName)
				}

				// Resolve the entry by setting the factory
				existingEntry.ResolveFactory(func(resolvedDeps, cfg map[string]any) any {
					return factory(resolvedDeps, cfg)
				})
			} else {
				// Not in unresolved registry - register as new LOCAL service with dependency resolution
				// Convert DependsOn to deps map
				deps := make(map[string]string)
				for _, depStr := range svc.DependsOn {
					// Parse "paramName:serviceName" or just "serviceName"
					// Note: serviceName can be "@config.key" for config-based resolution
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
					// Call original factory with resolved dependencies (eager injection)
					return factory(resolvedDeps, cfg)
				}, deps, svc.Config, deploy.WithRegistrationMode(deploy.LazyServiceSkip))
			}
		}
	}

	// Collect published services from current server's apps
	publishedServicesMap := make(map[string]bool)
	for _, appDef := range serverDef.Apps {
		for _, serviceName := range appDef.PublishedServices {
			publishedServicesMap[serviceName] = true
		}
	}

	// IMPORTANT: Force instantiate all published services BEFORE creating routers
	// This ensures all service dependencies are resolved before router creation
	for serviceName := range publishedServicesMap {
		_, ok := registry.GetServiceAny(serviceName)
		if !ok {
			logger.LogWarn("⚠️  Warning: Published service '%s' failed to instantiate (dependencies may be missing)", serviceName)
		} else {
			logger.LogInfo("✅ Instantiated published service: %s", serviceName)
		}
	}

	// Auto-generate router definitions for published services
	// Also update Apps.Routers to use normalized router names
	// Priority: service.router > router-definitions > metadata > auto-generate
	routerRenamings := make(map[string]string) // old router name -> new router name

	for serviceName := range publishedServicesMap {
		routerName := serviceName + "-router"

		// Get service definition from config (case-insensitive)
		serviceDef, exists := getServiceDef(config.ServiceDefinitions, serviceName)
		if !exists {
			return fmt.Errorf("published service '%s' not found in service-definitions after normalization", serviceName)
		}

		var pathPrefix string
		var pathRewrites []schema.PathRewriteDef
		var middlewares []string
		var hidden []string
		var custom []schema.RouteDef

		// Priority 1: Check if service has embedded router definition
		if serviceDef.Router != nil {
			pathPrefix = serviceDef.Router.PathPrefix
			pathRewrites = serviceDef.Router.PathRewrites
			middlewares = serviceDef.Router.Middlewares
			hidden = serviceDef.Router.Hidden
			custom = serviceDef.Router.Custom
		}

		// Priority 2: Check if router manually defined in router-definitions (override/standalone, case-insensitive)
		if yamlRouter, exists := getRouterDef(config.RouterDefinitions, routerName); exists {
			// Override only if not set in service.router
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

		// Define auto-generated router
		autoRouter := &schema.RouterDef{
			PathPrefix:   pathPrefix,
			PathRewrites: pathRewrites,
			Middlewares:  middlewares,
			Hidden:       hidden,
			Custom:       custom,
		}

		// Store to config.RouterDefinitions so it's available for later lookup (lowercase key)
		config.RouterDefinitions[strings.ToLower(routerName)] = autoRouter

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

	// Auto-create and register router instances from published services
	// This creates actual router instances from services that have router metadata
	for serviceName := range publishedServicesMap {
		routerName := serviceName + "-router"

		// Check if router instance already registered (manual override)
		if registry.GetRouter(routerName) != nil {
			continue
		}

		// Get service definition to find its type (case-insensitive)
		serviceDef, exists := getServiceDef(config.ServiceDefinitions, serviceName)
		if !exists {
			continue
		}

		// Get service type metadata (has router config from @RouterService annotation)
		metadata := registry.GetServiceMetadata(serviceDef.Type)
		if metadata == nil {
			// Skip services without metadata
			continue
		}

		// Check if metadata has router configuration
		hasRouterConfig := len(metadata.RouteOverrides) > 0 || metadata.PathPrefix != ""

		if !hasRouterConfig {
			// Skip services without router configuration
			continue
		}

		// Force instantiate service by calling GetService with type assertion
		// This will trigger dependency resolution and instantiation
		// We use a generic approach since we don't know the service type at compile time
		var serviceInstance any
		var ok bool

		// Try to get service instance (this should trigger instantiation if lazy)
		serviceInstance, ok = registry.GetServiceAny(serviceName)
		if !ok || serviceInstance == nil {
			// Service failed to instantiate - log detailed error
			logger.LogWarn("⚠️  Warning: Service '%s' failed to instantiate (dependencies may not be ready), creating lazy router factory instead", serviceName)

			// Create a lazy router factory that will try again when GetRouter is called
			registry.RegisterRouterFactory(routerName, func() router.Router {
				// Try to get service instance again
				svcInst, ok := registry.GetServiceAny(serviceName)
				if !ok || svcInst == nil {
					panic(fmt.Sprintf("Service '%s' still not instantiated when router '%s' requested", serviceName, routerName))
				}

				// Build router from service
				routerDef := registry.GetRouterDef(routerName)
				finalPrefix := metadata.PathPrefix
				if routerDef != nil && routerDef.PathPrefix != "" {
					finalPrefix = routerDef.PathPrefix
				}

				opts := &router.ServiceRouterOptions{
					Prefix:         finalPrefix,
					Middlewares:    metadata.MiddlewareNames,
					RouteOverrides: make(map[string]router.RouteMeta),
				}

				for methodName, routeMeta := range metadata.RouteOverrides {
					opts.RouteOverrides[methodName] = router.RouteMeta{
						HTTPMethod: routeMeta.Method,
						Path:       routeMeta.Path,
					}
				}

				return router.NewFromService(svcInst, opts)
			})

			logger.LogDebug("🔧 Registered lazy router factory for '%s' (will instantiate service on-demand)", routerName)
			continue
		}

		// Get RouterDef (may have PathPrefix from router-definitions YAML)
		routerDef := registry.GetRouterDef(routerName)

		// Determine final PathPrefix (priority: RouterDef > Metadata)
		finalPrefix := metadata.PathPrefix
		if routerDef != nil && routerDef.PathPrefix != "" {
			finalPrefix = routerDef.PathPrefix
		}

		// Build ServiceRouterOptions from metadata + RouterDef
		opts := &router.ServiceRouterOptions{
			Prefix:         finalPrefix, // Use final prefix (YAML overrides annotation)
			Middlewares:    metadata.MiddlewareNames,
			RouteOverrides: make(map[string]router.RouteMeta),
		}

		// Convert metadata.RouteOverrides to router.RouteMeta
		for methodName, routeMeta := range metadata.RouteOverrides {
			opts.RouteOverrides[methodName] = router.RouteMeta{
				HTTPMethod: routeMeta.Method,
				Path:       routeMeta.Path,
			}
		}

		// Create router from service
		r := router.NewFromService(serviceInstance, opts)

		// Register router instance
		registry.RegisterRouter(routerName, r)
		logger.LogDebug("🔧 Auto-created router '%s' from service '%s' (type: %s, prefix: %s)", routerName, serviceName, serviceDef.Type, finalPrefix)
	}

	return nil
}

// LoadConfig loads config and builds ALL deployments into Global registry
// Returns error only - deployments are stored in deploy.Global()
func LoadConfig(configPaths ...string) (*schema.DeployConfig, error) {
	if len(configPaths) == 0 {
		configPaths = []string{"config"}
	}

	config, err := loadConfig(configPaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	registry := deploy.Global()

	// Store original config for inline definitions normalization
	registry.StoreDeployConfig(config)

	// NOTE: normalizeServerDefinitions already called in LoadConfig STEP 9
	// No need to call again here

	// Store definitions to registry (NO runtime registration, just store data)
	// Runtime registration will happen in RunCurrentServer
	if err := StoreDefinitionsToRegistry(registry, config); err != nil {
		return nil, fmt.Errorf("failed to store definitions: %w", err)
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
				Name:            serverName,
				DeploymentName:  deploymentName,
				BaseURL:         serverDef.BaseURL,
				ConfigOverrides: serverDef.ConfigOverrides,
				Services:        make([]string, 0),
				RemoteServices:  make(map[string]string),
				Apps:            make([]*deploy.AppTopology, 0, len(serverDef.Apps)),
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

	logger.LogDebug("✅ Config loaded successfully from: %v", configPaths)
	return config, nil
}

// LoadDbPoolDefsFromConfig auto-discovers and sets up named DB pools from config
// Requires dbpool-manager service to be already registered
func LoadDbPoolDefsFromConfig() error {
	registry := deploy.Global()
	config := registry.GetDeployConfig()

	// Check if dbpool-definitions section exists
	if len(config.DbPoolDefinitions) == 0 {
		// No dbpool-definitions section, skip
		return nil
	}

	dpm, ok := deploy.Global().GetServiceAny("dbpool-manager")
	if !ok || dpm == nil {
		return fmt.Errorf("dbpool-manager service not found in registry")
	}
	dbPoolManager, ok := dpm.(serviceapi.DbPoolManager)
	if !ok {
		return fmt.Errorf("dbpool-manager service does not implement serviceapi.DbPoolManager interface")
	}

	// Setup each pool
	for poolName, poolConfig := range config.DbPoolDefinitions {
		// Extract DSN or build from components
		dsn := poolConfig.DSN

		// Extract optional pool parameters with best practice defaults
		minConns := poolConfig.MinConns
		if minConns == 0 {
			minConns = 2 // Best practice: 2 minimum connections
		}

		maxConns := poolConfig.MaxConns
		if maxConns == 0 {
			maxConns = 10 // Best practice: 10 max connections
		}

		maxIdleTime := 30 * time.Minute // Best practice default
		if poolConfig.MaxIdleTime != "" {
			if parsed, err := time.ParseDuration(poolConfig.MaxIdleTime); err == nil {
				maxIdleTime = parsed
			}
		}

		maxLifetime := time.Hour // Best practice default
		if poolConfig.MaxLifetime != "" {
			if parsed, err := time.ParseDuration(poolConfig.MaxLifetime); err == nil {
				maxLifetime = parsed
			}
		}

		// If no DSN, build from components
		if dsn == "" {
			host := poolConfig.Host
			port := poolConfig.Port
			if port == 0 {
				port = 5432 // Default PostgreSQL port
			}
			database := poolConfig.Database
			username := poolConfig.Username
			password := poolConfig.Password

			if host == "" || database == "" {
				return fmt.Errorf("dbpool-manager.%s: must provide either 'dsn' or 'host'+'database'", poolName)
			}

			// Build DSN with best practice defaults
			sslmode := poolConfig.SSLMode
			if sslmode == "" {
				sslmode = "disable"
			}

			dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_min_conns=%d&pool_max_conns=%d&pool_max_conn_idle_time=%s&pool_max_conn_lifetime=%s",
				username, password, host, port, database, sslmode, minConns, maxConns, maxIdleTime, maxLifetime)
		} else {
			// DSN provided - apply optional pool parameters if not already set
			opts := ""
			if !strings.Contains(dsn, "pool_min_conns=") {
				opts += fmt.Sprintf("&pool_min_conns=%d", minConns)
			}
			if !strings.Contains(dsn, "pool_max_conns=") {
				opts += fmt.Sprintf("&pool_max_conns=%d", maxConns)
			}
			if !strings.Contains(dsn, "pool_max_conn_idle_time=") {
				opts += fmt.Sprintf("&pool_max_conn_idle_time=%s", maxIdleTime)
			}
			if !strings.Contains(dsn, "pool_max_conn_lifetime=") {
				opts += fmt.Sprintf("&pool_max_conn_lifetime=%s", maxLifetime)
			}
			if strings.Contains(dsn, "?") {
				dsn += opts
			} else {
				dsn += "?" + strings.TrimPrefix(opts, "&")
			}
		}

		// Extract schema (default: public)
		schema := poolConfig.Schema
		if schema == "" {
			schema = "public"
		}

		// Set DSN and Schema for poolName
		// This also auto-registers the pool as a lazy service
		dbPoolManager.SetDbPoolManager(poolName, dsn, schema, poolConfig.RlsContext)

		logger.LogDebug("✅ Registered DB pool: %s (schema: %s)", poolName, schema)
	}

	return nil
}
