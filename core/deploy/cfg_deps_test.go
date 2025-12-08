package deploy_test

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Test interfaces
type Store interface {
	GetData() string
}

type PostgresStore struct {
	name string
}

func (s *PostgresStore) GetData() string {
	return "postgres:" + s.name
}

type MySQLStore struct {
	name string
}

func (s *MySQLStore) GetData() string {
	return "mysql:" + s.name
}

type UserService struct {
	Store  Store
	Logger string
}

func TestConfigBasedDependencyInjection(t *testing.T) {
	// Clear registry
	_ = deploy.Global()

	// Register store implementations
	lokstra_registry.RegisterLazyService("postgres-store", func() any {
		return &PostgresStore{name: "test-db"}
	}, nil)

	lokstra_registry.RegisterLazyService("mysql-store", func() any {
		return &MySQLStore{name: "test-db"}
	}, nil)

	// Register logger
	lokstra_registry.RegisterService("logger", "test-logger")

	// Register user service with cfg: prefix dependency
	lokstra_registry.RegisterLazyService("user-service",
		func(deps map[string]any, config map[string]any) any {
			return &UserService{
				Store:  deps["cfg:store.implementation"].(Store),
				Logger: deps["logger"].(string),
			}
		},
		map[string]any{
			"depends-on":           []string{"cfg:store.implementation", "logger"},
			"store.implementation": "postgres-store", // Config specifies postgres
		},
	)

	// Get service - should resolve postgres-store
	userSvc := lokstra_registry.MustGetService[*UserService]("user-service")

	if userSvc == nil {
		t.Fatal("expected user service to be created")
	}

	if userSvc.Logger != "test-logger" {
		t.Errorf("expected logger 'test-logger', got '%s'", userSvc.Logger)
	}

	data := userSvc.Store.GetData()
	if data != "postgres:test-db" {
		t.Errorf("expected 'postgres:test-db', got '%s'", data)
	}
}

func TestConfigBasedDependencyInjection_SwitchImplementation(t *testing.T) {
	// Clear registry
	_ = deploy.Global()

	// Register store implementations
	lokstra_registry.RegisterLazyService("postgres-store-2", func() any {
		return &PostgresStore{name: "pg-db"}
	}, nil)

	lokstra_registry.RegisterLazyService("mysql-store-2", func() any {
		return &MySQLStore{name: "my-db"}
	}, nil)

	// Register logger
	lokstra_registry.RegisterService("logger-2", "logger2")

	// Register user service with MySQL this time
	lokstra_registry.RegisterLazyService("user-service-2",
		func(deps map[string]any, config map[string]any) any {
			return &UserService{
				Store:  deps["cfg:store.implementation"].(Store),
				Logger: deps["logger-2"].(string),
			}
		},
		map[string]any{
			"depends-on":           []string{"cfg:store.implementation", "logger-2"},
			"store.implementation": "mysql-store-2", // Switch to MySQL!
		},
	)

	// Get service - should resolve mysql-store now
	userSvc := lokstra_registry.MustGetService[*UserService]("user-service-2")

	if userSvc == nil {
		t.Fatal("expected user service to be created")
	}

	data := userSvc.Store.GetData()
	if data != "mysql:my-db" {
		t.Errorf("expected 'mysql:my-db', got '%s'", data)
	}
}

func TestConfigBasedDependency_MissingConfig(t *testing.T) {
	// Clear registry
	_ = deploy.Global()

	// Register store
	lokstra_registry.RegisterLazyService("some-store", func() any {
		return &PostgresStore{name: "test"}
	}, nil)

	// Register service WITHOUT config for cfg: dependency
	lokstra_registry.RegisterLazyService("bad-service",
		func(deps map[string]any, config map[string]any) any {
			return &UserService{
				Store: deps["cfg:missing.config"].(Store),
			}
		},
		map[string]any{
			"depends-on": []string{"cfg:missing.config"},
			// Missing: "missing.config" key!
		},
	)

	// Should panic when trying to resolve
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when config key is missing")
		} else {
			// Check error message mentions config requirement
			if msg, ok := r.(string); ok {
				if !contains(msg, "config-based dependency") {
					t.Errorf("expected error about config-based dependency, got: %v", r)
				}
			}
		}
	}()

	lokstra_registry.MustGetService[*UserService]("bad-service")
}

func TestConfigBasedDependency_EmptyConfig(t *testing.T) {
	// Clear registry
	_ = deploy.Global()

	// Register service with EMPTY config value
	lokstra_registry.RegisterLazyService("bad-service-2",
		func(deps map[string]any, config map[string]any) any {
			return &UserService{
				Store: deps["cfg:empty.config"].(Store),
			}
		},
		map[string]any{
			"depends-on":   []string{"cfg:empty.config"},
			"empty.config": "", // Empty string!
		},
	)

	// Should panic when config value is empty
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when config value is empty")
		}
	}()

	lokstra_registry.MustGetService[*UserService]("bad-service-2")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
