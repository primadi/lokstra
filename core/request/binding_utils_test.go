package request_test

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/primadi/lokstra/core/request"
)

// Mock structures for testing binding utilities
type TestStruct struct {
	StringField  string
	IntField     int
	BoolField    bool
	Float64Field float64
	SliceField   []string
}

type TestUnmarshalJSONStruct struct {
	Value string
}

func (t *TestUnmarshalJSONStruct) UnmarshalJSON(data []byte) error {
	// Simple implementation for testing
	t.Value = string(data)
	return nil
}

func TestBindingUtilities_ConvertAndSetField_String(t *testing.T) {
	// We need to access unexported functions, so we'll test them indirectly
	// through the public binding methods that use them

	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?name=john&age=25&active=true&score=95.5", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test with a struct that exercises different field types
	type TestParams struct {
		Name   string  `query:"name"`
		Age    int     `query:"age"`
		Active bool    `query:"active"`
		Score  float64 `query:"score"`
	}

	var params TestParams
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

	if params.Score != 95.5 {
		t.Errorf("Expected Score to be 95.5, got %f", params.Score)
	}
}

func TestBindingUtilities_ConvertAndSetField_Slice(t *testing.T) {
	// Test slice handling with comma-separated values
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?tags=go,web,api&numbers=1,2,3", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	type TestParams struct {
		Tags    []string `query:"tags"`
		Numbers []int    `query:"numbers"`
	}

	var params TestParams
	err := ctx.BindQuery(&params)

	// Assertions
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

	expectedNumbers := []int{1, 2, 3}
	if len(params.Numbers) != len(expectedNumbers) {
		t.Errorf("Expected %d numbers, got %d", len(expectedNumbers), len(params.Numbers))
	}

	for i, num := range params.Numbers {
		if num != expectedNumbers[i] {
			t.Errorf("Expected number %d to be %d, got %d", i, expectedNumbers[i], num)
		}
	}
}

func TestBindingUtilities_UnsupportedTypes(t *testing.T) {
	// Test with unsupported field types
	type UnsupportedParams struct {
		ComplexField complex64 `query:"complex"`
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?complex=1+2i", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	var params UnsupportedParams
	err := ctx.BindQuery(&params)

	// Should return error for unsupported type
	if err == nil {
		t.Error("Expected error for unsupported field type")
	}
}

func TestBindingUtilities_InvalidConversions(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		paramType   string
	}{
		{
			name:        "Invalid int",
			queryString: "?age=notanumber",
			paramType:   "int",
		},
		{
			name:        "Invalid bool",
			queryString: "?active=notabool",
			paramType:   "bool",
		},
		{
			name:        "Invalid float",
			queryString: "?score=notafloat",
			paramType:   "float",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test"+tt.queryString, nil)
			ctx, cancel := request.NewContext(w, r)
			defer cancel()

			switch tt.paramType {
			case "int":
				type IntParams struct {
					Age int `query:"age"`
				}
				var params IntParams
				err := ctx.BindQuery(&params)
				if err == nil {
					t.Error("Expected error for invalid int conversion")
				}

			case "bool":
				type BoolParams struct {
					Active bool `query:"active"`
				}
				var params BoolParams
				err := ctx.BindQuery(&params)
				if err == nil {
					t.Error("Expected error for invalid bool conversion")
				}

			case "float":
				type FloatParams struct {
					Score float64 `query:"score"`
				}
				var params FloatParams
				err := ctx.BindQuery(&params)
				if err == nil {
					t.Error("Expected error for invalid float conversion")
				}
			}
		})
	}
}

func TestBindingUtilities_EmptyValues(t *testing.T) {
	// Test with empty query parameters
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?name=&age=", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	type TestParams struct {
		Name string `query:"name"`
		Age  int    `query:"age"`
	}

	var params TestParams
	err := ctx.BindQuery(&params)

	// Should handle empty string gracefully, but fail on empty int
	if err == nil {
		t.Error("Expected error for empty int value")
	}

	// Name should be empty string (valid)
	if params.Name != "" {
		t.Errorf("Expected empty Name, got '%s'", params.Name)
	}
}

