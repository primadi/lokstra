---
name: implementation-lokstra-create-migrations
description: Create SQL migration files for database schema. Generate .up.sql and .down.sql files using Lokstra CLI. Supports single and multi-database setups with migration.yaml configuration.
phase: implementation
order: 5
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.18"
  database: postgresql
---

# Implementation: Database Migrations

## When to Use

Use this skill when:
- Creating initial database schema from design spec
- Adding new tables, columns, or indexes
- Managing schema changes for single or multi-database systems
- Setting up migration file structure

Prerequisites:
- ✅ Database schema designed (see: design-lokstra-schema-design)
- ✅ Domain models finalized
- ✅ Database service configured in config.yaml (see: implementation-lokstra-yaml-config)
- ✅ Database pool defined in `service-definitions` section

---

## Quick Start: Create Migration

```bash
# Create new migration (auto-versioned)
lokstra migration create create_users_table

# Output:
# ✅ Created migration files:
#    migrations/001_create_users_table.up.sql
#    migrations/001_create_users_table.down.sql
```

This creates properly numbered migration files. Edit them with your SQL, then run:

```bash
lokstra migration up
```

---

## Migration File Structure

### Option 1: Flat Structure (Single Database)

Recommended for most applications with one database:

```
migrations/
├── migration.yaml              # Optional: DB pool config
├── 001_create_users.up.sql
├── 001_create_users.down.sql
├── 002_add_email_verified.up.sql
├── 002_add_email_verified.down.sql
├── 003_create_orders.up.sql
└── 003_create_orders.down.sql
```

### Option 2: Subfolder Structure (Multi-Database)

For systems with multiple databases (main-db, tenant-db, analytics-db):

```
migrations/
├── 01_main-db/                 # Runs first (alphabetical order)
│   ├── migration.yaml          # REQUIRED for subfolder detection
│   ├── 001_create_users.up.sql
│   └── 001_create_users.down.sql
├── 02_tenant-db/               # Runs second
│   ├── migration.yaml          # REQUIRED
│   ├── 001_create_tenants.up.sql
│   └── 001_create_tenants.down.sql
└── 03_analytics-db/            # Runs third
    ├── migration.yaml          # REQUIRED
    └── 001_create_events.up.sql
```

**Important**: 
- Each subfolder MUST contain `migration.yaml` to be detected as a migration folder.
- Use numeric prefixes (01_, 02_, 03_) for explicit execution order.
- Subfolders without `migration.yaml` are ignored.

---

## migration.yaml Configuration

### Fields (matches `lokstra_init.MigrationYamlConfig`)

```yaml
# Database pool name from config.yaml service-definitions
# REQUIRED for multi-database, recommended for single database
dbpool-name: db_main

# Schema migrations tracking table name
# Default: schema_migrations
schema-table: schema_migrations

# Enable/disable this migration folder
# Set to false to skip without deleting files (useful for maintenance)
# Default: true (nil = enabled)
enabled: true

# Description for documentation
description: Main application database
```

### Example: Single Database

File: `migrations/migration.yaml`

```yaml
dbpool-name: db_main
description: Application database migrations
```

### Example: Multi-Database Setup

File: `migrations/01_main-db/migration.yaml`

```yaml
dbpool-name: db_main
schema-table: schema_migrations
description: Main application database with users and orders
```

File: `migrations/02_analytics-db/migration.yaml`

```yaml
dbpool-name: db_analytics
schema-table: analytics_migrations
description: Analytics and reporting database
```

File: `migrations/03_ledger-db/migration.yaml`

```yaml
dbpool-name: db_ledger
schema-table: ledger_migrations
enabled: false  # Disabled - use CLI manually for critical ledger DB
description: General ledger (manual migrations only)
```

---

## Naming Convention

```
{version}_{description}.{up|down}.sql

Rules:
- version: 3-digit zero-padded number (001, 002, 003)
- description: lowercase snake_case (letters, numbers, underscores only)
- direction: up or down
```

**Valid Examples:**
```
001_create_users.up.sql          ✅
001_create_users.down.sql        ✅
002_add_email_verified.up.sql    ✅
015_create_audit_logs.up.sql     ✅
```

**Conflict Detection (will error):**
```
001_create_users.up.sql          ❌ Conflict!
001_create_orders.up.sql         ❌ Same version (001), different description
```

