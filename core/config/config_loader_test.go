package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================
// CONFIG LOADER TESTS
// ==============================================

func TestLoadConfigDir_EmptyDirectory(t *testing.T) {
	// Create empty temp directory
	tempDir, err := os.MkdirTemp("", "config_test_empty")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	cfg, err := LoadConfigDir(tempDir)
	require.NoError(t, err)

	// Should create default config
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Server)
	assert.Equal(t, "default", cfg.Server.Name)
	assert.NotNil(t, cfg.Server.Settings)
	assert.Empty(t, cfg.Apps)
	assert.Empty(t, cfg.Services)
	assert.Empty(t, cfg.Modules)
}

func TestLoadConfigDir_NonExistentDirectory(t *testing.T) {
	cfg, err := LoadConfigDir("/non/existent/path")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to read config dir")
}

func TestLoadConfigDir_SimpleYamlFile(t *testing.T) {
	// Create temp directory with yaml file
	tempDir, err := os.MkdirTemp("", "config_test_simple")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	yamlContent := `
server:
  name: test-server
  global_setting:
    debug: true
    port: 8080

apps:
  - name: test-app
    address: ":8080"
    listener_type: http
    routes:
      - method: GET
        path: /health
        handler: health.CheckHandler

services:
  - name: test-service
    type: test-factory
    config:
      param1: value1

modules:
  - name: test-module
    path: ./test-module
    settings:
      enabled: true
`

	err = os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(yamlContent), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfigDir(tempDir)
	require.NoError(t, err)

	// Test server config
	assert.NotNil(t, cfg.Server)
	assert.Equal(t, "test-server", cfg.Server.Name)
	assert.Equal(t, true, cfg.Server.Settings["debug"])
	assert.Equal(t, 8080, cfg.Server.Settings["port"])

	// Test apps config
	require.Len(t, cfg.Apps, 1)
	app := cfg.Apps[0]
	assert.Equal(t, "test-app", app.Name)
	assert.Equal(t, ":8080", app.Address)
	assert.Equal(t, "http", app.ListenerType)
	require.Len(t, app.Routes, 1)
	assert.Equal(t, "GET", app.Routes[0].Method)
	assert.Equal(t, "/health", app.Routes[0].Path)
	assert.Equal(t, "health.CheckHandler", app.Routes[0].Handler)

	// Test services config
	require.Len(t, cfg.Services, 1)
	svc := cfg.Services[0]
	assert.Equal(t, "test-service", svc.Name)
	assert.Equal(t, "test-factory", svc.Type)
	assert.Equal(t, "value1", svc.Config["param1"])

	// Test modules config
	require.Len(t, cfg.Modules, 1)
	mod := cfg.Modules[0]
	assert.Equal(t, "test-module", mod.Name)
	assert.Equal(t, "./test-module", mod.Path)
	assert.Equal(t, true, mod.Settings["enabled"])
}

func TestLoadConfigDir_MultipleYamlFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config_test_multiple")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// First file - server and one app
	config1 := `
server:
  name: multi-server
  global_setting:
    debug: true

apps:
  - name: app1
    address: ":8080"
    routes:
      - method: GET
        path: /api/v1
        handler: api.V1Handler
`

	// Second file - another app and services
	config2 := `
apps:
  - name: app2
    address: ":8081"
    routes:
      - method: POST
        path: /api/v2
        handler: api.V2Handler

services:
  - name: database
    type: postgres
    config:
      host: localhost
      port: 5432
`

	err = os.WriteFile(filepath.Join(tempDir, "01-server.yaml"), []byte(config1), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "02-apps.yaml"), []byte(config2), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfigDir(tempDir)
	require.NoError(t, err)

	// Should merge both files
	assert.Equal(t, "multi-server", cfg.Server.Name)
	assert.Equal(t, true, cfg.Server.Settings["debug"])

	require.Len(t, cfg.Apps, 2)

	// Check app1
	app1 := cfg.Apps[0]
	assert.Equal(t, "app1", app1.Name)
	assert.Equal(t, ":8080", app1.Address)
	require.Len(t, app1.Routes, 1)
	assert.Equal(t, "GET", app1.Routes[0].Method)

	// Check app2
	app2 := cfg.Apps[1]
	assert.Equal(t, "app2", app2.Name)
	assert.Equal(t, ":8081", app2.Address)
	require.Len(t, app2.Routes, 1)
	assert.Equal(t, "POST", app2.Routes[0].Method)

	// Check services
	require.Len(t, cfg.Services, 1)
	assert.Equal(t, "database", cfg.Services[0].Name)
	assert.Equal(t, "postgres", cfg.Services[0].Type)
}

