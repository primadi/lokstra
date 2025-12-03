package annotation

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/annotation/internal"
)

// ProcessComplexAnnotations processes annotations with parallel folder processing.
// rootPath is a slice of directories to scan. Each directory and its subdirectories
// will be scanned for .go files containing @RouterService annotations.
func ProcessComplexAnnotations(rootPath []string, maxWorkers int,
	onProcessRouterService func(*RouterServiceContext) error) (bool, error) {
	// Find all folders containing .go files from all root paths
	// Use map as set to avoid duplicate folders
	folderSet := make(map[string]bool)

	for _, path := range rootPath {
		if path == "" {
			continue
		}

		// Normalize the path
		normPath := utils.NormalizeWithBasePath(path)

		// Find folders in this path
		folders, err := findGoFolders(normPath)
		if err != nil {
			return false, fmt.Errorf("failed to find folders in %s: %w", path, err)
		}

		// Add to set (automatically deduplicates)
		for _, folder := range folders {
			folderSet[folder] = true
		}
	}

	if len(folderSet) == 0 {
		return false, nil
	}

	// Convert set to slice
	allFolders := make([]string, 0, len(folderSet))
	for folder := range folderSet {
		allFolders = append(allFolders, folder)
	}

	if maxWorkers == 0 {
		maxWorkers = runtime.NumCPU() * 2
	}

	// Create worker pool
	folderChan := make(chan string, len(allFolders))
	errChan := make(chan error, len(allFolders))
	changedChan := make(chan bool, len(allFolders))
	var wg sync.WaitGroup

	// Spawn workers
	for range maxWorkers {
		wg.Go(func() {
			for folder := range folderChan {
				codeChanged, err := ProcessPerFolder(folder, onProcessRouterService)
				if err != nil {
					errChan <- fmt.Errorf("folder %s: %w", folder, err)
				} else if codeChanged {
					changedChan <- true
				}
			}
		})
	}

	// Send folders to workers
	for _, folder := range allFolders {
		folderChan <- folder
	}
	close(folderChan)

	// Wait for all workers
	wg.Wait()
	close(errChan)
	close(changedChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// Check if any code changed
	anyCodeChanged := len(changedChan) > 0

	if len(errors) > 0 {
		return anyCodeChanged, fmt.Errorf("processing failed with %d errors: %v", len(errors), errors[0])
	}

	return anyCodeChanged, nil
}

// ProcessPerFolder processes a single folder
func ProcessPerFolder(folderPath string, onProcessRouterService func(*RouterServiceContext) error) (bool, error) {
	cachePath := filepath.Join(folderPath, internal.CacheFileName)
	genPath := filepath.Join(folderPath, internal.GeneratedFileName)

	// Step 1: Load cache if exists
	cache, err := loadCache(cachePath)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to load cache: %w", err)
	}
	if cache == nil {
		cache = &FolderCache{
			Version: 1,
			Files:   make(map[string]*FileCacheEntry),
		}
	}

	// Step 1.5: Check if generated file was manually altered or deleted (checksum mismatch)
	forceRegenerate := false
	if data, err := os.ReadFile(genPath); err == nil {
		// Generated file exists, check if checksum matches cache
		genChecksum := calculateChecksumFromBytes(data)
		if cache.GeneratedChecksum != "" && cache.GeneratedChecksum != genChecksum {
			// Checksum mismatch - file was manually altered
			forceRegenerate = true
		}
	} else if os.IsNotExist(err) && len(cache.Files) > 0 {
		// Generated file deleted but cache exists - force regenerate
		forceRegenerate = true
	}

	// Step 2: Scan .go files containing @RouterService
	skipped, updated, deleted, err := scanFolderFiles(folderPath, cache)
	if err != nil {
		return false, fmt.Errorf("failed to scan files: %w", err)
	}
	// If generated file was altered, force regenerate all files
	if forceRegenerate {
		// Move all skipped files to updated, but need to parse their annotations first
		for _, file := range skipped {
			// Parse annotations for skipped files
			annotations, err := ParseFileAnnotations(file.FullPath)
			if err != nil {
				// Cleanup before returning error
				cleanupFolder(folderPath)
				return false, fmt.Errorf("failed to parse annotations from %s: %w", file.Filename, err)
			}
			file.Annotations = annotations
			file.AnnotationCount = len(annotations)
			updated = append(updated, file)
		}
		skipped = nil
		// Clear cache to force full regeneration
		cache.Files = make(map[string]*FileCacheEntry)
	}

	// If no RouterService annotations found and no cache, skip
	if len(updated) == 0 && len(deleted) == 0 && len(skipped) == 0 {
		return false, nil
	}

	// Track if code changed (has updates or deletes)
	codeChanged := len(updated) > 0 || len(deleted) > 0

	// Step 3: Process RouterService annotations
	ctx := &RouterServiceContext{
		FolderPath:   folderPath,
		SkippedFiles: skipped,
		UpdatedFiles: updated,
		DeletedFiles: deleted,
		Cache:        cache,
		GeneratedCode: &GeneratedCode{
			Services:          make(map[string]*ServiceGeneration),
			PreservedSections: make(map[string]string),
		},
	}

	if err := onProcessRouterService(ctx); err != nil {
		// Cleanup before returning error
		cleanupFolder(folderPath)
		return false, fmt.Errorf("failed to process router service: %w", err)
	}

	// Step 4: Update cache in memory (if no errors)
	// Only update cache if there were actual changes
	if len(updated) > 0 || len(deleted) > 0 {
		// Get generated file checksum
		genChecksum := ""
		if data, err := os.ReadFile(genPath); err == nil {
			genChecksum = calculateChecksumFromBytes(data)
		}
		if genChecksum == "" {
			return false, nil
		}
		cache.GeneratedChecksum = genChecksum

		for _, file := range updated {
			cache.Files[file.Filename] = &FileCacheEntry{
				Filename:         file.Filename,
				Checksum:         file.Checksum,
				Annotations:      file.AnnotationCount,
				LastScan:         time.Now(),
				Generated:        []string{internal.GeneratedFileName},
				GeneratedModTime: time.Now(), // Keep for backward compatibility
			}
		}

		cache.UpdatedAt = time.Now()
	}

	// Remove deleted files from cache
	for _, filename := range deleted {
		delete(cache.Files, filename)
	}

	cache.UpdatedAt = time.Now()

	// Step 5: Write or delete cache file
	if len(cache.Files) > 0 {
		if err := saveCache(cachePath, cache); err != nil {
			return false, fmt.Errorf("failed to save cache: %w", err)
		}
	} else {
		// No files with annotations, remove cache if exists
		os.Remove(cachePath)
	}

	return codeChanged, nil
}