Lokstra enforces: **One version = One description**. If you need multiple changes, use separate version numbers.

---

## Lokstra CLI Commands

### Create New Migration

```bash
# Create in default directory (migrations/)
lokstra migration create create_users_table

# Create in specific directory
lokstra migration create add_orders -dir migrations/01_main-db
```

**Output:**
```
✅ Created migration files:
   migrations/001_create_users_table.up.sql
   migrations/001_create_users_table.down.sql

📝 Next steps:
   1. Edit the migration files with your SQL
   2. Run: lokstra migration up
```

### Run Migrations (Up)

```bash
# Run all pending migrations (uses migration.yaml if exists)
lokstra migration up

# Run in specific directory
lokstra migration up -dir migrations/01_main-db

# Specify database pool explicitly (overrides migration.yaml)
lokstra migration up -dir migrations -db db_main
```

### Rollback Migrations (Down)

```bash
# Rollback last migration
lokstra migration down

# Rollback N migrations
lokstra migration down -steps 3

# Rollback specific folder
lokstra migration down -dir migrations/01_main-db -steps 1
```

### Check Status

```bash
lokstra migration status

# Output:
# Migration Status:
# =================
# [✓] 001: create_users
# [✓] 002: add_email_verified
# [ ] 003: create_orders
# 
# Total: 3 migrations, 2 applied, 1 pending
```

### Get Current Version

```bash
lokstra migration version

# Output: Current version: 002
# (or "No migrations applied yet")
```

### Database Pool Resolution Order

1. If `-db <name>` flag specified → use it
2. Else if `migration.yaml` exists with `dbpool-name` → use it
3. Else → error (database pool must be specified)

---

## Programmatic Migration (main.go)

### Single Database (Auto-Migrate on Startup)

```go
package main

import (
    "github.com/primadi/lokstra/lokstra_init"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    // Bootstrap loads config and initializes services
    lokstra_init.Bootstrap()

    // Run migrations from default directory (migrations/)
    // Uses migration.yaml if present
    if err := lokstra_init.CheckDbMigration(&lokstra_init.MigrationConfig{
        MigrationsDir: "migrations",
        DbPoolName:    "db_main",  // Optional if migration.yaml has dbpool-name
    }); err != nil {
        panic(err)
    }

    // Start server
    lokstra_registry.RunServerFromConfig()
}
```

### Multi-Database (Auto-Scan All Subfolders)

```go
package main

import (
    "github.com/primadi/lokstra/lokstra_init"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    lokstra_init.Bootstrap()

    // Auto-scan and run all database migrations
    // Scans for subfolders with migration.yaml, runs in alphabetical order
    if err := lokstra_init.CheckDbMigrationsAuto("migrations"); err != nil {
        panic(err)
    }

    lokstra_registry.RunServerFromConfig()
}
```

**`CheckDbMigrationsAuto` behavior:**
- Scans `migrations/` for subfolders containing `migration.yaml`
- Runs them in alphabetical order (use numeric prefixes: 01_, 02_, 03_)
- Skips folders with `enabled: false`
- Reports success/error/skipped counts
- If no subfolders found, treats root as single migration folder

### Manual Control (Per Database)

```go
package main

import (
    "github.com/primadi/lokstra/lokstra_init"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    lokstra_init.Bootstrap()

    // Run migrations for each database manually (for specific ordering/control)
    lokstra_init.CheckDbMigration(&lokstra_init.MigrationConfig{
        MigrationsDir: "migrations/01_tenant-db",
    })

    lokstra_init.CheckDbMigration(&lokstra_init.MigrationConfig{
        MigrationsDir: "migrations/02_main-db",
    })

    lokstra_init.CheckDbMigration(&lokstra_init.MigrationConfig{
        MigrationsDir: "migrations/03_analytics-db",
    })

    lokstra_registry.RunServerFromConfig()
}
```

---

## Basic Migration Examples

### Create Table

File: `migrations/001_create_users.up.sql`

```sql
-- Migration: Create users table
-- Created: 2026-02-01

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,                              -- ULID recommended
    tenant_id TEXT NOT NULL,                          -- Multi-tenant (if needed)
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ                            -- Soft delete
);

-- Indexes
CREATE INDEX idx_users_tenant ON users(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email ON users(tenant_id, email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(tenant_id, created_at DESC);

-- Unique constraint scoped to tenant
CREATE UNIQUE INDEX uq_users_tenant_email 
    ON users(tenant_id, email) 
    WHERE deleted_at IS NULL;
```

