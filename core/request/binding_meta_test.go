package request_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/request"
)

// Test structures for binding metadata
type TaggedStruct struct {
	PathParam   string            `path:"id"`
	QueryParam  string            `query:"name"`
	HeaderParam string            `header:"Content-Type"`
	MapParam    map[string]string `query:"meta"`
	SliceParam  []string          `query:"tags"`
}

type UntaggedStruct struct {
	Field1 string
	Field2 int
	Field3 bool
}

type MixedStruct struct {
	Tagged   string `query:"tagged"`
	Untagged string
	Header   string `header:"Authorization"`
}

type IndexedStruct struct {
	Filters []FilterItem `query:"filter"`
}

type FilterItem struct {
	Key   string
	Value string
}

type CustomUnmarshalStruct struct {
	Value string `query:"custom"`
}

func (c *CustomUnmarshalStruct) UnmarshalJSON(data []byte) error {
	c.Value = "unmarshaled:" + string(data)
	return nil
}

func TestBindingMeta_TagParsing(t *testing.T) {
	// Test that binding metadata correctly identifies and parses tags
	w := httptest.NewRecorder()
	body := `{}` // Add empty JSON body to avoid EOF
	r := httptest.NewRequest("GET", "/test/123?name=john&tags=go,web&meta[key1]=value1", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.SetPathValue("id", "123")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	var params TaggedStruct
	err := ctx.BindAll(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if params.PathParam != "123" {
		t.Errorf("Expected PathParam to be '123', got '%s'", params.PathParam)
	}

	if params.QueryParam != "john" {
		t.Errorf("Expected QueryParam to be 'john', got '%s'", params.QueryParam)
	}

	if params.HeaderParam != "application/json" {
		t.Errorf("Expected HeaderParam to be 'application/json', got '%s'", params.HeaderParam)
	}

	if len(params.SliceParam) != 2 {
		t.Errorf("Expected 2 slice items, got %d", len(params.SliceParam))
	}

	if len(params.MapParam) != 1 {
		t.Errorf("Expected 1 map item, got %d", len(params.MapParam))
	}

	if params.MapParam["key1"] != "value1" {
		t.Errorf("Expected map[key1] to be 'value1', got '%s'", params.MapParam["key1"])
	}
}

func TestBindingMeta_UntaggedFields(t *testing.T) {
	// Test that untagged fields are ignored during binding
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?Field1=value1&Field2=42&Field3=true", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	var params UntaggedStruct
	err := ctx.BindQuery(&params)

	// Should not error, but fields should remain at zero values
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if params.Field1 != "" {
		t.Errorf("Expected Field1 to remain empty, got '%s'", params.Field1)
	}

	if params.Field2 != 0 {
		t.Errorf("Expected Field2 to remain 0, got %d", params.Field2)
	}

	if params.Field3 != false {
		t.Errorf("Expected Field3 to remain false, got %t", params.Field3)
	}
}

func TestBindingMeta_MixedTaggedUntagged(t *testing.T) {
	// Test struct with both tagged and untagged fields
	w := httptest.NewRecorder()
	body := `{}` // Add empty JSON body to avoid EOF
	r := httptest.NewRequest("GET", "/test?tagged=value&Untagged=ignored", strings.NewReader(body))
	r.Header.Set("Authorization", "Bearer token")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	var params MixedStruct
	err := ctx.BindAll(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Tagged field should be bound
	if params.Tagged != "value" {
		t.Errorf("Expected Tagged to be 'value', got '%s'", params.Tagged)
	}

	// Untagged field should remain empty
	if params.Untagged != "" {
		t.Errorf("Expected Untagged to remain empty, got '%s'", params.Untagged)
	}

	// Header field should be bound
	if params.Header != "Bearer token" {
		t.Errorf("Expected Header to be 'Bearer token', got '%s'", params.Header)
	}
}

func TestBindingMeta_CacheConsistency(t *testing.T) {
	// Test that binding metadata is cached and reused correctly
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("GET", "/test1?name=john", nil)
	ctx1, cancel1 := request.NewContext(w1, r1)
	defer cancel1()

	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/test2?name=jane", nil)
	ctx2, cancel2 := request.NewContext(w2, r2)
	defer cancel2()

	// Bind same struct type multiple times
	var params1, params2 TaggedStruct

	err1 := ctx1.BindQuery(&params1)
	err2 := ctx2.BindQuery(&params2)

	// Both should succeed
	if err1 != nil {
		t.Errorf("Expected no error for first binding, got: %v", err1)
	}

	if err2 != nil {
		t.Errorf("Expected no error for second binding, got: %v", err2)
	}

	// Values should be different but binding should work consistently
	if params1.QueryParam != "john" {
		t.Errorf("Expected first QueryParam to be 'john', got '%s'", params1.QueryParam)
	}

	if params2.QueryParam != "jane" {
		t.Errorf("Expected second QueryParam to be 'jane', got '%s'", params2.QueryParam)
	}
}

func TestBindingMeta_SliceDetection(t *testing.T) {
	// Test that slice fields are correctly detected and handled
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?tags=go&tags=web&tags=api", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	type SliceTestStruct struct {
		Tags     []string `query:"tags"`
		Numbers  []int    `query:"numbers"`
		Booleans []bool   `query:"booleans"`
	}

	var params SliceTestStruct
	err := ctx.BindQuery(&params)

	// Should handle multiple values for slice fields
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedTags := []string{"go", "web", "api"}
	if len(params.Tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(params.Tags))
	}

	for i, tag := range params.Tags {
		if tag != expectedTags[i] {
			t.Errorf("Expected tag %d to be '%s', got '%s'", i, expectedTags[i], tag)
		}
	}
}

func TestBindingMeta_MapDetection(t *testing.T) {
	// Test that map fields are correctly detected and handled
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?config[debug]=true&config[timeout]=30&config[host]=localhost", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	type MapTestStruct struct {
		Config map[string]string `query:"config"`
	}

	var params MapTestStruct
	err := ctx.BindQuery(&params)

	// Should handle map-style parameters
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(params.Config) != 3 {
		t.Errorf("Expected 3 config entries, got %d", len(params.Config))
	}

	expectedConfig := map[string]string{
		"debug":   "true",
		"timeout": "30",
		"host":    "localhost",
	}

	for key, expected := range expectedConfig {
		if actual, ok := params.Config[key]; !ok {
			t.Errorf("Expected config key '%s' to exist", key)
		} else if actual != expected {
			t.Errorf("Expected config[%s] to be '%s', got '%s'", key, expected, actual)
		}
	}
}

