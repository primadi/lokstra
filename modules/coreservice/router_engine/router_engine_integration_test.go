package router_engine

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/primadi/lokstra/serviceapi"
)

func TestRouterEngineInterface(t *testing.T) {
	// Test that both implementations satisfy the RouterEngine interface
	tests := []struct {
		name    string
		factory func(serviceName string, config any) (serviceapi.RouterEngine, error)
	}{
		{
			name: "HttpRouterEngine",
			factory: func(serviceName string, config any) (serviceapi.RouterEngine, error) {
				service, err := NewHttpRouterEngine(config)
				if err != nil {
					return nil, err
				}
				return service.(*HttpRouterEngine), nil
			},
		},
		{
			name: "ServeMuxEngine",
			factory: func(serviceName string, config any) (serviceapi.RouterEngine, error) {
				service, err := NewServeMuxEngine(config)
				if err != nil {
					return nil, err
				}
				return service.(*ServeMuxEngine), nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := tt.factory("test", nil)
			if err != nil {
				t.Fatalf("Failed to create %s: %v", tt.name, err)
			}

			if engine == nil {
				t.Fatalf("%s returned nil", tt.name)
			}

			// Test interface methods exist
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test"))
			})

			// Should not panic
			engine.HandleMethod(http.MethodGet, "/test", testHandler)

			// Create temp directory for static/SPA tests
			tempDir := t.TempDir()
			testFile := tempDir + "/test.txt"
			err = os.WriteFile(testFile, []byte("static content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			indexFile := tempDir + "/index.html"
			err = os.WriteFile(indexFile, []byte("<html><body>SPA</body></html>"), 0644)
			if err != nil {
				t.Fatalf("Failed to create index file: %v", err)
			}

			// Backend server for proxy testing
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("proxied"))
			}))
			defer backend.Close()

			// Test all interface methods
			engine.ServeStatic("/static", http.Dir(tempDir))
			engine.ServeSPA("/app", indexFile)

			proxyHandler := func(w http.ResponseWriter, r *http.Request) {
				// Simple proxy simulation
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("proxy response"))
			}
			engine.ServeReverseProxy("/proxy", proxyHandler)

			// Test ServeHTTP
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("%s ServeHTTP failed: status %d", tt.name, w.Code)
			}
		})
	}
}

func TestRouterEngineComparison(t *testing.T) {
	// Compare behavior between HttpRouter and ServeMux engines
	httpEngine, err := NewHttpRouterEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create HttpRouterEngine: %v", err)
	}

	serveMuxEngine, err := NewServeMuxEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create ServeMuxEngine: %v", err)
	}

	httpRouter := httpEngine.(*HttpRouterEngine)
	serveMuxRouter := serveMuxEngine.(*ServeMuxEngine)

	// Register same routes on both engines
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	httpRouter.HandleMethod(http.MethodGet, "/api/users/:id", testHandler)
	serveMuxRouter.HandleMethod(http.MethodGet, "/api/users/:id", testHandler)

	httpRouter.HandleMethod(http.MethodPost, "/api/data", testHandler)
	serveMuxRouter.HandleMethod(http.MethodPost, "/api/data", testHandler)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "GET_with_param",
			method:         http.MethodGet,
			path:           "/api/users/123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST_data",
			method:         http.MethodPost,
			path:           "/api/data",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not_found",
			method:         http.MethodGet,
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_HttpRouter", func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			httpRouter.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("HttpRouter status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})

		t.Run(tt.name+"_ServeMux", func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			serveMuxRouter.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("ServeMux status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestRouterEngineMethodNotAllowedBehavior(t *testing.T) {
	// Test different behavior between engines for method not allowed
	httpEngine, err := NewHttpRouterEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create HttpRouterEngine: %v", err)
	}

	serveMuxEngine, err := NewServeMuxEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create ServeMuxEngine: %v", err)
	}

	httpRouter := httpEngine.(*HttpRouterEngine)
	serveMuxRouter := serveMuxEngine.(*ServeMuxEngine)

	// Register only GET handler on both
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("get response"))
	})

	httpRouter.HandleMethod(http.MethodGet, "/test", testHandler)
	serveMuxRouter.HandleMethod(http.MethodGet, "/test", testHandler)

	// Test POST to GET-only route
	t.Run("HttpRouter_POST_to_GET_route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		w := httptest.NewRecorder()

		httpRouter.ServeHTTP(w, req)

		// HttpRouter returns 405 Method Not Allowed when path exists but method doesn't match
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("HttpRouter status = %v, expected %v (405 Method Not Allowed)", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("ServeMux_POST_to_GET_route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		w := httptest.NewRecorder()

		serveMuxRouter.ServeHTTP(w, req)

		// ServeMux should return 405 Method Not Allowed with Allow header
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("ServeMux status = %v, expected %v", w.Code, http.StatusMethodNotAllowed)
		}

		allowHeader := w.Header().Get("Allow")
		if allowHeader != "GET" {
			t.Errorf("Allow header = %v, expected %v", allowHeader, "GET")
		}
	})
}

