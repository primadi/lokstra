package loader

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/deploy/loader/internal"
	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// Provider interface for custom value providers
// Providers resolve keys to values from various sources (env, aws-secret, vault, k8s, etc.)
type Provider interface {
	// Name returns the provider name (e.g., "env", "aws-secret", "vault", "k8s")
	Name() string

	// Resolve resolves a key to its value
	// Returns the resolved value and whether it was found
	Resolve(key string) (string, bool)
}

// Provider registry
var (
	providers   = make(map[string]Provider)
	providersMu sync.RWMutex
)

// RegisterProvider registers a custom provider for config resolution
// Examples:
//   - RegisterProvider(&AWSSecretProvider{}) -> resolve ${@aws-secret:key}
//   - RegisterProvider(&VaultProvider{}) -> resolve ${@vault:path}
//   - RegisterProvider(&K8sConfigMapProvider{}) -> resolve ${@k8s:configmap/key}
func RegisterProvider(p Provider) {
	providersMu.Lock()
	defer providersMu.Unlock()
	providers[p.Name()] = p
}

// getProvider retrieves a provider by name
func getProvider(name string) Provider {
	providersMu.RLock()
	defer providersMu.RUnlock()
	return providers[name]
}

// init registers default providers
func init() {
	// Register default @env provider
	RegisterProvider(&envProvider{})
}

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

	// STEP 1: Resolve all ${...} EXCEPT ${@cfg:...} at YAML byte level
	// This resolves: ${ENV_VAR}, ${@env:VAR}, ${@aws-secret:key}, etc.
	step1Data := resolveYAMLBytesStep1(data)

	// Decode to get configs first (needed for step 2)
	var tempConfig schema.DeployConfig
	decoder := yaml.NewDecoder(bytes.NewReader(step1Data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&tempConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML (step 1): %w", err)
	}

	// STEP 2: Resolve ${@cfg:...} using configs from step 1
	// Re-marshal to YAML, resolve @cfg, then unmarshal again
	step1Bytes, err := yaml.Marshal(&tempConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config for step 2: %w", err)
	}

	step2Data := resolveYAMLBytesStep2(step1Bytes, tempConfig.Configs)

	// Final decode with all values resolved
	var config schema.DeployConfig
	decoder2 := yaml.NewDecoder(bytes.NewReader(step2Data))
	decoder2.KnownFields(true)
	if err := decoder2.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML (step 2): %w", err)
	}

	return &config, nil
}

// resolveYAMLBytesStep1 resolves all ${...} placeholders EXCEPT ${@cfg:...}
// This is STEP 1 of 2-step resolution process
// Resolves: ${ENV_VAR}, ${@env:VAR}, ${@aws-secret:key}, ${@vault:path}, etc.
// Skips: ${@cfg:KEY} (needs configs map from step 1 result)
func resolveYAMLBytesStep1(data []byte) []byte {
	content := string(data)

	// Find and replace all ${...} placeholders (except ${@cfg:...})
	for {
		start := strings.Index(content, "${")
		if start == -1 {
			break
		}

		end := strings.Index(content[start:], "}")
		if end == -1 {
			// Unclosed placeholder - leave as is
			break
		}
		end += start

		placeholder := content[start+2 : end]

		// Skip @cfg placeholders (will be resolved in step 2)
		if strings.HasPrefix(placeholder, "@cfg:") {
			// Continue searching after this placeholder
			nextStart := strings.Index(content[end+1:], "${")
			if nextStart == -1 {
				break
			}
			content = content[:end+1] + content[end+1:]
			continue
		}

		// Resolve using provider registry
		resolved := resolvePlaceholder(placeholder)

		// Replace placeholder with resolved value
		content = content[:start] + resolved + content[end+1:]
	}

	return []byte(content)
}

