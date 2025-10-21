package lokstra_registry_test

import (
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Mock service for testing
type MockUserService struct {
	Name string
}

func (s *MockUserService) GetName() string {
	return s.Name
}

func TestRegisterAndGetService(t *testing.T) {
	// Register a service
	mockService := &MockUserService{Name: "test-user-service"}
	lokstra_registry.RegisterService("user-service", mockService)

	// Get service with generic function
	retrieved := lokstra_registry.GetService[*MockUserService]("user-service")
	if retrieved == nil {
		t.Fatal("expected service to be retrieved, got nil")
	}

	if retrieved.GetName() != "test-user-service" {
		t.Errorf("expected name 'test-user-service', got '%s'", retrieved.GetName())
	}
}

func TestMustGetService(t *testing.T) {
	// Register a service
	mockService := &MockUserService{Name: "must-service"}
	lokstra_registry.RegisterService("must-service", mockService)

	// MustGetService should not panic
	retrieved := lokstra_registry.MustGetService[*MockUserService]("must-service")
	if retrieved.Name != "must-service" {
		t.Errorf("expected name 'must-service', got '%s'", retrieved.Name)
	}
}

func TestMustGetService_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when service not found")
		}
	}()

	// This should panic
	lokstra_registry.MustGetService[*MockUserService]("nonexistent-service")
}

func TestTryGetService(t *testing.T) {
	// Register a service
	mockService := &MockUserService{Name: "try-service"}
	lokstra_registry.RegisterService("try-service", mockService)

	// Try to get existing service
	retrieved, ok := lokstra_registry.TryGetService[*MockUserService]("try-service")
	if !ok {
		t.Fatal("expected service to be found")
	}
	if retrieved.Name != "try-service" {
		t.Errorf("expected name 'try-service', got '%s'", retrieved.Name)
	}

	// Try to get nonexistent service
	_, ok = lokstra_registry.TryGetService[*MockUserService]("nonexistent")
	if ok {
		t.Error("expected service not to be found")
	}
}

func TestGetServiceAny(t *testing.T) {
	// Register a service
	mockService := &MockUserService{Name: "any-service"}
	lokstra_registry.RegisterService("any-service", mockService)

	// Get service as any
	retrieved, ok := lokstra_registry.GetServiceAny("any-service")
	if !ok {
		t.Fatal("expected service to be found")
	}

	// Type assert
	if typed, ok := retrieved.(*MockUserService); ok {
		if typed.Name != "any-service" {
			t.Errorf("expected name 'any-service', got '%s'", typed.Name)
		}
	} else {
		t.Error("expected service to be *MockUserService")
	}
}

func TestRegisterAndGetMiddleware(t *testing.T) {
	// Create a mock middleware
	mockMiddleware := func(ctx *request.Context) error {
		ctx.Set("middleware-called", true)
		return nil
	}

	// Register middleware
	lokstra_registry.RegisterMiddleware("test-middleware", mockMiddleware)

	// Get middleware
	retrieved, ok := lokstra_registry.GetMiddleware("test-middleware")
	if !ok {
		t.Fatal("expected middleware to be found")
	}

	// Test middleware execution
	ctx := &request.Context{}
	err := retrieved(ctx)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if ctx.Get("middleware-called") != true {
		t.Error("expected middleware to set context value")
	}
}

func TestConfigDefineAndGet(t *testing.T) {
	// Define config
	lokstra_registry.DefineConfig("test-config", "test-value")

	// Resolve configs to make them available
	err := lokstra_registry.Global().ResolveConfigs()
	if err != nil {
		t.Fatalf("failed to resolve configs: %v", err)
	}

	// Get config with default
	value := lokstra_registry.GetConfig("test-config", "default")
	if value != "test-value" {
		t.Errorf("expected 'test-value', got '%s'", value)
	}

	// Get nonexistent config (should return default)
	defaultValue := lokstra_registry.GetConfig("nonexistent", "default-value")
	if defaultValue != "default-value" {
		t.Errorf("expected 'default-value', got '%s'", defaultValue)
	}
}

func TestGetConfig_TypeAssertion(t *testing.T) {
	// Define config with int value
	lokstra_registry.DefineConfig("int-config", 42)

	// Resolve configs
	err := lokstra_registry.Global().ResolveConfigs()
	if err != nil {
		t.Fatalf("failed to resolve configs: %v", err)
	}

	// Get as int
	intValue := lokstra_registry.GetConfig("int-config", 0)
	if intValue != 42 {
		t.Errorf("expected 42, got %d", intValue)
	}

	// Get as string (type mismatch, should return default)
	strValue := lokstra_registry.GetConfig("int-config", "default")
	if strValue != "default" {
		t.Errorf("expected 'default' due to type mismatch, got '%s'", strValue)
	}
}

func TestGlobal(t *testing.T) {
	// Verify Global() returns the same instance
	g1 := lokstra_registry.Global()
	g2 := lokstra_registry.Global()

	if g1 != g2 {
		t.Error("expected Global() to return the same instance")
	}
}
