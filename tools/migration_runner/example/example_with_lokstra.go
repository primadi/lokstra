package main

import (
	"log"

	"github.com/primadi/lokstra"
)

// Example of using lokstra.CheckDbMigration in your main application
// This is the recommended way for automatic migration in development
func ExampleUsage() {
	// Bootstrap Lokstra first
	lokstra.Bootstrap()

	// Option 1: Auto mode (RECOMMENDED for development)
	// - Runs migrations in dev/debug mode automatically
	// - Skips migrations in prod mode (use CLI: lokstra migration up)
	err := lokstra.CheckDbMigration(&lokstra.MigrationConfig{
		MigrationsDir: "migrations",
		Force:         lokstra.MigrationForceAuto, // or leave empty (default)
	})
	if err != nil {
		log.Fatalf("Migration check failed: %v", err)
	}

	// Option 2: Always run (even in prod - NOT RECOMMENDED)
	err = lokstra.CheckDbMigration(&lokstra.MigrationConfig{
		MigrationsDir: "migrations",
		Force:         lokstra.MigrationForceTrue,
	})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Option 3: Never run (use CLI only)
	err = lokstra.CheckDbMigration(&lokstra.MigrationConfig{
		MigrationsDir: "migrations",
		Force:         lokstra.MigrationForceFalse,
	})
	if err != nil {
		log.Fatalf("Migration check failed: %v", err)
	}

	// Option 4: Custom database pool
	err = lokstra.CheckDbMigration(&lokstra.MigrationConfig{
		MigrationsDir: "db/migrations",
		DbPoolName:    "analytics-db", // from config.yaml named-db-pools
		Force:         lokstra.MigrationForceAuto,
	})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Option 5: Silent mode (no logs)
	err = lokstra.CheckDbMigration(&lokstra.MigrationConfig{
		MigrationsDir: "migrations",
		Force:         lokstra.MigrationForceAuto,
		Silent:        true,
	})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Start your application
	// lokstra_registry.RunServerFromConfig()
}

// Typical usage in a real application
func TypicalMain() {
	// 1. Bootstrap Lokstra (loads config, detects mode, etc)
	lokstra.Bootstrap()

	// 2. Auto-run migrations in dev/debug, skip in prod
	if err := lokstra.CheckDbMigration(nil); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	// 3. Start your servers
	// lokstra_registry.RunServerFromConfig()
}

// What happens in each mode:
//
// DEV MODE (runtime.mode = "dev"):
//   - MigrationForceAuto  → ✅ Runs migrations automatically
//   - MigrationForceTrue  → ✅ Runs migrations
//   - MigrationForceFalse → ❌ Skips migrations
//
// DEBUG MODE (runtime.mode = "debug"):
//   - MigrationForceAuto  → ✅ Runs migrations automatically
//   - MigrationForceTrue  → ✅ Runs migrations
//   - MigrationForceFalse → ❌ Skips migrations
//
// PROD MODE (runtime.mode = "prod"):
//   - MigrationForceAuto  → ❌ Skips migrations (use CLI)
//   - MigrationForceTrue  → ✅ Runs migrations (dangerous!)
//   - MigrationForceFalse → ❌ Skips migrations
//
// For production, use CLI instead:
//   lokstra migration up -db main-db -dir migrations
//   lokstra migration down -steps 1
//   lokstra migration status
