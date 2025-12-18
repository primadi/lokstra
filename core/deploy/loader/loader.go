package loader

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/deploy/loader/internal"
	"github.com/primadi/lokstra/core/deploy/loader/resolver"
	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// loadConfig loads a deployment configuration from YAML file(s)
// Supports single file or multiple files that will be merged
// Paths can be files or folders - folders will be expanded to all *.yaml files
// This is a private function - external code should use LoadAndBuild instead
func loadConfig(paths ...string) (*schema.DeployConfig, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no config files specified")
	}

	var merged *schema.DeployConfig

	basePath := utils.GetBasePath()

	// STEP 0: Expand folders to files
	var expandedPaths []string
	for _, path := range paths {
		// If path is already absolute, use it directly; otherwise join with basePath
		normPath := path
		if !filepath.IsAbs(path) {
			normPath = filepath.Join(basePath, path)
		}

		// Check if path is a directory
		info, err := os.Stat(normPath)
		if err != nil {
			return nil, fmt.Errorf("failed to access %s: %w", path, err)
		}

		if info.IsDir() {
			// Expand directory to *.yaml files
			yamlFiles, err := filepath.Glob(filepath.Join(normPath, "*.yaml"))
			if err != nil {
				return nil, fmt.Errorf("failed to scan directory %s: %w", path, err)
			}
			if len(yamlFiles) == 0 {
				return nil, fmt.Errorf("no YAML files found in directory: %s", path)
			}
			expandedPaths = append(expandedPaths, yamlFiles...)
		} else {
			// It's a file, use as is
			expandedPaths = append(expandedPaths, normPath)
		}
	}

	// STEP 1: Load and merge all files (RAW, no resolution yet)
	for _, normPath := range expandedPaths {
		config, err := loadSingleFileRaw(normPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", normPath, err)
		}

		if merged == nil {
			merged = config
		} else {
			merged = mergeConfigs(merged, config)
		}
	}

	// STEP 2: Normalize shorthand servers (must be before getting server key)
	normalizeShorthandServers(merged)

	// STEP 3: Resolve configs.server to know which deployment/server to use
	if serverKeyRaw, ok := merged.Configs["server"]; ok {
		if serverKeyStr, ok := serverKeyRaw.(string); ok {
			resolved := resolver.ResolveSingleValue(serverKeyStr)
			merged.Configs["server"] = resolved
		}
	}

	// STEP 4: Apply config overrides (deployment → server)
	applyConfigOverrides(merged)

	// STEP 5: Marshal back to YAML for 2-phase resolution
	dataWithOverrides, err := yaml.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal merged config: %w", err)
	}

	// STEP 6: Resolve all ${...} EXCEPT ${@cfg:...}
	step1Data := resolver.ResolveYAMLBytesStep1(dataWithOverrides)

	var tempConfig schema.DeployConfig
	decoder := yaml.NewDecoder(bytes.NewReader(step1Data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&tempConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML (step 1): %w", err)
	}

	// STEP 7: Resolve ${@cfg:...} using configs from step 1
	step1Bytes, err := yaml.Marshal(&tempConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config for step 2: %w", err)
	}

	step2Data := resolver.ResolveYAMLBytesStep2(step1Bytes, tempConfig.Configs)

	// STEP 8: Final decode with all values resolved
	var finalConfig schema.DeployConfig
	decoder2 := yaml.NewDecoder(bytes.NewReader(step2Data))
	decoder2.KnownFields(true)
	if err := decoder2.Decode(&finalConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML (step 2): %w", err)
	}

	// STEP 9: Normalize server definitions (convert helper fields to apps)
	normalizeServerDefinitions(&finalConfig)

	// STEP 10: Validate final config
	if err := ValidateConfig(&finalConfig); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &finalConfig, nil
}

// loadSingleFileRaw loads a single YAML file WITHOUT any resolution
// Just parse the raw YAML structure
func loadSingleFileRaw(path string) (*schema.DeployConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var config schema.DeployConfig
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// applyConfigOverrides applies deployment and server config overrides to configs
func applyConfigOverrides(config *schema.DeployConfig) {
	if config.Configs == nil {
		return
	}

	// Get target server from configs.server
	serverKey, ok := config.Configs["server"].(string)
	if !ok || serverKey == "" {
		return
	}

	// Parse deployment.server
	var deploymentName, serverName string
	parts := strings.Split(serverKey, ".")
	if len(parts) == 2 {
		deploymentName = strings.ToLower(parts[0])
		serverName = strings.ToLower(parts[1])
	} else if len(parts) == 1 {
		deploymentName = "default"
		serverName = strings.ToLower(parts[0])
	} else {
		return
	}

	// Find deployment
	depDef, ok := config.Deployments[deploymentName]
	if !ok {
		return
	}

	// Apply deployment-level overrides
	for key, value := range depDef.ConfigOverrides {
		config.Configs[key] = value
	}

	// Find server
	serverDef, ok := depDef.Servers[serverName]
	if !ok {
		return
	}

	// Apply server-level overrides (highest priority)
	for key, value := range serverDef.ConfigOverrides {
		config.Configs[key] = value
	}
}

// mergeConfigs merges two configurations (target <- source)
// Source values override target values
func mergeConfigs(target, source *schema.DeployConfig) *schema.DeployConfig {
	result := &schema.DeployConfig{
		Configs:               mergeMap(target.Configs, source.Configs),
		DbPoolDefinitions:     mergeMaps(target.DbPoolDefinitions, source.DbPoolDefinitions),
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
