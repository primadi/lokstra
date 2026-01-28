package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// SkillsUpdater handles updating AI agent skills and templates in existing projects
type SkillsUpdater struct {
	ProjectDir string
	Branch     string
}

// NewSkillsUpdater creates a new SkillsUpdater instance
func NewSkillsUpdater(projectDir, branch string) *SkillsUpdater {
	return &SkillsUpdater{
		ProjectDir: projectDir,
		Branch:     branch,
	}
}

// Update downloads and updates skills and templates
func (su *SkillsUpdater) Update() error {
	// Step 1: Download framework from GitHub
	fmt.Println("📥 Downloading latest framework from GitHub...")
	tempDir, err := su.downloadFramework()
	if err != nil {
		return fmt.Errorf("failed to download framework: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Step 2: Backup existing files
	fmt.Println("💾 Backing up existing files...")
	if err := su.backupExisting(); err != nil {
		fmt.Printf("⚠️  Warning: failed to backup: %v\n", err)
	}

	// Step 3: Copy skills and templates
	fmt.Println("📋 Updating files...")
	if err := su.copySkillsAndTemplates(tempDir); err != nil {
		return fmt.Errorf("failed to update files: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Skills and templates updated successfully!")
	fmt.Println()
	fmt.Println("Updated files:")
	fmt.Println("  - .github/skills/")
	fmt.Println("  - .github/copilot-instructions.md")
	fmt.Println("  - docs/templates/")
	fmt.Println()

	return nil
}

// downloadFramework downloads the framework from GitHub
func (su *SkillsUpdater) downloadFramework() (string, error) {
	url := fmt.Sprintf("https://github.com/primadi/lokstra/archive/refs/heads/%s.tar.gz", su.Branch)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	tempDir, err := os.MkdirTemp("", "lokstra-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Extract tar.gz
	pc := &ProjectCreator{} // Reuse extraction logic
	if err := pc.extractTarGz(resp.Body, tempDir); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to extract archive: %w", err)
	}

	extractedDir := filepath.Join(tempDir, fmt.Sprintf("lokstra-%s", su.Branch))
	return extractedDir, nil
}

// backupExisting creates backup of existing files
func (su *SkillsUpdater) backupExisting() error {
	backupDir := filepath.Join(su.ProjectDir, ".github", "skills.backup")

	// Backup skills directory
	skillsDir := filepath.Join(su.ProjectDir, ".github", "skills")
	if _, err := os.Stat(skillsDir); err == nil {
		os.RemoveAll(backupDir) // Remove old backup
		if err := copyDirectoryRecursive(skillsDir, backupDir); err != nil {
			return err
		}
		fmt.Println("  ✓ Backed up .github/skills/ to .github/skills.backup/")
	}

	return nil
}

// copySkillsAndTemplates copies skills and templates from downloaded framework
func (su *SkillsUpdater) copySkillsAndTemplates(frameworkDir string) error {
	// Copy .github/skills/
	skillsSource := filepath.Join(frameworkDir, ".github", "skills")
	skillsTarget := filepath.Join(su.ProjectDir, ".github", "skills")

	if _, err := os.Stat(skillsSource); err == nil {
		// Remove existing skills
		os.RemoveAll(skillsTarget)

		if err := copyDirectoryRecursive(skillsSource, skillsTarget); err != nil {
			return fmt.Errorf("failed to copy skills: %w", err)
		}
		fmt.Println("  ✓ Updated .github/skills/")
	} else {
		return fmt.Errorf("skills directory not found in framework")
	}

	// Copy .github/copilot-instructions.md
	copilotSource := filepath.Join(frameworkDir, ".github", "copilot-instructions.md")
	copilotTarget := filepath.Join(su.ProjectDir, ".github", "copilot-instructions.md")

	if _, err := os.Stat(copilotSource); err == nil {
		os.MkdirAll(filepath.Join(su.ProjectDir, ".github"), 0755)

		if err := copyFile(copilotSource, copilotTarget, 0644); err != nil {
			return fmt.Errorf("failed to copy copilot-instructions.md: %w", err)
		}
		fmt.Println("  ✓ Updated .github/copilot-instructions.md")
	}

	// Copy docs/templates/
	templatesSource := filepath.Join(frameworkDir, "docs", "templates")
	templatesTarget := filepath.Join(su.ProjectDir, "docs", "templates")

	if _, err := os.Stat(templatesSource); err == nil {
		// Remove existing templates
		os.RemoveAll(templatesTarget)

		if err := copyDirectoryRecursive(templatesSource, templatesTarget); err != nil {
			return fmt.Errorf("failed to copy templates: %w", err)
		}
		fmt.Println("  ✓ Updated docs/templates/")
	} else {
		return fmt.Errorf("templates directory not found in framework")
	}

	return nil
}

// Helper functions

func copyDirectoryRecursive(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirectoryRecursive(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if err := copyFile(srcPath, dstPath, info.Mode()); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
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
