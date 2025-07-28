package router_engine

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewServeMuxEngine(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewServeMuxEngine(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServeMuxEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return // Expected error, so don't check further
			}

			if got == nil {
				t.Error("NewServeMuxEngine() returned nil")
				return
			}
		})
	}
}

func TestServeMuxEngine_HandleMethod(t *testing.T) {
	engine, err := NewServeMuxEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create ServeMuxEngine: %v", err)
	}

	router := engine.(*ServeMuxEngine)

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
			name:    "POST_api_path",
			method:  http.MethodPost,
			path:    "/api/users",
			handler: testHandler,
		},
		{
			name:    "PUT_complex_path",
			method:  http.MethodPut,
			path:    "/api/v1/users",
			handler: testHandler,
		},
		{
			name:    "DELETE_path",
			method:  http.MethodDelete,
			path:    "/api/data",
			handler: testHandler,
		},
		{
			name:    "PATCH_path",
			method:  http.MethodPatch,
			path:    "/api/patch",
			handler: testHandler,
		},
		{
			name:    "HEAD_path",
			method:  http.MethodHead,
			path:    "/api/head",
			handler: testHandler,
		},
		{
			name:    "OPTIONS_path",
			method:  http.MethodOptions,
			path:    "/api/options",
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

func TestServeMuxEngine_ServeHTTP(t *testing.T) {
	engine, err := NewServeMuxEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create ServeMuxEngine: %v", err)
	}

	router := engine.(*ServeMuxEngine)

	// Register test routes
	router.HandleMethod(http.MethodGet, "/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello world"))
	}))

	router.HandleMethod(http.MethodPost, "/data", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created: " + string(body)))
	}))

	router.HandleMethod(http.MethodPut, "/update", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("updated"))
	}))

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
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
			name:           "POST_data",
			method:         http.MethodPost,
			path:           "/data",
			body:           "test payload",
			expectedStatus: http.StatusCreated,
			expectedBody:   "created: test payload",
		},
		{
			name:           "PUT_update",
			method:         http.MethodPut,
			path:           "/update",
			expectedStatus: http.StatusOK,
			expectedBody:   "updated",
		},
		{
			name:           "not_found",
			method:         http.MethodGet,
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
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

			if tt.expectedBody != "" {
				body := w.Body.String()
				if body != tt.expectedBody {
					t.Errorf("Body = %v, want %v", body, tt.expectedBody)
				}
			}
		})
	}
}

func TestServeMuxEngine_MethodNotAllowed(t *testing.T) {
	engine, err := NewServeMuxEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create ServeMuxEngine: %v", err)
	}

	router := engine.(*ServeMuxEngine)

	// Register only GET handler
	router.HandleMethod(http.MethodGet, "/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("get response"))
	}))

	// Register only POST handler
	router.HandleMethod(http.MethodPost, "/api/data", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	}))

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedAllow  string
	}{
		{
			name:           "POST_to_GET_only_path",
			method:         http.MethodPost,
			path:           "/test",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedAllow:  "GET",
		},
		{
			name:           "PUT_to_GET_only_path",
			method:         http.MethodPut,
			path:           "/test",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedAllow:  "GET",
		},
		{
			name:           "GET_to_POST_only_path",
			method:         http.MethodGet,
			path:           "/api/data",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedAllow:  "POST",
		},
		{
			name:           "DELETE_to_POST_only_path",
			method:         http.MethodDelete,
			path:           "/api/data",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedAllow:  "POST",
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

			if allowHeader := w.Header().Get("Allow"); allowHeader != tt.expectedAllow {
				t.Errorf("Allow header = %v, want %v", allowHeader, tt.expectedAllow)
			}
		})
	}
}

func TestServeMuxEngine_HEADFallback(t *testing.T) {
	engine, err := NewServeMuxEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create ServeMuxEngine: %v", err)
	}

	router := engine.(*ServeMuxEngine)

	// Register GET handler
	router.HandleMethod(http.MethodGet, "/content", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", "12")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello world!"))
	}))

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectBody     bool
		expectHeaders  map[string]string
	}{
		{
			name:           "GET_request",
			method:         http.MethodGet,
			path:           "/content",
			expectedStatus: http.StatusOK,
			expectBody:     true,
			expectHeaders: map[string]string{
				"Content-Type":   "text/plain",
				"Content-Length": "12",
			},
		},
		{
			name:           "HEAD_request_fallback_to_GET",
			method:         http.MethodHead,
			path:           "/content",
			expectedStatus: http.StatusOK,
			expectBody:     false, // HEAD should not return body
			expectHeaders: map[string]string{
				"Content-Type":   "text/plain",
				"Content-Length": "12",
			},
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

			// Check headers
			for key, expectedValue := range tt.expectHeaders {
				if got := w.Header().Get(key); got != expectedValue {
					t.Errorf("Header %s = %v, want %v", key, got, expectedValue)
				}
			}

			// Check body presence
			body := w.Body.String()
			if tt.expectBody {
				if body != "hello world!" {
					t.Errorf("Expected body 'hello world!', got '%s'", body)
				}
			} else {
				if body != "" {
					t.Errorf("Expected no body for HEAD request, got '%s'", body)
				}
			}
		})
	}
}

