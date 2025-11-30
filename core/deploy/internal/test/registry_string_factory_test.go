package deploy_test

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy"

	"github.com/primadi/lokstra/core/deploy/schema"
)

// TestRegisterLazyService_StringFactoryType tests the new unified API
// that accepts both string factory type and inline functions
func TestRegisterLazyService_StringFactoryType(t *testing.T) {
	reg := deploy.NewGlobalRegistry()

	// Register a factory type first
	reg.RegisterRouterServiceType("user-factory",
		func(deps, cfg map[string]any) any {
			return map[string]any{
				"type": "user-service",
				"deps": deps,
				"cfg":  cfg,
			}
		},
		nil, // No remote factory
		&deploy.ServiceTypeConfig{
			PathPrefix: "/api/users",
		},
	)

	// Register another factory for dependency
	reg.RegisterServiceType("user-repo-factory",
		func(deps, cfg map[string]any) any {
			return map[string]any{
				"type": "user-repository",
			}
		},
	)

	// Register dependency first
	reg.RegisterLazyService("user-repository", "user-repo-factory", nil)

	// Now register service using STRING factory type (like YAML)
	reg.RegisterLazyService("user-service", "user-factory", map[string]any{
		"depends-on": []string{"user-repository"},
		"max-users":  1000,
	})

	// Verify service is registered
	if !reg.HasLazyService("user-service") {
		t.Fatal("user-service not registered")
	}

	// Verify definition is stored correctly
	def := reg.GetDeferredServiceDef("user-service")
	if def == nil {
		t.Fatal("user-service definition not found")
	}

	if def.Type != "user-factory" {
		t.Errorf("expected type 'user-factory', got '%s'", def.Type)
	}

	if len(def.DependsOn) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(def.DependsOn))
	}

	if def.Config["max-users"] != 1000 {
		t.Errorf("expected max-users=1000, got %v", def.Config["max-users"])
	}

	// Verify service can be retrieved and instantiated
	svc, ok := reg.GetServiceAny("user-service")
	if !ok {
		t.Fatal("user-service not found during retrieval")
	}

	if svc == nil {
		t.Fatal("user-service is nil")
	}

	// Verify the instance is correct
	svcMap, ok := svc.(map[string]any)
	if !ok {
		t.Fatal("user-service is not a map")
	}

	if svcMap["type"] != "user-service" {
		t.Errorf("expected type 'user-service', got '%v'", svcMap["type"])
	}

	// Verify dependencies were resolved
	deps, ok := svcMap["deps"].(map[string]any)
	if !ok || deps == nil {
		t.Fatal("dependencies not resolved")
	}

	if deps["user-repository"] == nil {
		t.Error("user-repository dependency not resolved")
	}

	// Verify config was passed
	cfg, ok := svcMap["cfg"].(map[string]any)
	if !ok || cfg == nil {
		t.Fatal("config not passed")
	}

	if cfg["max-users"] != 1000 {
		t.Errorf("expected max-users=1000 in config, got %v", cfg["max-users"])
	}
}

// TestRegisterLazyService_InlineFunction tests inline function registration
func TestRegisterLazyService_InlineFunction(t *testing.T) {
	reg := deploy.NewGlobalRegistry()

	// Register service with inline function (no string factory type)
	reg.RegisterLazyService("cache", func(deps, cfg map[string]any) any {
		return map[string]any{
			"type": "redis-cache",
			"addr": cfg["addr"],
		}
	}, map[string]any{
		"addr": "localhost:6379",
	})

	// Verify service is registered
	if !reg.HasLazyService("cache") {
		t.Fatal("cache not registered")
	}

	// Retrieve service
	svc, ok := reg.GetServiceAny("cache")
	if !ok {
		t.Fatal("cache not found during retrieval")
	}

	// Verify the instance
	svcMap, ok := svc.(map[string]any)
	if !ok {
		t.Fatal("cache is not a map")
	}

	if svcMap["type"] != "redis-cache" {
		t.Errorf("expected type 'redis-cache', got '%v'", svcMap["type"])
	}

	if svcMap["addr"] != "localhost:6379" {
		t.Errorf("expected addr 'localhost:6379', got '%v'", svcMap["addr"])
	}
}

// TestRegisterLazyService_YAMLEquivalence tests that string factory type
// registration is equivalent to YAML service-definitions
func TestRegisterLazyService_YAMLEquivalence(t *testing.T) {
	reg := deploy.NewGlobalRegistry()

	// Register factory type
	reg.RegisterServiceType("order-factory",
		func(deps, cfg map[string]any) any {
			return map[string]any{
				"type":   "order-service",
				"maxOrd": cfg["max-orders"],
			}
		},
	)

	// Simulate YAML loading (what loader/builder.go does)
	yamlServiceDef := &schema.ServiceDef{
		Name: "order-service",
		Type: "order-factory",
		Config: map[string]any{
			"max-orders": 500,
		},
	}

	// Load via new unified API (simulating what builder.go now does)
	configMap := make(map[string]any)
	for k, v := range yamlServiceDef.Config {
		configMap[k] = v
	}
	if len(yamlServiceDef.DependsOn) > 0 {
		configMap["depends-on"] = yamlServiceDef.DependsOn
	}

	reg.RegisterLazyService(yamlServiceDef.Name, yamlServiceDef.Type, configMap)

	// Verify equivalence
	def := reg.GetDeferredServiceDef("order-service")
	if def == nil {
		t.Fatal("order-service definition not found")
	}

	if def.Name != yamlServiceDef.Name {
		t.Errorf("name mismatch: expected '%s', got '%s'", yamlServiceDef.Name, def.Name)
	}

	if def.Type != yamlServiceDef.Type {
		t.Errorf("type mismatch: expected '%s', got '%s'", yamlServiceDef.Type, def.Type)
	}

	// Verify service works
	svc, ok := reg.GetServiceAny("order-service")
	if !ok {
		t.Fatal("order-service not found")
	}

	svcMap := svc.(map[string]any)
	if svcMap["maxOrd"] != 500 {
		t.Errorf("expected maxOrd=500, got %v", svcMap["maxOrd"])
	}
}

// TestRegisterLazyService_DependsOnParsing tests depends-on parsing
// from both []string and []any (YAML unmarshaling)
func TestRegisterLazyService_DependsOnParsing(t *testing.T) {
	reg := deploy.NewGlobalRegistry()

	// Test with []string (Go native)
	reg.RegisterLazyService("service-a", "factory-a", map[string]any{
		"depends-on": []string{"dep1", "dep2"},
	})

	defA := reg.GetDeferredServiceDef("service-a")
	if defA == nil || len(defA.DependsOn) != 2 {
		t.Error("service-a depends-on not parsed correctly from []string")
	}

	// Test with []any (YAML unmarshaling)
	reg.RegisterLazyService("service-b", "factory-b", map[string]any{
		"depends-on": []any{"dep3", "dep4"},
	})

	defB := reg.GetDeferredServiceDef("service-b")
	if defB == nil || len(defB.DependsOn) != 2 {
		t.Error("service-b depends-on not parsed correctly from []any")
	}

	if defB.DependsOn[0] != "dep3" || defB.DependsOn[1] != "dep4" {
		t.Errorf("service-b depends-on values incorrect: %v", defB.DependsOn)
	}
}
