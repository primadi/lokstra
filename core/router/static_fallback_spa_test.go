package router_test

import (
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/primadi/lokstra/core/router"
)

func TestStaticFallback_RawHandler_SPA_Mode(t *testing.T) {
	// Create a mock filesystem for SPA testing
	mockFS := fstest.MapFS{
		"index.html": {
			Data: []byte(`<!DOCTYPE html>
<html>
<head><title>SPA App</title></head>
<body><div id="app">SPA Application</div></body>
</html>`),
		},
		"assets/app.js": {
			Data: []byte(`console.log('SPA app loaded');`),
		},
		"assets/style.css": {
			Data: []byte(`body { font-family: Arial; }`),
		},
		"api/data.json": {
			Data: []byte(`{"message": "API response"}`),
		},
		"favicon.ico": {
			Data: []byte("fake-favicon-data"),
		},
	}

	// Create StaticFallback with SPA mode enabled
	fallback := router.NewStaticFallback(mockFS)
	handler := fallback.RawHandler(true) // SPA mode = true

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
		description    string
	}{
		{
			name:           "root_path",
			path:           "/",
			expectedStatus: 200,
			expectedBody:   "SPA Application", // Should serve index.html
			description:    "Root path should serve index.html",
		},
		{
			name:           "explicit_index",
			path:           "/index.html",
			expectedStatus: 200,
			expectedBody:   "SPA Application",
			description:    "Explicit index.html should serve index.html",
		},
		{
			name:           "spa_route_dashboard",
			path:           "/dashboard",
			expectedStatus: 200,
			expectedBody:   "SPA Application", // Should fallback to index.html
			description:    "SPA route /dashboard should fallback to index.html",
		},
		{
			name:           "spa_route_dashboard_admin",
			path:           "/dashboard/admin",
			expectedStatus: 200,
			expectedBody:   "SPA Application", // Should fallback to index.html
			description:    "SPA route /dashboard/admin should fallback to index.html",
		},
		{
			name:           "existing_asset_js",
			path:           "/assets/app.js",
			expectedStatus: 200,
			expectedBody:   "console.log('SPA app loaded');",
			description:    "Existing JS asset should be served directly",
		},
		{
			name:           "existing_asset_css",
			path:           "/assets/style.css",
			expectedStatus: 200,
			expectedBody:   "body { font-family: Arial; }",
			description:    "Existing CSS asset should be served directly",
		},
		{
			name:           "existing_api_endpoint",
			path:           "/api/data.json",
			expectedStatus: 200,
			expectedBody:   `{"message": "API response"}`,
			description:    "Existing API file should be served directly",
		},
		{
			name:           "existing_favicon",
			path:           "/favicon.ico",
			expectedStatus: 200,
			expectedBody:   "fake-favicon-data",
			description:    "Existing favicon should be served directly",
		},
		{
			name:           "missing_asset_with_extension",
			path:           "/dashboard/assets/image.jpg",
			expectedStatus: 404,
			expectedBody:   "404 page not found",
			description:    "Missing asset with extension should return 404, not fallback",
		},
		{
			name:           "missing_asset_js",
			path:           "/nonexistent/script.js",
			expectedStatus: 404,
			expectedBody:   "404 page not found",
			description:    "Missing JS file should return 404, not fallback",
		},
		{
			name:           "missing_asset_css",
			path:           "/missing/style.css",
			expectedStatus: 404,
			expectedBody:   "404 page not found",
			description:    "Missing CSS file should return 404, not fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()

			// Execute handler
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d - %s", tt.expectedStatus, rr.Code, tt.description)
			}

			// Check response body contains expected content
			body := rr.Body.String()
			if !containsString(body, tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', got '%s' - %s", tt.expectedBody, body, tt.description)
			}

			// Additional checks for successful responses
			if tt.expectedStatus == 200 {
				// Check Content-Type is set for known file types
				if isAssetFile(tt.path) {
					contentType := rr.Header().Get("Content-Type")
					if contentType == "" {
						t.Errorf("Expected Content-Type header for asset file %s", tt.path)
					}
				}
			}

			t.Logf("✅ %s: %d - %s", tt.path, rr.Code, tt.description)
		})
	}
}

