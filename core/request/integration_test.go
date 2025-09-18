package request_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/primadi/lokstra/core/request"
)

// CompleteRequestParams represents a comprehensive request structure
type CompleteRequestParams struct {
	// Path parameters
	UserID string `path:"user_id"`
	PostID string `path:"post_id"`

	// Query parameters
	Page     int               `query:"page"`
	Limit    int               `query:"limit"`
	Sort     string            `query:"sort"`
	Tags     []string          `query:"tags"`
	Active   bool              `query:"active"`
	Metadata map[string]string `query:"meta"`

	// Header parameters
	ContentType   string `header:"Content-Type"`
	Authorization string `header:"Authorization"`
	UserAgent     string `header:"User-Agent"`

	// Body parameters
	Title       string `body:"title"`
	Content     string `body:"content"`
	PublishedAt string `body:"published_at"`
	ViewCount   int    `body:"view_count"`
}

func TestIntegration_CompleteRequestBinding(t *testing.T) {
	// Setup a comprehensive request with all parameter types
	jsonBody := `{
		"title": "Test Post",
		"content": "This is a test post content",
		"published_at": "2024-01-15T10:30:00Z",
		"view_count": 1250
	}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/users/123/posts/456?page=2&limit=20&sort=created_at&tags=tech,go&active=true&meta[category]=blog&meta[priority]=high", strings.NewReader(jsonBody))

	// Set headers
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9")
	r.Header.Set("User-Agent", "Test-Client/1.0")

	// Set path parameters (simulating router behavior)
	r.SetPathValue("user_id", "123")
	r.SetPathValue("post_id", "456")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test complete binding
	var params CompleteRequestParams
	err := ctx.BindAll(&params)

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error in complete binding, got: %v", err)
	}

	// Validate path parameters
	if params.UserID != "123" {
		t.Errorf("Expected UserID to be '123', got '%s'", params.UserID)
	}

	if params.PostID != "456" {
		t.Errorf("Expected PostID to be '456', got '%s'", params.PostID)
	}

	// Validate query parameters
	if params.Page != 2 {
		t.Errorf("Expected Page to be 2, got %d", params.Page)
	}

	if params.Limit != 20 {
		t.Errorf("Expected Limit to be 20, got %d", params.Limit)
	}

	if params.Sort != "created_at" {
		t.Errorf("Expected Sort to be 'created_at', got '%s'", params.Sort)
	}

	if !params.Active {
		t.Error("Expected Active to be true")
	}

	expectedTags := []string{"tech", "go"}
	if len(params.Tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(params.Tags))
	}

	for i, tag := range params.Tags {
		if tag != expectedTags[i] {
			t.Errorf("Expected tag %d to be '%s', got '%s'", i, expectedTags[i], tag)
		}
	}

	// Validate metadata map
	expectedMeta := map[string]string{
		"category": "blog",
		"priority": "high",
	}

	if len(params.Metadata) != len(expectedMeta) {
		t.Errorf("Expected %d metadata entries, got %d", len(expectedMeta), len(params.Metadata))
	}

	for key, expected := range expectedMeta {
		if actual, ok := params.Metadata[key]; !ok {
			t.Errorf("Expected metadata key '%s' to exist", key)
		} else if actual != expected {
			t.Errorf("Expected metadata[%s] to be '%s', got '%s'", key, expected, actual)
		}
	}

	// Validate header parameters
	if params.ContentType != "application/json" {
		t.Errorf("Expected ContentType to be 'application/json', got '%s'", params.ContentType)
	}

	if params.Authorization != "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9" {
		t.Errorf("Expected Authorization to match, got '%s'", params.Authorization)
	}

	if params.UserAgent != "Test-Client/1.0" {
		t.Errorf("Expected UserAgent to be 'Test-Client/1.0', got '%s'", params.UserAgent)
	}

	// Validate body parameters
	if params.Title != "Test Post" {
		t.Errorf("Expected Title to be 'Test Post', got '%s'", params.Title)
	}

	if params.Content != "This is a test post content" {
		t.Errorf("Expected Content to match, got '%s'", params.Content)
	}

	if params.PublishedAt != "2024-01-15T10:30:00Z" {
		t.Errorf("Expected PublishedAt to be '2024-01-15T10:30:00Z', got '%s'", params.PublishedAt)
	}

	if params.ViewCount != 1250 {
		t.Errorf("Expected ViewCount to be 1250, got %d", params.ViewCount)
	}
}

func TestIntegration_RealWorldAPIEndpoint(t *testing.T) {
	// Simulate a real-world API endpoint: PUT /api/v1/users/:id
	type UpdateUserRequest struct {
		ID          string `path:"id"`
		Name        string `body:"name"`
		Email       string `body:"email"`
		Age         int    `body:"age"`
		Active      bool   `body:"active"`
		ContentType string `header:"Content-Type"`
		IfMatch     string `header:"If-Match"`
	}

	jsonBody := `{
		"name": "John Doe Updated",
		"email": "john.updated@example.com",
		"age": 31,
		"active": true
	}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/users/user123", strings.NewReader(jsonBody))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("If-Match", "\"etag-12345\"")
	r.SetPathValue("id", "user123")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	var req UpdateUserRequest
	err := ctx.BindAll(&req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Validate binding
	if req.ID != "user123" {
		t.Errorf("Expected ID to be 'user123', got '%s'", req.ID)
	}

	if req.Name != "John Doe Updated" {
		t.Errorf("Expected Name to be 'John Doe Updated', got '%s'", req.Name)
	}

	if req.Email != "john.updated@example.com" {
		t.Errorf("Expected Email to be 'john.updated@example.com', got '%s'", req.Email)
	}

	if req.Age != 31 {
		t.Errorf("Expected Age to be 31, got %d", req.Age)
	}

	if !req.Active {
		t.Error("Expected Active to be true")
	}

	if req.ContentType != "application/json" {
		t.Errorf("Expected ContentType to be 'application/json', got '%s'", req.ContentType)
	}

	if req.IfMatch != "\"etag-12345\"" {
		t.Errorf("Expected IfMatch to be '\"etag-12345\"', got '%s'", req.IfMatch)
	}
}