File: `migrations/001_create_users.down.sql`

```sql
-- Rollback: Drop users table
DROP INDEX IF EXISTS uq_users_tenant_email;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_tenant;
DROP TABLE IF EXISTS users;
```

### Add Column

File: `migrations/002_add_email_verified.up.sql`

```sql
-- Add new column with default value
ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;

-- Partial index for filtering verified users
CREATE INDEX idx_users_verified 
    ON users(tenant_id, email_verified) 
    WHERE email_verified = TRUE AND deleted_at IS NULL;
```

File: `migrations/002_add_email_verified.down.sql`

```sql
-- Rollback: Drop column
DROP INDEX IF EXISTS idx_users_verified;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
```

### Create Related Table with Foreign Key

File: `migrations/003_create_orders.up.sql`

```sql
CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    total_amount NUMERIC(12, 2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'confirmed', 'shipped', 'delivered', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Composite foreign key (includes tenant_id for multi-tenant isolation)
    CONSTRAINT fk_orders_user 
        FOREIGN KEY (tenant_id, user_id) 
        REFERENCES users(tenant_id, id) 
        ON DELETE RESTRICT
);

-- Indexes
CREATE INDEX idx_orders_tenant ON orders(tenant_id);
CREATE INDEX idx_orders_user ON orders(tenant_id, user_id);
CREATE INDEX idx_orders_status ON orders(tenant_id, status);
CREATE INDEX idx_orders_created_at ON orders(tenant_id, created_at DESC);
```

File: `migrations/003_create_orders.down.sql`

```sql
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user;
DROP INDEX IF EXISTS idx_orders_tenant;
DROP TABLE IF EXISTS orders CASCADE;
```

---

## Shared Utilities Migration

File: `migrations/000_init_utilities.up.sql`

```sql
-- Create lokstra_core schema (for internal tracking)
CREATE SCHEMA IF NOT EXISTS lokstra_core;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create updated_at trigger function (reusable)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

File: `migrations/000_init_utilities.down.sql`

```sql
-- Rollback utilities (be careful - may break other tables)
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
-- DROP EXTENSION IF EXISTS "uuid-ossp";  -- Usually keep this
-- DROP SCHEMA IF EXISTS lokstra_core CASCADE;  -- Never drop if tracking table exists
```

---

## Add Trigger to Table

File: `migrations/004_add_users_updated_trigger.up.sql`

```sql
-- Attach updated_at trigger to users table
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

File: `migrations/004_add_users_updated_trigger.down.sql`

```sql
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
```

---

## Data Migration (Backfill)

File: `migrations/005_backfill_email_verified.up.sql`

```sql
-- Backfill: Set email_verified = true for old users
-- Use batching for large tables to avoid lock issues

UPDATE users 
SET email_verified = TRUE
WHERE created_at < NOW() - INTERVAL '30 days'
  AND email_verified = FALSE;
```

File: `migrations/005_backfill_email_verified.down.sql`

```sql
-- Rollback: This is not reversible without data loss
-- Consider: Do nothing, or reset to false if acceptable

-- UPDATE users 
-- SET email_verified = FALSE
-- WHERE created_at < NOW() - INTERVAL '30 days';

-- Better: Log a warning
DO $$
BEGIN
    RAISE NOTICE 'Backfill rollback: No action taken. Data changes are not reversible.';
END $$;
```

---

## Best Practices

### 1. Use IF NOT EXISTS / IF EXISTS

```sql
-- Safe to run multiple times (idempotent)
CREATE TABLE IF NOT EXISTS users (...);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
DROP TABLE IF EXISTS old_table;
DROP INDEX IF EXISTS old_index;
```

### 2. Always Include Audit Columns

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    -- ... business columns
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  -- Soft delete support
);
```

### 3. One Change Per Migration

```
✅ 001_create_users.up.sql         -- Only creates users table
✅ 002_add_email_verified.up.sql   -- Only adds one column
✅ 003_create_orders.up.sql        -- Only creates orders table