func TestRouterEngineStaticFileHandling(t *testing.T) {
	// Test static file handling differences
	tempDir := t.TempDir()

	// Create test files
	testFile := tempDir + "/test.txt"
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	subDir := tempDir + "/css"
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	cssFile := subDir + "/style.css"
	err = os.WriteFile(cssFile, []byte("body { color: blue; }"), 0644)
	if err != nil {
		t.Fatalf("Failed to create CSS file: %v", err)
	}

	engines := []struct {
		name   string
		engine serviceapi.RouterEngine
	}{
		{
			name: "HttpRouter",
			engine: func() serviceapi.RouterEngine {
				e, _ := NewHttpRouterEngine(nil)
				return e.(*HttpRouterEngine)
			}(),
		},
		{
			name: "ServeMux",
			engine: func() serviceapi.RouterEngine {
				e, _ := NewServeMuxEngine(nil)
				return e.(*ServeMuxEngine)
			}(),
		},
	}

	for _, engineTest := range engines {
		t.Run(engineTest.name, func(t *testing.T) {
			engineTest.engine.ServeStatic("/assets", http.Dir(tempDir))

			tests := []struct {
				name           string
				path           string
				expectedStatus int
				expectedBody   string
			}{
				{
					name:           "root_file",
					path:           "/assets/test.txt",
					expectedStatus: http.StatusOK,
					expectedBody:   "test content",
				},
				{
					name:           "nested_file",
					path:           "/assets/css/style.css",
					expectedStatus: http.StatusOK,
					expectedBody:   "body { color: blue; }",
				},
				{
					name:           "not_found",
					path:           "/assets/nonexistent.txt",
					expectedStatus: http.StatusNotFound,
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					req := httptest.NewRequest(http.MethodGet, tt.path, nil)
					w := httptest.NewRecorder()

					engineTest.engine.ServeHTTP(w, req)

					if w.Code != tt.expectedStatus {
						t.Errorf("%s status = %v, want %v", engineTest.name, w.Code, tt.expectedStatus)
					}

					if tt.expectedBody != "" {
						body := w.Body.String()
						if body != tt.expectedBody {
							t.Errorf("%s body = %v, want %v", engineTest.name, body, tt.expectedBody)
						}
					}
				})
			}
		})
	}
}

func TestRouterEngineSPAHandling(t *testing.T) {
	// Test SPA handling across both engines
	tempDir := t.TempDir()

	indexFile := tempDir + "/index.html"
	err := os.WriteFile(indexFile, []byte("<!DOCTYPE html><html><body>SPA App</body></html>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create index file: %v", err)
	}

	// Create static assets
	jsFile := tempDir + "/app.js"
	err = os.WriteFile(jsFile, []byte("console.log('SPA initialized');"), 0644)
	if err != nil {
		t.Fatalf("Failed to create JS file: %v", err)
	}

	engines := []struct {
		name   string
		engine serviceapi.RouterEngine
	}{
		{
			name: "HttpRouter",
			engine: func() serviceapi.RouterEngine {
				e, _ := NewHttpRouterEngine(nil)
				return e.(*HttpRouterEngine)
			}(),
		},
		{
			name: "ServeMux",
			engine: func() serviceapi.RouterEngine {
				e, _ := NewServeMuxEngine(nil)
				return e.(*ServeMuxEngine)
			}(),
		},
	}

	for _, engineTest := range engines {
		t.Run(engineTest.name, func(t *testing.T) {
			engineTest.engine.ServeSPA("/app", indexFile)

			tests := []struct {
				name           string
				path           string
				expectedStatus int
				expectedBody   string
			}{
				{
					name:           "spa_root",
					path:           "/app",
					expectedStatus: http.StatusOK,
					expectedBody:   "<!DOCTYPE html><html><body>SPA App</body></html>",
				},
				{
					name:           "spa_route",
					path:           "/app/dashboard",
					expectedStatus: http.StatusOK,
					expectedBody:   "<!DOCTYPE html><html><body>SPA App</body></html>",
				},
				{
					name:           "spa_nested_route",
					path:           "/app/users/123/profile",
					expectedStatus: http.StatusOK,
					expectedBody:   "<!DOCTYPE html><html><body>SPA App</body></html>",
				},
				{
					name:           "static_asset",
					path:           "/app/app.js",
					expectedStatus: http.StatusOK,
					expectedBody:   "console.log('SPA initialized');",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					req := httptest.NewRequest(http.MethodGet, tt.path, nil)
					w := httptest.NewRecorder()

					engineTest.engine.ServeHTTP(w, req)

					if w.Code != tt.expectedStatus {
						t.Errorf("%s status = %v, want %v", engineTest.name, w.Code, tt.expectedStatus)
					}

					body := w.Body.String()
					if body != tt.expectedBody {
						t.Errorf("%s body = %v, want %v", engineTest.name, body, tt.expectedBody)
					}
				})
			}
		})
	}
}

