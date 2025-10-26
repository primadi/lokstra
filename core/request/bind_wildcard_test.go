package request

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/response/api_formatter"
)

func init() {
	// Setup global formatter for tests
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())
}

// TestBindBody_NoWildcard tests backward compatibility (no wildcard)
func TestBindBody_NoWildcard(t *testing.T) {
	type Request struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	bodyJSON := `{"name": "John", "email": "john@example.com"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx := NewContext(w, req, nil)

	var testReq Request
	err := ctx.Req.BindBody(&testReq)
	if err != nil {
		t.Fatalf("BindBody failed: %v", err)
	}

	if testReq.Name != "John" {
		t.Errorf("Expected Name 'John', got '%s'", testReq.Name)
	}

	if testReq.Email != "john@example.com" {
		t.Errorf("Expected Email 'john@example.com', got '%s'", testReq.Email)
	}
}

// TestBindBody_WildcardOnly tests binding entire body to map[string]any with json:"*"
func TestBindBody_WildcardOnly(t *testing.T) {
	type Request struct {
		BodyData map[string]any `json:"*"`
	}

	bodyJSON := `{"name": "John Doe", "email": "john@example.com", "age": 30, "active": true}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx := NewContext(w, req, nil)

	var testReq Request
	err := ctx.Req.BindBody(&testReq)
	if err != nil {
		t.Fatalf("BindBody failed: %v", err)
	}

	if testReq.BodyData == nil {
		t.Fatal("BodyData should not be nil")
	}

	// Verify all fields are captured
	if name, ok := testReq.BodyData["name"].(string); !ok || name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %v", testReq.BodyData["name"])
	}

	if email, ok := testReq.BodyData["email"].(string); !ok || email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got %v", testReq.BodyData["email"])
	}

	if age, ok := testReq.BodyData["age"].(float64); !ok || age != 30 {
		t.Errorf("Expected age 30, got %v", testReq.BodyData["age"])
	}

	if active, ok := testReq.BodyData["active"].(bool); !ok || !active {
		t.Errorf("Expected active true, got %v", testReq.BodyData["active"])
	}
}

