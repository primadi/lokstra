package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ProjectCreator handles the creation of new projects from templates
type ProjectCreator struct {
	ProjectName  string
	TemplatePath string
	Branch       string
	TargetDir    string
}

// NewProjectCreator creates a new ProjectCreator instance
func NewProjectCreator(projectName, templatePath, branch string) *ProjectCreator {
	return &ProjectCreator{
		ProjectName:  projectName,
		TemplatePath: templatePath,
		Branch:       branch,
		TargetDir:    filepath.Join(".", projectName),
	}
}

// Create executes the project creation process
func (pc *ProjectCreator) Create() error {
	// Step 1: Check if directory already exists
	if err := pc.checkTargetDir(); err != nil {
		return err
	}

	// Step 2: Download template from GitHub
	fmt.Println("📥 Downloading template from GitHub...")
	tempDir, err := pc.downloadTemplate()
	if err != nil {
		return fmt.Errorf("failed to download template: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Step 3: Copy template files to target directory
	fmt.Println("📋 Copying template files...")
	templateSourceDir := filepath.Join(tempDir, "project_templates", pc.TemplatePath)
	if err := pc.copyTemplate(templateSourceDir); err != nil {
		return fmt.Errorf("failed to copy template: %w", err)
	}

	// Step 4: Fix imports in all .go files
	fmt.Println("🔧 Fixing imports...")
	if err := pc.fixImports(); err != nil {
		return fmt.Errorf("failed to fix imports: %w", err)
	}

	// Step 5: Initialize go module
	fmt.Println("📦 Initializing Go module...")
	if err := pc.initGoModule(); err != nil {
		return fmt.Errorf("failed to initialize go module: %w", err)
	}

	// Step 6: Run go mod tidy
	fmt.Println("🧹 Running go mod tidy...")
	if err := pc.runGoModTidy(); err != nil {
		return fmt.Errorf("failed to run go mod tidy: %w", err)
	}

	fmt.Println()
	fmt.Printf("✅ Project created successfully!\n")
	fmt.Println()
	fmt.Printf("Next steps:\n")
	fmt.Printf("  cd %s\n", pc.ProjectName)
	fmt.Printf("  go run .\n")
	fmt.Println()

	return nil
}

// checkTargetDir checks if the target directory already exists
func (pc *ProjectCreator) checkTargetDir() error {
	if _, err := os.Stat(pc.TargetDir); err == nil {
		return fmt.Errorf("directory '%s' already exists", pc.TargetDir)
	}
	return nil
}

// downloadTemplate downloads the template from GitHub
func (pc *ProjectCreator) downloadTemplate() (string, error) {
	// GitHub archive URL format: https://github.com/owner/repo/archive/refs/heads/branch.tar.gz
	url := fmt.Sprintf("https://github.com/primadi/lokstra/archive/refs/heads/%s.tar.gz", pc.Branch)

	// Download the tar.gz file
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "lokstra-template-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Extract tar.gz
	if err := pc.extractTarGz(resp.Body, tempDir); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to extract archive: %w", err)
	}

	// GitHub extracts to a folder named "lokstra-{branch}"
	extractedDir := filepath.Join(tempDir, fmt.Sprintf("lokstra-%s", pc.Branch))

	return extractedDir, nil
}

// extractTarGz extracts a tar.gz archive
func (pc *ProjectCreator) extractTarGz(reader io.Reader, destDir string) error {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create parent directory if needed
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			// Create file
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// Copy contents
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}

// copyTemplate copies template files to target directory
func (pc *ProjectCreator) copyTemplate(sourceDir string) error {
	// Check if source exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("template not found: %s", pc.TemplatePath)
	}

	// Create target directory
	if err := os.MkdirAll(pc.TargetDir, 0755); err != nil {
		return err
	}

	// Walk through source directory and copy files
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from source
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Skip if it's the root directory
		if relPath == "." {
			return nil
		}

		// Target path
		targetPath := filepath.Join(pc.TargetDir, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Copy file
		return pc.copyFile(path, targetPath, info.Mode())
	})
}

// copyFile copies a single file
func (pc *ProjectCreator) copyFile(src, dst string, mode os.FileMode) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// fixImports fixes all import paths in .go files
func (pc *ProjectCreator) fixImports() error {
	return filepath.Walk(pc.TargetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		return pc.fixImportsInFile(path)
	})
}

// fixImportsInFile fixes imports in a single Go file
func (pc *ProjectCreator) fixImportsInFile(filePath string) error {
	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	originalContent := string(content)

	// Find and replace import paths
	// Pattern: "github.com/primadi/lokstra/project_templates/{template_path}/..."
	// Replace with: "{project_name}/..."

	templateImportPrefix := fmt.Sprintf("github.com/primadi/lokstra/project_templates/%s", pc.TemplatePath)
	newImportPrefix := pc.ProjectName

	updatedContent := strings.ReplaceAll(originalContent, templateImportPrefix, newImportPrefix)

	// Only write if content changed
	if updatedContent != originalContent {
		if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
			return err
		}
	}

	return nil
}

// initGoModule initializes a new Go module
func (pc *ProjectCreator) initGoModule() error {
	// Change to target directory
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(pc.TargetDir); err != nil {
		return err
	}

	// Run go mod init
	return runCommand("go", "mod", "init", pc.ProjectName)
}

// runGoModTidy runs go mod tidy
func (pc *ProjectCreator) runGoModTidy() error {
	// Change to target directory
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(pc.TargetDir); err != nil {
		return err
	}

	// Run go mod tidy
	return runCommand("go", "mod", "tidy")
}
