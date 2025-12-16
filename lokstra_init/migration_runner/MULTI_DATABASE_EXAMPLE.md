# Multi-Database Migration Structure Example
#
# For large systems with multiple databases, you can organize migrations like this:
#
# migrations/
# ├── main-db/                    # Main application database
# │   ├── migration.yaml
# │   ├── 001_create_users.up.sql
# │   ├── 001_create_users.down.sql
# │   └── 002_add_indexes.up.sql
# │
# ├── tenant-db/                  # Tenant management
# │   ├── migration.yaml
# │   ├── 001_create_tenants.up.sql
# │   └── 001_create_tenants.down.sql
# │
# ├── analytics-db/               # Analytics/reporting
# │   ├── migration.yaml
# │   └── 001_create_events.up.sql
# │
# └── ledger-db/                  # General ledger
#     ├── migration.yaml
#     └── 001_create_accounts.up.sql

# =============================================================================
# USAGE IN YOUR APPLICATION
# =============================================================================

# Single Database (Backward Compatible):
# 
# func main() {
#     lokstra.Bootstrap()
#     
#     // Automatically loads migration.yaml if exists
#     lokstra.CheckDbMigration(&lokstra.MigrationConfig{
#         MigrationsDir: "migrations/main-db",
#     })
#     
#     lokstra_registry.RunServerFromConfig()
# }

# Multiple Databases (Manual):
#
# func main() {
#     lokstra.Bootstrap()
#     
#     // Run migrations for each database
#     lokstra.CheckDbMigration(&lokstra.MigrationConfig{
#         MigrationsDir: "migrations/tenant-db",
#     })
#     
#     lokstra.CheckDbMigration(&lokstra.MigrationConfig{
#         MigrationsDir: "migrations/main-db",
#     })
#     
#     lokstra.CheckDbMigration(&lokstra.MigrationConfig{
#         MigrationsDir: "migrations/analytics-db",
#     })
#     
#     lokstra_registry.RunServerFromConfig()
# }

# =============================================================================
# MIGRATION.YAML CONFIGURATION
# =============================================================================

# Each migration folder should have its own migration.yaml:

# migrations/main-db/migration.yaml:
dbpool-name: main-db
schema-table: schema_migrations
force: auto
description: Main application database

# migrations/tenant-db/migration.yaml:
# dbpool-name: tenant-db
# schema-table: tenant_migrations
# force: auto
# description: Tenant management database

# migrations/ledger-db/migration.yaml:
# dbpool-name: ledger-db
# schema-table: ledger_migrations
# force: off  # Never auto-run, use CLI only
# description: General ledger database

# =============================================================================
# CLI COMMANDS FOR MULTI-DATABASE
# =============================================================================

# Run migrations for specific database:
# lokstra migration up -dir migrations/main-db -db main-db
# lokstra migration up -dir migrations/tenant-db -db tenant-db

# Check status:
# lokstra migration status -dir migrations/main-db
# lokstra migration status -dir migrations/tenant-db

# Rollback:
# lokstra migration down -dir migrations/ledger-db -steps 1
