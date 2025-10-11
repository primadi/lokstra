package config

import (
	"strings"
	"testing"
)

func TestValidateServices_SimpleMode_Success(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
			{
				Name: "db-service",
				Type: "db",
				Config: map[string]any{
					"host": "localhost",
				},
			},
			{
				Name:      "user-service",
				Type:      "user",
				DependsOn: []string{"db-service"},
				Config: map[string]any{
					"db_service": "db-service",
				},
			},
		},
	}

	err := ValidateServices(services)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateServices_SimpleMode_DependencyNotInConfig(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
			{
				Name: "db-service",
				Type: "db",
			},
			{
				Name:      "user-service",
				Type:      "user",
				DependsOn: []string{"db-service"}, // ← In depends-on
				Config: map[string]any{
					// ❌ NOT in config - should error!
					"password_min_length": 8,
				},
			},
		},
	}

	err := ValidateServices(services)
	if err == nil {
		t.Fatal("Expected error for dependency in depends-on but not used in config")
	}

	expectedMsg := "dependency 'db-service' in depends-on but not used in config"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestValidateServices_SimpleMode_DependencyNotExists(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
			{
				Name:      "user-service",
				Type:      "user",
				DependsOn: []string{"db-service"}, // ← Depends on non-existent service
				Config: map[string]any{
					"db_service": "db-service",
				},
			},
		},
	}

	err := ValidateServices(services)
	if err == nil {
		t.Fatal("Expected error for non-existent dependency")
	}

	expectedMsg := "depends on 'db-service' which does not exist"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestValidateServices_SimpleMode_ConfigReferenceNotInDependsOn(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
			{
				Name: "db-service",
				Type: "db",
			},
			{
				Name:      "user-service",
				Type:      "user",
				DependsOn: []string{}, // ← Empty depends-on
				Config: map[string]any{
					"db_service": "db-service", // ❌ References service but not in depends-on
				},
			},
		},
	}

	err := ValidateServices(services)
	if err == nil {
		t.Fatal("Expected error for config reference not in depends-on")
	}

	expectedMsg := "references service 'db-service' which is not in depends-on"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestValidateServices_SimpleMode_MultipleDependencies(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
			{
				Name: "db-service",
				Type: "db",
			},
			{
				Name: "cache-service",
				Type: "cache",
			},
			{
				Name:      "user-service",
				Type:      "user",
				DependsOn: []string{"db-service", "cache-service"},
				Config: map[string]any{
					"db_service":    "db-service",
					"cache_service": "cache-service",
				},
			},
		},
	}

	err := ValidateServices(services)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateServices_LayeredMode_Success(t *testing.T) {
	services := &ServicesConfig{
		Layered: map[string][]*Service{
			"infrastructure": {
				{
					Name: "db-service",
					Type: "db",
				},
			},
			"repository": {
				{
					Name:      "user-repo",
					Type:      "user_repo",
					DependsOn: []string{"db-service"},
					Config: map[string]any{
						"db_service": "db-service",
					},
				},
			},
		},
		Order: []string{"infrastructure", "repository"},
	}

	err := ValidateServices(services)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateServices_LayeredMode_DependencyNotInConfig(t *testing.T) {
	services := &ServicesConfig{
		Layered: map[string][]*Service{
			"infrastructure": {
				{
					Name: "db-service",
					Type: "db",
				},
			},
			"repository": {
				{
					Name:      "user-repo",
					Type:      "user_repo",
					DependsOn: []string{"db-service"}, // ← In depends-on
					Config: map[string]any{
						// ❌ NOT in config
						"table": "users",
					},
				},
			},
		},
		Order: []string{"infrastructure", "repository"},
	}

	err := ValidateServices(services)
	if err == nil {
		t.Fatal("Expected error for dependency in depends-on but not used in config")
	}

	expectedMsg := "dependency 'db-service' in depends-on but not used in config"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestValidateServices_LayeredMode_DependencyNotExists(t *testing.T) {
	services := &ServicesConfig{
		Layered: map[string][]*Service{
			"repository": {
				{
					Name:      "user-repo",
					Type:      "user_repo",
					DependsOn: []string{"db-service"}, // ← Non-existent
					Config: map[string]any{
						"db_service": "db-service",
					},
				},
			},
		},
		Order: []string{"repository"},
	}

	err := ValidateServices(services)
	if err == nil {
		t.Fatal("Expected error for non-existent dependency")
	}

	expectedMsg := "depends on 'db-service' which is not available yet"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestValidateServices_LayeredMode_ConfigReferenceNotInDependsOn(t *testing.T) {
	services := &ServicesConfig{
		Layered: map[string][]*Service{
			"infrastructure": {
				{
					Name: "db-service",
					Type: "db",
				},
			},
			"repository": {
				{
					Name:      "user-repo",
					Type:      "user_repo",
					DependsOn: []string{}, // ← Empty
					Config: map[string]any{
						"db_service": "db-service", // ❌ Not in depends-on
					},
				},
			},
		},
		Order: []string{"infrastructure", "repository"},
	}

	err := ValidateServices(services)
	if err == nil {
		t.Fatal("Expected error for config reference not in depends-on")
	}

	expectedMsg := "references service 'db-service' which is not in depends-on"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestValidateServices_EmptyConfig(t *testing.T) {
	services := &ServicesConfig{}

	err := ValidateServices(services)
	if err != nil {
		t.Errorf("Expected no error for empty config, got: %v", err)
	}
}

func TestValidateServices_NoDependencies(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
			{
				Name: "standalone-service",
				Type: "standalone",
				Config: map[string]any{
					"port": 8080,
				},
			},
		},
	}

	err := ValidateServices(services)
	if err != nil {
		t.Errorf("Expected no error for service with no dependencies, got: %v", err)
	}
}

// Test case from auth_service example in the codebase
func TestValidateServices_AuthServiceExample(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
			{
				Name: "user-service",
				Type: "user_service",
				Config: map[string]any{
					"storage": "memory",
				},
			},
			{
				Name:      "auth-service",
				Type:      "auth_service",
				DependsOn: []string{"user-service"},
				Config: map[string]any{
					"user_service": "user-service", // ← Must be in config if in depends-on
					"jwt_secret":   "dev-secret-key",
					"token_expiry": 3600,
				},
			},
		},
	}

	err := ValidateServices(services)
	if err != nil {
		t.Errorf("Expected no error for auth service example, got: %v", err)
	}
}

// Test case: depends-on without usage should fail
func TestValidateServices_AuthServiceExample_MissingInConfig(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
			{
				Name: "user-service",
				Type: "user_service",
			},
			{
				Name:      "auth-service",
				Type:      "auth_service",
				DependsOn: []string{"user-service"}, // ← In depends-on
				Config: map[string]any{
					// ❌ NOT in config!
					"jwt_secret":   "dev-secret-key",
					"token_expiry": 3600,
				},
			},
		},
	}

	err := ValidateServices(services)
	if err == nil {
		t.Fatal("Expected error when dependency in depends-on is not used in config")
	}

	expectedMsg := "dependency 'user-service' in depends-on but not used in config"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}
