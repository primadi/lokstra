package router_engine

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConvertToServeMuxParamPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple_parameter",
			input:    "/users/:id",
			expected: "/users/{id}",
		},
		{
			name:     "multiple_parameters",
			input:    "/users/:id/posts/:postId",
			expected: "/users/{id}/posts/{postId}",
		},
		{
			name:     "no_parameters",
			input:    "/api/users",
			expected: "/api/users",
		},
		{
			name:     "wildcard_parameter",
			input:    "/files/*filepath",
			expected: "/files/{filepath...}",
		},
		{
			name:     "parameter_at_end",
			input:    "/api/v1/resource/:resourceId",
			expected: "/api/v1/resource/{resourceId}",
		},
		{
			name:     "parameter_at_start",
			input:    "/:category/items",
			expected: "/{category}/items",
		},
		{
			name:     "multiple_wildcards",
			input:    "/static/*filepath/assets/*asset",
			expected: "/static/{filepath...}/assets/{asset...}",
		},
		{
			name:     "mixed_parameters",
			input:    "/users/:id/files/*filepath",
			expected: "/users/{id}/files/{filepath...}",
		},
		{
			name:     "empty_path",
			input:    "",
			expected: "",
		},
		{
			name:     "root_path",
			input:    "/",
			expected: "/",
		},
		{
			name:     "complex_path",
			input:    "/api/v1/users/:userId/posts/:postId/comments/:commentId",
			expected: "/api/v1/users/{userId}/posts/{postId}/comments/{commentId}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToServeMuxParamPath(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertToServeMuxParamPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHeadFallbackWriter(t *testing.T) {
	tests := []struct {
		name           string
		writeData      []byte
		expectedStatus int
		expectBody     bool
		headers        map[string]string
	}{
		{
			name:           "simple_write",
			writeData:      []byte("hello world"),
			expectedStatus: http.StatusOK,
			expectBody:     false,
			headers: map[string]string{
				"Content-Type": "text/plain",
			},
		},
		{
			name:           "json_response",
			writeData:      []byte(`{"message": "test"}`),
			expectedStatus: http.StatusOK,
			expectBody:     false,
			headers: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name:           "empty_write",
			writeData:      []byte(""),
			expectedStatus: http.StatusOK,
			expectBody:     false,
		},
		{
			name:           "large_content",
			writeData:      []byte(strings.Repeat("a", 1000)),
			expectedStatus: http.StatusOK,
			expectBody:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a regular ResponseWriter
			w := httptest.NewRecorder()

			// Set headers if provided
			for key, value := range tt.headers {
				w.Header().Set(key, value)
			}

			// Create head fallback writer
			headWriter := &headFallbackWriter{ResponseWriter: w}

			// Write status (if not 200)
			if tt.expectedStatus != http.StatusOK {
				headWriter.WriteHeader(tt.expectedStatus)
			}

			// Write data
			n, err := headWriter.Write(tt.writeData)
			if err != nil {
				t.Fatalf("Write failed: %v", err)
			}

			// Should return the length of data written
			if n != len(tt.writeData) {
				t.Errorf("Write returned %d bytes, expected %d", n, len(tt.writeData))
			}

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedStatus)
			}

			// Check that no body is written (HEAD should not have body)
			body := w.Body.String()
			if tt.expectBody {
				if body != string(tt.writeData) {
					t.Errorf("Expected body %q, got %q", string(tt.writeData), body)
				}
			} else {
				if body != "" {
					t.Errorf("Expected no body for HEAD request, got %q", body)
				}
			}

			// Check headers are preserved
			for key, expectedValue := range tt.headers {
				if got := w.Header().Get(key); got != expectedValue {
					t.Errorf("Header %s = %v, want %v", key, got, expectedValue)
				}
			}
		})
	}
}

