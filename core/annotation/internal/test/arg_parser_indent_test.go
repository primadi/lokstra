package annotation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/primadi/lokstra/core/annotation"
)

// TestParseFileAnnotations_IgnoreIndentedAnnotations tests that indented annotations
// (typically in documentation examples) are correctly ignored
func TestParseFileAnnotations_IgnoreIndentedAnnotations(t *testing.T) {
	// Create temp test file with indented annotation in doc comment
	// NOTE: The line "//\t@RouterService" uses actual TAB character for Go doc code example format
	content := "package middleware\n\n" +
		"import (\n" +
		"\t\"github.com/primadi/lokstra/lokstra_registry\"\n" +
		")\n\n" +
		"// Register is a placeholder for middleware registration\n" +
		"// TODO: Implement actual middleware registration when Lokstra framework supports it\n" +
		"//\n" +
		"// For now, middlewares should be applied manually in route setup or\n" +
		"// specified in @RouterService annotations for documentation purposes.\n" +
		"//\n" +
		"// Example @RouterService annotation:\n" +
		"//\n" +
		"//\t@RouterService name=\"tenant-service\", prefix=\"/api/tenants\", middlewares=[\"recovery\", \"request_logger\", \"auth\"]\n" +
		"//\n" +
		"// The \"auth\" middleware indicates that the endpoint requires authentication.\n" +
		"// Actual middleware implementation should be set up in your main.go or route configuration.\n" +
		"func Register() {\n" +
		"\tauthMw := NewAuthMiddleware(AuthMiddlewareConfig{})\n" +
		"\tlokstra_registry.RegisterMiddleware(\"auth\", authMw.Handler())\n" +
		"}\n"

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "register.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Parse annotations
	annotations, err := annotation.ParseFileAnnotations(filePath)
	if err != nil {
		t.Fatalf("ParseFileAnnotations() error = %v", err)
	}

	// Should find NO annotations (the @RouterService is indented with TAB, so should be ignored)
	if len(annotations) != 0 {
		t.Errorf("Expected 0 annotations (indented should be ignored), got %d", len(annotations))
		for _, ann := range annotations {
			t.Logf("  Found: @%s on line %d, target: %s", ann.Name, ann.Line, ann.TargetName)
		}
	}
}

// TestParseFileAnnotations_ValidAnnotations tests that non-indented annotations are correctly detected
func TestParseFileAnnotations_ValidAnnotations(t *testing.T) {
	content := `package application

// @RouterService name="user-service", prefix="/api/users"
type UserService struct {
	// @Inject "user-repository"
	UserRepo UserRepository
}

// @Route "GET /{id}"
func (s *UserService) GetByID(p *GetUserParams) (*User, error) {
	return s.UserRepo.GetByID(p.ID)
}
`

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "user_service.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	annotations, err := annotation.ParseFileAnnotations(filePath)
	if err != nil {
		t.Fatalf("ParseFileAnnotations() error = %v", err)
	}

	// Should find 3 annotations: @RouterService, @Inject, @Route
	if len(annotations) != 3 {
		t.Errorf("Expected 3 annotations, got %d", len(annotations))
		for _, ann := range annotations {
			t.Logf("  Found: @%s on line %d, target: %s", ann.Name, ann.Line, ann.TargetName)
		}
	}

	// Verify annotations
	expectedAnnotations := map[string]string{
		"RouterService": "UserService",
		"Inject":        "UserRepo",
		"Route":         "GetByID",
	}

	foundAnnotations := make(map[string]string)
	for _, ann := range annotations {
		foundAnnotations[ann.Name] = ann.TargetName
	}

	for name, expectedTarget := range expectedAnnotations {
		if target, found := foundAnnotations[name]; !found {
			t.Errorf("Expected annotation @%s not found", name)
		} else if target != expectedTarget {
			t.Errorf("Annotation @%s: expected target %s, got %s", name, expectedTarget, target)
		}
	}
}

// TestParseFileAnnotations_MultipleEmptyLinesAfterAnnotation tests that annotations
// with too many empty comment lines are discarded
func TestParseFileAnnotations_MultipleEmptyLinesAfterAnnotation(t *testing.T) {
	content := `package test

// @RouterService name="test-service"
//
//
//
//
// This is a very long documentation
// that spans multiple empty lines
// after the annotation
type TestService struct {}
`

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_service.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	annotations, err := annotation.ParseFileAnnotations(filePath)
	if err != nil {
		t.Fatalf("ParseFileAnnotations() error = %v", err)
	}

	// Should find NO annotations (too many empty lines after annotation)
	if len(annotations) != 0 {
		t.Errorf("Expected 0 annotations (too many empty lines), got %d", len(annotations))
		for _, ann := range annotations {
			t.Logf("  Found: @%s on line %d, target: %s", ann.Name, ann.Line, ann.TargetName)
		}
	}
}

// TestParseFileAnnotations_AnnotationWithFewEmptyLines tests that annotations
// with few empty comment lines (<=3) are still valid
func TestParseFileAnnotations_AnnotationWithFewEmptyLines(t *testing.T) {
	content := `package test

// @RouterService name="test-service"
//
// Some documentation
type TestService struct {}
`

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_service.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	annotations, err := annotation.ParseFileAnnotations(filePath)
	if err != nil {
		t.Fatalf("ParseFileAnnotations() error = %v", err)
	}

	// Should find 1 annotation (few empty lines are OK)
	if len(annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(annotations))
	} else {
		ann := annotations[0]
		if ann.Name != "RouterService" {
			t.Errorf("Expected @RouterService, got @%s", ann.Name)
		}
		if ann.TargetName != "TestService" {
			t.Errorf("Expected target TestService, got %s", ann.TargetName)
		}
	}
}
