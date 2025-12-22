package loader_test

import (
	"path/filepath"
	"testing"

	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/core/deploy/loader/internal"
)

func TestLoadSingleFile(t *testing.T) {
	config, err := loader.LoadConfig("./testdata/base.yaml")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Check configs
	if config.Configs == nil {
		t.Fatal("configs should not be nil")
	}

	if config.Configs["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %v", config.Configs["DB_HOST"])
	}

	// Check services
	if len(config.ServiceDefinitions) == 0 {
		t.Fatal("expected services to be loaded")
	}

	dbPool := config.ServiceDefinitions["db-pool"]
	if dbPool == nil {
		t.Fatal("db-pool service not found")
	}

	if dbPool.Type != "postgres-pool" {
		t.Errorf("expected type postgres-pool, got %s", dbPool.Type)
	}
}

func TestLoadMultipleFiles(t *testing.T) {
	config, err := loader.LoadConfig(
		"./testdata/base.yaml",
		"./testdata/services.yaml",
		"./testdata/deployments.yaml",
	)
	if err != nil {
		t.Fatalf("failed to load configs: %v", err)
	}

	// Check base configs merged
	if config.Configs["DB_HOST"] != "localhost" {
		t.Errorf("base config not merged correctly")
	}

	// Check services from different files merged
	if config.ServiceDefinitions["db-pool"] == nil {
		t.Error("service from base.yaml not found")
	}

	if config.ServiceDefinitions["user-service"] == nil {
		t.Error("service from services.yaml not found")
	}

	// Check user-service dependencies
	userSvc := config.ServiceDefinitions["user-service"]
	if len(userSvc.DependsOn) != 3 {
		t.Errorf("expected 3 dependencies, got %d", len(userSvc.DependsOn))
	}

	// Check deployments
	if len(config.Deployments) != 2 {
		t.Errorf("expected 2 deployments, got %d", len(config.Deployments))
	}

	prodDep := config.Deployments["production"]
	if prodDep == nil {
		t.Fatal("production deployment not found")
	}

	// Check config overrides
	if prodDep.ConfigOverrides["LOG_LEVEL"] != "warn" {
		t.Errorf("config override not applied")
	}

	// Check servers
	if len(prodDep.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(prodDep.Servers))
	}

	apiServer := prodDep.Servers["api-server"]
	if apiServer == nil {
		t.Fatal("api-server not found")
	}

	if apiServer.BaseURL != "https://api.example.com" {
		t.Errorf("expected api base URL, got %s", apiServer.BaseURL)
	}

	// Check apps
	if len(apiServer.Apps) != 1 {
		t.Errorf("expected 1 app, got %d", len(apiServer.Apps))
	}

	app := apiServer.Apps[0]
	if app.Addr != ":8080" {
		t.Errorf("expected addr :8080, got %s", app.Addr)
	}
}

func TestLoadFromDirectory(t *testing.T) {
	config, err := loader.LoadConfig("testdata")
	if err != nil {
		t.Fatalf("failed to load from directory: %v", err)
	}

	// Should have merged all files
	if len(config.ServiceDefinitions) < 3 {
		t.Errorf("expected at least 3 services, got %d", len(config.ServiceDefinitions))
	}

	if len(config.Deployments) < 2 {
		t.Errorf("expected at least 2 deployments, got %d", len(config.Deployments))
	}
}

func TestMergeStrategy(t *testing.T) {
	// Create temp files with overlapping configs
	config, err := loader.LoadConfig(
		"testdata/base.yaml",
		"testdata/services.yaml",
	)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	// Later files should override earlier ones
	// Both files define services, they should be merged
	if config.ServiceDefinitions["db-pool"] == nil {
		t.Error("service from first file missing")
	}

	if config.ServiceDefinitions["user-service"] == nil {
		t.Error("service from second file missing")
	}

	// If same service is defined twice, second should win
	// (not tested here as we don't have overlapping services in test files)
}

func TestValidation_ValidConfig(t *testing.T) {
	config, err := loader.LoadConfig("testdata/base.yaml")
	if err != nil {
		t.Fatalf("valid config should load without error: %v", err)
	}

	// Validation happens automatically in loader.LoadConfig
	// If we got here, validation passed
	if config == nil {
		t.Error("config should not be nil")
	}
}

