package flow

import (
	"testing"
)

func TestValidationRules(t *testing.T) {
	tests := []struct {
		name     string
		rule     ValidationRule
		value    any
		expected bool
		message  string
	}{
		// Required tests
		{"Required with string", Required(), "hello", true, ""},
		{"Required with empty string", Required(), "", false, "is required"},
		{"Required with nil", Required(), nil, false, "is required"},
		{"Required with zero int", Required(), 0, false, "is required"},
		{"Required with non-zero int", Required(), 42, true, ""},

		// MinLength tests
		{"MinLength valid", MinLength(3), "hello", true, ""},
		{"MinLength invalid", MinLength(5), "hi", false, "must be at least 5 characters"},
		{"MinLength exact", MinLength(3), "abc", true, ""},
		{"MinLength non-string", MinLength(3), 123, false, "must be a string"},

		// Email tests
		{"Email valid", Email(), "test@example.com", true, ""},
		{"Email invalid no @", Email(), "testexample.com", false, "must be a valid email address"},
		{"Email invalid no domain", Email(), "test@", false, "must be a valid email address"},
		{"Email empty", Email(), "", false, "must be a valid email address"},
		{"Email non-string", Email(), 123, false, "must be a string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, message := tt.rule(tt.value)
			if valid != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, valid)
			}
			if message != tt.message {
				t.Errorf("Expected message '%s', got '%s'", tt.message, message)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidationRules(b *testing.B) {
	rules := []ValidationRule{Required(), MinLength(5), Email()}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, rule := range rules {
			rule("test@example.com")
		}
	}
}
