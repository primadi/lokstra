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
func LoadConfigFs(fsys fs.FS, fileName string, config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Read file from filesystem
	data, err := fs.ReadFile(fsys, fileName)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", fileName, err)
	}

	// Parse YAML
	var tempConfig Config
	if err := yaml.Unmarshal(data, &tempConfig); err != nil {
		return fmt.Errorf("failed to parse YAML in %s: %w", fileName, err)
	}

	// Merge with existing config
	mergeConfigs(config, &tempConfig)

	return nil
}

// LoadConfigFile loads a single YAML configuration file from OS filesystem
func LoadConfigFile(fileName string, config *Config) error {
	// Use OS filesystem for backward compatibility
	return LoadConfigFs(os.DirFS("."), fileName, config)
}

// LoadConfigDirFs loads and merges multiple YAML configuration files from a filesystem directory
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

	// Load each file without individual validation (will validate at the end)
	for _, file := range yamlFiles {
		// Read file from filesystem
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			return fmt.Errorf("failed to read config file %s: %w", file, err)
		}

		// Parse YAML
		var tempConfig Config
		if err := yaml.Unmarshal(data, &tempConfig); err != nil {
			return fmt.Errorf("failed to parse YAML in %s: %w", file, err)
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
	target.Services = append(target.Services, source.Services...)

	// Merge middlewares
	target.Middlewares = append(target.Middlewares, source.Middlewares...)

	// Merge servers
	target.Servers = append(target.Servers, source.Servers...)
}