// FolderCache represents the cache json structure
type FolderCache struct {
	Version           int                        `json:"version"`
	Files             map[string]*FileCacheEntry `json:"files"`
	UpdatedAt         time.Time                  `json:"updated_at"`
	GeneratedChecksum string                     `json:"generated_checksum"` // Checksum of zz_generated.lokstra.go
}

// FileCacheEntry represents a single file in cache
type FileCacheEntry struct {
	Filename         string    `json:"filename"`
	Checksum         string    `json:"checksum"`
	Annotations      int       `json:"annotations"`
	LastScan         time.Time `json:"last_scan"`
	Generated        []string  `json:"generated"`
	GeneratedModTime time.Time `json:"generated_mod_time"` // Timestamp of zz_generated.lokstra.go
}

// FileToProcess represents a file that needs processing
type FileToProcess struct {
	Filename        string
	FullPath        string
	Checksum        string
	AnnotationCount int
	Annotations     []*ParsedAnnotation
}

// RouterServiceContext contains context for processing RouterService annotations
type RouterServiceContext struct {
	FolderPath    string
	SkippedFiles  []*FileToProcess
	UpdatedFiles  []*FileToProcess
	DeletedFiles  []string
	Cache         *FolderCache
	GeneratedCode *GeneratedCode
}

