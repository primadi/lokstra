package router_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

// Mock implementations for testing
type MockRegistrationContext struct{}

// GetRawHandler implements registration.Context.
func (m *MockRegistrationContext) GetRawHandler(name string) *registration.RawHandlerRegister {
	panic("unimplemented")
}

// RegisterRawHandler implements registration.Context.
func (m *MockRegistrationContext) RegisterRawHandler(name string, handler request.RawHandlerFunc) {
	panic("unimplemented")
}

// RegisterMiddlewareFactoryWithPriority implements registration.Context.
func (m *MockRegistrationContext) RegisterMiddlewareFactoryWithPriority(name string, middlewareFactory midware.Factory, priority int) error {
	return nil
}

// NewPermissionContextFromConfig implements registration.Context.
func (m *MockRegistrationContext) NewPermissionContextFromConfig(
	settings map[string]any,
	permission map[string]any) registration.Context {
	return &MockRegistrationContext{}
}

// RegisterMiddlewareFuncWithPriority implements registration.Context.
func (m *MockRegistrationContext) RegisterMiddlewareFuncWithPriority(name string, middlewareFunc midware.Func, priority int) error {
	return nil
}

// GetMiddlewareFactory implements registration.Context.
func (m *MockRegistrationContext) GetMiddlewareFactory(name string) (midware.Factory, int, bool) {
	return nil, 0, false
}

// RegisterCompiledModule implements registration.Context.
func (m *MockRegistrationContext) RegisterCompiledModule(pluginPath string) error {
	return nil
}

// RegisterCompiledModuleWithFuncName implements registration.Context.
func (m *MockRegistrationContext) RegisterCompiledModuleWithFuncName(pluginPath string, getModuleFuncName string) error {
	return nil
}

// RegisterMiddlewareFactory implements registration.Context.
func (m *MockRegistrationContext) RegisterMiddlewareFactory(name string, middlewareFactory midware.Factory) error {
	return nil
}

// RegisterMiddlewareFunc implements registration.Context.
func (m *MockRegistrationContext) RegisterMiddlewareFunc(name string, middlewareFunc midware.Func) error {
	return nil
}

// RegisterModule implements registration.Context.
func (m *MockRegistrationContext) RegisterModule(getModuleFunc func() registration.Module) error {
	return nil
}

func (m *MockRegistrationContext) CreateService(factoryName, serviceName string, config ...any) (service.Service, error) {
	if factoryName == "" || factoryName == "default" {
		return NewMockRouterEngine(), nil
	}
	return nil, nil
}

func (m *MockRegistrationContext) GetOrCreateService(factoryName, serviceName string, config ...any) (service.Service, error) {
	if svc, err := m.GetService(serviceName); err == nil {
		return svc, nil // Return existing service if found
	}
	return m.CreateService(factoryName, serviceName, config...)
}

func (m *MockRegistrationContext) GetValue(key string) (any, bool) { return nil, false }
func (m *MockRegistrationContext) SetValue(key string, value any)  {}

func (m *MockRegistrationContext) RegisterService(name string, svc service.Service) error { return nil }
func (m *MockRegistrationContext) GetService(serviceName string) (service.Service, error) {
	return nil, nil
}
func (m *MockRegistrationContext) RegisterServiceFactory(factoryName string, serviceFactory func(config any) (service.Service, error)) {
}
func (m *MockRegistrationContext) GetServiceFactory(factoryName string) (service.ServiceFactory, bool) {
	if factoryName == "lokstra.http_router.default" {
		return func(_ any) (service.Service, error) {
			return NewMockRouterEngine(), nil
		}, true
	}

	if factoryName == "lokstra.http_listener.default" {
		return func(_ any) (service.Service, error) {
			return NewMockListener(), nil
		}, true
	}

	return nil, false
}
func (m *MockRegistrationContext) GetServiceFactories(pattern string) []service.ServiceFactory {
	return []service.ServiceFactory{}
}
func (m *MockRegistrationContext) GetHandler(name string) *request.HandlerRegister { return nil }
func (m *MockRegistrationContext) RegisterHandler(name string, handler any)        {}

