package annotation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/annotation"
	"github.com/primadi/lokstra/core/annotation/internal"
)

// TestImportAlias_DifferentPathsSameAlias tests that when two different import paths
// use the same alias, the system automatically renames one of them to avoid conflict.
// Example:
//
//	service_a.go: import models "pkga"
//	service_b.go: import models "pkgb"
//
// Expected: One should be renamed to "models_1" or similar
func TestImportAlias_DifferentPathsSameAlias(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package A
	pkgaDir := filepath.Join(tmpDir, "pkga")
	if err := os.MkdirAll(pkgaDir, 0755); err != nil {
		t.Fatalf("Failed to create pkga dir: %v", err)
	}

	pkgaCode := `package pkga

type User struct {
	ID   string
	Name string
}
`
	if err := os.WriteFile(filepath.Join(pkgaDir, "models.go"), []byte(pkgaCode), 0644); err != nil {
		t.Fatalf("Failed to create pkga models: %v", err)
	}

	// Create package B
	pkgbDir := filepath.Join(tmpDir, "pkgb")
	if err := os.MkdirAll(pkgbDir, 0755); err != nil {
		t.Fatalf("Failed to create pkgb dir: %v", err)
	}

	pkgbCode := `package pkgb

type User struct {
	UserID   string
	FullName string
}
`
	if err := os.WriteFile(filepath.Join(pkgbDir, "models.go"), []byte(pkgbCode), 0644); err != nil {
		t.Fatalf("Failed to create pkgb models: %v", err)
	}

	// Create service A using pkga with alias "models"
	serviceACode := `package main

import (
	models "myapp/pkga"
)

// @RouterService name="service-a", prefix="/api/a"
type ServiceA struct {}

// @Route "GET /users"
func (s *ServiceA) GetUsers() (*models.User, error) {
	return nil, nil
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "service_a.go"), []byte(serviceACode), 0644); err != nil {
		t.Fatalf("Failed to create service_a.go: %v", err)
	}

	// Create service B using pkgb with alias "models" (CONFLICT!)
	serviceBCode := `package main

import (
	models "myapp/pkgb"
)

// @RouterService name="service-b", prefix="/api/b"
type ServiceB struct {}

// @Route "GET /users"
func (s *ServiceB) GetUsers() (*models.User, error) {
	return nil, nil
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "service_b.go"), []byte(serviceBCode), 0644); err != nil {
		t.Fatalf("Failed to create service_b.go: %v", err)
	}

	// Create go.mod
	goModContent := `module myapp

go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Parse annotations for both services
	annotationsA, err := annotation.ParseFileAnnotations(filepath.Join(tmpDir, "service_a.go"))
	if err != nil {
		t.Fatalf("Failed to parse service_a annotations: %v", err)
	}

	annotationsB, err := annotation.ParseFileAnnotations(filepath.Join(tmpDir, "service_b.go"))
	if err != nil {
		t.Fatalf("Failed to parse service_b annotations: %v", err)
	}

	// Create context
	ctx := &annotation.RouterServiceContext{
		FolderPath: tmpDir,
		UpdatedFiles: []*annotation.FileToProcess{
			{
				Filename:    "service_a.go",
				FullPath:    filepath.Join(tmpDir, "service_a.go"),
				Annotations: annotationsA,
			},
			{
				Filename:    "service_b.go",
				FullPath:    filepath.Join(tmpDir, "service_b.go"),
				Annotations: annotationsB,
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
	genFilePath := filepath.Join(tmpDir, internal.GeneratedFileName)
	genContent, err := os.ReadFile(genFilePath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	genCode := string(genContent)

	// Verify that both imports exist but with different aliases
	hasPkgaImport := strings.Contains(genCode, `"myapp/pkga"`)
	hasPkgbImport := strings.Contains(genCode, `"myapp/pkgb"`)

	if !hasPkgaImport {
		t.Error("Expected pkga import in generated code")
	}
	if !hasPkgbImport {
		t.Error("Expected pkgb import in generated code")
	}

	// Count how many times "models" appears as an alias
	// At least one should be renamed (e.g., models_1)
	modelsCount := strings.Count(genCode, "models \"myapp/")
	models1Count := strings.Count(genCode, "models_1 \"myapp/")

	t.Logf("Found 'models' alias: %d times", modelsCount)
	t.Logf("Found 'models_1' alias: %d times", models1Count)

	// One should keep "models", another should be "models_1"
	if modelsCount != 1 {
		t.Errorf("Expected exactly 1 'models' alias, got %d", modelsCount)
	}
	if models1Count != 1 {
		t.Errorf("Expected exactly 1 'models_1' alias, got %d", models1Count)
	}

	t.Logf("Generated code:\n%s", genCode)
}

// TestImportAlias_SamePathDifferentAliases tests that when the same import path
// is used with different aliases across services, they should be merged to use
// a single consistent alias (preferring the longer/more descriptive one).
// Example:
//
//	service_c.go: import userentity "pkga"
//	service_d.go: import pkgamodel "pkga"
//
// Expected: Should merge to one alias (e.g., "pkgamodel" or "userentity")
func TestImportAlias_SamePathDifferentAliases(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package A
	pkgaDir := filepath.Join(tmpDir, "pkga")
	if err := os.MkdirAll(pkgaDir, 0755); err != nil {
		t.Fatalf("Failed to create pkga dir: %v", err)
	}

	pkgaCode := `package pkga

