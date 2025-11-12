package deploy

import (
	"reflect"
	"testing"

	"github.com/primadi/lokstra/core/request"
)

func TestParseMiddlewareName(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedName   string
		expectedParams map[string]any
	}{
		{
			name:           "No parameters",
			input:          "recovery",
			expectedName:   "recovery",
			expectedParams: nil,
		},
		{
			name:           "Single parameter with quotes",
			input:          `rate-limit max="100"`,
			expectedName:   "rate-limit",
			expectedParams: map[string]any{"max": "100"},
		},
		{
			name:           "Single parameter without quotes",
			input:          `rate-limit max=100`,
			expectedName:   "rate-limit",
			expectedParams: map[string]any{"max": "100"},
		},
		{
			name:           "Multiple parameters with quotes",
			input:          `cors origins="https://example.com", methods="GET,POST"`,
			expectedName:   "cors",
			expectedParams: map[string]any{"origins": "https://example.com", "methods": "GET,POST"},
		},
		{
			name:           "Multiple parameters without quotes",
			input:          `rate-limit max=100, window=1m`,
			expectedName:   "rate-limit",
			expectedParams: map[string]any{"max": "100", "window": "1m"},
		},
		{
			name:           "Mixed quoted and unquoted",
			input:          `auth max=100, secret="my-secret!", issuer=local`,
			expectedName:   "auth",
			expectedParams: map[string]any{"max": "100", "secret": "my-secret!", "issuer": "local"},
		},
		{
			name:           "Parameters with spaces in quoted values",
			input:          `logger prefix="API Server", level=debug`,
			expectedName:   "logger",
			expectedParams: map[string]any{"prefix": "API Server", "level": "debug"},
		},
		{
			name:           "Parameters with special characters in quotes",
			input:          `auth secret="my-s3cr3t!", issuer="https://auth.example.com"`,
			expectedName:   "auth",
			expectedParams: map[string]any{"secret": "my-s3cr3t!", "issuer": "https://auth.example.com"},
		},
		{
			name:           "Extra whitespace",
			input:          `  timeout   duration=30s  ,  message="Request timeout"  `,
			expectedName:   "timeout",
			expectedParams: map[string]any{"duration": "30s", "message": "Request timeout"},
		},
		{
			name:           "Numeric values without quotes",
			input:          `cache ttl=3600, max_size=1000`,
			expectedName:   "cache",
			expectedParams: map[string]any{"ttl": "3600", "max_size": "1000"},
		},
		{
			name:           "Empty value with quotes",
			input:          `test-mw param=""`,
			expectedName:   "test-mw",
			expectedParams: map[string]any{"param": ""},
		},
		{
			name:           "URL without quotes (simple)",
			input:          `proxy target=http://localhost:3000`,
			expectedName:   "proxy",
			expectedParams: map[string]any{"target": "http://localhost:3000"},
		},
		{
			name:           "Boolean-like values without quotes",
			input:          `feature enabled=true, debug=false`,
			expectedName:   "feature",
			expectedParams: map[string]any{"enabled": "true", "debug": "false"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, params := parseMiddlewareName(tt.input)

			if name != tt.expectedName {
				t.Errorf("parseMiddlewareName(%q) name = %q, want %q", tt.input, name, tt.expectedName)
			}

			if !reflect.DeepEqual(params, tt.expectedParams) {
				t.Errorf("parseMiddlewareName(%q) params = %v, want %v", tt.input, params, tt.expectedParams)
			}
		})
	}
}

