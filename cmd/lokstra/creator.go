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
	UseLocal     bool // Use local framework files instead of downloading
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

	var tempDir string
	var err error

	// Step 2: Get template source (local or download)
	if pc.UseLocal {
		fmt.Println("📂 Using local framework files...")
		// Assume we're running from within lokstra-dev2
		// Find the root of lokstra framework
		tempDir, err = pc.getLocalFrameworkPath()
		if err != nil {
			return fmt.Errorf("failed to find local framework: %w", err)
		}
	} else {
		fmt.Println("📥 Downloading template from GitHub...")
		tempDir, err = pc.downloadTemplate()
		if err != nil {
			return fmt.Errorf("failed to download template: %w", err)
		}
		defer os.RemoveAll(tempDir)
	}

	// Step 3: Copy template files to target directory
	fmt.Println("📋 Copying template files...")
	templateSourceDir := filepath.Join(tempDir, "project_templates", pc.TemplatePath)
	if err := pc.copyTemplate(templateSourceDir); err != nil {
		return fmt.Errorf("failed to copy template: %w", err)
	}

	// Step 4: Copy AI agent skills and templates
	fmt.Println("🤖 Copying AI agent skills and templates...")
	if err := pc.copySkillsAndTemplates(tempDir); err != nil {
		// Non-fatal error - just warn
		fmt.Printf("⚠️  Warning: failed to copy skills and templates: %v\n", err)
	}

	// Step 5: Fix imports in all .go files
	fmt.Println("🔧 Fixing imports...")
	if err := pc.fixImports(); err != nil {
		return fmt.Errorf("failed to fix imports: %w", err)
	}

	// Step 6: Initialize go module
	fmt.Println("📦 Initializing Go module...")
	if err := pc.initGoModule(); err != nil {
		return fmt.Errorf("failed to initialize go module: %w", err)
	}

	// Step 7: Run go mod tidy
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
	// Normalize module name (replace backslash with forward slash for Windows)
	normalizedProjectName := strings.ReplaceAll(pc.ProjectName, "\\", "/")
	newImportPrefix := normalizedProjectName

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

	// Normalize module name (replace backslash with forward slash for Windows)
	moduleName := strings.ReplaceAll(pc.ProjectName, "\\", "/")

	// Run go mod init
	return runCommand("go", "mod", "init", moduleName)
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

// copySkillsAndTemplates copies .github/skills/ and docs/templates/ to the new project
func (pc *ProjectCreator) copySkillsAndTemplates(tempDir string) error {
	// Source paths in the downloaded repository
	lokstraRoot := tempDir // tempDir is already "lokstra-{branch}"

	// Copy .github/skills/
	skillsSource := filepath.Join(lokstraRoot, ".github", "skills")
	skillsTarget := filepath.Join(pc.TargetDir, ".github", "skills")

	if _, err := os.Stat(skillsSource); err == nil {
		if err := pc.copyDirectory(skillsSource, skillsTarget); err != nil {
			return fmt.Errorf("failed to copy skills: %w", err)
		}
		fmt.Println("  ✓ Copied AI agent skills to .github/skills/")
	} else {
		fmt.Println("  ⚠️  Skills directory not found in framework")
	}

	// Copy .github/copilot-instructions.md
	copilotSource := filepath.Join(lokstraRoot, ".github", "copilot-instructions.md")
	copilotTarget := filepath.Join(pc.TargetDir, ".github", "copilot-instructions.md")

	if _, err := os.Stat(copilotSource); err == nil {
		// Create .github directory if not exists
		if err := os.MkdirAll(filepath.Join(pc.TargetDir, ".github"), 0755); err != nil {
			return fmt.Errorf("failed to create .github directory: %w", err)
		}

		if err := pc.copyFile(copilotSource, copilotTarget, 0644); err != nil {
			return fmt.Errorf("failed to copy copilot-instructions.md: %w", err)
		}
		fmt.Println("  ✓ Copied .github/copilot-instructions.md")
	}

	// Copy docs/templates/
	templatesSource := filepath.Join(lokstraRoot, "docs", "templates")
	templatesTarget := filepath.Join(pc.TargetDir, "docs", "templates")

	if _, err := os.Stat(templatesSource); err == nil {
		if err := pc.copyDirectory(templatesSource, templatesTarget); err != nil {
			return fmt.Errorf("failed to copy templates: %w", err)
		}
		fmt.Println("  ✓ Copied document templates to docs/templates/")
	} else {
		fmt.Println("  ⚠️  Templates directory not found in framework")
	}

	return nil
}

// copyDirectory recursively copies a directory
func (pc *ProjectCreator) copyDirectory(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := pc.copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Get file info for mode
			info, err := entry.Info()
			if err != nil {
				return err
			}
			// Copy file
			if err := pc.copyFile(srcPath, dstPath, info.Mode()); err != nil {
				return err
			}
		}
	}

	return nil
}

// getLocalFrameworkPath finds the local framework root directory
func (pc *ProjectCreator) getLocalFrameworkPath() (string, error) {
	// Try to find go.mod with module github.com/primadi/lokstra
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Search up the directory tree
	dir := currentDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Read go.mod to check if it's lokstra framework
			content, err := os.ReadFile(goModPath)
			if err == nil && strings.Contains(string(content), "module github.com/primadi/lokstra") {
				return dir, nil
			}
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("lokstra framework root not found. Run with -local only from within lokstra-dev2 directory")
}
