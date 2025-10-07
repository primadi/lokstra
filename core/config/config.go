package config

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level YAML configuration structure
type Config struct {
	Configs     []*GeneralConfig `yaml:"configs,omitempty" json:"configs,omitempty"`
	Services    ServicesConfig   `yaml:"services,omitempty" json:"services,omitempty"`
	Middlewares []*Middleware    `yaml:"middlewares,omitempty" json:"middlewares,omitempty"`
	Servers     []*Server        `yaml:"servers,omitempty" json:"servers,omitempty"`
}

// ServicesConfig supports both simple (array) and layered (map) service definitions
type ServicesConfig struct {
	Simple  []*Service            // Simple mode: flat array of services
	Layered map[string][]*Service // Layered mode: services grouped by layer name
	Order   []string              // Order of layers (for layered mode)
}

// MarshalJSON implements custom JSON marshaling to output only the active mode
func (sc ServicesConfig) MarshalJSON() ([]byte, error) {
	if sc.IsSimple() {
		return json.Marshal(sc.Simple)
	}
	if sc.IsLayered() {
		return json.Marshal(sc.Layered)
	}
	// Empty config
	return []byte("[]"), nil
}

// UnmarshalYAML implements custom YAML unmarshaling to support both array and map formats
func (sc *ServicesConfig) UnmarshalYAML(node *yaml.Node) error {
	// Try to unmarshal as array (simple mode)
	var simpleServices []*Service
	errSimple := node.Decode(&simpleServices)
	if errSimple == nil {
		sc.Simple = simpleServices
		sc.Layered = nil
		sc.Order = nil
		return nil
	}

	// Try to unmarshal as map (layered mode)
	var layeredServices map[string][]*Service
	errLayered := node.Decode(&layeredServices)
	if errLayered == nil {
		sc.Layered = layeredServices
		sc.Simple = nil

		// Preserve order from YAML
		sc.Order = make([]string, 0, len(layeredServices))
		for i := 0; i < len(node.Content); i += 2 {
			if i < len(node.Content) {
				layerName := node.Content[i].Value
				sc.Order = append(sc.Order, layerName)
			}
		}

		return nil
	}

	// Both parsing attempts failed
	return fmt.Errorf("services must be either an array or a map; array error: %v; map error: %v", errSimple, errLayered)
}

// MarshalYAML implements custom YAML marshaling
func (sc ServicesConfig) MarshalYAML() (interface{}, error) {
	if sc.Layered != nil {
		return sc.Layered, nil
	}
	return sc.Simple, nil
}

// IsSimple returns true if using simple (array) mode
func (sc *ServicesConfig) IsSimple() bool {
	return len(sc.Simple) > 0
}

// IsLayered returns true if using layered (map) mode
func (sc *ServicesConfig) IsLayered() bool {
	return len(sc.Layered) > 0
}

// GetAllServices returns all services in registration order (simple or layered)
func (sc *ServicesConfig) GetAllServices() []*Service {
	if sc.IsSimple() {
		return sc.Simple
	}

	if sc.IsLayered() {
		var all []*Service
		for _, layerName := range sc.Order {
			all = append(all, sc.Layered[layerName]...)
		}
		return all
	}

	return nil
}

// GeneralConfig represents general configuration key-value pairs
// Used for hardcoded middleware, services, or any components that need configuration
type GeneralConfig struct {
	Name  string `yaml:"name" json:"name"`   // Configuration key
	Value any    `yaml:"value" json:"value"` // Configuration value (can be string, number, bool, object, etc.)
}

// Service configuration
type Service struct {
	Name      string         `yaml:"name" json:"name"`
	Type      string         `yaml:"type" json:"type"`
	Enable    *bool          `yaml:"enable,omitempty" json:"enable,omitempty"` // default: true
	DependsOn []string       `yaml:"depends-on,omitempty" json:"depends-on,omitempty"`
	Config    map[string]any `yaml:"config,omitempty" json:"config,omitempty"`
}

// Middleware configuration
type Middleware struct {
	Name   string         `yaml:"name" json:"name"`
	Type   string         `yaml:"type" json:"type"`
	Enable *bool          `yaml:"enable,omitempty" json:"enable,omitempty"` // default: true
	Config map[string]any `yaml:"config,omitempty" json:"config,omitempty"`
}

// Server configuration
type Server struct {
	Name         string `yaml:"name" json:"name"`
	BaseUrl      string `yaml:"baseUrl,omitempty" json:"baseUrl,omitempty"`             // Base URL of the server
	DeploymentID string `yaml:"deployment-id,omitempty" json:"deployment-id,omitempty"` // Deployment ID for grouping servers
	Apps         []*App `yaml:"apps" json:"apps"`
}

// RouterWithPrefix defines a router with its path prefix
type RouterWithPrefix struct {
	Name   string `yaml:"name" json:"name"`
	Prefix string `yaml:"prefix" json:"prefix"`
}

// App configuration within a server
type App struct {
	Name              string              `yaml:"name" json:"name,omitempty"`
	Addr              string              `yaml:"addr" json:"addr"`
	ListenerType      string              `yaml:"listener-type,omitempty" json:"listener-type,omitempty"`             // default: "default"
	Routers           []string            `yaml:"routers,omitempty" json:"routers,omitempty"`                         // router names
	RoutersWithPrefix []*RouterWithPrefix `yaml:"routers-with-prefix,omitempty" json:"routers-with-prefix,omitempty"` // routers with custom prefix
}

// creates a new empty Config
func New() *Config { return &Config{} }

// Helper functions for default values

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

func (a *App) GetName(index int) string {
	if a.Name == "" {
		return fmt.Sprintf("app-%02d", index+1)
	}
	return a.Name
}

func (s *Server) GetBaseUrl() string {
	if s.BaseUrl == "" {
		return "http://localhost"
	}
	return s.BaseUrl
}

func (s *Server) GetDeploymentID() string {
	return s.DeploymentID
}
