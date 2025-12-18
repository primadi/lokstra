# Migration Runner - Multi-Database Support

This document explains how to use `migration.yaml` configuration files for managing multiple database migrations in large systems.

## Overview

For large systems with multiple databases (e.g., tenant-db, ledger-db, analytics-db), you can organize migrations into separate folders, each with its own `migration.yaml` configuration.

## Directory Structure

```
migrations/
├── main-db/
│   ├── migration.yaml
│   ├── 001_create_users.up.sql
│   ├── 001_create_users.down.sql
│   └── 002_add_indexes.up.sql
│
├── tenant-db/
│   ├── migration.yaml
│   ├── 001_create_tenants.up.sql
│   └── 001_create_tenants.down.sql
│
└── ledger-db/
    ├── migration.yaml
    ├── 001_create_accounts.up.sql
    └── 001_create_accounts.down.sql
```

## migration.yaml Configuration

Each migration folder can have a `migration.yaml` file to configure database-specific settings:

```yaml
# Database pool name from config.yaml dbpool-definitions
dbpool-name: main-db

# Schema migrations tracking table
schema-table: schema_migrations

# Migration execution mode
# - "auto": Run in dev/debug, skip in prod (recommended)
# - "on": Always run (even in prod)
# - "off": Never run (use CLI only)
force: auto

# Description (optional)
description: Main application database
```

### Configuration Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `dbpool-name` | string | `"main-db"` | Database pool name from config.yaml |
| `schema-table` | string | `"schema_migrations"` | Table for tracking migrations |
| `force` | string | `"auto"` | Execution mode: `auto`, `on`, `off` |
| `description` | string | `""` | Documentation description |
| `depends-on` | []string | `[]` | Dependencies (future feature) |

## Usage in Code

### Single Database (Backward Compatible)

```go
func main() {
    lokstra.Bootstrap()
    
    // Automatically loads migration.yaml if it exists
    lokstra.CheckDbMigration(&lokstra.MigrationConfig{
        MigrationsDir: "migrations",
    })
    
    lokstra_registry.RunServerFromConfig()
}
```

### Multiple Databases

```go
func main() {
    lokstra.Bootstrap()
    
    // Each database loads its own migration.yaml
    lokstra.CheckDbMigration(&lokstra.MigrationConfig{
        MigrationsDir: "migrations/tenant-db",
    })
    
    lokstra.CheckDbMigration(&lokstra.MigrationConfig{
        MigrationsDir: "migrations/main-db",
    })
    
    lokstra.CheckDbMigration(&lokstra.MigrationConfig{
        MigrationsDir: "migrations/ledger-db",
    })
    
    lokstra_registry.RunServerFromConfig()
}
```

### Override YAML Config

Code parameters override YAML values:

```go
lokstra.CheckDbMigration(&lokstra.MigrationConfig{
    MigrationsDir: "migrations/ledger-db",
    Force: lokstra.MigrationForceTrue,  // Override yaml
    DbPoolName: "custom-db",            // Override yaml
})
```

## Force Modes Explained

### `auto` (Recommended for Development)

- **Dev/Debug Mode**: ✅ Runs migrations automatically
- **Production Mode**: ❌ Skips migrations
- **Use Case**: Safe for development, controlled in production

### `on` (Always Run)

- **Dev/Debug Mode**: ✅ Runs migrations
- **Production Mode**: ✅ Runs migrations
- **Use Case**: Testing, or if you're sure about auto-migrations in prod
- **⚠️ Warning**: Risky for production!

### `off` (Never Run)

- **Dev/Debug Mode**: ❌ Skips migrations
- **Production Mode**: ❌ Skips migrations
- **Use Case**: Critical databases that need manual control via CLI

## CLI Commands

Run migrations manually using the CLI:

```bash
# Single database
lokstra migration up -dir migrations/main-db
lokstra migration down -dir migrations/main-db -steps 1
lokstra migration status -dir migrations/main-db

# Different databases
lokstra migration up -dir migrations/tenant-db
lokstra migration up -dir migrations/ledger-db

# Create new migration
lokstra migration create add_user_profile -dir migrations/main-db
```

## Real-World Examples

### Example 1: E-Commerce System

```
migrations/
├── catalog-db/          # force: auto
│   └── migration.yaml   # Products, categories
├── orders-db/           # force: auto
│   └── migration.yaml   # Orders, cart
└── analytics-db/        # force: off
    └── migration.yaml   # Analytics (CLI only)
```

### Example 2: Multi-Tenant SaaS

```
migrations/
├── tenant-management/   # force: auto
│   └── migration.yaml   # Tenant registry
├── shared-data/         # force: auto
│   └── migration.yaml   # Shared lookup tables
└── tenant-template/     # force: off
    └── migration.yaml   # Template for new tenants
```

### Example 3: Financial System

```
migrations/
├── user-db/            # force: auto
│   └── migration.yaml  # User management
├── transaction-db/     # force: off
│   └── migration.yaml  # Transactions (critical!)
└── report-db/          # force: auto
    └── migration.yaml  # Reporting
```

## Best Practices

1. **Use `auto` for most databases** - Safe for development
2. **Use `off` for critical databases** - Ledger, transactions, billing
3. **Keep migration.yaml in version control** - Track configuration changes
4. **Document dependencies** - Use description field
5. **Test migrations in staging first** - Before production
6. **Use CLI for production** - Manual control and verification

## Backward Compatibility

The system is fully backward compatible:

- If `migration.yaml` doesn't exist, uses default values
- Old code without YAML support continues to work
- Explicit config parameters always override YAML

## Migration Priority

Configuration loading order (later overrides earlier):

1. **YAML config** (`migration.yaml` in migrations folder)
2. **Code config** (MigrationConfig parameters)
3. **Default values** (if nothing specified)

Example:

```go
// migration.yaml: dbpool-name: tenant-db, force: auto

lokstra.CheckDbMigration(&lokstra.MigrationConfig{
    MigrationsDir: "migrations/tenant-db",
    Force: lokstra.MigrationForceTrue,  // Overrides yaml
    // DbPoolName not specified -> uses yaml: tenant-db
})

// Result:
// - DbPoolName: "tenant-db" (from yaml)
// - Force: "true" (from code, overrides yaml)
// - SchemaTable: "schema_migrations" (default)
```

## See Also

- [MULTI_DATABASE_EXAMPLE.md](./MULTI_DATABASE_EXAMPLE.md) - Complete examples
- [main_multi_db.go](./example/main_multi_db.go) - Working code example
- [multi_db/](./example/multi_db/) - Example migration folders
