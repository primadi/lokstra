package loader

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

//go:embed lokstra.schema.json
var schemaFS embed.FS

// LoadConfig loads a deployment configuration from YAML file(s)
// Supports single file or multiple files that will be merged
func LoadConfig(paths ...string) (*schema.DeployConfig, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no config files specified")
	}

	var merged *schema.DeployConfig

	// Load and merge each file
	for _, path := range paths {
		config, err := loadSingleFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", path, err)
		}

		if merged == nil {
			merged = config
		} else {
			merged = mergeConfigs(merged, config)
		}
	}

	// Validate merged config
	if err := ValidateConfig(merged); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return merged, nil
}

// loadSingleFile loads and parses a single YAML file
func loadSingleFile(path string) (*schema.DeployConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var config schema.DeployConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// mergeConfigs merges two configurations (target <- source)
// Source values override target values
func mergeConfigs(target, source *schema.DeployConfig) *schema.DeployConfig {
	result := &schema.DeployConfig{
		Configs:                  mergeMap(target.Configs, source.Configs),
		ServiceDefinitions:       mergeMaps(target.ServiceDefinitions, source.ServiceDefinitions),
		Routers:                  mergeMaps(target.Routers, source.Routers),
		RemoteServiceDefinitions: mergeMaps(target.RemoteServiceDefinitions, source.RemoteServiceDefinitions),
		Deployments:              mergeMaps(target.Deployments, source.Deployments),
	}
	return result
}

// mergeMap merges two maps (any values)
func mergeMap(target, source map[string]any) map[string]any {
	if target == nil {
		target = make(map[string]any)
	}
	if source == nil {
		return target
	}

	result := make(map[string]any, len(target)+len(source))
	for k, v := range target {
		result[k] = v
	}
	for k, v := range source {
		result[k] = v // Source overrides target
	}
	return result
}

// mergeMaps merges two maps of pointers
func mergeMaps[T any](target, source map[string]*T) map[string]*T {
	if target == nil {
		target = make(map[string]*T)
	}
	if source == nil {
		return target
	}

	result := make(map[string]*T, len(target)+len(source))
	for k, v := range target {
		result[k] = v
	}
	for k, v := range source {
		result[k] = v // Source overrides target
	}
	return result
}

// ValidateConfig validates a deployment configuration against JSON schema
func ValidateConfig(config *schema.DeployConfig) error {
	// Load embedded schema
	schemaData, err := schemaFS.ReadFile("lokstra.schema.json")
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaData)

	// Convert config to map for validation
	configMap := configToMap(config)
	documentLoader := gojsonschema.NewGoLoader(configMap)

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, err := range result.Errors() {
			errors = append(errors, fmt.Sprintf("  - %s: %s", err.Field(), err.Description()))
		}
		return fmt.Errorf("schema validation failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// configToMap converts DeployConfig to map[string]any for JSON schema validation
func configToMap(config *schema.DeployConfig) map[string]any {
	result := make(map[string]any)

	if len(config.Configs) > 0 {
		result["configs"] = config.Configs
	}

	if len(config.ServiceDefinitions) > 0 {
		services := make(map[string]any)
		for name, svc := range config.ServiceDefinitions {
			svcMap := map[string]any{
				"type": svc.Type,
			}
			if len(svc.DependsOn) > 0 {
				svcMap["depends-on"] = svc.DependsOn
			}
			if len(svc.Config) > 0 {
				svcMap["config"] = svc.Config
			}
			services[name] = svcMap
		}
		result["service-definitions"] = services
	}

	if len(config.Routers) > 0 {
		routers := make(map[string]any)
		for name, rtr := range config.Routers {
			rtrMap := map[string]any{
				"service": rtr.Service,
			}
			if len(rtr.Overrides) > 0 {
				rtrMap["overrides"] = rtr.Overrides
			}
			routers[name] = rtrMap
		}
		result["routers"] = routers
	}

	if len(config.RemoteServiceDefinitions) > 0 {
		remotes := make(map[string]any)
		for name, rs := range config.RemoteServiceDefinitions {
			rsMap := map[string]any{
				"url":      rs.URL,
				"resource": rs.Resource,
			}
			if rs.ResourcePlural != "" {
				rsMap["resource-plural"] = rs.ResourcePlural
			}
			remotes[name] = rsMap
		}
		result["remote-service-definitions"] = remotes
	}

	if len(config.Deployments) > 0 {
		deployments := make(map[string]any)
		for name, dep := range config.Deployments {
			depMap := make(map[string]any)

			if len(dep.ConfigOverrides) > 0 {
				depMap["config-overrides"] = dep.ConfigOverrides
			}

			if len(dep.Servers) > 0 {
				servers := make(map[string]any)
				for srvName, srv := range dep.Servers {
					srvMap := map[string]any{
						"base-url": srv.BaseURL,
					}

					if len(srv.Apps) > 0 {
						apps := make([]any, len(srv.Apps))
						for i, app := range srv.Apps {
							appMap := map[string]any{
								"addr": app.Addr,
							}
							if len(app.Services) > 0 {
								appMap["required-services"] = app.Services
							}
							if len(app.Routers) > 0 {
								appMap["routers"] = app.Routers
							}
							if len(app.RemoteServices) > 0 {
								appMap["required-remote-services"] = app.RemoteServices
							}
							apps[i] = appMap
						}
						srvMap["apps"] = apps
					}

					servers[srvName] = srvMap
				}
				depMap["servers"] = servers
			}

			deployments[name] = depMap
		}
		result["deployments"] = deployments
	}

	return result
}

// LoadConfigFromDir loads all .yaml and .yml files from a directory and merges them
func LoadConfigFromDir(dirPath string) (*schema.DeployConfig, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
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
		return nil, fmt.Errorf("no YAML files found in directory: %s", dirPath)
	}

	return LoadConfig(paths...)
}
