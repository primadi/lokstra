package config_test

import (
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/config"
)

func TestExpandVariablesWithCFG(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "CFG resolver with configs section",
			input: `
configs:
  - name: database.host
    value: postgres.example.com
  - name: database.port
    value: 5432
  - name: api.baseUrl
    value: https://api.example.com

services:
  - name: postgres
    config:
      host: ${@CFG:database.host}
      port: ${@CFG:database.port}
  - name: api
    config:
      baseUrl: ${@CFG:api.baseUrl}
`,
			expected: `
configs:
  - name: database.host
    value: postgres.example.com
  - name: database.port
    value: 5432
  - name: api.baseUrl
    value: https://api.example.com

services:
  - name: postgres
    config:
      host: postgres.example.com
      port: 5432
  - name: api
    config:
      baseUrl: https://api.example.com
`,
		},
		{
			name: "CFG resolver with default value",
			input: `
configs:
  - name: existing.key
    value: exists

services:
  - name: test
    config:
      existing: ${@CFG:existing.key}
      missing: ${@CFG:missing.key:default-value}
`,
			expected: `
configs:
  - name: existing.key
    value: exists

services:
  - name: test
    config:
      existing: exists
      missing: default-value
`,
		},
		{
			name: "No configs section - CFG placeholders remain",
			input: `
services:
  - name: test
    config:
      value: ${@CFG:some.key:fallback}
`,
			expected: `
services:
  - name: test
    config:
      value: ${@CFG:some.key:fallback}
`,
		},
		{
			name: "Mixed resolvers - ENV and CFG",
			input: `
configs:
  - name: app.name
    value: MyApp

services:
  - name: test
    config:
      appName: ${@CFG:app.name}
      port: ${PORT:8080}
`,
			expected: `
configs:
  - name: app.name
    value: MyApp

services:
  - name: test
    config:
      appName: MyApp
      port: 8080
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.ExpandVariables(tt.input)

			// For debugging
			if result != tt.expected {
				t.Logf("Input:\n%s", tt.input)
				t.Logf("Expected:\n%s", tt.expected)
				t.Logf("Got:\n%s", result)
			}

			// Compare trimmed strings (ignore leading/trailing whitespace)
			if strings.TrimSpace(result) != strings.TrimSpace(tt.expected) {
				t.Errorf("expandVariables() mismatch\nGot:\n%s\nWant:\n%s", result, tt.expected)
			}
		})
	}
}

func TestExpandVariablesWithCFGComplexValues(t *testing.T) {
	input := `
configs:
  - name: database.dsn
    value: postgresql://user:pass@localhost:5432/db
  - name: redis.url
    value: redis://:password@localhost:6379/0

services:
  - name: postgres
    config:
      dsn: ${@CFG:database.dsn}
  - name: redis
    config:
      url: ${@CFG:redis.url}
`

	result := config.ExpandVariables(input)

	if !contains(result, "postgresql://user:pass@localhost:5432/db") {
		t.Errorf("Expected expanded DSN with colons, got:\n%s", result)
	}

	if !contains(result, "redis://:password@localhost:6379/0") {
		t.Errorf("Expected expanded Redis URL with colons, got:\n%s", result)
	}
}

func contains(s, substr string) bool {
	return findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
