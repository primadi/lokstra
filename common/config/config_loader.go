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

	// Normalize middleware for apps, routes, and groups
	for _, app := range cfg.Apps {
		if app.MiddlewareRaw != nil {
			mw, err := NormalizeMiddlewareConfig(app.MiddlewareRaw)
			if err != nil {
				return nil, fmt.Errorf("normalize middleware for app %s: %w", app.Name, err)
			}
			app.Middleware = mw
		}

		// Normalize middleware for routes
		for _, route := range app.Routes {
			if route.MiddlewareRaw != nil {
				mw, err := NormalizeMiddlewareConfig(route.MiddlewareRaw)
				if err != nil {
					return nil, fmt.Errorf("normalize middleware for route %s in app %s: %w",
						route.Path, app.Name, err)
				}
				route.Middleware = mw
				_ = route.Middleware
			}
		}

		// Normalize middleware for groups
		for _, group := range app.Groups {
			if group.MiddlewareRaw != nil {
				mw, err := NormalizeMiddlewareConfig(group.MiddlewareRaw)
				if err != nil {
					return nil, fmt.Errorf("normalize middleware for group %s in app %s: %w",
						group.Prefix, app.Name, err)
				}
				group.Middleware = mw
				_ = group.Middleware
			}
			for _, route := range group.Routes {
				if route.MiddlewareRaw != nil {
					mw, err := NormalizeMiddlewareConfig(route.MiddlewareRaw)
					if err != nil {
						return nil, fmt.Errorf("normalize middleware for route %s in group %s of app %s: %w",
							route.Path, group.Prefix, app.Name, err)
					}
					route.Middleware = mw
					_ = route.Middleware
				}
			}
		}
	}
	if cfg.Server == nil {
		cfg.Server = &ServerConfig{
			Name:     "default",
			Settings: map[string]any{},
		}
	}
	if cfg.Server.Name == "" {
		cfg.Server.Name = "default"
	}
	if cfg.Server.Settings == nil {
		cfg.Server.Settings = map[string]any{}
	}
	if cfg.Apps == nil {
		cfg.Apps = []*AppConfig{}
	}
	if cfg.Services == nil {
		cfg.Services = []*ServiceConfig{}
	}
	if cfg.Modules == nil {
		cfg.Modules = []*ModuleConfig{}
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