type MockRouterEngine struct {
	routes         map[string]map[string]http.Handler
	staticMounts   map[string]http.Dir
	spaMounts      map[string]string
	reverseProxies map[string]http.HandlerFunc
}

func NewMockRouterEngine() *MockRouterEngine {
	return &MockRouterEngine{
		routes:         make(map[string]map[string]http.Handler),
		staticMounts:   make(map[string]http.Dir),
		spaMounts:      make(map[string]string),
		reverseProxies: make(map[string]http.HandlerFunc),
	}
}

func (m *MockRouterEngine) HandleMethod(method request.HTTPMethod, path string, handler http.Handler) {
	if m.routes[string(method)] == nil {
		m.routes[string(method)] = make(map[string]http.Handler)
	}
	m.routes[string(method)][path] = handler
}

// RawHandle implements serviceapi.RouterEngine.
func (m *MockRouterEngine) RawHandle(pattern string, handler http.Handler) {
	panic("unimplemented")
}

// RawHandleFunc implements serviceapi.RouterEngine.
func (m *MockRouterEngine) RawHandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	panic("unimplemented")
}

// ServeStatic implements serviceapi.RouterEngine.
func (m *MockRouterEngine) ServeStatic(prefix string, spa bool, sources ...fs.FS) {
	panic("unimplemented")
}

// ServeHtmxPage implements serviceapi.RouterEngine.
func (m *MockRouterEngine) ServeHtmxPage(pageDataRouter http.Handler,
	prefix string, sources ...fs.FS) {
	panic("unimplemented")
}

func (m *MockRouterEngine) ServeReverseProxy(prefix string, handler http.HandlerFunc) {
	m.reverseProxies[prefix] = handler
}

func (m *MockRouterEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if methodRoutes, ok := m.routes[r.Method]; ok {
		if handler, ok := methodRoutes[r.URL.Path]; ok {
			handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

type MockHttpListener struct{}

// GetStartMessage implements serviceapi.HttpListener.
func (m *MockHttpListener) GetStartMessage(addr string) string {
	panic("unimplemented")
}

func (m *MockHttpListener) ListenAndServe(addr string, handler http.Handler) error { return nil }
func (m *MockHttpListener) Shutdown(shutdownTimeout time.Duration) error           { return nil }
func (m *MockHttpListener) IsRunning() bool                                        { return false }
func (m *MockHttpListener) ActiveRequest() int                                     { return 0 }

var _ registration.Context = (*MockRegistrationContext)(nil)
var _ serviceapi.RouterEngine = (*MockRouterEngine)(nil)
var _ serviceapi.HttpListener = (*MockHttpListener)(nil)
var _ service.Service = (*MockRouterEngine)(nil)
var _ service.Service = (*MockHttpListener)(nil)

func NewMockListener() serviceapi.HttpListener {
	return &MockHttpListener{}
}

func TestNewRouter(t *testing.T) {
	ctx := &MockRegistrationContext{}

	tests := []struct {
		name   string
		config map[string]any
	}{
		{
			name:   "basic router creation",
			config: map[string]any{},
		},
		{
			name:   "router with config",
			config: map[string]any{"port": 8080},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := router.NewRouter(ctx, tt.config)

			if r == nil {
				t.Error("Expected router to be created, got nil")
			}

			if r.Prefix() != "/" {
				t.Errorf("Expected empty prefix, got %s", r.Prefix())
			}
		})
	}
}

func TestNewRouterWithEngine(t *testing.T) {
	ctx := &MockRegistrationContext{}

	tests := []struct {
		name       string
		engineType string
		config     map[string]any
	}{
		{
			name:       "default engine",
			engineType: "default",
			config:     map[string]any{},
		},
		{
			name:       "empty engine type",
			engineType: "",
			config:     map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := router.NewRouterWithEngine(ctx, tt.engineType, tt.config)

			if r == nil {
				t.Error("Expected router to be created, got nil")
			}

			meta := r.GetMeta()
			if meta == nil {
				t.Error("Expected router meta to be available, got nil")
			}
		})
	}
}

func TestNewListener(t *testing.T) {
	ctx := &MockRegistrationContext{}

	tests := []struct {
		name   string
		config map[string]any
	}{
		{
			name:   "basic listener creation",
			config: map[string]any{},
		},
		{
			name:   "listener with config",
			config: map[string]any{"port": 8080},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener := router.NewListener(ctx, tt.config)

			if listener == nil {
				t.Error("Expected listener to be created, got nil")
			}
		})
	}
}

