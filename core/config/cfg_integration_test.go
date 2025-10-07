package config

import (
	"os"
	"testing"
)

// TestLoadConfigWithCFGResolver tests full config loading with CFG resolver
func TestLoadConfigWithCFGResolver(t *testing.T) {
	// Create a temporary config file with CFG resolver usage
	configYAML := `
configs:
  - name: database.host
    value: postgres.example.com
  - name: database.port
    value: 5432
  - name: database.name
    value: myapp
  - name: redis.host
    value: redis.example.com
  - name: redis.port
    value: 6379

servers:
  - name: "${@CFG:database.host}"
    baseUrl: "http://${@CFG:database.host}:${@CFG:database.port}"
    apps:
      - name: api
        addr: /api

services:
  - name: cache
    type: redis
    config:
      host: "${@CFG:redis.host}"
      port: ${@CFG:redis.port}
`

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configYAML); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	filePath := tmpFile.Name()
	tmpFile.Close()

	// Load config
	config := &Config{}
	err = LoadConfigFile(filePath, config)
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	// Verify server name expanded correctly
	if len(config.Servers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(config.Servers))
	}
	expectedName := "postgres.example.com"
	if config.Servers[0].Name != expectedName {
		t.Errorf("Server name: expected %q, got %q", expectedName, config.Servers[0].Name)
	}

	// Verify baseUrl expanded correctly
	expectedBaseUrl := "http://postgres.example.com:5432"
	if config.Servers[0].BaseUrl != expectedBaseUrl {
		t.Errorf("Server baseUrl: expected %q, got %q", expectedBaseUrl, config.Servers[0].BaseUrl)
	}

	// Verify service config expanded correctly
	if len(config.Services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(config.Services))
	}
	serviceConfig := config.Services[0].Config
	if serviceConfig == nil {
		t.Fatalf("Service config is nil")
	}

	host, ok := serviceConfig["host"].(string)
	if !ok {
		t.Fatalf("host is not a string")
	}
	if host != "redis.example.com" {
		t.Errorf("service host: expected %q, got %q", "redis.example.com", host)
	}

	// Port should be expanded as integer
	port, ok := serviceConfig["port"].(int)
	if !ok {
		// Might be float64 from YAML parsing
		portFloat, ok := serviceConfig["port"].(float64)
		if !ok {
			t.Fatalf("port is not int or float64: %T", serviceConfig["port"])
		}
		port = int(portFloat)
	}
	if port != 6379 {
		t.Errorf("service port: expected 6379, got %d", port)
	}
}

