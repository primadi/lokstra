package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/annotation"
)

const version = "1.0.2"

func main() {
	logger.SetLogLevel(logger.LogLevelInfo)

	// for debugging purpose
	if lokstra.DetectRunMode() != lokstra.RunModeProd {
		// use 04_sync_config template for testing
		os.Chdir(filepath.Join(utils.GetBasePath(), "../../project_templates/02_app_framework/04_sync_config"))
		// os.Args = slices.Concat(os.Args[:1], []string{"migration", "status"})
		os.Args = slices.Concat(os.Args[:1], []string{"version"})
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "new":
		newCmd()
	case "autogen", "generate":
		autogenCmd()
	case "migration", "migrate":
		migrationCmd()
	case "version":
		fmt.Printf("Lokstra CLI v%s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Lokstra CLI - Create new Lokstra projects from templates")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  lokstra new <project-name> [flags]")
	fmt.Println("  lokstra autogen|generate [folder] [flags]")
	fmt.Println("  lokstra migration|migrate <command> [flags]")
	fmt.Println("  lokstra version")
	fmt.Println("  lokstra help")
	fmt.Println()
	fmt.Println("Flags for 'new' command:")
	fmt.Println("  -template <name>    Template to use (optional, interactive if not specified)")
	fmt.Println("  -branch <name>      Git branch to download from (default: main)")
	fmt.Println()
	fmt.Println("Flags for 'generate' command:")
	fmt.Println("  -force              Force rebuild by deleting all cache files")
	fmt.Println()
	fmt.Println("Migration commands:")
	fmt.Println("  lokstra migration create <name>        Create new migration files")
	fmt.Println("  lokstra migration up [flags]           Run pending migrations")
	fmt.Println("  lokstra migration down [flags]         Rollback last migration")
	fmt.Println("  lokstra migration status [flags]       Show migration status")
	fmt.Println("  lokstra migration version [flags]      Show current version")
	fmt.Println()
	fmt.Println("Migration flags:")
	fmt.Println("  -dir <path>         Migrations directory (default: migrations)")
	fmt.Println("  -db <name>          Database pool name (default: main-db)")
	fmt.Println("  -steps <n>          Number of migrations to rollback (default: 1)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  lokstra new myapp")
	fmt.Println("  lokstra new myapp -template 02_app_framework/01_medium_system")
	fmt.Println("  lokstra new myapp -template 01_router/01_router_only -branch main")
	fmt.Println()
	fmt.Println("  lokstra autogen                 # Generate code in current directory")
	fmt.Println("  lokstra generate                # Generate code in current directory")
	fmt.Println("  lokstra autogen ./myproject     # Generate code in specific folder")
	fmt.Println("  lokstra generate ./myproject    # Generate code in specific folder")
	fmt.Println()
	fmt.Println("  lokstra migration create create_users_table")
	fmt.Println("  lokstra migration up")
	fmt.Println("  lokstra migration down -steps=2")
	fmt.Println("  lokstra migration status -db=replica-db")
}

func newCmd() {
	// Parse flags for 'new' command
	newFlags := flag.NewFlagSet("new", flag.ExitOnError)
	templateFlag := newFlags.String("template", "", "Template to use")
	branchFlag := newFlags.String("branch", "main", "Git branch to download from")

	// Get project name (first argument after 'new')
	if len(os.Args) < 3 {
		fmt.Println("Error: project name is required")
		fmt.Println()
		fmt.Println("Usage: lokstra new <project-name> [flags]")
		os.Exit(1)
	}

	projectName := os.Args[2]

	// Parse remaining flags
	newFlags.Parse(os.Args[3:])

	// Execute new command
	if err := executeNew(projectName, *templateFlag, *branchFlag); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func executeNew(projectName, templatePath, branch string) error {
	fmt.Printf("🚀 Creating new Lokstra project: %s\n\n", projectName)

	// If template not specified, show interactive selection
	if templatePath == "" {
		var err error
		templatePath, err = selectTemplate(branch)
		if err != nil {
			return err
		}
	}

	fmt.Printf("📦 Selected template: %s\n", templatePath)
	fmt.Printf("🌿 Branch: %s\n\n", branch)

	// Execute the creation process
	creator := NewProjectCreator(projectName, templatePath, branch)
	return creator.Create()
}

func autogenCmd() {
	// Parse flags
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	force := fs.Bool("force", false, "Force rebuild by deleting all cache files")
	fs.Parse(os.Args[2:])

	// Get target folder (optional, defaults to current directory)
	targetFolder := "."
	if fs.NArg() > 0 {
		targetFolder = fs.Arg(0)
	}

	// Execute autogen
	if err := executeAutogen(targetFolder, *force); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func executeAutogen(targetFolder string, force bool) error {
	fmt.Printf("🔧 Running code generation in: %s\n\n", targetFolder)

	// Convert to absolute path first
	absPath, err := filepath.Abs(targetFolder)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if target folder exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("folder does not exist: %s", absPath)
	}

	// Import annotation processor
	// Instead of running "go run . --generate-only", call annotation processor directly
	return generateCodeForFolder(absPath, force)
}

// generateCodeForFolder calls the annotation processor to generate code
func generateCodeForFolder(absPath string, force bool) error {
	// Delete all cache files if --force flag is set
	if force {
		fmt.Println("🗑️  Force rebuild: deleting all cache files")
		fmt.Println()
		if err := deleteAllCacheFilesInFolder(absPath); err != nil {
			return fmt.Errorf("failed to delete cache files: %w", err)
		}
	}

	// Process the folder recursively using annotation processor
	// Cache will be used automatically unless files changed or generated code was manually modified
	_, err := annotation.ProcessComplexAnnotations(
		[]string{absPath},
		0, // Use default worker count (CPU * 2)
		func(ctx *annotation.RouterServiceContext) error {
			fmt.Printf("Processing folder: %s\n", ctx.FolderPath)
			fmt.Printf("  - Skipped: %d files\n", len(ctx.SkippedFiles))
			fmt.Printf("  - Updated: %d files\n", len(ctx.UpdatedFiles))
			fmt.Printf("  - Deleted: %d files\n", len(ctx.DeletedFiles))

			// Generate code
			if err := annotation.GenerateCodeForFolder(ctx); err != nil {
				return err
			}

			return nil
		},
	)

	if err != nil {
		return fmt.Errorf("code generation failed: %w", err)
	}

	fmt.Println("✅ Code generation completed successfully")
	return nil
}

// deleteAllCacheFilesInFolder removes all zz_cache.lokstra.json files in target folder
func deleteAllCacheFilesInFolder(targetFolder string) error {
	return filepath.Walk(targetFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}
		if !info.IsDir() && info.Name() == "zz_cache.lokstra.json" {
			fmt.Printf("  Deleting cache: %s\n", path)
			os.Remove(path)
		}
		return nil
	})
}
