package annotation_test

import (
	"testing"

	"github.com/primadi/lokstra/core/annotation"
)

func TestParseArrayWithCommaInString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple array",
			input:    `["mw1", "mw2"]`,
			expected: []string{"mw1", "mw2"},
		},
		{
			name:     "array with comma in string",
			input:    `["mw-test param1=123, param2=abc"]`,
			expected: []string{"mw-test param1=123, param2=abc"},
		},
		{
			name:     "mixed array with comma in string",
			input:    `["mw-test param1=123, param2=abc", "mw2"]`,
			expected: []string{"mw-test param1=123, param2=abc", "mw2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := annotation.ParseArrayValue(tt.input)
			if err != nil {
				t.Fatalf("ParseArrayValue() error = %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("ParseArrayValue() got %d items, want %d items\nGot: %v\nWant: %v",
					len(result), len(tt.expected), result, tt.expected)
			}

			for i, val := range result {
				if val != tt.expected[i] {
					t.Errorf("ParseArrayValue()[%d] = %q, want %q", i, val, tt.expected[i])
				}
			}
		})
	}
}

func TestSmartSplitDebug(t *testing.T) {
	input := `"GET /users/{id}", ["mw-test param1=123, param2=abc"]`

	t.Logf("Input: %q", input)
	result := annotation.SmartSplit(input, ',')

	t.Logf("Got %d parts:", len(result))
	for i, part := range result {
		t.Logf("  [%d] %q", i, part)
	}

	// Expected: 2 parts
	// Part 0: "GET /users/{id}"
	// Part 1: ["mw-test param1=123, param2=abc"]

	if len(result) != 2 {
		t.Fatalf("Expected 2 parts, got %d", len(result))
	}
}

func TestParseAnnotationArgs(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedNamed      map[string]interface{}
		expectedPositional []interface{}
	}{
		{
			name:          "route with middleware array containing comma",
			input:         `"GET /users/{id}", ["mw-test param1=123, param2=abc"]`,
			expectedNamed: map[string]interface{}{},
			expectedPositional: []interface{}{
				"GET /users/{id}",
				[]string{"mw-test param1=123, param2=abc"},
			},
		},
		{
			name:          "route with multiple middlewares",
			input:         `"GET /users", ["mw1", "mw2"]`,
			expectedNamed: map[string]interface{}{},
			expectedPositional: []interface{}{
				"GET /users",
				[]string{"mw1", "mw2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, positional, err := annotation.ParseAnnotationArgs(tt.input)
			if err != nil {
				t.Fatalf("ParseAnnotationArgs() error = %v", err)
			}

			t.Logf("Input: %q", tt.input)
			t.Logf("Got %d positional args:", len(positional))
			for i, val := range positional {
				t.Logf("  [%d] type=%T value=%#v", i, val, val)
			}

			// Check positional args
			if len(positional) != len(tt.expectedPositional) {
				t.Fatalf("positional args: got %d items, want %d items\nGot: %#v\nWant: %#v",
					len(positional), len(tt.expectedPositional), positional, tt.expectedPositional)
			}

			for i, val := range positional {
				expected := tt.expectedPositional[i]

				// Compare based on type
				switch exp := expected.(type) {
				case string:
					if val != exp {
						t.Errorf("positional[%d] = %q, want %q", i, val, exp)
					}
				case []string:
					arr, ok := val.([]string)
					if !ok {
						t.Errorf("positional[%d] is not []string, got %T", i, val)
						continue
					}
					if len(arr) != len(exp) {
						t.Errorf("positional[%d] has %d items, want %d items", i, len(arr), len(exp))
						continue
					}
					for j, item := range arr {
						if item != exp[j] {
							t.Errorf("positional[%d][%d] = %q, want %q", i, j, item, exp[j])
						}
					}
				}
			}
		})
	}
}
