package deploy_test

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

// TestMiddlewareFactory_UnnamedFuncType tests the fix for unnamed function type conversion
// This was causing panic in lokstra-auth project when middleware factory returned
// func(*request.Context) error instead of request.HandlerFunc
func TestMiddlewareFactory_UnnamedFuncType(t *testing.T) {
	// Reset registry
	deploy.ResetGlobalRegistryForTesting()

	// Register middleware factory that returns UNNAMED func type
	// This is the signature that was causing panic before the fix
	lokstra_registry.RegisterMiddlewareFactory("test-unnamed", func(config map[string]any) func(*request.Context) error {
		return func(ctx *request.Context) error {
			// Middleware logic
			return nil
		}
	})

	// Create middleware - should NOT panic with the fix
	mw := deploy.Global().CreateMiddleware("test-unnamed")

	if mw == nil {
		t.Fatal("expected middleware to be created, got nil (type conversion failed)")
	}

	// Verify type is request.HandlerFunc after conversion
	if _, ok := any(mw).(request.HandlerFunc); !ok {
		t.Errorf("expected request.HandlerFunc after conversion, got %T", mw)
	}
}

// TestMiddlewareFactory_NamedHandlerFunc tests that named type still works
func TestMiddlewareFactory_NamedHandlerFunc(t *testing.T) {
	// Reset registry
	deploy.ResetGlobalRegistryForTesting()

	// Register middleware factory that returns NAMED HandlerFunc type
	lokstra_registry.RegisterMiddlewareFactory("test-named", func(config map[string]any) request.HandlerFunc {
		return func(ctx *request.Context) error {
			return nil
		}
	})

	// Create middleware
	mw := deploy.Global().CreateMiddleware("test-named")

	if mw == nil {
		t.Fatal("expected middleware to be created, got nil")
	}

	// Verify type is request.HandlerFunc
	if _, ok := any(mw).(request.HandlerFunc); !ok {
		t.Errorf("expected request.HandlerFunc, got %T", mw)
	}
}

// TestMiddlewareFactory_UnnamedWithParams tests unnamed type with inline params
func TestMiddlewareFactory_UnnamedWithParams(t *testing.T) {
	// Reset registry
	deploy.ResetGlobalRegistryForTesting()

	var capturedConfig map[string]any

	// Register middleware factory with config
	lokstra_registry.RegisterMiddlewareFactory("test-params", func(config map[string]any) func(*request.Context) error {
		capturedConfig = config
		return func(ctx *request.Context) error {
			return nil
		}
	})

	// Create middleware with inline params
	mw := deploy.Global().CreateMiddleware("test-params key=value, num=123")

	if mw == nil {
		t.Fatal("expected middleware to be created, got nil")
	}

	// Verify config was captured
	if capturedConfig == nil {
		t.Fatal("expected config to be passed to factory")
	}

	if capturedConfig["key"] != "value" {
		t.Errorf("expected key='value', got %v", capturedConfig["key"])
	}

	if capturedConfig["num"] != "123" {
		t.Errorf("expected num='123', got %v", capturedConfig["num"])
	}
}

// TestMiddlewareFactory_RegisteredName tests unnamed type via RegisterMiddlewareName
func TestMiddlewareFactory_RegisteredName(t *testing.T) {
	// Reset registry
	deploy.ResetGlobalRegistryForTesting()

	// Register factory type (unnamed)
	lokstra_registry.RegisterMiddlewareFactory("auth-factory", func(config map[string]any) func(*request.Context) error {
		return func(ctx *request.Context) error {
			return nil
		}
	})

	// Register middleware name with config
	lokstra_registry.RegisterMiddlewareName("my-auth", "auth-factory", map[string]any{
		"enabled": true,
	})

	// Create middleware by name - should use RegisterMiddlewareName flow
	mw := deploy.Global().CreateMiddleware("my-auth")

	if mw == nil {
		t.Fatal("expected middleware to be created from registered name, got nil")
	}

	// Verify type conversion worked
	if _, ok := any(mw).(request.HandlerFunc); !ok {
		t.Errorf("expected request.HandlerFunc after conversion, got %T", mw)
	}
}
