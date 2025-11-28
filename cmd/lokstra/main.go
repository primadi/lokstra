package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/annotation"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/tools/migration_runner"
)

const version = "1.0.1"

func main() {
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
	fmt.Println("  lokstra autogen|generate [folder]")
	fmt.Println("  lokstra migration|migrate <command> [flags]")
	fmt.Println("  lokstra version")
	fmt.Println("  lokstra help")
	fmt.Println()
	fmt.Println("Flags for 'new' command:")
	fmt.Println("  -template <name>    Template to use (optional, interactive if not specified)")
	fmt.Println("  -branch <name>      Git branch to download from (default: main)")
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
	// Get target folder (optional, defaults to current directory)
	targetFolder := "."
	if len(os.Args) >= 3 {
		targetFolder = os.Args[2]
	}

	// Execute autogen
	if err := executeAutogen(targetFolder); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func executeAutogen(targetFolder string) error {
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
	return generateCodeForFolder(absPath)
}

// generateCodeForFolder calls the annotation processor to generate code
func generateCodeForFolder(absPath string) error {
	// Delete all cache files first to force rebuild
	if err := deleteAllCacheFilesInFolder(absPath); err != nil {
		return fmt.Errorf("failed to delete cache files: %w", err)
	}

	// Process the folder recursively using annotation processor
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

func migrationCmd() {
	// Get migration subcommand
	if len(os.Args) < 3 {
		fmt.Println("Error: migration command is required")
		fmt.Println()
		fmt.Println("Available commands:")
		fmt.Println("  create <name>    Create new migration files")
		fmt.Println("  up               Run pending migrations")
		fmt.Println("  down             Rollback last migration")
		fmt.Println("  status           Show migration status")
		fmt.Println("  version          Show current version")
		os.Exit(1)
	}

	subCmd := os.Args[2]

	// Parse flags
	migrationFlags := flag.NewFlagSet("migration", flag.ExitOnError)
	dirFlag := migrationFlags.String("dir", "migrations", "Migrations directory")
	dbFlag := migrationFlags.String("db", "main-db", "Database pool name")
	stepsFlag := migrationFlags.Int("steps", 1, "Number of migrations to rollback")

	// Handle create command separately (doesn't need DB connection)
	if subCmd == "create" {
		if len(os.Args) < 4 {
			fmt.Println("Error: migration name is required")
			fmt.Println()
			fmt.Println("Usage: lokstra migration create <name>")
			fmt.Println("Example: lokstra migration create create_users_table")
			os.Exit(1)
		}

		migrationFlags.Parse(os.Args[4:])
		executeMigrationCreate(os.Args[3], *dirFlag)
		return
	}

	// Parse flags for other commands
	migrationFlags.Parse(os.Args[3:])

	// Execute migration command
	if err := executeMigration(subCmd, *dirFlag, *dbFlag, *stepsFlag); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func executeMigrationCreate(name, dir string) {
	// For create command, we don't need DB connection
	// Just create the files directly
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Error creating migrations directory: %v\n", err)
		os.Exit(1)
	}

	// Find next version number
	files, _ := os.ReadDir(dir)
	nextVersion := 1
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		var ver int
		fmt.Sscanf(file.Name(), "%03d_", &ver)
		if ver >= nextVersion {
			nextVersion = ver + 1
		}
	}

	// Create migration files
	upFile := fmt.Sprintf("%s/%03d_%s.up.sql", dir, nextVersion, name)
	downFile := fmt.Sprintf("%s/%03d_%s.down.sql", dir, nextVersion, name)

	upContent := fmt.Sprintf("-- Migration: %s\n-- Created: %s\n\n-- Add your UP migration SQL here\n",
		name, time.Now().Format("2006-01-02 15:04:05"))
	downContent := fmt.Sprintf("-- Migration: %s\n-- Created: %s\n\n-- Add your DOWN migration SQL here\n",
		name, time.Now().Format("2006-01-02 15:04:05"))

	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		fmt.Printf("Error creating UP migration: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		fmt.Printf("Error creating DOWN migration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Created migration files:\n")
	fmt.Printf("   %s\n", upFile)
	fmt.Printf("   %s\n", downFile)
}

func executeMigration(subCmd, dir, dbPoolName string, steps int) error {
	// Load Lokstra configuration
	lokstra.Bootstrap()

	// Get database pool
	poolAny, ok := lokstra_registry.GetServiceAny(dbPoolName)
	if !ok {
		return fmt.Errorf("database pool '%s' not found in registry", dbPoolName)
	}

	pool, ok := poolAny.(serviceapi.DbPoolWithSchema)
	if !ok {
		return fmt.Errorf("database pool '%s' does not implement DbPoolWithSchema interface", dbPoolName)
	}

	// Create migration runner
	runner := migration_runner.New(pool, dir)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute command
	switch subCmd {
	case "up":
		if err := runner.Up(ctx); err != nil {
			return fmt.Errorf("migration up failed: %w", err)
		}
		fmt.Println("✅ All migrations applied successfully")

	case "down":
		if err := runner.DownN(ctx, steps); err != nil {
			return fmt.Errorf("migration down failed: %w", err)
		}
		fmt.Printf("✅ Rolled back %d migration(s)\n", steps)

	case "status":
		status, err := runner.Status(ctx)
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}
		fmt.Println(status)

	case "version":
		version, err := runner.Version(ctx)
		if err != nil {
			return fmt.Errorf("failed to get version: %w", err)
		}
		if version == 0 {
			fmt.Println("No migrations applied yet")
		} else {
			fmt.Printf("Current version: %03d\n", version)
		}

	default:
		return fmt.Errorf("unknown migration command: %s", subCmd)
	}

	return nil
}