func TestRouterEngineProxyHandling(t *testing.T) {
	// Test reverse proxy handling across both engines
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Backend-Method", r.Method)
		w.WriteHeader(http.StatusOK)

		body, _ := io.ReadAll(r.Body)
		response := `{"path": "` + r.URL.Path + `", "method": "` + r.Method + `", "body": "` + string(body) + `"}`
		w.Write([]byte(response))
	}))
	defer backend.Close()

	engines := []struct {
		name   string
		engine serviceapi.RouterEngine
	}{
		{
			name: "HttpRouter",
			engine: func() serviceapi.RouterEngine {
				e, _ := NewHttpRouterEngine(nil)
				return e.(*HttpRouterEngine)
			}(),
		},
		{
			name: "ServeMux",
			engine: func() serviceapi.RouterEngine {
				e, _ := NewServeMuxEngine(nil)
				return e.(*ServeMuxEngine)
			}(),
		},
	}

	for _, engineTest := range engines {
		t.Run(engineTest.name, func(t *testing.T) {
			proxyHandler2 := func(w http.ResponseWriter, r *http.Request) {
				// Simple proxy simulation that mimics backend response
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Backend-Method", r.Method)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"path": "` + r.URL.Path + `", "method": "` + r.Method + `"}`))
			}
			engineTest.engine.ServeReverseProxy("/api", proxyHandler2)

			tests := []struct {
				name           string
				method         string
				path           string
				body           string
				expectedStatus int
				expectedPath   string
			}{
				{
					name:           "GET_request",
					method:         http.MethodGet,
					path:           "/api/users",
					expectedStatus: http.StatusOK,
					expectedPath:   "/users",
				},
				{
					name:           "POST_request",
					method:         http.MethodPost,
					path:           "/api/data",
					body:           "test data",
					expectedStatus: http.StatusOK,
					expectedPath:   "/data",
				},
				{
					name:           "nested_path",
					method:         http.MethodGet,
					path:           "/api/v1/users/123",
					expectedStatus: http.StatusOK,
					expectedPath:   "/v1/users/123",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					var req *http.Request
					if tt.body != "" {
						req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
					} else {
						req = httptest.NewRequest(tt.method, tt.path, nil)
					}

					w := httptest.NewRecorder()
					engineTest.engine.ServeHTTP(w, req)

					if w.Code != tt.expectedStatus {
						t.Errorf("%s status = %v, want %v", engineTest.name, w.Code, tt.expectedStatus)
					}

					// Check backend headers were forwarded
					if method := w.Header().Get("X-Backend-Method"); method != tt.method {
						t.Errorf("%s backend method = %v, want %v", engineTest.name, method, tt.method)
					}

					// Check Content-Type was forwarded
					if ct := w.Header().Get("Content-Type"); ct != "application/json" {
						t.Errorf("%s content-type = %v, want %v", engineTest.name, ct, "application/json")
					}

					// Check response contains expected path
					body := w.Body.String()
					if !strings.Contains(body, `"path": "`+tt.expectedPath+`"`) {
						t.Errorf("%s response doesn't contain expected path %v: %v", engineTest.name, tt.expectedPath, body)
					}

					// Check method was proxied correctly
					if !strings.Contains(body, `"method": "`+tt.method+`"`) {
						t.Errorf("%s response doesn't contain expected method %v: %v", engineTest.name, tt.method, body)
					}
				})
			}
		})
	}
}

