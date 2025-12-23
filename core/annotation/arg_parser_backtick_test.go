package annotation

import (
	"testing"
)

func TestParseAnnotationArgs_Backtick(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectNamed      map[string]any
		expectPositional []any
	}{
		{
			name:  "backtick simple string",
			input: "key=`myvalue`",
			expectNamed: map[string]any{
				"key": "myvalue",
			},
			expectPositional: nil,
		},
		{
			name:  "backtick with double quotes inside",
			input: "default=`Config{Name: \"myapp\", Port: 8080}`",
			expectNamed: map[string]any{
				"default": `Config{Name: "myapp", Port: 8080}`,
			},
			expectPositional: nil,
		},
		{
			name:        "backtick positional arg",
			input:       "`POST /users/{id}`",
			expectNamed: map[string]any{},
			expectPositional: []any{
				"POST /users/{id}",
			},
		},
		{
			name:  "mixed backtick and double quote",
			input: "name=\"user-service\", default=`Config{Name: \"app\"}`",
			expectNamed: map[string]any{
				"name":    "user-service",
				"default": `Config{Name: "app"}`,
			},
			expectPositional: nil,
		},
		{
			name:  "backtick with newlines",
			input: "sql=`SELECT * FROM users WHERE name = \"john\"`",
			expectNamed: map[string]any{
				"sql": `SELECT * FROM users WHERE name = "john"`,
			},
			expectPositional: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			named, positional, err := ParseAnnotationArgs(tt.input)
			if err != nil {
				t.Fatalf("ParseAnnotationArgs failed: %v", err)
			}

			// Check named args
			if len(named) != len(tt.expectNamed) {
				t.Errorf("Named args count = %d, want %d", len(named), len(tt.expectNamed))
			}
			for key, expectedVal := range tt.expectNamed {
				if actualVal, ok := named[key]; !ok {
					t.Errorf("Missing named arg: %s", key)
				} else if actualVal != expectedVal {
					t.Errorf("Named arg %s = %v, want %v", key, actualVal, expectedVal)
				}
			}

			// Check positional args
			if len(positional) != len(tt.expectPositional) {
				t.Errorf("Positional args count = %d, want %d", len(positional), len(tt.expectPositional))
			}
			for i, expectedVal := range tt.expectPositional {
				if i >= len(positional) {
					break
				}
				if positional[i] != expectedVal {
					t.Errorf("Positional arg[%d] = %v, want %v", i, positional[i], expectedVal)
				}
			}
		})
	}
}

func TestSmartSplit_Backtick(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		delim  rune
		expect []string
	}{
		{
			name:  "split with backtick string",
			input: "name=`value`, port=8080",
			delim: ',',
			expect: []string{
				"name=`value`",
				" port=8080",
			},
		},
		{
			name:  "backtick with comma inside",
			input: "sql=`SELECT a, b, c FROM users`, limit=10",
			delim: ',',
			expect: []string{
				"sql=`SELECT a, b, c FROM users`",
				" limit=10",
			},
		},
		{
			name:  "mixed quotes",
			input: "a=\"value1\", b=`value2`, c='value3'",
			delim: ',',
			expect: []string{
				"a=\"value1\"",
				" b=`value2`",
				" c='value3'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SmartSplit(tt.input, tt.delim)
			if len(result) != len(tt.expect) {
				t.Errorf("Split result count = %d, want %d", len(result), len(tt.expect))
				t.Logf("Got: %v", result)
				t.Logf("Want: %v", tt.expect)
				return
			}

			for i, expected := range tt.expect {
				if result[i] != expected {
					t.Errorf("Split[%d] = %q, want %q", i, result[i], expected)
				}
			}
		})
	}
}
