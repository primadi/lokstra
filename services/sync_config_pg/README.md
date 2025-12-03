# SyncConfig PostgreSQL Service

PostgreSQL-based synchronized configuration store with real-time updates across multiple instances using LISTEN/NOTIFY and CRC heartbeat validation.

## Features

- **Real-time Sync**: Uses PostgreSQL LISTEN/NOTIFY for instant config updates across instances
- **CRC Heartbeat**: Periodic CRC32 checksum validation every 5 minutes to detect missed notifications
- **Auto Recovery**: Automatically re-syncs when CRC mismatch is detected
- **Local Cache**: Fast in-memory access with database persistence
- **Change Callbacks**: Subscribe to configuration changes
- **Type-safe Getters**: Convenient methods for string, int, and bool values
- **Connection Resilience**: Auto-reconnect on connection loss

## Use Cases

- Multi-instance application configuration
- Feature flags that need instant propagation
- Dynamic settings without restart
- Distributed cache invalidation triggers
- Real-time configuration management

## Configuration

### YAML Configuration

```yaml
service-definitions:
  # First, define the database pool
  my-db-pool:
    type: dbpool_pg
    params:
      dsn: postgres://user:pass@localhost:5432/mydb?sslmode=disable
      schema: public
  
  # Then, define sync config that uses the pool
  config-service:
    type: sync_config_pg
    depends-on: [dbpool-manager]
    params:
      dbpool_name: my-db-pool
      table_name: app_config
      channel: config_updates
      heartbeat_interval: 5  # minutes
      sync_on_mismatch: true
      enable_notification: true
```

### Configuration Options

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `dbpool_name` | string | **required** | Name of the DbPool to use (from DbPoolManager) |
| `table_name` | string | `sync_config` | Table name for storing configs |
| `channel` | string | `config_changes` | PostgreSQL NOTIFY channel |
| `heartbeat_interval` | int | `5` | CRC heartbeat interval in minutes |
| `reconnect_interval` | int | `10` | Reconnect attempt interval in seconds |
| `sync_on_mismatch` | bool | `true` | Auto sync when CRC mismatch detected |
| `enable_notification` | bool | `true` | Enable LISTEN/NOTIFY (disable for single instance) |

## How It Works

### 1. Initial Load
- Creates table if not exists
- Loads all existing configs into memory
- Calculates initial CRC32 checksum
- Starts LISTEN on notification channel

### 2. Set/Delete Operations
```
Instance A: Set("key", "value")
    ↓
1. Update database
2. Update local cache
3. Calculate new CRC
4. NOTIFY other instances
5. Trigger local callbacks
```

### 3. Notification Handling
```
Instance B receives NOTIFY
    ↓
1. Update local cache
2. Calculate new CRC
3. Compare CRC with sender
4. Trigger local callbacks
5. If CRC mismatch → Full sync
```

### 4. CRC Heartbeat
```
Every 5 minutes:
    ↓
1. Broadcast current CRC
2. Other instances compare
3. If mismatch → Full sync
```

This ensures no missed notifications due to network issues or brief disconnections.

## Usage

### Basic Usage

```go
import (
    "context"
    "github.com/primadi/lokstra/serviceapi"
    "github.com/primadi/lokstra/services/sync_config_pg"
    "github.com/primadi/lokstra/services/dbpool_manager"
)

// Create DB Pool Manager
poolManager := dbpool_manager.NewPgxPoolManager()
poolManager.RegisterNamed("my-db-pool", "postgres://localhost:5432/mydb?sslmode=disable", "public")

cfg := &sync_config_pg.Config{
    DbPoolName: "my-db-pool",
    TableName:  "app_config",
}

configService, err := sync_config_pg.Service(cfg, poolManager)
if err != nil {
    panic(err)
}
defer configService.Shutdown()

ctx := context.Background()

// Set configuration
err = configService.Set(ctx, "feature_flag", true)

// Get with type safety
enabled := configService.GetBool(ctx, "feature_flag", false)
maxUsers := configService.GetInt(ctx, "max_users", 100)
appName := configService.GetString(ctx, "app_name", "MyApp")
```

### With Annotation (@RouterService)

```go
// @RouterService name="settings-service"
type SettingsService struct {
    // @Inject "config-service"
    Config serviceapi.SyncConfig
}

// @Route "PUT /settings/{key}"
func (s *SettingsService) UpdateSetting(p *UpdateSettingParams) error {
    return s.Config.Set(context.Background(), p.Key, p.Value)
}

// @Route "GET /settings/{key}"
func (s *SettingsService) GetSetting(p *GetSettingParams) (any, error) {
    return s.Config.Get(context.Background(), p.Key)
}

// @Route "GET /settings"
func (s *SettingsService) GetAllSettings() (map[string]any, error) {
    return s.Config.GetAll(context.Background())
}
```

### Subscribe to Changes

