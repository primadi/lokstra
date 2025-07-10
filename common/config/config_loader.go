package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func LoadConfigDir(dir string) (*LokstraConfig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read config dir: %w", err)
	}

	cfg := &LokstraConfig{}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", path, err)
		}

		expanded := ExpandVariables(string(data))

		var raw map[string]any
		if err := yaml.Unmarshal([]byte(expanded), &raw); err != nil {
			return nil, fmt.Errorf("unmarshal yaml %s: %w", path, err)
		}

		if err := mergePartialYaml(cfg, raw); err != nil {
			return nil, fmt.Errorf("merge yaml %s: %w", path, err)
		}
	}

	return cfg, nil
}

func mergePartialYaml(cfg *LokstraConfig, raw map[string]any) error {
	if rawApps, ok := raw["apps"]; ok {
		data, _ := yaml.Marshal(rawApps)
		var apps []*AppConfig
		if err := yaml.Unmarshal(data, &apps); err != nil {
			return fmt.Errorf("unmarshal apps: %w", err)
		}
		cfg.Apps = mergeApps(cfg.Apps, apps)
	}

	if rawServices, ok := raw["services"]; ok {
		data, _ := yaml.Marshal(rawServices)
		var services []*ServiceConfig
		if err := yaml.Unmarshal(data, &services); err != nil {
			return fmt.Errorf("unmarshal services: %w", err)
		}
		cfg.Services = append(cfg.Services, services...)
	}

	if rawModules, ok := raw["modules"]; ok {
		data, _ := yaml.Marshal(rawModules)
		var modules []*ModuleConfig
		if err := yaml.Unmarshal(data, &modules); err != nil {
			return fmt.Errorf("unmarshal modules: %w", err)
		}
		cfg.Modules = append(cfg.Modules, modules...)
	}

	if rawServer, ok := raw["server"]; ok {
		data, _ := yaml.Marshal(rawServer)
		var server ServerConfig
		if err := yaml.Unmarshal(data, &server); err != nil {
			return fmt.Errorf("unmarshal server: %w", err)
		}
		cfg.Server = &server
	}

	return nil
}

func mergeApps(existing, incoming []*AppConfig) []*AppConfig {
	byName := map[string]*AppConfig{}
	for _, app := range existing {
		byName[app.Name] = app
	}
	for _, app := range incoming {
		if ex, found := byName[app.Name]; found {
			ex.Routes = append(ex.Routes, app.Routes...)
			ex.Middleware = append(ex.Middleware, app.Middleware...)
			if ex.Settings == nil {
				ex.Settings = map[string]any{}
			}
			maps.Copy(ex.Settings, app.Settings)
		} else {
			existing = append(existing, app)
			byName[app.Name] = app
		}
	}
	return existing
}
