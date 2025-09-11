package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================
// INTEGRATION TESTS
// ==============================================

func TestLoadConfigDir_CompleteConfiguration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config_integration_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set environment variables for testing
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DB_HOST", "testdb.example.com")
	os.Setenv("REDIS_PORT", "6380")
	defer func() {
		os.Unsetenv("APP_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("REDIS_PORT")
	}()

	// Create external API routes file
	apiRoutes := `
routes:
  - method: GET
    path: /users
    handler: user.ListHandler
    middleware:
      - rate_limit
  - method: POST
    path: /users
    handler: user.CreateHandler

groups:
  - prefix: /admin
    middleware:
      - name: auth
        config:
          role: admin
    routes:
      - method: GET
        path: /stats
        handler: admin.StatsHandler

mount_static:
  - prefix: /uploads
    folder: [./uploads]

mount_rpc_service:
  - base_path: /rpc
    service_name: user_service
`

	err = os.WriteFile(filepath.Join(tempDir, "api-routes.yaml"), []byte(apiRoutes), 0644)
	require.NoError(t, err)

	// Main configuration files
	serverConfig := `
server:
  name: ${APP_NAME:integration-test-server}
  global_setting:
    debug: ${DEBUG:true}
    max_connections: ${MAX_CONN:1000}
`

	appsConfig := `
apps:
  - name: main-app
    address: ":${APP_PORT}"
    listener_type: ${LISTENER_TYPE:http}
    router_engine_type: httprouter
    middleware:
      - cors
      - name: logging
        enabled: true
        config:
          level: ${LOG_LEVEL:info}
    setting:
      read_timeout: ${READ_TIMEOUT:30}
      write_timeout: 30
    routes:
      - method: GET
        path: /health
        handler: health.CheckHandler
      - method: GET
        path: /version
        handler: version.Handler
        middleware:
          - cache
    groups:
      - prefix: /api/v1
        load_from:
          - api-routes.yaml
        middleware:
          - name: api_auth
            config:
              secret: ${API_SECRET:default-secret}
      - prefix: /api/v2
        routes:
          - method: GET
            path: /info
            handler: info.Handler
        mount_reverse_proxy:
          - prefix: /proxy
            target: http://backend.example.com

  - name: admin-app
    address: ":${ADMIN_PORT:8081}"
    routes:
      - method: GET
        path: /admin/health
        handler: admin.HealthHandler
`

	servicesConfig := `
services:
  - name: database
    type: postgres
    config:
      host: ${DB_HOST}
      port: ${DB_PORT:5432}
      database: ${DB_NAME:testdb}
      username: ${DB_USER:testuser}
      password: ${DB_PASS:testpass}
      
  - name: cache
    type: redis
    config:
      host: ${REDIS_HOST:localhost}
      port: ${REDIS_PORT}
      database: ${REDIS_DB:0}

  - name: user_service
    type: user_service_factory
    config:
      api_key: ${USER_API_KEY:default-key}
`

	modulesConfig := `
modules:
  - name: auth_module
    path: ./modules/auth
    entry: AuthMain
    settings:
      jwt_secret: ${JWT_SECRET:default-jwt-secret}
      token_ttl: ${TOKEN_TTL:3600}
    permissions:
      file_access: ${FILE_ACCESS:read}
      network_access: ${NETWORK_ACCESS:restricted}

  - name: metrics_module
    path: ./modules/metrics
    settings:
      enabled: ${METRICS_ENABLED:true}
      endpoint: /metrics
`

	// Write all configuration files
	err = os.WriteFile(filepath.Join(tempDir, "01-server.yaml"), []byte(serverConfig), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "02-apps.yaml"), []byte(appsConfig), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "03-services.yaml"), []byte(servicesConfig), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "04-modules.yaml"), []byte(modulesConfig), 0644)
	require.NoError(t, err)

	// Load and test the configuration
	cfg, err := LoadConfigDir(tempDir)
	require.NoError(t, err)

	// Test server configuration
	assert.Equal(t, "integration-test-server", cfg.Server.Name)
	assert.Equal(t, true, cfg.Server.Settings["debug"])           // Should be bool
	assert.Equal(t, 1000, cfg.Server.Settings["max_connections"]) // Should be int

	// Test apps configuration
	require.Len(t, cfg.Apps, 2)

	// Test main app
	mainApp := cfg.Apps[0]
	assert.Equal(t, "main-app", mainApp.Name)
	assert.Equal(t, ":9090", mainApp.Address) // From environment variable
	assert.Equal(t, "http", mainApp.ListenerType)
	assert.Equal(t, "httprouter", mainApp.RouterEngineType)

	// Test main app middleware
	require.Len(t, mainApp.Middleware, 2)
	assert.Equal(t, "cors", mainApp.Middleware[0].Name)
	assert.Equal(t, "logging", mainApp.Middleware[1].Name)
	assert.Equal(t, "info", mainApp.Middleware[1].Config["level"])

	// Test main app settings
	assert.Equal(t, 30, mainApp.Settings["read_timeout"]) // Should be int
	assert.Equal(t, 30, mainApp.Settings["write_timeout"])

	// Test main app routes
	require.Len(t, mainApp.Routes, 2)
	assert.Equal(t, "/health", mainApp.Routes[0].Path)
	assert.Equal(t, "/version", mainApp.Routes[1].Path)

	// Test route middleware
	versionRoute := mainApp.Routes[1]
	require.Len(t, versionRoute.Middleware, 1)
	assert.Equal(t, "cache", versionRoute.Middleware[0].Name)

	// Test main app groups
	require.Len(t, mainApp.Groups, 2)

	// Test API v1 group (with included routes)
	apiV1Group := mainApp.Groups[0]
	assert.Equal(t, "/api/v1", apiV1Group.Prefix)

	// Test group middleware
	require.Len(t, apiV1Group.Middleware, 1)
	assert.Equal(t, "api_auth", apiV1Group.Middleware[0].Name)
	assert.Equal(t, "default-secret", apiV1Group.Middleware[0].Config["secret"])

	// Test included routes
	require.Len(t, apiV1Group.Routes, 2)
	assert.Equal(t, "/users", apiV1Group.Routes[0].Path)
	assert.Equal(t, "GET", apiV1Group.Routes[0].Method)
	assert.Equal(t, "POST", apiV1Group.Routes[1].Method)

	// Test included route middleware
	usersGetRoute := apiV1Group.Routes[0]
	require.Len(t, usersGetRoute.Middleware, 1)
	assert.Equal(t, "rate_limit", usersGetRoute.Middleware[0].Name)

	// Test included nested groups
	require.Len(t, apiV1Group.Groups, 1)
	adminGroup := apiV1Group.Groups[0]
	assert.Equal(t, "/admin", adminGroup.Prefix)
	require.Len(t, adminGroup.Middleware, 1)
	assert.Equal(t, "auth", adminGroup.Middleware[0].Name)
	assert.Equal(t, "admin", adminGroup.Middleware[0].Config["role"])

	// Test mount configurations
	require.Len(t, apiV1Group.MountStatic, 1)
	assert.Equal(t, "/uploads", apiV1Group.MountStatic[0].Prefix)
	assert.Equal(t, "./uploads", apiV1Group.MountStatic[0].Folder[0])

	require.Len(t, apiV1Group.MountRpcService, 1)
	assert.Equal(t, "/rpc", apiV1Group.MountRpcService[0].BasePath)
	assert.Equal(t, "user_service", apiV1Group.MountRpcService[0].ServiceName)

	// Test API v2 group
	apiV2Group := mainApp.Groups[1]
	assert.Equal(t, "/api/v2", apiV2Group.Prefix)
	require.Len(t, apiV2Group.Routes, 1)
	assert.Equal(t, "/info", apiV2Group.Routes[0].Path)

	require.Len(t, apiV2Group.MountReverseProxy, 1)
	assert.Equal(t, "/proxy", apiV2Group.MountReverseProxy[0].Prefix)
	assert.Equal(t, "http://backend.example.com", apiV2Group.MountReverseProxy[0].Target)

	// Test admin app
	adminApp := cfg.Apps[1]
	assert.Equal(t, "admin-app", adminApp.Name)
	assert.Equal(t, ":8081", adminApp.Address)

	// Test services configuration
	require.Len(t, cfg.Services, 3)

	// Test database service
	dbService := cfg.Services[0]
	assert.Equal(t, "database", dbService.Name)
	assert.Equal(t, "postgres", dbService.Type)
	assert.Equal(t, "testdb.example.com", dbService.Config["host"])
	assert.Equal(t, 5432, dbService.Config["port"]) // Should be int
	assert.Equal(t, "testdb", dbService.Config["database"])

	// Test cache service
	cacheService := cfg.Services[1]
	assert.Equal(t, "cache", cacheService.Name)
	assert.Equal(t, "redis", cacheService.Type)
	assert.Equal(t, "localhost", cacheService.Config["host"])
	assert.Equal(t, 6380, cacheService.Config["port"]) // Should be int - from environment

	// Test user service
	userService := cfg.Services[2]
	assert.Equal(t, "user_service", userService.Name)
	assert.Equal(t, "user_service_factory", userService.Type)
	assert.Equal(t, "default-key", userService.Config["api_key"])

	// Test modules configuration
	require.Len(t, cfg.Modules, 2)

	// Test auth module
	authModule := cfg.Modules[0]
	assert.Equal(t, "auth_module", authModule.Name)
	assert.Equal(t, "./modules/auth", authModule.Path)
	assert.Equal(t, "AuthMain", authModule.Entry)
	assert.Equal(t, "default-jwt-secret", authModule.Settings["jwt_secret"])
	assert.Equal(t, 3600, authModule.Settings["token_ttl"]) // Should be int
	assert.Equal(t, "read", authModule.Permissions["file_access"])

	// Test metrics module
	metricsModule := cfg.Modules[1]
	assert.Equal(t, "metrics_module", metricsModule.Name)
	assert.Equal(t, "./modules/metrics", metricsModule.Path)
	assert.Equal(t, true, metricsModule.Settings["enabled"]) // Should be bool
	assert.Equal(t, "/metrics", metricsModule.Settings["endpoint"])
}

