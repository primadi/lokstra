package engine_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/router/engine"
	chiengine "github.com/primadi/lokstra/core/router/engine/chi"
)

// Benchmark handlers
var (
	simpleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	pathValueHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id))
	})

	wildcardHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.PathValue("path")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(path))
	})
)

// setupRouters creates routers with common test routes
func setupRouters() (serveMux, serveMuxPlus, chiRouter engine.RouterEngine) {
	// ServeMux
	sm := engine.NewServeMux()
	sm.Handle("GET /", simpleHandler)
	sm.Handle("GET /users", simpleHandler)
	sm.Handle("GET /users/{id}", pathValueHandler)
	sm.Handle("POST /users", simpleHandler)
	sm.Handle("PUT /users/{id}", pathValueHandler)
	sm.Handle("DELETE /users/{id}", pathValueHandler)
	sm.Handle("GET /api/{path...}", wildcardHandler)

	// ServeMuxPlus
	smp := engine.NewServeMuxPlus()
	smp.Handle("GET /", simpleHandler)
	smp.Handle("GET /users", simpleHandler)
	smp.Handle("GET /users/{id}", pathValueHandler)
	smp.Handle("POST /users", simpleHandler)
	smp.Handle("PUT /users/{id}", pathValueHandler)
	smp.Handle("DELETE /users/{id}", pathValueHandler)
	smp.Handle("GET /api/{path...}", wildcardHandler)

	// ChiRouter
	chi := chiengine.NewChiRouter()
	chi.Handle("GET /", simpleHandler)
	chi.Handle("GET /users", simpleHandler)
	chi.Handle("GET /users/{id}", pathValueHandler)
	chi.Handle("POST /users", simpleHandler)
	chi.Handle("PUT /users/{id}", pathValueHandler)
	chi.Handle("DELETE /users/{id}", pathValueHandler)
	chi.Handle("GET /api/{path...}", wildcardHandler)

	return sm, smp, chi
}

// Benchmark static routes (no path parameters)
func BenchmarkStaticRoute_ServeMux(b *testing.B) {
	sm, _, _ := setupRouters()
	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sm.ServeHTTP(w, req)
	}
}

func BenchmarkStaticRoute_ServeMuxPlus(b *testing.B) {
	_, smp, _ := setupRouters()
	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		smp.ServeHTTP(w, req)
	}
}

func BenchmarkStaticRoute_ChiRouter(b *testing.B) {
	_, _, chi := setupRouters()
	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		chi.ServeHTTP(w, req)
	}
}

// Benchmark routes with path parameters
func BenchmarkPathParam_ServeMux(b *testing.B) {
	sm, _, _ := setupRouters()
	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sm.ServeHTTP(w, req)
	}
}

func BenchmarkPathParam_ServeMuxPlus(b *testing.B) {
	_, smp, _ := setupRouters()
	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		smp.ServeHTTP(w, req)
	}
}

func BenchmarkPathParam_ChiRouter(b *testing.B) {
	_, _, chi := setupRouters()
	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		chi.ServeHTTP(w, req)
	}
}

// Benchmark wildcard routes
func BenchmarkWildcard_ServeMux(b *testing.B) {
	sm, _, _ := setupRouters()
	req := httptest.NewRequest("GET", "/api/v1/users/123/posts", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sm.ServeHTTP(w, req)
	}
}

func BenchmarkWildcard_ServeMuxPlus(b *testing.B) {
	_, smp, _ := setupRouters()
	req := httptest.NewRequest("GET", "/api/v1/users/123/posts", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		smp.ServeHTTP(w, req)
	}
}

func BenchmarkWildcard_ChiRouter(b *testing.B) {
	_, _, chi := setupRouters()
	req := httptest.NewRequest("GET", "/api/v1/users/123/posts", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		chi.ServeHTTP(w, req)
	}
}

// Benchmark OPTIONS requests (auto-generated)
func BenchmarkOPTIONS_ServeMux(b *testing.B) {
	sm, _, _ := setupRouters()
	req := httptest.NewRequest("OPTIONS", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sm.ServeHTTP(w, req)
	}
}

func BenchmarkOPTIONS_ServeMuxPlus(b *testing.B) {
	_, smp, _ := setupRouters()
	req := httptest.NewRequest("OPTIONS", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		smp.ServeHTTP(w, req)
	}
}

func BenchmarkOPTIONS_ChiRouter(b *testing.B) {
	_, _, chi := setupRouters()
	req := httptest.NewRequest("OPTIONS", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		chi.ServeHTTP(w, req)
	}
}

// Benchmark mixed routes (simulate real-world scenario)
func BenchmarkMixedRoutes_ServeMux(b *testing.B) {
	sm, _, _ := setupRouters()
	requests := []*http.Request{
		httptest.NewRequest("GET", "/", nil),
		httptest.NewRequest("GET", "/users", nil),
		httptest.NewRequest("GET", "/users/123", nil),
		httptest.NewRequest("POST", "/users", nil),
		httptest.NewRequest("PUT", "/users/456", nil),
		httptest.NewRequest("DELETE", "/users/789", nil),
		httptest.NewRequest("GET", "/api/v1/resources", nil),
		httptest.NewRequest("OPTIONS", "/users/123", nil),
	}
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := requests[i%len(requests)]
		sm.ServeHTTP(w, req)
	}
}

