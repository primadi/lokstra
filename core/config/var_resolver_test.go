package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==============================================
// ENV RESOLVER TESTS
// ==============================================

func TestEnvResolver_Resolve_ExistingVariable(t *testing.T) {
	resolver := EnvResolver{}
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result, found := resolver.Resolve("ENV", "TEST_VAR", "")
	assert.True(t, found)
	assert.Equal(t, "test_value", result)
}

func TestEnvResolver_Resolve_NonExistentVariable(t *testing.T) {
	resolver := EnvResolver{}

	result, found := resolver.Resolve("ENV", "NON_EXISTENT", "")
	assert.False(t, found)
	assert.Equal(t, "", result)
}

func TestEnvResolver_Resolve_EmptyVariable(t *testing.T) {
	resolver := EnvResolver{}
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	result, found := resolver.Resolve("ENV", "EMPTY_VAR", "default")
	assert.False(t, found)             // Empty env var returns false
	assert.Equal(t, "default", result) // Returns default value
}

func TestEnvResolver_Resolve_WrongSource(t *testing.T) {
	resolver := EnvResolver{}
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result, found := resolver.Resolve("file", "TEST_VAR", "")
	assert.False(t, found)
	assert.Equal(t, "", result)
}

// ==============================================
// VARIABLE EXPANSION TESTS
// ==============================================

func TestExpandVariables_SimpleEnvVar(t *testing.T) {
	os.Setenv("TEST_VAR", "hello")
	defer os.Unsetenv("TEST_VAR")

	result := expandVariables("${TEST_VAR}")
	assert.Equal(t, "hello", result)
}

func TestExpandVariables_WithDefault(t *testing.T) {
	result := expandVariables("${NON_EXISTENT:default_value}")
	assert.Equal(t, "default_value", result)
}

func TestExpandVariables_ExplicitEnvSource(t *testing.T) {
	os.Setenv("TEST_VAR", "env_value")
	defer os.Unsetenv("TEST_VAR")

	result := expandVariables("${ENV:TEST_VAR}")
	assert.Equal(t, "env_value", result)
}

func TestExpandVariables_ExplicitEnvSourceWithDefault(t *testing.T) {
	result := expandVariables("${ENV:NON_EXISTENT:default_value}")
	assert.Equal(t, "default_value", result)
}

func TestExpandVariables_MultipleVariables(t *testing.T) {
	os.Setenv("VAR1", "hello")
	os.Setenv("VAR2", "world")
	defer func() {
		os.Unsetenv("VAR1")
		os.Unsetenv("VAR2")
	}()

	result := expandVariables("${VAR1} ${VAR2}!")
	assert.Equal(t, "hello world!", result)
}

func TestExpandVariables_NoVariables(t *testing.T) {
	result := expandVariables("no variables here")
	assert.Equal(t, "no variables here", result)
}

func TestExpandVariables_EmptyString(t *testing.T) {
	result := expandVariables("")
	assert.Equal(t, "", result)
}

func TestExpandVariables_OnlyDefault(t *testing.T) {
	result := expandVariables("${NON_EXISTENT:}")
	assert.Equal(t, "", result)
}

func TestExpandVariables_ColonInValue(t *testing.T) {
	os.Setenv("TEST_VAR", "value:with:colons")
	defer os.Unsetenv("TEST_VAR")

	result := expandVariables("${TEST_VAR}")
	assert.Equal(t, "value:with:colons", result)
}

// ==============================================
// CUSTOM RESOLVER TESTS
// ==============================================

func TestAddVariableResolver_Success(t *testing.T) {
	// Create custom resolver for testing
	customResolver := &customTestResolver{
		values: map[string]string{
			"test_key": "test_value",
		},
	}

	// Add resolver
	AddVariableResolver("TEST", customResolver)

	// Test resolution
	result := expandVariables("${TEST:test_key}")
	assert.Equal(t, "test_value", result)

	// Clean up by removing the resolver
	delete(variableResolvers, "TEST")
}

func TestAddVariableResolver_Panic(t *testing.T) {
	assert.Panics(t, func() {
		AddVariableResolver("ENV", &customTestResolver{})
	})
}

func TestExpandVariables_WithCustomResolver(t *testing.T) {
	customResolver := &customTestResolver{
		values: map[string]string{
			"config_key": "config_value",
		},
	}

	AddVariableResolver("CONFIG", customResolver)
	defer delete(variableResolvers, "CONFIG")

	result := expandVariables("${CONFIG:config_key:default}")
	assert.Equal(t, "config_value", result)
}

// ==============================================
// EDGE CASES
// ==============================================

func TestExpandVariables_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		setup    func()
		cleanup  func()
	}{
		{
			name:     "Variable with no colon",
			input:    "${NO_COLON}",
			expected: "",
			setup:    func() {},
			cleanup:  func() {},
		},
		{
			name:     "Variable name with underscores and numbers",
			input:    "${VAR_123}",
			expected: "test_value",
			setup: func() {
				os.Setenv("VAR_123", "test_value")
			},
			cleanup: func() {
				os.Unsetenv("VAR_123")
			},
		},
		{
			name:     "Default value with colons - split only on first colon",
			input:    "${NON_EXISTENT:a:b:c}",
			expected: "b:c", // SplitN(key, ":", 3) splits into [NON_EXISTENT, a, b:c], so default = "b:c"
			setup:    func() {},
			cleanup:  func() {},
		},
		{
			name:     "Nested variable syntax - Go's os.Expand limitation",
			input:    "${OUTER_${INNER:inner_value}}",
			expected: "inner_value}", // os.Expand finds ${INNER:inner_value} first, replaces with "inner_value"
			setup:    func() {},
			cleanup:  func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			result := expandVariables(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==============================================
// HELPER STRUCTS FOR TESTING
// ==============================================

type customTestResolver struct {
	values map[string]string
}

func (r *customTestResolver) Resolve(source, name, defaultValue string) (string, bool) {
	if source != "TEST" && source != "CONFIG" {
		return "", false
	}
	value, exists := r.values[name]
	if !exists || value == "" {
		return defaultValue, false
	}
	return value, true
}
