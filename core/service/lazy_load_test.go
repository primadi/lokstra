package service_test

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

type TestService struct {
	Name string
}

func TestLazyLoad_WithGlobalRegistry(t *testing.T) {
	// Initialize global registry
	_ = deploy.Global()

	// Register a service
	testSvc := &TestService{Name: "test-service"}
	lokstra_registry.RegisterService("lazy-test-service", testSvc)

	// Create lazy loader
	lazy := service.LazyLoad[*TestService]("lazy-test-service")

	// Verify not loaded yet
	if lazy.IsLoaded() {
		t.Error("expected service not to be loaded yet")
	}

	// Get service (should load from registry)
	retrieved := lazy.Get()
	if retrieved == nil {
		t.Fatal("expected service to be retrieved")
	}

	if retrieved.Name != "test-service" {
		t.Errorf("expected name 'test-service', got '%s'", retrieved.Name)
	}

	// Verify now loaded
	if !lazy.IsLoaded() {
		t.Error("expected service to be loaded after Get()")
	}

	// Get again (should return cached)
	retrieved2 := lazy.Get()
	if retrieved2 != retrieved {
		t.Error("expected same instance (cached)")
	}
}

func TestLazyLoad_NotFound(t *testing.T) {
	// Initialize global registry
	_ = deploy.Global()

	// Create lazy loader for non-existent service
	lazy := service.LazyLoad[*TestService]("non-existent")

	// Should return nil (zero value)
	retrieved := lazy.Get()
	if retrieved != nil {
		t.Errorf("expected nil for non-existent service, got %v", retrieved)
	}
}

func TestLazyLoadWith_CustomLoader(t *testing.T) {
	called := false
	testSvc := &TestService{Name: "custom"}

	lazy := service.LazyLoadWith(func() *TestService {
		called = true
		return testSvc
	})

	// Verify not called yet
	if called {
		t.Error("expected loader not to be called yet")
	}

	// Get service
	retrieved := lazy.Get()
	if !called {
		t.Error("expected loader to be called")
	}

	if retrieved.Name != "custom" {
		t.Errorf("expected name 'custom', got '%s'", retrieved.Name)
	}

	// Reset flag and get again (should NOT call loader again)
	called = false
	lazy.Get()
	if called {
		t.Error("expected loader not to be called second time (should use cache)")
	}
}

func TestMustGet_Success(t *testing.T) {
	_ = deploy.Global()

	testSvc := &TestService{Name: "must-test"}
	lokstra_registry.RegisterService("must-test-service", testSvc)

	lazy := service.LazyLoad[*TestService]("must-test-service")
	retrieved := lazy.MustGet()

	if retrieved.Name != "must-test" {
		t.Errorf("expected name 'must-test', got '%s'", retrieved.Name)
	}
}

func TestMustGet_Panic(t *testing.T) {
	_ = deploy.Global()

	lazy := service.LazyLoad[*TestService]("does-not-exist")

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when service not found")
		}
	}()

	lazy.MustGet()
}

func TestValue(t *testing.T) {
	testSvc := &TestService{Name: "preloaded"}
	cached := service.Value(testSvc)

	// Should be marked as loaded
	if !cached.IsLoaded() {
		t.Error("expected Value() to mark as loaded")
	}

	// Get should return the preloaded value
	retrieved := cached.Get()
	if retrieved.Name != "preloaded" {
		t.Errorf("expected name 'preloaded', got '%s'", retrieved.Name)
	}
}

func TestCast(t *testing.T) {
	testSvc := &TestService{Name: "cast-test"}
	cached := service.Value[any](testSvc)

	// Cast to specific type
	typedCached := service.Cast[*TestService](cached)
	retrieved := typedCached.Get()

	if retrieved.Name != "cast-test" {
		t.Errorf("expected name 'cast-test', got '%s'", retrieved.Name)
	}
}