func TestLoadConfigDir_MergeConflicts(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config_merge_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// First config file
	config1 := `
apps:
  - name: shared-app
    address: ":8080"
    routes:
      - method: GET
        path: /route1
        handler: handler1
    middleware:
      - cors
    setting:
      key1: value1
      shared: from_config1

services:
  - name: shared-service
    type: postgres
    config:
      host: localhost
      key1: value1

modules:
  - name: shared-module
    path: ./module1
    settings:
      key1: value1
      shared: from_config1
`

	// Second config file with conflicts
	config2 := `
apps:
  - name: shared-app  # Same app name - should merge
    routes:
      - method: POST
        path: /route2
        handler: handler2
    middleware:
      - auth
    setting:
      key2: value2
      shared: from_config2  # Should override

services:
  - name: shared-service  # Same service name - should merge
    config:
      port: 5432
      key2: value2

modules:
  - name: shared-module  # Same module name - should merge
    settings:
      key2: value2
      shared: from_config2  # Should override
`

	err = os.WriteFile(filepath.Join(tempDir, "01-first.yaml"), []byte(config1), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "02-second.yaml"), []byte(config2), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfigDir(tempDir)
	require.NoError(t, err)

	// Test app merging
	require.Len(t, cfg.Apps, 1)
	app := cfg.Apps[0]
	assert.Equal(t, "shared-app", app.Name)

	// Should have routes from both files
	require.Len(t, app.Routes, 2)
	assert.Equal(t, "/route1", app.Routes[0].Path)
	assert.Equal(t, "/route2", app.Routes[1].Path)

	// Should have middleware from both files
	require.Len(t, app.Middleware, 2)
	assert.Equal(t, "cors", app.Middleware[0].Name)
	assert.Equal(t, "auth", app.Middleware[1].Name)

	// Settings should be merged with later file overriding
	assert.Equal(t, "value1", app.Settings["key1"])
	assert.Equal(t, "value2", app.Settings["key2"])
	assert.Equal(t, "from_config2", app.Settings["shared"])

	// Test service merging
	require.Len(t, cfg.Services, 1)
	service := cfg.Services[0]
	assert.Equal(t, "shared-service", service.Name)
	assert.Equal(t, "localhost", service.Config["host"])
	assert.Equal(t, 5432, service.Config["port"])
	assert.Equal(t, "value1", service.Config["key1"])
	assert.Equal(t, "value2", service.Config["key2"])

	// Test module merging
	require.Len(t, cfg.Modules, 1)
	module := cfg.Modules[0]
	assert.Equal(t, "shared-module", module.Name)
	assert.Equal(t, "./module1", module.Path) // From first file
	assert.Equal(t, "value1", module.Settings["key1"])
	assert.Equal(t, "value2", module.Settings["key2"])
	assert.Equal(t, "from_config2", module.Settings["shared"])
}

