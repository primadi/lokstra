package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
)

// Benchmark: Fast path vs reflection for common patterns

func BenchmarkHandler_ContextError_FastPath(b *testing.B) {
	// This should use fast path (Tier 0 - zero cost)
	handler := func(c *request.Context) error {
		return c.Api.Ok(map[string]string{"test": "data"})
	}

	r := New("bench")
	r.GET("/test", handler)
	r.Build()

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_ContextAnyError_FastPath(b *testing.B) {
	// This should use fast path (Tier 1)
	handler := func(c *request.Context) (any, error) {
		return map[string]string{"test": "data"}, nil
	}

	r := New("bench")
	r.GET("/test", handler)
	r.Build()

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_ContextResponseError_FastPath(b *testing.B) {
	// This should use fast path (Tier 1)
	handler := func(c *request.Context) (*response.Response, error) {
		resp := response.NewResponse()
		resp.Json(map[string]string{"test": "data"})
		return resp, nil
	}

	r := New("bench")
	r.GET("/test", handler)
	r.Build()

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_ContextApiHelperError_FastPath(b *testing.B) {
	// This should use fast path (Tier 1)
	handler := func(c *request.Context) (*response.ApiHelper, error) {
		api := response.NewApiHelper()
		api.Ok(map[string]string{"test": "data"})
		return api, nil
	}

	r := New("bench")
	r.GET("/test", handler)
	r.Build()

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_ContextAny_FastPath(b *testing.B) {
	// This should use fast path (Tier 1) - no error variant
	handler := func(c *request.Context) any {
		return map[string]string{"test": "data"}
	}

	r := New("bench")
	r.GET("/test", handler)
	r.Build()

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_NoContextAnyError_FastPath(b *testing.B) {
	// This should use fast path (Tier 1)
	handler := func() (any, error) {
		return map[string]string{"test": "data"}, nil
	}

	r := New("bench")
	r.GET("/test", handler)
	r.Build()

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_StructParam_Reflection(b *testing.B) {
	// This uses reflection (Tier 2) - no fast path available
	type params struct {
		ID int `path:"id"`
	}

	handler := func(p *params) (any, error) {
		return map[string]any{"id": p.ID}, nil
	}

	r := New("bench")
	r.GET("/test/{id}", handler)
	r.Build()

	req := httptest.NewRequest("GET", "/test/123", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_HttpHandlerFunc(b *testing.B) {
	// Standard HTTP handler (Tier 1)
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test":"data"}`))
	}

	r := New("bench")
	r.GET("/test", handler)
	r.Build()

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// Comparison benchmarks - measure overhead of different tiers

func BenchmarkOverhead_Tier0_Direct(b *testing.B) {
	handler := func(c *request.Context) error {
		return nil
	}

	r := New("bench")
	r.GET("/test", handler)
	r.Build()

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkOverhead_Tier1_Wrapper(b *testing.B) {
	handler := func(c *request.Context) (any, error) {
		return nil, nil
	}

	r := New("bench")
	r.GET("/test", handler)
	r.Build()

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkOverhead_Tier2_Reflection(b *testing.B) {
	type params struct{}

	handler := func(p *params) (any, error) {
		return nil, nil
	}

	r := New("bench")
	r.GET("/test", handler)
	r.Build()

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