// TestLoadConfigWithCFGAndENV tests CFG and ENV resolvers working together
func TestLoadConfigWithCFGAndENV(t *testing.T) {
	// Set environment variable
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	configYAML := `
configs:
  - name: features.debug
    value: false
  - name: features.timeout
    value: 30

servers:
  - name: "${APP_ENV}-server"
    baseUrl: "http://localhost:8080"
    apps:
      - name: api
        addr: /

middlewares:
  - name: custom
    type: custom
    config:
      environment: "${APP_ENV}"
      debug: ${@CFG:features.debug}
      timeout: ${@CFG:features.timeout}
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configYAML); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	filePath := tmpFile.Name()
	tmpFile.Close()

	config := &Config{}
	err = LoadConfigFile(filePath, config)
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	// Verify ENV variable expanded
	if config.Servers[0].Name != "production-server" {
		t.Errorf("Server name: expected %q, got %q", "production-server", config.Servers[0].Name)
	}

	// Verify CFG variables expanded in middleware
	if len(config.Middlewares) != 1 {
		t.Fatalf("Expected 1 middleware, got %d", len(config.Middlewares))
	}
	mwConfig := config.Middlewares[0].Config

	env, ok := mwConfig["environment"].(string)
	if !ok {
		t.Fatalf("environment is not a string")
	}
	if env != "production" {
		t.Errorf("environment: expected %q, got %q", "production", env)
	}

	debug, ok := mwConfig["debug"].(bool)
	if !ok {
		t.Fatalf("debug is not a bool")
	}
	if debug != false {
		t.Errorf("debug: expected false, got %v", debug)
	}

	timeout, ok := mwConfig["timeout"].(int)
	if !ok {
		// Might be float64
		timeoutFloat, ok := mwConfig["timeout"].(float64)
		if !ok {
			t.Fatalf("timeout is not int or float64: %T", mwConfig["timeout"])
		}
		timeout = int(timeoutFloat)
	}
	if timeout != 30 {
		t.Errorf("timeout: expected 30, got %d", timeout)
	}
}

// TestLoadConfigWithNestedCFG tests CFG resolver with nested config paths
func TestLoadConfigWithNestedCFG(t *testing.T) {
	configYAML := `
configs:
  - name: database.primary.host
    value: db1.example.com
  - name: database.primary.port
    value: 5432
  - name: database.replica.host
    value: db2.example.com
  - name: database.replica.port
    value: 5433
  - name: app.name
    value: MyApp
  - name: app.version
    value: 1.0.0

servers:
  - name: "${@CFG:app.name}"
    baseUrl: "http://localhost:8080"
    apps:
      - name: api
        addr: /

services:
  - name: db-info
    type: custom
    config:
      primary: "${@CFG:database.primary.host}:${@CFG:database.primary.port}"
      replica: "${@CFG:database.replica.host}:${@CFG:database.replica.port}"
      version: "${@CFG:app.version}"
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configYAML); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	filePath := tmpFile.Name()
	tmpFile.Close()

	config := &Config{}
	err = LoadConfigFile(filePath, config)
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	// Verify server name
	if config.Servers[0].Name != "MyApp" {
		t.Errorf("Server name: expected %q, got %q", "MyApp", config.Servers[0].Name)
	}

	// Verify nested CFG paths in service config
	if len(config.Services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(config.Services))
	}
	serviceConfig := config.Services[0].Config

	primary, ok := serviceConfig["primary"].(string)
	if !ok {
		t.Fatalf("primary is not a string")
	}
	if primary != "db1.example.com:5432" {
		t.Errorf("primary: expected %q, got %q", "db1.example.com:5432", primary)
	}

	replica, ok := serviceConfig["replica"].(string)
	if !ok {
		t.Fatalf("replica is not a string")
	}
	if replica != "db2.example.com:5433" {
		t.Errorf("replica: expected %q, got %q", "db2.example.com:5433", replica)
	}

	version, ok := serviceConfig["version"].(string)
	if !ok {
		t.Fatalf("version is not a string")
	}
	if version != "1.0.0" {
		t.Errorf("version: expected %q, got %q", "1.0.0", version)
	}
}

// TestLoadConfigWithCFGDefaults tests CFG resolver with default values
func TestLoadConfigWithCFGDefaults(t *testing.T) {
	configYAML := `
configs:
  - name: database.host
    value: postgres.example.com
  # port is missing - will use default

servers:
  - name: api-server
    baseUrl: "http://${@CFG:database.host}:${@CFG:database.port:5432}"
    apps:
      - name: api
        addr: /

services:
  - name: db-service
    type: database
    config:
      db_host: "${@CFG:database.host}"
      db_port: ${@CFG:database.port:5432}
      db_name: "${@CFG:database.name:defaultdb}"
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configYAML); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	filePath := tmpFile.Name()
	tmpFile.Close()

	config := &Config{}
	err = LoadConfigFile(filePath, config)
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	// Verify server baseUrl uses default for missing port
	expectedBaseUrl := "http://postgres.example.com:5432"
	if config.Servers[0].BaseUrl != expectedBaseUrl {
		t.Errorf("Server baseUrl: expected %q, got %q", expectedBaseUrl, config.Servers[0].BaseUrl)
	}

	// Verify service config
	if len(config.Services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(config.Services))
	}
	serviceConfig := config.Services[0].Config

	dbHost, ok := serviceConfig["db_host"].(string)
	if !ok {
		t.Fatalf("db_host is not a string")
	}
	if dbHost != "postgres.example.com" {
		t.Errorf("db_host: expected %q, got %q", "postgres.example.com", dbHost)
	}

	// Port should use default
	dbPort, ok := serviceConfig["db_port"].(int)
	if !ok {
		portFloat, ok := serviceConfig["db_port"].(float64)
		if !ok {
			t.Fatalf("db_port is not int or float64: %T", serviceConfig["db_port"])
		}
		dbPort = int(portFloat)
	}
	if dbPort != 5432 {
		t.Errorf("db_port: expected 5432, got %d", dbPort)
	}

	// Name should use default
	dbName, ok := serviceConfig["db_name"].(string)
	if !ok {
		t.Fatalf("db_name is not a string")
	}
	if dbName != "defaultdb" {
		t.Errorf("db_name: expected %q, got %q", "defaultdb", dbName)
	}
}

// TestLoadConfigWithoutCFGSection tests that configs without CFG resolver work normally
func TestLoadConfigWithoutCFGSection(t *testing.T) {
	os.Setenv("TEST_PORT", "9090")
	defer os.Unsetenv("TEST_PORT")

	configYAML := `
servers:
  - name: test-server
    baseUrl: "http://localhost:${TEST_PORT:8080}"
    apps:
      - name: api
        addr: /
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configYAML); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	filePath := tmpFile.Name()
	tmpFile.Close()

	config := &Config{}
	err = LoadConfigFile(filePath, config)
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	// Verify ENV variable expanded
	if config.Servers[0].BaseUrl != "http://localhost:9090" {
		t.Errorf("Server baseUrl: expected %q, got %q", "http://localhost:9090", config.Servers[0].BaseUrl)
	}
}

