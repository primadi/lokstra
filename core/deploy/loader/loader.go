package loader

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/deploy/loader/internal"
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
	// Load and merge each file
	for _, path := range paths {
		// If path is already absolute, use it directly; otherwise join with basePath
		normPath := path
		if !filepath.IsAbs(path) {
			normPath = filepath.Join(basePath, path)
		}
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
		Configs:               mergeMap(target.Configs, source.Configs),
		NamedDbPools:          mergeMaps(target.NamedDbPools, source.NamedDbPools),
		MiddlewareDefinitions: mergeMaps(target.MiddlewareDefinitions, source.MiddlewareDefinitions),
		ServiceDefinitions:    mergeMaps(target.ServiceDefinitions, source.ServiceDefinitions),
		RouterDefinitions:     mergeMaps(target.RouterDefinitions, source.RouterDefinitions), // Renamed from Routers
		Deployments:           mergeMaps(target.Deployments, source.Deployments),
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
	configMap := internal.ConfigToMap(config)
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
