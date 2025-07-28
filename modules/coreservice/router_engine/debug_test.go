package router_engine

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestHttpRouterDefaultBehavior(t *testing.T) {
	// Test plain httprouter behavior
	router := httprouter.New()

	// Register only GET
	router.GET("/test", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("get response"))
	})

	// Test POST to GET-only route
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	t.Logf("Plain httprouter status for POST to GET-only route: %d", w.Code)
	t.Logf("HandleMethodNotAllowed setting: %v", router.HandleMethodNotAllowed)
}

func TestOurHttpRouterBehavior(t *testing.T) {
	// Test our HttpRouterEngine behavior
	engine, err := NewHttpRouterEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create HttpRouterEngine: %v", err)
	}

	httpRouter := engine.(*HttpRouterEngine)

	// Register only GET
	httpRouter.HandleMethod(http.MethodGet, "/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("get response"))
	}))

	// Test POST to GET-only route
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	httpRouter.ServeHTTP(w, req)

	t.Logf("Our HttpRouterEngine status for POST to GET-only route: %d", w.Code)
	t.Logf("HandleMethodNotAllowed setting: %v", httpRouter.hr.HandleMethodNotAllowed)
	t.Logf("Has ServeMux fallback: %v", httpRouter.sm != nil)
}
