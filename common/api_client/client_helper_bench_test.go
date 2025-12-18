package api_client_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/primadi/lokstra/common/api_client"
	"github.com/primadi/lokstra/common/cast"
	"github.com/primadi/lokstra/core/response/api_formatter"
)

type BenchResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Helper to create a test ClientRouter pointing to a URL
func newTestClientRouter(url string) *api_client.ClientRouter {
	return &api_client.ClientRouter{
		RouterName: "test-router",
		ServerName: "test-server",
		FullURL:    url,
		IsLocal:    false,
	}
}

// FetchAndCastUnoptimized - Original version WITHOUT caching for comparison
func FetchAndCastUnoptimized[T any](client *api_client.ClientRouter, path string, opts ...api_client.FetchOption) (T, error) {
	cfg := &api_client.FetchConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	method := cfg.Method
	if method == "" {
		method = "GET"
	}

	var zero T

	resp, err := client.Method(method, path, cfg.Body, cfg.Headers)
	if err != nil {
		return zero, fmt.Errorf("failed to fetch: %v", err)
	}

	formatter := cfg.Formatter
	if formatter == nil {
		formatter = api_formatter.GetGlobalFormatter()
	}

	clientResp := &api_formatter.ClientResponse{}
	if err := formatter.ParseClientResponse(resp, clientResp); err != nil {
		return zero, fmt.Errorf("failed to parse response: %v", err)
	}

	if cfg.CustomFunc != nil {
		customResult, err := cfg.CustomFunc(resp, clientResp)
		if err != nil {
			return zero, err
		}

		if customResult != nil {
			if result, ok := customResult.(T); ok {
				return result, nil
			}

			// NO CACHING - repeated reflection every call
			var result T
			resultType := reflect.TypeOf((*T)(nil)).Elem()

			if resultType.Kind() == reflect.Pointer {
				elemType := resultType.Elem()
				newValue := reflect.New(elemType)

				if err := cast.ToStruct(customResult, newValue.Interface(), true); err != nil {
					return zero, fmt.Errorf("failed to cast custom result: %v", err)
				}

				result = newValue.Interface().(T)
			} else {
				if err := cast.ToStruct(customResult, &result, true); err != nil {
					return zero, fmt.Errorf("failed to cast custom result: %v", err)
				}
			}

			return result, nil
		}
	}

	if clientResp.Status != "success" {
		code := "API_ERROR"
		message := clientResp.Message
		if message == "" {
			message = "Downstream API returned error"
		}

		if clientResp.Error != nil {
			if clientResp.Error.Code != "" {
				code = clientResp.Error.Code
			}
			if clientResp.Error.Message != "" {
				message = clientResp.Error.Message
			}
		}

		return zero, &api_client.ApiError{
			StatusCode: resp.StatusCode,
			Code:       code,
			Message:    message,
		}
	}

	var result T

	// NO CACHING - repeated reflection every call
	resultType := reflect.TypeOf((*T)(nil)).Elem()
	if resultType.Kind() == reflect.Pointer {
		elemType := resultType.Elem()
		newValue := reflect.New(elemType)

		if err := cast.ToStruct(clientResp.Data, newValue.Interface(), true); err != nil {
			return zero, fmt.Errorf("failed to cast data: %v", err)
		}

		result = newValue.Interface().(T)
	} else {
		if err := cast.ToStruct(clientResp.Data, &result, true); err != nil {
			return zero, fmt.Errorf("failed to cast data: %v", err)
		}
	}

	return result, nil
}

// Benchmark for FetchAndCast with struct type
func BenchmarkFetchAndCast_Struct(b *testing.B) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","data":{"id":123,"name":"test"}}`))
	}))
	defer server.Close()

	client := newTestClientRouter(server.URL)

	for b.Loop() {
		_, err := api_client.FetchAndCast[BenchResponse](client, "/test")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark UNOPTIMIZED version (no cache) for comparison
func BenchmarkFetchAndCastUnoptimized_Struct(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","data":{"id":123,"name":"test"}}`))
	}))
	defer server.Close()

	client := newTestClientRouter(server.URL)

	for b.Loop() {
		_, err := FetchAndCastUnoptimized[BenchResponse](client, "/test")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for FetchAndCast with pointer type
func BenchmarkFetchAndCast_Pointer(b *testing.B) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","data":{"id":123,"name":"test"}}`))
	}))
	defer server.Close()

	client := newTestClientRouter(server.URL)

	for b.Loop() {
		_, err := api_client.FetchAndCast[*BenchResponse](client, "/test")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark UNOPTIMIZED pointer version
func BenchmarkFetchAndCastUnoptimized_Pointer(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","data":{"id":123,"name":"test"}}`))
	}))
	defer server.Close()

	client := newTestClientRouter(server.URL)

	for b.Loop() {
		_, err := FetchAndCastUnoptimized[*BenchResponse](client, "/test")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark with CustomFunc
func BenchmarkFetchAndCast_CustomFunc(b *testing.B) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","data":{"id":123,"name":"test"}}`))
	}))
	defer server.Close()

	client := newTestClientRouter(server.URL)

	customFunc := func(resp *http.Response, clientResp *api_formatter.ClientResponse) (any, error) {
		return clientResp.Data, nil
	}

	for b.Loop() {
		_, err := api_client.FetchAndCast[BenchResponse](client, "/test", api_client.WithCustomFunc(customFunc))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark concurrent calls (realistic scenario)
func BenchmarkFetchAndCast_Concurrent(b *testing.B) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success","data":{"id":123,"name":"test"}}`))
	}))
	defer server.Close()

	client := newTestClientRouter(server.URL)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := api_client.FetchAndCast[BenchResponse](client, "/test")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Micro-benchmark: Direct reflection (single call)
func BenchmarkReflection_SingleCall(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		resultType := reflect.TypeOf((*BenchResponse)(nil)).Elem()
		_ = resultType.Kind()
	}
}

// Micro-benchmark: Reflection with type checking (as used in FetchAndCast)
func BenchmarkReflection_WithTypeCheck(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		resultType := reflect.TypeOf((*BenchResponse)(nil)).Elem()
		isPointer := resultType.Kind() == reflect.Pointer
		var elemType reflect.Type
		if isPointer {
			elemType = resultType.Elem()
		}
		_ = elemType
	}
}

// Realistic benchmark: Multiple reflection calls per operation (as in original FetchAndCast)
func BenchmarkReflection_MultipleCallsSequential(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		// First reflection call (CustomFunc path)
		resultType1 := reflect.TypeOf((*BenchResponse)(nil)).Elem()
		isPointer1 := resultType1.Kind() == reflect.Pointer
		var elemType1 reflect.Type
		if isPointer1 {
			elemType1 = resultType1.Elem()
		}
		_ = elemType1

		// Second reflection call (main result path)
		resultType2 := reflect.TypeOf((*BenchResponse)(nil)).Elem()
		isPointer2 := resultType2.Kind() == reflect.Pointer
		var elemType2 reflect.Type
		if isPointer2 {
			elemType2 = resultType2.Elem()
		}
		_ = elemType2
	}
}

// Test concurrent reflection calls (realistic for high-throughput API)
func BenchmarkReflection_Concurrent(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resultType1 := reflect.TypeOf((*BenchResponse)(nil)).Elem()
			_ = resultType1.Kind()
			resultType2 := reflect.TypeOf((*BenchResponse)(nil)).Elem()
			_ = resultType2.Kind()
		}
	})
}
