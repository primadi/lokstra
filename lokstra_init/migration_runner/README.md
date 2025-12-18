# Migration Runner

Simple database migration tool for Lokstra framework.

## Features

- ✅ File-based SQL migrations (up/down)
- ✅ Version tracking in database
- ✅ Supports multiple named databases
- ✅ Idempotent (safe to run multiple times)
- ✅ Transaction support per migration
- ✅ Integrates with Lokstra's dbpool-manager
- ✅ **Auto-generate migration files** with `create` command
- ✅ **Conflict detection** - prevents version collisions
- ✅ **Multiple statements** per file (tables, indexes, views, procedures)

## Usage

### 1. Create New Migration (Auto-versioned)

```bash
# Auto-generates version number and file pair
go run main.go --cmd=create --name="create_users_table"

# Output:
# ✅ Created migration files:
#    → 001_create_users_table.up.sql
#    → 001_create_users_table.down.sql
```

**Benefits:**
- 🎯 Auto-detects next version number
- 🎯 Creates both UP and DOWN files
- 🎯 Includes helpful templates
- 🎯 No manual version numbering needed

### 2. Migration File Naming

```
migrations/
├── 001_create_users_table.up.sql
├── 001_create_users_table.down.sql
├── 002_add_email_index.up.sql
├── 002_add_email_index.down.sql
└── 003_create_orders_table.up.sql
    003_create_orders_table.down.sql
```

### 2. Migration File Naming

Format: `{version}_{description}.{up|down}.sql`

- **version**: 3-digit numeric (001, 002, 003, etc.)
- **description**: Snake_case description
- **direction**: `up` or `down`

**Rules:**
- ✅ One version = One description (strictly enforced)
- ✅ One .up.sql and one .down.sql per version
- ❌ Duplicate versions with different descriptions = ERROR
- ❌ Multiple .up.sql or .down.sql for same version = ERROR

**Examples:**
```
001_create_users_table.up.sql    ✅
001_create_users_table.down.sql  ✅
002_add_email_index.up.sql       ✅
002_add_email_index.down.sql     ✅
```

**Conflicts (will error):**
```
001_create_users.up.sql     ❌ Conflict!
001_create_orders.up.sql    ❌ Same version, different description
```

### 3. Setup Migrations Directory (Manual)

```go
package main

import (
    "context"
    "log"
    
    "github.com/primadi/lokstra/tools/migration_runner"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

func main() {
    // Setup your app and load config
    lokstra_registry.LoadConfigFromFolder("config")
    
    // Get DB pool from named pools
    dbPool := lokstra_registry.GetService[serviceapi.DbPoolWithSchema]("main-db")
    
    // Create runner
    runner := migration_runner.New(dbPool, "migrations")
    
    // Run migrations
    ctx := context.Background()
    if err := runner.Up(ctx); err != nil {
        log.Fatal(err)
    }
    
    log.Println("✅ Migrations completed")
}
```

### 4. Available Commands

```bash
# Create new migration (recommended way)
go run main.go --cmd=create --name="add_user_phone_column"

# Run all pending UP migrations
go run main.go --cmd=up

# Rollback last migration (DOWN)
go run main.go --cmd=down

# Rollback N migrations
go run main.go --cmd=down --steps=2

# Get current version
go run main.go --cmd=version

# Get migration status
go run main.go --cmd=status
```

## Example Migration Files

### 001_create_users_table.up.sql
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

### 001_create_users_table.down.sql
```sql
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

## Integration with Lokstra Config

### config.yaml
```yaml
dbpool-definitions:
  main-db:
    host: localhost
    port: 5432
    database: myapp
    username: postgres
    password: secret
    schema: public
    min-conns: 2
    max-conns: 10
```

### main.go
```go
import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/tools/migration_runner"
    _ "github.com/primadi/lokstra/services/dbpool_manager"
)

func main() {
    lokstra.Bootstrap()
    
    // Load config (auto-discovers dbpool-definitions)
    lokstra_registry.LoadConfigFromFolder("config")
    
    // Run migrations before starting server
    dbPool := lokstra_registry.GetService[serviceapi.DbPoolWithSchema]("main-db")
    runner := migration_runner.New(dbPool, "migrations")
    
    if err := runner.Up(context.Background()); err != nil {
        log.Fatal("Migration failed:", err)
    }
    
    // Start server
    lokstra_registry.RunServerFromConfig()
}
```

## Best Practices

1. **Use `create` command** - Auto-generates version numbers, prevents conflicts
2. **Always include DOWN migrations** - Enables rollback capability
3. **Keep migrations small** - One logical change per migration
4. **Test locally first** - Verify migrations work before deploying
5. **Use transactions** - Each migration runs in a transaction (automatic)
6. **Version control** - Commit migration files to git
7. **Run before deployment** - Automate in CI/CD pipeline
8. **One version = One description** - Enforced by conflict detection

## Error Handling & Conflict Detection

### Version Conflict Detection

If two developers create migrations with the same version:

```bash
# Developer A
001_create_users.up.sql

# Developer B  
001_create_orders.up.sql

# When loading migrations:
❌ Migration conflict detected for version 001:
  Found: 001_create_users.*.sql
  Found: 001_create_orders.*.sql
  → Same version cannot have different descriptions!
  → Please rename one of the migrations to use a different version number.
```

**Solution:** Rename one migration to next available version:
```bash
mv 001_create_orders.up.sql 002_create_orders.up.sql
mv 001_create_orders.down.sql 002_create_orders.down.sql
```

### Missing DOWN Migration

```bash
# Current version: 004
# Trying to rollback to 003
# But 003_xxx.down.sql is missing:

❌ Cannot rollback migration 003_add_indexes
  → Missing file: 003_add_indexes.down.sql
  → Create the DOWN migration file or manually remove this version from schema_migrations table
```

**Solutions:**
1. Create the missing `.down.sql` file
2. Manually remove from database: `DELETE FROM schema_migrations WHERE version = 3`

## Schema Tracking

The runner creates a `schema_migrations` table:

```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);
```

This table tracks which migrations have been applied.

## Error Handling

- If a migration fails, the transaction is rolled back
- The migration version is NOT recorded
- Subsequent `Up()` calls will retry the failed migration
- Use `Status()` to check which migrations are pending
