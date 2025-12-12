package deploy_test

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy"
)

// Test RegisterLazyService with all three factory signatures
func TestRegisterLazyService_ThreeModes(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	// Mode 1: func() any (simplest)
	g.RegisterLazyService("simple-service", func() any {
		return "simple-result"
	}, nil)

	// Mode 2: func(cfg map[string]any) any (config only)
	g.RegisterLazyService("config-service", func(cfg map[string]any) any {
		return cfg["value"]
	}, map[string]any{"value": "config-result"})

	// Mode 3: func(deps, cfg map[string]any) any (full control)
	g.RegisterLazyService("full-service", func(deps, cfg map[string]any) any {
		// deps should be nil for lazy services (they resolve via GetService)
		if deps != nil {
			t.Error("deps should be nil for lazy services")
		}
		return "full-" + cfg["suffix"].(string)
	}, map[string]any{"suffix": "result"})

	// Retrieve simple service
	result1, ok := g.GetServiceAny("simple-service")
	if !ok {
		t.Fatal("simple-service not found")
	}
	if result1 != "simple-result" {
		t.Errorf("expected 'simple-result', got '%v'", result1)
	}

	// Retrieve config service
	result2, ok := g.GetServiceAny("config-service")
	if !ok {
		t.Fatal("config-service not found")
	}
	if result2 != "config-result" {
		t.Errorf("expected 'config-result', got '%v'", result2)
	}

	// Retrieve full service
	result3, ok := g.GetServiceAny("full-service")
	if !ok {
		t.Fatal("full-service not found")
	}
	if result3 != "full-result" {
		t.Errorf("expected 'full-result', got '%v'", result3)
	}
}

// Test dependency resolution with all factory modes
func TestRegisterLazyService_Dependencies(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	// Database service (with config - mode 2)
	g.RegisterLazyService("db", func(cfg map[string]any) any {
		return map[string]any{
			"type": "postgres",
			"dsn":  cfg["dsn"],
		}
	}, map[string]any{"dsn": "postgresql://localhost/test"})

	// Repository service (no params - mode 1)
	g.RegisterLazyService("user-repo", func() any {
		db, _ := g.GetServiceAny("db")
		dbMap := db.(map[string]any)
		return "UserRepo connected to " + dbMap["dsn"].(string)
	}, nil)

	// Service with repository dependency (full signature - mode 3)
	g.RegisterLazyService("user-service", func(deps, cfg map[string]any) any {
		repo, _ := g.GetServiceAny("user-repo")
		return "UserService using " + repo.(string)
	}, nil)

	// Access user-service (should resolve all dependencies)
	result, ok := g.GetServiceAny("user-service")
	if !ok {
		t.Fatal("user-service not found")
	}

	expected := "UserService using UserRepo connected to postgresql://localhost/test"
	if result != expected {
		t.Errorf("expected '%s', got '%v'", expected, result)
	}
}

// Test invalid factory signature
func TestRegisterLazyService_InvalidSignature(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid factory signature")
		}
	}()

	// Invalid signature (string parameter instead of map)
	g.RegisterLazyService("invalid", func(s string) any {
		return s
	}, nil)
}

// Test singleton behavior
func TestRegisterLazyService_Singleton(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	counter := 0
	g.RegisterLazyService("counter", func() any {
		counter++
		return counter
	}, nil)

	// First access
	result1, _ := g.GetServiceAny("counter")
	if result1 != 1 {
		t.Errorf("expected 1, got %v", result1)
	}

	// Second access (should return cached value)
	result2, _ := g.GetServiceAny("counter")
	if result2 != 1 {
		t.Errorf("expected 1 (cached), got %v", result2)
	}

	// Factory should only be called once
	if counter != 1 {
		t.Errorf("factory called %d times, expected 1", counter)
	}
}

// Test multiple instances with different configs
func TestRegisterLazyService_MultipleInstances(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	// Register multiple DB instances with different DSN
	g.RegisterLazyService("db_main", func(cfg map[string]any) any {
		return "Connection to " + cfg["dsn"].(string)
	}, map[string]any{"dsn": "main-db"})

	g.RegisterLazyService("db-analytics", func(cfg map[string]any) any {
		return "Connection to " + cfg["dsn"].(string)
	}, map[string]any{"dsn": "analytics-db"})

	g.RegisterLazyService("db-cache", func(cfg map[string]any) any {
		return "Connection to " + cfg["dsn"].(string)
	}, map[string]any{"dsn": "cache-db"})

	// Access each instance
	main, _ := g.GetServiceAny("db_main")
	if main != "Connection to main-db" {
		t.Errorf("unexpected db_main: %v", main)
	}

	analytics, _ := g.GetServiceAny("db-analytics")
	if analytics != "Connection to analytics-db" {
		t.Errorf("unexpected db-analytics: %v", analytics)
	}

	cache, _ := g.GetServiceAny("db-cache")
	if cache != "Connection to cache-db" {
		t.Errorf("unexpected db-cache: %v", cache)
	}
}
