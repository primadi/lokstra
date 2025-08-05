package router_engine

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewHttpRouterEngine(t *testing.T) {
	tests := []struct {
		name    string
		config  any
		wantErr bool
	}{
		{
			name:    "valid_creation",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "with_config",
			config:  map[string]any{"key": "value"},
			wantErr: false,
		},
		{
			name:    "empty_service_name",
			config:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHttpRouterEngine(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHttpRouterEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return // Expected error, so don't check further
			}

			if got == nil {
				t.Error("NewHttpRouterEngine() returned nil")
				return
			}
		})
	}
}

func TestHttpRouterEngine_HandleMethod(t *testing.T) {
	engine, err := NewHttpRouterEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create HttpRouterEngine: %v", err)
	}

	router := engine.(*HttpRouterEngine)

	// Test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	tests := []struct {
		name    string
		method  string
		path    string
		handler http.Handler
	}{
		{
			name:    "GET_simple_path",
			method:  http.MethodGet,
			path:    "/test",
			handler: testHandler,
		},
		{
			name:    "POST_with_param",
			method:  http.MethodPost,
			path:    "/users/:id",
			handler: testHandler,
		},
		{
			name:    "PUT_complex_path",
			method:  http.MethodPut,
			path:    "/api/v1/users/:id/profile",
			handler: testHandler,
		},
		{
			name:    "DELETE_wildcard",
			method:  http.MethodDelete,
			path:    "/files/*filepath",
			handler: testHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			router.HandleMethod(tt.method, tt.path, tt.handler)
		})
	}
}

func TestHttpRouterEngine_ServeHTTP(t *testing.T) {
	engine, err := NewHttpRouterEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create HttpRouterEngine: %v", err)
	}

	router := engine.(*HttpRouterEngine)

	// Register some test routes
	router.HandleMethod(http.MethodGet, "/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello world"))
	}))

	router.HandleMethod(http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("user id: " + id))
	}))

	router.HandleMethod(http.MethodPost, "/data", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	}))

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET_hello",
			method:         http.MethodGet,
			path:           "/hello",
			expectedStatus: http.StatusOK,
			expectedBody:   "hello world",
		},
		{
			name:           "GET_user_with_param",
			method:         http.MethodGet,
			path:           "/users/123",
			expectedStatus: http.StatusOK,
			expectedBody:   "user id: 123",
		},
		{
			name:           "POST_data",
			method:         http.MethodPost,
			path:           "/data",
			expectedStatus: http.StatusCreated,
			expectedBody:   "created",
		},
		{
			name:           "not_found",
			method:         http.MethodGet,
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			name:           "method_not_allowed",
			method:         http.MethodPut,
			path:           "/hello",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedBody != "" {
				body := w.Body.String()
				if body != tt.expectedBody {
					t.Errorf("Body = %v, want %v", body, tt.expectedBody)
				}
			}
		})
	}
}

func TestHttpRouterEngine_ServeStatic(t *testing.T) {
	engine, err := NewHttpRouterEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create HttpRouterEngine: %v", err)
	}

	router := engine.(*HttpRouterEngine)

	// Create a temporary directory with test files
	tempDir := t.TempDir()
	testFile := tempDir + "/test.txt"
	err = os.WriteFile(testFile, []byte("test file content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test ServeStatic
	router.ServeStatic("/static", http.Dir(tempDir))

	// Test request to static file
	req := httptest.NewRequest(http.MethodGet, "/static/test.txt", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if body != "test file content" {
		t.Errorf("Body = %v, want %v", body, "test file content")
	}
}

func TestHttpRouterEngine_ServeSPA(t *testing.T) {
	engine, err := NewHttpRouterEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create HttpRouterEngine: %v", err)
	}

	router := engine.(*HttpRouterEngine)

	// Create a temporary directory with SPA files
	tempDir := t.TempDir()
	indexFile := tempDir + "/index.html"
	err = os.WriteFile(indexFile, []byte("<html><body>SPA Index</body></html>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create index file: %v", err)
	}

	// Create a static file
	jsFile := tempDir + "/app.js"
	err = os.WriteFile(jsFile, []byte("console.log('app');"), 0644)
	if err != nil {
		t.Fatalf("Failed to create js file: %v", err)
	}

	// Test ServeSPA
	router.ServeSPA("/app", indexFile)

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
			expectedBody:   "<html><body>SPA Index</body></html>",
		},
		{
			name:           "spa_route",
			path:           "/app/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>SPA Index</body></html>",
		},
		{
			name:           "static_file",
			path:           "/app/app.js",
			expectedStatus: http.StatusOK,
			expectedBody:   "console.log('app');",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedStatus)
			}

			body := w.Body.String()
			if body != tt.expectedBody {
				t.Errorf("Body = %v, want %v", body, tt.expectedBody)
			}
		})
	}
}

