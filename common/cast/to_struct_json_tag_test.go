package cast_test

import (
	"testing"

	"github.com/primadi/lokstra/common/cast"
)

func TestToStruct_JSONTagRequirement(t *testing.T) {
	// Test 1: Struct WITHOUT json tags (using field names directly)
	type ConfigNoTag struct {
		Name string
		Port int
	}

	t.Run("without json tag - using field name", func(t *testing.T) {
		input := map[string]any{
			"Name": "myapp", // Must match field name exactly
			"Port": 8080,
		}

		var result ConfigNoTag
		err := cast.ToStruct(input, &result, false)
		if err != nil {
			t.Fatalf("ToStruct failed: %v", err)
		}

		if result.Name != "myapp" {
			t.Errorf("Name = %v, want myapp", result.Name)
		}
		if result.Port != 8080 {
			t.Errorf("Port = %v, want 8080", result.Port)
		}
	})

	// Test 2: Struct WITH json tags
	type ConfigWithTag struct {
		Name string `json:"app_name"`
		Port int    `json:"server_port"`
	}

	t.Run("with json tag", func(t *testing.T) {
		input := map[string]any{
			"app_name":    "myapp", // Uses json tag
			"server_port": 8080,
		}

		var result ConfigWithTag
		err := cast.ToStruct(input, &result, false)
		if err != nil {
			t.Fatalf("ToStruct failed: %v", err)
		}

		if result.Name != "myapp" {
			t.Errorf("Name = %v, want myapp", result.Name)
		}
		if result.Port != 8080 {
			t.Errorf("Port = %v, want 8080", result.Port)
		}
	})

	// Test 3: json tag ignored field
	type ConfigIgnored struct {
		Name   string `json:"name"`
		Secret string `json:"-"` // Ignored
		Port   int    `json:"port"`
	}

	t.Run("json tag with ignored field", func(t *testing.T) {
		input := map[string]any{
			"name":   "myapp",
			"Secret": "should-be-ignored",
			"port":   8080,
		}

		var result ConfigIgnored
		err := cast.ToStruct(input, &result, false)
		if err != nil {
			t.Fatalf("ToStruct failed: %v", err)
		}

		if result.Name != "myapp" {
			t.Errorf("Name = %v, want myapp", result.Name)
		}
		if result.Secret != "" {
			t.Errorf("Secret = %v, want empty (should be ignored)", result.Secret)
		}
		if result.Port != 8080 {
			t.Errorf("Port = %v, want 8080", result.Port)
		}
	})

	// Test 4: Mixed - some with tags, some without
	type ConfigMixed struct {
		Name string `json:"app_name"`
		Port int    // No tag, uses field name
		Host string // No tag, uses field name
	}

	t.Run("mixed json tags", func(t *testing.T) {
		input := map[string]any{
			"app_name": "myapp",     // Uses json tag
			"Port":     8080,        // Uses field name
			"Host":     "localhost", // Uses field name
		}

		var result ConfigMixed
		err := cast.ToStruct(input, &result, false)
		if err != nil {
			t.Fatalf("ToStruct failed: %v", err)
		}

		if result.Name != "myapp" {
			t.Errorf("Name = %v, want myapp", result.Name)
		}
		if result.Port != 8080 {
			t.Errorf("Port = %v, want 8080", result.Port)
		}
		if result.Host != "localhost" {
			t.Errorf("Host = %v, want localhost", result.Host)
		}
	})
}
