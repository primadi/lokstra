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

	// Create a service definition with config references
	// This simulates how service configs with @cfg references are handled
	reg.RegisterLazyService("test-service", "test-factory", map[string]any{
		"max-conns": "${@cfg:MAX_CONNECTIONS}",
		"level":     "${@cfg:LOG_LEVEL}",
		"static":    "static-value",
	})

	// Resolve all configs - this should resolve service configs too
	if err := reg.ResolveConfigs(); err != nil {
		t.Fatalf("failed to resolve configs: %v", err)
	}

	// Get the resolved service definition
	serviceDef := reg.GetDeferredServiceDef("test-service")
	if serviceDef == nil {
		t.Fatal("service definition not found")
	}

	// Check resolved values
	if serviceDef.Config["max-conns"] != 50 {
		t.Errorf("expected 50, got %v (type %T)", serviceDef.Config["max-conns"], serviceDef.Config["max-conns"])
	}

	if serviceDef.Config["level"] != "debug" {
		t.Errorf("expected 'debug', got %v", serviceDef.Config["level"])
	}

	if serviceDef.Config["static"] != "static-value" {
		t.Errorf("expected 'static-value', got %v", serviceDef.Config["static"])
	}
}

func TestGlobalRegistry_ServiceDefinition(t *testing.T) {
	reg := NewGlobalRegistry()

	// Register a service with string factory type (new unified API)
	reg.RegisterLazyService("user-service", "user-factory", map[string]any{
		"depends-on": []string{"db-user", "logger"},
		"cache-ttl":  300,
	})

	// Check if service is registered
	if !reg.HasLazyService("user-service") {
		t.Fatal("user-service not found")
	}

	// Retrieve service definition
	svc := reg.GetDeferredServiceDef("user-service")
	if svc == nil {
		t.Fatal("user-service definition not found")
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

// TestGlobalRegistry_RouterOverride - DEPRECATED
// Router overrides are now inline in RouterDef, this test is kept for reference
// but the functionality has been removed in favor of inline overrides
func TestGlobalRegistry_RouterOverride(t *testing.T) {
	t.Skip("Router overrides are now inline in RouterDef, DefineRouterOverride/GetRouterOverride have been removed")

	// This test is now obsolete because:
	// 1. RouterOverrideDef struct has been deleted
	// 2. DefineRouterOverride() method has been removed
	// 3. GetRouterOverride() method has been removed
	// 4. Overrides are now specified directly in RouterDef with inline fields:
	//    - PathPrefix, Middlewares, Hidden, Custom
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
