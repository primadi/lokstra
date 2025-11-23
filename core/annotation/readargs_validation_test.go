package annotation_test

import (
	"testing"

	"github.com/primadi/lokstra/core/annotation"
)

func TestReadArgsValidation(t *testing.T) {
	tests := []struct {
		name          string
		annotation    *annotation.ParsedAnnotation
		expectedArgs  []string
		shouldError   bool
		errorContains string
	}{
		{
			name: "valid named args",
			annotation: &annotation.ParsedAnnotation{
				Args: map[string]interface{}{
					"route":       "GET /users",
					"middlewares": []string{"auth"},
				},
			},
			expectedArgs: []string{"route", "middlewares"},
			shouldError:  false,
		},
		{
			name: "invalid named arg",
			annotation: &annotation.ParsedAnnotation{
				Args: map[string]interface{}{
					"invalid": "value",
				},
			},
			expectedArgs:  []string{"route", "middlewares"},
			shouldError:   true,
			errorContains: "unexpected argument 'invalid'",
		},
		{
			name: "valid positional args",
			annotation: &annotation.ParsedAnnotation{
				PositionalArgs: []interface{}{"GET /users", []string{"auth"}},
			},
			expectedArgs: []string{"route", "middlewares"},
			shouldError:  false,
		},
		{
			name: "too many positional args",
			annotation: &annotation.ParsedAnnotation{
				PositionalArgs: []interface{}{"GET /users", []string{"auth"}, "extra"},
			},
			expectedArgs:  []string{"route", "middlewares"},
			shouldError:   true,
			errorContains: "too many arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.annotation.ReadArgs(tt.expectedArgs...)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("expected result but got nil")
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)*2 && s[len(s)/2-len(substr)/2:len(s)/2+len(substr)/2+len(substr)%2] == substr ||
			containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
