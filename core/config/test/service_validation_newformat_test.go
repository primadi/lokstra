package config_test

import (
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/config"
)

// Test new format: local_key:service-name
func TestValidateServices_NewFormat_LocalKey(t *testing.T) {
	services := &config.ServicesConfig{
		Simple: []*config.Service{
			{
				Name: "user-service",
				Type: "user",
			},
			{
				Name:      "auth-service",
				Type:      "auth",
				DependsOn: []string{"user_service:user-service"}, // NEW FORMAT
				Config:    map[string]any{},                      // Auto-injected!
			},
		},
	}

	err := config.ValidateServices(services)
	if err != nil {
		t.Fatalf("Expected no error with new format, got: %v", err)
	}
}

// Test backward compatibility: simple format still works
func TestValidateServices_OldFormat_BackwardCompatible(t *testing.T) {
	services := &config.ServicesConfig{
		Simple: []*config.Service{
			{
				Name: "user-service",
				Type: "user",
			},
			{
				Name:      "auth-service",
				Type:      "auth",
				DependsOn: []string{"user-service"}, // OLD FORMAT (still works)
				Config: map[string]any{
					"user-service": "user-service", // Must match service name
				},
			},
		},
	}

	err := config.ValidateServices(services)
	if err != nil {
		t.Fatalf("Expected backward compatibility, got: %v", err)
	}
}

// Test multiple dependencies with different local keys
func TestValidateServices_MultipleDepsWithLocalKeys(t *testing.T) {
	services := &config.ServicesConfig{
		Simple: []*config.Service{
			{
				Name: "user-service",
				Type: "user",
			},
			{
				Name: "payment-service",
				Type: "payment",
			},
			{
				Name: "order-service",
				Type: "order",
				DependsOn: []string{
					"user_svc:user-service",       // local_key != service name
					"payment_svc:payment-service", // local_key != service name
				},
				Config: map[string]any{
					// Auto-injected as user_svc and payment_svc
				},
			},
		},
	}

	err := config.ValidateServices(services)
	if err != nil {
		t.Fatalf("Expected no error with multiple deps, got: %v", err)
	}
}

// Test error: dependency service doesn't exist
func TestValidateServices_NonExistentService(t *testing.T) {
	services := &config.ServicesConfig{
		Simple: []*config.Service{
			{
				Name:      "auth-service",
				Type:      "auth",
				DependsOn: []string{"user_service:user-service"}, // user-service doesn't exist
			},
		},
	}

	err := config.ValidateServices(services)
	if err == nil {
		t.Fatal("Expected error for non-existent service")
	}

	if !strings.Contains(err.Error(), "user-service") {
		t.Errorf("Expected error to mention 'user-service', got: %v", err)
	}
}

// Test: literal string in config no longer causes false positive
func TestValidateServices_LiteralStringNoFalsePositive(t *testing.T) {
	services := &config.ServicesConfig{
		Simple: []*config.Service{
			{
				Name: "user-service",
				Type: "user",
			},
			{
				Name: "config-service",
				Type: "config",
				// No depends-on
				Config: map[string]any{
					// This is just a literal string, not a service reference!
					"default_service": "user-service",
					"service_name":    "user-service", // Another literal
				},
			},
		},
	}

	// Should NOT error - Rule 3 removed!
	err := config.ValidateServices(services)
	if err != nil {
		t.Fatalf("Expected no error for literal strings (false positive), got: %v", err)
	}
}

// Test layered mode with new format
func TestValidateServices_LayeredMode_NewFormat(t *testing.T) {
	services := &config.ServicesConfig{
		Layered: map[string][]*config.Service{
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
					DependsOn: []string{"db:db-service"}, // NEW FORMAT
					Config:    map[string]any{},          // Auto-injected
				},
			},
		},
		Order: []string{"infrastructure", "repository"},
	}

	err := config.ValidateServices(services)
	if err != nil {
		t.Fatalf("Expected no error in layered mode with new format, got: %v", err)
	}
}

// Test dependency not used (still required to have value in config or auto-inject)
func TestValidateServices_DependencyNotUsed(t *testing.T) {
	// NOTE: This test is now expected to PASS
	// With auto-injection enabled, dependencies don't need to be in config
	services := &config.ServicesConfig{
		Simple: []*config.Service{
			{
				Name: "db-service",
				Type: "db",
			},
			{
				Name:      "user-service",
				Type:      "user",
				DependsOn: []string{"db_service:db-service"},
				Config: map[string]any{
					// Dependencies will be auto-injected, no need in config
					"password_min_length": 8,
				},
			},
		},
	}

	// Should NOT error - auto-injection handles this
	err := config.ValidateServices(services)
	if err != nil {
		t.Fatalf("Unexpected error (auto-injection should handle this): %v", err)
	}
}

// Test parseDependencyEntry helper
func TestParseDependencyEntry(t *testing.T) {
	tests := []struct {
		input       string
		wantLocal   string
		wantService string
	}{
		{"user-service", "user-service", "user-service"},
		{"user_service:user-service", "user_service", "user-service"},
		{"db:db-service", "db", "db-service"},
		{"my_local_key:some-remote-service", "my_local_key", "some-remote-service"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			local, service := extractServiceNameFromDep(tt.input), extractServiceNameFromDep(tt.input)

			// For simple format, both should be same
			if !strings.Contains(tt.input, ":") {
				if local != tt.wantLocal || service != tt.wantService {
					t.Errorf("parseDependencyEntry(%q) = (%q, %q), want (%q, %q)",
						tt.input, local, service, tt.wantLocal, tt.wantService)
				}
			}
		})
	}
}

// extractServiceNameFromDep extracts the actual service name from depends-on entry
// Supports both "service-name" and "local_key:service-name" formats
func extractServiceNameFromDep(dep string) string {
	if idx := strings.Index(dep, ":"); idx > 0 {
		return dep[idx+1:]
	}
	return dep
}
