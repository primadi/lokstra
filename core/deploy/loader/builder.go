package loader

import (
	"fmt"
	"strings"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/schema"
)

// normalizeServerDefinitions converts server-level helper fields to a new app
// This allows shorthand syntax: addr/routers/published-services at server level
// for the common case of 1 server = 1 app
// If helper fields are present, a NEW app is created and PREPENDED to Apps array
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

			// Create a new app from helper fields
			newApp := &schema.AppDefMap{
				Addr:              serverDef.HelperAddr,
				Routers:           serverDef.HelperRouters,
				PublishedServices: serverDef.HelperPublishedServices,
			}

			// PREPEND new app to Apps array (so it becomes first)
			serverDef.Apps = append([]*schema.AppDefMap{newApp}, serverDef.Apps...)

			// Clear helper fields
			serverDef.HelperAddr = ""
			serverDef.HelperRouters = nil
			serverDef.HelperPublishedServices = nil
		}
	}
}

// LoadAndBuild loads config and builds ALL deployments into Global registry
// Returns error only - deployments are stored in deploy.Global()
func LoadAndBuild(configPaths []string) error {
	config, err := LoadConfig(configPaths...)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	registry := deploy.Global()

	// Get external service definitions
	externalServices := config.ExternalServiceDefinitions

	// Register configs from YAML
	for name, value := range config.Configs {
		registry.DefineConfig(&schema.ConfigDef{
			Name:  name,
			Value: value,
		})
	}

	// Resolve configs
	if err := registry.ResolveConfigs(); err != nil {
		return fmt.Errorf("failed to resolve configs: %w", err)
	}

	// Register middlewares from YAML
	for name, mw := range config.MiddlewareDefinitions {
		mw.Name = name // Set name from map key
		registry.DefineMiddleware(mw)
	}

	// Normalize server definitions (convert helper fields to apps)
	normalizeServerDefinitions(config)

	// Auto-create service wrappers for external services with factory type
	// This allows external-service-definitions to directly specify the factory
	// without needing a separate service-definitions entry
	for name, extSvc := range externalServices {
		if extSvc.Type != "" {
			// Check if service definition already exists (manual override)
			if _, exists := config.ServiceDefinitions[name]; exists {
				continue // Skip auto-creation, use manual definition
			}

			// Auto-create service definition
			autoServiceDef := &schema.ServiceDef{
				Name:      name,
				Type:      extSvc.Type,
				DependsOn: nil, // External services have no dependencies
				Config:    make(map[string]any),
			}

			// Copy external service config if provided
			if extSvc.Config != nil {
				for k, v := range extSvc.Config {
					autoServiceDef.Config[k] = v
				}
			}

			// Add URL, resource, convention metadata to config
			autoServiceDef.Config["url"] = extSvc.URL
			if extSvc.Resource != "" {
				autoServiceDef.Config["resource"] = extSvc.Resource
			}
			if extSvc.ResourcePlural != "" {
				autoServiceDef.Config["resource_plural"] = extSvc.ResourcePlural
			}
			if extSvc.Convention != "" {
				autoServiceDef.Config["convention"] = extSvc.Convention
			}
			if extSvc.Overrides != "" {
				autoServiceDef.Config["overrides"] = extSvc.Overrides
			}

			// Add to service definitions (will be registered below)
			config.ServiceDefinitions[name] = autoServiceDef
		}
	}

	// Register services from YAML (includes auto-created external services)
	for name, svc := range config.ServiceDefinitions {
		svc.Name = name // Set name from map key
		registry.DefineService(svc)
	}

	// Auto-generate router definitions for published services
	// This scans ALL deployments to collect ALL published services
	publishedServicesMap := make(map[string]bool)
	for _, depDef := range config.Deployments {
		for _, serverDef := range depDef.Servers {
			for _, appDef := range serverDef.Apps {
				for _, serviceName := range appDef.PublishedServices {
					publishedServicesMap[serviceName] = true
				}
			}
		}
	}

	// Define routers for each published service
	for serviceName := range publishedServicesMap {
		routerName := serviceName + "-router"

		// Get service definition to find service type
		serviceDef, ok := config.ServiceDefinitions[serviceName]
		if !ok {
			return fmt.Errorf("published service '%s' not found in service-definitions", serviceName)
		}

		// Get service metadata from factory registration (RegisterServiceType options)
		metadata := registry.GetServiceMetadata(serviceDef.Type)

		// Metadata Resolution Priority (3 levels):
		//   1. YAML routers: section (highest - deployment-specific override)
		//   2. RegisterServiceType options (medium - framework default)
		//   3. Auto-generate from service name (lowest - fallback)
		//
		// Note: XXXRemote struct metadata (RemoteServiceMeta) is checked at runtime
		// by BuildRouterFromDefinition, not here during config loading.
		var resourceName, resourcePlural, convention, overrides string

		// Priority 1: Check if router manually defined in YAML (for overrides)
		if yamlRouter, exists := config.Routers[routerName]; exists {
			// Use YAML definition (allows override)
			resourceName = yamlRouter.Resource
			resourcePlural = yamlRouter.ResourcePlural
			convention = yamlRouter.Convention
			overrides = yamlRouter.Overrides
		}

		// Priority 2: Fallback to metadata from RegisterServiceType
		if resourceName == "" && metadata != nil && metadata.Resource != "" {
			resourceName = metadata.Resource
			resourcePlural = metadata.ResourcePlural
			convention = metadata.Convention
		}

		// Priority 3: Final fallback - auto-generate from service name
		if resourceName == "" {
			resourceName = strings.TrimSuffix(serviceName, "-service")
			resourcePlural = resourceName + "s" // Simple pluralization
			convention = "rest"
		}

		// Define router (will override if exists in YAML)
		registry.DefineRouter(routerName, &schema.RouterDef{
			Service:        serviceName,
			Convention:     convention,
			Resource:       resourceName,
			ResourcePlural: resourcePlural,
			Overrides:      overrides,
		})
	}

	// Register router overrides from YAML
	for name, overrideDef := range config.RouterOverrides {
		registry.DefineRouterOverride(name, overrideDef)
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
			} // TODO: Auto-detect remote services from factory dependencies
			// For now, remote service resolution happens during service registration
			// based on published-services across servers in deployment

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

	return nil
}

