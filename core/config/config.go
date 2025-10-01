package config

import (
	"fmt"
)

// Config represents the top-level YAML configuration structure
type Config struct {
	Configs     []GeneralConfig `yaml:"configs,omitempty"` // General configuration values
	Routers     []Router        `yaml:"routers,omitempty"`
	Services    []Service       `yaml:"services,omitempty"`
	Middlewares []Middleware    `yaml:"middlewares,omitempty"`
	Servers     []Server        `yaml:"servers,omitempty"`
}

// GeneralConfig represents general configuration key-value pairs
// Used for hardcoded middleware, services, or any components that need configuration
type GeneralConfig struct {
	Name  string `yaml:"name"`  // Configuration key
	Value any    `yaml:"value"` // Configuration value (can be string, number, bool, object, etc.)
}

// Router configuration
type Router struct {
	Name       string   `yaml:"name"`
	EngineType string   `yaml:"engine-type,omitempty"` // default: "default"
	Enable     *bool    `yaml:"enable,omitempty"`      // default: true
	Use        []string `yaml:"use,omitempty"`         // middleware names to ADD to router
	Routes     []Route  `yaml:"routes,omitempty"`
}

// Route configuration - ONLY for middleware management on existing routes
// Routes in YAML config can ONLY:
// 1. Add middleware to existing routes (use) - middleware will be ADDED to existing route middleware
// 2. Disable existing routes (enable: false)
//
// Path and Method are NOT configurable - routes must be registered in code first.
// Middleware inheritance is ALWAYS additive - no override capability to avoid confusion.
type Route struct {
	Name   string   `yaml:"name"`             // Required: must match route name registered in code
	Enable *bool    `yaml:"enable,omitempty"` // default: true, set false to disable route
	Use    []string `yaml:"use,omitempty"`    // middleware names to ADD to existing route middleware
} // Service configuration
type Service struct {
	Name   string         `yaml:"name"`
	Type   string         `yaml:"type"`
	Enable *bool          `yaml:"enable,omitempty"` // default: true
	Config map[string]any `yaml:"config,omitempty"`
}

// Middleware configuration
type Middleware struct {
	Name   string         `yaml:"name"`
	Type   string         `yaml:"type"`
	Enable *bool          `yaml:"enable,omitempty"` // default: true
	Config map[string]any `yaml:"config,omitempty"`
}

// Server configuration
type Server struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Services    []string `yaml:"services,omitempty"` // service names
	Apps        []App    `yaml:"apps"`
}

// App configuration within a server
type App struct {
	Name           string         `yaml:"name"`
	Addr           string         `yaml:"addr"`
	ListenerType   string         `yaml:"listener-type,omitempty"` // default: "default"
	Routers        []string       `yaml:"routers,omitempty"`       // router names
	ReverseProxies []ReverseProxy `yaml:"reverse-proxies,omitempty"`
}

// ReverseProxy configuration
type ReverseProxy struct {
	Path        string `yaml:"path"`
	StripPrefix string `yaml:"strip-prefix,omitempty"` // default: ""
	Target      string `yaml:"target"`
}

// Helper functions for default values
func (r *Router) IsEnabled() bool {
	if r.Enable == nil {
		return true // default value
	}
	return *r.Enable
}

func (r *Router) GetEngineType() string {
	if r.EngineType == "" {
		return "default"
	}
	return r.EngineType
}

// ShouldOverrideParentMw removed - middleware is always additive

func (rt *Route) IsEnabled() bool {
	if rt.Enable == nil {
		return true // default value
	}
	return *rt.Enable
}

// GetMethod removed - method is not configurable via YAML

// ShouldOverrideParentMw removed - middleware is always additive

func (s *Service) IsEnabled() bool {
	if s.Enable == nil {
		return true // default value
	}
	return *s.Enable
}

func (m *Middleware) IsEnabled() bool {
	if m.Enable == nil {
		return true // default value
	}
	return *m.Enable
}

func (a *App) GetListenerType() string {
	if a.ListenerType == "" {
		return "default"
	}
	return a.ListenerType
}

func (rp *ReverseProxy) GetStripPrefix() string {
	return rp.StripPrefix // empty string is the default
}

