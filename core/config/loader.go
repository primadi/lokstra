package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadConfigFs loads a single YAML configuration file from any filesystem
// Variable expansion (including two-pass for CFG resolver) is handled automatically
func LoadConfigFs(fsys fs.FS, fileName string, config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Read file from filesystem
	data, err := fs.ReadFile(fsys, fileName)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", fileName, err)
	}

	// Expand all variables (two-pass expansion for CFG is automatic)
	expanded := ExpandVariables(string(data))

	// Parse YAML
	var tempConfig Config
	if err := yaml.Unmarshal([]byte(expanded), &tempConfig); err != nil {
		return fmt.Errorf("failed to parse YAML in %s: %w", fileName, err)
	}

	// Validate against JSON schema
	if err := ValidateConfig(&tempConfig); err != nil {
		return fmt.Errorf("validation failed for %s: %w", fileName, err)
	}

	// Merge with existing config
	mergeConfigs(config, &tempConfig)

	return nil
}

// LoadConfigFile loads a single YAML configuration file from OS filesystem
func LoadConfigFile(fileName string, config *Config) error {
	// Check if path is absolute
	if filepath.IsAbs(fileName) {
		// For absolute paths, use os.DirFS from the root directory
		dir := filepath.Dir(fileName)
		base := filepath.Base(fileName)
		return LoadConfigFs(os.DirFS(dir), base, config)
	}
	// Use OS filesystem for relative paths
	return LoadConfigFs(os.DirFS("."), fileName, config)
}

// LoadConfigDirFs loads and merges multiple YAML configuration files from a filesystem directory
// Variable expansion (including two-pass for CFG resolver) is handled automatically
func LoadConfigDirFs(fsys fs.FS, dirName string, config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Find all .yaml and .yml files in directory
	var yamlFiles []string
	err := fs.WalkDir(fsys, dirName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".yaml" || ext == ".yml" {
				yamlFiles = append(yamlFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory %s: %w", dirName, err)
	}

	if len(yamlFiles) == 0 {
		return fmt.Errorf("no YAML files found in directory %s", dirName)
	}

	// Sort files for consistent loading order
	sort.Strings(yamlFiles)

	// Load each file
	for _, file := range yamlFiles {
		// Read file from filesystem
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			return fmt.Errorf("failed to read config file %s: %w", file, err)
		}

		// Expand all variables (two-pass expansion for CFG is automatic)
		expanded := ExpandVariables(string(data))

		// Parse YAML
		var tempConfig Config
		if err := yaml.Unmarshal([]byte(expanded), &tempConfig); err != nil {
			return fmt.Errorf("failed to parse YAML in %s: %w", file, err)
		}

		// Validate against JSON schema
		if err := ValidateConfig(&tempConfig); err != nil {
			return fmt.Errorf("validation failed for %s: %w", file, err)
		}

		// Merge with existing config
		mergeConfigs(config, &tempConfig)
	}

	return nil
}

// LoadConfigDir loads and merges multiple YAML configuration files from an OS directory
func LoadConfigDir(dirName string, config *Config) error {
	// Use OS filesystem for backward compatibility
	return LoadConfigDirFs(os.DirFS("."), dirName, config)
}

// mergeConfigs merges source config into target config
func mergeConfigs(target *Config, source *Config) {
	// Merge general configs
	target.Configs = append(target.Configs, source.Configs...)

	// Merge services
	mergeServices(&target.Services, &source.Services)

	// Merge middlewares
	target.Middlewares = append(target.Middlewares, source.Middlewares...)

	// Merge servers
	target.Servers = append(target.Servers, source.Servers...)
}

// mergeServices merges source services into target services
func mergeServices(target *ServicesConfig, source *ServicesConfig) {
	// If target is empty (no services at all), just copy source
	if len(target.Simple) == 0 && len(target.Layered) == 0 {
		*target = *source
		return
	}

	// Both simple: merge arrays
	if target.IsSimple() && source.IsSimple() {
		target.Simple = append(target.Simple, source.Simple...)
		return
	}

	// Both layered: merge layers
	if target.IsLayered() && source.IsLayered() {
		if target.Layered == nil {
			target.Layered = make(map[string][]*Service)
		}
		for layerName, services := range source.Layered {
			target.Layered[layerName] = append(target.Layered[layerName], services...)
		}
		// Merge layer order (keep target order, add new layers from source)
		existingLayers := make(map[string]bool)
		for _, layer := range target.Order {
			existingLayers[layer] = true
		}
		for _, layer := range source.Order {
			if !existingLayers[layer] {
				target.Order = append(target.Order, layer)
			}
		}
		return
	}

	// Mixed mode: cannot merge simple and layered
	// This is a configuration error - for now, we'll keep target unchanged
	// In the future, we could panic or return an error
}
