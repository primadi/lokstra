package deploy

import (
	"testing"
)

// Test struct for concrete return type
type TestService struct {
	Name string
}

func NewTestService() *TestService {
	return &TestService{Name: "test"}
}

func NewTestServiceWithConfig(cfg map[string]any) *TestService {
	return &TestService{Name: cfg["name"].(string)}
}

func NewTestServiceFull(deps, cfg map[string]any) *TestService {
	suffix := ""
	if cfg != nil {
		if s, ok := cfg["suffix"]; ok {
			suffix = s.(string)
		}
	}
	return &TestService{Name: "full-" + suffix}
}

// Test concrete return types are accepted
func TestNormalizeServiceFactory_ConcreteReturnTypes(t *testing.T) {
	g := NewGlobalRegistry()

	// Mode 1: func() *TestService (concrete return type!)
	g.RegisterServiceType("test-service-1",
		NewTestService, // func() *TestService
		nil,
	)

	// Mode 2: func(cfg map[string]any) *TestService
	g.RegisterServiceType("test-service-2",
		NewTestServiceWithConfig,
		nil,
	)

	// Mode 3: func(deps, cfg map[string]any) *TestService
	g.RegisterServiceType("test-service-3",
		NewTestServiceFull,
		nil,
	)

	// Test retrieval and invocation
	factory1 := g.GetServiceFactory("test-service-1", true)
	result1 := factory1(nil, nil)
	if svc, ok := result1.(*TestService); !ok || svc.Name != "test" {
		t.Errorf("expected TestService with name 'test', got %v", result1)
	}

	factory2 := g.GetServiceFactory("test-service-2", true)
	result2 := factory2(nil, map[string]any{"name": "custom"})
	if svc, ok := result2.(*TestService); !ok || svc.Name != "custom" {
		t.Errorf("expected TestService with name 'custom', got %v", result2)
	}

	factory3 := g.GetServiceFactory("test-service-3", true)
	result3 := factory3(nil, map[string]any{"suffix": "test"})
	if svc, ok := result3.(*TestService); !ok || svc.Name != "full-test" {
		t.Errorf("expected TestService with name 'full-test', got %v", result3)
	}
}

// Test with RegisterLazyService
func TestRegisterLazyService_ConcreteReturnTypes(t *testing.T) {
	g := NewGlobalRegistry()

	// Register with concrete return types
	g.RegisterLazyService("lazy-test-1", NewTestService, nil)
	g.RegisterLazyService("lazy-test-2", NewTestServiceWithConfig, map[string]any{"name": "lazy"})
	g.RegisterLazyService("lazy-test-3", NewTestServiceFull, map[string]any{"suffix": "lazy"})

	// Test retrieval
	result1, ok := g.GetServiceAny("lazy-test-1")
	if !ok {
		t.Fatal("lazy-test-1 not found")
	}
	if svc, ok := result1.(*TestService); !ok || svc.Name != "test" {
		t.Errorf("expected TestService with name 'test', got %v", result1)
	}

	result2, ok := g.GetServiceAny("lazy-test-2")
	if !ok {
		t.Fatal("lazy-test-2 not found")
	}
	if svc, ok := result2.(*TestService); !ok || svc.Name != "lazy" {
		t.Errorf("expected TestService with name 'lazy', got %v", result2)
	}

	result3, ok := g.GetServiceAny("lazy-test-3")
	if !ok {
		t.Fatal("lazy-test-3 not found")
	}
	if svc, ok := result3.(*TestService); !ok || svc.Name != "full-lazy" {
		t.Errorf("expected TestService with name 'full-lazy', got %v", result3)
	}
}

// Test invalid signatures
func TestNormalizeServiceFactory_InvalidSignatures(t *testing.T) {
	g := NewGlobalRegistry()

	testCases := []struct {
		name    string
		factory any
	}{
		{"not a function", "not a function"},
		{"wrong param type", func(s string) any { return s }},
		{"too many params", func(a, b, c map[string]any) any { return nil }},
		{"no return value", func() {}},
		{"multiple returns", func() (any, error) { return nil, nil }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic for %s", tc.name)
				}
			}()
			g.RegisterServiceType("invalid-"+tc.name, tc.factory, nil)
		})
	}
}
