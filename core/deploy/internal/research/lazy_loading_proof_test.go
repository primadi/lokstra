package research

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy"
)

// Test to PROVE that lazy loading only creates services on-demand
func TestLazyLoadingProof_OnDemandCreation(t *testing.T) {
	reg := deploy.NewGlobalRegistry()

	// Track which factories were called
	var callLog []string

	// Register 3 services with tracking
	reg.RegisterServiceType("service-a-factory",
		func(deps, cfg map[string]any) any {
			callLog = append(callLog, "service-a-factory")
			return &struct{ Name string }{Name: "ServiceA"}
		},
		nil,
	)

	reg.RegisterServiceType("service-b-factory",
		func(deps, cfg map[string]any) any {
			callLog = append(callLog, "service-b-factory")
			return &struct{ Name string }{Name: "ServiceB"}
		},
		nil,
	)

	reg.RegisterServiceType("service-c-factory",
		func(deps, cfg map[string]any) any {
			callLog = append(callLog, "service-c-factory")
			return &struct{ Name string }{Name: "ServiceC"}
		},
		nil,
	)

	// Register lazy services
	reg.RegisterLazyService("service-a", "service-a-factory", nil)
	reg.RegisterLazyService("service-b", "service-b-factory", nil)
	reg.RegisterLazyService("service-c", "service-c-factory", nil)

	// ✅ PROOF 1: After registration, NO factories called yet
	if len(callLog) != 0 {
		t.Fatalf("Expected 0 factory calls after registration, got %d: %v", len(callLog), callLog)
	}
	t.Logf("✅ PROOF 1: After registration, NO factories called (callLog is empty)")

	// Access only service-a
	svcA, ok := reg.GetServiceAny("service-a")
	if !ok {
		t.Fatal("service-a not found")
	}

	// ✅ PROOF 2: Only service-a factory was called
	if len(callLog) != 1 || callLog[0] != "service-a-factory" {
		t.Fatalf("Expected only service-a-factory to be called, got: %v", callLog)
	}
	t.Logf("✅ PROOF 2: Only service-a factory was called: %v", callLog)

	// Access service-a again
	svcA2, ok := reg.GetServiceAny("service-a")
	if !ok {
		t.Fatal("service-a not found on second access")
	}

	// ✅ PROOF 3: Factory NOT called again (same instance returned)
	if len(callLog) != 1 {
		t.Fatalf("Expected factory to be called only once, got %d calls: %v", len(callLog), callLog)
	}
	if svcA != svcA2 {
		t.Fatal("Expected same instance on second access")
	}
	t.Logf("✅ PROOF 3: Factory NOT called again, same instance returned")

	// Access service-b
	svcB, ok := reg.GetServiceAny("service-b")
	if !ok {
		t.Fatal("service-b not found")
	}

	// ✅ PROOF 4: Now service-b factory was called (service-c still not called)
	if len(callLog) != 2 || callLog[1] != "service-b-factory" {
		t.Fatalf("Expected service-b-factory to be called, got: %v", callLog)
	}
	t.Logf("✅ PROOF 4: service-b factory was called, service-c still not created: %v", callLog)

	// ✅ PROOF 5: service-c is NEVER created (not accessed)
	expectedLog := []string{"service-a-factory", "service-b-factory"}
	if len(callLog) != len(expectedLog) {
		t.Fatalf("Expected only 2 factory calls, got %d: %v", len(callLog), callLog)
	}
	for i, expected := range expectedLog {
		if callLog[i] != expected {
			t.Fatalf("Expected callLog[%d] = %s, got %s", i, expected, callLog[i])
		}
	}
	t.Logf("✅ PROOF 5: service-c factory was NEVER called (on-demand loading confirmed)")

	// Final verification
	t.Logf("\n📊 Final Call Log: %v", callLog)
	t.Logf("📊 Services Created: 2/3 (66%% - only what was accessed)")
	t.Logf("📊 Services NOT Created: service-c (never accessed)")

	_ = svcA
	_ = svcB
}

