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
	Routers     []*Router        `yaml:"routers,omitempty" json:"routers,omitempty"`
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
func (sc ServicesConfig) MarshalYAML() (any, error) {
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

// Flatten converts layered services to flat array.
// This is the preferred method for processing services, as it provides
// a consistent interface regardless of the configuration format.
// Returns simple array if already flat, or flattens layered structure while preserving order.
func (sc *ServicesConfig) Flatten() []*Service {
	if sc.IsSimple() {
		return sc.Simple
	}

	if sc.IsLayered() {
		var flattened []*Service
		for _, layerName := range sc.Order {
			flattened = append(flattened, sc.Layered[layerName]...)
		}
		return flattened
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
	Name       string         `yaml:"name" json:"name"`
	Type       string         `yaml:"type" json:"type"`
	Enable     *bool          `yaml:"enable,omitempty" json:"enable,omitempty"` // default: true
	DependsOn  []string       `yaml:"depends-on,omitempty" json:"depends-on,omitempty"`
	Config     map[string]any `yaml:"config,omitempty" json:"config,omitempty"`
	AutoRouter *AutoRouter    `yaml:"auto-router,omitempty" json:"auto-router,omitempty"` // Auto-router configuration
}

// AutoRouter configuration for auto-generating routes from service
type AutoRouter struct {
	Convention         string           `yaml:"convention,omitempty" json:"convention,omitempty"`                     // Convention name (e.g., "rest", "rpc")
	PathPrefix         string           `yaml:"path-prefix,omitempty" json:"path-prefix,omitempty"`                   // Path prefix for routes
	ResourceName       string           `yaml:"resource-name,omitempty" json:"resource-name,omitempty"`               // Resource name (singular) for convention (e.g., "user", "person")
	PluralResourceName string           `yaml:"plural-resource-name,omitempty" json:"plural-resource-name,omitempty"` // Plural resource name override (e.g., "people" for "person")
	Routes             []*RouteOverride `yaml:"routes,omitempty" json:"routes,omitempty"`                             // Route overrides
}

// RouteOverride allows overriding specific route behavior
type RouteOverride struct {
	Name   string `yaml:"name" json:"name"`                         // Function/method name
	Method string `yaml:"method,omitempty" json:"method,omitempty"` // HTTP method override
	Path   string `yaml:"path,omitempty" json:"path,omitempty"`     // Path override
}

// Middleware configuration
type Middleware struct {
	Name   string         `yaml:"name" json:"name"`
	Type   string         `yaml:"type" json:"type"`
	Enable *bool          `yaml:"enable,omitempty" json:"enable,omitempty"` // default: true
	Config map[string]any `yaml:"config,omitempty" json:"config,omitempty"`
}

// Router configuration
type Router struct {
	Name        string   `yaml:"name" json:"name"`
	PathPrefix  string   `yaml:"path-prefix,omitempty" json:"path-prefix,omitempty"` // Path prefix (prepended to code prefix)
	Middlewares []string `yaml:"middlewares,omitempty" json:"middlewares,omitempty"` // Middleware names to apply
}

// Server configuration
type Server struct {
	Name         string `yaml:"name" json:"name"`
	BaseUrl      string `yaml:"base-url,omitempty" json:"base-url,omitempty"`           // Base URL of the server
	DeploymentID string `yaml:"deployment-id,omitempty" json:"deployment-id,omitempty"` // Deployment ID for grouping servers
	Apps         []*App `yaml:"apps" json:"apps"`
}

// ReverseProxyRewrite represents path rewrite rules for reverse proxy
type ReverseProxyRewrite struct {
	From string `yaml:"from" json:"from"` // Pattern to match in path (regex supported)
	To   string `yaml:"to" json:"to"`     // Replacement pattern
}

// ReverseProxyConfig represents reverse proxy configuration for an app
type ReverseProxyConfig struct {
	Prefix      string               `yaml:"prefix" json:"prefix"`                                 // URL prefix to match (e.g., "/api")
	StripPrefix bool                 `yaml:"strip-prefix,omitempty" json:"strip-prefix,omitempty"` // Whether to strip the prefix before forwarding
	Target      string               `yaml:"target" json:"target"`                                 // Target backend URL (e.g., "http://api-server:8080")
	Rewrite     *ReverseProxyRewrite `yaml:"rewrite,omitempty" json:"rewrite,omitempty"`           // Path rewrite rules
}

// App configuration within a server
type App struct {
	Name           string                `yaml:"name" json:"name,omitempty"`
	Addr           string                `yaml:"addr" json:"addr"`
	ListenerType   string                `yaml:"listener-type,omitempty" json:"listener-type,omitempty"`     // default: "default"
	Services       []string              `yaml:"services,omitempty" json:"services,omitempty"`               // service names deployed in this app
	Routers        []string              `yaml:"routers,omitempty" json:"routers,omitempty"`                 // router names
	ReverseProxies []*ReverseProxyConfig `yaml:"reverse-proxies,omitempty" json:"reverse-proxies,omitempty"` // reverse proxy configurations
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

// Service convention helper methods

func (s *Service) GetConvention(globalDefault string) string {
	// Priority: AutoRouter > legacy field > global default > fallback
	if s.AutoRouter != nil && s.AutoRouter.Convention != "" {
		return s.AutoRouter.Convention
	}
	if globalDefault != "" {
		return globalDefault
	}
	return "rest" // Final fallback
}

func (s *Service) GetPathPrefix() string {
	// Priority: AutoRouter > legacy field
	if s.AutoRouter != nil && s.AutoRouter.PathPrefix != "" {
		return s.AutoRouter.PathPrefix
	}
	return "" // Default is empty
}

func (s *Service) GetResourceName() string {
	// Priority: AutoRouter > derived from type
	if s.AutoRouter != nil && s.AutoRouter.ResourceName != "" {
		return s.AutoRouter.ResourceName
	}
	// Extract resource name from service type if not explicitly set
	// E.g., "user_service" -> "user"
	if s.Type != "" {
		return extractResourceNameFromType(s.Type)
	}
	return s.Name
}

func (s *Service) GetPluralResourceName() string {
	// Only check AutoRouter (no legacy field for plural)
	if s.AutoRouter != nil && s.AutoRouter.PluralResourceName != "" {
		return s.AutoRouter.PluralResourceName
	}
	return "" // Empty means will be auto-pluralized from ResourceName
}

func (s *Service) GetRouteOverrides() []*RouteOverride {
	if s.AutoRouter != nil {
		return s.AutoRouter.Routes
	}
	return nil
}
