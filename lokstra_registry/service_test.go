package lokstra_registry_test

import (
	"testing"

	"github.com/primadi/lokstra/lokstra_registry"
)

type namedService interface {
	Name() string
}

type descService interface {
	Desc() string
}

type serviceA struct {
	name string
}

func (s *serviceA) Name() string {
	return s.name
}

func newServiceA(name string) namedService {
	return &serviceA{name: name}
}

func TestRegisterAndGetService(t *testing.T) {
	t.Run("register and get by concrete type", func(t *testing.T) {
		lokstra_registry.RegisterService("serviceA", newServiceA("Service A"),
			lokstra_registry.AllowOverride(true))
		svcA := lokstra_registry.GetService[*serviceA]("serviceA")
		if svcA == nil {
			t.Fatalf("svcA should not be nil after LazyGetService")
		}
		if svcA.Name() != "Service A" {
			t.Errorf("Name of svcA should be 'Service A', got %s", svcA.Name())
		}
	})

	t.Run("get by interface type", func(t *testing.T) {
		namedSvc := lokstra_registry.GetService[namedService]("serviceA")
		if namedSvc == nil {
			t.Fatalf("namedSvc should not be nil after LazyGetService")
		}
		if namedSvc.Name() != "Service A" {
			t.Errorf("Name of namedSvc should be 'Service A', got %s", namedSvc.Name())
		}
	})

	t.Run("try get with wrong interface", func(t *testing.T) {
		_, ok := lokstra_registry.TryGetService[descService]("serviceA")
		if ok {
			t.Errorf("serviceA should not be found with DescService interface")
		}
		_, ok = lokstra_registry.TryGetService[descService]("serviceC")
		if ok {
			t.Errorf("serviceC should not be found with DescService interface")
		}
	})
}

func TestNewService(t *testing.T) {
	t.Run("register factory and create service", func(t *testing.T) {
		lokstra_registry.RegisterServiceFactory("typeA", func(config map[string]any) any {
			return &serviceA{name: config["name"].(string)}
		}, lokstra_registry.AllowOverride(true))
		svcA := lokstra_registry.NewService[namedService]("newServiceA", "typeA", map[string]any{"name": "Service A"})
		if svcA == nil {
			t.Fatalf("Failed to create serviceA")
		}
		var newServiceA *serviceA
		var ok bool
		if newServiceA, ok = lokstra_registry.TryGetService[*serviceA]("newServiceA"); !ok {
			t.Errorf("newServiceA should be registered")
		}
		if newServiceA.Name() != "Service A" {
			t.Errorf("Name of newServiceA should be 'Service A', got %s", newServiceA.Name())
		}
	})

	t.Run("create service with unregistered factory", func(t *testing.T) {
		svcB := lokstra_registry.NewService[*serviceA]("newServiceB", "typeB", map[string]any{"name": "Service B"})
		if svcB != nil {
			t.Errorf("Failed to create serviceB")
		}
	})
}

func TestGetService_PanicNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when service not found")
		}
	}()
	_ = lokstra_registry.MustGetService[*serviceA]("notExistService")
}

func TestGetService_PanicTypeMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when type mismatch")
		}
	}()
	// Register serviceA as *serviceA
	lokstra_registry.RegisterService("serviceA", newServiceA("Service A"))
	// Try to get as descService (should panic)
	_ = lokstra_registry.GetService[descService]("serviceA")
}

func TestRegisterLazyServiceAndGetService(t *testing.T) {
	// Register factory for lazy type
	lokstra_registry.RegisterServiceFactory("lazyTypeA", func(config map[string]any) any {
		return &serviceA{name: config["name"].(string)}
	})

	// Register lazy service
	lokstra_registry.RegisterLazyService("lazyServiceA", "lazyTypeA",
		map[string]any{"name": "Lazy Service A"})

	t.Run("GetService from lazy service", func(t *testing.T) {
		svcA := lokstra_registry.GetService[*serviceA]("lazyServiceA")
		if svcA == nil {
			t.Fatalf("svcA should not be nil after GetService from lazy service")
		}
		if svcA.Name() != "Lazy Service A" {
			t.Errorf("Name of svcA should be 'Lazy Service A', got %s", svcA.Name())
		}
	})

	t.Run("GetService with wrong type should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic when type mismatch from lazy service")
			}
		}()
		_ = lokstra_registry.MustGetService[descService]("lazyServiceA")
	})
}