func TestRouterEngineComplexIntegration(t *testing.T) {
	// Test complex scenarios with multiple route types on both engines
	tempDir := t.TempDir()

	// Create files
	staticFile := tempDir + "/style.css"
	err := os.WriteFile(staticFile, []byte("body { font-family: Arial; }"), 0644)
	if err != nil {
		t.Fatalf("Failed to create CSS file: %v", err)
	}

	indexFile := tempDir + "/index.html"
	err = os.WriteFile(indexFile, []byte("<!DOCTYPE html><html><body>Full App</body></html>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create HTML file: %v", err)
	}

	// Backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service": "external", "path": "` + r.URL.Path + `"}`))
	}))
	defer backend.Close()

	engines := []struct {
		name   string
		engine serviceapi.RouterEngine
	}{
		{
			name: "HttpRouter",
			engine: func() serviceapi.RouterEngine {
				e, _ := NewHttpRouterEngine(nil)
				return e.(*HttpRouterEngine)
			}(),
		},
		{
			name: "ServeMux",
			engine: func() serviceapi.RouterEngine {
				e, _ := NewServeMuxEngine(nil)
				return e.(*ServeMuxEngine)
			}(),
		},
	}

	for _, engineTest := range engines {
		t.Run(engineTest.name, func(t *testing.T) {
			// Register all types of routes
			engineTest.engine.HandleMethod(http.MethodGet, "/api/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "healthy"}`))
			}))

			engineTest.engine.HandleMethod(http.MethodPost, "/api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				// Properly escape the JSON content
				escapedBody := strings.ReplaceAll(string(body), `"`, `\"`)
				w.Write([]byte(`{"created": true, "data": "` + escapedBody + `"}`))
			}))

			engineTest.engine.ServeStatic("/assets", http.Dir(tempDir))
			engineTest.engine.ServeSPA("/app", indexFile)

			proxyHandler3 := func(w http.ResponseWriter, r *http.Request) {
				// Simple proxy simulation for external service
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				// Remove the /external prefix to get the path that would be sent to external service
				path := strings.TrimPrefix(r.URL.Path, "/external")
				w.Write([]byte(`{"service": "external", "path": "` + path + `"}`))
			}
			engineTest.engine.ServeReverseProxy("/external", proxyHandler3)

			// Test all route types
			tests := []struct {
				name           string
				method         string
				path           string
				body           string
				expectedStatus int
				expectedBody   string
				contentType    string
			}{
				{
					name:           "api_health",
					method:         http.MethodGet,
					path:           "/api/health",
					expectedStatus: http.StatusOK,
					expectedBody:   `{"status": "healthy"}`,
					contentType:    "application/json",
				},
				{
					name:           "api_create_user",
					method:         http.MethodPost,
					path:           "/api/users",
					body:           `{"name": "John"}`,
					expectedStatus: http.StatusCreated,
					expectedBody:   `{"created": true, "data": "{\"name\": \"John\"}"}`,
					contentType:    "application/json",
				},
				{
					name:           "static_css",
					method:         http.MethodGet,
					path:           "/assets/style.css",
					expectedStatus: http.StatusOK,
					expectedBody:   "body { font-family: Arial; }",
				},
				{
					name:           "spa_root",
					method:         http.MethodGet,
					path:           "/app",
					expectedStatus: http.StatusOK,
					expectedBody:   "<!DOCTYPE html><html><body>Full App</body></html>",
				},
				{
					name:           "spa_route",
					method:         http.MethodGet,
					path:           "/app/dashboard",
					expectedStatus: http.StatusOK,
					expectedBody:   "<!DOCTYPE html><html><body>Full App</body></html>",
				},
				{
					name:           "external_proxy",
					method:         http.MethodGet,
					path:           "/external/api/data",
					expectedStatus: http.StatusOK,
					expectedBody:   `{"service": "external", "path": "/api/data"}`,
					contentType:    "application/json",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					var req *http.Request
					if tt.body != "" {
						req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
					} else {
						req = httptest.NewRequest(tt.method, tt.path, nil)
					}

					w := httptest.NewRecorder()
					engineTest.engine.ServeHTTP(w, req)

					if w.Code != tt.expectedStatus {
						t.Errorf("%s %s status = %v, want %v", engineTest.name, tt.name, w.Code, tt.expectedStatus)
					}

					body := w.Body.String()
					if body != tt.expectedBody {
						t.Errorf("%s %s body = %v, want %v", engineTest.name, tt.name, body, tt.expectedBody)
					}

					if tt.contentType != "" {
						if ct := w.Header().Get("Content-Type"); ct != tt.contentType {
							t.Errorf("%s %s content-type = %v, want %v", engineTest.name, tt.name, ct, tt.contentType)
						}
					}
				})
			}
		})
	}
}