func TestHttpRouterEngine_ServeReverseProxy(t *testing.T) {
	engine, err := NewHttpRouterEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create HttpRouterEngine: %v", err)
	}

	router := engine.(*HttpRouterEngine)

	// Create a test backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend response: " + r.URL.Path))
	}))
	defer backend.Close()

	// Test ServeReverseProxy
	proxyHandler := func(w http.ResponseWriter, r *http.Request) {
		// Simple proxy simulation - in real implementation this would forward to backend.URL
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend response: " + r.URL.Path))
	}
	router.ServeReverseProxy("/api", proxyHandler)

	// Test request to proxied path
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	expectedBody := "backend response: /test"
	if body != expectedBody {
		t.Errorf("Body = %v, want %v", body, expectedBody)
	}
}

func TestHttpRouterEngine_NotFoundFallback(t *testing.T) {
	engine, err := NewHttpRouterEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create HttpRouterEngine: %v", err)
	}

	router := engine.(*HttpRouterEngine)

	// Add a static file route (which creates the servemux fallback)
	tempDir := t.TempDir()
	router.ServeStatic("/static", http.Dir(tempDir))

	// Add a regular route
	router.HandleMethod(http.MethodGet, "/api/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("api response"))
	}))

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "existing_api_route",
			path:           "/api/test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "static_not_found_but_handled_by_servemux",
			path:           "/static/nonexistent.txt",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "completely_unknown_route",
			path:           "/unknown",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestHttpRouterEngine_Integration(t *testing.T) {
	engine, err := NewHttpRouterEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create HttpRouterEngine: %v", err)
	}

	router := engine.(*HttpRouterEngine)

	// Create temp directory for static files
	tempDir := t.TempDir()
	staticFile := tempDir + "/style.css"
	err = os.WriteFile(staticFile, []byte("body { margin: 0; }"), 0644)
	if err != nil {
		t.Fatalf("Failed to create CSS file: %v", err)
	}

	indexFile := tempDir + "/index.html"
	err = os.WriteFile(indexFile, []byte("<!DOCTYPE html><html><body>App</body></html>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create HTML file: %v", err)
	}

	// Create backend server for proxy
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "backend data"}`))
	}))
	defer backend.Close()

	// Register different types of routes
	router.HandleMethod(http.MethodGet, "/api/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user_id": "` + id + `"}`))
	}))

	router.HandleMethod(http.MethodPost, "/api/data", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Created: " + string(body)))
	}))

	router.ServeStatic("/assets", http.Dir(tempDir))
	router.ServeSPA("/app", indexFile)

	proxyHandler2 := func(w http.ResponseWriter, r *http.Request) {
		// Simple proxy simulation
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "backend data"}`))
	}
	router.ServeReverseProxy("/proxy", proxyHandler2)

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
			name:           "api_with_param",
			method:         http.MethodGet,
			path:           "/api/users/123",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"user_id": "123"}`,
			contentType:    "application/json",
		},
		{
			name:           "api_post",
			method:         http.MethodPost,
			path:           "/api/data",
			body:           "test data",
			expectedStatus: http.StatusCreated,
			expectedBody:   "Created: test data",
		},
		{
			name:           "static_file",
			method:         http.MethodGet,
			path:           "/assets/style.css",
			expectedStatus: http.StatusOK,
			expectedBody:   "body { margin: 0; }",
		},
		{
			name:           "spa_route",
			method:         http.MethodGet,
			path:           "/app/dashboard",
			expectedStatus: http.StatusOK,
			expectedBody:   "<!DOCTYPE html><html><body>App</body></html>",
		},
		{
			name:           "reverse_proxy",
			method:         http.MethodGet,
			path:           "/proxy/test",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message": "backend data"}`,
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
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedStatus)
			}

			body := w.Body.String()
			if body != tt.expectedBody {
				t.Errorf("Body = %v, want %v", body, tt.expectedBody)
			}

			if tt.contentType != "" {
				if ct := w.Header().Get("Content-Type"); ct != tt.contentType {
					t.Errorf("Content-Type = %v, want %v", ct, tt.contentType)
				}
			}
		})
	}
}
