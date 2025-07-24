package request_test

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/request"
)

// Test structures for binding
type PathParams struct {
	ID   string `path:"id"`
	Name string `path:"name"`
}

type QueryParams struct {
	Name   string   `query:"name"`
	Age    int      `query:"age"`
	Active bool     `query:"active"`
	Tags   []string `query:"tags"`
}

type HeaderParams struct {
	ContentType   string   `header:"Content-Type"`
	Authorization string   `header:"Authorization"`
	CustomHeaders []string `header:"X-Custom-Header"`
}

type BodyParams struct {
	Title       string `body:"title"`
	Description string `body:"description"`
	Count       int    `body:"count"`
}

type AllParams struct {
	ID            string   `path:"id"`
	Name          string   `query:"name"`
	ContentType   string   `header:"Content-Type"`
	Authorization string   `header:"Authorization"`
	Title         string   `body:"title"`
	Description   string   `body:"description"`
	Tags          []string `query:"tags"`
}

type IndexedParams struct {
	Filters []FilterParam `query:"filter"`
}

type FilterParam struct {
	Key   string
	Value string
}

type MapParams struct {
	Metadata map[string]string `query:"meta"`
}

func TestContext_BindPath(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users/123/profile/john", nil)

	// Simulate path parameters set by router
	r.SetPathValue("id", "123")
	r.SetPathValue("name", "john")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	var params PathParams
	err := ctx.BindPath(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if params.ID != "123" {
		t.Errorf("Expected ID to be '123', got '%s'", params.ID)
	}

	if params.Name != "john" {
		t.Errorf("Expected Name to be 'john', got '%s'", params.Name)
	}
}

func TestContext_BindPath_EmptyParams(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	var params PathParams
	err := ctx.BindPath(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Parameters should be empty strings since no path values are set
	if params.ID != "" {
		t.Errorf("Expected empty ID, got '%s'", params.ID)
	}

	if params.Name != "" {
		t.Errorf("Expected empty Name, got '%s'", params.Name)
	}
}

func TestContext_BindQuery(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?name=john&age=25&active=true&tags=go,web,api", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	var params QueryParams
	err := ctx.BindQuery(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if params.Name != "john" {
		t.Errorf("Expected Name to be 'john', got '%s'", params.Name)
	}

	if params.Age != 25 {
		t.Errorf("Expected Age to be 25, got %d", params.Age)
	}

	if !params.Active {
		t.Error("Expected Active to be true")
	}

	if len(params.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(params.Tags))
	}

	expectedTags := []string{"go", "web", "api"}
	for i, tag := range params.Tags {
		if tag != expectedTags[i] {
			t.Errorf("Expected tag %d to be '%s', got '%s'", i, expectedTags[i], tag)
		}
	}
}

func TestContext_BindQuery_MultipleValues(t *testing.T) {
	// Setup - Multiple values for the same parameter
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	// Manually set multiple values for the same parameter
	values := url.Values{}
	values.Add("tags", "go")
	values.Add("tags", "web")
	values.Add("tags", "api")
	r.URL.RawQuery = values.Encode()

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	type SimpleParams struct {
		Tags []string `query:"tags"`
	}

	var params SimpleParams
	err := ctx.BindQuery(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(params.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(params.Tags))
	}

	expectedTags := []string{"go", "web", "api"}
	for i, tag := range params.Tags {
		if tag != expectedTags[i] {
			t.Errorf("Expected tag %d to be '%s', got '%s'", i, expectedTags[i], tag)
		}
	}
}

func TestContext_BindQuery_InvalidTypes(t *testing.T) {
	// Setup - Invalid values for type conversion
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?age=invalid&active=notbool", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	var params QueryParams
	err := ctx.BindQuery(&params)

	// Assertions - should return error due to invalid type conversion
	if err == nil {
		t.Error("Expected error due to invalid type conversion")
	}
}

func TestContext_BindHeader(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer token123")
	r.Header.Add("X-Custom-Header", "value1")
	r.Header.Add("X-Custom-Header", "value2")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	var params HeaderParams
	err := ctx.BindHeader(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if params.ContentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got '%s'", params.ContentType)
	}

	if params.Authorization != "Bearer token123" {
		t.Errorf("Expected Authorization to be 'Bearer token123', got '%s'", params.Authorization)
	}

	if len(params.CustomHeaders) != 2 {
		t.Errorf("Expected 2 custom headers, got %d", len(params.CustomHeaders))
	}

	expectedHeaders := []string{"value1", "value2"}
	for i, header := range params.CustomHeaders {
		if header != expectedHeaders[i] {
			t.Errorf("Expected custom header %d to be '%s', got '%s'", i, expectedHeaders[i], header)
		}
	}
}

func TestContext_BindBody(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	body := `{"title":"Test Title","description":"Test Description","count":42}`
	r := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	var params BodyParams
	err := ctx.BindBody(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if params.Title != "Test Title" {
		t.Errorf("Expected Title to be 'Test Title', got '%s'", params.Title)
	}

	if params.Description != "Test Description" {
		t.Errorf("Expected Description to be 'Test Description', got '%s'", params.Description)
	}

	if params.Count != 42 {
		t.Errorf("Expected Count to be 42, got %d", params.Count)
	}
}

func TestContext_BindBody_EmptyBody(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	var params BodyParams
	err := ctx.BindBody(&params)

	// Assertions - should return error due to empty body
	if err == nil {
		t.Error("Expected error due to empty body")
	}
}

func TestContext_BindBody_InvalidJSON(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	body := `{"title":"Test Title","description":invalid json}`
	r := httptest.NewRequest("POST", "/test", strings.NewReader(body))

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	var params BodyParams
	err := ctx.BindBody(&params)

	// Assertions - should return error due to invalid JSON
	if err == nil {
		t.Error("Expected error due to invalid JSON")
	}
}

func TestContext_BindAll(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	body := `{"title":"Test Title","description":"Test Description"}`
	r := httptest.NewRequest("POST", "/users/123?name=john&tags=go,web", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer token123")
	r.SetPathValue("id", "123")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	var params AllParams
	err := ctx.BindAll(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Path params
	if params.ID != "123" {
		t.Errorf("Expected ID to be '123', got '%s'", params.ID)
	}

	// Query params
	if params.Name != "john" {
		t.Errorf("Expected Name to be 'john', got '%s'", params.Name)
	}

	if len(params.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(params.Tags))
	}

	// Header params
	if params.ContentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got '%s'", params.ContentType)
	}

	if params.Authorization != "Bearer token123" {
		t.Errorf("Expected Authorization to be 'Bearer token123', got '%s'", params.Authorization)
	}

	// Body params
	if params.Title != "Test Title" {
		t.Errorf("Expected Title to be 'Test Title', got '%s'", params.Title)
	}

	if params.Description != "Test Description" {
		t.Errorf("Expected Description to be 'Test Description', got '%s'", params.Description)
	}
}

func TestContext_BindMapParams(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?meta[key1]=value1&meta[key2]=value2&meta[key3]=value3", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	var params MapParams
	err := ctx.BindQuery(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(params.Metadata) != 3 {
		t.Errorf("Expected 3 metadata entries, got %d", len(params.Metadata))
	}

	expectedMeta := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, expected := range expectedMeta {
		if actual, ok := params.Metadata[key]; !ok {
			t.Errorf("Expected metadata key '%s' to exist", key)
		} else if actual != expected {
			t.Errorf("Expected metadata[%s] to be '%s', got '%s'", key, expected, actual)
		}
	}
}

func TestContext_BindAll_PartialErrors(t *testing.T) {
	// Setup - Valid path and query, but invalid body JSON
	w := httptest.NewRecorder()
	body := `{"title":"Test Title","invalid":json}`
	r := httptest.NewRequest("POST", "/users/123?name=john", strings.NewReader(body))
	r.SetPathValue("id", "123")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	var params AllParams
	err := ctx.BindAll(&params)

	// Assertions - should return error due to invalid JSON in body
	if err == nil {
		t.Error("Expected error due to invalid JSON in body")
	}

	// Path and query params should still be bound correctly before the body error
	if params.ID != "123" {
		t.Errorf("Expected ID to be '123' even with body error, got '%s'", params.ID)
	}

	if params.Name != "john" {
		t.Errorf("Expected Name to be 'john' even with body error, got '%s'", params.Name)
	}
}

func TestContext_BindingWithNoMatchingTags(t *testing.T) {
	// Test struct with no matching binding tags
	type NoTagParams struct {
		Field1 string
		Field2 int
	}

	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?name=john&age=25", nil)
	r.SetPathValue("id", "123")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test each binding method
	var params NoTagParams

	err := ctx.BindPath(&params)
	if err != nil {
		t.Errorf("Expected no error for path binding with no tags, got: %v", err)
	}

	err = ctx.BindQuery(&params)
	if err != nil {
		t.Errorf("Expected no error for query binding with no tags, got: %v", err)
	}

	err = ctx.BindHeader(&params)
	if err != nil {
		t.Errorf("Expected no error for header binding with no tags, got: %v", err)
	}

	// Fields should remain at zero values
	if params.Field1 != "" {
		t.Errorf("Expected Field1 to remain empty, got '%s'", params.Field1)
	}

	if params.Field2 != 0 {
		t.Errorf("Expected Field2 to remain 0, got %d", params.Field2)
	}
}
