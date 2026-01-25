package deploy_test

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Test interfaces
type Repository interface {
	GetData() string
}

type PostgresRepository struct {
	name string
}

func (s *PostgresRepository) GetData() string {
	return "postgres:" + s.name
}

type MySQLRepository struct {
	name string
}

func (s *MySQLRepository) GetData() string {
	return "mysql:" + s.name
}

type UserService struct {
	Repository Repository
	Logger     string
}

func TestConfigBasedDependencyInjection(t *testing.T) {
	// Clear registry for isolated test
	deploy.ResetGlobalRegistryForTesting()

	// Register repository implementations
	lokstra_registry.RegisterLazyService("postgres-repository", func() any {
		return &PostgresRepository{name: "test-db"}
	}, nil)

	lokstra_registry.RegisterLazyService("mysql-repository", func() any {
		return &MySQLRepository{name: "test-db"}
	}, nil)

	// Register logger
	lokstra_registry.RegisterService("logger", "test-logger")

	// Set GLOBAL config (not service config!)
	lokstra_registry.SetConfig("repository.implementation", "postgres-repository")

	// DEBUG: Verify config is set
	if val, ok := deploy.Global().GetConfig("repository.implementation"); ok {
		t.Logf("✅ Config set successfully: repository.implementation = %v", val)
	} else {
		t.Fatalf("❌ Config NOT set: repository.implementation")
	}

	// Register user service with @ prefix for config-based dependency
	lokstra_registry.RegisterLazyService("user-service",
		func(deps map[string]any, config map[string]any) any {
			return &UserService{
				Repository: deps["cfg"].(Repository),
				Logger:     deps["logger"].(string),
			}
		},
		map[string]any{
			"depends-on": []string{"cfg:@repository.implementation", "logger"},
		},
	)

	// Get service - should resolve postgres-repository
	userSvc := lokstra_registry.MustGetService[*UserService]("user-service")

	if userSvc == nil {
		t.Fatal("expected user service to be created")
	}

	if userSvc.Logger != "test-logger" {
		t.Errorf("expected logger 'test-logger', got '%s'", userSvc.Logger)
	}

	data := userSvc.Repository.GetData()
	if data != "postgres:test-db" {
		t.Errorf("expected 'postgres:test-db', got '%s'", data)
	}
}

func TestConfigBasedDependencyInjection_SwitchImplementation(t *testing.T) {
	// Clear registry for isolated test
	deploy.ResetGlobalRegistryForTesting()

	// Register repository implementations
	lokstra_registry.RegisterLazyService("postgres-repository-2", func() any {
		return &PostgresRepository{name: "pg-db"}
	}, nil)

	lokstra_registry.RegisterLazyService("mysql-repository-2", func() any {
		return &MySQLRepository{name: "my-db"}
	}, nil)

	// Register logger
	lokstra_registry.RegisterService("logger-2", "logger2")

	// Set GLOBAL config to use MySQL
	lokstra_registry.SetConfig("repository.implementation", "mysql-repository-2")

	// Register user service with MySQL this time
	lokstra_registry.RegisterLazyService("user-service-2",
		func(deps map[string]any, config map[string]any) any {
			return &UserService{
				Repository: deps["cfg"].(Repository),
				Logger:     deps["logger-2"].(string),
			}
		},
		map[string]any{
			"depends-on": []string{"cfg:@repository.implementation", "logger-2"},
		},
	)

	// Get service - should resolve mysql-repository now
	userSvc := lokstra_registry.MustGetService[*UserService]("user-service-2")

	if userSvc == nil {
		t.Fatal("expected user service to be created")
	}

	data := userSvc.Repository.GetData()
	if data != "mysql:my-db" {
		t.Errorf("expected 'mysql:my-db', got '%s'", data)
	}
}

func TestConfigBasedDependency_MissingConfig(t *testing.T) {
	// Clear registry for isolated test
	deploy.ResetGlobalRegistryForTesting()

	// Register repository
	lokstra_registry.RegisterLazyService("some-repository", func() any {
		return &PostgresRepository{name: "test"}
	}, nil)

	// DO NOT set global config for "missing.config"

	// Register service with @ dependency that won't be found
	lokstra_registry.RegisterLazyService("bad-service",
		func(deps map[string]any, config map[string]any) any {
			return &UserService{
				Repository: deps["cfg"].(Repository),
			}
		},
		map[string]any{
			"depends-on": []string{"cfg:@missing.config"},
		},
	)

	// Should panic when trying to resolve - config key not found
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when config key is missing")
		}
	}()

	lokstra_registry.MustGetService[*UserService]("bad-service")
}

func TestConfigBasedDependency_EmptyConfig(t *testing.T) {
	// Clear registry for isolated test
	deploy.ResetGlobalRegistryForTesting()

	// Set global config with EMPTY value
	lokstra_registry.SetConfig("empty.config", "")

	// Register service with @ dependency pointing to empty config
	lokstra_registry.RegisterLazyService("bad-service-2",
		func(deps map[string]any, config map[string]any) any {
			return &UserService{
				Repository: deps["cfg"].(Repository),
			}
		},
		map[string]any{
			"depends-on": []string{"cfg:@empty.config"},
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
