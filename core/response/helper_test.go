package response_test

import (
	"net/http"
	"testing"

	"github.com/primadi/lokstra/core/response"
)

func TestResponse_Ok(t *testing.T) {
	r := response.NewResponse()
	data := "test data"

	err := r.Ok(data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if r.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, r.StatusCode)
	}

	if r.ResponseCode != response.CodeOK {
		t.Errorf("Expected response code %s, got %s", response.CodeOK, r.ResponseCode)
	}

	if !r.Success {
		t.Error("Expected success to be true")
	}

	if r.Data != data {
		t.Errorf("Expected data %v, got %v", data, r.Data)
	}
}

func TestResponse_OkCreated(t *testing.T) {
	r := response.NewResponse()
	data := map[string]string{"id": "123"}

	err := r.OkCreated(data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if r.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, r.StatusCode)
	}

	if r.ResponseCode != response.CodeCreated {
		t.Errorf("Expected response code %s, got %s", response.CodeCreated, r.ResponseCode)
	}

	if !r.Success {
		t.Error("Expected success to be true")
	}

	if r.Data == nil {
		t.Error("Expected data to be set")
	}
}

func TestResponse_OkUpdated(t *testing.T) {
	r := response.NewResponse()
	data := "updated item"

	err := r.OkUpdated(data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if r.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, r.StatusCode)
	}

	if r.ResponseCode != response.CodeUpdated {
		t.Errorf("Expected response code %s, got %s", response.CodeUpdated, r.ResponseCode)
	}

	if !r.Success {
		t.Error("Expected success to be true")
	}

	if r.Message != "Updated successfully" {
		t.Errorf("Expected default message 'Updated successfully', got '%s'", r.Message)
	}

	if r.Data != data {
		t.Errorf("Expected data %v, got %v", data, r.Data)
	}
}

func TestResponse_OkUpdated_WithExistingMessage(t *testing.T) {
	r := response.NewResponse()
	r.Message = "Custom update message"
	data := "updated item"

	err := r.OkUpdated(data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if r.Message != "Custom update message" {
		t.Errorf("Expected existing message to be preserved, got '%s'", r.Message)
	}
}

func TestResponse_OkList(t *testing.T) {
	r := response.NewResponse()
	data := []string{"item1", "item2"}
	meta := map[string]int{"total": 2, "page": 1}

	err := r.OkList(data, meta)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if r.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, r.StatusCode)
	}

	if r.ResponseCode != response.CodeOK {
		t.Errorf("Expected response code %s, got %s", response.CodeOK, r.ResponseCode)
	}

	if !r.Success {
		t.Error("Expected success to be true")
	}

	if r.Data == nil {
		t.Error("Expected data to be set")
	}

	if r.Meta == nil {
		t.Error("Expected meta to be set")
	}
}

func TestResponse_ErrorNotFound(t *testing.T) {
	r := response.NewResponse()
	msg := "Resource not found"

	err := r.ErrorNotFound(msg)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if r.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, r.StatusCode)
	}

	if r.ResponseCode != response.CodeNotFound {
		t.Errorf("Expected response code %s, got %s", response.CodeNotFound, r.ResponseCode)
	}

	if r.Success {
		t.Error("Expected success to be false")
	}

	if r.Message != msg {
		t.Errorf("Expected message '%s', got '%s'", msg, r.Message)
	}
}

func TestResponse_ErrorDuplicate(t *testing.T) {
	r := response.NewResponse()
	msg := "Duplicate entry"

	err := r.ErrorDuplicate(msg)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if r.StatusCode != http.StatusConflict {
		t.Errorf("Expected status code %d, got %d", http.StatusConflict, r.StatusCode)
	}

	if r.ResponseCode != response.CodeDuplicate {
		t.Errorf("Expected response code %s, got %s", response.CodeDuplicate, r.ResponseCode)
	}

	if r.Success {
		t.Error("Expected success to be false")
	}

	if r.Message != msg {
		t.Errorf("Expected message '%s', got '%s'", msg, r.Message)
	}
}

func TestResponse_ErrorBadRequest(t *testing.T) {
	r := response.NewResponse()
	msg := "Invalid request"

	err := r.ErrorBadRequest(msg)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if r.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, r.StatusCode)
	}

	if r.ResponseCode != response.CodeBadRequest {
		t.Errorf("Expected response code %s, got %s", response.CodeBadRequest, r.ResponseCode)
	}

	if r.Success {
		t.Error("Expected success to be false")
	}

	if r.Message != msg {
		t.Errorf("Expected message '%s', got '%s'", msg, r.Message)
	}
}

func TestResponse_ErrorValidation(t *testing.T) {
	r := response.NewResponse()
	globalMsg := "Validation failed"
	fieldErrors := map[string]string{
		"email":    "Invalid email format",
		"password": "Password too short",
	}

	err := r.ErrorValidation(globalMsg, fieldErrors)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if r.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, r.StatusCode)
	}

	if r.ResponseCode != response.CodeBadRequest {
		t.Errorf("Expected response code %s, got %s", response.CodeBadRequest, r.ResponseCode)
	}

	if r.Success {
		t.Error("Expected success to be false")
	}

	if r.Message != globalMsg {
		t.Errorf("Expected message '%s', got '%s'", globalMsg, r.Message)
	}

	if r.FieldErrors == nil {
		t.Error("Expected field errors to be set")
	}

	if r.FieldErrors["email"] != "Invalid email format" {
		t.Error("Expected email field error to be set correctly")
	}

	if r.FieldErrors["password"] != "Password too short" {
		t.Error("Expected password field error to be set correctly")
	}
}

func TestResponse_ErrorInternal(t *testing.T) {
	r := response.NewResponse()
	msg := "Internal server error"

	err := r.ErrorInternal(msg)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if r.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, r.StatusCode)
	}

	if r.ResponseCode != response.CodeInternal {
		t.Errorf("Expected response code %s, got %s", response.CodeInternal, r.ResponseCode)
	}

	if r.Success {
		t.Error("Expected success to be false")
	}

	if r.Message != msg {
		t.Errorf("Expected message '%s', got '%s'", msg, r.Message)
	}
}

func TestResponse_WriteRaw(t *testing.T) {
	r := response.NewResponse()
	contentType := "text/plain"
	status := http.StatusOK
	data := []byte("raw response data")

	err := r.WriteRaw(contentType, status, data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if r.StatusCode != status {
		t.Errorf("Expected status code %d, got %d", status, r.StatusCode)
	}

	if !r.Success {
		t.Error("Expected success to be true")
	}

	if r.Headers.Get("Content-Type") != contentType {
		t.Errorf("Expected Content-Type '%s', got '%s'", contentType, r.Headers.Get("Content-Type"))
	}

	if string(r.RawData) != string(data) {
		t.Errorf("Expected raw data '%s', got '%s'", string(data), string(r.RawData))
	}

	// Data should also be set
	if r.Data == nil {
		t.Error("Expected data to be set")
	}
}

func TestResponse_WriteRaw_NilHeaders(t *testing.T) {
	r := response.NewResponse()
	// Ensure headers is nil initially
	r.Headers = nil

	defer func() {
		if recover() != nil {
			t.Error("WriteRaw should handle nil headers gracefully")
		}
	}()

	err := r.WriteRaw("text/plain", http.StatusOK, []byte("data"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