func TestMergeConfig(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]any
		override map[string]any
		expected map[string]any
	}{
		{
			name:     "Both nil",
			base:     nil,
			override: nil,
			expected: nil,
		},
		{
			name:     "Base nil",
			base:     nil,
			override: map[string]any{"key": "value"},
			expected: map[string]any{"key": "value"},
		},
		{
			name:     "Override nil",
			base:     map[string]any{"key": "value"},
			override: nil,
			expected: map[string]any{"key": "value"},
		},
		{
			name:     "Merge without conflicts",
			base:     map[string]any{"key1": "value1"},
			override: map[string]any{"key2": "value2"},
			expected: map[string]any{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "Override takes precedence",
			base:     map[string]any{"key": "base_value"},
			override: map[string]any{"key": "override_value"},
			expected: map[string]any{"key": "override_value"},
		},
		{
			name:     "Complex merge",
			base:     map[string]any{"key1": "value1", "key2": "value2"},
			override: map[string]any{"key2": "override2", "key3": "value3"},
			expected: map[string]any{"key1": "value1", "key2": "override2", "key3": "value3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeConfig(tt.base, tt.override)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("mergeConfig(%v, %v) = %v, want %v", tt.base, tt.override, result, tt.expected)
			}
		})
	}
}

func TestCreateMiddlewareWithInlineParams(t *testing.T) {
	registry := NewGlobalRegistry()

	// Track received config
	var receivedConfig map[string]any

	// Register a test middleware factory
	registry.RegisterMiddlewareType("test-mw", func(config map[string]any) any {
		receivedConfig = config
		return request.HandlerFunc(func(ctx *request.Context) error {
			return ctx.Next()
		})
	})

	tests := []struct {
		name           string
		middlewareName string
		expectedConfig map[string]any
	}{
		{
			name:           "No parameters",
			middlewareName: "test-mw",
			expectedConfig: nil,
		},
		{
			name:           "Single parameter",
			middlewareName: `test-mw param1="value1"`,
			expectedConfig: map[string]any{"param1": "value1"},
		},
		{
			name:           "Multiple parameters",
			middlewareName: `test-mw param1="value1", param2="value2"`,
			expectedConfig: map[string]any{"param1": "value1", "param2": "value2"},
		},
		{
			name:           "Parameters with special characters",
			middlewareName: `test-mw url="https://example.com", secret="my-s3cr3t!"`,
			expectedConfig: map[string]any{"url": "https://example.com", "secret": "my-s3cr3t!"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receivedConfig = nil // Reset

			mw := registry.CreateMiddleware(tt.middlewareName)

			if mw == nil {
				t.Fatal("CreateMiddleware returned nil")
			}

			// Verify received config
			if !reflect.DeepEqual(receivedConfig, tt.expectedConfig) {
				t.Errorf("Factory received config = %v, want %v", receivedConfig, tt.expectedConfig)
			}
		})
	}
}

func TestCreateMiddlewareWithRegisteredName(t *testing.T) {
	registry := NewGlobalRegistry()

	// Track received config
	var receivedConfig map[string]any

	// Register factory
	registry.RegisterMiddlewareType("logger", func(config map[string]any) any {
		receivedConfig = config
		return request.HandlerFunc(func(ctx *request.Context) error {
			return ctx.Next()
		})
	})

	// Register named middleware with base config
	registry.RegisterMiddlewareName("api-logger", "logger", map[string]any{
		"prefix": "API",
	})

	tests := []struct {
		name           string
		middlewareName string
		expectedConfig map[string]any
	}{
		{
			name:           "Use registered name",
			middlewareName: "api-logger",
			expectedConfig: map[string]any{"prefix": "API"},
		},
		{
			name:           "Override with inline params",
			middlewareName: `api-logger prefix="CUSTOM"`,
			expectedConfig: map[string]any{"prefix": "CUSTOM"}, // Inline takes precedence
		},
		{
			name:           "Merge inline params with base config",
			middlewareName: `api-logger level="debug"`,
			expectedConfig: map[string]any{"prefix": "API", "level": "debug"}, // Merged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receivedConfig = nil // Reset

			mw := registry.CreateMiddleware(tt.middlewareName)

			if mw == nil {
				t.Fatal("CreateMiddleware returned nil")
			}

			// Verify received config
			if !reflect.DeepEqual(receivedConfig, tt.expectedConfig) {
				t.Errorf("Factory received config = %v, want %v", receivedConfig, tt.expectedConfig)
			}
		})
	}
}
