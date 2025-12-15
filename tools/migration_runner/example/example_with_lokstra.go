package main

import (
	"log"

	"github.com/primadi/lokstra/lokstra_init"
)

// Example of using lokstra.CheckDbMigration in your main application
// This is the recommended way for automatic migration in development
func ExampleUsage() {
	// Bootstrap Lokstra first
	lokstra_init.Bootstrap()

	// Option 1: Auto mode (RECOMMENDED for development)
	// - Runs migrations in dev/debug mode automatically
	// - Skips migrations in prod mode (use CLI: lokstra migration up)
	err := lokstra_init.CheckDbMigration(&lokstra_init.MigrationConfig{
		MigrationsDir: "migrations",
	})
	if err != nil {
		log.Fatalf("Migration check failed: %v", err)
	}

	// Option 2: Always run (even in prod - NOT RECOMMENDED)
	err = lokstra_init.CheckDbMigration(&lokstra_init.MigrationConfig{
		MigrationsDir: "migrations",
	})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Option 3: Custom database pool
	err = lokstra_init.CheckDbMigration(&lokstra_init.MigrationConfig{
		MigrationsDir: "db/migrations",
		DbPoolName:    "analytics-db", // from config.yaml named-db-pools
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
	lokstra_init.Bootstrap()

	// 2. Auto-run migrations in dev/debug, skip in prod
	if err := lokstra_init.CheckDbMigration(nil); err != nil {
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