func TestHeadFallbackWriter_WriteHeader(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedStatus int
	}{
		{
			name:           "status_ok",
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "status_created",
			statusCode:     http.StatusCreated,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "status_not_found",
			statusCode:     http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "status_internal_error",
			statusCode:     http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			headWriter := &headFallbackWriter{ResponseWriter: w}

			headWriter.WriteHeader(tt.statusCode)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestHeadFallbackWriter_Header(t *testing.T) {
	w := httptest.NewRecorder()
	headWriter := &headFallbackWriter{ResponseWriter: w}

	// Test that Header() returns the underlying header
	header := headWriter.Header()
	if header == nil {
		t.Error("Header() returned nil")
	}

	// Test setting and getting headers
	header.Set("Content-Type", "application/json")
	header.Set("X-Test", "test-value")

	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %v, want %v", got, "application/json")
	}

	if got := w.Header().Get("X-Test"); got != "test-value" {
		t.Errorf("X-Test = %v, want %v", got, "test-value")
	}
}

func TestCleanPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple_path",
			input:    "/api",
			expected: "/api",
		},
		{
			name:     "path_with_trailing_slash",
			input:    "/api/",
			expected: "/api",
		},
		{
			name:     "root_path",
			input:    "/",
			expected: "/",
		},
		{
			name:     "empty_path",
			input:    "",
			expected: "/",
		},
		{
			name:     "complex_path",
			input:    "/api/v1/users",
			expected: "/api/v1/users",
		},
		{
			name:     "path_with_multiple_slashes",
			input:    "/api//v1/",
			expected: "/api//v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("cleanPrefix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHelperFunctionIntegration(t *testing.T) {
	tests := []struct {
		name         string
		originalPath string
		testPath     string
		expectMatch  bool
	}{
		{
			name:         "simple_parameter_conversion",
			originalPath: "/users/:id",
			testPath:     "/users/123",
			expectMatch:  true,
		},
		{
			name:         "wildcard_conversion",
			originalPath: "/files/*filepath",
			testPath:     "/files/documents/test.txt",
			expectMatch:  true,
		},
		{
			name:         "multiple_parameters",
			originalPath: "/users/:id/posts/:postId",
			testPath:     "/users/123/posts/456",
			expectMatch:  true,
		},
		{
			name:         "no_match",
			originalPath: "/users/:id",
			testPath:     "/posts/123",
			expectMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert path using helper function
			convertedPath := ConvertToServeMuxParamPath(tt.originalPath)

			// Create a test ServeMux to verify the conversion works
			mux := http.NewServeMux()

			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("matched"))
			})

			mux.Handle(convertedPath, testHandler)

			// Test with converted path
			req := httptest.NewRequest(http.MethodGet, tt.testPath, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if tt.expectMatch {
				if !handlerCalled {
					t.Errorf("Expected handler to be called for path %s -> %s with test %s",
						tt.originalPath, convertedPath, tt.testPath)
				}
				if w.Code != http.StatusOK {
					t.Errorf("Expected status 200, got %d", w.Code)
				}
			} else {
				if handlerCalled {
					t.Errorf("Expected handler NOT to be called for path %s -> %s with test %s",
						tt.originalPath, convertedPath, tt.testPath)
				}
				if w.Code != http.StatusNotFound {
					t.Errorf("Expected status 404, got %d", w.Code)
				}
			}
		})
	}
}

func TestHeadFallbackWriter_MultipleWrites(t *testing.T) {
	w := httptest.NewRecorder()
	headWriter := &headFallbackWriter{ResponseWriter: w}

	// Set some headers
	headWriter.Header().Set("Content-Type", "text/plain")
	headWriter.Header().Set("Content-Length", "11")

	// Multiple writes should all be discarded
	data1 := []byte("hello ")
	data2 := []byte("world")

	n1, err1 := headWriter.Write(data1)
	if err1 != nil {
		t.Fatalf("First write failed: %v", err1)
	}
	if n1 != len(data1) {
		t.Errorf("First write returned %d bytes, expected %d", n1, len(data1))
	}

	n2, err2 := headWriter.Write(data2)
	if err2 != nil {
		t.Fatalf("Second write failed: %v", err2)
	}
	if n2 != len(data2) {
		t.Errorf("Second write returned %d bytes, expected %d", n2, len(data2))
	}

	// Check that no body was written
	body := w.Body.String()
	if body != "" {
		t.Errorf("Expected no body, got %q", body)
	}

	// Check that headers are preserved
	if ct := w.Header().Get("Content-Type"); ct != "text/plain" {
		t.Errorf("Content-Type = %v, want %v", ct, "text/plain")
	}

	if cl := w.Header().Get("Content-Length"); cl != "11" {
		t.Errorf("Content-Length = %v, want %v", cl, "11")
	}
}

func TestConvertPathEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "parameter_with_extension",
			input:    "/files/:filename.txt",
			expected: "/files/{filename}.txt",
		},
		{
			name:     "parameter_with_dash",
			input:    "/api/:user-id",
			expected: "/api/{user}-id",
		},
		{
			name:     "parameter_with_underscore",
			input:    "/api/:user_id",
			expected: "/api/{user_id}",
		},
		{
			name:     "wildcard_with_extension",
			input:    "/assets/*file.css",
			expected: "/assets/{file...}.css",
		},
		{
			name:     "consecutive_parameters",
			input:    "/api/:version/:id",
			expected: "/api/{version}/{id}",
		},
		{
			name:     "parameter_in_query_position",
			input:    "/search/:query?",
			expected: "/search/{query}?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToServeMuxParamPath(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertToServeMuxParamPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
