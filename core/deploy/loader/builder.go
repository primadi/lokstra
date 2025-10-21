package loader

import (
	"fmt"
	"strings"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/schema"
)

// BuildDeployment builds a Deployment from loaded configuration
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
			app := server.NewApp(appDef.Addr)

			// Add services
			if len(appDef.Services) > 0 {
				app.AddServices(appDef.Services...)
			}

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

	// Register services from YAML
	for name, svc := range config.ServiceDefinitions {
		svc.Name = name // Set name from map key
		registry.DefineService(svc)
	}

	// Build ALL deployments
	for deploymentName, depDef := range config.Deployments {
		// Create deployment
		dep := deploy.NewWithRegistry(deploymentName, registry)

		// Apply config overrides
		for key, value := range depDef.ConfigOverrides {
			dep.SetConfigOverride(key, value)
		}

		// Create servers and apps
		for serverName, serverDef := range depDef.Servers {
			server := dep.NewServer(serverName, serverDef.BaseURL)

			// Create apps
			for _, appDef := range serverDef.Apps {
				app := server.NewApp(appDef.Addr)

				// Add server-level services (shared across all apps on this server)
				if len(serverDef.Services) > 0 {
					app.AddServices(serverDef.Services...)
				}

				// Add app-level services (specific to this app)
				if len(appDef.Services) > 0 {
					app.AddServices(appDef.Services...)
				}

				// Add server-level remote services
				for _, remoteServiceName := range serverDef.RemoteServices {
					// Look up remote service definition
					remoteDef, ok := config.RemoteServiceDefinitions[remoteServiceName]
					if !ok {
						return fmt.Errorf("remote service '%s' not found in remote-service-definitions", remoteServiceName)
					}

					// Extract the actual service name (remove -remote suffix if present)
					actualServiceName := strings.TrimSuffix(remoteServiceName, "-remote")

					// Add as remote service with base URL
					app.AddRemoteServiceByName(actualServiceName, remoteDef.URL)
				}

				// Add app-level remote services
				for _, remoteServiceName := range appDef.RemoteServices {
					// Look up remote service definition
					remoteDef, ok := config.RemoteServiceDefinitions[remoteServiceName]
					if !ok {
						return fmt.Errorf("remote service '%s' not found in remote-service-definitions", remoteServiceName)
					}

					// Extract the actual service name
					actualServiceName := strings.TrimSuffix(remoteServiceName, "-remote")

					// Add as remote service with base URL
					app.AddRemoteServiceByName(actualServiceName, remoteDef.URL)
				}

				// Add router names (routers are registered separately in code)
				for _, routerName := range appDef.Routers {
					app.AddRouter(routerName, nil) // nil because actual router is in global registry
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