// LoadAndBuildFromDir loads all YAML files from a directory and builds ALL deployments
func LoadAndBuildFromDir(dirPath string) error {
	// Use LoadConfigFromDir to scan directory for YAML files
	config, err := LoadConfigFromDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to load config from directory: %w", err)
	}

	// Process the config (same logic as LoadAndBuild but starting from loaded config)
	registry := deploy.Global()

	// Get external service definitions
	externalServices := config.ExternalServiceDefinitions

	// Register configs from YAML
	for name, value := range config.Configs {
		registry.DefineConfig(&schema.ConfigDef{
			Name:  name,
			Value: value,
		})
	}

	// Resolve configs
	if err := registry.ResolveConfigs(); err != nil {
		return fmt.Errorf("failed to resolve configs: %w", err)
	}

	// Register middlewares from YAML
	for name, mw := range config.MiddlewareDefinitions {
		mw.Name = name // Set name from map key
		registry.DefineMiddleware(mw)
	}

	// Normalize server definitions (convert helper fields to apps)
	normalizeServerDefinitions(config)

	// Auto-create service wrappers for external services with factory type
	// This allows external-service-definitions to directly specify the factory
	// without needing a separate service-definitions entry
	for name, extSvc := range externalServices {
		if extSvc.Type != "" {
			// Check if service definition already exists (manual override)
			if _, exists := config.ServiceDefinitions[name]; exists {
				continue // Skip auto-creation, use manual definition
			}

			// Auto-create service definition
			autoServiceDef := &schema.ServiceDef{
				Name:      name,
				Type:      extSvc.Type,
				DependsOn: nil, // External services have no dependencies
				Config:    make(map[string]any),
			}

			// Copy external service config if provided
			if extSvc.Config != nil {
				for k, v := range extSvc.Config {
					autoServiceDef.Config[k] = v
				}
			}

			// Add URL, resource, convention metadata to config
			autoServiceDef.Config["url"] = extSvc.URL
			if extSvc.Resource != "" {
				autoServiceDef.Config["resource"] = extSvc.Resource
			}
			if extSvc.ResourcePlural != "" {
				autoServiceDef.Config["resource_plural"] = extSvc.ResourcePlural
			}
			if extSvc.Convention != "" {
				autoServiceDef.Config["convention"] = extSvc.Convention
			}
			if extSvc.Overrides != "" {
				autoServiceDef.Config["overrides"] = extSvc.Overrides
			}

			// Add to service definitions (will be registered below)
			config.ServiceDefinitions[name] = autoServiceDef
		}
	}

	// Register services from YAML (includes auto-created external services)
	for name, svc := range config.ServiceDefinitions {
		svc.Name = name // Set name from map key
		registry.DefineService(svc)
	}

	// Auto-generate router definitions for published services
	publishedServicesMap := make(map[string]bool)
	for _, depDef := range config.Deployments {
		for _, serverDef := range depDef.Servers {
			for _, appDef := range serverDef.Apps {
				for _, serviceName := range appDef.PublishedServices {
					publishedServicesMap[serviceName] = true
				}
			}
		}
	}

	// Define routers for each published service
	for serviceName := range publishedServicesMap {
		routerName := serviceName + "-router"

		serviceDef, ok := config.ServiceDefinitions[serviceName]
		if !ok {
			return fmt.Errorf("published service '%s' not found in service-definitions", serviceName)
		}

		metadata := registry.GetServiceMetadata(serviceDef.Type)

		var resourceName, resourcePlural, convention, overrides string

		if yamlRouter, exists := config.Routers[routerName]; exists {
			resourceName = yamlRouter.Resource
			resourcePlural = yamlRouter.ResourcePlural
			convention = yamlRouter.Convention
			overrides = yamlRouter.Overrides
		}

		if resourceName == "" && metadata != nil && metadata.Resource != "" {
			resourceName = metadata.Resource
			resourcePlural = metadata.ResourcePlural
			convention = metadata.Convention
		}

		if resourceName == "" {
			resourceName = strings.TrimSuffix(serviceName, "-service")
			resourcePlural = resourceName + "s"
			convention = "rest"
		}

		registry.DefineRouter(routerName, &schema.RouterDef{
			Service:        serviceName,
			Convention:     convention,
			Resource:       resourceName,
			ResourcePlural: resourcePlural,
			Overrides:      overrides,
		})
	}

	// Register router overrides from YAML
	for name, overrideDef := range config.RouterOverrides {
		registry.DefineRouterOverride(name, overrideDef)
	}

	// Build ALL deployments (create topology only)
	for deploymentName, depDef := range config.Deployments {
		// Build service location registry
		serviceLocations := make(map[string]string)
		for _, serverDef := range depDef.Servers {
			for _, appDef := range serverDef.Apps {
				for _, serviceName := range appDef.PublishedServices {
					fullURL := serverDef.BaseURL + appDef.Addr
					serviceLocations[serviceName] = fullURL
				}
			}
		}

		// Create and store topology
		deployTopo := &deploy.DeploymentTopology{
			Name:            deploymentName,
			ConfigOverrides: make(map[string]any),
			Servers:         make(map[string]*deploy.ServerTopology),
		}

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
			serviceMap := make(map[string]bool)
			for _, appDef := range serverDef.Apps {
				for _, svcName := range appDef.PublishedServices {
					serviceMap[svcName] = true
				}
			}
			for svcName := range serviceMap {
				serverTopo.Services = append(serverTopo.Services, svcName)
			} // TODO: Auto-detect remote services from factory dependencies
			// For now, remote service resolution happens during service registration

			// Build app topologies
			for _, appDef := range serverDef.Apps {
				appTopo := &deploy.AppTopology{
					Addr:    appDef.Addr,
					Routers: make([]string, 0, len(appDef.Routers)+len(appDef.PublishedServices)),
				}

				appTopo.Routers = append(appTopo.Routers, appDef.Routers...)
				for _, serviceName := range appDef.PublishedServices {
					routerName := serviceName + "-router"
					appTopo.Routers = append(appTopo.Routers, routerName)
				}

				serverTopo.Apps = append(serverTopo.Apps, appTopo)
			}

			deployTopo.Servers[serverName] = serverTopo
		}

		registry.StoreDeploymentTopology(deployTopo)
	}

	return nil
}