```go
// Subscribe to all config changes
subscriptionID := configService.Subscribe(func(key string, value any) {
    log.Printf("Config changed: %s = %v", key, value)
    
    // React to specific keys
    if key == "maintenance_mode" {
        if value.(bool) {
            startMaintenanceMode()
        } else {
            stopMaintenanceMode()
        }
    }
})

// Unsubscribe when done
defer configService.Unsubscribe(subscriptionID)
```

### Feature Flags

```go
type FeatureFlagService struct {
    Config serviceapi.SyncConfig
}

func (s *FeatureFlagService) IsEnabled(flagName string) bool {
    return s.Config.GetBool(
        context.Background(),
        "feature:"+flagName,
        false,
    )
}

func (s *FeatureFlagService) Enable(flagName string) error {
    return s.Config.Set(
        context.Background(),
        "feature:"+flagName,
        true,
    )
}

// Usage
if featureFlags.IsEnabled("new_ui") {
    renderNewUI()
} else {
    renderOldUI()
}
```

### Dynamic Rate Limiting

```go
type RateLimiterService struct {
    Config serviceapi.SyncConfig
}

func (s *RateLimiterService) GetLimit(endpoint string) int {
    return s.Config.GetInt(
        context.Background(),
        "rate_limit:"+endpoint,
        100, // default
    )
}

// Admin endpoint to update rate limits
func (s *RateLimiterService) UpdateLimit(endpoint string, limit int) error {
    return s.Config.Set(
        context.Background(),
        "rate_limit:"+endpoint,
        limit,
    )
}
```

### Multi-Instance Application

```go
// Instance 1
config1.Set(ctx, "max_connections", 500)

// Instance 2, 3, 4, ... all receive update within milliseconds
// No restart needed!

// Even if a notification is missed:
// - CRC heartbeat every 5 minutes will detect mismatch
// - Automatic full sync will be triggered
// - All instances stay in sync
```

## Database Schema

```sql
CREATE TABLE sync_config (
    key VARCHAR(255) PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sync_config_updated_at ON sync_config (updated_at);
```

The table is created automatically if it doesn't exist.

## CRC Validation

The service uses CRC32 checksum for data integrity:

```go
// Get current CRC
crc := configService.GetCRC()

// CRC is calculated from sorted keys and values
// Any change in data = different CRC
// Heartbeat broadcasts CRC every 5 minutes
// Instances compare and sync if mismatch detected
```

## Performance

- **Read**: O(1) from in-memory cache
- **Write**: O(1) database update + O(n) for CRC recalculation
- **Sync**: O(n) full table scan (only on mismatch or startup)
- **Memory**: ~50-100 bytes per config entry

## Best Practices

1. **Use Namespaces**: Prefix keys with category (`feature:`, `limit:`, `cache:`)
2. **Keep Values Small**: Store references, not large objects
3. **Handle Callbacks Async**: Don't block in subscription callbacks
4. **Graceful Shutdown**: Always call `Shutdown()` to cleanup connections
5. **Monitor CRC**: Log CRC mismatches to detect issues
6. **Connection Pooling**: Use appropriate `max_connections` in DSN
7. **Indexes**: Keep the updated_at index for efficient sync queries

## Error Handling

```go
// Set with error handling
if err := config.Set(ctx, "key", "value"); err != nil {
    log.Printf("Failed to set config: %v", err)
    // Config may be temporarily unavailable
    // Use cached value or retry
}

// Get with default fallback
value := config.GetString(ctx, "key", "fallback")
// Always succeeds, returns default if key not found

// Check if key exists
_, err := config.Get(ctx, "key")
if err != nil {
    // Key doesn't exist
}
```

## Monitoring

```go
// Health check endpoint
func (s *HealthService) CheckConfig() map[string]any {
    crc := s.Config.GetCRC()
    all, _ := s.Config.GetAll(context.Background())
    
    return map[string]any{
        "status": "healthy",
        "crc":    crc,
        "count":  len(all),
    }
}
```

## Troubleshooting

### Configs Not Syncing

1. Check LISTEN/NOTIFY is enabled: `enable_notification: true`
2. Verify PostgreSQL version supports LISTEN/NOTIFY (9.0+)
3. Check network between instances and database
4. Look for listener errors in logs
5. Verify all instances use same `channel` name

### CRC Mismatches

- Normal during brief network issues
- Auto-recovers via heartbeat sync
- If persistent, check:
  - Database connectivity
  - Concurrent modifications outside the service
  - Clock skew between instances

### High Memory Usage

- Reduce number of config entries
- Use external storage for large values
- Implement key expiration if needed

### Slow Startup

- Too many config entries to load
- Slow database query
- Add pagination if > 10,000 entries

## Example: Complete Application

See [EXAMPLES.md](./EXAMPLES.md) for complete working examples including:
- Feature flag system
- Dynamic rate limiter
- Multi-tenant configuration
- Cache invalidation
- A/B testing configuration
