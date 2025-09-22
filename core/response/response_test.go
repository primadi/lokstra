package response_test

import (
	"net/http"
	"testing"

	"github.com/primadi/lokstra/core/response"
)

func TestNewResponse(t *testing.T) {
	r := response.NewResponse()

	if r == nil {
		t.Error("Expected response to be created, got nil")
		return
	}

	// Check default values
	if r.StatusCode != 0 {
		t.Errorf("Expected default status code 0, got %d", r.StatusCode)
	}

	if r.Success != false {
		t.Errorf("Expected default success false, got %v", r.Success)
	}

	if r.Data != nil {
		t.Errorf("Expected default data nil, got %v", r.Data)
	}
}

func TestResponse_WithMessage(t *testing.T) {
	r := response.NewResponse()
	msg := "Test message"

	result := r.WithMessage(msg)

	if result != r {
		t.Error("Expected WithMessage to return the same response instance")
	}

	if r.Message != msg {
		t.Errorf("Expected message '%s', got '%s'", msg, r.Message)
	}
}

func TestResponse_WithData(t *testing.T) {
	r := response.NewResponse()
	data := map[string]any{"key": "value"}

	result := r.WithData(data)

	if result != r {
		t.Error("Expected WithData to return the same response instance")
	}

	if r.Data == nil {
		t.Error("Expected data to be set")
	}

	// Verify data content
	dataMap, ok := r.Data.(map[string]any)
	if !ok {
		t.Error("Expected data to be a map")
	}

	if dataMap["key"] != "value" {
		t.Error("Expected data to contain correct key-value pair")
	}
}

func TestResponse_WithMeta(t *testing.T) {
	r := response.NewResponse()
	meta := map[string]any{"total": 100, "page": 1}

	result := r.WithMeta(meta)

	if result != r {
		t.Error("Expected WithMeta to return the same response instance")
	}

	if r.Meta == nil {
		t.Error("Expected meta to be set")
	}

	// Verify meta content
	metaMap, ok := r.Meta.(map[string]any)
	if !ok {
		t.Error("Expected meta to be a map")
	}

	if metaMap["total"] != 100 {
		t.Error("Expected meta to contain correct total value")
	}

	if metaMap["page"] != 1 {
		t.Error("Expected meta to contain correct page value")
	}
}

func TestResponse_WithHeader(t *testing.T) {
	r := response.NewResponse()
	key := "X-Custom-Header"
	value := "custom-value"

	result := r.WithHeader(key, value)

	if result != r {
		t.Error("Expected WithHeader to return the same response instance")
	}

	if r.Headers == nil {
		t.Error("Expected headers to be initialized")
	}

	if r.Headers.Get(key) != value {
		t.Errorf("Expected header '%s' to be '%s', got '%s'", key, value, r.Headers.Get(key))
	}
}

func TestResponse_WithHeader_Multiple(t *testing.T) {
	r := response.NewResponse()

	r.WithHeader("X-Header-1", "value1").
		WithHeader("X-Header-2", "value2")

	if r.Headers.Get("X-Header-1") != "value1" {
		t.Error("Expected first header to be set")
	}

	if r.Headers.Get("X-Header-2") != "value2" {
		t.Error("Expected second header to be set")
	}
}

func TestResponse_GetHeaders(t *testing.T) {
	r := response.NewResponse()

	// Test when headers is nil
	headers := r.GetHeaders()
	if headers == nil {
		t.Error("Expected GetHeaders to initialize and return headers")
	}

	// Test when headers already exists
	r.WithHeader("Test", "value")
	headers2 := r.GetHeaders()
	if headers2.Get("Test") != "value" {
		t.Error("Expected existing headers to be returned")
	}
}

func TestResponse_GetStatusCode(t *testing.T) {
	r := response.NewResponse()

	// Test default status code
	statusCode := r.GetStatusCode()
	if statusCode != http.StatusOK {
		t.Errorf("Expected default status code %d, got %d", http.StatusOK, statusCode)
	}

	// Test custom status code
	r.StatusCode = http.StatusCreated
	statusCode = r.GetStatusCode()
	if statusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, statusCode)
	}
}

func TestResponse_SetStatusCode(t *testing.T) {
	r := response.NewResponse()

	result := r.SetStatusCode(http.StatusNotFound)

	if result != r {
		t.Error("Expected SetStatusCode to return the same response instance")
	}

	if r.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, r.StatusCode)
	}
}

func TestResponse_ChainedMethods(t *testing.T) {
	r := response.NewResponse()

	// Test method chaining
	result := r.WithMessage("Test").
		WithData("data").
		WithMeta("meta").
		WithHeader("X-Test", "test").
		SetStatusCode(http.StatusCreated)

	if result != r {
		t.Error("Expected all methods to return the same response instance")
	}

	if r.Message != "Test" {
		t.Error("Expected message to be set")
	}

	if r.Data != "data" {
		t.Error("Expected data to be set")
	}

	if r.Meta != "meta" {
		t.Error("Expected meta to be set")
	}

	if r.Headers.Get("X-Test") != "test" {
		t.Error("Expected header to be set")
	}

	if r.StatusCode != http.StatusCreated {
		t.Error("Expected status code to be set")
	}
}

func TestResponse_FieldsStructure(t *testing.T) {
	r := response.NewResponse()

	// Test all field types
	r.StatusCode = http.StatusOK
	r.ResponseCode = response.CodeOK
	r.Success = true
	r.Message = "test message"
	r.Data = map[string]string{"key": "value"}
	r.Meta = map[string]int{"total": 10}
	r.WithHeader("X-Test", "value")
	r.RawData = []byte("raw data")
	r.FieldErrors = map[string]string{"field1": "error1"}

	// Verify all fields are set correctly
	if r.StatusCode != http.StatusOK {
		t.Error("StatusCode not set correctly")
	}

	if r.ResponseCode != response.CodeOK {
		t.Error("ResponseCode not set correctly")
	}

	if !r.Success {
		t.Error("Success not set correctly")
	}

	if r.Message != "test message" {
		t.Error("Message not set correctly")
	}

	if r.Data == nil {
		t.Error("Data not set correctly")
	}

	if r.Meta == nil {
		t.Error("Meta not set correctly")
	}

	if r.Headers == nil {
		t.Error("Headers not set correctly")
	}

	if r.RawData == nil {
		t.Error("RawData not set correctly")
	}

	if r.FieldErrors == nil {
		t.Error("FieldErrors not set correctly")
	}
}
