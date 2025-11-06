package loader

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// LoadConfig loads a deployment configuration from YAML file(s)
// Supports single file or multiple files that will be merged
func LoadConfig(paths ...string) (*schema.DeployConfig, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no config files specified")
	}

	var merged *schema.DeployConfig

	basePath := utils.GetBasePath()
	log.Println(basePath)
	// Load and merge each file
	for _, path := range paths {
		normPath := filepath.Join(basePath, path)
		config, err := loadSingleFile(normPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", path, err)
		}

		if merged == nil {
			merged = config
		} else {
			merged = mergeConfigs(merged, config)
		}
	}

	// Normalize server definitions (convert helper fields to apps) BEFORE validation
	normalizeServerDefinitions(merged)

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

	// Use strict YAML decoder to catch unknown fields
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true) // This will error on unknown fields like "services" instead of "service-definitions"

	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// mergeConfigs merges two configurations (target <- source)
// Source values override target values
func mergeConfigs(target, source *schema.DeployConfig) *schema.DeployConfig {
	result := &schema.DeployConfig{
		Configs:                    mergeMap(target.Configs, source.Configs),
		MiddlewareDefinitions:      mergeMaps(target.MiddlewareDefinitions, source.MiddlewareDefinitions),
		ServiceDefinitions:         mergeMaps(target.ServiceDefinitions, source.ServiceDefinitions),
		RouterDefinitions:          mergeMaps(target.RouterDefinitions, source.RouterDefinitions), // Renamed from Routers
		ExternalServiceDefinitions: mergeMaps(target.ExternalServiceDefinitions, source.ExternalServiceDefinitions),
		Deployments:                mergeMaps(target.Deployments, source.Deployments),
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
	// Load embedded schema from schema package
	schemaData := schema.GetSchemaBytes()
	schemaLoader := gojsonschema.NewBytesLoader(schemaData)

	// Convert config to map for validation
	configMap := configToMap(config)
	documentLoader := gojsonschema.NewGoLoader(configMap) // Validate
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

	if len(config.RouterDefinitions) > 0 {
		routers := make(map[string]any)
		for name, rtr := range config.RouterDefinitions {
			rtrMap := make(map[string]any)

			// Basic fields
			if rtr.Convention != "" {
				rtrMap["convention"] = rtr.Convention
			}
			if rtr.Resource != "" {
				rtrMap["resource"] = rtr.Resource
			}
			if rtr.ResourcePlural != "" {
				rtrMap["resource-plural"] = rtr.ResourcePlural
			}

			// Inline override fields
			if rtr.PathPrefix != "" {
				rtrMap["path-prefix"] = rtr.PathPrefix
			}
			if len(rtr.Middlewares) > 0 {
				rtrMap["middlewares"] = rtr.Middlewares
			}
			if len(rtr.Hidden) > 0 {
				rtrMap["hidden"] = rtr.Hidden
			}
			if len(rtr.Custom) > 0 {
				rtrMap["custom"] = rtr.Custom
			}

			routers[name] = rtrMap
		}
		result["router-definitions"] = routers // Renamed from "routers"
	}

	if len(config.ExternalServiceDefinitions) > 0 {
		externals := make(map[string]any)
		for name, es := range config.ExternalServiceDefinitions {
			esMap := map[string]any{
				"url": es.URL,
			}
			if es.Router != nil {
				if es.Router.Resource != "" {
					esMap["resource"] = es.Router.Resource
				}
				if es.Router.ResourcePlural != "" {
					esMap["resource-plural"] = es.Router.ResourcePlural
				}
				if es.Router.Convention != "" {
					esMap["convention"] = es.Router.Convention
				}
				// Inline override fields
				if es.Router.PathPrefix != "" {
					esMap["path-prefix"] = es.Router.PathPrefix
				}
				if len(es.Router.Middlewares) > 0 {
					esMap["middlewares"] = es.Router.Middlewares
				}
				if len(es.Router.Hidden) > 0 {
					esMap["hidden"] = es.Router.Hidden
				}
				if len(es.Router.Custom) > 0 {
					esMap["custom"] = es.Router.Custom
				}
			}
			externals[name] = esMap
		}
		result["external-service-definitions"] = externals
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
							if len(app.Routers) > 0 {
								appMap["routers"] = app.Routers
							}
							if len(app.PublishedServices) > 0 {
								appMap["published-services"] = app.PublishedServices
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
