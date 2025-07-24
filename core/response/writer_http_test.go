package response_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/response"
)

func TestResponse_WriteHttp(t *testing.T) {
	r := response.NewResponse()
	r.Ok("test data")
	r.WithHeader("X-Custom", "custom-value")

	w := httptest.NewRecorder()

	err := r.WriteHttp(w)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type is set to JSON by default
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Check custom header
	customHeader := w.Header().Get("X-Custom")
	if customHeader != "custom-value" {
		t.Errorf("Expected X-Custom header 'custom-value', got '%s'", customHeader)
	}

	// Check response body
	body := w.Body.String()
	if !strings.Contains(body, `"success":true`) {
		t.Error("Expected response body to contain success:true")
	}

	if !strings.Contains(body, `"data":"test data"`) {
		t.Error("Expected response body to contain data")
	}
}

func TestResponse_WriteHttp_WithExistingContentType(t *testing.T) {
	r := response.NewResponse()
	r.Ok("test data")
	r.WithHeader("Content-Type", "application/xml")

	w := httptest.NewRecorder()

	err := r.WriteHttp(w)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should preserve existing Content-Type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/xml" {
		t.Errorf("Expected Content-Type 'application/xml', got '%s'", contentType)
	}
}

func TestResponse_WriteHttp_WithMultipleHeaders(t *testing.T) {
	r := response.NewResponse()
	r.Ok("test")
	r.WithHeader("X-Header-1", "value1")
	r.WithHeader("X-Header-2", "value2")

	// Add multiple values for the same header
	r.GetHeaders().Add("X-Multi", "value1")
	r.GetHeaders().Add("X-Multi", "value2")

	w := httptest.NewRecorder()

	err := r.WriteHttp(w)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that all headers are set
	if w.Header().Get("X-Header-1") != "value1" {
		t.Error("Expected X-Header-1 to be set")
	}

	if w.Header().Get("X-Header-2") != "value2" {
		t.Error("Expected X-Header-2 to be set")
	}

	// Check multiple values for same header
	multiValues := w.Header().Values("X-Multi")
	if len(multiValues) != 2 {
		t.Errorf("Expected 2 values for X-Multi header, got %d", len(multiValues))
	}
}

func TestResponse_WriteHttp_WithRawData(t *testing.T) {
	r := response.NewResponse()
	rawData := []byte("This is raw data")

	r.WriteRaw("text/plain", http.StatusOK, rawData)

	w := httptest.NewRecorder()

	err := r.WriteHttp(w)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Expected Content-Type 'text/plain', got '%s'", contentType)
	}

	// Check raw data is written directly
	body := w.Body.String()
	if body != string(rawData) {
		t.Errorf("Expected body '%s', got '%s'", string(rawData), body)
	}
}

func TestResponse_WriteHttp_ErrorResponse(t *testing.T) {
	r := response.NewResponse()
	r.ErrorNotFound("Resource not found")

	w := httptest.NewRecorder()

	err := r.WriteHttp(w)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check status code
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}

	// Check response body
	body := w.Body.String()
	if !strings.Contains(body, `"success":false`) {
		t.Error("Expected response body to contain success:false")
	}

	if !strings.Contains(body, `"message":"Resource not found"`) {
		t.Error("Expected response body to contain error message")
	}
}

func TestResponse_WriteHttp_ValidationError(t *testing.T) {
	r := response.NewResponse()
	fieldErrors := map[string]string{
		"email": "Invalid email",
		"name":  "Name required",
	}
	r.ErrorValidation("Validation failed", fieldErrors)

	w := httptest.NewRecorder()

	err := r.WriteHttp(w)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Check response body contains field errors
	body := w.Body.String()
	if !strings.Contains(body, `"errors"`) {
		t.Error("Expected response body to contain errors field")
	}
}

func TestResponse_WriteHttp_EmptyResponse(t *testing.T) {
	r := response.NewResponse()

	w := httptest.NewRecorder()

	err := r.WriteHttp(w)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should use default status code from GetStatusCode()
	if w.Code != http.StatusOK {
		t.Errorf("Expected default status code %d, got %d", http.StatusOK, w.Code)
	}

	// Should set default content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected default Content-Type 'application/json', got '%s'", contentType)
	}

	// Should produce valid JSON
	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}

func TestResponse_WriteHttp_NilHeaders(t *testing.T) {
	r := response.NewResponse()
	r.Ok("test")
	// Ensure Headers is nil
	r.Headers = nil

	w := httptest.NewRecorder()

	err := r.WriteHttp(w)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should still set default content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected default Content-Type 'application/json', got '%s'", contentType)
	}
}

func TestResponse_WriteHttp_ZeroStatusCode(t *testing.T) {
	r := response.NewResponse()
	r.Success = true
	r.Data = "test"
	// StatusCode remains 0

	w := httptest.NewRecorder()

	err := r.WriteHttp(w)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should use default status code from GetStatusCode()
	if w.Code != http.StatusOK {
		t.Errorf("Expected default status code %d, got %d", http.StatusOK, w.Code)
	}
}
