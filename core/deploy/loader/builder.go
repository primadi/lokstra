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
} // BuildDeployment builds a Deployment from loaded configuration
func BuildDeployment(config *schema.DeployConfig, deploymentName string, registry *deploy.GlobalRegistry) (*deploy.Deployment, error) {
	// Get deployment definition
	depDef, ok := config.Deployments[deploymentName]
	if !ok {
		return nil, fmt.Errorf("deployment %s not found in config", deploymentName)
	}

	// Register configs from YAML
	for name, value := range config.Configs {
		registry.DefineConfig(&schema.ConfigDef{
			Name:  name,
			Value: value,
		})
	}

	// Resolve configs
	if err := registry.ResolveConfigs(); err != nil {
		return nil, fmt.Errorf("failed to resolve configs: %w", err)
	}

	// Register services from YAML
	for name, svc := range config.ServiceDefinitions {
		svc.Name = name // Set name from map key
		registry.DefineService(svc)
	}

	// Create deployment
	dep := deploy.NewWithRegistry(deploymentName, registry)

	// Apply config overrides
	for key, value := range depDef.ConfigOverrides {
		dep.SetConfigOverride(key, value)
	}

	// Create servers and apps
	for serverName, serverDef := range depDef.Servers {
		server := dep.NewServer(serverName, serverDef.BaseURL)

		for _, appDef := range serverDef.Apps {
			_ = server.NewApp(appDef.Addr)

			// Services are at server-level only (shared across all apps)
			// No app-level services

			// TODO: Add routers when router system is ready
			// TODO: Add remote services when remote service system is ready
		}
	}

	return dep, nil
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

	// Normalize server definitions (convert helper fields to apps)
	normalizeServerDefinitions(config)

	// Register services from YAML
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

	// Build ALL deployments
	for deploymentName, depDef := range config.Deployments {
		// Create deployment
		dep := deploy.NewWithRegistry(deploymentName, registry)

		// Apply config overrides
		for key, value := range depDef.ConfigOverrides {
			dep.SetConfigOverride(key, value)
		}

		// First pass: Build service location registry (service-name → base-url)
		// This maps published services to their server URLs
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

		// Second pass: Create servers and apps
		for serverName, serverDef := range depDef.Servers {
			server := dep.NewServer(serverName, serverDef.BaseURL)

			// Create apps
			for _, appDef := range serverDef.Apps {
				app := server.NewApp(appDef.Addr)

				// Add server-level services (shared across all apps on this server)
				if len(serverDef.Services) > 0 {
					app.AddServices(serverDef.Services...)
				}

				// Auto-add published services (they must be local)
				if len(appDef.PublishedServices) > 0 {
					app.AddServices(appDef.PublishedServices...)
				}

				// Add server-level remote services (auto-resolved)
				for _, remoteServiceName := range serverDef.RemoteServices {
					// Extract the actual service name (remove -remote suffix if present)
					actualServiceName := strings.TrimSuffix(remoteServiceName, "-remote")

					// Auto-resolve URL from service locations
					remoteURL, found := serviceLocations[actualServiceName]
					if !found {
						// Fallback to external-service-definitions if exists
						if remoteDef, ok := externalServices[remoteServiceName]; ok {
							remoteURL = remoteDef.URL
						} else {
							return fmt.Errorf("remote service '%s' not found - not published in any server and not in external-service-definitions", actualServiceName)
						}
					}

					// Get optional overrides from YAML
					var remoteDef *schema.RemoteServiceSimple
					if def, ok := externalServices[remoteServiceName]; ok {
						remoteDef = def
					}

					// Add as remote service with auto-resolved URL
					app.AddRemoteServiceByName(actualServiceName, remoteURL, remoteDef)
				}

				// Add router names (routers are registered separately in code)
				for _, routerName := range appDef.Routers {
					app.AddRouter(routerName, nil) // nil because actual router is in global registry
				}

				// Auto-generate routers from published-services
				for _, serviceName := range appDef.PublishedServices {
					// Auto-generated router name: serviceName + "-router"
					routerName := serviceName + "-router"
					app.AddRouter(routerName, nil)
				}
			}
		}

		// Register deployment in global registry
		registry.RegisterDeployment(deploymentName, dep)
	}

	return nil
}

// LoadAndBuildFromDir loads all YAML files from a directory and builds ALL deployments
func LoadAndBuildFromDir(dirPath string) error {
	// For now, just return an error suggesting to use LoadAndBuild
	return fmt.Errorf("LoadAndBuildFromDir not yet implemented for new pattern, use LoadAndBuild with explicit paths")
}