func TestLoadConfigDir_ComplexMiddlewareNormalization(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config_middleware_complex_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	groupFile := `
groups:
  - prefix: /internal
    middleware:
      - name: internal_auth
        config:
          type: jwt
      - rate_limit
    routes:
      - method: GET
        path: /metrics
        handler: metrics.Handler
        middleware:
          - name: metrics_auth
            enabled: false
`

	err = os.WriteFile(filepath.Join(tempDir, "groups.yaml"), []byte(groupFile), 0644)
	require.NoError(t, err)

	config := `
apps:
  - name: complex-app
    middleware:
      - cors
      - name: global_auth
        enabled: true
        config:
          providers:
            - oauth2
            - basic
          timeout: 30
    routes:
      - method: GET
        path: /public
        handler: public.Handler
        override_middleware: true
        middleware:
          - public_only
    groups:
      - prefix: /api
        load_from:
          - groups.yaml
        override_middleware: false
        middleware:
          - api_auth
`

	err = os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(config), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfigDir(tempDir)
	require.NoError(t, err)

	app := cfg.Apps[0]

	// Test app-level middleware
	require.Len(t, app.Middleware, 2)
	assert.Equal(t, "cors", app.Middleware[0].Name)
	assert.Equal(t, "global_auth", app.Middleware[1].Name)
	assert.True(t, app.Middleware[1].Enabled)

	providers := app.Middleware[1].Config["providers"].([]any)
	require.Len(t, providers, 2)
	assert.Equal(t, "oauth2", providers[0])
	assert.Equal(t, "basic", providers[1])

	// Test route-level middleware
	publicRoute := app.Routes[0]
	assert.True(t, publicRoute.OverrideMiddleware)
	require.Len(t, publicRoute.Middleware, 1)
	assert.Equal(t, "public_only", publicRoute.Middleware[0].Name)

	// Test group middleware hierarchy
	apiGroup := app.Groups[0]
	assert.False(t, apiGroup.OverrideMiddleware)

	// Should have middleware from load_from file + local middleware
	require.Len(t, apiGroup.Middleware, 1)
	assert.Equal(t, "api_auth", apiGroup.Middleware[0].Name) // Local middleware

	// Test nested group from included file
	require.Len(t, apiGroup.Groups, 1)
	internalGroup := apiGroup.Groups[0]
	assert.Equal(t, "/internal", internalGroup.Prefix)
	require.Len(t, internalGroup.Middleware, 2)
	assert.Equal(t, "internal_auth", internalGroup.Middleware[0].Name)
	assert.Equal(t, "jwt", internalGroup.Middleware[0].Config["type"])
	assert.Equal(t, "rate_limit", internalGroup.Middleware[1].Name) // Simple string middleware

	// Test route in nested group
	require.Len(t, internalGroup.Routes, 1)
	metricsRoute := internalGroup.Routes[0]
	require.Len(t, metricsRoute.Middleware, 1)
	assert.Equal(t, "metrics_auth", metricsRoute.Middleware[0].Name)
	assert.False(t, metricsRoute.Middleware[0].Enabled)
}
