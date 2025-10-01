package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFile(t *testing.T) {
	// Create temporary config file
	configContent := `
routers:
  - name: test-router
    routes:
      - name: test-route
        path: /test
        handler: TestHandler

services:
  - name: test-service
    type: memory

middlewares:
  - name: test-middleware
    type: logger

servers:
  - name: test-server
    apps:
      - name: test-app
        addr: ":8080"
        routers: [test-router]
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Test loading
	var cfg Config
	err = LoadConfigFile(tmpFile.Name(), &cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify content
	if len(cfg.Routers) != 1 {
		t.Errorf("Expected 1 router, got %d", len(cfg.Routers))
	}

	if cfg.Routers[0].Name != "test-router" {
		t.Errorf("Expected router name 'test-router', got %s", cfg.Routers[0].Name)
	}

	if len(cfg.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(cfg.Services))
	}

	if len(cfg.Middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(cfg.Middlewares))
	}

	if len(cfg.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(cfg.Servers))
	}
}

func TestLoadConfigDir(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "test-config-dir-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple config files
	routersConfig := `
routers:
  - name: router1
    routes:
      - name: route1
        path: /route1
        handler: Handler1
`

	servicesConfig := `
services:
  - name: service1
    type: memory
  - name: service2
    type: file
`

	middlewaresConfig := `
middlewares:
  - name: middleware1
    type: logger
  - name: middleware2
    type: cors
`

	serversConfig := `
servers:
  - name: server1
    services: [service1, service2]
    apps:
      - name: app1
        addr: ":8080"
        routers: [router1]
`

	// Write files
	files := map[string]string{
		"routers.yaml":     routersConfig,
		"services.yaml":    servicesConfig,
		"middlewares.yaml": middlewaresConfig,
		"servers.yaml":     serversConfig,
	}

	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test loading directory
	var cfg Config
	err = LoadConfigDir(tmpDir, &cfg)
	if err != nil {
		t.Fatalf("Failed to load config dir: %v", err)
	}

	// Verify merged content
	if len(cfg.Routers) != 1 {
		t.Errorf("Expected 1 router, got %d", len(cfg.Routers))
	}

	if len(cfg.Services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(cfg.Services))
	}

	if len(cfg.Middlewares) != 2 {
		t.Errorf("Expected 2 middlewares, got %d", len(cfg.Middlewares))
	}

	if len(cfg.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(cfg.Servers))
	}

	// Verify cross-references
	server := cfg.Servers[0]
	if len(server.Services) != 2 {
		t.Errorf("Expected server to reference 2 services, got %d", len(server.Services))
	}

	if len(server.Apps) != 1 {
		t.Errorf("Expected 1 app in server, got %d", len(server.Apps))
	}

	app := server.Apps[0]
	if len(app.Routers) != 1 {
		t.Errorf("Expected app to reference 1 router, got %d", len(app.Routers))
	}
}

func TestApplyAllConfig(t *testing.T) {
	t.Skip("Skipping ApplyAllConfig test - requires registry setup")

	// Test with missing server
	validConfig := Config{}
	err := ApplyAllConfig(&validConfig, "missing-server")
	if err == nil {
		t.Error("Expected error for missing server, got nil")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: Config{
				Routers: []Router{
					{
						Name: "test-router",
						Routes: []Route{
							{Name: "test-route", Use: []string{"test-middleware"}},
						},
					},
				},
				Services: []Service{
					{Name: "test-service", Type: "memory"},
				},
				Middlewares: []Middleware{
					{Name: "test-middleware", Type: "logger"},
				},
				Servers: []Server{
					{
						Name: "test-server",
						Apps: []App{
							{
								Name:    "test-app",
								Addr:    ":8080",
								Routers: []string{"test-router"},
							},
						},
						Services: []string{"test-service"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "duplicate router names",
			config: Config{
				Routers: []Router{
					{Name: "router1"},
					{Name: "router1"},
				},
			},
			expectError: true,
			errorMsg:    "duplicate router name",
		},
		{
			name: "empty router name",
			config: Config{
				Routers: []Router{
					{Name: ""},
				},
			},
			expectError: true,
			errorMsg:    "router name cannot be empty",
		},
		{
			name: "undefined router reference",
			config: Config{
				Servers: []Server{
					{
						Name: "test-server",
						Apps: []App{
							{
								Name:    "test-app",
								Addr:    ":8080",
								Routers: []string{"undefined-router"},
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "undefined router",
		},
		{
			name: "undefined service reference",
			config: Config{
				Servers: []Server{
					{
						Name:     "test-server",
						Services: []string{"undefined-service"},
						Apps: []App{
							{Name: "test-app", Addr: ":8080"},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "undefined service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error, got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

func TestDefaultValues(t *testing.T) {
	router := Router{}

	// Test router defaults
	if !router.IsEnabled() {
		t.Error("Router should be enabled by default")
	}

	if router.GetEngineType() != "default" {
		t.Errorf("Router engine type should default to 'default', got %s", router.GetEngineType())
	}

	// OverrideParentMw removed - middleware is always additive

	route := Route{}

	// Test route defaults
	if !route.IsEnabled() {
		t.Error("Route should be enabled by default")
	}

	// Method and override-parent-mw removed - not configurable via YAML

	service := Service{}

	// Test service defaults
	if !service.IsEnabled() {
		t.Error("Service should be enabled by default")
	}

	middleware := Middleware{}

	// Test middleware defaults
	if !middleware.IsEnabled() {
		t.Error("Middleware should be enabled by default")
	}

	app := App{}

	// Test app defaults
	if app.GetListenerType() != "default" {
		t.Errorf("App listener type should default to 'default', got %s", app.GetListenerType())
	}

	rp := ReverseProxy{}

	// Test reverse proxy defaults
	if rp.GetStripPrefix() != "" {
		t.Errorf("Reverse proxy strip prefix should default to empty string, got %s", rp.GetStripPrefix())
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(substr) > 0 && len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				func() bool {
					for i := 0; i <= len(s)-len(substr); i++ {
						if s[i:i+len(substr)] == substr {
							return true
						}
					}
					return false
				}())))
}
