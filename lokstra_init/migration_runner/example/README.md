# Migration Runner Example

This example demonstrates how to use the Lokstra migration runner tool.

## Setup

1. **Create PostgreSQL database:**
```bash
createdb migration_example
```

2. **Update config if needed:**
Edit `config/config.yaml` to match your database credentials.

## Usage

### Create new migration (Recommended)
```bash
# Auto-generates version and file pair
go run main.go --cmd=create --name="create_products_table"

# Output:
# ✅ Created migration files:
#    → 004_create_products_table.up.sql
#    → 004_create_products_table.down.sql
#
# 📝 Next steps:
#    1. Edit the migration files with your SQL
#    2. Run: go run main.go --cmd=up
```

### Run all pending migrations
```bash
go run main.go --cmd=up
```

### Check migration status
```bash
go run main.go --cmd=status
```

Output:
```
Migration Status:
=================

[✓] 001: create_users_table
[✓] 002: create_orders_table
[ ] 003: add_user_contact_info

Total: 3 migrations, 2 applied, 1 pending
```

### Get current version
```bash
go run main.go --cmd=version
```

### Rollback last migration
```bash
go run main.go --cmd=down
```

### Rollback multiple migrations
```bash
go run main.go --cmd=down --steps=2
```

### Use different database pool
```bash
go run main.go --cmd=up --db=replica-db
```

### Use different migrations directory
```bash
go run main.go --cmd=up --dir=../other-migrations
```

## Migration Files

Migrations are located in `migrations/` directory with naming pattern:
```
{version}_{description}.{up|down}.sql
```

Examples:
- `001_create_users_table.up.sql` - Creates users table
- `001_create_users_table.down.sql` - Drops users table
- `002_create_orders_table.up.sql` - Creates orders table
- `002_create_orders_table.down.sql` - Drops orders table

## Integration with Application

To run migrations automatically when your application starts:

```go
package main

import (
    "context"
    "log"
    
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
    "github.com/primadi/lokstra/tools/migration_runner"
    
    _ "github.com/primadi/lokstra/services/dbpool_manager"
)

func main() {
    lokstra.Bootstrap()
    lokstra_registry.LoadConfigFromFolder("config")
    
    // Run migrations before starting server
    dbPool := lokstra_registry.GetService[serviceapi.DbPoolWithSchema]("main-db")
    runner := migration_runner.New(dbPool, "migrations")
    
    ctx := context.Background()
    if err := runner.Up(ctx); err != nil {
        log.Fatal("Migration failed:", err)
    }
    
    // Start your application
    lokstra_registry.RunServerFromConfig()
}
```

## Troubleshooting

**Error: "Migration conflict detected for version XXX"**
- Two migration files have the same version but different descriptions
- This happens when multiple developers create migrations simultaneously
- Solution: Rename one of the conflicting migrations to use the next available version number

**Error: "Cannot rollback migration XXX - Missing file"**
- The `.down.sql` file is missing for a migration you're trying to rollback
- Solution 1: Create the missing `.down.sql` file with rollback SQL
- Solution 2: Manually remove from database: `DELETE FROM schema_migrations WHERE version = XXX`

**Error: "Database pool 'main-db' not found"**
- Check that `dbpool-definitions.main-db` is defined in your config.yaml
- Make sure you imported `_ "github.com/primadi/lokstra/services/dbpool_manager"`

**Error: "Migrations directory not found"**
- Check that the migrations directory exists
- Use `--dir` flag to specify a different directory

**Error: "Migration failed: failed to execute migration SQL"**
- Check the SQL syntax in your migration file
- Verify database connection and permissions
- Check previous migration logs for context