// TestBindBody_WildcardWithPathParam tests wildcard with path parameter
func TestBindBody_WildcardWithPathParam(t *testing.T) {
	type UpdateUserRequest struct {
		ID       string         `path:"id"`
		BodyData map[string]any `json:"*"`
	}

	bodyJSON := `{"name": "Jane", "email": "jane@example.com"}`
	req := httptest.NewRequest("PUT", "/users/123", bytes.NewBufferString(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "123")

	w := httptest.NewRecorder()
	ctx := NewContext(w, req, nil)

	var updateReq UpdateUserRequest
	err := ctx.Req.BindAll(&updateReq)
	if err != nil {
		t.Fatalf("BindAll failed: %v", err)
	}

	// Verify path parameter
	if updateReq.ID != "123" {
		t.Errorf("Expected ID '123', got '%s'", updateReq.ID)
	}

	// Verify wildcard body data
	if updateReq.BodyData == nil {
		t.Fatal("BodyData should not be nil")
	}

	if name, ok := updateReq.BodyData["name"].(string); !ok || name != "Jane" {
		t.Errorf("Expected name 'Jane', got %v", updateReq.BodyData["name"])
	}

	if email, ok := updateReq.BodyData["email"].(string); !ok || email != "jane@example.com" {
		t.Errorf("Expected email 'jane@example.com', got %v", updateReq.BodyData["email"])
	}
}

// TestBindBody_WildcardWithTypedFields tests mixing typed fields with wildcard
func TestBindBody_WildcardWithTypedFields(t *testing.T) {
	type CreateResourceRequest struct {
		Name     string         `json:"name"`
		Type     string         `json:"type"`
		Metadata map[string]any `json:"*"`
	}

	bodyJSON := `{
		"name": "MyResource",
		"type": "document",
		"author": "Jane",
		"tags": ["important", "review"],
		"priority": 5
	}`

	req := httptest.NewRequest("POST", "/resources", bytes.NewBufferString(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx := NewContext(w, req, nil)

	var createReq CreateResourceRequest
	err := ctx.Req.BindBody(&createReq)
	if err != nil {
		t.Fatalf("BindBody failed: %v", err)
	}

	// Verify typed fields
	if createReq.Name != "MyResource" {
		t.Errorf("Expected Name 'MyResource', got '%s'", createReq.Name)
	}

	if createReq.Type != "document" {
		t.Errorf("Expected Type 'document', got '%s'", createReq.Type)
	}

	// Verify wildcard metadata (should contain all fields including name and type)
	if createReq.Metadata == nil {
		t.Fatal("Metadata should not be nil")
	}

	// The wildcard captures entire body
	if author, ok := createReq.Metadata["author"].(string); !ok || author != "Jane" {
		t.Errorf("Expected author 'Jane', got %v", createReq.Metadata["author"])
	}

	if priority, ok := createReq.Metadata["priority"].(float64); !ok || priority != 5 {
		t.Errorf("Expected priority 5, got %v", createReq.Metadata["priority"])
	}

	// Name and Type should also be in metadata
	if name, ok := createReq.Metadata["name"].(string); !ok || name != "MyResource" {
		t.Errorf("Expected metadata name 'MyResource', got %v", createReq.Metadata["name"])
	}
}

// TestBindBody_WildcardEmptyBody tests wildcard with empty body
func TestBindBody_WildcardEmptyBody(t *testing.T) {
	type Request struct {
		BodyData map[string]any `json:"*"`
	}

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx := NewContext(w, req, nil)

	var testReq Request
	err := ctx.Req.BindBody(&testReq)
	if err != nil {
		t.Fatalf("BindBody with empty body should not fail: %v", err)
	}

	// Empty body should result in nil or empty map
	if len(testReq.BodyData) > 0 {
		t.Errorf("Expected empty or nil BodyData, got %v", testReq.BodyData)
	}
}

// TestBindBody_WildcardInvalidJSON tests wildcard with invalid JSON
func TestBindBody_WildcardInvalidJSON(t *testing.T) {
	type Request struct {
		BodyData map[string]any `json:"*"`
	}

	invalidJSON := `{"name": "test", "invalid": }`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx := NewContext(w, req, nil)

	var testReq Request
	err := ctx.Req.BindBody(&testReq)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	// Should be a ValidationError
	if _, ok := err.(*ValidationError); !ok {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

// TestBindBody_WildcardNestedObjects tests wildcard with nested objects
func TestBindBody_WildcardNestedObjects(t *testing.T) {
	type Request struct {
		Data map[string]any `json:"*"`
	}

	bodyJSON := `{
		"user": {
			"name": "John",
			"age": 30
		},
		"metadata": {
			"created": "2024-01-01",
			"tags": ["a", "b", "c"]
		}
	}`

	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx := NewContext(w, req, nil)

	var testReq Request
	err := ctx.Req.BindBody(&testReq)
	if err != nil {
		t.Fatalf("BindBody failed: %v", err)
	}

	if testReq.Data == nil {
		t.Fatal("Data should not be nil")
	}

	// Verify nested object
	user, ok := testReq.Data["user"].(map[string]any)
	if !ok {
		t.Fatal("Expected user to be map[string]any")
	}

	if userName, ok := user["name"].(string); !ok || userName != "John" {
		t.Errorf("Expected user.name 'John', got %v", user["name"])
	}

	// Verify nested array
	metadata, ok := testReq.Data["metadata"].(map[string]any)
	if !ok {
		t.Fatal("Expected metadata to be map[string]any")
	}

	tags, ok := metadata["tags"].([]any)
	if !ok {
		t.Fatal("Expected metadata.tags to be []any")
	}

	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}
}

// TestBindBody_WildcardWithValidation tests wildcard doesn't affect validation on typed fields
func TestBindBody_WildcardWithValidation(t *testing.T) {
	type Request struct {
		Email string         `json:"email" validate:"required,email"`
		Data  map[string]any `json:"*"`
	}

	// Valid email
	bodyJSON := `{"email": "test@example.com", "extra": "data"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx := NewContext(w, req, nil)

	var testReq Request
	err := ctx.Req.BindBody(&testReq)
	if err != nil {
		t.Fatalf("BindBody with valid email should not fail: %v", err)
	}

	if testReq.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", testReq.Email)
	}

	// Invalid email should fail validation
	invalidJSON := `{"email": "not-an-email", "extra": "data"}`
	req2 := httptest.NewRequest("POST", "/test", bytes.NewBufferString(invalidJSON))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	ctx2 := NewContext(w2, req2, nil)

	var testReq2 Request
	err2 := ctx2.Req.BindBody(&testReq2)
	if err2 == nil {
		t.Fatal("Expected validation error for invalid email")
	}
}
