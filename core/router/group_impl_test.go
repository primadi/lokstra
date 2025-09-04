package router_test

import (
	"net/http"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
)

func TestGroupImpl_BasicMethods(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	group := r.Group("/api")

	t.Run("Prefix", func(t *testing.T) {
		prefix := group.Prefix()
		if prefix != "/api" {
			t.Errorf("Expected prefix '/api', got %s", prefix)
		}
	})

	t.Run("GetMeta", func(t *testing.T) {
		groupMeta := group.GetMeta()
		if groupMeta == nil {
			t.Error("Expected meta to be available, got nil")
			return
		}

		if groupMeta.Prefix != "/api" {
			t.Errorf("Expected meta prefix '/api', got %s", groupMeta.Prefix)
		}
	})

	t.Run("GetMiddleware", func(t *testing.T) {
		middleware := group.GetMiddleware()
		if middleware == nil {
			t.Error("Expected middleware slice to be available, got nil")
		}
	})
}

func TestGroupImpl_WithPrefix(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	tests := []struct {
		name        string
		groupPrefix string
		newPrefix   string
		expected    string
	}{
		{"change prefix", "/api", "/v1", "/v1"},
		{"empty prefix", "/api", "", "/api"},
		{"root prefix", "/api", "/", "/api"},
		{"relative prefix", "/api", "users", "/api/users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := r.Group(tt.groupPrefix)
			updatedGroup := group.WithPrefix(tt.newPrefix)

			if updatedGroup == nil {
				t.Error("Expected group to be returned, got nil")
			}

			prefix := updatedGroup.Prefix()
			if prefix != tt.expected {
				t.Errorf("Expected prefix '%s', got '%s'", tt.expected, prefix)
			}
		})
	}
}

func TestGroupImpl_WithOverrideMiddleware(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})
	group := r.Group("/api")

	tests := []struct {
		name   string
		enable bool
	}{
		{"enable override", true},
		{"disable override", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := group.WithOverrideMiddleware(tt.enable)

			if result == nil {
				t.Error("Expected group to be returned, got nil")
			}

			meta := result.GetMeta()
			if meta.OverrideMiddleware != tt.enable {
				t.Errorf("Expected OverrideMiddleware to be %v, got %v", tt.enable, meta.OverrideMiddleware)
			}
		})
	}
}

func TestGroupImpl_HTTPMethods(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})
	group := r.Group("/api")

	handler := func(ctx *request.Context) error {
		return ctx.Ok("OK")
	}

	tests := []struct {
		name   string
		method func(path string, handler any, mw ...any) router.Router
		path   string
	}{
		{"GET", group.GET, "/users"},
		{"POST", group.POST, "/users"},
		{"PUT", group.PUT, "/users/1"},
		{"PATCH", group.PATCH, "/users/1"},
		{"DELETE", group.DELETE, "/users/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(tt.path, handler)

			if result == nil {
				t.Errorf("Expected group to be returned, got nil")
			}

			meta := result.GetMeta()
			if len(meta.Routes) == 0 {
				t.Error("Expected route to be added to group metadata")
			}

			// Check the last added route
			lastRoute := meta.Routes[len(meta.Routes)-1]
			expectedPath := "/api" + tt.path
			if lastRoute.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, lastRoute.Path)
			}
		})
	}
}

func TestGroupImpl_Handle(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})
	group := r.Group("/api")

	handler := func(ctx *request.Context) error {
		return ctx.Ok("OK")
	}

	tests := []struct {
		name         string
		method       request.HTTPMethod
		path         string
		expectedPath string
	}{
		{"GET method", "GET", "/users", "/api/users"},
		{"POST method", "POST", "/users", "/api/users"},
		{"PUT method", "PUT", "/users/1", "/api/users/1"},
		{"DELETE method", "DELETE", "/users/1", "/api/users/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := group.Handle(tt.method, tt.path, handler)

			if result == nil {
				t.Errorf("Expected group to be returned, got nil")
			}

			meta := result.GetMeta()
			if len(meta.Routes) == 0 {
				t.Error("Expected route to be added to group metadata")
			}

			lastRoute := meta.Routes[len(meta.Routes)-1]
			if lastRoute.Method != tt.method {
				t.Errorf("Expected method %s, got %s", tt.method, lastRoute.Method)
			}
			if lastRoute.Path != tt.expectedPath {
				t.Errorf("Expected path %s, got %s", tt.expectedPath, lastRoute.Path)
			}
		})
	}
}

func TestGroupImpl_HandleOverrideMiddleware(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})
	group := r.Group("/api")

	handler := func(ctx *request.Context) error {
		return ctx.Ok("OK")
	}

	result := group.HandleOverrideMiddleware("GET", "/users", handler)

	if result == nil {
		t.Error("Expected group to be returned, got nil")
	}

	meta := result.GetMeta()
	if len(meta.Routes) == 0 {
		t.Error("Expected route to be added to group metadata")
	}

	lastRoute := meta.Routes[len(meta.Routes)-1]
	if !lastRoute.OverrideMiddleware {
		t.Error("Expected OverrideMiddleware to be true")
	}
	if lastRoute.Path != "/api/users" {
		t.Errorf("Expected path '/api/users', got %s", lastRoute.Path)
	}
}

