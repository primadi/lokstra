package annotation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/primadi/lokstra/core/annotation"
)

// TestRouterServiceValidation_MustBeOnStruct tests that @RouterService
// annotation must be placed above a struct declaration, not a function
func TestRouterServiceValidation_MustBeOnStruct(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid - struct",
			content: `package application

// @RouterService name="user-service", prefix="/api/users"
type UserService struct {
	UserRepo UserRepository
}

type UserRepository interface {
	GetByID(id string) (*User, error)
}

type User struct {
	ID string
}
`,
			expectError: false,
		},
		{
			name: "invalid - function",
			content: `package application

// @RouterService name="register-service", prefix="/api/register"
func Register() {
	// This should fail validation
}
`,
			expectError: true,
			errorMsg:    "must be placed directly above a struct declaration, found 'Register' instead",
		},
		{
			name: "invalid - interface",
			content: `package application

// @RouterService name="user-service", prefix="/api/users"
type UserService interface {
	GetByID(id string) error
}
`,
			expectError: true,
			errorMsg:    "must be placed directly above a struct declaration, found 'UserService' instead",
		},
		{
			name: "invalid - type alias",
			content: `package application

// @RouterService name="user-service", prefix="/api/users"
type UserService = string
`,
			expectError: true,
			errorMsg:    "must be placed directly above a struct declaration, found 'UserService' instead",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and file
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.go")

			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Parse annotations
			annotations, err := annotation.ParseFileAnnotations(filePath)
			if err != nil {
				t.Fatalf("ParseFileAnnotations() error = %v", err)
			}

			// Create context for code generation
			ctx := &annotation.RouterServiceContext{
				FolderPath: tmpDir,
				UpdatedFiles: []*annotation.FileToProcess{
					{
						Filename:    "test.go",
						FullPath:    filePath,
						Annotations: annotations,
					},
				},
				SkippedFiles: []*annotation.FileToProcess{},
				DeletedFiles: []string{},
				GeneratedCode: &annotation.GeneratedCode{
					Services:          make(map[string]*annotation.ServiceGeneration),
					PreservedSections: make(map[string]string),
				},
			}

			// Try to generate code
			err = annotation.GenerateCodeForFolder(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