// GeneratedCode holds all generated code for a folder
type GeneratedCode struct {
	Services          map[string]*ServiceGeneration
	PreservedSections map[string]string // filename -> code section from existing zz_generated.lokstra.go
}

// ServiceGeneration holds generation data for one service
type ServiceGeneration struct {
	ServiceName      string
	Prefix           string
	Middlewares      []string
	Routes           map[string]string           // methodName -> "METHOD /path"
	RouteMiddlewares map[string][]string         // methodName -> []middleware (per-route middleware)
	Methods          map[string]*MethodSignature // methodName -> signature
	Dependencies     map[string]*DependencyInfo  // serviceName -> field info
	Imports          map[string]string           // alias -> import path (e.g., "domain" -> ".../.../domain")
	StructName       string
	InterfaceName    string
	RemoteTypeName   string
	SourceFile       string
}

// DependencyInfo holds field injection information
type DependencyInfo struct {
	ServiceName string // e.g., "user-repository"
	FieldName   string // e.g., "UserRepo"
	FieldType   string // e.g., "domain.UserRepository" (interface type)
}

// MethodSignature holds method signature information
type MethodSignature struct {
	Name       string
	ParamType  string // e.g., "*domain.GetUserRequest"
	ReturnType string // e.g., "*domain.User" or empty for error-only
	HasData    bool   // true if returns data, false if only error
}

// ParsedAnnotation represents a parsed annotation with arguments
type ParsedAnnotation struct {
	Name           string
	Args           map[string]any
	PositionalArgs []any
	Line           int
	TargetName     string
	TargetType     string
}

// findGoFolders finds all folders containing .go files
func findGoFolders(root string) ([]string, error) {
	folderSet := make(map[string]bool)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			folderSet[filepath.Dir(path)] = true
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	folders := make([]string, 0, len(folderSet))
	for folder := range folderSet {
		folders = append(folders, folder)
	}

	return folders, nil
}

// scanFolderFiles scans files in a folder and categorizes them
func scanFolderFiles(folderPath string, cache *FolderCache) ([]*FileToProcess, []*FileToProcess, []string, error) {
	var skipped, updated []*FileToProcess
	seenFiles := make(map[string]bool)

	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		// Skip generated files
		if file.Name() == internal.GeneratedFileName || strings.HasSuffix(file.Name(), "_test.go") {
			continue
		}

		fullPath := filepath.Join(folderPath, file.Name())

		// Quick check: does file contain @RouterService?
		hasRouterService, err := fileContainsRouterService(fullPath)
		if err != nil {
			// Cleanup before returning error
			cleanupFolder(folderPath)
			return nil, nil, nil, err
		}
		if !hasRouterService {
			continue
		}

		seenFiles[file.Name()] = true

		// Calculate checksum
		checksum, err := calculateChecksum(fullPath)
		if err != nil {
			return nil, nil, nil, err
		}

		// Check cache
		cached, exists := cache.Files[file.Name()]
		if exists && cached.Checksum == checksum {
			// File unchanged, skip but preserve annotation count from cache
			skipped = append(skipped, &FileToProcess{
				Filename:        file.Name(),
				FullPath:        fullPath,
				Checksum:        checksum,
				AnnotationCount: cached.Annotations, // Preserve from cache!
			})
			continue
		}

		// File changed or new, parse annotations
		annotations, err := ParseFileAnnotations(fullPath)
		if err != nil {
			// Cleanup before returning error
			cleanupFolder(folderPath)
			return nil, nil, nil, err
		}

		updated = append(updated, &FileToProcess{
			Filename:        file.Name(),
			FullPath:        fullPath,
			Checksum:        checksum,
			AnnotationCount: len(annotations),
			Annotations:     annotations,
		})
	}

	// Find deleted files (in cache but not in folder)
	var deleted []string
	for filename := range cache.Files {
		if !seenFiles[filename] {
			deleted = append(deleted, filename)
		}
	}

	return skipped, updated, deleted, nil
}

