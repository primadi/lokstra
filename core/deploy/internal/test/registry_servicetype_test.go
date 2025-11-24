package deploy_test

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy"
)

// Test RegisterServiceType with all three factory signatures
func TestRegisterServiceType_ThreeModes(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	// Mode 1: func() any (simplest)
	g.RegisterServiceType("simple-service",
		func() any {
			return "simple-local"
		},
		func() any {
			return "simple-remote"
		},
	)

	// Mode 2: func(cfg map[string]any) any (config only)
	g.RegisterServiceType("config-service",
		func(cfg map[string]any) any {
			return "local-" + cfg["env"].(string)
		},
		func(cfg map[string]any) any {
			return "remote-" + cfg["env"].(string)
		},
	)

	// Mode 3: func(deps, cfg map[string]any) any (full control)
	g.RegisterServiceType("full-service",
		func(deps, cfg map[string]any) any {
			return "full-local-" + cfg["mode"].(string)
		},
		func(deps, cfg map[string]any) any {
			return "full-remote-" + cfg["mode"].(string)
		},
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

// Test RegisterServiceType with nil remote factory
func TestRegisterServiceType_NilRemote(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	g.RegisterServiceType("local-only",
		func() any {
			return "local-service"
		},
		nil, // No remote factory
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

// Test RegisterServiceType with metadata
func TestRegisterServiceType_WithMetadata(t *testing.T) {
	g := deploy.NewGlobalRegistry()

	g.RegisterServiceType("user-service",
		func() any {
			return "user-service-instance"
		},
		nil,
		deploy.WithResource("user", "users"),
		deploy.WithConvention("rest"),
	)

	metadata := g.GetServiceMetadata("user-service")
	if metadata == nil {
		t.Fatal("metadata not found")
	}

	if metadata.Resource != "user" {
		t.Errorf("expected resource 'user', got '%s'", metadata.Resource)
	}

	if metadata.ResourcePlural != "users" {
		t.Errorf("expected resource plural 'users', got '%s'", metadata.ResourcePlural)
	}

	if metadata.Convention != "rest" {
		t.Errorf("expected convention 'rest', got '%s'", metadata.Convention)
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
		nil,
	)
}
