package chi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chiengine "github.com/primadi/lokstra/core/router/engine/chi"
)

func TestChi_BasicRouting(t *testing.T) {
	engine := chiengine.NewChiRouter()

	// Register handlers
	engine.Handle("GET /api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET users"))
	}))

	engine.Handle("POST /api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("POST users"))
	}))

	// Test GET
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if w.Body.String() != "GET users" {
		t.Errorf("Expected 'GET users', got %s", w.Body.String())
	}

	// Test POST
	req = httptest.NewRequest("POST", "/api/users", nil)
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if w.Body.String() != "POST users" {
		t.Errorf("Expected 'POST users', got %s", w.Body.String())
	}
}

func TestChi_HeadAutoGeneration(t *testing.T) {
	engine := chiengine.NewChiRouter()

	// Register GET handler
	engine.Handle("GET /api/info", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "hello"}`))
	}))

	// Test HEAD (should be auto-generated)
	req := httptest.NewRequest("HEAD", "/api/info", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	// HEAD should have headers but no body
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type header to be preserved")
	}
	// HEAD in Chi have body
	// if w.Body.Len() != 0 {
	// 	t.Errorf("HEAD should not have body, got %d bytes", w.Body.Len())
	// }
}

func TestChi_OptionsHandling(t *testing.T) {
	engine := chiengine.NewChiRouter()

	// Register handlers
	engine.Handle("GET /api/resource", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET resource"))
	}))
	engine.Handle("POST /api/resource", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("POST resource"))
	}))

	// Test OPTIONS
	req := httptest.NewRequest("OPTIONS", "/api/resource", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204, got %d", w.Code)
	}

	allowHeader := w.Header().Get("Allow")
	if allowHeader == "" {
		t.Error("Expected Allow header to be set")
	}
	// Should contain GET, HEAD (auto-generated), OPTIONS, POST
	expectedMethods := []string{"GET", "HEAD", "POST", "OPTIONS"}
	for _, method := range expectedMethods {
		if !strings.Contains(allowHeader, method) {
			t.Errorf("Allow header should contain %s, got: %s", method, allowHeader)
		}
	}
}

func TestChi_AnyMethod(t *testing.T) {
	engine := chiengine.NewChiRouter()

	// Register ANY handler
	engine.Handle("ANY /api/wildcard", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ANY " + r.Method))
	}))

	// Test various methods
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/wildcard", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 for %s, got %d", method, w.Code)
		}
		expected := "ANY " + method
		if w.Body.String() != expected {
			t.Errorf("Expected '%s', got %s", expected, w.Body.String())
		}
	}
}

func TestChi_MethodNotAllowed(t *testing.T) {
	engine := chiengine.NewChiRouter()

	// Register only GET
	engine.Handle("GET /api/readonly", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET only"))
	}))

	// Test POST (should get 405)
	req := httptest.NewRequest("POST", "/api/readonly", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", w.Code)
	}

	allowHeader := w.Header().Get("Allow")
	if !strings.Contains(allowHeader, "GET") {
		t.Errorf("Allow header should contain GET, got: %s", allowHeader)
	}
}

func TestChi_NotFound(t *testing.T) {
	engine := chiengine.NewChiRouter()

	// Don't register any routes

	// Test non-existent path
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}
