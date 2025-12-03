package cors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/middleware/cors"
)

func TestCorsMiddleware_AllOrigins(t *testing.T) {
	h := cors.Middleware("*")

	// Test GET with Origin
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	ctx := request.NewContext(w, req, nil)
	h(ctx)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Errorf("Allow-Origin header not set correctly: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Allow-Credentials header not set correctly")
	}
}

func TestCorsMiddleware_AllowedOrigin(t *testing.T) {
	h := cors.Middleware("http://allowed.com")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://allowed.com")
	w := httptest.NewRecorder()
	ctx := request.NewContext(w, req, nil)
	h(ctx)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://allowed.com" {
		t.Errorf("Allow-Origin header not set correctly for allowed origin")
	}
}

func TestCorsMiddleware_DisallowedOrigin(t *testing.T) {
	h := cors.Middleware("http://allowed.com")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://forbidden.com")
	w := httptest.NewRecorder()
	ctx := request.NewContext(w, req, nil)
	h(ctx)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403 for forbidden origin, got %d", w.Code)
	}
}

func TestCorsMiddleware_OPTIONS(t *testing.T) {
	h := cors.Middleware("*")

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header, Authorization")
	w := httptest.NewRecorder()
	ctx := request.NewContext(w, req, nil)
	h(ctx)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Errorf("Allow-Origin header not set correctly on OPTIONS")
	}
	if w.Header().Get("Access-Control-Allow-Headers") != "X-Custom-Header, Authorization" {
		t.Errorf("Allow-Headers header not set correctly on OPTIONS: %s", w.Header().Get("Access-Control-Allow-Headers"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("Allow-Methods header not set correctly on OPTIONS: %s", w.Header().Get("Access-Control-Allow-Methods"))
	}
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204 for OPTIONS, got %d", w.Code)
	}
}