func BenchmarkMixedRoutes_ServeMuxPlus(b *testing.B) {
	_, smp, _ := setupRouters()
	requests := []*http.Request{
		httptest.NewRequest("GET", "/", nil),
		httptest.NewRequest("GET", "/users", nil),
		httptest.NewRequest("GET", "/users/123", nil),
		httptest.NewRequest("POST", "/users", nil),
		httptest.NewRequest("PUT", "/users/456", nil),
		httptest.NewRequest("DELETE", "/users/789", nil),
		httptest.NewRequest("GET", "/api/v1/resources", nil),
		httptest.NewRequest("OPTIONS", "/users/123", nil),
	}
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := requests[i%len(requests)]
		smp.ServeHTTP(w, req)
	}
}

func BenchmarkMixedRoutes_ChiRouter(b *testing.B) {
	_, _, chi := setupRouters()
	requests := []*http.Request{
		httptest.NewRequest("GET", "/", nil),
		httptest.NewRequest("GET", "/users", nil),
		httptest.NewRequest("GET", "/users/123", nil),
		httptest.NewRequest("POST", "/users", nil),
		httptest.NewRequest("PUT", "/users/456", nil),
		httptest.NewRequest("DELETE", "/users/789", nil),
		httptest.NewRequest("GET", "/api/v1/resources", nil),
		httptest.NewRequest("OPTIONS", "/users/123", nil),
	}
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := requests[i%len(requests)]
		chi.ServeHTTP(w, req)
	}
}

// Benchmark large route table (100 routes)
func setupLargeRouters() (serveMux, serveMuxPlus, chiRouter engine.RouterEngine) {
	sm := engine.NewServeMux()
	smp := engine.NewServeMuxPlus()
	chi := chiengine.NewChiRouter()

	for i := 0; i < 100; i++ {
		pattern := fmt.Sprintf("GET /resource%d/{id}", i)
		sm.Handle(pattern, pathValueHandler)
		smp.Handle(pattern, pathValueHandler)
		chi.Handle(pattern, pathValueHandler)
	}

	return sm, smp, chi
}

func BenchmarkLargeRouteTable_ServeMux(b *testing.B) {
	sm, _, _ := setupLargeRouters()
	// Test middle route
	req := httptest.NewRequest("GET", "/resource50/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sm.ServeHTTP(w, req)
	}
}

func BenchmarkLargeRouteTable_ServeMuxPlus(b *testing.B) {
	_, smp, _ := setupLargeRouters()
	// Test middle route
	req := httptest.NewRequest("GET", "/resource50/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		smp.ServeHTTP(w, req)
	}
}

func BenchmarkLargeRouteTable_ChiRouter(b *testing.B) {
	_, _, chi := setupLargeRouters()
	// Test middle route
	req := httptest.NewRequest("GET", "/resource50/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		chi.ServeHTTP(w, req)
	}
}

// Benchmark router creation overhead
func BenchmarkRouterCreation_ServeMux(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = engine.NewServeMux()
	}
}

func BenchmarkRouterCreation_ServeMuxPlus(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = engine.NewServeMuxPlus()
	}
}

func BenchmarkRouterCreation_ChiRouter(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = chiengine.NewChiRouter()
	}
}

// Benchmark route registration
func BenchmarkRouteRegistration_ServeMux(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sm := engine.NewServeMux()
		sm.Handle("GET /users/{id}", pathValueHandler)
	}
}

func BenchmarkRouteRegistration_ServeMuxPlus(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		smp := engine.NewServeMuxPlus()
		smp.Handle("GET /users/{id}", pathValueHandler)
	}
}

func BenchmarkRouteRegistration_ChiRouter(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		chi := chiengine.NewChiRouter()
		chi.Handle("GET /users/{id}", pathValueHandler)
	}
}

// Benchmark concurrent requests (parallel)
func BenchmarkParallel_ServeMux(b *testing.B) {
	sm, _, _ := setupRouters()

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/users/123", nil)
		w := httptest.NewRecorder()
		for pb.Next() {
			sm.ServeHTTP(w, req)
		}
	})
}

func BenchmarkParallel_ServeMuxPlus(b *testing.B) {
	_, smp, _ := setupRouters()

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/users/123", nil)
		w := httptest.NewRecorder()
		for pb.Next() {
			smp.ServeHTTP(w, req)
		}
	})
}

func BenchmarkParallel_ChiRouter(b *testing.B) {
	_, _, chi := setupRouters()

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/users/123", nil)
		w := httptest.NewRecorder()
		for pb.Next() {
			chi.ServeHTTP(w, req)
		}
	})
}