func TestNewListenerWithEngine(t *testing.T) {
	ctx := &MockRegistrationContext{}

	tests := []struct {
		name         string
		listenerType string
		config       map[string]any
	}{
		{
			name:         "default listener",
			listenerType: "default",
			config:       map[string]any{},
		},
		{
			name:         "empty listener type",
			listenerType: "",
			config:       map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener := router.NewListenerWithEngine(ctx, tt.listenerType, tt.config)

			if listener == nil {
				t.Error("Expected listener to be created, got nil")
			}
		})
	}
}

func TestRouterImpl_BasicMethods(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	t.Run("Prefix", func(t *testing.T) {
		prefix := r.Prefix()
		if prefix != "/" {
			t.Errorf("Expected empty prefix, got %s", prefix)
		}
	})

	t.Run("WithPrefix", func(t *testing.T) {
		newRouter := r.WithPrefix("/api")
		newPrefix := newRouter.Prefix()
		if newPrefix != "/api" {
			t.Errorf("Expected prefix '/api', got %s", newPrefix)
		}
	})

	t.Run("GetMeta", func(t *testing.T) {
		meta := r.GetMeta()
		if meta == nil {
			t.Error("Expected meta to be available, got nil")
		}
	})

	t.Run("GetMiddleware", func(t *testing.T) {
		middleware := r.GetMiddleware()
		if middleware == nil {
			t.Error("Expected middleware slice to be available, got nil")
		}
	})
}

func TestRouterImpl_HTTPMethods(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	handler := func(ctx *request.Context) error {
		return ctx.Ok("OK")
	}

	tests := []struct {
		name   string
		method func(path string, handler any, mw ...any) router.Router
		path   string
	}{
		{"GET", r.GET, "/users"},
		{"POST", r.POST, "/users"},
		{"PUT", r.PUT, "/users/1"},
		{"PATCH", r.PATCH, "/users/1"},
		{"DELETE", r.DELETE, "/users/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(tt.path, handler)

			if result == nil {
				t.Errorf("Expected router to be returned, got nil")
			}

			// Check if the route was added to metadata
			meta := result.GetMeta()
			if len(meta.Routes) == 0 {
				t.Error("Expected route to be added to router metadata")
			}
		})
	}
}

func TestRouterImpl_Handle(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	handler := func(ctx *request.Context) error {
		return ctx.Ok("OK")
	}

	tests := []struct {
		name   string
		method request.HTTPMethod
		path   string
	}{
		{"GET method", "GET", "/test"},
		{"POST method", "POST", "/test"},
		{"PUT method", "PUT", "/test"},
		{"DELETE method", "DELETE", "/test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.Handle(tt.method, tt.path, handler)

			if result == nil {
				t.Errorf("Expected router to be returned, got nil")
			}

			meta := result.GetMeta()
			if len(meta.Routes) == 0 {
				t.Error("Expected route to be added to router metadata")
			}

			// Check the last added route
			lastRoute := meta.Routes[len(meta.Routes)-1]
			if lastRoute.Method != tt.method {
				t.Errorf("Expected method %s, got %s", tt.method, lastRoute.Method)
			}
			if lastRoute.Path != tt.path {
				t.Errorf("Expected path %s, got %s", tt.path, lastRoute.Path)
			}
		})
	}
}

