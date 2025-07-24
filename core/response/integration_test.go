package response_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/response"
)

func TestResponse_CompleteWorkflow_Success(t *testing.T) {
	// Test a complete success workflow
	r := response.NewResponse()

	// Chain methods
	r.WithMessage("Operation successful").
		WithHeader("X-Request-ID", "123").
		WithHeader("X-Version", "1.0").
		Ok(map[string]string{"id": "user123", "name": "John Doe"})

	// Write to HTTP response
	w := httptest.NewRecorder()
	err := r.WriteHttp(w)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify complete response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"success":true`) {
		t.Error("Expected success to be true")
	}

	if !strings.Contains(body, `"message":"Operation successful"`) {
		t.Error("Expected custom message")
	}

	if !strings.Contains(body, `"code":"OK"`) {
		t.Error("Expected OK response code")
	}

	if !strings.Contains(body, `"id":"user123"`) {
		t.Error("Expected data to be included")
	}

	if w.Header().Get("X-Request-ID") != "123" {
		t.Error("Expected custom headers to be set")
	}
}

func TestResponse_CompleteWorkflow_Error(t *testing.T) {
	// Test a complete error workflow
	r := response.NewResponse()

	fieldErrors := map[string]string{
		"email":    "Email is required",
		"password": "Password must be at least 8 characters",
	}

	r.WithHeader("X-Error-ID", "err456").
		ErrorValidation("Validation failed", fieldErrors)

	// Write to HTTP response
	w := httptest.NewRecorder()
	err := r.WriteHttp(w)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify complete error response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"success":false`) {
		t.Error("Expected success to be false")
	}

	if !strings.Contains(body, `"code":"BAD_REQUEST"`) {
		t.Error("Expected BAD_REQUEST response code")
	}

	if !strings.Contains(body, `"message":"Validation failed"`) {
		t.Error("Expected validation message")
	}

	if !strings.Contains(body, `"errors"`) {
		t.Error("Expected field errors")
	}

	if !strings.Contains(body, `"email":"Email is required"`) {
		t.Error("Expected email field error")
	}

	if w.Header().Get("X-Error-ID") != "err456" {
		t.Error("Expected error header to be set")
	}
}

func TestResponse_RawDataWorkflow(t *testing.T) {
	// Test raw data workflow
	r := response.NewResponse()

	rawData := []byte(`<html><body>Hello World</body></html>`)

	r.WithHeader("X-Content-Source", "template").
		WriteRaw("text/html", http.StatusOK, rawData)

	// Write to HTTP response
	w := httptest.NewRecorder()
	err := r.WriteHttp(w)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify raw response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "text/html" {
		t.Errorf("Expected Content-Type text/html, got %s", w.Header().Get("Content-Type"))
	}

	body := w.Body.String()
	if body != string(rawData) {
		t.Errorf("Expected raw data, got %s", body)
	}

	if w.Header().Get("X-Content-Source") != "template" {
		t.Error("Expected custom header to be preserved")
	}
}

func TestResponse_JSONFormatterIntegration(t *testing.T) {
	// Test integration with JSON formatter
	formatter := response.NewJSONFormatter()
	r := response.NewResponse()

	r.OkList(
		[]map[string]string{
			{"id": "1", "name": "Item 1"},
			{"id": "2", "name": "Item 2"},
		},
		map[string]interface{}{
			"total":       2,
			"page":        1,
			"per_page":    10,
			"total_pages": 1,
		},
	)

	// Test HTTP writing
	w := httptest.NewRecorder()
	err := formatter.WriteHttp(w, r)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"total":2`) {
		t.Error("Expected meta data in response")
	}

	if !strings.Contains(body, `"name":"Item 1"`) {
		t.Error("Expected list data in response")
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected JSON content type")
	}
}

func TestResponse_ChainedMethodsIntegration(t *testing.T) {
	// Test that all methods can be chained properly
	r := response.NewResponse()

	result := r.
		WithMessage("Test message").
		WithData("test data").
		WithMeta(map[string]int{"count": 1}).
		WithHeader("X-Test-1", "value1").
		WithHeader("X-Test-2", "value2").
		SetStatusCode(http.StatusCreated)

	// All methods should return the same instance
	if result != r {
		t.Error("Expected all methods to return the same instance for chaining")
	}

	// Verify all values are set
	if r.Message != "Test message" {
		t.Error("Message not set correctly")
	}

	if r.Data != "test data" {
		t.Error("Data not set correctly")
	}

	if r.Meta == nil {
		t.Error("Meta not set correctly")
	}

	if r.StatusCode != http.StatusCreated {
		t.Error("Status code not set correctly")
	}

	if r.Headers.Get("X-Test-1") != "value1" || r.Headers.Get("X-Test-2") != "value2" {
		t.Error("Headers not set correctly")
	}
}

func TestResponse_MultipleResponseTypes(t *testing.T) {
	// Test all response type methods work correctly
	testCases := []struct {
		name            string
		setupFunc       func(r *response.Response) error
		expectedStatus  int
		expectedCode    response.ResponseCode
		expectedSuccess bool
	}{
		{
			name:            "Ok response",
			setupFunc:       func(r *response.Response) error { return r.Ok("data") },
			expectedStatus:  http.StatusOK,
			expectedCode:    response.CodeOK,
			expectedSuccess: true,
		},
		{
			name:            "Created response",
			setupFunc:       func(r *response.Response) error { return r.OkCreated("data") },
			expectedStatus:  http.StatusCreated,
			expectedCode:    response.CodeCreated,
			expectedSuccess: true,
		},
		{
			name:            "Updated response",
			setupFunc:       func(r *response.Response) error { return r.OkUpdated("data") },
			expectedStatus:  http.StatusOK,
			expectedCode:    response.CodeUpdated,
			expectedSuccess: true,
		},
		{
			name:            "Not found error",
			setupFunc:       func(r *response.Response) error { return r.ErrorNotFound("not found") },
			expectedStatus:  http.StatusNotFound,
			expectedCode:    response.CodeNotFound,
			expectedSuccess: false,
		},
		{
			name:            "Bad request error",
			setupFunc:       func(r *response.Response) error { return r.ErrorBadRequest("bad request") },
			expectedStatus:  http.StatusBadRequest,
			expectedCode:    response.CodeBadRequest,
			expectedSuccess: false,
		},
		{
			name:            "Internal error",
			setupFunc:       func(r *response.Response) error { return r.ErrorInternal("internal error") },
			expectedStatus:  http.StatusInternalServerError,
			expectedCode:    response.CodeInternal,
			expectedSuccess: false,
		},
		{
			name:            "Duplicate error",
			setupFunc:       func(r *response.Response) error { return r.ErrorDuplicate("duplicate") },
			expectedStatus:  http.StatusConflict,
			expectedCode:    response.CodeDuplicate,
			expectedSuccess: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := response.NewResponse()

			err := tc.setupFunc(r)
			if err != nil {
				t.Errorf("Setup function returned error: %v", err)
			}

			if r.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, r.StatusCode)
			}

			if r.ResponseCode != tc.expectedCode {
				t.Errorf("Expected code %s, got %s", tc.expectedCode, r.ResponseCode)
			}

			if r.Success != tc.expectedSuccess {
				t.Errorf("Expected success %v, got %v", tc.expectedSuccess, r.Success)
			}
		})
	}
}