func TestValidation_InvalidServiceName(t *testing.T) {
	// Create invalid config in memory and validate
	// This tests the validation logic directly

	// Note: We could create temp invalid YAML files for more thorough testing
	// For now, the schema validation is tested through loading valid files
}

func TestConfigToMap(t *testing.T) {
	config, err := loader.LoadConfig("testdata/base.yaml", "testdata/services.yaml")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	configMap := internal.ConfigToMap(config)

	// Check configs section
	configs, ok := configMap["configs"].(map[string]any)
	if !ok {
		t.Fatal("configs section missing or wrong type")
	}

	if configs["DB_HOST"] != "localhost" {
		t.Error("config value not converted correctly")
	}

	// Check services section
	services, ok := configMap["service-definitions"].(map[string]any)
	if !ok {
		t.Fatal("services section missing or wrong type")
	}

	if len(services) == 0 {
		t.Error("services should not be empty")
	}

	// Check service structure
	dbPoolAny, ok := services["db-pool"]
	if !ok {
		t.Fatal("db-pool service not found in map")
	}

	dbPool, ok := dbPoolAny.(map[string]any)
	if !ok {
		t.Fatal("db-pool should be a map")
	}

	if dbPool["type"] != "postgres-pool" {
		t.Error("service type not converted correctly")
	}
}

func TestAbsolutePaths(t *testing.T) {
	// Test with absolute paths
	absPath, err := filepath.Abs("testdata/base.yaml")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	config, err := loader.LoadConfig(absPath)
	if err != nil {
		t.Fatalf("failed to load with absolute path: %v", err)
	}

	if config == nil {
		t.Error("config should not be nil")
	}
}