func TestBindingUtilities_SliceFromMultipleValues(t *testing.T) {
	// Test slice binding from multiple query parameters with same name
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

	type TestParams struct {
		Tags []string `query:"tags"`
	}

	var params TestParams
	err := ctx.BindQuery(&params)

	// Assertions
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

func TestBindingUtilities_IntegerTypes(t *testing.T) {
	// Test various integer types
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?int8=127&int16=32767&int32=2147483647&uint8=255&uint16=65535", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	type TestParams struct {
		Int8   int8   `query:"int8"`
		Int16  int16  `query:"int16"`
		Int32  int32  `query:"int32"`
		Uint8  uint8  `query:"uint8"`
		Uint16 uint16 `query:"uint16"`
	}

	var params TestParams
	err := ctx.BindQuery(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if params.Int8 != 127 {
		t.Errorf("Expected Int8 to be 127, got %d", params.Int8)
	}

	if params.Int16 != 32767 {
		t.Errorf("Expected Int16 to be 32767, got %d", params.Int16)
	}

	if params.Int32 != 2147483647 {
		t.Errorf("Expected Int32 to be 2147483647, got %d", params.Int32)
	}

	if params.Uint8 != 255 {
		t.Errorf("Expected Uint8 to be 255, got %d", params.Uint8)
	}

	if params.Uint16 != 65535 {
		t.Errorf("Expected Uint16 to be 65535, got %d", params.Uint16)
	}
}

func TestBindingUtilities_FloatTypes(t *testing.T) {
	// Test float32 and float64
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?float32=3.14&float64=2.718281828", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	type TestParams struct {
		Float32 float32 `query:"float32"`
		Float64 float64 `query:"float64"`
	}

	var params TestParams
	err := ctx.BindQuery(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if params.Float32 != 3.14 {
		t.Errorf("Expected Float32 to be 3.14, got %f", params.Float32)
	}

	if params.Float64 != 2.718281828 {
		t.Errorf("Expected Float64 to be 2.718281828, got %f", params.Float64)
	}
}

func TestBindingUtilities_BooleanValues(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"1", "1", true},
		{"0", "0", false},
		{"True", "True", true},
		{"False", "False", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test?active="+tt.value, nil)
			ctx, cancel := request.NewContext(w, r)
			defer cancel()

			type TestParams struct {
				Active bool `query:"active"`
			}

			var params TestParams
			err := ctx.BindQuery(&params)

			if err != nil {
				t.Errorf("Expected no error for value '%s', got: %v", tt.value, err)
			}

			if params.Active != tt.expected {
				t.Errorf("Expected Active to be %t for value '%s', got %t", tt.expected, tt.value, params.Active)
			}
		})
	}
}

func TestBindingUtilities_CommaSeparatedValues(t *testing.T) {
	// Test comma-separated values with various whitespace
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?tags=go%2Cweb%2C+api+%2C+microservice%2C++docker++", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	type TestParams struct {
		Tags []string `query:"tags"`
	}

	var params TestParams
	err := ctx.BindQuery(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should trim whitespace from each tag
	expectedTags := []string{"go", "web", "api", "microservice", "docker"}
	if len(params.Tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(params.Tags))
	}

	for i, tag := range params.Tags {
		if tag != expectedTags[i] {
			t.Errorf("Expected tag %d to be '%s', got '%s'", i, expectedTags[i], tag)
		}
	}
}

func TestBindingUtilities_EmptyCommaSeparated(t *testing.T) {
	// Test comma-separated values with empty elements
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?tags=go,,web,,,api", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	type TestParams struct {
		Tags []string `query:"tags"`
	}

	var params TestParams
	err := ctx.BindQuery(&params)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should skip empty elements
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