❌ 001_create_users_and_orders.up.sql  -- Too many changes
```

### 4. Test Rollback Before Commit

```bash
# Apply migration
lokstra migration up

# Verify schema
lokstra migration status

# Test rollback
lokstra migration down -steps 1

# Verify clean rollback
lokstra migration status

# Re-apply if good
lokstra migration up
```

### 5. Keep Migrations Small

- Easier to troubleshoot failures
- Easier to rollback partial changes
- Clearer change history
- Faster to apply in CI/CD

### 6. Include tenant_id for Multi-Tenant Apps

```sql
-- Every business table needs tenant_id
CREATE TABLE products (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,  -- ← MANDATORY
    name VARCHAR(255) NOT NULL,
    -- ...
);

-- Indexes should include tenant_id first
CREATE INDEX idx_products_tenant_name 
    ON products(tenant_id, name);
```

### 7. Use Composite Foreign Keys for Multi-Tenant

```sql
-- Prevents cross-tenant data references
CONSTRAINT fk_orders_user 
    FOREIGN KEY (tenant_id, user_id) 
    REFERENCES users(tenant_id, id)
```

---

## Production Recommendations

### Never Auto-Migrate in Production

```yaml
# migrations/migration.yaml for production
dbpool-name: db_prod
enabled: false  # ← Disable auto-migration
description: Production database - use CLI only
```

### Use CLI for Production Deployments

```bash
# In CI/CD pipeline or deployment script
lokstra migration status -dir migrations -db db_prod

# Review pending migrations before applying
lokstra migration up -dir migrations -db db_prod
```

### Backup Before Migration

```bash
# Create backup before migrating
pg_dump -h localhost -U postgres -d myapp > backup_before_migration.sql

# Then apply migrations
lokstra migration up
```

---

## Multi-Database Example (E-Commerce)

```
migrations/
├── 01_main-db/           # Core application
│   ├── migration.yaml    # dbpool-name: db_main
│   ├── 001_create_users.up.sql
│   └── 002_create_products.up.sql
│
├── 02_orders-db/         # Orders & inventory
│   ├── migration.yaml    # dbpool-name: db_orders
│   ├── 001_create_orders.up.sql
│   └── 002_create_inventory.up.sql
│
└── 03_analytics-db/      # Reporting
    ├── migration.yaml    # dbpool-name: db_analytics
    └── 001_create_events.up.sql
```

Run all with:
```go
lokstra_init.CheckDbMigrationsAuto("migrations")
```

Or individually:
```bash
lokstra migration up -dir migrations/01_main-db
lokstra migration up -dir migrations/02_orders-db
lokstra migration up -dir migrations/03_analytics-db
```

---

## Troubleshooting

### Error: "database pool not found"

Check that your `config.yaml` has the database pool defined:

```yaml
service-definitions:
  db_main:
    type: dbpool
    driver: postgres
    dsn: "postgres://user:pass@localhost:5432/myapp?sslmode=disable"
```

### Error: "migrations directory not found"

Ensure the migrations directory exists:

```bash
mkdir -p migrations
```

### Error: "migration conflict detected"

You have two migration files with the same version number but different descriptions. Rename one to use a different version:

```bash
# Before (conflict):
001_create_users.up.sql
001_create_orders.up.sql  # ← Conflict!

# After (fixed):
001_create_users.up.sql
002_create_orders.up.sql  # ← Use next version
```

### Error: "no UP SQL for migration"

Every version needs both `.up.sql` and `.down.sql` files. Create the missing file:

```bash
lokstra migration create my_migration_name
# This creates both files automatically
```

---

## Next Steps

1. **Create @Handler endpoints** (see: implementation-lokstra-create-handler)
2. **Create @Service repositories** (see: implementation-lokstra-create-service)
3. **Write integration tests** (see: advanced-lokstra-tests)

---

## Related Skills

- [design-lokstra-schema-design](../design-lokstra-schema-design/SKILL.md) - Schema planning with multi-tenant patterns
- [implementation-lokstra-yaml-config](../implementation-lokstra-yaml-config/SKILL.md) - Database pool configuration
- [implementation-lokstra-create-handler](../implementation-lokstra-create-handler/SKILL.md) - HTTP endpoints
- [implementation-lokstra-create-service](../implementation-lokstra-create-service/SKILL.md) - Data access layer
