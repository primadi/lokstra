---
layout: default
title: DbPool Manager - Lokstra Services
description: Centralized database connection pool management with multi-tenancy and named pool support
---

# DbPool Manager

The `dbpool_manager` service provides centralized management of database connection pools with support for multi-tenancy and named pool configurations. It allows dynamic creation and management of multiple database pools based on tenant IDs or custom names.

## Table of Contents

- [Overview](#overview)
- [Configuration](#configuration)
- [Registration](#registration)
- [Core Concepts](#core-concepts)
- [DSN-Based Pool Management](#dsn-based-pool-management)
- [Tenant-Based Pool Management](#tenant-based-pool-management)
- [Named Pool Management](#named-pool-management)
- [Advanced Features](#advanced-features)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

**Service Type:** `dbpool_manager`

**Interface:** `serviceapi.DbPoolManager`

**Key Features:**

```
✓ Dynamic Pool Creation   - Create pools on-demand
✓ Multi-Tenancy Support   - Isolate tenant databases
✓ Named Pool Management   - Custom pool configurations
✓ DSN-Based Pooling       - Share pools across tenants
✓ Thread-Safe Operations  - Concurrent-safe pool access
✓ Automatic Cleanup       - Graceful shutdown handling
```

## Configuration

### YAML Configuration

Lokstra uses a special `dbpool-definitions:` section in the root of YAML config for named pool definitions:

```yaml
# Named database pools section (root level)
dbpool-definitions:
  main-db:
    dsn: "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
    schema: "public"
    min_conns: 2
    max_conns: 10
    max_idle_time: "30m"
    max_lifetime: "1h"
  
  analytics-db:
    host: localhost
    port: 5432
    database: analytics
    username: user
    password: pass
    schema: "analytics"
```

### DbPool Manager Modes

Lokstra provides two implementation types for `dbpool-manager`:

#### 1. Local Map Mode (Default)

Uses in-memory `map[string]serviceapi.DbPool` and `map[string]*serviceapi.DbPoolInfo` to store pools and configurations.

**Implementation:** `services/dbpool_manager/dbpool_manager.go`

**Characteristics:**
- Uses `sync.RWMutex` for thread-safe access
- Pools stored in memory map (keyed by DSN)
- Named pool configs stored in memory map (keyed by name)
- Suitable for single-instance applications

**Configuration:**
```yaml
# Default mode - no special config needed
dbpool-definitions:
  main-db:
    dsn: "postgres://localhost/mydb"
```

**Code:**
```go
import "github.com/primadi/lokstra/services/dbpool_manager"

// Create local pool manager
manager := dbpool_manager.NewPgxPoolManager()
lokstra_registry.RegisterService("dbpool-manager", manager)
```

#### 2. Distributed Sync Mode

Uses `common/syncmap.SyncMap` to store named pool configurations, allowing sharing across multiple instances via PostgreSQL.

**Implementation:** `services/dbpool_manager/sync_pool_manager.go`

**Characteristics:**
- Pools still stored in memory map (keyed by DSN) for performance
- Named pool configs stored in `syncmap.SyncMap[*serviceapi.DbPoolInfo]`
- Configurations sync across instances via PostgreSQL LISTEN/NOTIFY
- Changes persist to database
- Suitable for multi-instance deployments

**Configuration:**
```yaml

# Define pools as usual
dbpool-definitions:
  main-db:
    dsn: "postgres://localhost/mydb"
    schema: "public"
```

**Code:**
```go
import (
    "github.com/primadi/lokstra/services/dbpool_manager"
    "github.com/primadi/lokstra/services/sync_config_pg"
)

// 1. Register sync-config service first (required)
sync_config_pg.Register("dbpool", 5, 10) // name, heartbeat_interval, reconnect_interval

// 2. Create sync pool manager
manager := dbpool_manager.NewPgxSyncDbPoolManager()
lokstra_registry.RegisterService("dbpool-manager", manager)
```

**Differences:**

| Feature | Local Map Mode | Sync Mode |
|---------|---------------|-----------|
| Storage | In-memory `map[string]*DbPoolInfo` | `syncmap.SyncMap[*DbPoolInfo]` |
| Persistence | Lost on restart | Persisted to database |
| Multi-instance | Not shared | Shared across instances |
| Real-time sync | N/A | Via PostgreSQL LISTEN/NOTIFY |
| Initialization | Immediate | Lazy (on first use) |

### Programmatic Configuration

```go
import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/services/dbpool_manager"
    "github.com/primadi/lokstra/lokstra_init"
)

// Option 1: Use helper function (recommended)
lokstra_init.UsePgxDbPoolManager(false) // false = local, true = sync

// Option 2: Manual registration
manager := dbpool_manager.NewPgxPoolManager() // or NewPgxSyncDbPoolManager()
lokstra_registry.RegisterService("dbpool-manager", manager)

// Get manager
manager := lokstra_registry.GetService[serviceapi.DbPoolManager]("dbpool-manager")
```

## Registration

### Basic Registration

```go
import "github.com/primadi/lokstra/services/dbpool_manager"

func init() {
    dbpool_manager.Register()
}
```

### Bulk Registration

```go
import "github.com/primadi/lokstra/services"

func main() {
    // Registers all services including dbpool_manager
    services.RegisterAllServices()
}
```

## Core Concepts

### Pool Management Strategy

The DbPool Manager uses three main strategies for pool management:

1. **DSN-Based Pools** - Shared pools identified by connection string
2. **Tenant-Based Pools** - Tenant-specific database configurations
3. **Named Pools** - Custom-named pool configurations

All three strategies share the same underlying pool instances when DSNs match, ensuring efficient resource usage.

### Interface Overview

```go
type DbPoolManager interface {
    // DSN-based pool management
    GetDsnPool(dsn string) (DbPool, error)
    
    // Tenant-based pool management
    SetTenantDsn(tenant string, dsn string, schema string)
    GetTenantDsn(tenant string) (string, string, error)
    GetTenantPool(tenant string) (DbPool, error)
    RemoveTenant(tenant string)
    AcquireTenantConn(ctx context.Context, tenant string) (DbConn, error)
    
    // Named pool management
    SetNamedDsn(name string, dsn string, schema string)
    GetNamedDsn(name string) (string, string, error)
    GetNamedPool(name string) (DbPool, error)
    RemoveNamed(name string)
    AcquireNamedConn(ctx context.Context, name string) (DbConn, error)
    
    // Shutdown
    Shutdown() error
}
```

## DSN-Based Pool Management

### Get or Create Pool

The `GetDsnPool` method returns an existing pool or creates a new one if it doesn't exist.

```go
import (
    "context"
    "github.com/primadi/lokstra/serviceapi"
)

manager := lokstra_registry.GetService[serviceapi.DbPoolManager]("db-pool-manager")

// Get or create pool for DSN
dsn := "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
pool, err := manager.GetDsnPool(dsn)
if err != nil {
    log.Fatal(err)
}

// Acquire connection from pool
conn, err := pool.Acquire(context.Background(), "public")
if err != nil {
    log.Fatal(err)
}
defer conn.Release()

// Use connection
rows, err := conn.Query(context.Background(), "SELECT * FROM users")
```

### Pool Sharing

Pools with identical DSNs are automatically shared:

```go
// First call creates the pool
pool1, _ := manager.GetDsnPool("postgres://localhost/db1")

// Second call with same DSN returns the same pool instance
pool2, _ := manager.GetDsnPool("postgres://localhost/db1")

// pool1 == pool2 (same instance)
```

## Tenant-Based Pool Management

### Setting Tenant Configuration

Configure database connections for specific tenants:

```go
manager := lokstra_registry.GetService[serviceapi.DbPoolManager]("db-pool-manager")

// Configure tenant database
manager.SetTenantDsn(
    "tenant-123",                                           // Tenant ID
    "postgres://user:pass@localhost:5432/tenant_db",       // DSN
    "tenant_123",                                          // Schema name
)

manager.SetTenantDsn(
    "tenant-456",
    "postgres://user:pass@localhost:5432/tenant_db",
    "tenant_456",
)
```

### Getting Tenant Pool

```go
// Get pool for specific tenant
pool, err := manager.GetTenantPool("tenant-123")
if err != nil {
    log.Fatal(err)
}

// Acquire connection
conn, err := pool.Acquire(context.Background(), "tenant_123")
if err != nil {
    log.Fatal(err)
}
defer conn.Release()
```

### Acquiring Tenant Connection (Recommended)

The `AcquireTenantConn` method is the recommended way to get tenant connections:

```go
// Acquire connection with tenant context automatically set
conn, err := manager.AcquireTenantConn(context.Background(), "tenant-123")
if err != nil {
    log.Fatal(err)
}
defer conn.Release()

// Connection has:
// - Correct schema set
// - Tenant context set for RLS
// - Ready to use immediately
rows, err := conn.Query(context.Background(), "SELECT * FROM users")
```

### Getting Tenant Configuration

```go
// Retrieve tenant DSN and schema
dsn, schema, err := manager.GetTenantDsn("tenant-123")
if err != nil {
    log.Printf("Tenant not configured: %v", err)
    return
}

log.Printf("Tenant DSN: %s, Schema: %s", dsn, schema)
```

### Removing Tenant

```go
// Remove tenant configuration
manager.RemoveTenant("tenant-123")

// Pool is not removed (might be used by other tenants)
// Only the tenant->DSN mapping is removed
```

## Named Pool Management

### Setting Named Configuration

Named pools allow custom configurations with meaningful names:

```go
manager := lokstra_registry.GetService[serviceapi.DbPoolManager]("db-pool-manager")

// Configure analytics database
manager.SetNamedDsn(
    "analytics",                                     // Pool name
    "postgres://user:pass@analytics-db:5432/stats", // DSN
    "public",                                        // Schema
)

// Configure reporting database
manager.SetNamedDsn(
    "reporting",
    "postgres://user:pass@reporting-db:5432/reports",
    "public",
)

// Configure read-replica
manager.SetNamedDsn(
    "read-replica",
    "postgres://user:pass@replica:5432/mydb",
    "public",
)
```

### Getting Named Pool

```go
// Get pool by name
pool, err := manager.GetNamedPool("analytics")
if err != nil {
    log.Fatal(err)
}

// Acquire connection
conn, err := pool.Acquire(context.Background(), "public")
if err != nil {
    log.Fatal(err)
}
defer conn.Release()
```

### Acquiring Named Connection (Recommended)

```go
// Acquire connection with schema automatically set
conn, err := manager.AcquireNamedConn(context.Background(), "analytics")
if err != nil {
    log.Fatal(err)
}
defer conn.Release()

// Use for analytics queries
stats, err := conn.SelectManyRowMap(context.Background(),
    "SELECT date, count, revenue FROM daily_stats WHERE date >= $1",
    startDate,
)
```

### Getting Named Configuration

```go
// Retrieve named pool configuration
dsn, schema, err := manager.GetNamedDsn("analytics")
if err != nil {
    log.Printf("Named pool not found: %v", err)
    return
}

log.Printf("Analytics DSN: %s, Schema: %s", dsn, schema)
```

### Removing Named Pool

```go
// Remove named pool configuration
manager.RemoveNamed("analytics")

// Pool is not removed (might be used by other names/tenants)
// Only the name->DSN mapping is removed
```

## Transaction Management

Lokstra provides lazy transaction management through context. Transactions are created on-demand when the first database operation occurs, and all subsequent operations in the same context automatically join the transaction.

### BeginTransaction

The `BeginTransaction` method marks a context as needing a transaction for a specific pool name. The transaction is created lazily on first database operation.

**Package:** `serviceapi`

**Signature:**
```go
func BeginTransaction(ctx context.Context, poolName string) (context.Context, func(*error))
```

**Parameters:**
- `ctx`: The context to attach transaction to
- `poolName`: Name of the database pool (e.g., `"main-db"`, `"analytics-db"`)

**Returns:**
- New context with transaction marker
- Finalize function that should be deferred, accepts pointer to error for commit/rollback decision

**Usage Pattern:**
```go
func (s *Service) DoWork(ctx context.Context) (err error) {
    ctx, finish := serviceapi.BeginTransaction(ctx, "main-db")
    defer finish(&err) // Auto commit if err == nil, rollback if err != nil
    
    // First DB operation creates the transaction
    s.repo1.Create(ctx, data1)
    
    // Subsequent operations join the same transaction
    s.repo2.Update(ctx, data2)
    s.repo3.Delete(ctx, id)
    
    return nil // Transaction commits
}
```

### Request Context Integration

In HTTP handlers, use `request.Context.BeginTransaction()`:

```go
func (s *UserService) CreateUser(ctx *request.Context, user *User) (err error) {
    defer ctx.BeginTransaction("main-db")(&err)
    
    // All operations join the transaction
    if err := s.userRepo.Create(ctx, user); err != nil {
        return err // Auto rollback
    }
    
    if err := s.auditRepo.Log(ctx, "user_created", user.ID); err != nil {
        return err // Auto rollback
    }
    
    return nil // Auto commit
}
```

### Transaction Context Details

**TxContext Structure:**
```go
type TxContext struct {
    PoolName   string              // Pool name (e.g., "db_auth")
    Tx         DbTx                // Transaction instance (lazy created)
    Conn       DbConn              // Connection instance
    Counter    int                 // Nested call counter
    committed  bool                // Commit flag
    rolledBack bool                // Rollback flag
}
```

**How It Works:**

1. **Marking Phase**: `BeginTransaction` adds `TxContext` marker to context
2. **Lazy Creation**: First DB operation checks context, finds marker, creates transaction
3. **Auto-Join**: Subsequent DB operations detect existing transaction and reuse it
4. **Finalization**: Deferred function commits (on success) or rolls back (on error)

**Transaction Flow:**
```
BeginTransaction("main-db")
    ↓
Context with TxContext marker (Tx == nil)
    ↓
First DB Operation
    ↓
getExecutor() checks context → finds TxContext
    ↓
Lazy create transaction (Tx = Begin())
    ↓
Second DB Operation
    ↓
getExecutor() checks context → reuses existing Tx
    ↓
defer finish(&err) → Commit() or Rollback()
```

### Pool Name Based Tracking

Transactions are tracked by pool name, not pool instance. This allows:
- Multiple pools to have independent transactions
- Services to reference pools by name without direct injection
- Transaction context to survive service boundaries

**Example - Multiple Pool Transactions:**
```go
func (s *Service) Transfer(ctx context.Context) (err error) {
    // Transaction for main database
    ctx, finish1 := serviceapi.BeginTransaction(ctx, "main-db")
    defer finish1(&err)
    
    // Transaction for analytics database (separate, independent)
    ctx, finish2 := serviceapi.BeginTransaction(ctx, "analytics-db")
    defer finish2(&err)
    
    // Operations on main-db join main-db transaction
    s.mainRepo.Update(ctx, ...) // Uses "main-db" transaction
    
    // Operations on analytics-db join analytics-db transaction
    s.analyticsRepo.Log(ctx, ...) // Uses "analytics-db" transaction
    
    return nil
}
```

### WithoutTransaction

Use `WithoutTransaction` to explicitly ignore parent transaction:

```go
import "github.com/primadi/lokstra/serviceapi"

func (s *Service) CreateWithAudit(ctx context.Context) (err error) {
    ctx, finish := serviceapi.BeginTransaction(ctx, "main-db")
    defer finish(&err)
    
    // This joins the transaction
    s.repo.Create(ctx, data)
    
    // This uses a separate connection (no transaction)
    isolatedCtx := serviceapi.WithoutTransaction(ctx)
    s.auditRepo.Log(isolatedCtx, "created") // Commits immediately
    
    return nil
}
```

### Nested Transaction Support

Lokstra supports pseudo-nested transactions using a counter mechanism:

```go
func (s *Service) Outer(ctx context.Context) (err error) {
    ctx, finish := serviceapi.BeginTransaction(ctx, "main-db")
    defer finish(&err) // Counter = 1
    
    s.Inner(ctx) // Calls BeginTransaction again
    
    return nil
}

func (s *Service) Inner(ctx context.Context) (err error) {
    // Counter increments to 2
    ctx, finish := serviceapi.BeginTransaction(ctx, "main-db")
    defer finish(&err) // Counter decrements to 1
    
    s.repo.Update(ctx, data)
    
    return nil // Transaction commits only when counter reaches 0
}
```

## Advanced Features

### Custom Pool Factory

You can create a pool manager with a custom pool factory function:

```go
import "github.com/primadi/lokstra/services/dbpool_manager"

// Custom factory function
customFactory := func(dsn string) (serviceapi.DbPool, error) {
    // Custom pool creation logic
    log.Printf("Creating pool for: %s", dsn)
    
    // Use custom configuration
    cfg := &CustomPoolConfig{
        DSN:            dsn,
        MaxConnections: 50,
        MinConnections: 5,
    }
    
    return NewCustomPool(cfg)
}

// Create manager with custom factory
manager := dbpool_manager.NewPoolManager(customFactory)
```

### Pool Reuse Across Strategies

Pools are automatically shared when DSNs match:

```go
// Tenant and named pools share the same DSN
manager.SetTenantDsn("tenant-1", "postgres://localhost/db", "schema1")
manager.SetNamedDsn("main", "postgres://localhost/db", "schema2")

// Both use the same underlying pool
tenantPool, _ := manager.GetTenantPool("tenant-1")
namedPool, _ := manager.GetNamedPool("main")

// tenantPool and namedPool share the same connection pool
// Only the schema differs when acquiring connections
```

### Graceful Shutdown

The pool manager implements graceful shutdown:

```go
// Shutdown all managed pools
if err := manager.Shutdown(); err != nil {
    log.Printf("Error during shutdown: %v", err)
}

// All pools are closed
// All active connections are released
```

### Thread-Safe Operations

All operations are thread-safe using `sync.Map`:

```go
// Safe to call from multiple goroutines
go manager.SetTenantDsn("tenant-1", dsn1, "schema1")
go manager.SetTenantDsn("tenant-2", dsn2, "schema2")
go manager.AcquireTenantConn(ctx, "tenant-1")
go manager.AcquireTenantConn(ctx, "tenant-2")
```

## Best Practices

### Pool Management

```go
✓ DO: Use tenant-based pools for multi-tenant applications
manager.SetTenantDsn("tenant-id", dsn, schema)
conn, _ := manager.AcquireTenantConn(ctx, "tenant-id")

✓ DO: Use named pools for different database purposes
manager.SetNamedDsn("analytics", analyticsDsn, "public")
manager.SetNamedDsn("cache", cacheDsn, "public")

✓ DO: Share pools with identical DSNs
// Same DSN = same pool = efficient resource usage
manager.SetTenantDsn("tenant-1", dsn, "schema1")
manager.SetTenantDsn("tenant-2", dsn, "schema2")

✗ DON'T: Create unnecessary pools
// BAD: Different names for same database
manager.SetNamedDsn("pool1", dsn, "public")
manager.SetNamedDsn("pool2", dsn, "public")
// Instead, use the same name or rely on DSN-based sharing
```

### Connection Acquisition

```go
✓ DO: Use AcquireTenantConn for tenant connections
conn, _ := manager.AcquireTenantConn(ctx, "tenant-id")
// Schema and tenant context automatically set

✓ DO: Use AcquireNamedConn for named pools
conn, _ := manager.AcquireNamedConn(ctx, "analytics")
// Schema automatically set

✓ DO: Always release connections
conn, _ := manager.AcquireTenantConn(ctx, "tenant-id")
defer conn.Release()

✗ DON'T: Manually manage schema/tenant context
// BAD: Manual management is error-prone
pool, _ := manager.GetTenantPool("tenant-id")
conn, _ := pool.Acquire(ctx, "wrong_schema") // Easy to make mistakes
```

### Configuration Management

```go
✓ DO: Configure tenants at application startup
func initializeTenants() {
    manager := lokstra_registry.GetService[serviceapi.DbPoolManager]("db-pool-manager")
    
    tenants := []Tenant{
        {ID: "tenant-1", DSN: dsn1, Schema: "schema1"},
        {ID: "tenant-2", DSN: dsn2, Schema: "schema2"},
    }
    
    for _, t := range tenants {
        manager.SetTenantDsn(t.ID, t.DSN, t.Schema)
    }
}

✓ DO: Configure named pools for different purposes
manager.SetNamedDsn("main", mainDsn, "public")
manager.SetNamedDsn("analytics", analyticsDsn, "public")
manager.SetNamedDsn("read-replica", replicaDsn, "public")

✓ DO: Remove tenant configurations when tenant is deleted
manager.RemoveTenant("deleted-tenant-id")

✗ DON'T: Configure tenants dynamically for every request
// BAD: Performance overhead
func handler(tenantID string) {
    manager.SetTenantDsn(tenantID, dsn, schema) // Repeated configuration
    conn, _ := manager.AcquireTenantConn(ctx, tenantID)
}
// Instead, configure once at startup or on tenant creation
```

### Error Handling

```go
✓ DO: Check for configuration errors
conn, err := manager.AcquireTenantConn(ctx, tenantID)
if err != nil {
    if err.Error() == "tenant pool not found: "+tenantID {
        return ErrTenantNotConfigured
    }
    return err
}

✓ DO: Validate tenant existence before use
dsn, schema, err := manager.GetTenantDsn(tenantID)
if err != nil {
    return fmt.Errorf("tenant %s not found", tenantID)
}

✗ DON'T: Ignore configuration errors
conn, _ := manager.AcquireTenantConn(ctx, tenantID) // BAD: Ignoring errors
```

## Examples

### Multi-Tenant Application

```go
package main

import (
    "context"
    "fmt"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

type TenantService struct {
    poolManager serviceapi.DbPoolManager
}

func NewTenantService() *TenantService {
    return &TenantService{
        poolManager: lokstra_registry.GetService[serviceapi.DbPoolManager]("db-pool-manager"),
    }
}

// Initialize tenant database configuration
func (s *TenantService) AddTenant(tenantID, dsn, schema string) error {
    s.poolManager.SetTenantDsn(tenantID, dsn, schema)
    
    // Test connection
    conn, err := s.poolManager.AcquireTenantConn(context.Background(), tenantID)
    if err != nil {
        s.poolManager.RemoveTenant(tenantID)
        return fmt.Errorf("failed to connect tenant database: %w", err)
    }
    defer conn.Release()
    
    if err := conn.Ping(context.Background()); err != nil {
        s.poolManager.RemoveTenant(tenantID)
        return fmt.Errorf("tenant database ping failed: %w", err)
    }
    
    return nil
}

// Get users for specific tenant
func (s *TenantService) GetTenantUsers(ctx context.Context, tenantID string) ([]User, error) {
    // Acquire tenant-specific connection
    conn, err := s.poolManager.AcquireTenantConn(ctx, tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to acquire tenant connection: %w", err)
    }
    defer conn.Release()
    
    // Query with automatic RLS filtering
    mapper := func(row serviceapi.Row) (any, error) {
        var user User
        err := row.Scan(&user.ID, &user.Name, &user.Email)
        return user, err
    }
    
    result, err := conn.SelectManyWithMapper(ctx, mapper,
        "SELECT id, name, email FROM users WHERE active = true",
    )
    
    if err != nil {
        return nil, err
    }
    
    return result.([]User), nil
}

// Remove tenant
func (s *TenantService) RemoveTenant(tenantID string) {
    s.poolManager.RemoveTenant(tenantID)
}
```

### Named Pool Usage

```go
package repository

import (
    "context"
    "github.com/primadi/lokstra/serviceapi"
)

type AnalyticsRepository struct {
    poolManager serviceapi.DbPoolManager
}

func NewAnalyticsRepository(manager serviceapi.DbPoolManager) *AnalyticsRepository {
    return &AnalyticsRepository{
        poolManager: manager,
    }
}

// Initialize analytics database
func (r *AnalyticsRepository) Initialize() error {
    // Configure analytics pool
    r.poolManager.SetNamedDsn(
        "analytics",
        "postgres://user:pass@analytics-db:5432/stats",
        "public",
    )
    
    // Test connection
    conn, err := r.poolManager.AcquireNamedConn(context.Background(), "analytics")
    if err != nil {
        return err
    }
    defer conn.Release()
    
    return conn.Ping(context.Background())
}

// Get daily statistics
func (r *AnalyticsRepository) GetDailyStats(ctx context.Context, date string) ([]Stat, error) {
    conn, err := r.poolManager.AcquireNamedConn(ctx, "analytics")
    if err != nil {
        return nil, err
    }
    defer conn.Release()
    
    rows, err := conn.SelectManyRowMap(ctx,
        `SELECT metric, value, timestamp 
         FROM daily_stats 
         WHERE date = $1 
         ORDER BY timestamp`,
        date,
    )
    
    if err != nil {
        return nil, err
    }
    
    stats := make([]Stat, len(rows))
    for i, row := range rows {
        stats[i] = Stat{
            Metric:    row["metric"].(string),
            Value:     row["value"].(float64),
            Timestamp: row["timestamp"].(time.Time),
        }
    }
    
    return stats, nil
}
```

### Mixed Strategy Application

```go
package main

import (
    "context"
    "github.com/primadi/lokstra/serviceapi"
)

type DatabaseService struct {
    poolManager serviceapi.DbPoolManager
}

func (s *DatabaseService) Initialize() error {
    // Configure main application database
    s.poolManager.SetNamedDsn(
        "main",
        "postgres://user:pass@localhost:5432/app_db",
        "public",
    )
    
    // Configure read replica
    s.poolManager.SetNamedDsn(
        "read-replica",
        "postgres://user:pass@replica:5432/app_db",
        "public",
    )
    
    // Configure analytics database
    s.poolManager.SetNamedDsn(
        "analytics",
        "postgres://user:pass@analytics:5432/stats_db",
        "public",
    )
    
    // Configure multi-tenant databases
    tenants := []struct {
        ID     string
        DSN    string
        Schema string
    }{
        {"tenant-1", "postgres://localhost:5432/tenant_db", "tenant_1"},
        {"tenant-2", "postgres://localhost:5432/tenant_db", "tenant_2"},
        {"tenant-3", "postgres://localhost:5432/tenant_db", "tenant_3"},
    }
    
    for _, t := range tenants {
        s.poolManager.SetTenantDsn(t.ID, t.DSN, t.Schema)
    }
    
    return nil
}

// Write to main database
func (s *DatabaseService) CreateUser(ctx context.Context, user *User) error {
    conn, err := s.poolManager.AcquireNamedConn(ctx, "main")
    if err != nil {
        return err
    }
    defer conn.Release()
    
    return conn.QueryRow(ctx,
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        user.Name, user.Email,
    ).Scan(&user.ID)
}

// Read from replica
func (s *DatabaseService) GetUsers(ctx context.Context) ([]User, error) {
    conn, err := s.poolManager.AcquireNamedConn(ctx, "read-replica")
    if err != nil {
        return nil, err
    }
    defer conn.Release()
    
    rows, err := conn.SelectManyRowMap(ctx, "SELECT id, name, email FROM users")
    if err != nil {
        return nil, err
    }
    
    users := make([]User, len(rows))
    for i, row := range rows {
        users[i] = User{
            ID:    row["id"].(int),
            Name:  row["name"].(string),
            Email: row["email"].(string),
        }
    }
    
    return users, nil
}

// Write analytics data
func (s *DatabaseService) LogEvent(ctx context.Context, event *Event) error {
    conn, err := s.poolManager.AcquireNamedConn(ctx, "analytics")
    if err != nil {
        return err
    }
    defer conn.Release()
    
    _, err = conn.Exec(ctx,
        "INSERT INTO events (type, data, timestamp) VALUES ($1, $2, $3)",
        event.Type, event.Data, event.Timestamp,
    )
    
    return err
}

// Tenant-specific operation
func (s *DatabaseService) GetTenantData(ctx context.Context, tenantID string) ([]Data, error) {
    conn, err := s.poolManager.AcquireTenantConn(ctx, tenantID)
    if err != nil {
        return nil, err
    }
    defer conn.Release()
    
    rows, err := conn.SelectManyRowMap(ctx, "SELECT * FROM tenant_data")
    if err != nil {
        return nil, err
    }
    
    // Process rows...
    return processRows(rows), nil
}
```

### Dynamic Tenant Onboarding

```go
package service

import (
    "context"
    "fmt"
)

type TenantOnboardingService struct {
    poolManager serviceapi.DbPoolManager
}

// Onboard new tenant
func (s *TenantOnboardingService) OnboardTenant(ctx context.Context, tenant *Tenant) error {
    // Configure tenant database
    s.poolManager.SetTenantDsn(tenant.ID, tenant.DSN, tenant.Schema)
    
    // Acquire connection to verify
    conn, err := s.poolManager.AcquireTenantConn(ctx, tenant.ID)
    if err != nil {
        s.poolManager.RemoveTenant(tenant.ID)
        return fmt.Errorf("failed to connect: %w", err)
    }
    defer conn.Release()
    
    // Run migrations/setup
    if err := s.runTenantSetup(ctx, conn, tenant); err != nil {
        s.poolManager.RemoveTenant(tenant.ID)
        return fmt.Errorf("setup failed: %w", err)
    }
    
    return nil
}

// Offboard tenant
func (s *TenantOnboardingService) OffboardTenant(ctx context.Context, tenantID string) error {
    // Remove tenant configuration
    s.poolManager.RemoveTenant(tenantID)
    
    // Note: Underlying pool is not removed if other tenants share the same DSN
    // This is the desired behavior for resource efficiency
    
    return nil
}

// Migrate tenant to new database
func (s *TenantOnboardingService) MigrateTenant(ctx context.Context, 
    tenantID, newDSN, newSchema string) error {
    
    // Get old configuration
    oldDSN, oldSchema, err := s.poolManager.GetTenantDsn(tenantID)
    if err != nil {
        return err
    }
    
    // Update to new configuration
    s.poolManager.SetTenantDsn(tenantID, newDSN, newSchema)
    
    // Verify new connection
    conn, err := s.poolManager.AcquireTenantConn(ctx, tenantID)
    if err != nil {
        // Rollback to old configuration
        s.poolManager.SetTenantDsn(tenantID, oldDSN, oldSchema)
        return fmt.Errorf("failed to connect to new database: %w", err)
    }
    defer conn.Release()
    
    if err := conn.Ping(ctx); err != nil {
        // Rollback to old configuration
        s.poolManager.SetTenantDsn(tenantID, oldDSN, oldSchema)
        return fmt.Errorf("new database ping failed: %w", err)
    }
    
    return nil
}

func (s *TenantOnboardingService) runTenantSetup(ctx context.Context, 
    conn serviceapi.DbConn, tenant *Tenant) error {
    
    // Create tables, indexes, etc.
    migrations := []string{
        "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)",
        "CREATE TABLE IF NOT EXISTS orders (id SERIAL PRIMARY KEY, user_id INT, amount DECIMAL)",
        "CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
    }
    
    for _, migration := range migrations {
        if _, err := conn.Exec(ctx, migration); err != nil {
            return fmt.Errorf("migration failed: %w", err)
        }
    }
    
    return nil
}
```

## Related Documentation

- [DbPool (PostgreSQL)](dbpool-pg) - PostgreSQL connection pooling
- [Services Overview](index) - Service architecture and patterns
- [Configuration](../03-configuration/config) - YAML configuration system
- [Multi-Tenancy Guide](../../02-framework-guide/multi-tenancy) - Multi-tenant application patterns

---

**Next:** [Metrics Service](metrics-prometheus) - Prometheus metrics integration
