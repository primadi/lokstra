package response_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/response"
)

func TestJSONFormatter_NewJSONFormatter(t *testing.T) {
	formatter := response.NewJSONFormatter()

	if formatter == nil {
		t.Error("Expected formatter to be created, got nil")
	}
}

func TestJSONFormatter_ContentType(t *testing.T) {
	formatter := response.NewJSONFormatter()

	contentType := formatter.ContentType()
	expected := "application/json"

	if contentType != expected {
		t.Errorf("Expected content type '%s', got '%s'", expected, contentType)
	}
}

func TestJSONFormatter_WriteHttp(t *testing.T) {
	formatter := response.NewJSONFormatter()
	r := response.NewResponse()

	// Set up response data
	r.Ok("test data")
	r.WithHeader("X-Custom", "custom-value")

	// Create test HTTP response writer
	w := httptest.NewRecorder()

	err := formatter.WriteHttp(w, r)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Check custom header
	customHeader := w.Header().Get("X-Custom")
	if customHeader != "custom-value" {
		t.Errorf("Expected X-Custom header 'custom-value', got '%s'", customHeader)
	}

	// Check response body contains JSON
	body := w.Body.String()
	if !strings.Contains(body, `"success":true`) {
		t.Error("Expected response body to contain success:true")
	}

	if !strings.Contains(body, `"data":"test data"`) {
		t.Error("Expected response body to contain data")
	}
}

func TestJSONFormatter_WriteHttp_WithMultipleHeaders(t *testing.T) {
	formatter := response.NewJSONFormatter()
	r := response.NewResponse()

	r.Ok("test")
	r.WithHeader("X-Header-1", "value1")
	r.WithHeader("X-Header-2", "value2")

	// Add multiple values for the same header
	r.GetHeaders().Add("X-Multi", "value1")
	r.GetHeaders().Add("X-Multi", "value2")

	w := httptest.NewRecorder()

	err := formatter.WriteHttp(w, r)

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

func TestJSONFormatter_WriteBuffer(t *testing.T) {
	formatter := response.NewJSONFormatter()
	r := response.NewResponse()

	r.Ok("buffer test")

	var buf bytes.Buffer
	err := formatter.WriteBuffer(&buf, r)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"success":true`) {
		t.Error("Expected buffer output to contain success:true")
	}

	if !strings.Contains(output, `"data":"buffer test"`) {
		t.Error("Expected buffer output to contain data")
	}
}

func TestJSONFormatter_WriteStdout(t *testing.T) {
	formatter := response.NewJSONFormatter()
	r := response.NewResponse()

	r.Ok("stdout test")

	// This test just ensures the method doesn't panic
	// We can't easily capture stdout in a unit test
	err := formatter.WriteStdout(r)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestJSONFormatter_WriteHttp_ErrorResponse(t *testing.T) {
	formatter := response.NewJSONFormatter()
	r := response.NewResponse()

	r.ErrorNotFound("Resource not found")

	w := httptest.NewRecorder()

	err := formatter.WriteHttp(w, r)

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

	if !strings.Contains(body, `"code":"NOT_FOUND"`) {
		t.Error("Expected response body to contain error code")
	}
}

func TestJSONFormatter_WriteHttp_ValidationError(t *testing.T) {
	formatter := response.NewJSONFormatter()
	r := response.NewResponse()

	fieldErrors := map[string]string{
		"email": "Invalid email",
		"name":  "Name required",
	}
	r.ErrorValidation("Validation failed", fieldErrors)

	w := httptest.NewRecorder()

	err := formatter.WriteHttp(w, r)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"errors"`) {
		t.Error("Expected response body to contain errors field")
	}

	if !strings.Contains(body, `"email":"Invalid email"`) {
		t.Error("Expected response body to contain email error")
	}

	if !strings.Contains(body, `"name":"Name required"`) {
		t.Error("Expected response body to contain name error")
	}
}

func TestJSONFormatter_WriteHttp_EmptyResponse(t *testing.T) {
	formatter := response.NewJSONFormatter()
	r := response.NewResponse()

	// Empty response - just default values
	w := httptest.NewRecorder()

	err := formatter.WriteHttp(w, r)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should still produce valid JSON
	body := w.Body.String()
	if !strings.Contains(body, `"success":false`) {
		t.Error("Expected response body to contain success:false for empty response")
	}

	// Should have default status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected default status code %d, got %d", http.StatusOK, w.Code)
	}
}
