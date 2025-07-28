package router_test

import (
	"net/http"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/service"
)

// Mock RPC Service for testing
type MockRpcService struct {
	name string
}

var _ service.Service = (*MockRpcService)(nil)

func TestRouterImpl_MountRpcService_StringService(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, "test-router", map[string]any{})

	serviceName := "user.service"
	result := r.MountRpcService("/rpc", serviceName, false)

	if result == nil {
		t.Error("Expected router to be returned, got nil")
	}

	meta := result.GetMeta()
	if len(meta.Routes) == 0 {
		t.Error("Expected RPC route to be added to router metadata")
	}

	route := meta.Routes[len(meta.Routes)-1]
	if route.Method != "POST" {
		t.Errorf("Expected RPC route method to be POST, got %s", route.Method)
	}
	if route.Path != "/rpc/:method" {
		t.Errorf("Expected RPC route path to be '/rpc/:method', got %s", route.Path)
	}
}

func TestRouterImpl_MountRpcService_ServiceInterface(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, "test-router", map[string]any{})

	mockService := &MockRpcService{name: "user.service"}
	result := r.MountRpcService("/rpc", mockService, false)

	if result == nil {
		t.Error("Expected router to be returned, got nil")
	}

	resMeta := result.GetMeta()
	if len(resMeta.Routes) == 0 {
		t.Error("Expected RPC route to be added to router metadata")
	}

	route := resMeta.Routes[len(resMeta.Routes)-1]
	if route.Handler.Extension == nil {
		t.Error("Expected RPC route to have extension metadata")
	}

	rpcMeta, ok := route.Handler.Extension.(*service.RpcServiceMeta)
	if !ok {
		t.Error("Expected RPC route extension to be *RpcServiceMeta")
	}
	if rpcMeta.ServiceInst.(*MockRpcService).name != "user.service" {
		t.Errorf("Expected RPC service Name to be 'user.service', got %s", rpcMeta.ServiceName)
	}
}

func TestRouterImpl_MountRpcService_RpcServiceMeta(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, "test-router", map[string]any{})

	rpcMeta := &service.RpcServiceMeta{
		ServiceName: "custom.service",
		MethodParam: ":action",
	}

	result := r.MountRpcService("/rpc", rpcMeta, false)

	if result == nil {
		t.Error("Expected router to be returned, got nil")
	}

	routerMeta := result.GetMeta()
	if len(routerMeta.Routes) == 0 {
		t.Error("Expected RPC route to be added to router metadata")
	}

	route := routerMeta.Routes[len(routerMeta.Routes)-1]
	storedRpcMeta, ok := route.Handler.Extension.(*service.RpcServiceMeta)
	if !ok {
		t.Error("Expected RPC route extension to be *RpcServiceMeta")
	}
	if storedRpcMeta.ServiceName != "custom.service" {
		t.Errorf("Expected RPC service name to be 'custom.service', got %s", storedRpcMeta.ServiceName)
	}
	if storedRpcMeta.MethodParam != ":action" {
		t.Errorf("Expected RPC method param to be ':action', got %s", storedRpcMeta.MethodParam)
	}
}

func TestRouterImpl_HandlerTypes(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, "test-router", map[string]any{})

	tests := []struct {
		name    string
		handler any
		isValid bool
	}{
		{
			name: "HandlerFunc",
			handler: func(ctx *request.Context) error {
				return ctx.Ok("OK")
			},
			isValid: true,
		},
		{
			name:    "string handler name",
			handler: "user.getAll",
			isValid: true,
		},
		{
			name: "HandlerMeta",
			handler: &request.HandlerMeta{
				Name: "user.getAll",
				HandlerFunc: func(ctx *request.Context) error {
					return ctx.Ok("OK")
				},
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				result := r.GET("/test", tt.handler)
				if result == nil {
					t.Error("Expected router to be returned for valid handler, got nil")
				}
			} else {
				// For invalid handlers, we expect a panic
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic for invalid handler type, but it didn't panic")
					}
				}()
				r.GET("/test", tt.handler)
			}
		})
	}
}

func TestRouterImpl_StaticMounts(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, "test-router", map[string]any{})

	// Test static mount
	staticDir := http.Dir("./public")
	result := r.MountStatic("/static", staticDir)

	if result == nil {
		t.Error("Expected router to be returned, got nil")
	}

	// Test SPA mount
	result = r.MountSPA("/app", "index.html")

	if result == nil {
		t.Error("Expected router to be returned, got nil")
	}

	// Test reverse proxy mount
	result = r.MountReverseProxy("/api", "http://backend:8080")

	if result == nil {
		t.Error("Expected router to be returned, got nil")
	}
}

func TestRouterImpl_PrefixCleaning(t *testing.T) {
	ctx := &MockRegistrationContext{}

	tests := []struct {
		name         string
		routerPrefix string
		pathPrefix   string
	}{
		{"empty router prefix", "", "/users"},
		{"root router prefix", "/", "/users"},
		{"normal prefixes", "/api", "/users"},
		{"empty path prefix", "/api", ""},
		{"root path prefix", "/api", "/"},
		{"both with slashes", "/api/", "/users/"},
		{"nested prefixes", "/api/v1", "/users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := router.NewRouter(ctx, "test-router", map[string]any{})
			r = r.WithPrefix(tt.routerPrefix)

			handler := func(ctx *request.Context) error {
				return ctx.Ok("OK")
			}

			r.GET(tt.pathPrefix, handler)

			// Since cleanPrefix is internal, we test through the route metadata
			meta := r.GetMeta()
			if len(meta.Routes) == 0 {
				t.Error("Expected route to be added")
				return
			}

			route := meta.Routes[len(meta.Routes)-1]
			// The path should be properly cleaned
			if route.Path == "" {
				t.Error("Expected non-empty path after prefix cleaning")
			}
		})
	}
}
