package lokstra_registry

import "github.com/primadi/lokstra/core/deploy"

// DeploymentConfig defines a deployment topology from code
// This is the code-equivalent of YAML deployment structure
type DeploymentConfig struct {
	ConfigOverrides map[string]any
	Servers         map[string]*ServerConfig
}

// GetConfigOverrides implements deploy.DeploymentConfig interface
func (d *DeploymentConfig) GetConfigOverrides() map[string]any {
	if d.ConfigOverrides == nil {
		return make(map[string]any)
	}
	return d.ConfigOverrides
}

// GetServers implements deploy.DeploymentConfig interface
func (d *DeploymentConfig) GetServers() map[string]deploy.ServerConfig {
	result := make(map[string]deploy.ServerConfig, len(d.Servers))
	for name, server := range d.Servers {
		result[name] = server
	}
	return result
}

// ServerConfig defines a server in a deployment
type ServerConfig struct {
	BaseURL         string
	ConfigOverrides map[string]any // Server-level config overrides
	Apps            []*AppConfig

	// Shorthand: If only one app, you can define it directly here
	// If set, a new app will be created and prepended to Apps array
	Addr              string
	Routers           []string
	PublishedServices []string
}

// GetBaseURL implements deploy.ServerConfig interface
func (s *ServerConfig) GetBaseURL() string {
	return s.BaseURL
}

// GetConfigOverrides implements deploy.ServerConfig interface
func (s *ServerConfig) GetConfigOverrides() map[string]any {
	if s.ConfigOverrides == nil {
		return make(map[string]any)
	}
	return s.ConfigOverrides
}

// GetApps implements deploy.ServerConfig interface
func (s *ServerConfig) GetApps() []deploy.AppConfig {
	result := make([]deploy.AppConfig, len(s.Apps))
	for i, app := range s.Apps {
		result[i] = app
	}
	return result
}

// GetAddr implements deploy.ServerConfig interface
func (s *ServerConfig) GetAddr() string {
	return s.Addr
}

// GetRouters implements deploy.ServerConfig interface
func (s *ServerConfig) GetRouters() []string {
	return s.Routers
}

// GetPublishedServices implements deploy.ServerConfig interface
func (s *ServerConfig) GetPublishedServices() []string {
	return s.PublishedServices
}

// AppConfig defines an app (listener) in a server
type AppConfig struct {
	Addr              string   // e.g., ":8080", "127.0.0.1:8080", "unix:/tmp/app.sock"
	Routers           []string // Router names to include
	PublishedServices []string // Service names to auto-generate routers for
}

// GetAddr implements deploy.AppConfig interface
func (a *AppConfig) GetAddr() string {
	return a.Addr
}

// GetRouters implements deploy.AppConfig interface
func (a *AppConfig) GetRouters() []string {
	return a.Routers
}

// GetPublishedServices implements deploy.AppConfig interface
func (a *AppConfig) GetPublishedServices() []string {
	return a.PublishedServices
}
