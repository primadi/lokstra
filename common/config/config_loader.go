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
	for i, app := range cfg.Apps {
		if app.MiddlewareRaw != nil {
			mw, err := NormalizeMiddlewareConfig(app.MiddlewareRaw)
			if err != nil {
				return nil, fmt.Errorf("normalize middleware for app %s: %w", app.Name, err)
			}
			cfg.Apps[i].Middleware = mw
		}

		// Normalize middleware for routes
		for i, route := range app.Routes {
			if route.MiddlewareRaw != nil {
				mw, err := NormalizeMiddlewareConfig(route.MiddlewareRaw)
				if err != nil {
					return nil, fmt.Errorf("normalize middleware for route %s in app %s: %w",
						route.Path, app.Name, err)
				}
				app.Routes[i].Middleware = mw
			}
		}

		// Normalize middleware for groups
		if len(app.Groups) > 0 {
			if err := loadGroupConfig(&app.Groups); err != nil {
				return nil, fmt.Errorf("load group config for app %s: %w", app.Name, err)
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

func loadGroupConfig(groupArr *[]GroupConfig) error {
	for idx, group := range *groupArr {
		if group.MiddlewareRaw != nil {
			mw, err := NormalizeMiddlewareConfig(group.MiddlewareRaw)
			if err != nil {
				return fmt.Errorf("normalize middleware for group %s: %w", group.Prefix, err)
			}
			(*groupArr)[idx].Middleware = mw
		}

		for i, route := range group.Routes {
			if route.MiddlewareRaw != nil {
				mw, err := NormalizeMiddlewareConfig(route.MiddlewareRaw)
				if err != nil {
					return fmt.Errorf("normalize middleware for route %s in group %s: %w",
						route.Path, group.Prefix, err)
				}
				group.Routes[i].Middleware = mw
			}
		}

		if len(group.Groups) > 0 {
			if err := loadGroupConfig(&group.Groups); err != nil {
				return err
			}
		}
	}

	return nil
}

func mergePartialYaml(cfg *LokstraConfig, raw map[string]any) error {
	if rawApps, ok := raw["apps"]; ok {
		data, err := yaml.Marshal(rawApps)
		if err != nil {
			return fmt.Errorf("marshal apps: %w", err)
		}

		var apps []*AppConfig
		if err := yaml.Unmarshal(data, &apps); err != nil {
			return fmt.Errorf("unmarshal apps: %w", err)
		}
		cfg.Apps = mergeApps(cfg.Apps, apps)
	}

	if rawServices, ok := raw["services"]; ok {
		data, err := yaml.Marshal(rawServices)
		if err != nil {
			return fmt.Errorf("marshal services: %w", err)
		}

		var services []*ServiceConfig
		if err := yaml.Unmarshal(data, &services); err != nil {
			return fmt.Errorf("unmarshal services: %w", err)
		}
		cfg.Services = mergeServices(cfg.Services, services)
	}

	if rawModules, ok := raw["modules"]; ok {
		data, err := yaml.Marshal(rawModules)
		if err != nil {
			return fmt.Errorf("marshal modules: %w", err)
		}

		var modules []*ModuleConfig
		if err := yaml.Unmarshal(data, &modules); err != nil {
			return fmt.Errorf("unmarshal modules: %w", err)
		}
		cfg.Modules = mergeModules(cfg.Modules, modules)
	}

	if rawServer, ok := raw["server"]; ok {
		data, err := yaml.Marshal(rawServer)
		if err != nil {
			return fmt.Errorf("marshal server: %w", err)
		}
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
			if ex.MiddlewareRaw == nil {
				ex.MiddlewareRaw = []any{}
			}
			if app.MiddlewareRaw != nil {
				ex.MiddlewareRaw = append(ex.MiddlewareRaw.([]any), app.MiddlewareRaw.([]any)...)
			}
			// ex.Middleware = append(ex.Middleware, app.Middleware...)
			// if ex.Settings == nil {
			// 	ex.Settings = map[string]any{}
			// }
			maps.Copy(ex.Settings, app.Settings)
		} else {
			existing = append(existing, app)
			byName[app.Name] = app
		}
	}
	return existing
}

func mergeServices(existing, incoming []*ServiceConfig) []*ServiceConfig {
	byName := map[string]*ServiceConfig{}
	for _, svc := range existing {
		byName[svc.Name] = svc
	}
	for _, svc := range incoming {
		if ex, found := byName[svc.Name]; found {
			// Merge logic if needed, e.g., settings
			if ex.Config == nil {
				ex.Config = map[string]any{}
			}
			maps.Copy(ex.Config, svc.Config)
		} else {
			existing = append(existing, svc)
			byName[svc.Name] = svc
		}
	}
	return existing
}

func mergeModules(existing, incoming []*ModuleConfig) []*ModuleConfig {
	byName := map[string]*ModuleConfig{}
	for _, mod := range existing {
		byName[mod.Name] = mod
	}
	for _, mod := range incoming {
		if ex, found := byName[mod.Name]; found {
			// Merge logic if needed
			if ex.Settings == nil {
				ex.Settings = map[string]any{}
			}
			maps.Copy(ex.Settings, mod.Settings)
		} else {
			existing = append(existing, mod)
			byName[mod.Name] = mod
		}
	}
	return existing
}
