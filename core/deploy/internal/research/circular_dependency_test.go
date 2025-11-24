package research

import (
	"fmt"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/service"
)

// Test to PROVE that Cached does NOT prevent circular dependency crash
func TestCircularDependency_StillCrashes(t *testing.T) {
	reg := deploy.NewGlobalRegistry()

	// Service A depends on B
	reg.RegisterServiceType("service-a-factory",
		func(deps, cfg map[string]any) any {
			return &struct {
				B *service.Cached[any]
			}{
				B: service.Cast[any](deps["service-b"]),
			}
		},
		nil,
	)

	// Service B depends on A (CIRCULAR!)
	reg.RegisterServiceType("service-b-factory",
		func(deps, cfg map[string]any) any {
			return &struct {
				A *service.Cached[any]
			}{
				A: service.Cast[any](deps["service-a"]),
			}
		},
		nil,
	)

	// Register with circular dependencies
	reg.RegisterLazyService("service-a", "service-a-factory", map[string]any{
		"depends-on": []string{"service-b"},
	})

	reg.RegisterLazyService("service-b", "service-b-factory", map[string]any{
		"depends-on": []string{"service-a"},
	})

	// Try to access service-a
	defer func() {
		if r := recover(); r != nil {
			panicMsg := fmt.Sprint(r)
			t.Logf("✅ PROOF: Circular dependency PANICS with clear error message")
			t.Logf("   Panic message: %s", panicMsg)

			// Verify panic message contains circular dependency info
			if !strings.Contains(panicMsg, "circular dependency detected") {
				t.Errorf("Expected panic message to contain 'circular dependency detected', got: %s", panicMsg)
			}
			if !strings.Contains(panicMsg, "service-a") && !strings.Contains(panicMsg, "service-b") {
				t.Errorf("Expected panic message to show service names, got: %s", panicMsg)
			}
			return
		}
		t.Fatal("❌ Expected panic from circular dependency, but got none")
	}()

	// This WILL panic due to circular dependency
	_, _ = reg.GetServiceAny("service-a")
}

// Test to show when Cached IS actually useful
func TestCached_UsefulForConditionalLoading(t *testing.T) {
	reg := deploy.NewGlobalRegistry()

	var expensiveAPICalled bool

	// Cheap service
	reg.RegisterServiceType("cache-factory",
		func(deps, cfg map[string]any) any {
			return &struct{ Name string }{Name: "Cache"}
		},
		nil,
	)

	// Expensive service
	reg.RegisterServiceType("expensive-api-factory",
		func(deps, cfg map[string]any) any {
			expensiveAPICalled = true
			t.Logf("💰 Expensive API factory called")
			return &struct{ Name string }{Name: "ExpensiveAPI"}
		},
		nil,
	)

	// Analytics service with LAZY dependencies
	reg.RegisterServiceType("analytics-factory",
		func(deps, cfg map[string]any) any {
			return &struct {
				Cache        *service.Cached[any]
				ExpensiveAPI *service.Cached[any]
			}{
				Cache:        service.Cast[any](deps["cache"]),
				ExpensiveAPI: service.Cast[any](deps["expensive-api"]),
			}
		},
		nil,
	)

	reg.RegisterLazyService("cache", "cache-factory", nil)
	reg.RegisterLazyService("expensive-api", "expensive-api-factory", nil)
	reg.RegisterLazyService("analytics", "analytics-factory", map[string]any{
		"depends-on": []string{"cache", "expensive-api"},
	})

	// Access analytics service
	svc, ok := reg.GetServiceAny("analytics")
	if !ok {
		t.Fatal("analytics service not found")
	}

	// ✅ PROOF: expensive-api was created during analytics creation
	if !expensiveAPICalled {
		t.Fatal("Expected expensive-api to be called during dependency resolution")
	}
	t.Logf("✅ PROOF: ExpensiveAPI was created when analytics service was created")
	t.Logf("   (NOT lazy at dependency level - loaded during service creation)")

	_ = svc
}

// Test to show EAGER injection also works fine
func TestEagerInjection_WorksNormally(t *testing.T) {
	reg := deploy.NewGlobalRegistry()

	var paymentAPICalled bool
	var emailSvcCalled bool

	// Payment API
	reg.RegisterServiceType("payment-api-factory",
		func(deps, cfg map[string]any) any {
			paymentAPICalled = true
			return &struct{ Name string }{Name: "PaymentAPI"}
		},
		nil,
	)

	// Email Service
	reg.RegisterServiceType("email-service-factory",
		func(deps, cfg map[string]any) any {
			emailSvcCalled = true
			return &struct{ Name string }{Name: "EmailService"}
		},
		nil,
	)

	// Order service with EAGER dependencies (no service.Cached)
	type OrderService struct {
		PaymentAPI any // Direct injection
		EmailSvc   any // Direct injection
	}

	reg.RegisterServiceType("order-service-factory",
		func(deps, cfg map[string]any) any {
			return &OrderService{
				PaymentAPI: deps["payment-api"],   // Direct, no Cast
				EmailSvc:   deps["email-service"], // Direct, no Cast
			}
		},
		nil,
	)

	reg.RegisterLazyService("payment-api", "payment-api-factory", nil)
	reg.RegisterLazyService("email-service", "email-service-factory", nil)
	reg.RegisterLazyService("order-service", "order-service-factory", map[string]any{
		"depends-on": []string{"payment-api", "email-service"},
	})

	// Access order service
	svc, ok := reg.GetServiceAny("order-service")
	if !ok {
		t.Fatal("order-service not found")
	}

	orderSvc := svc.(*OrderService)

	// ✅ PROOF: Both dependencies created when order-service created
	if !paymentAPICalled {
		t.Fatal("Expected payment-api to be called")
	}
	if !emailSvcCalled {
		t.Fatal("Expected email-service to be called")
	}

	// ✅ PROOF: Direct access works (no .MustGet() needed)
	if orderSvc.PaymentAPI == nil {
		t.Fatal("Expected PaymentAPI to be available")
	}
	if orderSvc.EmailSvc == nil {
		t.Fatal("Expected EmailSvc to be available")
	}

	t.Logf("✅ PROOF: Eager injection works fine")
	t.Logf("   PaymentAPI and EmailService both created when OrderService created")
	t.Logf("   Direct access without .MustGet() works perfectly")
}
