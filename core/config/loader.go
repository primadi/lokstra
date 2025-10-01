package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/primadi/lokstra/lokstra_registry"
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

	// Validate merged config
	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed after loading %s: %w", fileName, err)
	}

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

	// Validate merged config once at the end
	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed after loading directory %s: %w", dirName, err)
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

	// Merge routers
	target.Routers = append(target.Routers, source.Routers...)

	// Merge services
	target.Services = append(target.Services, source.Services...)

	// Merge middlewares
	target.Middlewares = append(target.Middlewares, source.Middlewares...)

	// Merge servers
	target.Servers = append(target.Servers, source.Servers...)
}

// ApplyRoutersConfig modifies existing routers in lokstra_registry based on YAML config
func ApplyRoutersConfig(cfg *Config, routerNames ...string) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// If no router names specified, apply all routers
	var targetRouters []Router
	if len(routerNames) == 0 {
		targetRouters = cfg.Routers
	} else {
		// Filter routers by name
		routerMap := make(map[string]Router)
		for _, router := range cfg.Routers {
			routerMap[router.Name] = router
		}

		for _, name := range routerNames {
			if router, exists := routerMap[name]; exists {
				if router.IsEnabled() {
					targetRouters = append(targetRouters, router)
				}
			} else {
				return fmt.Errorf("router config not found: %s", name)
			}
		}
	}

	// Modify each router in registry
	for _, routerCfg := range targetRouters {
		if !routerCfg.IsEnabled() {
			continue
		}

		// Check if router exists in registry
		existingRouter := lokstra_registry.GetRouter(routerCfg.Name)
		if existingRouter == nil {
			panic(fmt.Sprintf("router %s not found in registry - must be registered in code first", routerCfg.Name))
		}

		// Apply router-level middleware modifications
		for _, mwName := range routerCfg.Use {
			mw := lokstra_registry.CreateMiddleware(mwName)
			if mw == nil {
				panic(fmt.Sprintf("middleware %s not found in registry for router %s", mwName, routerCfg.Name))
			}
			existingRouter.Use(mw)
		}

		// Apply route middleware modifications (routes in config can ONLY add middleware or disable routes)
		for _, routeCfg := range routerCfg.Routes {
			if !routeCfg.IsEnabled() {
				// TODO: Implement route removal/disabling
				// This would require router interface to support route removal
				continue
			}

			// Route configuration can ONLY add middleware to existing routes
			// No path/method modification - routes must be registered in code first

			// Apply middleware to route (this is the main purpose of route config)
			if len(routeCfg.Use) > 0 {
				// Get middlewares for this route
				var routeMiddlewares []any
				for _, mwName := range routeCfg.Use {
					mw := lokstra_registry.CreateMiddleware(mwName)
					if mw == nil {
						panic(fmt.Sprintf("middleware %s not found in registry for route %s", mwName, routeCfg.Name))
					}
					routeMiddlewares = append(routeMiddlewares, mw)
				}

				// TODO: Apply route-specific middleware when router interface supports it
				// For now, we validate that middleware exists but can't apply it to specific routes
				// This requires extending the router interface to support:
				// - GetRoute(name) *route.Route
				// - AddRouteMiddleware(routeName, middleware)
			}
		}
	}

	return nil
}

// ApplyServerConfig modifies existing server in lokstra_registry based on YAML config
func ApplyServerConfig(cfg *Config, serverName string) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Find server config
	var serverCfg *Server
	for _, s := range cfg.Servers {
		if s.Name == serverName {
			serverCfg = &s
			break
		}
	}
	if serverCfg == nil {
		return fmt.Errorf("server config not found: %s", serverName)
	}

	// Check if server exists in registry
	existingServer := lokstra_registry.GetServer(serverName)
	if existingServer == nil {
		panic(fmt.Sprintf("server %s not found in registry - must be registered in code first", serverName))
	}

	// Apply services first
	if err := ApplyServicesConfig(cfg, serverCfg.Services...); err != nil {
		return fmt.Errorf("failed to apply services for server %s: %w", serverName, err)
	}

	// Validate routers referenced by apps exist in registry
	for _, appCfg := range serverCfg.Apps {
		for _, routerName := range appCfg.Routers {
			if lokstra_registry.GetRouter(routerName) == nil {
				panic(fmt.Sprintf("router %s referenced by app %s not found in registry - must be registered in code first",
					routerName, appCfg.Name))
			}
		}

		// Validate reverse proxy configurations
		for _, proxyCfg := range appCfg.ReverseProxies {
			if proxyCfg.Path == "" || proxyCfg.Target == "" {
				return fmt.Errorf("invalid reverse proxy configuration in app %s", appCfg.Name)
			}
		}
	}

	// Note: Server modification would be implemented here
	// This would require extending the server interface to support dynamic configuration
	// For now, we just validate that the server exists and configuration is valid

	return nil
}