func TestBindingMeta_MultipleTagTypes(t *testing.T) {
	// Test struct with multiple tag types on same field (should use first valid one)
	type MultiTagStruct struct {
		// Only the first valid tag should be used
		Value string `path:"from_path" query:"from_query" header:"from_header"`
	}

	w := httptest.NewRecorder()
	body := `{}` // Add empty JSON body to avoid EOF
	r := httptest.NewRequest("GET", "/test?from_query=query_value", strings.NewReader(body))
	r.Header.Set("from_header", "header_value")
	r.SetPathValue("from_path", "path_value")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	var params MultiTagStruct
	err := ctx.BindAll(&params)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should use the first tag found (path in this case since tags are processed in order)
	if params.Value != "path_value" {
		t.Errorf("Expected Value to be 'path_value', got '%s'", params.Value)
	}
}

func TestBindingMeta_PointerTypes(t *testing.T) {
	// Test that binding works with pointer to struct
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?name=john", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	type SimpleStruct struct {
		Name string `query:"name"`
	}

	// Test with pointer to struct
	params := &SimpleStruct{}
	err := ctx.BindQuery(params)

	if err != nil {
		t.Errorf("Expected no error with pointer type, got: %v", err)
	}

	if params.Name != "john" {
		t.Errorf("Expected Name to be 'john', got '%s'", params.Name)
	}
}

func TestBindingMeta_NestedStructs(t *testing.T) {
	// Test that nested structs are not automatically handled (only top-level fields)
	type NestedStruct struct {
		Value string `query:"nested_value"`
	}

	type ParentStruct struct {
		Name   string       `query:"name"`
		Nested NestedStruct // No tag, should be ignored
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?name=parent&nested_value=ignored", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	var params ParentStruct
	err := ctx.BindQuery(&params)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Parent field should be bound
	if params.Name != "parent" {
		t.Errorf("Expected Name to be 'parent', got '%s'", params.Name)
	}

	// Nested field should remain empty (not bound)
	if params.Nested.Value != "" {
		t.Errorf("Expected nested Value to remain empty, got '%s'", params.Nested.Value)
	}
}

func TestBindingMeta_EmptyTagValues(t *testing.T) {
	// Test fields with empty tag values (should be ignored)
	type EmptyTagStruct struct {
		Field1 string `query:""`     // Empty tag value
		Field2 string `query:"name"` // Valid tag
		Field3 string // No tag
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?name=value&Field1=ignored&Field3=ignored", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	var params EmptyTagStruct
	err := ctx.BindQuery(&params)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Only Field2 should be bound
	if params.Field1 != "" {
		t.Errorf("Expected Field1 to remain empty, got '%s'", params.Field1)
	}

	if params.Field2 != "value" {
		t.Errorf("Expected Field2 to be 'value', got '%s'", params.Field2)
	}

	if params.Field3 != "" {
		t.Errorf("Expected Field3 to remain empty, got '%s'", params.Field3)
	}
}
