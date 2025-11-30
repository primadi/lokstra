package deploy_test

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy"
)

// Test RegisterRouterServiceType with all three factory signatures
func TestRegisterRouterServiceType_ThreeModes(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	// Mode 1: func() any (simplest)
	g.RegisterRouterServiceType("simple-service",
		func() any {
			return "simple-local"
		},
		func() any {
			return "simple-remote"
		},
		nil,
	)

	// Mode 2: func(cfg map[string]any) any (config only)
	g.RegisterRouterServiceType("config-service",
		func(cfg map[string]any) any {
			return "local-" + cfg["env"].(string)
		},
		func(cfg map[string]any) any {
			return "remote-" + cfg["env"].(string)
		},
		nil,
	)

	// Mode 3: func(deps, cfg map[string]any) any (full control)
	g.RegisterRouterServiceType("full-service",
		func(deps, cfg map[string]any) any {
			return "full-local-" + cfg["mode"].(string)
		},
		func(deps, cfg map[string]any) any {
			return "full-remote-" + cfg["mode"].(string)
		},
		nil,
	)

	// Test local factory retrieval
	localSimple := g.GetServiceFactory("simple-service", true)
	if localSimple == nil {
		t.Fatal("simple-service local factory not found")
	}
	result1 := localSimple(nil, nil)
	if result1 != "simple-local" {
		t.Errorf("expected 'simple-local', got '%v'", result1)
	}

	// Test remote factory retrieval
	remoteSimple := g.GetServiceFactory("simple-service", false)
	if remoteSimple == nil {
		t.Fatal("simple-service remote factory not found")
	}
	result2 := remoteSimple(nil, nil)
	if result2 != "simple-remote" {
		t.Errorf("expected 'simple-remote', got '%v'", result2)
	}

	// Test config factory
	localConfig := g.GetServiceFactory("config-service", true)
	result3 := localConfig(nil, map[string]any{"env": "production"})
	if result3 != "local-production" {
		t.Errorf("expected 'local-production', got '%v'", result3)
	}

	// Test full factory
	localFull := g.GetServiceFactory("full-service", true)
	result4 := localFull(map[string]any{"dep": "test"}, map[string]any{"mode": "sync"})
	if result4 != "full-local-sync" {
		t.Errorf("expected 'full-local-sync', got '%v'", result4)
	}
}

// Test RegisterServiceType simple signature (infrastructure services)
func TestRegisterServiceType_Simple(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	// Simple infrastructure service registration
	g.RegisterServiceType("db-pool",
		func(cfg map[string]any) any {
			return "db-pool-instance"
		},
	)

	local := g.GetServiceFactory("db-pool", true)
	if local == nil {
		t.Fatal("db-pool factory not found")
	}

	result := local(nil, map[string]any{"dsn": "test"})
	if result != "db-pool-instance" {
		t.Errorf("expected 'db-pool-instance', got '%v'", result)
	}

	// Remote should be nil for simple registration
	remote := g.GetServiceFactory("db-pool", false)
	if remote != nil {
		t.Error("remote factory should be nil for simple RegisterServiceType")
	}
}

// Test RegisterRouterServiceType with nil remote factory
func TestRegisterRouterServiceType_NilRemote(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	g.RegisterRouterServiceType("local-only",
		func() any {
			return "local-service"
		},
		nil, // No remote factory
		nil, // No config
	)

	local := g.GetServiceFactory("local-only", true)
	if local == nil {
		t.Fatal("local factory not found")
	}

	remote := g.GetServiceFactory("local-only", false)
	if remote != nil {
		t.Error("remote factory should be nil")
	}
}

// Test RegisterRouterServiceType with metadata
func TestRegisterRouterServiceType_WithMetadata(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	g.RegisterRouterServiceType("user-service",
		func() any {
			return "user-service-instance"
		},
		nil,
		&deploy.ServiceTypeConfig{
			PathPrefix:  "/api/users",
			Middlewares: []string{"auth", "cors"},
		},
	)

	metadata := g.GetServiceMetadata("user-service")
	if metadata == nil {
		t.Fatal("metadata not found")
	}

	if metadata.PathPrefix != "/api/users" {
		t.Errorf("expected PathPrefix '/api/users', got '%s'", metadata.PathPrefix)
	}

	if len(metadata.MiddlewareNames) != 2 {
		t.Errorf("expected 2 middlewares, got %d", len(metadata.MiddlewareNames))
	}
}

// Test invalid factory signature
func TestRegisterServiceType_InvalidSignature(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid factory signature")
		}
	}()

	// Invalid signature (string parameter)
	g.RegisterServiceType("invalid",
		func(s string) any {
			return s
		},
	)
}
