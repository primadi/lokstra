package annotation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/annotation"
	"github.com/primadi/lokstra/core/annotation/internal"
)

// TestUnusedImportsNotIncluded verifies that imports from source file
// that are not used in handler methods/dependencies are not included in generated code
func TestUnusedImportsNotIncluded(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a service file with imports that are NOT used in handlers
	serviceCode := `package application

import (
	"github.com/primadi/lokstra-auth/credential/domain"  // NOT USED in handlers
	"github.com/primadi/lokstra/core/request"              // NOT USED in handlers
	core_repository "github.com/primadi/lokstra-auth/infrastructure/repository"  // USED in dependency
)

// @EndpointService name="test-service", prefix="/api"
type TestService struct {
	// @Inject "user-repository"
	Repo core_repository.UserRepository
}

// @Route "GET /users/{id}"
// Returns string, no domain types used
func (s *TestService) GetUser(id string) (string, error) {
	// Method doesn't use domain or request packages
	return "user-" + id, nil
}

// @Route "POST /users"
// Takes simple struct, no domain types
func (s *TestService) CreateUser(p *CreateUserParams) (string, error) {
	return "created", nil
}

type CreateUserParams struct {
	Name string
}
`

	// Write service file
	servicePath := filepath.Join(tmpDir, "test_service.go")
	if err := os.WriteFile(servicePath, []byte(serviceCode), 0644); err != nil {
		t.Fatalf("Failed to create service file: %v", err)
	}

	// Parse annotations
	annotations, err := annotation.ParseFileAnnotations(servicePath)
	if err != nil {
		t.Fatalf("ParseFileAnnotations failed: %v", err)
	}

	// Create context
	ctx := &annotation.RouterServiceContext{
		FolderPath: tmpDir,
		UpdatedFiles: []*annotation.FileToProcess{
			{
				Filename:    "test_service.go",
				FullPath:    servicePath,
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

	// Generate code
	if err := annotation.GenerateCodeForFolder(ctx); err != nil {
		t.Fatalf("GenerateCodeForFolder failed: %v", err)
	}

	// Read generated file
	genPath := filepath.Join(tmpDir, internal.GeneratedFileName)
	content, err := os.ReadFile(genPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	genCode := string(content)

	// Verify that unused imports are NOT included
	if strings.Contains(genCode, `"github.com/primadi/lokstra-auth/credential/domain"`) {
		t.Error("Generated code should NOT import unused package 'domain'")
	}

	if strings.Contains(genCode, `"github.com/primadi/lokstra/core/request"`) {
		t.Error("Generated code should NOT import unused package 'request'")
	}

	// Verify that used import IS included
	if !strings.Contains(genCode, `"github.com/primadi/lokstra-auth/infrastructure/repository"`) {
		t.Error("Generated code should import used package 'core_repository'")
	}

	// Verify core lokstra imports are included
	if !strings.Contains(genCode, `"github.com/primadi/lokstra/lokstra_registry"`) {
		t.Error("Generated code should import lokstra_registry")
	}

	// Should include deploy and proxy for @EndpointService
	if !strings.Contains(genCode, `"github.com/primadi/lokstra/core/deploy"`) {
		t.Error("Generated code should import deploy for @EndpointService")
	}

	if !strings.Contains(genCode, `"github.com/primadi/lokstra/core/proxy"`) {
		t.Error("Generated code should import proxy for @EndpointService")
	}

	t.Logf("✅ Generated code correctly filtered unused imports")
	t.Logf("Generated file:\n%s", genCode)
}

// TestOnlyMethodTypesIncluded verifies that only types used in method signatures
// are considered when filtering imports
func TestOnlyMethodTypesIncluded(t *testing.T) {
	tmpDir := t.TempDir()

	// Create domain package
	domainDir := filepath.Join(tmpDir, "domain")
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		t.Fatalf("Failed to create domain dir: %v", err)
	}

	domainCode := `package domain

type User struct {
	ID   string
	Name string
}
`
	if err := os.WriteFile(filepath.Join(domainDir, "models.go"), []byte(domainCode), 0644); err != nil {
		t.Fatalf("Failed to create domain models: %v", err)
	}

	// Create helper package
	helperDir := filepath.Join(tmpDir, "helper")
	if err := os.MkdirAll(helperDir, 0755); err != nil {
		t.Fatalf("Failed to create helper dir: %v", err)
	}

	helperCode := `package helper

func Query(q string) string {
	return ""
}
`
	if err := os.WriteFile(filepath.Join(helperDir, "helper.go"), []byte(helperCode), 0644); err != nil {
		t.Fatalf("Failed to create helper: %v", err)
	}

	// Create application package
	appDir := filepath.Join(tmpDir, "application")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("Failed to create application dir: %v", err)
	}

	// Create a service file with helper types not used in handlers
	serviceCode := `package application

import (
	"testapp/domain"  // Used in handler
	"testapp/helper"  // NOT used in handler, only in helper method
)

// @EndpointService name="test-service", prefix="/api"
type TestService struct {
}

// @Route "GET /users"
func (s *TestService) GetUsers() ([]*domain.User, error) {
	users := s.fetchFromDB()  // Uses helper internally
	return users, nil
}

// Helper method (not a handler)
func (s *TestService) fetchFromDB() []*domain.User {
	// This uses helper.Query but this method is not a handler
	_ = helper.Query("SELECT * FROM users")
	return nil
}
`

	// Write service file
	servicePath := filepath.Join(appDir, "test_service.go")
	if err := os.WriteFile(servicePath, []byte(serviceCode), 0644); err != nil {
		t.Fatalf("Failed to create service file: %v", err)
	}

	// Create go.mod for module resolution
	goModContent := `module testapp

go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Parse annotations
	annotations, err := annotation.ParseFileAnnotations(servicePath)
	if err != nil {
		t.Fatalf("ParseFileAnnotations failed: %v", err)
	}

	// Create context
	ctx := &annotation.RouterServiceContext{
		FolderPath: appDir,
		UpdatedFiles: []*annotation.FileToProcess{
			{
				Filename:    "test_service.go",
				FullPath:    servicePath,
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

	// Generate code
	if err := annotation.GenerateCodeForFolder(ctx); err != nil {
		t.Fatalf("GenerateCodeForFolder failed: %v", err)
	}

	// Read generated file
	genPath := filepath.Join(appDir, internal.GeneratedFileName)
	content, err := os.ReadFile(genPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	genCode := string(content)

	// Verify that domain (used in handler) IS included
	if !strings.Contains(genCode, `"testapp/domain"`) {
		t.Error("Generated code should import 'domain' used in handler")
		t.Logf("Generated imports:\n%s", genCode[:strings.Index(genCode, "// Auto-register")])
	}

	// Verify that helper (NOT used in handler signature) is NOT included
	if strings.Contains(genCode, `"testapp/helper"`) {
		t.Error("Generated code should NOT import 'helper' - only used in non-handler methods")
	}

	t.Logf("✅ Generated code only includes types from handler signatures")
}