func TestRouterImpl_HandleOverrideMiddleware(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	handler := func(ctx *request.Context) error {
		return ctx.Ok("OK")
	}

	result := r.HandleOverrideMiddleware("GET", "/test", handler)

	if result == nil {
		t.Error("Expected router to be returned, got nil")
	}

	meta := result.GetMeta()
	if len(meta.Routes) == 0 {
		t.Error("Expected route to be added to router metadata")
	}

	lastRoute := meta.Routes[len(meta.Routes)-1]
	if !lastRoute.OverrideMiddleware {
		t.Error("Expected OverrideMiddleware to be true")
	}
}

func TestRouterImpl_WithOverrideMiddleware(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	tests := []struct {
		name   string
		enable bool
	}{
		{"enable override", true},
		{"disable override", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.WithOverrideMiddleware(tt.enable)

			if result == nil {
				t.Error("Expected router to be returned, got nil")
			}

			meta := result.GetMeta()
			if meta.OverrideMiddleware != tt.enable {
				t.Errorf("Expected OverrideMiddleware to be %v, got %v", tt.enable, meta.OverrideMiddleware)
			}
		})
	}
}

func TestRouterImpl_Group(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	tests := []struct {
		name   string
		prefix string
	}{
		{"api group", "/api"},
		{"v1 group", "/v1"},
		{"empty prefix", ""},
		{"root prefix", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := r.Group(tt.prefix)

			if group == nil {
				t.Error("Expected group to be created, got nil")
			}

			meta := r.GetMeta()
			if len(meta.Groups) == 0 {
				t.Error("Expected group to be added to router metadata")
			}
		})
	}
}

func TestRouterImpl_GroupBlock(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	handler := func(ctx *request.Context) error {
		return ctx.Ok("OK")
	}

	result := r.GroupBlock("/api", func(gr router.Router) {
		gr.GET("/users", handler)
		gr.POST("/users", handler)
	})

	if result == nil {
		t.Error("Expected router to be returned, got nil")
	}

	meta := result.GetMeta()
	if len(meta.Groups) == 0 {
		t.Error("Expected group to be added to router metadata")
	}

	groupMeta := meta.Groups[0]
	if len(groupMeta.Routes) != 2 {
		t.Errorf("Expected 2 routes in group, got %d", len(groupMeta.Routes))
	}
}

func TestRouterImpl_MountStatic(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	result := r.MountStatic("/static", false, os.DirFS("./public"))

	if result == nil {
		t.Error("Expected router to be returned, got nil")
	}
}

func TestRouterImpl_MountSPA(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	result := r.MountStatic("/app", true, os.DirFS("./"))

	if result == nil {
		t.Error("Expected router to be returned, got nil")
	}
}

func TestRouterImpl_MountReverseProxy(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	result := r.MountReverseProxy("/api", "http://backend:8080", false)

	if result == nil {
		t.Error("Expected router to be returned, got nil")
	}
}

func TestRouterImpl_ServeHTTP(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// This should not panic
	r.ServeHTTP(w, req)
}

func TestRouterImpl_OverrideMiddleware(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	result := r.WithOverrideMiddleware(true)

	if result == nil {
		t.Error("Expected router to be returned, got nil")
	}

	meta := result.GetMeta()
	if !meta.OverrideMiddleware {
		t.Error("Expected OverrideMiddleware to be true")
	}
}

func TestRouterImpl_RecurseAllHandler(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	handler := func(ctx *request.Context) error {
		return ctx.Ok("OK")
	}

	r.GET("/test1", handler)
	r.POST("/test2", handler)

	var routes []*router.RouteMeta
	r.RecurseAllHandler(func(rt *router.RouteMeta) {
		routes = append(routes, rt)
	})

	if len(routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(routes))
	}
}

func TestRouterImpl_DumpRoutes(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	handler := func(ctx *request.Context) error {
		return ctx.Ok("OK")
	}

	r.GET("/test", handler)

	// Should not panic
	r.DumpRoutes()
}

func TestRouterImpl_FastHttpHandler(t *testing.T) {
	ctx := &MockRegistrationContext{}
	r := router.NewRouter(ctx, map[string]any{})

	handler := r.FastHttpHandler()

	if handler == nil {
		t.Error("Expected FastHTTP handler to be returned, got nil")
	}
}