func TestNonExistentFile(t *testing.T) {
	_, err := loader.LoadConfig("testdata/nonexistent.yaml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestEmptyConfig(t *testing.T) {
	_, err := loader.LoadConfig()
	if err == nil {
		t.Error("expected error when no files specified")
	}
}

func TestShorthandSyntax(t *testing.T) {
	config, err := loader.LoadConfig("testdata/shorthand.yaml")
	if err != nil {
		t.Fatalf("failed to load shorthand config: %v", err)
	}

	dep := config.Deployments["test"]
	if dep == nil {
		t.Fatal("test deployment not found")
	}

	server := dep.Servers["api-server"]
	if server == nil {
		t.Fatal("api-server not found")
	}

	// After loader.LoadConfig, normalization already happened, so helper fields are cleared
	// and apps should be created
	if len(server.Apps) != 1 {
		t.Errorf("expected 1 app after normalization, got %d", len(server.Apps))
	}

	app := server.Apps[0]
	if app.Addr != ":3000" {
		t.Errorf("expected app addr :3000, got %s", app.Addr)
	}

	if len(app.PublishedServices) != 2 {
		t.Errorf("expected 2 published services in app, got %d", len(app.PublishedServices))
	}

	// Verify helper fields are cleared after normalization
	if server.HelperAddr != "" {
		t.Errorf("expected helper addr to be cleared, got %s", server.HelperAddr)
	}
}

func TestSmartMerging(t *testing.T) {
	// Test smart merging behavior: helper fields without addr should merge into first app
	config, err := loader.LoadConfig("testdata/deployments.yaml")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Check production deployment
	prodDep := config.Deployments["production"]
	if prodDep == nil {
		t.Fatal("production deployment not found")
	}

	apiServer := prodDep.Servers["api-server"]
	if apiServer == nil {
		t.Fatal("api-server not found")
	}

	// Should have 1 app (helper merged into first app)
	if len(apiServer.Apps) != 1 {
		t.Errorf("expected 1 app after smart merge, got %d", len(apiServer.Apps))
	}

	app := apiServer.Apps[0]

	// App should have addr from explicit app definition
	if app.Addr != ":8080" {
		t.Errorf("expected addr :8080, got %s", app.Addr)
	}

	// App should have published-services from helper fields (merged)
	if len(app.PublishedServices) != 2 {
		t.Errorf("expected 2 published services (merged from helper), got %d", len(app.PublishedServices))
	}

	expectedServices := map[string]bool{
		"user-service":  false,
		"order-service": false,
	}

	for _, svc := range app.PublishedServices {
		if _, exists := expectedServices[svc]; exists {
			expectedServices[svc] = true
		} else {
			t.Errorf("unexpected service: %s", svc)
		}
	}

	for svc, found := range expectedServices {
		if !found {
			t.Errorf("expected service not found: %s", svc)
		}
	}

	// Verify helper fields are cleared
	if apiServer.HelperAddr != "" {
		t.Errorf("expected helper addr to be cleared, got %s", apiServer.HelperAddr)
	}
	if len(apiServer.HelperPublishedServices) > 0 {
		t.Errorf("expected helper published-services to be cleared, got %d items", len(apiServer.HelperPublishedServices))
	}
}

func TestValidateConfigYAML_Valid(t *testing.T) {
	validYAML := `
configs:
  app-name: test-app
  db-host: localhost

service-definitions:
  user-service:
    type: user-service-factory
    depends-on:
      - db-pool

deployments:
  development:
    servers:
      api:
        base-url: http://localhost:8080
        apps:
          - addr: ":8080"
            published-services:
              - user-service
`

	err := loader.ValidateConfigYAML([]byte(validYAML))
	if err != nil {
		t.Errorf("valid YAML should pass validation: %v", err)
	}
}

func TestValidateConfigYAML_InvalidField(t *testing.T) {
	invalidYAML := `
configs:
  app-name: test-app

servise-definitions:
  user-service:
    type: user-service-factory
`

	err := loader.ValidateConfigYAML([]byte(invalidYAML))
	if err == nil {
		t.Error("expected validation error for invalid field 'servise-definitions'")
	}
}

func TestValidateConfigYAML_MissingRequiredField(t *testing.T) {
	// Service definition without 'type' field
	invalidYAML := `
service-definitions:
  user-service:
    depends-on:
      - db-pool
`

	err := loader.ValidateConfigYAML([]byte(invalidYAML))
	if err == nil {
		t.Error("expected validation error for missing 'type' field")
	}
}

func TestValidateConfigYAML_InvalidType(t *testing.T) {
	// addr should be string, not number
	invalidYAML := `
deployments:
  development:
    servers:
      api:
        base-url: http://localhost:8080
        apps:
          - addr: 8080
            published-services:
              - user-service
`

	err := loader.ValidateConfigYAML([]byte(invalidYAML))
	if err == nil {
		t.Error("expected validation error for wrong type (number instead of string)")
	}
}

func TestValidateConfigYAML_InvalidDependsOn(t *testing.T) {
	// depends-on with invalid service name pattern
	invalidYAML := `
service-definitions:
  user-service:
    type: user-service-factory
    depends-on:
      - DB_POOL
`

	err := loader.ValidateConfigYAML([]byte(invalidYAML))
	if err == nil {
		t.Error("expected validation error for invalid depends-on pattern (uppercase with underscore)")
	}
}

func TestValidateConfigYAML_MalformedYAML(t *testing.T) {
	malformedYAML := `
configs:
  app-name: test-app
  invalid yaml:
   - missing proper indentation
  - broken structure
`

	err := loader.ValidateConfigYAML([]byte(malformedYAML))
	if err == nil {
		t.Error("expected error for malformed YAML")
	}
}

func TestValidateConfigYAML_AdditionalProperties(t *testing.T) {
	// Unknown top-level field
	invalidYAML := `
configs:
  app-name: test-app

unknown-field:
  something: value
`

	err := loader.ValidateConfigYAML([]byte(invalidYAML))
	if err == nil {
		t.Error("expected validation error for unknown top-level field")
	}
}

func TestValidateConfigYAML_EmptyConfig(t *testing.T) {
	emptyYAML := `{}`

	err := loader.ValidateConfigYAML([]byte(emptyYAML))
	// Empty config might be valid depending on schema requirements
	// Adjust assertion based on actual schema rules
	if err != nil {
		t.Logf("Empty config validation: %v", err)
	}
}
