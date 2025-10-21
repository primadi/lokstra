package old_registry

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/primadi/lokstra/core/request"
)

type TestCounterService struct {
	count int
}

// Test concurrent access to service registry - verifies thread-safety
func TestConcurrentServiceAccess(t *testing.T) {
	// Clear registries
	serviceRegistry = sync.Map{}
	lazyServiceConfigRegistry = sync.Map{}
	serviceFactoryRegistry = sync.Map{}

	// Register factory
	RegisterServiceType("counter", func(cfg map[string]any) any {
		time.Sleep(10 * time.Millisecond) // Simulate slow creation
		return &TestCounterService{count: 0}
	}, AllowOverride(true))

	// Register lazy service
	RegisterLazyService("counter-svc", "counter", map[string]any{}, AllowOverride(true))

	// Launch 100 goroutines trying to get the same service
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make([]*TestCounterService, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			svc := GetService[*TestCounterService]("counter-svc")
			results[idx] = svc
		}(i)
	}

	wg.Wait()

	// All goroutines should get the SAME instance (singleton)
	firstInstance := results[0]
	for i := 1; i < numGoroutines; i++ {
		if results[i] != firstInstance {
			t.Errorf("Expected same instance, got different at index %d", i)
		}
	}

	t.Logf("✓ All %d goroutines got the same service instance (thread-safe singleton)", numGoroutines)
}

// Test concurrent registration of different services
func TestConcurrentServiceRegistration(t *testing.T) {
	// Clear registries
	serviceRegistry = sync.Map{}
	lazyServiceConfigRegistry = sync.Map{}
	serviceFactoryRegistry = sync.Map{}

	// Register factory
	RegisterServiceType("counter", func(cfg map[string]any) any {
		return &TestCounterService{count: 0}
	}, AllowOverride(true))

	const numServices = 100
	var wg sync.WaitGroup
	wg.Add(numServices)

	// Register many services concurrently
	for i := 0; i < numServices; i++ {
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("counter-%d", idx)
			RegisterLazyService(name, "counter", map[string]any{}, AllowOverride(true))
		}(i)
	}

	wg.Wait()

	// Verify all services are registered
	count := 0
	lazyServiceConfigRegistry.Range(func(key, value any) bool {
		count++
		return true
	})

	if count != numServices {
		t.Errorf("Expected %d services, got %d", numServices, count)
	}

	t.Logf("✓ Successfully registered %d services concurrently (thread-safe registration)", count)
}

// Test concurrent config access
func TestConcurrentConfigAccess(t *testing.T) {
	// Clear config registry
	configRegistry = sync.Map{}

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // readers and writers

	// Writers
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", idx)
			SetConfig(key, idx)
		}(i)
	}

	// Readers
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			time.Sleep(5 * time.Millisecond) // Let writers run first
			key := fmt.Sprintf("key-%d", idx)
			_ = GetConfig(key, 0)
		}(i)
	}

	wg.Wait()

	// Verify all configs are written
	count := 0
	configRegistry.Range(func(key, value any) bool {
		count++
		return true
	})

	if count != numGoroutines {
		t.Errorf("Expected %d configs, got %d", numGoroutines, count)
	}

	t.Logf("✓ Successfully handled %d concurrent config reads/writes (thread-safe)", numGoroutines*2)
}

// Test concurrent middleware access
func TestConcurrentMiddlewareAccess(t *testing.T) {
	// Clear middleware registries
	mwFactoryRegistry = sync.Map{}
	mwEntryRegistry = sync.Map{}

	// Register middleware factory
	RegisterMiddlewareFactory("test-mw", func(cfg map[string]any) request.HandlerFunc {
		return func(c *request.Context) error {
			return nil
		}
	}, AllowOverride(true))

	// Register middleware entries
	const numMiddlewares = 50
	var wg sync.WaitGroup
	wg.Add(numMiddlewares)

	for i := 0; i < numMiddlewares; i++ {
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("mw-%d", idx)
			RegisterMiddlewareName(name, "test-mw", map[string]any{}, AllowOverride(true))
		}(i)
	}

	wg.Wait()

	// Create middlewares concurrently
	wg.Add(numMiddlewares)
	for i := 0; i < numMiddlewares; i++ {
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("mw-%d", idx)
			mw := CreateMiddleware(name)
			if mw == nil {
				t.Errorf("Expected middleware, got nil for %s", name)
			}
		}(i)
	}

	wg.Wait()
	t.Logf("✓ Successfully handled %d concurrent middleware operations (thread-safe)", numMiddlewares*2)
}
