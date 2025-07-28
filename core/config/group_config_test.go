package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================
// GROUP CONFIG TESTS
// ==============================================

func TestLoadGroupConfig_SimpleGroup(t *testing.T) {
	groups := []GroupConfig{
		{
			Prefix: "/api",
			MiddlewareRaw: []any{
				"cors",
				map[string]any{
					"name":    "auth",
					"enabled": true,
					"config":  map[string]any{"secret": "test"},
				},
			},
			Routes: []RouteConfig{
				{
					Method:        "GET",
					Path:          "/users",
					Handler:       "user.ListHandler",
					MiddlewareRaw: []any{"rate_limit"},
				},
			},
		},
	}

	err := loadGroupConfig(&groups)
	require.NoError(t, err)

	group := groups[0]

	// Test group middleware normalization
	require.Len(t, group.Middleware, 2)
	assert.Equal(t, "cors", group.Middleware[0].Name)
	assert.Equal(t, "auth", group.Middleware[1].Name)
	assert.Equal(t, "test", group.Middleware[1].Config["secret"])

	// Test route middleware normalization
	route := group.Routes[0]
	require.Len(t, route.Middleware, 1)
	assert.Equal(t, "rate_limit", route.Middleware[0].Name)
}

func TestLoadGroupConfig_NestedGroups(t *testing.T) {
	groups := []GroupConfig{
		{
			Prefix:        "/api",
			MiddlewareRaw: []any{"cors"},
			Groups: []GroupConfig{
				{
					Prefix:        "/v1",
					MiddlewareRaw: []any{"auth"},
					Routes: []RouteConfig{
						{
							Method:        "GET",
							Path:          "/users",
							Handler:       "user.ListHandler",
							MiddlewareRaw: []any{"cache"},
						},
					},
				},
			},
		},
	}

	err := loadGroupConfig(&groups)
	require.NoError(t, err)

	// Test parent group
	parentGroup := groups[0]
	require.Len(t, parentGroup.Middleware, 1)
	assert.Equal(t, "cors", parentGroup.Middleware[0].Name)

	// Test nested group
	require.Len(t, parentGroup.Groups, 1)
	nestedGroup := parentGroup.Groups[0]
	require.Len(t, nestedGroup.Middleware, 1)
	assert.Equal(t, "auth", nestedGroup.Middleware[0].Name)

	// Test nested route
	require.Len(t, nestedGroup.Routes, 1)
	route := nestedGroup.Routes[0]
	require.Len(t, route.Middleware, 1)
	assert.Equal(t, "cache", route.Middleware[0].Name)
}

func TestLoadGroupConfig_ErrorInGroupMiddleware(t *testing.T) {
	groups := []GroupConfig{
		{
			Prefix: "/api",
			MiddlewareRaw: []any{
				map[string]any{
					// Missing name - should cause error
					"enabled": true,
				},
			},
		},
	}

	err := loadGroupConfig(&groups)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "normalize middleware for group /api")
}

func TestLoadGroupConfig_ErrorInRouteMiddleware(t *testing.T) {
	groups := []GroupConfig{
		{
			Prefix: "/api",
			Routes: []RouteConfig{
				{
					Method:  "GET",
					Path:    "/users",
					Handler: "user.Handler",
					MiddlewareRaw: []any{
						map[string]any{
							// Missing name - should cause error
							"enabled": true,
						},
					},
				},
			},
		},
	}

	err := loadGroupConfig(&groups)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "normalize middleware for route /users in group /api")
}

// ==============================================
// GROUP INCLUDES TESTS
// ==============================================

func TestExpandGroupIncludes_SimpleInclude(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "group_include_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create external group file
	externalGroup := `
routes:
  - method: GET
    path: /external
    handler: external.Handler
  - method: POST
    path: /create
    handler: external.CreateHandler

mount_static:
  - prefix: /assets
    folder: ./static

groups:
  - prefix: /nested
    routes:
      - method: GET
        path: /item
        handler: nested.Handler
`

	externalPath := filepath.Join(tempDir, "external.yaml")
	err = os.WriteFile(externalPath, []byte(externalGroup), 0644)
	require.NoError(t, err)

	// Group config that includes external file
	groups := []GroupConfig{
		{
			Prefix:   "/api",
			LoadFrom: []string{"external.yaml"},
			Routes: []RouteConfig{
				{Method: "GET", Path: "/local", Handler: "local.Handler"},
			},
		},
	}

	err = expandGroupIncludes(tempDir, &groups)
	require.NoError(t, err)

	group := groups[0]

	// Should have local route + 2 external routes
	require.Len(t, group.Routes, 3)
	assert.Equal(t, "/local", group.Routes[0].Path)    // Original local route
	assert.Equal(t, "/external", group.Routes[1].Path) // From external file
	assert.Equal(t, "/create", group.Routes[2].Path)   // From external file

	// Should have mount_static from external file
	require.Len(t, group.MountStatic, 1)
	assert.Equal(t, "/assets", group.MountStatic[0].Prefix)
	assert.Equal(t, "./static", group.MountStatic[0].Folder)

	// Should have nested group from external file
	require.Len(t, group.Groups, 1)
	assert.Equal(t, "/nested", group.Groups[0].Prefix)
	require.Len(t, group.Groups[0].Routes, 1)
	assert.Equal(t, "/item", group.Groups[0].Routes[0].Path)
}

