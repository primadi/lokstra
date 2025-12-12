package main

import (
	"log"

	"github.com/primadi/lokstra"
)

// Example: Multi-database migration setup
// This demonstrates how to use migration.yaml for different databases
func main() {
	// Bootstrap Lokstra (loads config, detects runtime mode)
	lokstra.Bootstrap()

	// load database and other configurations
	if err := lokstra.LoadConfigFromFolder("config"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	lokstra.UsePgxDbPoolManager(true)

	lokstra.UsePgxSyncConfig("db_main")
	lokstra.LoadNamedDbPoolsFromConfig()

	// OPTION 1: Auto-scan all migration folders (RECOMMENDED)
	// Scans multi_db/ for subdirectories, runs them in alphabetical order
	// Each folder loads its own migration.yaml configuration
	if err := lokstra.CheckDbMigrationsAuto("multi_db"); err != nil {
		log.Fatalf("Multi-database migrations failed: %v", err)
	}

	// OPTION 2: Manual per-folder (if you need custom control)
	/*
		if err := lokstra.CheckDbMigration(&lokstra.MigrationConfig{
			MigrationsDir: "multi_db/01_main-db",
		}); err != nil {
			log.Fatalf("Main DB migration failed: %v", err)
		}

		if err := lokstra.CheckDbMigration(&lokstra.MigrationConfig{
			MigrationsDir: "multi_db/02_tenant-db",
		}); err != nil {
			log.Fatalf("Tenant DB migration failed: %v", err)
		}

		if err := lokstra.CheckDbMigration(&lokstra.MigrationConfig{
			MigrationsDir: "multi_db/03_ledger-db",
		}); err != nil {
			log.Fatalf("Ledger DB migration failed: %v", err)
		}
	*/

	log.Println("Application ready")

	// For production ledger-db migrations, use CLI:
	// lokstra migration up -dir multi_db/ledger-db -db ledger-db
	// lokstra migration status -dir multi_db/ledger-db

	// Start your application servers
	if err := lokstra.RunConfiguredServer(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

/*
DIRECTORY STRUCTURE:

multi_db/
├── 01_main-db/                 # Runs FIRST (alphabetical order)
│   ├── migration.yaml          # force: auto
│   ├── 001_create_users.up.sql
│   └── 001_create_users.down.sql
│
├── 02_tenant-db/               # Runs SECOND
│   ├── migration.yaml          # force: auto
│   ├── 001_create_tenants.up.sql
│   └── 001_create_tenants.down.sql
│
└── 03_ledger-db/               # Runs THIRD
    ├── migration.yaml          # force: off (CLI only)
    ├── 001_create_accounts.up.sql
    └── 001_create_accounts.down.sql

EXECUTION ORDER:
- Folders are processed in alphabetical order
- Use numeric prefixes (01_, 02_, 03_) for explicit ordering
- No dependencies needed - order is explicit from folder names

MIGRATION.YAML EXAMPLES:

# multi_db/01_main-db/migration.yaml
dbpool-name: main-db
schema-table: schema_migrations
force: auto
description: Main application database

# multi_db/02_tenant-db/migration.yaml
dbpool-name: tenant-db
schema-table: tenant_migrations
force: auto
description: Tenant management database

# multi_db/03_ledger-db/migration.yaml
dbpool-name: ledger-db
schema-table: ledger_migrations
force: off  # Never auto-run, critical data
description: General ledger database

RUNTIME BEHAVIOR:

DEV/DEBUG MODE (with CheckDbMigrationsAuto):
1. 01_main-db: ✅ Auto-runs (force=auto)
2. 02_tenant-db: ✅ Auto-runs (force=auto)
3. 03_ledger-db: ❌ Skipped (force=off)

PRODUCTION MODE (with CheckDbMigrationsAuto):
1. 01_main-db: ❌ Skipped (force=auto, prod detected)
2. 02_tenant-db: ❌ Skipped (force=auto, prod detected)
3. 03_ledger-db: ❌ Skipped (force=off)

Execution is SEQUENTIAL in alphabetical order

CLI COMMANDS FOR PRODUCTION:
lokstra migration up -dir multi_db/01_main-db
lokstra migration up -dir multi_db/02_tenant-db
lokstra migration up -dir multi_db/03_ledger-db

OVERRIDE EXAMPLE:
You can still override yaml config in code:

lokstra.CheckDbMigration(&lokstra.MigrationConfig{
    MigrationsDir: "multi_db/03_ledger-db",
    Force: lokstra.MigrationForceTrue,  // Override yaml force=off
})

AUTO-SCAN BEHAVIOR:
- CheckDbMigrationsAuto("multi_db") scans all subdirectories
- Skips folders without .sql files
- Processes in alphabetical order
- Each folder loads its own migration.yaml
- Reports: "X successful, Y skipped"
*/