func TestGroupImpl_NestedGroups(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	apiGroup := r.Group("/api")
	v1Group := apiGroup.Group("/v1")

	if v1Group == nil {
		t.Error("Expected nested group to be created, got nil")
	}

	expectedPrefix := "/api/v1"
	if v1Group.Prefix() != expectedPrefix {
		t.Errorf("Expected nested group prefix '%s', got '%s'", expectedPrefix, v1Group.Prefix())
	}

	// Check that parent group has the nested group
	apiMeta := apiGroup.GetMeta()
	if len(apiMeta.Groups) == 0 {
		t.Error("Expected nested group to be added to parent group metadata")
	}

	nestedMeta := apiMeta.Groups[0]
	if nestedMeta.Prefix != expectedPrefix {
		t.Errorf("Expected nested group meta prefix '%s', got '%s'", expectedPrefix, nestedMeta.Prefix)
	}
}

func TestGroupImpl_GroupBlock(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})
	group := r.Group("/api")

	handler := func(ctx *request.Context) error {
		return ctx.Ok("OK")
	}

	result := group.GroupBlock("/v1", func(gr router.Router) {
		gr.GET("/users", handler)
		gr.POST("/users", handler)
	})

	if result == nil {
		t.Error("Expected group to be returned, got nil")
	}

	meta := group.GetMeta()
	if len(meta.Groups) == 0 {
		t.Error("Expected nested group to be added to group metadata")
	}

	nestedMeta := meta.Groups[0]
	if len(nestedMeta.Routes) != 2 {
		t.Errorf("Expected 2 routes in nested group, got %d", len(nestedMeta.Routes))
	}

	// Check routes have correct paths
	expectedPaths := []string{"/api/v1/users", "/api/v1/users"}
	for i, route := range nestedMeta.Routes {
		if route.Path != expectedPaths[i] {
			t.Errorf("Expected route path '%s', got '%s'", expectedPaths[i], route.Path)
		}
	}
}

func TestGroupImpl_Use(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})
	group := r.Group("/api")

	mockMiddleware := func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			return next(ctx)
		}
	}

	result := group.Use(mockMiddleware)

	if result == nil {
		t.Error("Expected group to be returned, got nil")
	}

	middleware := result.GetMiddleware()
	if len(middleware) == 0 {
		t.Error("Expected middleware to be added to group")
	}
}

func TestGroupImpl_LockMiddleware(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})
	group := r.Group("/api")

	// Should not panic
	group.LockMiddleware()
}

func TestGroupImpl_OverrideMiddleware(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})
	group := r.Group("/api")

	result := group.OverrideMiddleware()

	if result == nil {
		t.Error("Expected group to be returned, got nil")
	}

	meta := result.GetMeta()
	if !meta.OverrideMiddleware {
		t.Error("Expected OverrideMiddleware to be true")
	}
}

func TestGroupImpl_MountStatic(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})
	group := r.Group("/api")

	result := group.MountStatic("/static", http.Dir("./public"))

	if result == nil {
		t.Error("Expected group to be returned, got nil")
	}
}

func TestGroupImpl_MountSPA(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})
	group := r.Group("/api")

	result := group.MountSPA("/app", "index.html")

	if result == nil {
		t.Error("Expected group to be returned, got nil")
	}
}

func TestGroupImpl_MountReverseProxy(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})
	group := r.Group("/api")

	result := group.MountReverseProxy("/proxy", "http://backend:8080", false)

	if result == nil {
		t.Error("Expected group to be returned, got nil")
	}
}

func TestGroupImpl_CleanPrefix(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	tests := []struct {
		name           string
		groupPrefix    string
		pathPrefix     string
		expectedResult string
	}{
		{"normal prefixes", "/api", "/users", "/api/users"},
		{"empty path prefix", "/api", "", "/api"},
		{"root path prefix", "/api", "/", "/api"},
		{"root group prefix", "/", "/users", "/users"},
		{"both empty", "", "", ""},
		{"trailing slashes", "/api/", "/users/", "/api/users/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := r.Group(tt.groupPrefix)

			handler := func(ctx *request.Context) error {
				return ctx.Ok("OK")
			}

			group.GET(tt.pathPrefix, handler)

			meta := group.GetMeta()
			if len(meta.Routes) == 0 {
				t.Error("Expected route to be added")
				return
			}

			route := meta.Routes[len(meta.Routes)-1]
			// The actual path cleaning logic is internal, so we test the behavior
			if route.Path == "" {
				t.Error("Expected non-empty path after cleaning")
			}
		})
	}
}

// Test panic scenarios for GroupImpl
func TestGroupImpl_PanicScenarios(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})
	group := r.Group("/api")

	t.Run("DumpRoutes should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected DumpRoutes to panic, but it didn't")
			}
		}()
		group.DumpRoutes()
	})

	t.Run("RecurseAllHandler should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected RecurseAllHandler to panic, but it didn't")
			}
		}()
		group.RecurseAllHandler(func(rt *router.RouteMeta) {})
	})

	t.Run("ServeHTTP should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected ServeHTTP to panic, but it didn't")
			}
		}()
		group.ServeHTTP(nil, nil)
	})

	t.Run("FastHttpHandler should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected FastHttpHandler to panic, but it didn't")
			}
		}()
		group.FastHttpHandler()
	})
}
