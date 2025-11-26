package config_test

import (
	"os"
	"testing"

	"github.com/primadi/lokstra/core/config"
)

func TestExpandVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "Simple variable without default",
			input:    "${PORT}",
			envVars:  map[string]string{"PORT": "8080"},
			expected: "8080",
		},
		{
			name:     "Variable with default, env not set",
			input:    "${PORT:3000}",
			envVars:  map[string]string{},
			expected: "3000",
		},
		{
			name:     "Variable with default, env set",
			input:    "${PORT:3000}",
			envVars:  map[string]string{"PORT": "8080"},
			expected: "8080",
		},
		{
			name:     "Default value with colon (URL)",
			input:    "${BASE_URL:http://localhost}",
			envVars:  map[string]string{},
			expected: "http://localhost",
		},
		{
			name:     "Default value with multiple colons (URL with port)",
			input:    "${BASE_URL:http://localhost:8080}",
			envVars:  map[string]string{},
			expected: "http://localhost:8080",
		},
		{
			name:     "Complex default with colon (DSN)",
			input:    "${DSN:postgresql://user:pass@localhost:5432/db}",
			envVars:  map[string]string{},
			expected: "postgresql://user:pass@localhost:5432/db",
		},
		{
			name:     "Explicit ENV resolver",
			input:    "${@ENV:API_KEY}",
			envVars:  map[string]string{"API_KEY": "secret123"},
			expected: "secret123",
		},
		{
			name:     "Explicit ENV resolver with default",
			input:    "${@ENV:API_KEY:default-key}",
			envVars:  map[string]string{},
			expected: "default-key",
		},
		{
			name:     "Explicit ENV resolver with colon in default",
			input:    "${@ENV:SERVICE_URL:http://localhost:9090}",
			envVars:  map[string]string{},
			expected: "http://localhost:9090",
		},
		{
			name:     "Multiple variables in string",
			input:    "Host: ${HOST:localhost}, Port: ${PORT:8080}",
			envVars:  map[string]string{"HOST": "0.0.0.0"},
			expected: "Host: 0.0.0.0, Port: 8080",
		},
		{
			name:     "Variable in middle of string",
			input:    "Server running on http://${HOST:localhost}:${PORT:8080}",
			envVars:  map[string]string{"PORT": "3000"},
			expected: "Server running on http://localhost:3000",
		},
		{
			name:     "Unknown resolver falls back to default",
			input:    "${@UNKNOWN:KEY:fallback}",
			envVars:  map[string]string{},
			expected: "fallback",
		},
		{
			name:     "Empty default value",
			input:    "${VAR:}",
			envVars:  map[string]string{},
			expected: "",
		},
		{
			name:     "No default, env not set",
			input:    "${MISSING_VAR}",
			envVars:  map[string]string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set environment variables
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			result := config.ExpandVariables(tt.input)
			if result != tt.expected {
				t.Errorf("expandVariables(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandVariables_RealWorldExamples(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:  "Database DSN with credentials",
			input: "${DATABASE_URL:postgresql://user:password@localhost:5432/mydb?sslmode=disable}",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://prod_user:prod_pass@db.example.com:5432/production",
			},
			expected: "postgresql://prod_user:prod_pass@db.example.com:5432/production",
		},
		{
			name:     "Redis URL",
			input:    "${REDIS_URL:redis://localhost:6379/0}",
			envVars:  map[string]string{},
			expected: "redis://localhost:6379/0",
		},
		{
			name:     "API endpoint with path",
			input:    "${API_ENDPOINT:https://api.example.com/v1/endpoint}",
			envVars:  map[string]string{},
			expected: "https://api.example.com/v1/endpoint",
		},
		{
			name:  "Complete server config",
			input: "Server: ${SERVER_NAME:my-server} running on ${BASE_URL:http://localhost:8080}",
			envVars: map[string]string{
				"SERVER_NAME": "production-server",
			},
			expected: "Server: production-server running on http://localhost:8080",
		},
		{
			name:     "Unix socket path",
			input:    "${SOCKET_PATH:unix:///var/run/app.sock}",
			envVars:  map[string]string{},
			expected: "unix:///var/run/app.sock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			result := config.ExpandVariables(tt.input)
			if result != tt.expected {
				t.Errorf("expandVariables(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandVariables_CustomResolvers(t *testing.T) {
	// Register a test resolver
	config.AddVariableResolver("TEST", &testResolver{
		values: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Custom resolver without default",
			input:    "${@TEST:key1}",
			expected: "value1",
		},
		{
			name:     "Custom resolver with default, key exists",
			input:    "${@TEST:key2:default}",
			expected: "value2",
		},
		{
			name:     "Custom resolver with default, key missing",
			input:    "${@TEST:missing:fallback-value}",
			expected: "fallback-value",
		},
		{
			name:     "Custom resolver with colon in default",
			input:    "${@TEST:missing:http://localhost:8080}",
			expected: "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.ExpandVariables(tt.input)
			if result != tt.expected {
				t.Errorf("expandVariables(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// testResolver is a mock resolver for testing
type testResolver struct {
	values map[string]string
}

func (r *testResolver) Resolve(source string, key string, defaultValue string) (string, bool) {
	if val, ok := r.values[key]; ok {
		return val, true
	}
	return defaultValue, false
}
