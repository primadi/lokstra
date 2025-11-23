package annotation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/annotation"
	"github.com/primadi/lokstra/core/annotation/internal"
)

// Helper function to run codegen on a folder
func runCodegen(folderPath string) error {
	_, err := annotation.ProcessPerFolder(folderPath, annotation.GenerateCodeForFolder)
	return err
}

// TestMultiFileCodegen tests multi-file code generation in testdata/multifile_test
// You can inspect the generated files in that folder
func TestMultiFileCodegen(t *testing.T) {
	testDir := filepath.Join("testdata", "multifile_test")

	// Clean up generated files from previous runs
	genPath := filepath.Join(testDir, internal.GeneratedFileName)
	cachePath := filepath.Join(testDir, internal.CacheFileName)
	os.Remove(genPath)
	os.Remove(cachePath)

	// First generation - should generate all 3 services
	t.Run("Initial generation", func(t *testing.T) {
		err := runCodegen(testDir)
		if err != nil {
			t.Fatalf("Initial generation failed: %v", err)
		}

		// Verify generated file exists
		if _, err := os.Stat(genPath); os.IsNotExist(err) {
			t.Fatal("Generated file was not created")
		}

		// Read and verify content
		content, err := os.ReadFile(genPath)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		genCode := string(content)

		// Verify all 3 services are registered in init()
		if !strings.Contains(genCode, "RegisterUserService()") {
			t.Error("RegisterUserService() not found in init()")
		}
		if !strings.Contains(genCode, "RegisterProfileService()") {
			t.Error("RegisterProfileService() not found in init()")
		}
		if !strings.Contains(genCode, "RegisterAuthService()") {
			t.Error("RegisterAuthService() not found in init()")
		}

		t.Logf("✓ Generated file contains all 3 Register calls in init()")
		t.Logf("✓ Check file: %s", genPath)
	})

	// Modify one file and regenerate
	t.Run("Modify one file", func(t *testing.T) {
		// Read current user_service.go
		userServicePath := filepath.Join(testDir, "user_service.go")
		originalContent, err := os.ReadFile(userServicePath)
		if err != nil {
			t.Fatalf("Failed to read user_service.go: %v", err)
		}

		// Modify it - add a new method
		modifiedContent := string(originalContent) + `
// @Route "DELETE /users/{id}"
func (s *UserService) Delete(id string) error {
	return nil
}
`
		if err := os.WriteFile(userServicePath, []byte(modifiedContent), 0644); err != nil {
			t.Fatalf("Failed to modify user_service.go: %v", err)
		}

		// Restore original content after test
		defer os.WriteFile(userServicePath, originalContent, 0644)

		// Regenerate
		err = runCodegen(testDir)
		if err != nil {
			t.Fatalf("Regeneration failed: %v", err)
		}

		// Read generated file
		content, err := os.ReadFile(genPath)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		genCode := string(content)

		// CRITICAL: All 3 services should STILL be in init()
		if !strings.Contains(genCode, "RegisterUserService()") {
			t.Error("After modifying user_service.go: RegisterUserService() missing!")
		}
		if !strings.Contains(genCode, "RegisterProfileService()") {
			t.Error("After modifying user_service.go: RegisterProfileService() missing - BUG!")
		}
		if !strings.Contains(genCode, "RegisterAuthService()") {
			t.Error("After modifying user_service.go: RegisterAuthService() missing - BUG!")
		}

		// Verify new Delete method is in generated code
		if !strings.Contains(genCode, "Delete") {
			t.Error("New Delete method not found in generated code")
		}

		t.Logf("✓ After modifying one file, all 3 services still registered")
		t.Logf("✓ New Delete method added to generated code")
	})

	// Test manual edit detection
	t.Run("Manual edit detection", func(t *testing.T) {
		// Manually edit generated file
		content, err := os.ReadFile(genPath)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		manualEdit := "// MANUAL EDIT - SHOULD BE REMOVED\n" + string(content)
		if err := os.WriteFile(genPath, []byte(manualEdit), 0644); err != nil {
			t.Fatalf("Failed to manually edit file: %v", err)
		}

		// Regenerate - should detect timestamp mismatch
		err = runCodegen(testDir)
		if err != nil {
			t.Fatalf("Regeneration after manual edit failed: %v", err)
		}

		// Read regenerated content
		regenContent, err := os.ReadFile(genPath)
		if err != nil {
			t.Fatalf("Failed to read regenerated file: %v", err)
		}

		genCode := string(regenContent)

		// Manual comment should be REMOVED
		if strings.Contains(genCode, "MANUAL EDIT") {
			t.Error("Manual edit was NOT removed - timestamp validation failed!")
		}

		// All services should still be registered
		if !strings.Contains(genCode, "RegisterUserService()") ||
			!strings.Contains(genCode, "RegisterProfileService()") ||
			!strings.Contains(genCode, "RegisterAuthService()") {
			t.Error("Services missing after manual edit regeneration")
		}

		t.Logf("✓ Manual edits correctly removed on regeneration")
		t.Logf("✓ All services still registered correctly")
	})

	t.Logf("\n=== Generated files location ===")
	t.Logf("Generated: %s", genPath)
	t.Logf("Cache:     %s", cachePath)
	t.Logf("\nYou can inspect these files to verify the fix works correctly!")
}