// resolveYAMLBytesStep2 resolves ${@cfg:...} placeholders using configs map
// This is STEP 2 of 2-step resolution process
//
// Format: ${@cfg:KEY} or ${@cfg:KEY:default} or ${@cfg:'KEY:with:colons'}
//
// QUOTE ESCAPING:
// If config key contains ':' characters, wrap it in SINGLE quotes ('):
//   - ${@cfg:db.host} -> key="db.host"
//   - ${@cfg:db.host:localhost} -> key="db.host", default="localhost"
//   - ${@cfg:'db:host'} -> key="db:host" (quoted)
//   - ${@cfg:'db:url':fallback} -> key="db:url", default="fallback"
func resolveYAMLBytesStep2(data []byte, configs map[string]any) []byte {
	content := string(data)

	// Find and replace all ${@cfg:...} placeholders
	for {
		start := strings.Index(content, "${@cfg:")
		if start == -1 {
			break
		}

		end := strings.Index(content[start:], "}")
		if end == -1 {
			// Unclosed placeholder - leave as is
			break
		}
		end += start

		// Extract key: ${@cfg:KEY} -> KEY
		key := content[start+7 : end] // 7 = len("${@cfg:")

		// Parse key:default format with quote escaping support
		// Examples:
		//   "db.host" -> key="db.host", default=""
		//   "db.host:localhost" -> key="db.host", default="localhost"
		//   `'db:host'` -> key="db:host", default=""
		//   `'db:url':fallback` -> key="db:url", default="fallback"
		configKey, defaultValue := internal.ParseKeyDefault(key)

		// Lookup in configs (case-insensitive)
		var resolved string
		if val, ok := configs[strings.ToLower(configKey)]; ok {
			resolved = fmt.Sprintf("%v", val)
		} else if val, ok := configs[configKey]; ok {
			resolved = fmt.Sprintf("%v", val)
		} else if defaultValue != "" {
			resolved = defaultValue
		} else {
			// Not found - keep original for debugging
			resolved = "${@cfg:" + key + "}"
		}

		// Replace placeholder with resolved value
		content = content[:start] + resolved + content[end+1:]
	}

	return []byte(content)
}

// resolvePlaceholder resolves a placeholder using provider registry
// Formats supported:
//   - VAR_NAME -> @env provider (default)
//   - VAR_NAME:default -> @env provider with default
//   - @provider:key -> custom provider
//   - @provider:key:default -> custom provider with default
//   - @provider:'key:with:colons' -> quoted key (no default)
//   - @provider:'key:with:colons':default -> quoted key with default
//
// QUOTE ESCAPING:
// If key contains ':' characters, wrap it in SINGLE quotes (') to avoid ambiguity:
//   - Without quotes: FIRST ':' after provider is key/default separator
//   - With single quotes: ':' inside quotes is part of the key
//   - Use single quote (') not double quote (") to avoid YAML syntax conflict
//
// Examples:
//
//	${DB_HOST} -> key="DB_HOST" (env provider)
//	${DB_HOST:localhost} -> key="DB_HOST", default="localhost"
//	${@env:DB_HOST} -> explicit env provider
//	${@vault:secret/data/db:password} -> key="secret/data/db", default="password"
//	${@vault:'secret/data/db:password'} -> key="secret/data/db:password", default=""
//	${@vault:'secret/data/db:password':fallback} -> key="secret/data/db:password", default="fallback"
//	${@aws-secret:'arn:aws:secretsmanager:region:account:secret:name'} -> key contains ':'
//	${DB_URL:postgresql://localhost:5432/db} -> key="DB_URL", default="postgresql://localhost:5432/db"
//	${@cfg:'db:url'} -> key="db:url" from configs
func resolvePlaceholder(placeholder string) string {
	var providerName string
	var key string
	var defaultValue string

	// Check if it's a custom provider (@provider:key)
	if strings.HasPrefix(placeholder, "@") {
		// Format: @provider:key:default or @provider:key
		// Parse: @provider first, then key:default
		afterAt := placeholder[1:]
		firstColon := strings.Index(afterAt, ":")
		if firstColon == -1 {
			// Invalid format - no ':' after @provider
			return "${" + placeholder + "}"
		}

		providerName = afterAt[:firstColon]
		restAfterProvider := afterAt[firstColon+1:]

		// Parse key:default (default is LAST part after ':')
		key, defaultValue = internal.ParseKeyDefault(restAfterProvider)
	} else {
		// Default to @env provider
		// Format: VAR_NAME:default or VAR_NAME
		providerName = "env"
		key, defaultValue = internal.ParseKeyDefault(placeholder)
	}

	// Get provider from registry
	provider := getProvider(providerName)
	if provider == nil {
		// Provider not found - return original or default
		if defaultValue != "" {
			return defaultValue
		}
		return "${" + placeholder + "}"
	}

	// Resolve using provider
	if value, ok := provider.Resolve(key); ok {
		return value
	}

	// Use default value if provided
	if defaultValue != "" {
		return defaultValue
	}

	// Not found - return original placeholder for debugging
	return "${" + placeholder + "}"
}

// getCommandLineParam extracts value from command-line arguments
// Uses flag parsing to properly handle -KEY=value format
func getCommandLineParam(key string) (string, bool) {
	// Simple implementation - parse os.Args for -KEY=value or --KEY=value
	keyLower := strings.ToLower(key)

	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				argKey := strings.TrimPrefix(parts[0], "--")
				argKey = strings.TrimPrefix(argKey, "-")
				argKey = strings.ToLower(argKey)

				if argKey == keyLower {
					return parts[1], true
				}
			}
		}
	}

	return "", false
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