func TestServeMuxEngine_ServeStatic(t *testing.T) {
	engine, err := NewServeMuxEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create ServeMuxEngine: %v", err)
	}

	router := engine.(*ServeMuxEngine)

	// Create temporary directory with test files
	tempDir := t.TempDir()
	testFile := tempDir + "/test.txt"
	err = os.WriteFile(testFile, []byte("static file content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create subdirectory with file
	subDir := tempDir + "/css"
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	cssFile := subDir + "/style.css"
	err = os.WriteFile(cssFile, []byte("body { color: red; }"), 0644)
	if err != nil {
		t.Fatalf("Failed to create CSS file: %v", err)
	}

	// Test ServeStatic
	router.ServeStatic("/static", http.Dir(tempDir))

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "static_file",
			path:           "/static/test.txt",
			expectedStatus: http.StatusOK,
			expectedBody:   "static file content",
		},
		{
			name:           "static_css_file",
			path:           "/static/css/style.css",
			expectedStatus: http.StatusOK,
			expectedBody:   "body { color: red; }",
		},
		{
			name:           "static_not_found",
			path:           "/static/nonexistent.txt",
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

			if tt.expectedBody != "" {
				body := w.Body.String()
				if body != tt.expectedBody {
					t.Errorf("Body = %v, want %v", body, tt.expectedBody)
				}
			}
		})
	}
}

