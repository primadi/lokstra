package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/tools/migration_runner"

	// Import services
	_ "github.com/primadi/lokstra/services/dbpool_manager"
)

func MainTest() {
	// CLI flags
	command := flag.String("cmd", "up", "Migration command: up, down, status, version, create")
	dbName := flag.String("db", "main-db", "Named database pool from config")
	migrationsDir := flag.String("dir", "migrations", "Migrations directory")
	steps := flag.Int("steps", 1, "Number of migrations to rollback (for 'down' command)")
	name := flag.String("name", "", "Migration name for 'create' command (snake_case)")
	flag.Parse()

	// Create command doesn't need database connection
	if *command == "create" {
		if *name == "" {
			log.Fatal("❌ --name is required for create command\n" +
				"   Example: --cmd=create --name=\"create_users_table\"")
		}

		runner := migration_runner.New(nil, *migrationsDir)
		if err := runner.Create(*name); err != nil {
			log.Fatalf("❌ Failed to create migration: %v", err)
		}
		return
	}

	// Bootstrap and load config for other commands
	lokstra_init.Bootstrap()
	lokstra_registry.LoadConfig("config")

	// Get database pool
	pool, ok := lokstra_registry.GetServiceAny(*dbName)
	if !ok {
		log.Fatalf("❌ Database pool '%s' not found. Check your config.yaml named-db-pools section", *dbName)
	}

	dbPool, ok := pool.(serviceapi.DbPool)
	if !ok {
		log.Fatalf("❌ Service '%s' is not a DbPool", *dbName)
	}

	// Create migration runner
	runner := migration_runner.New(dbPool, *migrationsDir)
	ctx := context.Background()

	// Execute command
	switch *command {
	case "up":
		if err := runner.Up(ctx); err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}

	case "down":
		if err := runner.DownN(ctx, *steps); err != nil {
			log.Fatalf("❌ Rollback failed: %v", err)
		}

	case "status":
		status, err := runner.Status(ctx)
		if err != nil {
			log.Fatalf("❌ Failed to get status: %v", err)
		}
		fmt.Println(status)

	case "version":
		version, err := runner.Version(ctx)
		if err != nil {
			log.Fatalf("❌ Failed to get version: %v", err)
		}
		fmt.Printf("Current migration version: %d\n", version)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", *command)
		fmt.Fprintf(os.Stderr, "Valid commands: up, down, status, version\n")
		os.Exit(1)
	}
}
