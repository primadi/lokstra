package annotation

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/primadi/lokstra/core/annotation/internal"
)

// TestErrorCleanup_InvalidAnnotation tests cleanup when annotation parsing fails
func TestErrorCleanup_InvalidAnnotation(t *testing.T) {
	tmpDir := t.TempDir()

	// Step 1: Create valid service with cache and generated file
	validContent := `package application

// @RouterService name="user-service", prefix="/api/users"
type UserService struct {
	ID string
}
`
	filePath := filepath.Join(tmpDir, "user_service.go")
	if err := os.WriteFile(filePath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Process to create cache and generated file
	_, err := ProcessPerFolder(tmpDir, GenerateCodeForFolder)
	if err != nil {
		t.Fatalf("Initial ProcessPerFolder() error = %v", err)
	}

	cachePath := filepath.Join(tmpDir, internal.CacheFileName)
	genPath := filepath.Join(tmpDir, internal.GeneratedFileName)

	// Verify files created
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatalf("Expected cache file to exist: %s", cachePath)
	}
	if _, err := os.Stat(genPath); os.IsNotExist(err) {
		t.Fatalf("Expected generated file to exist: %s", genPath)
	}

	t.Logf("✓ Step 1: Cache and generated files created")

	// Step 2: Create invalid annotation that will cause parsing error
	// Invalid: @RouterService on function instead of struct
	invalidContent := `package application

// @RouterService name="invalid-service", prefix="/api/invalid"
func InvalidService() {
	// This should cause validation error
}
`
	if err := os.WriteFile(filePath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	// Process again - should fail and cleanup
	_, err = ProcessPerFolder(tmpDir, GenerateCodeForFolder)
	if err == nil {
		t.Fatalf("Expected error for invalid annotation, got nil")
	}

	t.Logf("✓ Step 2: Got expected error: %v", err)

	// Step 3: Verify cleanup - both cache and generated files should be removed
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Errorf("Expected cache file to be cleaned up after error, but it still exists: %s", cachePath)
	} else {
		t.Logf("✓ Step 3a: Cache file cleaned up: %s", cachePath)
	}

	if _, err := os.Stat(genPath); !os.IsNotExist(err) {
		t.Errorf("Expected generated file to be cleaned up after error, but it still exists: %s", genPath)
	} else {
		t.Logf("✓ Step 3b: Generated file cleaned up: %s", genPath)
	}
}

// TestErrorCleanup_FileReadError tests cleanup when file reading fails
func TestErrorCleanup_FileReadError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with valid annotation first
	validContent := `package application

// @RouterService name="user-service", prefix="/api/users"
type UserService struct {}
`
	filePath := filepath.Join(tmpDir, "user_service.go")
	if err := os.WriteFile(filePath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Process to create cache and generated file
	_, err := ProcessPerFolder(tmpDir, GenerateCodeForFolder)
	if err != nil {
		t.Fatalf("Initial ProcessPerFolder() error = %v", err)
	}

	cachePath := filepath.Join(tmpDir, internal.CacheFileName)
	genPath := filepath.Join(tmpDir, internal.GeneratedFileName)

	// Verify files created
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatalf("Expected cache file to exist")
	}
	if _, err := os.Stat(genPath); os.IsNotExist(err) {
		t.Fatalf("Expected generated file to exist")
	}

	t.Logf("✓ Step 1: Cache and generated files created")

	// Create another file with annotation that will fail processing
	invalidGoFile := filepath.Join(tmpDir, "invalid.go")
	invalidContent := `package application

// @RouterService name="bad-service"
// This is not valid Go syntax
type BadService struct {
	unclosed string "json:"name
}
`
	if err := os.WriteFile(invalidGoFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	// Process again - should fail and cleanup
	_, err = ProcessPerFolder(tmpDir, GenerateCodeForFolder)
	if err == nil {
		t.Fatalf("Expected error for invalid Go syntax, got nil")
	}

	t.Logf("✓ Step 2: Got expected error: %v", err)

	// Verify cleanup
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Errorf("Expected cache file to be cleaned up after error")
	} else {
		t.Logf("✓ Step 3a: Cache file cleaned up")
	}

	if _, err := os.Stat(genPath); !os.IsNotExist(err) {
		t.Errorf("Expected generated file to be cleaned up after error")
	} else {
		t.Logf("✓ Step 3b: Generated file cleaned up")
	}
}

// TestErrorCleanup_ProcessingError tests cleanup when processing callback fails
func TestErrorCleanup_ProcessingError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid file
	validContent := `package application

// @RouterService name="user-service", prefix="/api/users"
type UserService struct {}
`
	filePath := filepath.Join(tmpDir, "user_service.go")
	if err := os.WriteFile(filePath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// First process with normal callback
	_, err := ProcessPerFolder(tmpDir, GenerateCodeForFolder)
	if err != nil {
		t.Fatalf("Initial ProcessPerFolder() error = %v", err)
	}

	cachePath := filepath.Join(tmpDir, internal.CacheFileName)
	genPath := filepath.Join(tmpDir, internal.GeneratedFileName)

	// Verify files created
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatalf("Expected cache file to exist")
	}
	if _, err := os.Stat(genPath); os.IsNotExist(err) {
		t.Fatalf("Expected generated file to exist")
	}

	t.Logf("✓ Step 1: Cache and generated files created")

	// Update file to trigger reprocessing
	if err := os.WriteFile(filePath, []byte(validContent+"// updated\n"), 0644); err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// Process with callback that returns error
	errorCallback := func(ctx *RouterServiceContext) error {
		return fmt.Errorf("simulated processing error")
	}

	_, err = ProcessPerFolder(tmpDir, errorCallback)
	if err == nil {
		t.Fatalf("Expected processing error, got nil")
	}

	t.Logf("✓ Step 2: Got expected error: %v", err)

	// Verify cleanup
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Errorf("Expected cache file to be cleaned up after processing error")
	} else {
		t.Logf("✓ Step 3a: Cache file cleaned up")
	}

	if _, err := os.Stat(genPath); !os.IsNotExist(err) {
		t.Errorf("Expected generated file to be cleaned up after processing error")
	} else {
		t.Logf("✓ Step 3b: Generated file cleaned up")
	}
}