// Validation methods
func (c *Config) Validate() error {
	// Validate general configs
	configNames := make(map[string]bool)
	for _, config := range c.Configs {
		if config.Name == "" {
			return fmt.Errorf("config name cannot be empty")
		}
		if configNames[config.Name] {
			return fmt.Errorf("duplicate config name: %s", config.Name)
		}
		configNames[config.Name] = true
	}

	// Validate routers
	routerNames := make(map[string]bool)
	for _, router := range c.Routers {
		if router.Name == "" {
			return fmt.Errorf("router name cannot be empty")
		}
		if routerNames[router.Name] {
			return fmt.Errorf("duplicate router name: %s", router.Name)
		}
		routerNames[router.Name] = true

		// Validate routes within router
		routeNames := make(map[string]bool)
		for _, route := range router.Routes {
			if route.Name == "" {
				return fmt.Errorf("route name cannot be empty in router %s", router.Name)
			}

			// Check for duplicate route names (route names must be unique within router)
			if routeNames[route.Name] {
				return fmt.Errorf("duplicate route name '%s' in router %s", route.Name, router.Name)
			}
			routeNames[route.Name] = true

			// Path and Method validation removed - not configurable via YAML
			// Routes in config only manage middleware and enable/disable functionality
		}
	}

	// Validate services
	serviceNames := make(map[string]bool)
	for _, service := range c.Services {
		if service.Name == "" {
			return fmt.Errorf("service name cannot be empty")
		}
		if service.Type == "" {
			return fmt.Errorf("service type cannot be empty for service %s", service.Name)
		}
		if serviceNames[service.Name] {
			return fmt.Errorf("duplicate service name: %s", service.Name)
		}
		serviceNames[service.Name] = true
	}

	// Validate middlewares
	middlewareNames := make(map[string]bool)
	for _, middleware := range c.Middlewares {
		if middleware.Name == "" {
			return fmt.Errorf("middleware name cannot be empty")
		}
		if middleware.Type == "" {
			return fmt.Errorf("middleware type cannot be empty for middleware %s", middleware.Name)
		}
		if middlewareNames[middleware.Name] {
			return fmt.Errorf("duplicate middleware name: %s", middleware.Name)
		}
		middlewareNames[middleware.Name] = true
	}

	// Validate servers
	serverNames := make(map[string]bool)
	for _, server := range c.Servers {
		if server.Name == "" {
			return fmt.Errorf("server name cannot be empty")
		}
		if serverNames[server.Name] {
			return fmt.Errorf("duplicate server name: %s", server.Name)
		}
		serverNames[server.Name] = true

		// Validate apps within server
		appNames := make(map[string]bool)
		addrs := make(map[string]bool)
		for _, app := range server.Apps {
			if app.Name == "" {
				return fmt.Errorf("app name cannot be empty in server %s", server.Name)
			}
			if app.Addr == "" {
				return fmt.Errorf("app addr cannot be empty for app %s in server %s", app.Name, server.Name)
			}
			if appNames[app.Name] {
				return fmt.Errorf("duplicate app name %s in server %s", app.Name, server.Name)
			}
			if addrs[app.Addr] {
				return fmt.Errorf("duplicate app addr %s in server %s", app.Addr, server.Name)
			}
			appNames[app.Name] = true
			addrs[app.Addr] = true

			// Validate reverse proxies
			proxyPaths := make(map[string]bool)
			for _, proxy := range app.ReverseProxies {
				if proxy.Path == "" {
					return fmt.Errorf("reverse proxy path cannot be empty in app %s", app.Name)
				}
				if proxy.Target == "" {
					return fmt.Errorf("reverse proxy target cannot be empty for path %s in app %s", proxy.Path, app.Name)
				}
				if proxyPaths[proxy.Path] {
					return fmt.Errorf("duplicate reverse proxy path %s in app %s", proxy.Path, app.Name)
				}
				proxyPaths[proxy.Path] = true
			}
		}

		// Validate service references
		for _, serviceName := range server.Services {
			if !serviceNames[serviceName] {
				return fmt.Errorf("server %s references undefined service: %s", server.Name, serviceName)
			}
		}
	}

	// Cross-validate router references in apps
	for _, server := range c.Servers {
		for _, app := range server.Apps {
			for _, routerName := range app.Routers {
				if !routerNames[routerName] {
					return fmt.Errorf("app %s in server %s references undefined router: %s", app.Name, server.Name, routerName)
				}
			}
		}
	}

	// Cross-validate middleware references in routers
	for _, router := range c.Routers {
		for _, mwName := range router.Use {
			if !middlewareNames[mwName] {
				return fmt.Errorf("router %s references undefined middleware: %s", router.Name, mwName)
			}
		}
		for _, route := range router.Routes {
			for _, mwName := range route.Use {
				if !middlewareNames[mwName] {
					return fmt.Errorf("route %s in router %s references undefined middleware: %s", route.Name, router.Name, mwName)
				}
			}
		}
	}

	return nil
}
