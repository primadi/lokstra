package annotation

import (
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
)

// ProcessComplexAnnotations processes annotations with parallel folder processing
func ProcessComplexAnnotations(rootPath string, maxWorkers int,
	onProcessRouterService func(*RouterServiceContext) error) (bool, error) {
	// Find all folders containing .go files
	normPath := utils.NormalizeWithBasePath(rootPath)
	folders, err := findGoFolders(normPath)
	if err != nil {
		return false, fmt.Errorf("failed to find folders: %w", err)
	}
	if maxWorkers == 0 {
		maxWorkers = runtime.NumCPU() * 2
	}

	// Create worker pool
	folderChan := make(chan string, len(folders))
	errChan := make(chan error, len(folders))
	changedChan := make(chan bool, len(folders))
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
	for _, folder := range folders {
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
	cachePath := filepath.Join(folderPath, cacheFileName)

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

	// Step 2: Scan .go files containing @RouterService
	skipped, updated, deleted, err := scanFolderFiles(folderPath, cache)
	if err != nil {
		return false, fmt.Errorf("failed to scan files: %w", err)
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
		return false, fmt.Errorf("failed to process router service: %w", err)
	}

	// Step 4: Update cache in memory (if no errors)
	for _, file := range updated {
		cache.Files[file.Filename] = &FileCacheEntry{
			Filename:    file.Filename,
			Checksum:    file.Checksum,
			Annotations: file.AnnotationCount,
			LastScan:    time.Now(),
			Generated:   []string{generatedFileName},
		}
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
	Version   int                        `json:"version"`
	Files     map[string]*FileCacheEntry `json:"files"`
	UpdatedAt time.Time                  `json:"updated_at"`
}

// FileCacheEntry represents a single file in cache
type FileCacheEntry struct {
	Filename    string    `json:"filename"`
	Checksum    string    `json:"checksum"`
	Annotations int       `json:"annotations"`
	LastScan    time.Time `json:"last_scan"`
	Generated   []string  `json:"generated"`
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
	FieldType   string // e.g., "*service.Cached[domain.UserRepository]"
	InnerType   string // e.g., "domain.UserRepository" (extracted from generic)
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
	Args           map[string]interface{}
	PositionalArgs []interface{}
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
		if file.Name() == generatedFileName || strings.HasSuffix(file.Name(), "_test.go") {
			continue
		}

		fullPath := filepath.Join(folderPath, file.Name())

		// Quick check: does file contain @RouterService?
		hasRouterService, err := fileContainsRouterService(fullPath)
		if err != nil {
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
			// File unchanged, skip
			skipped = append(skipped, &FileToProcess{
				Filename: file.Name(),
				FullPath: fullPath,
				Checksum: checksum,
			})
			continue
		}

		// File changed or new, parse annotations
		annotations, err := parseFileAnnotations(fullPath)
		if err != nil {
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

// fileContainsRouterService quickly checks if file contains @RouterService
func fileContainsRouterService(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(content), "@RouterService"), nil
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