// Test to PROVE that dependencies are resolved on-demand
func TestLazyLoadingProof_DependencyResolution(t *testing.T) {
	reg := deploy.NewGlobalRegistry()

	// Track resolution order
	var resolutionOrder []string

	// Register repository (leaf dependency)
	reg.RegisterServiceType("repository-factory",
		func(deps, cfg map[string]any) any {
			resolutionOrder = append(resolutionOrder, "repository")
			return &struct{ Name string }{Name: "Repository"}
		},
		nil,
	)

	// Register service (depends on repository)
	reg.RegisterServiceType("service-factory",
		func(deps, cfg map[string]any) any {
			resolutionOrder = append(resolutionOrder, "service")
			return &struct{ Name string }{Name: "Service"}
		},
		nil,
	)

	// Register as lazy services
	reg.RegisterLazyService("repository", "repository-factory", nil)
	reg.RegisterLazyService("service", "service-factory", map[string]any{
		"depends-on": []string{"repository"},
	})

	// ✅ PROOF 1: After registration, nothing is created
	if len(resolutionOrder) != 0 {
		t.Fatalf("Expected 0 resolutions after registration, got %d: %v", len(resolutionOrder), resolutionOrder)
	}
	t.Logf("✅ PROOF 1: After registration, NO services created")

	// Access only the service (not repository directly)
	svc, ok := reg.GetServiceAny("service")
	if !ok {
		t.Fatal("service not found")
	}

	// ✅ PROOF 2: Both repository AND service were created (dependency chain)
	expectedOrder := []string{"repository", "service"}
	if len(resolutionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d resolutions, got %d: %v", len(expectedOrder), len(resolutionOrder), resolutionOrder)
	}
	for i, expected := range expectedOrder {
		if resolutionOrder[i] != expected {
			t.Fatalf("Expected resolutionOrder[%d] = %s, got %s", i, expected, resolutionOrder[i])
		}
	}
	t.Logf("✅ PROOF 2: Dependency chain resolved on-demand: %v", resolutionOrder)

	// ✅ PROOF 3: Dependencies are resolved BEFORE service factory
	if resolutionOrder[0] != "repository" {
		t.Fatal("Expected repository to be resolved first (dependency)")
	}
	if resolutionOrder[1] != "service" {
		t.Fatal("Expected service to be resolved second (dependent)")
	}
	t.Logf("✅ PROOF 3: Dependencies resolved in correct order (repository → service)")

	// Access service again
	svc2, ok := reg.GetServiceAny("service")
	if !ok {
		t.Fatal("service not found on second access")
	}

	// ✅ PROOF 4: NO additional resolutions (cached)
	if len(resolutionOrder) != 2 {
		t.Fatalf("Expected resolution count to remain 2, got %d: %v", len(resolutionOrder), resolutionOrder)
	}
	if svc != svc2 {
		t.Fatal("Expected same instance on second access")
	}
	t.Logf("✅ PROOF 4: Second access returns cached instance, no re-resolution")

	t.Logf("\n📊 Final Resolution Order: %v", resolutionOrder)
	t.Logf("📊 Total Resolutions: %d (1 dependency + 1 service)", len(resolutionOrder))
}

// Test to PROVE that unused dependencies are NEVER created
func TestLazyLoadingProof_UnusedDependencies(t *testing.T) {
	reg := deploy.NewGlobalRegistry()

	var createdServices []string

	// Service A with no dependencies
	reg.RegisterServiceType("service-a-factory",
		func(deps, cfg map[string]any) any {
			createdServices = append(createdServices, "service-a")
			return &struct{ Name string }{Name: "ServiceA"}
		},
		nil,
	)

	// Service B depends on Service A
	reg.RegisterServiceType("service-b-factory",
		func(deps, cfg map[string]any) any {
			createdServices = append(createdServices, "service-b")
			return &struct{ Name string }{Name: "ServiceB"}
		},
		nil,
	)

	// Service C depends on Service B (transitive dependency on A)
	reg.RegisterServiceType("service-c-factory",
		func(deps, cfg map[string]any) any {
			createdServices = append(createdServices, "service-c")
			return &struct{ Name string }{Name: "ServiceC"}
		},
		nil,
	)

	// Register lazy services
	reg.RegisterLazyService("service-a", "service-a-factory", nil)
	reg.RegisterLazyService("service-b", "service-b-factory", map[string]any{
		"depends-on": []string{"service-a"},
	})
	reg.RegisterLazyService("service-c", "service-c-factory", map[string]any{
		"depends-on": []string{"service-b"},
	})

	// ✅ PROOF 1: Nothing created after registration
	if len(createdServices) != 0 {
		t.Fatalf("Expected 0 services created, got %d: %v", len(createdServices), createdServices)
	}
	t.Logf("✅ PROOF 1: After registration, NO services created")

	// Access only service-a (leaf, no dependencies)
	svcA, ok := reg.GetServiceAny("service-a")
	if !ok {
		t.Fatal("service-a not found")
	}

	// ✅ PROOF 2: Only service-a created, NOT service-b or service-c
	if len(createdServices) != 1 || createdServices[0] != "service-a" {
		t.Fatalf("Expected only service-a to be created, got: %v", createdServices)
	}
	t.Logf("✅ PROOF 2: Only service-a created, service-b and service-c NOT created: %v", createdServices)

	// Now access service-c (depends on b, which depends on a)
	svcC, ok := reg.GetServiceAny("service-c")
	if !ok {
		t.Fatal("service-c not found")
	}

	// ✅ PROOF 3: Dependency chain resolved (b and c created, a reused)
	expectedOrder := []string{"service-a", "service-b", "service-c"}
	if len(createdServices) != len(expectedOrder) {
		t.Fatalf("Expected %d services created, got %d: %v", len(expectedOrder), len(createdServices), createdServices)
	}
	for i, expected := range expectedOrder {
		if createdServices[i] != expected {
			t.Fatalf("Expected createdServices[%d] = %s, got %s", i, expected, createdServices[i])
		}
	}
	t.Logf("✅ PROOF 3: Dependency chain resolved: %v", createdServices)

	t.Logf("\n📊 Final Created Services: %v", createdServices)
	t.Logf("📊 Total Services: 3 (all created because dependency chain)")
	t.Logf("📊 Key Insight: service-a was created ONCE (reused when service-c needed it via service-b)")

	_ = svcA
	_ = svcC
}
