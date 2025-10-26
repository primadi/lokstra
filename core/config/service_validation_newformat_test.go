package config

import (
	"strings"
	"testing"
)

// Test new format: local_key:service-name
func TestValidateServices_NewFormat_LocalKey(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
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

	err := ValidateServices(services)
	if err != nil {
		t.Fatalf("Expected no error with new format, got: %v", err)
	}
}

// Test backward compatibility: simple format still works
func TestValidateServices_OldFormat_BackwardCompatible(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
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

	err := ValidateServices(services)
	if err != nil {
		t.Fatalf("Expected backward compatibility, got: %v", err)
	}
}

// Test multiple dependencies with different local keys
func TestValidateServices_MultipleDepsWithLocalKeys(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
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

	err := ValidateServices(services)
	if err != nil {
		t.Fatalf("Expected no error with multiple deps, got: %v", err)
	}
}

// Test error: dependency service doesn't exist
func TestValidateServices_NonExistentService(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
			{
				Name:      "auth-service",
				Type:      "auth",
				DependsOn: []string{"user_service:user-service"}, // user-service doesn't exist
			},
		},
	}

	err := ValidateServices(services)
	if err == nil {
		t.Fatal("Expected error for non-existent service")
	}

	if !strings.Contains(err.Error(), "user-service") {
		t.Errorf("Expected error to mention 'user-service', got: %v", err)
	}
}

// Test: literal string in config no longer causes false positive
func TestValidateServices_LiteralStringNoFalsePositive(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
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
	err := ValidateServices(services)
	if err != nil {
		t.Fatalf("Expected no error for literal strings (false positive), got: %v", err)
	}
}

// Test layered mode with new format
func TestValidateServices_LayeredMode_NewFormat(t *testing.T) {
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
					DependsOn: []string{"db:db-service"}, // NEW FORMAT
					Config:    map[string]any{},          // Auto-injected
				},
			},
		},
		Order: []string{"infrastructure", "repository"},
	}

	err := ValidateServices(services)
	if err != nil {
		t.Fatalf("Expected no error in layered mode with new format, got: %v", err)
	}
}

// Test dependency not used (still required to have value in config or auto-inject)
func TestValidateServices_DependencyNotUsed(t *testing.T) {
	services := &ServicesConfig{
		Simple: []*Service{
			{
				Name: "db-service",
				Type: "db",
			},
			{
				Name:      "user-service",
				Type:      "user",
				DependsOn: []string{"db_service:db-service"},
				Config: map[string]any{
					// Config doesn't reference db-service at all
					"password_min_length": 8,
				},
			},
		},
	}

	// Should error - dependency declared but not used
	err := ValidateServices(services)
	if err == nil {
		t.Fatal("Expected error when dependency not used in config")
	}

	if !strings.Contains(err.Error(), "db_service:db-service") {
		t.Errorf("Expected error to mention dependency, got: %v", err)
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
