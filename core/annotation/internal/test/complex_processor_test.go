package annotation_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/primadi/lokstra/core/annotation"
)

// TestProcessComplexAnnotations_NoDuplicateFolders tests that duplicate folders
// are only processed once when multiple overlapping paths are provided
func TestProcessComplexAnnotations_NoDuplicateFolders(t *testing.T) {
	// Create temporary test directory structure
	tmpDir := t.TempDir()

	// Create nested folder structure with .go files
	testStructure := map[string]string{
		"module1/service.go":           "package module1\n\n// @Handler name=\"service1\"\ntype Service1 struct {}",
		"module1/submodule/handler.go": "package submodule\n\n// @Handler name=\"service2\"\ntype Service2 struct {}",
		"module2/api.go":               "package module2\n\n// @Handler name=\"service3\"\ntype Service3 struct {}",
	}

	for filePath, content := range testStructure {
		fullPath := filepath.Join(tmpDir, filePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	// Track which folders were processed
	processedFolders := make(map[string]int)
	var mu sync.Mutex

	// Define overlapping paths that should result in same folders being scanned
	scanPaths := []string{
		tmpDir,                           // Root directory
		filepath.Join(tmpDir, "module1"), // Subdirectory (overlaps with root)
		filepath.Join(tmpDir, "module1/submodule"), // Nested subdirectory (overlaps with both)
		tmpDir, // Duplicate root (should be deduplicated)
	}

	// Process with tracking
	_, err := annotation.ProcessComplexAnnotations(scanPaths, 1,
		func(ctx *annotation.RouterServiceContext) error {
			mu.Lock()
			processedFolders[ctx.FolderPath]++
			mu.Unlock()
			return nil
		})

	if err != nil {
		t.Fatalf("ProcessComplexAnnotations failed: %v", err)
	}

	// Verify each folder was processed exactly once
	for folder, count := range processedFolders {
		if count != 1 {
			t.Errorf("Folder %s was processed %d times, expected 1", folder, count)
		}
	}

	// Verify we processed exactly 3 unique folders
	expectedFolderCount := 3 // module1, module1/submodule, module2
	if len(processedFolders) != expectedFolderCount {
		t.Errorf("Expected %d unique folders to be processed, got %d", expectedFolderCount, len(processedFolders))
		t.Logf("Processed folders: %v", processedFolders)
	}
}

// TestProcessComplexAnnotations_EmptyPaths tests handling of empty paths
func TestProcessComplexAnnotations_EmptyPaths(t *testing.T) {
	processedCount := 0

	// Process with empty and nil paths
	_, err := annotation.ProcessComplexAnnotations([]string{"", "", ""}, 1,
		func(ctx *annotation.RouterServiceContext) error {
			processedCount++
			return nil
		})

	if err != nil {
		t.Fatalf("ProcessComplexAnnotations failed with empty paths: %v", err)
	}

	if processedCount != 0 {
		t.Errorf("Expected 0 folders to be processed with empty paths, got %d", processedCount)
	}
}

// TestProcessComplexAnnotations_DifferentPathFormats tests that different path
// representations of the same directory are deduplicated
func TestProcessComplexAnnotations_DifferentPathFormats(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package test\n\n// @Handler name=\"test\"\ntype Test struct {}"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	processedFolders := make(map[string]int)
	var mu sync.Mutex

	// Different representations of the same path
	scanPaths := []string{
		tmpDir,
		tmpDir + string(filepath.Separator), // With trailing separator
		filepath.Clean(tmpDir),              // Cleaned path
	}

	_, err := annotation.ProcessComplexAnnotations(scanPaths, 1,
		func(ctx *annotation.RouterServiceContext) error {
			mu.Lock()
			processedFolders[ctx.FolderPath]++
			mu.Unlock()
			return nil
		})

	if err != nil {
		t.Fatalf("ProcessComplexAnnotations failed: %v", err)
	}

	// Should process exactly once despite multiple path formats
	totalProcessed := 0
	for _, count := range processedFolders {
		totalProcessed += count
	}

	if totalProcessed != 1 {
		t.Errorf("Expected folder to be processed exactly 1 time, got %d times", totalProcessed)
		t.Logf("Processed folders: %v", processedFolders)
	}
}

// TestProcessComplexAnnotations_NoGoFiles tests that folders without .go files are skipped
func TestProcessComplexAnnotations_NoGoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create folder structure without .go files
	if err := os.MkdirAll(filepath.Join(tmpDir, "emptyfolder"), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("No Go files"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	processedCount := 0

	_, err := annotation.ProcessComplexAnnotations([]string{tmpDir}, 1,
		func(ctx *annotation.RouterServiceContext) error {
			processedCount++
			return nil
		})

	if err != nil {
		t.Fatalf("ProcessComplexAnnotations failed: %v", err)
	}

	if processedCount != 0 {
		t.Errorf("Expected 0 folders to be processed (no .go files), got %d", processedCount)
	}
}

// TestProcessComplexAnnotations_ParallelProcessing tests that parallel processing
// still deduplicates correctly
func TestProcessComplexAnnotations_ParallelProcessing(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple folders with .go files
	for i := 1; i <= 10; i++ {
		folderPath := filepath.Join(tmpDir, fmt.Sprintf("module%d", i))
		if err := os.MkdirAll(folderPath, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		testFile := filepath.Join(folderPath, "service.go")
		content := fmt.Sprintf("package module%d\n\n// @Handler name=\"service%d\"\ntype Service%d struct {}", i, i, i)
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	processedFolders := make(map[string]int)
	var mu sync.Mutex

	// Create overlapping paths
	scanPaths := []string{
		tmpDir,                           // All modules
		tmpDir,                           // Duplicate
		filepath.Join(tmpDir, "module1"), // Specific module (overlaps)
		filepath.Join(tmpDir, "module5"), // Another specific module (overlaps)
	}

	// Use multiple workers
	_, err := annotation.ProcessComplexAnnotations(scanPaths, 4,
		func(ctx *annotation.RouterServiceContext) error {
			mu.Lock()
			processedFolders[ctx.FolderPath]++
			mu.Unlock()
			return nil
		})

	if err != nil {
		t.Fatalf("ProcessComplexAnnotations failed: %v", err)
	}

	// Verify no duplicates even with parallel processing
	for folder, count := range processedFolders {
		if count != 1 {
			t.Errorf("Folder %s was processed %d times with parallel processing, expected 1", folder, count)
		}
	}

	// Should process exactly 10 folders
	if len(processedFolders) != 10 {
		t.Errorf("Expected 10 unique folders to be processed, got %d", len(processedFolders))
	}
}

// TestScenario1_OverlappingPaths tests Bootstrap(".", "./modules")
// where ./modules is inside ".", should not duplicate
func TestScenario1_OverlappingPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create structure: root with modules subfolder
	testStructure := map[string]string{
		"main.go":                "package main\n\n// @Handler name=\"main\"\ntype MainService struct {}",
		"modules/user/user.go":   "package user\n\n// @Handler name=\"user\"\ntype UserService struct {}",
		"modules/order/order.go": "package order\n\n// @Handler name=\"order\"\ntype OrderService struct {}",
	}

	for filePath, content := range testStructure {
		fullPath := filepath.Join(tmpDir, filePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	processedFolders := make(map[string]int)
	var mu sync.Mutex

	// Scenario 1: Bootstrap(".", "./modules")
	// Root path "." includes everything, "./modules" overlaps
	scanPaths := []string{
		tmpDir,                           // Equivalent to "."
		filepath.Join(tmpDir, "modules"), // Equivalent to "./modules"
	}

	_, err := annotation.ProcessComplexAnnotations(scanPaths, 1,
		func(ctx *annotation.RouterServiceContext) error {
			mu.Lock()
			processedFolders[ctx.FolderPath]++
			mu.Unlock()
			return nil
		})

	if err != nil {
		t.Fatalf("ProcessComplexAnnotations failed: %v", err)
	}

	// Verify each folder processed exactly once
	for folder, count := range processedFolders {
		if count != 1 {
			t.Errorf("Scenario 1: Folder %s was processed %d times, expected 1", folder, count)
		}
	}

	// Should process 3 unique folders: root, modules/user, modules/order
	expectedCount := 3
	if len(processedFolders) != expectedCount {
		t.Errorf("Scenario 1: Expected %d unique folders, got %d", expectedCount, len(processedFolders))
		t.Logf("Processed folders: %v", processedFolders)
	}
}

// TestScenario2_SamePathDifferentForms tests Bootstrap("./modules", "modules")
// where both paths normalize to the same directory
func TestScenario2_SamePathDifferentForms(t *testing.T) {
	tmpDir := t.TempDir()

	// Create modules folder
	modulesPath := filepath.Join(tmpDir, "modules")
	if err := os.MkdirAll(modulesPath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	testFile := filepath.Join(modulesPath, "service.go")
	if err := os.WriteFile(testFile, []byte("package modules\n\n// @Handler name=\"modules\"\ntype Service struct {}"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	processedFolders := make(map[string]int)
	var mu sync.Mutex

	// Change to tmpDir to simulate relative paths from project root
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	// Scenario 2: Bootstrap("./modules", "modules")
	// Both should normalize to the same path
	scanPaths := []string{
		"./modules",                      // Relative with ./
		"modules",                        // Relative without ./
		filepath.Join(tmpDir, "modules"), // Absolute path
	}

	_, err := annotation.ProcessComplexAnnotations(scanPaths, 1,
		func(ctx *annotation.RouterServiceContext) error {
			mu.Lock()
			processedFolders[ctx.FolderPath]++
			mu.Unlock()
			return nil
		})

	if err != nil {
		t.Fatalf("ProcessComplexAnnotations failed: %v", err)
	}

	// Should have exactly 1 unique folder despite 3 different path representations
	totalProcessed := 0
	for folder, count := range processedFolders {
		totalProcessed += count
		if count != 1 {
			t.Errorf("Scenario 2: Folder %s was processed %d times, expected 1", folder, count)
		}
	}

	if totalProcessed != 1 {
		t.Errorf("Scenario 2: Expected 1 total folder processing, got %d", totalProcessed)
		t.Logf("Processed folders: %v", processedFolders)
	}
}

// TestScenario3_MultipleNestedPaths tests Bootstrap(".", "./modules/user", "./modules/order")
// All folders scanned, but each unique folder processed only once
func TestScenario3_MultipleNestedPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested structure
	testStructure := map[string]string{
		"main.go":                     "package main\n\n// @Handler name=\"main\"\ntype MainService struct {}",
		"modules/user/service.go":     "package user\n\n// @Handler name=\"user\"\ntype UserService struct {}",
		"modules/user/repository.go":  "package user\n\n// No RouterService here",
		"modules/order/service.go":    "package order\n\n// @Handler name=\"order\"\ntype OrderService struct {}",
		"modules/order/repository.go": "package order\n\n// No RouterService here",
		"modules/payment/service.go":  "package payment\n\n// @Handler name=\"payment\"\ntype PaymentService struct {}",
	}

	for filePath, content := range testStructure {
		fullPath := filepath.Join(tmpDir, filePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	processedFolders := make(map[string]int)
	var mu sync.Mutex

	// Scenario 3: Bootstrap(".", "./modules/user", "./modules/order")
	// Root "." scans everything, but specific modules are also listed
	scanPaths := []string{
		tmpDir,                                   // Equivalent to "." (scans all)
		filepath.Join(tmpDir, "modules", "user"), // Specific module (overlaps with ".")
		filepath.Join(tmpDir, "modules", "order"), // Specific module (overlaps with ".")
	}

	_, err := annotation.ProcessComplexAnnotations(scanPaths, 1,
		func(ctx *annotation.RouterServiceContext) error {
			mu.Lock()
			processedFolders[ctx.FolderPath]++
			mu.Unlock()
			return nil
		})

	if err != nil {
		t.Fatalf("ProcessComplexAnnotations failed: %v", err)
	}

	// Verify each folder processed exactly once
	for folder, count := range processedFolders {
		if count != 1 {
			t.Errorf("Scenario 3: Folder %s was processed %d times, expected 1", folder, count)
		}
	}

	// Should process 4 unique folders: root, modules/user, modules/order, modules/payment
	// Note: modules/payment is also scanned because root "." includes it
	expectedCount := 4
	if len(processedFolders) != expectedCount {
		t.Errorf("Scenario 3: Expected %d unique folders, got %d", expectedCount, len(processedFolders))
		t.Logf("Processed folders: %v", processedFolders)
	}
}

// TestFileContainsRouterService tests the quick check function for @Handler annotation
func TestFileContainsRouterService(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name: "standard format",
			content: `package app

// @Handler name="user-service"
type UserService struct {}
`,
			expected: true,
		},
		{
			name: "with_spaces_after_//",
			content: `package app

//   @Handler name="user-service"
type UserService struct {}
`,
			expected: false, // Changed: Multiple spaces (>1) are treated as indented (code example)
		},
		{
			name: "no space after //",
			content: `package app

//@Handler name="user-service"
type UserService struct {}
`,
			expected: true,
		},
		{
			name: "in string (should not match - no comment)",
			content: `package app

const x = "@Handler name=\"test\""
`,
			expected: false,
		},
		{
			name: "in descriptive comment (should NOT match - not at start)",
			content: `package app

// This is about @Handler annotation
type UserService struct {}
`,
			expected: false, // Should NOT match - @Handler is not at the start of comment
		},
		{
			name: "block comment (should not match)",
			content: `package app

/* @Handler name="test" */
type UserService struct {}
`,
			expected: false,
		},
		{
			name: "no annotation",
			content: `package app

type UserService struct {}
`,
			expected: false,
		},
		{
			name: "different annotation only",
			content: `package app

// @Route "GET /users"
func GetUsers() {}
`,
			expected: false,
		},
		{
			name: "tab before comment",
			content: `package app

	// @Handler name="test"
type Service struct {}
`,
			expected: true,
		},
		{
			name: "multiple annotations including RouterService",
			content: `package app

// @Route "GET /test"
// @Handler name="test"
type Service struct {}
`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.go")

			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			// Test (access internal function via annotation package)
			result, err := annotation.TestFileContainsRouterService(tmpFile)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("annotation.FileContainsRouterService() = %v, want %v", result, tt.expected)
				t.Logf("File content:\n%s", tt.content)
			}
		})
	}
}

// TestFileContainsRouterService_Error tests error handling
func TestFileContainsRouterService_Error(t *testing.T) {
	// Test with non-existent file
	_, err := annotation.TestFileContainsRouterService("/nonexistent/file.go")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}
