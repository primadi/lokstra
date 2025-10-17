package deploy

import (
	"os"
	"testing"

	"github.com/primadi/lokstra/core/deploy/schema"
)

func TestGlobalRegistry_ConfigResolution(t *testing.T) {
	// Create fresh registry
	reg := NewGlobalRegistry()

	// Set environment variable
	os.Setenv("TEST_DB_URL", "postgres://test-server/testdb")
	defer os.Unsetenv("TEST_DB_URL")

	// Define configs
	reg.DefineConfig(&schema.ConfigDef{
		Name:  "DB_MAX_CONNS",
		Value: 20,
	})

	reg.DefineConfig(&schema.ConfigDef{
		Name:  "DB_URL",
		Value: "${TEST_DB_URL}",
	})

	// Resolve configs
	if err := reg.ResolveConfigs(); err != nil {
		t.Fatalf("failed to resolve configs: %v", err)
	}

	// Check DB_MAX_CONNS (should preserve type)
	maxConns, ok := reg.GetResolvedConfig("DB_MAX_CONNS")
	if !ok {
		t.Fatal("DB_MAX_CONNS not found")
	}
	if maxConns != 20 {
		t.Errorf("expected 20, got %v (type %T)", maxConns, maxConns)
	}

	// Check DB_URL (should resolve env var)
	dbURL, ok := reg.GetResolvedConfig("DB_URL")
	if !ok {
		t.Fatal("DB_URL not found")
	}
	if dbURL != "postgres://test-server/testdb" {
		t.Errorf("expected postgres URL, got %v", dbURL)
	}
}

func TestGlobalRegistry_ConfigReference(t *testing.T) {
	reg := NewGlobalRegistry()

	// Define configs
	reg.DefineConfig(&schema.ConfigDef{
		Name:  "MAX_CONNECTIONS",
		Value: 50,
	})

	reg.DefineConfig(&schema.ConfigDef{
		Name:  "LOG_LEVEL",
		Value: "debug",
	})

	// Resolve all configs first
	if err := reg.ResolveConfigs(); err != nil {
		t.Fatalf("failed to resolve configs: %v", err)
	}

	// Now resolve a service config that references @cfg
	serviceConfig := map[string]any{
		"max-conns": "${@cfg:MAX_CONNECTIONS}",
		"level":     "${@cfg:LOG_LEVEL}",
		"static":    "static-value",
	}

	// Resolve each config value
	resolvedMaxConns, err := reg.ResolveConfigValue(serviceConfig["max-conns"])
	if err != nil {
		t.Fatalf("failed to resolve max-conns: %v", err)
	}
	if resolvedMaxConns != 50 {
		t.Errorf("expected 50, got %v (type %T)", resolvedMaxConns, resolvedMaxConns)
	}

	resolvedLevel, err := reg.ResolveConfigValue(serviceConfig["level"])
	if err != nil {
		t.Fatalf("failed to resolve level: %v", err)
	}
	if resolvedLevel != "debug" {
		t.Errorf("expected 'debug', got %v", resolvedLevel)
	}

	resolvedStatic, err := reg.ResolveConfigValue(serviceConfig["static"])
	if err != nil {
		t.Fatalf("failed to resolve static: %v", err)
	}
	if resolvedStatic != "static-value" {
		t.Errorf("expected 'static-value', got %v", resolvedStatic)
	}
}

func TestGlobalRegistry_ServiceDefinition(t *testing.T) {
	reg := NewGlobalRegistry()

	// Define a service
	reg.DefineService(&schema.ServiceDef{
		Name:      "user-service",
		Type:      "user-factory",
		DependsOn: []string{"db-user", "logger"},
		Config: map[string]any{
			"cache-ttl": 300,
		},
	})

	// Retrieve service definition
	svc := reg.GetService("user-service")
	if svc == nil {
		t.Fatal("user-service not found")
	}

	if svc.Name != "user-service" {
		t.Errorf("expected name 'user-service', got '%s'", svc.Name)
	}

	if svc.Type != "user-factory" {
		t.Errorf("expected type 'user-factory', got '%s'", svc.Type)
	}

	if len(svc.DependsOn) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(svc.DependsOn))
	}
}

func TestGlobalRegistry_RouterOverride(t *testing.T) {
	reg := NewGlobalRegistry()

	// Define router override
	enabledFalse := false
	reg.DefineRouterOverride(&schema.RouterOverrideDef{
		Name:        "user-public-api",
		PathPrefix:  "/api/v1",
		Middlewares: []string{"cors", "rate-limit"},
		Hidden:      []string{"Delete", "BulkDelete"},
		Routes: []schema.RouteDef{
			{
				Name:        "Create",
				Path:        "/register",
				Middlewares: []string{"recaptcha"},
			},
			{
				Name:    "AdminReset",
				Enabled: &enabledFalse,
			},
		},
	})

	// Retrieve override
	override := reg.GetRouterOverride("user-public-api")
	if override == nil {
		t.Fatal("user-public-api override not found")
	}

	if override.PathPrefix != "/api/v1" {
		t.Errorf("expected path prefix '/api/v1', got '%s'", override.PathPrefix)
	}

	if len(override.Hidden) != 2 {
		t.Errorf("expected 2 hidden methods, got %d", len(override.Hidden))
	}

	if len(override.Routes) != 2 {
		t.Errorf("expected 2 route overrides, got %d", len(override.Routes))
	}
}

func TestGlobalRegistry_FactoryRegistration(t *testing.T) {
	reg := NewGlobalRegistry()

	// Mock factories
	localFactory := func(deps map[string]any, config map[string]any) any {
		return "local-instance"
	}
	remoteFactory := func(deps map[string]any, config map[string]any) any {
		return "remote-instance"
	}

	// Register service type
	reg.RegisterServiceType("test-service", localFactory, remoteFactory)

	// Get local factory
	local := reg.GetServiceFactory("test-service", true)
	if local == nil {
		t.Fatal("local factory not found")
	}

	result := local(nil, nil)
	if result != "local-instance" {
		t.Errorf("expected 'local-instance', got %v", result)
	}

	// Get remote factory
	remote := reg.GetServiceFactory("test-service", false)
	if remote == nil {
		t.Fatal("remote factory not found")
	}

	result = remote(nil, nil)
	if result != "remote-instance" {
		t.Errorf("expected 'remote-instance', got %v", result)
	}
}

func TestGlobalRegistry_MiddlewareFactory(t *testing.T) {
	reg := NewGlobalRegistry()

	// Mock middleware factory
	mwFactory := func(config map[string]any) any {
		return "middleware-instance"
	}

	// Register middleware type
	reg.RegisterMiddlewareType("test-middleware", mwFactory)

	// Get middleware factory
	factory := reg.GetMiddlewareFactory("test-middleware")
	if factory == nil {
		t.Fatal("middleware factory not found")
	}

	result := factory(nil)
	if result != "middleware-instance" {
		t.Errorf("expected 'middleware-instance', got %v", result)
	}
}

func TestGlobalSingleton(t *testing.T) {
	// Get global registry
	reg1 := Global()
	reg2 := Global()

	// Should be the same instance
	if reg1 != reg2 {
		t.Error("Global() should return singleton instance")
	}

	// Test registration via global
	reg1.RegisterServiceType("test", nil, nil)

	// Should be accessible from reg2
	factory := reg2.GetServiceFactory("test", true)
	if factory != nil {
		t.Error("expected nil factory (we registered nil)")
	}
	// But the type should be registered
	if reg2.GetServiceFactory("test", false) != nil {
		t.Error("expected nil remote factory")
	}
}