func TestStaticFallback_SPA_vs_Static_Mode(t *testing.T) {
	// Test the difference between SPA mode (true) and static mode (false)
	mockFS := fstest.MapFS{
		"index.html": {Data: []byte("<html>Index</html>")},
		"about.html": {Data: []byte("<html>About</html>")},
	}

	fallback := router.NewStaticFallback(mockFS)

	tests := []struct {
		name        string
		path        string
		spaMode     bool
		expectCode  int
		expectBody  string
		description string
	}{
		{
			name:        "static_mode_missing_route",
			path:        "/dashboard",
			spaMode:     false,
			expectCode:  404,
			expectBody:  "404 page not found",
			description: "Static mode should return 404 for missing routes",
		},
		{
			name:        "spa_mode_missing_route",
			path:        "/dashboard",
			spaMode:     true,
			expectCode:  200,
			expectBody:  "<html>Index</html>",
			description: "SPA mode should fallback to index.html for missing routes",
		},
		{
			name:        "both_modes_existing_file",
			path:        "/about.html",
			spaMode:     true, // Test with SPA mode, should still serve the file directly
			expectCode:  200,
			expectBody:  "<html>About</html>",
			description: "Both modes should serve existing files directly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := fallback.RawHandler(tt.spaMode)

			req := httptest.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectCode {
				t.Errorf("Expected status %d, got %d - %s", tt.expectCode, rr.Code, tt.description)
			}

			body := rr.Body.String()
			if !containsString(body, tt.expectBody) {
				t.Errorf("Expected body to contain '%s', got '%s' - %s", tt.expectBody, body, tt.description)
			}

			t.Logf("✅ %s (SPA:%v): %d - %s", tt.path, tt.spaMode, rr.Code, tt.description)
		})
	}
}

func TestStaticFallback_MultipleSources_SPA(t *testing.T) {
	// Test SPA mode with multiple fallback sources
	primaryFS := fstest.MapFS{
		"index.html":    {Data: []byte("<html>Primary SPA</html>")},
		"assets/app.js": {Data: []byte("console.log('primary');")},
	}

	fallbackFS := fstest.MapFS{
		"assets/fallback.js": {Data: []byte("console.log('fallback');")},
		"api/data.json":      {Data: []byte(`{"source": "fallback"}`)},
	}

	fallback := router.NewStaticFallback(primaryFS, fallbackFS)
	handler := fallback.RawHandler(true) // SPA mode

	tests := []struct {
		name        string
		path        string
		expectCode  int
		expectBody  string
		description string
	}{
		{
			name:        "spa_fallback_from_primary",
			path:        "/dashboard",
			expectCode:  200,
			expectBody:  "<html>Primary SPA</html>",
			description: "SPA route should use index.html from primary source",
		},
		{
			name:        "asset_from_primary",
			path:        "/assets/app.js",
			expectCode:  200,
			expectBody:  "console.log('primary');",
			description: "Should serve asset from primary source",
		},
		{
			name:        "asset_from_fallback",
			path:        "/assets/fallback.js",
			expectCode:  200,
			expectBody:  "console.log('fallback');",
			description: "Should serve asset from fallback source",
		},
		{
			name:        "missing_asset",
			path:        "/assets/missing.jpg",
			expectCode:  404,
			expectBody:  "404 page not found",
			description: "Missing asset should return 404 even in SPA mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectCode {
				t.Errorf("Expected status %d, got %d - %s", tt.expectCode, rr.Code, tt.description)
			}

			body := rr.Body.String()
			if !containsString(body, tt.expectBody) {
				t.Errorf("Expected body to contain '%s', got '%s' - %s", tt.expectBody, body, tt.description)
			}
		})
	}
}

// Helper functions
func containsString(haystack, needle string) bool {
	return len(needle) == 0 || len(haystack) >= len(needle) &&
		(haystack == needle || strings.Contains(haystack, needle))
}

func isAssetFile(path string) bool {
	extensions := []string{".js", ".css", ".jpg", ".jpeg", ".png", ".gif", ".ico", ".svg", ".json"}
	for _, ext := range extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
