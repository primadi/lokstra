# SyncMap - Distributed Synchronized Map

SyncMap provides a distributed, type-safe map that automatically synchronizes across multiple nodes/servers.

## Features

- **Type-Safe**: Generic interface for compile-time type checking
- **Distributed**: Automatic synchronization across nodes via PostgreSQL NOTIFY
- **Real-time**: Subscribe to change notifications
- **Isolated**: Prefix-based key isolation for multiple maps
- **Singleton Backend**: Multiple SyncMaps share single synchronizer (efficient heartbeat)

## Installation

```go
import "github.com/primadi/lokstra/common/syncmap"
```

## Quick Start

```go
package main

import (
    "context"
    "github.com/primadi/lokstra/services/sync_config_pg"
    "github.com/primadi/lokstra/common/syncmap"
)

func main() {
    // Create backend SyncConfig (singleton per config)
    cfg := &sync_config_pg.Config{
        DbPoolName: "main-db",
        TableName:  "sync_config",
        Channel:    "config_changes",
    }
    
    backend, _ := sync_config_pg.NewSyncConfigPG(cfg)
    defer backend.Shutdown()
    
    // Create SyncMaps with different prefixes
    userCache := syncmap.NewSyncMap[User](backend, "users")
    settings := syncmap.NewSyncMap[string](backend, "settings")
    
    ctx := context.Background()
    
    // Use like a regular map
    userCache.Set(ctx, "john", User{Name: "John", Age: 30})
    settings.Set(ctx, "theme", "dark")
    
    // Get values
    user, exists, _ := userCache.Get(ctx, "john")
    theme, exists, _ := settings.Get(ctx, "theme")
}
```

## Usage Examples

### Basic Operations

```go
ctx := context.Background()
m := syncmap.NewSyncMap[string](backend, "my_prefix")

// Set
m.Set(ctx, "key1", "value1")

// Get
value, exists, err := m.Get(ctx, "key1")

// Delete
m.Delete(ctx, "key1")

// Has
if m.Has(ctx, "key1") {
    // key exists
}
```

### Complex Types

```go
type Config struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

configs := syncmap.NewSyncMap[Config](backend, "configs")

configs.Set(ctx, "db", Config{
    Host: "localhost",
    Port: 5432,
})

dbConfig, exists, _ := configs.Get(ctx, "db")
```

### Bulk Operations

```go
// Get all keys
keys, _ := m.Keys(ctx)

// Get all values
values, _ := m.Values(ctx)

// Get all entries
all, _ := m.All(ctx)

// Iterate
m.Range(ctx, func(key string, value string) bool {
    fmt.Printf("%s = %s\n", key, value)
    return true // continue iteration
})

// Clear all
m.Clear(ctx)

// Get count
count := m.Len(ctx)
```

### Subscriptions

```go
// Subscribe to changes
subID := m.Subscribe(func(key string, value string) {
    fmt.Printf("Changed: %s = %s\n", key, value)
})

// Unsubscribe later
defer m.Unsubscribe(subID)
```

### Prefix Isolation

```go
// Multiple maps can share the same backend
// Each map only sees its own prefixed keys
userMap := syncmap.NewSyncMap[User](backend, "users")
orderMap := syncmap.NewSyncMap[Order](backend, "orders")

// Both can have "id123" key without conflict
userMap.Set(ctx, "id123", user)   // Repositoryd as "users:id123"
orderMap.Set(ctx, "id123", order) // Repositoryd as "orders:id123"
```

## Real-World Example: Pool Manager

```go
type PoolManager struct {
    tenantPools syncmap.SyncMap[*DsnSchema]
    namedPools  syncmap.SyncMap[*DsnSchema]
}

func NewPoolManager(backend serviceapi.SyncConfig) *PoolManager {
    return &PoolManager{
        tenantPools: syncmap.NewSyncMap[*DsnSchema](backend, "tenant"),
        namedPools:  syncmap.NewSyncMap[*DsnSchema](backend, "pool"),
    }
}

func (pm *PoolManager) AddTenantPool(ctx, tenant string, dsn *DsnSchema) error {
    return pm.tenantPools.Set(ctx, tenant, dsn)
}

func (pm *PoolManager) GetTenantPool(ctx, tenant string) (*DsnSchema, bool, error) {
    return pm.tenantPools.Get(ctx, tenant)
}

func (pm *PoolManager) ListTenants(ctx context.Context) ([]string, error) {
    return pm.tenantPools.Keys(ctx)
}
```

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  SyncMap A  │     │  SyncMap B  │     │  SyncMap C  │
│ prefix:     │     │ prefix:     │     │ prefix:     │
│  "users"    │     │  "orders"   │     │  "cache"    │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                    │
       └───────────────────┼────────────────────┘
                           │
                    ┌──────▼──────┐
                    │  SyncConfig │  (Singleton)
                    │   Backend   │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ PostgreSQL  │
                    │   + NOTIFY  │
                    └─────────────┘
```

## Benefits

1. **Singleton Backend**: Multiple SyncMaps share 1 synchronizer
   - 10 servers × 5 SyncMaps = Still only 10 heartbeat queries (not 50)
   
2. **Type Safety**: Compile-time type checking via generics

3. **Automatic Sync**: Changes propagate to all nodes automatically

4. **Clean API**: Map-like interface, familiar to Go developers

5. **Extensible**: Easy to add Redis, etcd, etc. backends in the future

## Performance Notes

- Subscribe callbacks run in goroutines (non-blocking)
- Prefix filtering happens in-memory (fast)
- Single LISTEN connection per backend (efficient)
- CRC heartbeat checks data consistency every 5 minutes

## Future Backends

Coming soon:
- `NewSyncMapRedis()` - Redis Pub/Sub backend
- `NewSyncMapEtcd()` - etcd watch backend
- `NewSyncMapMemory()` - In-memory for testing

## License

Part of the Lokstra framework.
