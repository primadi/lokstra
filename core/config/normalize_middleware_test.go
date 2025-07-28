package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================
// NORMALIZE MIDDLEWARE TESTS
// ==============================================

func TestNormalizeMiddlewareConfig_StringArray(t *testing.T) {
	input := []any{"cors", "auth", "logging"}

	result, err := normalizeMiddlewareConfig(input)
	require.NoError(t, err)
	require.Len(t, result, 3)

	assert.Equal(t, "cors", result[0].Name)
	assert.True(t, result[0].Enabled)
	assert.Nil(t, result[0].Config)

	assert.Equal(t, "auth", result[1].Name)
	assert.True(t, result[1].Enabled)
	assert.Nil(t, result[1].Config)

	assert.Equal(t, "logging", result[2].Name)
	assert.True(t, result[2].Enabled)
	assert.Nil(t, result[2].Config)
}

func TestNormalizeMiddlewareConfig_MapArray(t *testing.T) {
	input := []any{
		map[string]any{
			"name":    "cors",
			"enabled": true,
		},
		map[string]any{
			"name":    "auth",
			"enabled": false,
			"config": map[string]any{
				"secret": "mysecret",
				"ttl":    3600,
			},
		},
	}

	result, err := normalizeMiddlewareConfig(input)
	require.NoError(t, err)
	require.Len(t, result, 2)

	// First middleware
	assert.Equal(t, "cors", result[0].Name)
	assert.True(t, result[0].Enabled)
	assert.Nil(t, result[0].Config)

	// Second middleware
	assert.Equal(t, "auth", result[1].Name)
	assert.False(t, result[1].Enabled)
	require.NotNil(t, result[1].Config)
	assert.Equal(t, "mysecret", result[1].Config["secret"])
	assert.Equal(t, 3600, result[1].Config["ttl"])
}

func TestNormalizeMiddlewareConfig_MixedArray(t *testing.T) {
	input := []any{
		"cors",
		map[string]any{
			"name":   "auth",
			"config": map[string]any{"secret": "test"},
		},
		"logging",
	}

	result, err := normalizeMiddlewareConfig(input)
	require.NoError(t, err)
	require.Len(t, result, 3)

	assert.Equal(t, "cors", result[0].Name)
	assert.True(t, result[0].Enabled)

	assert.Equal(t, "auth", result[1].Name)
	assert.True(t, result[1].Enabled) // default to true
	assert.Equal(t, "test", result[1].Config["secret"])

	assert.Equal(t, "logging", result[2].Name)
	assert.True(t, result[2].Enabled)
}

func TestNormalizeMiddlewareConfig_MapWithoutName(t *testing.T) {
	input := []any{
		map[string]any{
			"enabled": true,
			"config":  map[string]any{"key": "value"},
		},
	}

	result, err := normalizeMiddlewareConfig(input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "middleware entry missing 'name'")
}

func TestNormalizeMiddlewareConfig_MapWithEmptyName(t *testing.T) {
	input := []any{
		map[string]any{
			"name":    "",
			"enabled": true,
		},
	}

	result, err := normalizeMiddlewareConfig(input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "middleware entry missing 'name'")
}

func TestNormalizeMiddlewareConfig_MapWithNonStringName(t *testing.T) {
	input := []any{
		map[string]any{
			"name":    123, // Non-string name
			"enabled": true,
		},
	}

	result, err := normalizeMiddlewareConfig(input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "middleware entry missing 'name'")
}

func TestNormalizeMiddlewareConfig_InvalidArrayItemType(t *testing.T) {
	input := []any{
		"cors",
		123, // Invalid type
		"auth",
	}

	result, err := normalizeMiddlewareConfig(input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid middleware entry type")
}

func TestNormalizeMiddlewareConfig_NotAnArray(t *testing.T) {
	input := "not an array"

	result, err := normalizeMiddlewareConfig(input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "middleware config must be a list")
}

func TestNormalizeMiddlewareConfig_EmptyArray(t *testing.T) {
	input := []any{}

	result, err := normalizeMiddlewareConfig(input)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestNormalizeMiddlewareConfig_MapWithDefaultEnabled(t *testing.T) {
	// Test that enabled defaults to true when not specified
	input := []any{
		map[string]any{
			"name": "test-middleware",
			// enabled not specified
		},
	}

	result, err := normalizeMiddlewareConfig(input)
	require.NoError(t, err)
	require.Len(t, result, 1)

	assert.Equal(t, "test-middleware", result[0].Name)
	assert.True(t, result[0].Enabled) // Should default to true
	assert.Nil(t, result[0].Config)
}

func TestNormalizeMiddlewareConfig_MapWithNonBoolEnabled(t *testing.T) {
	// Test behavior when enabled is not a boolean
	input := []any{
		map[string]any{
			"name":    "test-middleware",
			"enabled": "true", // String instead of bool
		},
	}

	result, err := normalizeMiddlewareConfig(input)
	require.NoError(t, err)
	require.Len(t, result, 1)

	assert.Equal(t, "test-middleware", result[0].Name)
	assert.True(t, result[0].Enabled) // Should use default since type assertion fails
}

func TestNormalizeMiddlewareConfig_MapWithNonMapConfig(t *testing.T) {
	// Test behavior when config is not a map
	input := []any{
		map[string]any{
			"name":   "test-middleware",
			"config": "not a map",
		},
	}

	result, err := normalizeMiddlewareConfig(input)
	require.NoError(t, err)
	require.Len(t, result, 1)

	assert.Equal(t, "test-middleware", result[0].Name)
	assert.True(t, result[0].Enabled)
	assert.Nil(t, result[0].Config) // Should be nil since type assertion fails
}

func TestNormalizeMiddlewareConfig_ComplexConfig(t *testing.T) {
	input := []any{
		map[string]any{
			"name":    "rate-limiter",
			"enabled": true,
			"config": map[string]any{
				"requests_per_minute": 100,
				"burst_size":          10,
				"whitelist": []any{
					"127.0.0.1",
					"::1",
				},
				"redis": map[string]any{
					"host": "localhost",
					"port": 6379,
				},
			},
		},
	}

	result, err := normalizeMiddlewareConfig(input)
	require.NoError(t, err)
	require.Len(t, result, 1)

	middleware := result[0]
	assert.Equal(t, "rate-limiter", middleware.Name)
	assert.True(t, middleware.Enabled)
	require.NotNil(t, middleware.Config)

	// Test nested config values
	assert.Equal(t, 100, middleware.Config["requests_per_minute"])
	assert.Equal(t, 10, middleware.Config["burst_size"])

	whitelist := middleware.Config["whitelist"].([]any)
	require.Len(t, whitelist, 2)
	assert.Equal(t, "127.0.0.1", whitelist[0])
	assert.Equal(t, "::1", whitelist[1])

	redis := middleware.Config["redis"].(map[string]any)
	assert.Equal(t, "localhost", redis["host"])
	assert.Equal(t, 6379, redis["port"])
}

func TestNormalizeMiddlewareConfig_NilInput(t *testing.T) {
	result, err := normalizeMiddlewareConfig(nil)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "middleware config must be a list")
}