func TestExpandGroupIncludes_MultipleIncludes(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "group_multiple_include_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create first external group file
	external1 := `
routes:
  - method: GET
    path: /from-file1
    handler: file1.Handler
`

	// Create second external group file
	external2 := `
routes:
  - method: POST
    path: /from-file2
    handler: file2.Handler
`

	err = os.WriteFile(filepath.Join(tempDir, "file1.yaml"), []byte(external1), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "file2.yaml"), []byte(external2), 0644)
	require.NoError(t, err)

	// Group config that includes both files
	groups := []GroupConfig{
		{
			Prefix:   "/api",
			LoadFrom: []string{"file1.yaml", "file2.yaml"},
		},
	}

	err = expandGroupIncludes(tempDir, &groups)
	require.NoError(t, err)

	group := groups[0]
	require.Len(t, group.Routes, 2)
	assert.Equal(t, "/from-file1", group.Routes[0].Path)
	assert.Equal(t, "/from-file2", group.Routes[1].Path)
}

func TestExpandGroupIncludes_NonExistentFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "group_include_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	groups := []GroupConfig{
		{
			Prefix:   "/api",
			LoadFrom: []string{"non-existent.yaml"},
		},
	}

	err = expandGroupIncludes(tempDir, &groups)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read load_from file")
}

func TestExpandGroupIncludes_InvalidYaml(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "group_include_invalid_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	invalidYaml := `
routes:
  - method: GET
    path: /test
    invalid: [unclosed
`

	err = os.WriteFile(filepath.Join(tempDir, "invalid.yaml"), []byte(invalidYaml), 0644)
	require.NoError(t, err)

	groups := []GroupConfig{
		{
			Prefix:   "/api",
			LoadFrom: []string{"invalid.yaml"},
		},
	}

	err = expandGroupIncludes(tempDir, &groups)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal load_from file")
}

func TestExpandGroupIncludes_WithVariableExpansion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "group_include_vars_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set environment variable
	os.Setenv("API_PREFIX", "/dynamic")
	defer os.Unsetenv("API_PREFIX")

	externalGroup := `
routes:
  - method: GET
    path: ${API_PREFIX}/endpoint
    handler: dynamic.Handler
`

	err = os.WriteFile(filepath.Join(tempDir, "dynamic.yaml"), []byte(externalGroup), 0644)
	require.NoError(t, err)

	groups := []GroupConfig{
		{
			Prefix:   "/api",
			LoadFrom: []string{"dynamic.yaml"},
		},
	}

	err = expandGroupIncludes(tempDir, &groups)
	require.NoError(t, err)

	group := groups[0]
	require.Len(t, group.Routes, 1)
	assert.Equal(t, "/dynamic/endpoint", group.Routes[0].Path)
}

func TestExpandGroupIncludes_ForbiddenRootLevelPrefix(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "group_include_forbidden_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	externalGroup := `
prefix: /forbidden  # This should cause an error
routes:
  - method: GET
    path: /test
    handler: test.Handler
`

	err = os.WriteFile(filepath.Join(tempDir, "forbidden.yaml"), []byte(externalGroup), 0644)
	require.NoError(t, err)

	groups := []GroupConfig{
		{
			Prefix:   "/api",
			LoadFrom: []string{"forbidden.yaml"},
		},
	}

	err = expandGroupIncludes(tempDir, &groups)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prefix not allowed at root level")
}

func TestExpandGroupIncludes_ForbiddenRootLevelOverrideMiddleware(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "group_include_override_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	externalGroup := `
override_middleware: true  # This should cause an error
routes:
  - method: GET
    path: /test
    handler: test.Handler
`

	err = os.WriteFile(filepath.Join(tempDir, "override.yaml"), []byte(externalGroup), 0644)
	require.NoError(t, err)

	groups := []GroupConfig{
		{
			Prefix:   "/api",
			LoadFrom: []string{"override.yaml"},
		},
	}

	err = expandGroupIncludes(tempDir, &groups)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "override_middleware not allowed at root level")
}

func TestExpandGroupIncludes_ForbiddenRootLevelMiddleware(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "group_include_middleware_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	externalGroup := `
middleware:  # This should cause an error
  - cors
routes:
  - method: GET
    path: /test
    handler: test.Handler
`

	err = os.WriteFile(filepath.Join(tempDir, "middleware.yaml"), []byte(externalGroup), 0644)
	require.NoError(t, err)

	groups := []GroupConfig{
		{
			Prefix:   "/api",
			LoadFrom: []string{"middleware.yaml"},
		},
	}

	err = expandGroupIncludes(tempDir, &groups)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "middleware not allowed at root level")
}

func TestExpandGroupIncludes_RecursiveIncludes(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "group_recursive_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create nested directory structure
	subDir := filepath.Join(tempDir, "sub")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Nested group file
	nestedGroup := `
routes:
  - method: GET
    path: /nested
    handler: nested.Handler
`

	err = os.WriteFile(filepath.Join(subDir, "nested.yaml"), []byte(nestedGroup), 0644)
	require.NoError(t, err)

	// Parent group file that includes nested
	parentGroup := `
groups:
  - prefix: /sub
    load_from:
      - sub/nested.yaml
`

	err = os.WriteFile(filepath.Join(tempDir, "parent.yaml"), []byte(parentGroup), 0644)
	require.NoError(t, err)

	// Main group that includes parent
	groups := []GroupConfig{
		{
			Prefix:   "/api",
			LoadFrom: []string{"parent.yaml"},
		},
	}

	err = expandGroupIncludes(tempDir, &groups)
	require.NoError(t, err)

	group := groups[0]
	require.Len(t, group.Groups, 1)

	subGroup := group.Groups[0]
	assert.Equal(t, "/sub", subGroup.Prefix)
	require.Len(t, subGroup.Routes, 1)
	assert.Equal(t, "/nested", subGroup.Routes[0].Path)
}