// fileContainsRouterService quickly checks if file contains @RouterService annotation.
// Uses same parsing logic as ParseFileAnnotations for consistency.
// Only matches when @RouterService is at the start of comment content (after // and spaces).
// Ignores TAB-indented annotations (Go code examples in documentation).
func fileContainsRouterService(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		trimmedLine := bytes.TrimSpace(line)

		// Check for // comment
		if after, ok := bytes.CutPrefix(trimmedLine, []byte("//")); ok {
			// CRITICAL: Detect Go code examples (TAB-indented after //)
			// Valid annotation:   // @RouterService
			// Invalid annotation: //	@RouterService (TAB - code example)

			// Find the position of // in original line
			commentPos := bytes.Index(line, []byte("//"))
			if commentPos != -1 && commentPos+2 < len(line) {
				afterComment := line[commentPos+2:]

				// Check for TAB immediately after // (Go code example convention)
				if len(afterComment) > 0 && afterComment[0] == '\t' {
					// This is a code example (TAB-indented) - skip
					continue
				}

				// Check for multiple spaces or single TAB
				trimmedAfter := bytes.TrimLeft(afterComment, " \t")

				if bytes.HasPrefix(trimmedAfter, []byte("@RouterService")) {
					leadingWhitespace := afterComment[:len(afterComment)-len(trimmedAfter)]

					// Allow single space only (normal comment: "// @RouterService")
					// Reject TAB or multiple spaces
					if len(leadingWhitespace) > 1 || (len(leadingWhitespace) == 1 && leadingWhitespace[0] == '\t') {
						// Indented - skip it
						continue
					}

					return true, nil
				}
			}

			// Also check trimmed version for backward compatibility
			after = bytes.TrimSpace(after)
			if bytes.HasPrefix(after, []byte("@RouterService")) {
				// Double-check: make sure it's not TAB-indented
				commentPos := bytes.Index(line, []byte("//"))
				if commentPos != -1 && commentPos+2 < len(line) {
					afterComment := line[commentPos+2:]
					if len(afterComment) > 0 && afterComment[0] == '\t' {
						// TAB-indented - skip
						continue
					}

					trimmedAfter := bytes.TrimLeft(afterComment, " \t")
					leadingWhitespace := afterComment[:len(afterComment)-len(trimmedAfter)]

					if len(leadingWhitespace) > 1 || (len(leadingWhitespace) == 1 && leadingWhitespace[0] == '\t') {
						continue
					}

					return true, nil
				}
			}
		}
	}

	return false, scanner.Err()
} // TestFileContainsRouterService is exported for testing purposes only.
// It wraps the internal fileContainsRouterService function.
func TestFileContainsRouterService(path string) (bool, error) {
	return fileContainsRouterService(path)
}

// calculateChecksum calculates SHA256 checksum of a file
func calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// calculateChecksumFromBytes calculates checksum from byte slice
func calculateChecksumFromBytes(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// loadCache loads cache from JSON file
func loadCache(path string) (*FolderCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache FolderCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

// saveCache saves cache to JSON file
func saveCache(path string, cache *FolderCache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// cleanupFolder removes cache and generated files from a folder
func cleanupFolder(folderPath string) {
	cachePath := filepath.Join(folderPath, internal.CacheFileName)
	genPath := filepath.Join(folderPath, internal.GeneratedFileName)

	// Remove cache file
	if err := os.Remove(cachePath); err == nil {
		fmt.Fprintf(os.Stderr, "[lokstra-annotation] 🗑️  Cleaned up %s in %s\n",
			internal.CacheFileName, folderPath)
	}

	// Remove generated file
	if err := os.Remove(genPath); err == nil {
		fmt.Fprintf(os.Stderr, "[lokstra-annotation] 🗑️  Cleaned up %s in %s\n",
			internal.GeneratedFileName, folderPath)
	}
}
