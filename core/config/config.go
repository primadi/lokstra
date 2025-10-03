package config

import (
	"fmt"
)

// Config represents the top-level YAML configuration structure
type Config struct {
	Configs     []*GeneralConfig `yaml:"configs,omitempty"`
	Services    []*Service       `yaml:"services,omitempty"`
	Middlewares []*Middleware    `yaml:"middlewares,omitempty"`
	Servers     []*Server        `yaml:"servers,omitempty"`
}

// GeneralConfig represents general configuration key-value pairs
// Used for hardcoded middleware, services, or any components that need configuration
type GeneralConfig struct {
	Name  string `yaml:"name"`  // Configuration key
	Value any    `yaml:"value"` // Configuration value (can be string, number, bool, object, etc.)
}

// Service configuration
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
	Name         string `yaml:"name"`
	BaseUrl      string `yaml:"baseUrl,omitempty"`       // Base URL of the server
	DeploymentID string `yaml:"deployment-id,omitempty"` // Deployment ID for grouping servers
	Apps         []*App `yaml:"apps"`
}

// App configuration within a server
type App struct {
	Name         string   `yaml:"name"`
	Addr         string   `yaml:"addr"`
	ListenerType string   `yaml:"listener-type,omitempty"` // default: "default"
	Routers      []string `yaml:"routers,omitempty"`       // router names
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