type User struct {
	ID   string
	Name string
}
`
	if err := os.WriteFile(filepath.Join(pkgaDir, "models.go"), []byte(pkgaCode), 0644); err != nil {
		t.Fatalf("Failed to create pkga models: %v", err)
	}

	// Create service C using pkga with alias "userentity"
	serviceCCode := `package main

import (
	userentity "myapp/pkga"
)

// @RouterService name="service-c", prefix="/api/c"
type ServiceC struct {}

// @Route "GET /entity"
func (s *ServiceC) GetEntity() (*userentity.User, error) {
	return nil, nil
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "service_c.go"), []byte(serviceCCode), 0644); err != nil {
		t.Fatalf("Failed to create service_c.go: %v", err)
	}

	// Create service D using pkga with alias "pkgamodel" (same path, different alias)
	serviceDCode := `package main

import (
	pkgamodel "myapp/pkga"
)

// @RouterService name="service-d", prefix="/api/d"
type ServiceD struct {}

// @Route "GET /data"
func (s *ServiceD) GetData() (*pkgamodel.User, error) {
	return nil, nil
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "service_d.go"), []byte(serviceDCode), 0644); err != nil {
		t.Fatalf("Failed to create service_d.go: %v", err)
	}

	// Create go.mod
	goModContent := `module myapp

go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Parse annotations for both services
	annotationsC, err := annotation.ParseFileAnnotations(filepath.Join(tmpDir, "service_c.go"))
	if err != nil {
		t.Fatalf("Failed to parse service_c annotations: %v", err)
	}

	annotationsD, err := annotation.ParseFileAnnotations(filepath.Join(tmpDir, "service_d.go"))
	if err != nil {
		t.Fatalf("Failed to parse service_d annotations: %v", err)
	}

	// Create context
	ctx := &annotation.RouterServiceContext{
		FolderPath: tmpDir,
		UpdatedFiles: []*annotation.FileToProcess{
			{
				Filename:    "service_c.go",
				FullPath:    filepath.Join(tmpDir, "service_c.go"),
				Annotations: annotationsC,
			},
			{
				Filename:    "service_d.go",
				FullPath:    filepath.Join(tmpDir, "service_d.go"),
				Annotations: annotationsD,
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
	genFilePath := filepath.Join(tmpDir, internal.GeneratedFileName)
	genContent, err := os.ReadFile(genFilePath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	genCode := string(genContent)

	// Verify that pkga is imported only once
	pkgaImportCount := strings.Count(genCode, `"myapp/pkga"`)

	if pkgaImportCount != 1 {
		t.Errorf("Expected pkga to be imported exactly once, got %d times", pkgaImportCount)
	}

	// Check which alias was chosen (should prefer longer one)
	hasUserEntity := strings.Contains(genCode, "userentity \"myapp/pkga\"")
	hasPkgaModel := strings.Contains(genCode, "pkgamodel \"myapp/pkga\"")

	if !hasUserEntity && !hasPkgaModel {
		t.Error("Expected either 'userentity' or 'pkgamodel' alias for pkga")
	}

	// Should have exactly one alias
	if hasUserEntity && hasPkgaModel {
		t.Error("Expected only one alias for pkga, found both 'userentity' and 'pkgamodel'")
	}

	// Log which alias was chosen
	if hasUserEntity {
		t.Log("Merged to alias: 'userentity'")
	} else if hasPkgaModel {
		t.Log("Merged to alias: 'pkgamodel'")
	}

	t.Logf("Generated code:\n%s", genCode)
}