func TestLoadConfigDir_InvalidYaml(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config_test_invalid")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	invalidYaml := `
server:
  name: test
  invalid: [unclosed bracket
`

	err = os.WriteFile(filepath.Join(tempDir, "invalid.yaml"), []byte(invalidYaml), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfigDir(tempDir)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "unmarshal yaml")
}

func TestLoadConfigDir_WithVariableExpansion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config_test_vars")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set environment variable for testing
	os.Setenv("TEST_PORT", "9090")
	os.Setenv("TEST_DEBUG", "false")
	defer func() {
		os.Unsetenv("TEST_PORT")
		os.Unsetenv("TEST_DEBUG")
	}()

	yamlWithVars := `
server:
  name: ${TEST_APP_NAME:default-server}
  global_setting:
    port: ${TEST_PORT}
    debug: ${TEST_DEBUG}
    timeout: ${TEST_TIMEOUT:30}

apps:
  - name: test-app
    address: ":${TEST_PORT}"
`

	err = os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(yamlWithVars), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfigDir(tempDir)
	require.NoError(t, err)

	// Variable expansion should work
	assert.Equal(t, "default-server", cfg.Server.Name)   // default value used
	assert.Equal(t, 9090, cfg.Server.Settings["port"])   // env var used - should be int
	assert.Equal(t, false, cfg.Server.Settings["debug"]) // env var used - should be bool
	assert.Equal(t, 30, cfg.Server.Settings["timeout"])  // default value used - should be int
	assert.Equal(t, ":9090", cfg.Apps[0].Address)        // env var in address
}

func TestLoadConfigDir_WithMiddleware(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config_test_middleware")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	yamlWithMiddleware := `
apps:
  - name: middleware-app
    address: ":8080"
    middleware:
      - cors
      - name: auth
        enabled: true
        config:
          secret: mysecret
      - name: logging
        enabled: false
    routes:
      - method: GET
        path: /api
        handler: api.Handler
        middleware:
          - rate_limit
`

	err = os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(yamlWithMiddleware), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfigDir(tempDir)
	require.NoError(t, err)

	app := cfg.Apps[0]

	// Test app-level middleware
	require.Len(t, app.Middleware, 3)

	assert.Equal(t, "cors", app.Middleware[0].Name)
	assert.True(t, app.Middleware[0].Enabled)

	assert.Equal(t, "auth", app.Middleware[1].Name)
	assert.True(t, app.Middleware[1].Enabled)
	assert.Equal(t, "mysecret", app.Middleware[1].Config["secret"])

	assert.Equal(t, "logging", app.Middleware[2].Name)
	assert.False(t, app.Middleware[2].Enabled)

	// Test route-level middleware
	route := app.Routes[0]
	require.Len(t, route.Middleware, 1)
	assert.Equal(t, "rate_limit", route.Middleware[0].Name)
	assert.True(t, route.Middleware[0].Enabled)
}

func TestLoadConfigDir_NonYamlFilesIgnored(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config_test_non_yaml")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create YAML file
	yamlContent := `
server:
  name: test-server
`
	err = os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Create non-YAML files (should be ignored)
	err = os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("ignore me"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "config.json"), []byte(`{"ignore": true}`), 0644)
	require.NoError(t, err)

	// Create subdirectory (should be ignored)
	err = os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	require.NoError(t, err)

	cfg, err := LoadConfigDir(tempDir)
	require.NoError(t, err)

	// Should only process the YAML file
	assert.Equal(t, "test-server", cfg.Server.Name)
}

func TestMergeApps(t *testing.T) {
	existing := []*AppConfig{
		{
			Name:    "app1",
			Address: ":8080",
			Routes: []RouteConfig{
				{Method: "GET", Path: "/existing", Handler: "existing.Handler"},
			},
			Settings: map[string]any{"key1": "value1"},
		},
	}

	incoming := []*AppConfig{
		{
			Name:    "app1", // Same name - should merge
			Address: ":8080",
			Routes: []RouteConfig{
				{Method: "POST", Path: "/new", Handler: "new.Handler"},
			},
			Settings: map[string]any{"key2": "value2"},
		},
		{
			Name:    "app2", // New app - should add
			Address: ":8081",
			Routes: []RouteConfig{
				{Method: "GET", Path: "/app2", Handler: "app2.Handler"},
			},
		},
	}

	result := mergeApps(existing, incoming)

	require.Len(t, result, 2)

	// Check merged app1
	app1 := result[0]
	assert.Equal(t, "app1", app1.Name)
	require.Len(t, app1.Routes, 2) // Should have both routes
	assert.Equal(t, "/existing", app1.Routes[0].Path)
	assert.Equal(t, "/new", app1.Routes[1].Path)
	assert.Equal(t, "value1", app1.Settings["key1"])
	assert.Equal(t, "value2", app1.Settings["key2"])

	// Check new app2
	app2 := result[1]
	assert.Equal(t, "app2", app2.Name)
	assert.Equal(t, ":8081", app2.Address)
	require.Len(t, app2.Routes, 1)
}

func TestMergeServices(t *testing.T) {
	existing := []*ServiceConfig{
		{
			Name:   "db",
			Type:   "postgres",
			Config: map[string]any{"host": "localhost"},
		},
	}

	incoming := []*ServiceConfig{
		{
			Name:   "db", // Same name - should merge config
			Type:   "postgres",
			Config: map[string]any{"port": 5432},
		},
		{
			Name:   "cache", // New service - should add
			Type:   "redis",
			Config: map[string]any{"host": "redis-host"},
		},
	}

	result := mergeServices(existing, incoming)

	require.Len(t, result, 2)

	// Check merged db service
	db := result[0]
	assert.Equal(t, "db", db.Name)
	assert.Equal(t, "localhost", db.Config["host"])
	assert.Equal(t, 5432, db.Config["port"])

	// Check new cache service
	cache := result[1]
	assert.Equal(t, "cache", cache.Name)
	assert.Equal(t, "redis", cache.Type)
}

func TestMergeModules(t *testing.T) {
	existing := []*ModuleConfig{
		{
			Name:     "mod1",
			Path:     "./mod1",
			Settings: map[string]any{"key1": "value1"},
		},
	}

	incoming := []*ModuleConfig{
		{
			Name:     "mod1", // Same name - should merge settings
			Settings: map[string]any{"key2": "value2"},
		},
		{
			Name: "mod2", // New module - should add
			Path: "./mod2",
		},
	}

	result := mergeModules(existing, incoming)

	require.Len(t, result, 2)

	// Check merged mod1
	mod1 := result[0]
	assert.Equal(t, "mod1", mod1.Name)
	assert.Equal(t, "value1", mod1.Settings["key1"])
	assert.Equal(t, "value2", mod1.Settings["key2"])

	// Check new mod2
	mod2 := result[1]
	assert.Equal(t, "mod2", mod2.Name)
	assert.Equal(t, "./mod2", mod2.Path)
}