// ApplyServicesConfig modifies existing services in lokstra_registry based on YAML config
func ApplyServicesConfig(cfg *Config, serviceNames ...string) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// If no service names specified, apply all services
	var targetServices []Service
	if len(serviceNames) == 0 {
		targetServices = cfg.Services
	} else {
		// Filter services by name
		serviceMap := make(map[string]Service)
		for _, service := range cfg.Services {
			serviceMap[service.Name] = service
		}

		for _, name := range serviceNames {
			if service, exists := serviceMap[name]; exists {
				if service.IsEnabled() {
					targetServices = append(targetServices, service)
				}
			} else {
				return fmt.Errorf("service config not found: %s", name)
			}
		}
	}

	// Apply each service configuration
	for _, serviceCfg := range targetServices {
		if !serviceCfg.IsEnabled() {
			continue
		}

		// Validate service type is not empty
		if serviceCfg.Type == "" {
			return fmt.Errorf("service %s has empty type", serviceCfg.Name)
		}

		// Register or update lazy service configuration
		lokstra_registry.RegisterLazyService(
			serviceCfg.Name,
			serviceCfg.Type,
			serviceCfg.Config,
			lokstra_registry.AllowOverride(true), // Allow override of existing services
		)
	}

	return nil
}

// ApplyMiddlewareConfig modifies existing middlewares in lokstra_registry based on YAML config
func ApplyMiddlewareConfig(cfg *Config, middlewareNames ...string) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// If no middleware names specified, apply all middlewares
	var targetMiddlewares []Middleware
	if len(middlewareNames) == 0 {
		targetMiddlewares = cfg.Middlewares
	} else {
		// Filter middlewares by name
		middlewareMap := make(map[string]Middleware)
		for _, middleware := range cfg.Middlewares {
			middlewareMap[middleware.Name] = middleware
		}

		for _, name := range middlewareNames {
			if middleware, exists := middlewareMap[name]; exists {
				if middleware.IsEnabled() {
					targetMiddlewares = append(targetMiddlewares, middleware)
				}
			} else {
				return fmt.Errorf("middleware config not found: %s", name)
			}
		}
	}

	// Apply each middleware configuration
	for _, middlewareCfg := range targetMiddlewares {
		if !middlewareCfg.IsEnabled() {
			continue
		}

		// Validate middleware type is not empty
		if middlewareCfg.Type == "" {
			return fmt.Errorf("middleware %s has empty type", middlewareCfg.Name)
		}

		// Register middleware name with config
		lokstra_registry.RegisterMiddlewareName(
			middlewareCfg.Name,
			middlewareCfg.Type,
			middlewareCfg.Config,
			lokstra_registry.AllowOverride(true), // Allow override of existing middlewares
		)
	}

	return nil
}

// ApplyGeneralConfig applies general configuration values to lokstra_registry
func ApplyGeneralConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Apply each general configuration
	for _, configEntry := range cfg.Configs {
		if configEntry.Name == "" {
			return fmt.Errorf("config name cannot be empty")
		}

		// Register configuration value (allow override for configs from YAML)
		lokstra_registry.RegisterConfig(
			configEntry.Name,
			configEntry.Value,
			lokstra_registry.AllowOverride(true),
		)
	}

	return nil
}

// ApplyAllConfig validates all configuration for a specific server
// Note: This is a placeholder implementation for configuration validation.
// In production, this would create and return an actual server instance.
func ApplyAllConfig(cfg *Config, serverName string) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Find server config to get dependencies
	var serverCfg *Server
	for _, s := range cfg.Servers {
		if s.Name == serverName {
			serverCfg = &s
			break
		}
	}
	if serverCfg == nil {
		return fmt.Errorf("server not found: %s", serverName)
	}

	// Collect all required router names from apps
	var requiredRouters []string
	for _, appCfg := range serverCfg.Apps {
		requiredRouters = append(requiredRouters, appCfg.Routers...)
	}

	// Collect all required middleware names from routers and routes
	requiredMiddlewares := make(map[string]bool)
	for _, routerName := range requiredRouters {
		for _, router := range cfg.Routers {
			if router.Name == routerName {
				// Add router-level middlewares
				for _, mwName := range router.Use {
					requiredMiddlewares[mwName] = true
				}
				// Add route-level middlewares
				for _, route := range router.Routes {
					for _, mwName := range route.Use {
						requiredMiddlewares[mwName] = true
					}
				}
			}
		}
	}

	// Convert middleware map to slice
	var middlewareNames []string
	for mwName := range requiredMiddlewares {
		middlewareNames = append(middlewareNames, mwName)
	}

	// Apply configurations in order:
	// 0. Apply general configs first (may be needed by other components)
	if err := ApplyGeneralConfig(cfg); err != nil {
		return fmt.Errorf("failed to apply general configs: %w", err)
	}

	// 1. Apply middlewares (dependencies for routers)
	if err := ApplyMiddlewareConfig(cfg, middlewareNames...); err != nil {
		return fmt.Errorf("failed to apply middlewares: %w", err)
	}

	// 2. Apply services (dependencies for server)
	if err := ApplyServicesConfig(cfg, serverCfg.Services...); err != nil {
		return fmt.Errorf("failed to apply services: %w", err)
	}

	// 3. Apply routers (dependencies for server apps) - only if router configs exist
	if len(cfg.Routers) > 0 {
		if err := ApplyRoutersConfig(cfg, requiredRouters...); err != nil {
			return fmt.Errorf("failed to apply routers: %w", err)
		}
	} else {
		// No router configuration - using middleware from code registration only
		fmt.Printf("ℹ️  No router configuration found - using middleware from code\n")
	}

	// 4. Apply server configuration
	if err := ApplyServerConfig(cfg, serverName); err != nil {
		return fmt.Errorf("failed to apply server config: %w", err)
	}

	return nil
}
