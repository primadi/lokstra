# SyncConfig PostgreSQL Migrations

This directory contains SQL migration files for the SyncConfig service.

## Migration Files

### 001_create_sync_config.sql
Creates the main `sync_config` table with:
- `key` (VARCHAR 255, PRIMARY KEY) - Configuration key
- `value` (JSONB) - Configuration value in JSON format
- `updated_at` (TIMESTAMP) - Last update timestamp
- Index on `updated_at` for query optimization

## How to Use

### Option 1: Automatic (Code Creates Table)
The service automatically creates the table on first run. No manual migration needed.

### Option 2: Manual Migration
If you prefer to run migrations manually before starting the service:

```bash
# Using psql
psql -U postgres -d yourdb -f migrations/001_create_sync_config.sql

# Or using migration runner
go run tools/migration_runner/main.go -dir services/sync_config_pg/migrations
```

### Option 3: Custom Table Name
If you use a custom table name in config, modify the table name in the SQL:

```yaml
service-definitions:
  sync-config:
    type: sync_config_pg
    params:
      table_name: my_custom_config  # Custom table name
```

Then edit the SQL file to use `my_custom_config` instead of `sync_config`.

## PostgreSQL LISTEN/NOTIFY Channel

The service uses PostgreSQL LISTEN/NOTIFY for real-time synchronization:
- **Default channel**: `config_changes`
- **Configurable** via `channel` parameter

No migration needed for LISTEN/NOTIFY - it's a PostgreSQL built-in feature.

## Notes

- The service auto-creates the table if it doesn't exist
- Manual migrations are optional but recommended for production
- JSONB type allows flexible value storage (strings, numbers, objects, arrays)
- GIN index on `value` column is optional (uncomment if you need JSONB queries)