// TestLoadConfigWithInvalidCFGKey tests CFG keys with default values when key doesn't exist
func TestLoadConfigWithInvalidCFGKey(t *testing.T) {
	configYAML := `
configs:
  - name: database.host
    value: postgres.example.com

servers:
  - name: "${@CFG:database.nonexistent:default-server}"
    baseUrl: "http://localhost:8080"
    apps:
      - name: api
        addr: /
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configYAML); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	filePath := tmpFile.Name()
	tmpFile.Close()

	config := &Config{}
	err = LoadConfigFile(filePath, config)
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	// Invalid CFG key should use default value
	if config.Servers[0].Name != "default-server" {
		t.Errorf("Server name: expected %q for invalid CFG key with default, got %q", "default-server", config.Servers[0].Name)
	}
}

// TestLoadConfigCFGWithComplexValues tests CFG with various data types
func TestLoadConfigCFGWithComplexValues(t *testing.T) {
	configYAML := `
configs:
  - name: cors.allowedOrigin
    value: "https://app.example.com"
  - name: cors.allowCredentials
    value: true
  - name: cors.maxAge
    value: 3600

middlewares:
  - name: cors
    type: cors
    config:
      allowedOrigin: "${@CFG:cors.allowedOrigin}"
      allowCredentials: ${@CFG:cors.allowCredentials}
      maxAge: ${@CFG:cors.maxAge}
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configYAML); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	filePath := tmpFile.Name()
	tmpFile.Close()

	config := &Config{}
	err = LoadConfigFile(filePath, config)
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	if len(config.Middlewares) == 0 {
		t.Fatalf("Expected middleware config")
	}

	mwConfig := config.Middlewares[0].Config
	if mwConfig == nil {
		t.Fatalf("Middleware config is nil")
	}

	// Verify string value
	allowedOrigin, ok := mwConfig["allowedOrigin"].(string)
	if !ok {
		t.Fatalf("allowedOrigin is not a string: %T", mwConfig["allowedOrigin"])
	}
	if allowedOrigin != "https://app.example.com" {
		t.Errorf("allowedOrigin: expected %q, got %q", "https://app.example.com", allowedOrigin)
	}

	// Verify boolean value
	allowCredentials, ok := mwConfig["allowCredentials"].(bool)
	if !ok {
		t.Fatalf("allowCredentials is not a bool: %T", mwConfig["allowCredentials"])
	}
	if allowCredentials != true {
		t.Errorf("allowCredentials: expected true, got %v", allowCredentials)
	}

	// Verify integer value
	maxAge, ok := mwConfig["maxAge"].(int)
	if !ok {
		maxAgeFloat, ok := mwConfig["maxAge"].(float64)
		if !ok {
			t.Fatalf("maxAge is not int or float64: %T", mwConfig["maxAge"])
		}
		maxAge = int(maxAgeFloat)
	}
	if maxAge != 3600 {
		t.Errorf("maxAge: expected 3600, got %d", maxAge)
	}
}