func TestServeMuxEngine_ServeSPA(t *testing.T) {
	engine, err := NewServeMuxEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create ServeMuxEngine: %v", err)
	}

	router := engine.(*ServeMuxEngine)

	// Create temporary files for SPA
	tempDir := t.TempDir()
	indexFile := tempDir + "/index.html"
	err = os.WriteFile(indexFile, []byte("<!DOCTYPE html><html><body>SPA Content</body></html>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create index file: %v", err)
	}

	// Create some static assets
	jsFile := tempDir + "/app.js"
	err = os.WriteFile(jsFile, []byte("console.log('SPA loaded');"), 0644)
	if err != nil {
		t.Fatalf("Failed to create JS file: %v", err)
	}

	cssFile := tempDir + "/app.css"
	err = os.WriteFile(cssFile, []byte("body { font-family: Arial; }"), 0644)
	if err != nil {
		t.Fatalf("Failed to create CSS file: %v", err)
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
			expectedBody:   "<!DOCTYPE html><html><body>SPA Content</body></html>",
		},
		{
			name:           "spa_route",
			path:           "/app/dashboard",
			expectedStatus: http.StatusOK,
			expectedBody:   "<!DOCTYPE html><html><body>SPA Content</body></html>",
		},
		{
			name:           "spa_nested_route",
			path:           "/app/users/123",
			expectedStatus: http.StatusOK,
			expectedBody:   "<!DOCTYPE html><html><body>SPA Content</body></html>",
		},
		{
			name:           "static_js_file",
			path:           "/app/app.js",
			expectedStatus: http.StatusOK,
			expectedBody:   "console.log('SPA loaded');",
		},
		{
			name:           "static_css_file",
			path:           "/app/app.css",
			expectedStatus: http.StatusOK,
			expectedBody:   "body { font-family: Arial; }",
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

func TestServeMuxEngine_ServeReverseProxy(t *testing.T) {
	engine, err := NewServeMuxEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create ServeMuxEngine: %v", err)
	}

	router := engine.(*ServeMuxEngine)

	// Create test backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Backend", "test")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"path": "` + r.URL.Path + `", "method": "` + r.Method + `"}`))
	}))
	defer backend.Close()

	// Test ServeReverseProxy
	router.ServeReverseProxy("/api", backend.URL)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		expectedBody   string
		expectedHeader string
	}{
		{
			name:           "proxy_GET",
			method:         http.MethodGet,
			path:           "/api/users",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"path": "/users", "method": "GET"}`,
			expectedHeader: "application/json",
		},
		{
			name:           "proxy_POST",
			method:         http.MethodPost,
			path:           "/api/data",
			body:           "test data",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"path": "/data", "method": "POST"}`,
			expectedHeader: "application/json",
		},
		{
			name:           "proxy_nested_path",
			method:         http.MethodGet,
			path:           "/api/v1/users/123",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"path": "/v1/users/123", "method": "GET"}`,
			expectedHeader: "application/json",
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

			if ct := w.Header().Get("Content-Type"); ct != tt.expectedHeader {
				t.Errorf("Content-Type = %v, want %v", ct, tt.expectedHeader)
			}

			if backend := w.Header().Get("X-Backend"); backend != "test" {
				t.Errorf("X-Backend header = %v, want %v", backend, "test")
			}
		})
	}
}

func TestServeMuxEngine_MultipleMethodsOnSamePath(t *testing.T) {
	engine, err := NewServeMuxEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create ServeMuxEngine: %v", err)
	}

	router := engine.(*ServeMuxEngine)

	// Register multiple methods on same path
	router.HandleMethod(http.MethodGet, "/api/resource", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GET response"))
	}))

	router.HandleMethod(http.MethodPost, "/api/resource", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("POST response"))
	}))

	router.HandleMethod(http.MethodPut, "/api/resource", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("PUT response"))
	}))

	router.HandleMethod(http.MethodDelete, "/api/resource", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET_method",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   "GET response",
		},
		{
			name:           "POST_method",
			method:         http.MethodPost,
			expectedStatus: http.StatusCreated,
			expectedBody:   "POST response",
		},
		{
			name:           "PUT_method",
			method:         http.MethodPut,
			expectedStatus: http.StatusOK,
			expectedBody:   "PUT response",
		},
		{
			name:           "DELETE_method",
			method:         http.MethodDelete,
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name:           "PATCH_method_not_allowed",
			method:         http.MethodPatch,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/resource", nil)
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

			// Check Allow header for 405 responses
			if tt.expectedStatus == http.StatusMethodNotAllowed {
				allowHeader := w.Header().Get("Allow")
				expectedMethods := []string{"GET", "POST", "PUT", "DELETE"}
				for _, method := range expectedMethods {
					if !strings.Contains(allowHeader, method) {
						t.Errorf("Allow header '%s' should contain '%s'", allowHeader, method)
					}
				}
			}
		})
	}
}

func TestServeMuxEngine_Integration(t *testing.T) {
	engine, err := NewServeMuxEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create ServeMuxEngine: %v", err)
	}

	router := engine.(*ServeMuxEngine)

	// Create temp directory for files
	tempDir := t.TempDir()

	// Static file
	staticFile := tempDir + "/style.css"
	err = os.WriteFile(staticFile, []byte("body { margin: 0; }"), 0644)
	if err != nil {
		t.Fatalf("Failed to create CSS file: %v", err)
	}

	// SPA index
	indexFile := tempDir + "/index.html"
	err = os.WriteFile(indexFile, []byte("<!DOCTYPE html><html><body>App</body></html>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create HTML file: %v", err)
	}

	// Backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service": "backend", "path": "` + r.URL.Path + `"}`))
	}))
	defer backend.Close()

	// Register all types of routes
	router.HandleMethod(http.MethodGet, "/api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users": []}`))
	}))

	router.HandleMethod(http.MethodPost, "/api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Properly escape the JSON content
		escapedBody := strings.ReplaceAll(string(body), `"`, `\"`)
		w.Write([]byte(`{"created": true, "data": "` + escapedBody + `"}`))
	}))

	router.ServeStatic("/assets", http.Dir(tempDir))
	router.ServeSPA("/app", indexFile)
	router.ServeReverseProxy("/proxy", backend.URL)

	// Test comprehensive integration
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
			name:           "api_get_users",
			method:         http.MethodGet,
			path:           "/api/users",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"users": []}`,
			contentType:    "application/json",
		},
		{
			name:           "api_post_users",
			method:         http.MethodPost,
			path:           "/api/users",
			body:           `{"name": "John"}`,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"created": true, "data": "{\"name\": \"John\"}"}`,
			contentType:    "application/json",
		},
		{
			name:           "static_asset",
			method:         http.MethodGet,
			path:           "/assets/style.css",
			expectedStatus: http.StatusOK,
			expectedBody:   "body { margin: 0; }",
		},
		{
			name:           "spa_root",
			method:         http.MethodGet,
			path:           "/app",
			expectedStatus: http.StatusOK,
			expectedBody:   "<!DOCTYPE html><html><body>App</body></html>",
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
			expectedBody:   `{"service": "backend", "path": "/test"}`,
			contentType:    "application/json",
		},
		{
			name:           "method_not_allowed",
			method:         http.MethodDelete,
			path:           "/api/users",
			expectedStatus: http.StatusMethodNotAllowed,
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

			if tt.expectedBody != "" {
				body := w.Body.String()
				if body != tt.expectedBody {
					t.Errorf("Body = %v, want %v", body, tt.expectedBody)
				}
			}

			if tt.contentType != "" {
				if ct := w.Header().Get("Content-Type"); ct != tt.contentType {
					t.Errorf("Content-Type = %v, want %v", ct, tt.contentType)
				}
			}

			// Check Allow header for method not allowed
			if tt.expectedStatus == http.StatusMethodNotAllowed {
				allowHeader := w.Header().Get("Allow")
				if allowHeader == "" {
					t.Error("Allow header should be set for 405 responses")
				}
			}
		})
	}
}
