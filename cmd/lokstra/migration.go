package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/tools/migration_runner"
)

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
	configFileFlag := migrationFlags.String("config", "config.yaml", "Lokstra config file")
	migDirFlag := migrationFlags.String("dir", "migrations", "Migrations directory")
	dbFlag := migrationFlags.String("db", "global-db", "Database pool name")
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
		executeMigrationCreate(os.Args[3], *migDirFlag)
		return
	}

	// Parse flags for other commands
	migrationFlags.Parse(os.Args[3:])

	// Execute migration command
	if err := executeMigration(subCmd, *configFileFlag, *migDirFlag,
		*dbFlag, *stepsFlag); err != nil {
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

func executeMigration(subCmd, configFile, migrationDir, dbPoolName string, steps int) error {
	cfgFile := utils.NormalizeWithWordkingDir(configFile)
	if !utils.IsFileExists(cfgFile) {
		cfgFile = utils.NormalizeWithWordkingDir("/config/config.yaml")
		if !utils.IsFileExists(cfgFile) {
			return fmt.Errorf("config file not found: %s", configFile)
		}
	}

	migDir := utils.NormalizeWithWordkingDir(migrationDir)
	if !utils.IsFileExists(migDir) {
		return fmt.Errorf("migrations directory not found: %s", migrationDir)
	}

	// load named-db-pools from config file
	if err := lokstra.LoadConfig(cfgFile); err != nil {
		return fmt.Errorf("failed to load config file '%s': %w", filepath.Base(cfgFile), err)
	}

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
	runner := migration_runner.New(pool, migDir)
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