func TestIntegration_HandlerFunctionWorkflow(t *testing.T) {
	// Test complete workflow with a handler function
	type SearchRequest struct {
		Query    string   `query:"q"`
		Category string   `query:"category"`
		Tags     []string `query:"tags"`
		Page     int      `query:"page"`
		Size     int      `query:"size"`
	}

	// Create a handler function
	handler := func(ctx *request.Context) error {
		var req SearchRequest
		if err := ctx.BindQuery(&req); err != nil {
			return err
		}

		// Set response based on parsed request
		ctx.WithMessage("Search completed").WithData(map[string]any{
			"query":    req.Query,
			"category": req.Category,
			"tags":     req.Tags,
			"page":     req.Page,
			"size":     req.Size,
			"results":  []string{"result1", "result2", "result3"},
		})

		return nil
	}

	// Setup request
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search?q=golang&category=programming&tags=web,api&page=1&size=10", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Execute handler
	err := handler(ctx)

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error from handler, got: %v", err)
	}

	if ctx.Message != "Search completed" {
		t.Errorf("Expected message to be 'Search completed', got '%s'", ctx.Message)
	}

	if ctx.Data == nil {
		t.Fatal("Expected Data to be set")
	}

	data, ok := ctx.Data.(map[string]any)
	if !ok {
		t.Fatal("Expected Data to be map[string]any")
	}

	if data["query"] != "golang" {
		t.Errorf("Expected query to be 'golang', got '%v'", data["query"])
	}

	if data["category"] != "programming" {
		t.Errorf("Expected category to be 'programming', got '%v'", data["category"])
	}
}

func TestIntegration_ContextCancellationWithBinding(t *testing.T) {
	// Test that context cancellation doesn't interfere with binding
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?name=john&age=25", nil)

	baseCtx, baseCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer baseCancel()

	r = r.WithContext(baseCtx)
	ctx, cancel := request.NewContext(w, r)

	type TestParams struct {
		Name string `query:"name"`
		Age  int    `query:"age"`
	}

	var params TestParams
	err := ctx.BindQuery(&params)

	// Binding should work before cancellation
	if err != nil {
		t.Errorf("Expected no error before cancellation, got: %v", err)
	}

	if params.Name != "john" {
		t.Errorf("Expected Name to be 'john', got '%s'", params.Name)
	}

	if params.Age != 25 {
		t.Errorf("Expected Age to be 25, got %d", params.Age)
	}

	// Cancel context
	cancel()

	// Context should be cancelled
	if ctx.Err() != context.Canceled {
		t.Error("Expected context to be cancelled")
	}

	// But binding should still work with already parsed data
	var params2 TestParams
	err2 := ctx.BindQuery(&params2)
	if err2 != nil {
		t.Errorf("Expected binding to work even after cancellation, got: %v", err2)
	}
}

