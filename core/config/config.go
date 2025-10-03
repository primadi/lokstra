package config

import (
	"fmt"
)

// Config represents the top-level YAML configuration structure
type Config struct {
	Configs     []*GeneralConfig `yaml:"configs,omitempty" json:"configs,omitempty"`
	Services    []*Service       `yaml:"services,omitempty" json:"services,omitempty"`
	Middlewares []*Middleware    `yaml:"middlewares,omitempty" json:"middlewares,omitempty"`
	Servers     []*Server        `yaml:"servers,omitempty" json:"servers,omitempty"`
}

// GeneralConfig represents general configuration key-value pairs
// Used for hardcoded middleware, services, or any components that need configuration
type GeneralConfig struct {
	Name  string `yaml:"name" json:"name"`   // Configuration key
	Value any    `yaml:"value" json:"value"` // Configuration value (can be string, number, bool, object, etc.)
}

// Service configuration
type Service struct {
	Name   string         `yaml:"name" json:"name"`
	Type   string         `yaml:"type" json:"type"`
	Enable *bool          `yaml:"enable,omitempty" json:"enable,omitempty"` // default: true
	Config map[string]any `yaml:"config,omitempty" json:"config,omitempty"`
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

// App configuration within a server
type App struct {
	Name         string   `yaml:"name" json:"name,omitempty"`
	Addr         string   `yaml:"addr" json:"addr"`
	ListenerType string   `yaml:"listener-type,omitempty" json:"listener-type,omitempty"` // default: "default"
	Routers      []string `yaml:"routers,omitempty" json:"routers,omitempty"`             // router names
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
