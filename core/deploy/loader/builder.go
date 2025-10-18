package loader

import (
	"fmt"

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

// LoadAndBuild is a convenience function that loads config and builds a deployment
func LoadAndBuild(configPaths []string, deploymentName string, registry *deploy.GlobalRegistry) (*deploy.Deployment, error) {
	config, err := LoadConfig(configPaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return BuildDeployment(config, deploymentName, registry)
}

// LoadAndBuildFromDir loads all YAML files from a directory and builds a deployment
func LoadAndBuildFromDir(dirPath string, deploymentName string, registry *deploy.GlobalRegistry) (*deploy.Deployment, error) {
	config, err := LoadConfigFromDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from directory: %w", err)
	}

	return BuildDeployment(config, deploymentName, registry)
}