func TestIntegration_MultipleBindingCalls(t *testing.T) {
	// Test multiple binding calls on the same context
	jsonBody := `{"title": "Test", "content": "Content"}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test?name=john&age=25", strings.NewReader(jsonBody))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer token")
	r.SetPathValue("id", "123")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Define separate structs for different binding types
	type PathParams struct {
		ID string `path:"id"`
	}

	type QueryParams struct {
		Name string `query:"name"`
		Age  int    `query:"age"`
	}

	type HeaderParams struct {
		ContentType   string `header:"Content-Type"`
		Authorization string `header:"Authorization"`
	}

	type BodyParams struct {
		Title   string `body:"title"`
		Content string `body:"content"`
	}

	// Bind each type separately
	var pathParams PathParams
	var queryParams QueryParams
	var headerParams HeaderParams
	var bodyParams BodyParams

	err1 := ctx.BindPath(&pathParams)
	err2 := ctx.BindQuery(&queryParams)
	err3 := ctx.BindHeader(&headerParams)
	err4 := ctx.BindBody(&bodyParams)

	// All should succeed
	if err1 != nil {
		t.Errorf("Expected no error for path binding, got: %v", err1)
	}

	if err2 != nil {
		t.Errorf("Expected no error for query binding, got: %v", err2)
	}

	if err3 != nil {
		t.Errorf("Expected no error for header binding, got: %v", err3)
	}

	if err4 != nil {
		t.Errorf("Expected no error for body binding, got: %v", err4)
	}

	// Validate all bindings
	if pathParams.ID != "123" {
		t.Errorf("Expected path ID to be '123', got '%s'", pathParams.ID)
	}

	if queryParams.Name != "john" {
		t.Errorf("Expected query name to be 'john', got '%s'", queryParams.Name)
	}

	if queryParams.Age != 25 {
		t.Errorf("Expected query age to be 25, got %d", queryParams.Age)
	}

	if headerParams.ContentType != "application/json" {
		t.Errorf("Expected content type to be 'application/json', got '%s'", headerParams.ContentType)
	}

	if headerParams.Authorization != "Bearer token" {
		t.Errorf("Expected authorization to be 'Bearer token', got '%s'", headerParams.Authorization)
	}

	if bodyParams.Title != "Test" {
		t.Errorf("Expected body title to be 'Test', got '%s'", bodyParams.Title)
	}

	if bodyParams.Content != "Content" {
		t.Errorf("Expected body content to be 'Content', got '%s'", bodyParams.Content)
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	// Test error scenarios in realistic request handling

	tests := []struct {
		name        string
		method      string
		url         string
		body        string
		headers     map[string]string
		pathParams  map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Invalid JSON body",
			method:      "POST",
			url:         "/test",
			body:        `{"invalid": json}`,
			headers:     map[string]string{"Content-Type": "application/json"},
			expectError: true,
			errorMsg:    "JSON parsing error",
		},
		{
			name:        "Invalid query parameter type",
			method:      "GET",
			url:         "/test?age=notanumber",
			body:        "",
			expectError: true,
			errorMsg:    "Type conversion error",
		},
		{
			name:        "Valid request",
			method:      "GET",
			url:         "/test?name=john&age=25",
			body:        `{"title": "Test"}`,
			headers:     map[string]string{"Content-Type": "application/json"},
			pathParams:  map[string]string{"id": "123"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tt.body != "" {
				body = bytes.NewBufferString(tt.body)
			} else {
				body = nil
			}

			w := httptest.NewRecorder()
			var r *http.Request
			if body != nil {
				r = httptest.NewRequest(tt.method, tt.url, body)
			} else {
				r = httptest.NewRequest(tt.method, tt.url, nil)
			}

			// Set headers
			for key, value := range tt.headers {
				r.Header.Set(key, value)
			}

			// Set path parameters
			for key, value := range tt.pathParams {
				r.SetPathValue(key, value)
			}

			ctx, cancel := request.NewContext(w, r)
			defer cancel()

			type TestRequest struct {
				ID    string `path:"id"`
				Name  string `query:"name"`
				Age   int    `query:"age"`
				Title string `body:"title"`
			}

			var req TestRequest
			err := ctx.BindAll(&req)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// For valid requests, verify binding worked
			if !tt.expectError {
				if req.Name != "john" {
					t.Errorf("Expected name to be 'john', got '%s'", req.Name)
				}

				if req.Age != 25 {
					t.Errorf("Expected age to be 25, got %d", req.Age)
				}

				if tt.body != "" && req.Title != "Test" {
					t.Errorf("Expected title to be 'Test', got '%s'", req.Title)
				}

				if req.ID != "123" && tt.pathParams["id"] == "123" {
					t.Errorf("Expected ID to be '123', got '%s'", req.ID)
				}
			}
		})
	}
}
