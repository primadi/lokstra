---
title: Database Pools
layout: default
parent: Framework Guide
nav_order: 8
---

# Database Pools

Lokstra provides built-in support for database connection pooling with automatic configuration from YAML files. Understanding the hierarchy of database abstractions helps you use them effectively.

## Database Components Hierarchy

Lokstra's database system has four main components:

```
DbPoolManager (service)
    ↓ manages
DbPool (connection pool, injectable service)
    ↓ provides
DbConn (individual connection)
    ↓ can create
DbTx (transaction)
```

### 1. DbPoolManager

**Service** that manages multiple named database pools.

- Manages pool configurations (DSN, schema, etc.)
- Creates and caches pool instances
- Provides access to pools by name
- Supports local map or distributed sync storage

### 2. DbPool

**Interface** representing a connection pool. Can be injected into services.

```go
type DbPool interface {
    Acquire(ctx context.Context) (DbConn, error)
    DbConn // Can also execute queries directly
}
```

- Provides connections from the pool
- Can be injected as `@Inject "pool-name"`
- Shared across services using the same pool name

### 3. DbConn

**Interface** representing an individual database connection.

```go
type DbConn interface {
    Begin(ctx context.Context) (DbTx, error)
    Transaction(ctx context.Context, fn func(tx DbExecutor) error) error
    Release() error
    DbExecutor // Can execute queries
}
```

- Represents a single connection from the pool
- Must be released when done
- Can create transactions

### 4. DbTx (Transaction)

**Interface** representing a database transaction.

```go
type DbTx interface {
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
    DbExecutor // Can execute queries
}
```

- Represents an ongoing transaction
- Created from DbConn or via context (recommended)
- Must be committed or rolled back

## Transaction via Context (Recommended)

The recommended way to handle transactions is using `ctx.BeginTransaction(poolName)` from `request.Context`. Transactions are automatically finalized (commit/rollback) when the response is written.

## Setup Database Pools

### 1. Define DB Pools in Config

Lokstra has a special `dbpool-definitions:` section in YAML config for defining named database pools:

**config.yaml:**
```yaml
# Named database pools configuration
dbpool-definitions:
  main-db:
    dsn: "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
    min_conns: 2
    max_conns: 10
    max_idle_time: "30m"
    max_lifetime: "1h"
    schema: "public"

  analytics-db:
    host: localhost
    port: 5432
    database: analytics
    username: analytics_user
    password: secret
    sslmode: disable
    min_conns: 2
    max_conns: 20
    schema: "analytics"
```

This section is automatically loaded and pools are registered as services that can be injected.

### 2. Recommended: Use lokstra_init

**The recommended way** is to use `lokstra_init.BootstrapAndRun()` which handles everything in the correct order:

```go
package main

import "github.com/primadi/lokstra/lokstra_init"

func main() {
    // Handles all initialization including dbpool-definitions setup
    if err := lokstra_init.BootstrapAndRun(); err != nil {
        log.Fatal(err)
    }
}
```

With options for sync mode:
```go
err := lokstra_init.BootstrapAndRun(
    lokstra_init.WithDbPoolManager(true, true), // enable, useSync
    lokstra_init.WithPgSyncMap(true, "db_main"),
)
```

**See [Lokstra Initialization](./09-lokstra-init.md) for details.**

### 3. Manual Setup (Advanced)

If you need more control, you can set up manually (not recommended unless you understand the initialization order):

```go
func main() {
    lokstra.Bootstrap()
    
    // 1. Load config
    if err := lokstra_registry.LoadConfig("config.yaml"); err != nil {
        log.Fatal(err)
    }
    
    // 2. Setup sync-config first (if using sync mode)
    sync_config_pg.Register("db_main", 5*time.Minute, 5*time.Second)
    
    // 3. Setup definitions
    lokstra_init.UsePgxDbPoolManager(true) // true = sync mode
    
    // 4. Load pools from config
    if err := loader.LoadDbPoolManagerFromConfig(); err != nil {
        log.Fatal(err)
    }
    
    // 5. Run server
    lokstra_registry.InitAndRunServer()
}
```

## Inject DB Pool into Service

### Using @Inject Annotation

```go
// @Service "user-repository"
type UserRepository struct {
    // @Inject "main-db"
    DB serviceapi.DbPool
}

func (r *UserRepository) GetUser(id string) (*User, error) {
    var user User
    err := r.DB.QueryRow(context.Background(), 
        "SELECT id, name, email FROM users WHERE id = $1", id,
    ).Scan(&user.ID, &user.Name, &user.Email)
    return &user, err
}
```

### Using Manual Injection

```go
func UserRepositoryFactory(deps map[string]any, config map[string]any) any {
    return &UserRepository{
        DB: deps["main-db"].(serviceapi.DbPool),
    }
}

// In register.go
lokstra_registry.RegisterServiceType("user-repository-factory", 
    UserRepositoryFactory, nil)
```

**config.yaml:**
```yaml
service-definitions:
  user-repository:
    type: user-repository-factory
    depends-on:
      - DB:main-db  # Inject DB pool named "main-db"
```

## DSN Configuration

### Option 1: Direct DSN

```yaml
dbpool-definitions:
  mydb:
    dsn: "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
```

### Option 2: Component-Based (Recommended)

```yaml
dbpool-definitions:
  mydb:
    host: ${DB_HOST:localhost}
    port: ${DB_PORT:5432}
    database: ${DB_NAME:mydb}
    username: ${DB_USER:user}
    password: ${DB_PASS:secret}
    sslmode: ${DB_SSLMODE:disable}
```

## Pool Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `min_conns` | 2 | Minimum connections in pool |
| `max_conns` | 10 | Maximum connections in pool |
| `max_idle_time` | 30m | Max time a connection can be idle |
| `max_lifetime` | 1h | Max lifetime of a connection |
| `schema` | public | Default PostgreSQL schema |

## Best Practices

### 1. Separate Config from Code

✅ **Good:**
```go
// Load config first, setup DB later
lokstra_registry.LoadConfig("config.yaml")
lokstra.SetupDbPoolManager()
```

❌ **Bad:**
```go
// Auto-setup couples config loading with infrastructure
lokstra_registry.RunServerFromConfig("config.yaml")
```

### 2. Use Named Pools for Different Purposes

```yaml
dbpool-definitions:
  transactional-db:  # For OLTP workloads
    max_conns: 10
    
  analytics-db:      # For OLAP workloads
    max_conns: 50
    
  cache-db:          # For caching
    max_conns: 5
```

### 3. Environment-Specific Configuration

```yaml
dbpool-definitions:
  main-db:
    host: ${DB_HOST:localhost}
    port: ${DB_PORT:5432}
    database: ${DB_NAME}          # Required in production
    username: ${DB_USER}          # Required in production
    password: ${DB_PASS}          # Required in production
    sslmode: ${DB_SSLMODE:require}
```

**Development:**
```bash
export DB_NAME=myapp_dev
export DB_USER=dev_user
export DB_PASS=dev_pass
export DB_SSLMODE=disable
```

**Production:**
```bash
export DB_NAME=myapp_prod
export DB_USER=prod_user
export DB_PASS=secure_password
export DB_SSLMODE=require
```

## Testing Without DB

```go
func TestUserService(t *testing.T) {
    // Load config without setting up DB pools
    lokstra_registry.LoadConfig("config.yaml")
    
    // Mock DB pool
    mockDB := &MockDbPool{}
    lokstra_registry.RegisterService("main-db", mockDB)
    
    // Test service
    service := lokstra_registry.GetService[*UserService]("user-service")
    // ...
}
```

## Multiple Databases

```go
// @Service "reporting-service"
type ReportingService struct {
    // @Inject "transactional-db"
    TransactionalDB serviceapi.DbPool
    
    // @Inject "analytics-db"
    AnalyticsDB serviceapi.DbPool
}

func (s *ReportingService) GenerateReport() (*Report, error) {
    // Read from transactional DB
    users, _ := s.TransactionalDB.Query(...)
    
    // Read from analytics DB
    metrics, _ := s.AnalyticsDB.Query(...)
    
    return &Report{Users: users, Metrics: metrics}, nil
}
```

## DbPool Manager Modes

Lokstra supports two types of `dbpool-manager` implementations:

### 1. Local Map Mode (Default)

Uses in-memory map to store pool configurations. Suitable for single-instance applications.

```yaml
# No special config needed - this is the default
dbpool-definitions:
  main-db:
    dsn: "postgres://localhost/mydb"
```

**Characteristics:**
- ✅ Fast access (in-memory map)
- ✅ Simple configuration
- ❌ Pool configs not shared across instances
- ❌ Changes lost on restart

### 2. Distributed Sync Mode

Uses PostgreSQL-based SyncMap to share pool configurations across multiple instances. Requires `sync_config_pg` service.

```yaml
# Configure in configs section
configs:
  dbpool-definitions:
    use_sync: true

# Then define pools as usual
dbpool-definitions:
  main-db:
    dsn: "postgres://localhost/mydb"
    schema: "public"
```

**Characteristics:**
- ✅ Pool configs shared across all instances
- ✅ Changes persist and sync in real-time
- ✅ Suitable for multi-instance deployments
- ⚠️ Requires `sync_config_pg` service to be registered

**When to use:**
- Multiple application instances running
- Need dynamic pool management across instances
- Want pool configurations to persist and sync

**Setup for Sync Mode:**

**Recommended:** Use `lokstra_init`:

```go
import "github.com/primadi/lokstra/lokstra_init"

func main() {
    err := lokstra_init.BootstrapAndRun(
        lokstra_init.WithDbPoolManager(true, true), // enable, useSync=true
        lokstra_init.WithPgSyncMap(true, "db_main"),
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

**Manual setup** (advanced, not recommended):
```go
// 1. Register sync-config first (required for sync mode)
sync_config_pg.Register("db_main", 5*time.Minute, 5*time.Second)

// 2. Load config (contains use_sync: true)
lokstra_registry.LoadConfig("config.yaml")

// 3. Setup dbpool-manager with sync mode
lokstra_init.UsePgxDbPoolManager(true)

// 4. Load pools from config
loader.LoadDbPoolManagerFromConfig()

// 5. Run server
lokstra_registry.InitAndRunServer()
```

**See [Lokstra Initialization](./09-lokstra-init.md) for recommended approach.**

## Transaction Management

**Recommended approach:** Use `ctx.BeginTransaction(poolName)` from `request.Context`. This creates a transaction for the specified pool name lazily (only when first database operation occurs) and automatically commits/rolls back based on error state.

### Basic Transaction Usage

```go
func (s *UserService) CreateUser(ctx *request.Context, user *User) error {
    // Begin transaction for pool named "main-db"
    // Transaction will be created automatically on first DB operation
    // Will auto-commit on success or rollback on error in FinalizeResponse
    ctx.BeginTransaction("main-db")
    
    // All database operations using this ctx will join the same transaction
    if err := s.userRepo.Create(ctx, user); err != nil {
        return err // Auto rollback on error
    }
    
    if err := s.auditRepo.Log(ctx, "user_created", user.ID); err != nil {
        return err // Auto rollback on error
    }
    
    return nil // Auto commit on success
}
```

**Example from real code:**

```go
// @Route "POST /"
func (s *TenantService) CreateTenant(ctx *request.Context,
    req *domain.CreateTenantRequest) (*domain.Tenant, error) {

    // Begin transaction for pool "db_auth"
    // All subsequent DB operations using ctx will join this transaction
    ctx.BeginTransaction("db_auth")

    // These operations will automatically join the transaction
    existing, err := s.TenantStore.GetByName(ctx, req.Name)
    if err == nil && existing != nil {
        return nil, fmt.Errorf("tenant already exists")
    }
    
    // Create tenant (also joins transaction)
    tenant, err := s.TenantStore.Create(ctx, ...)
    
    return tenant, err // Auto commit if nil, rollback if error
}
```

### How It Works

1. **Marking**: `BeginTransaction("pool-name")` marks the context with transaction intent
2. **Lazy Creation**: Transaction is created only when first DB operation occurs (not immediately)
3. **Pool Name Based**: Transaction is tracked by pool name, not pool instance
4. **Auto-Join**: All DB operations using the same context automatically join the transaction
5. **Auto-Finalization**: Happens automatically in `FinalizeResponse()`:
   - Returns `nil` + status < 400 → **Commit**
   - Returns error OR status >= 400 → **Rollback**

**Key Points:**
- No need to manually inject DbPool - just use the pool name
- Transaction is created lazily (zero overhead if no DB operations)
- All operations in the same context automatically share the transaction
- Rollback happens on **any** error status (400+), even if handler returns nil error

**Example - Status-based rollback:**
```go
func (s *Service) Create(ctx *request.Context, req *Request) error {
    ctx.BeginTransaction("db")
    
    s.repo.Create(ctx, data)
    
    // Even though error is nil, transaction will rollback because status = 400
    return ctx.Api.BadRequest("Validation failed") // ← Triggers rollback!
}
```

### Manual Transaction Control

For advanced scenarios (dry-run, testing, conditional commit):

```go
// Dry-run: Execute operations but don't persist
func (s *Service) DryRun(ctx *request.Context, req *Request) error {
    ctx.BeginTransaction("main-db")
    
    // Execute all operations
    result, err := s.repo.Create(ctx, req)
    if err != nil {
        return err
    }
    
    // Manual rollback - changes discarded
    ctx.RollbackTransaction("main-db")
    
    // Return 200 OK with results
    return ctx.Api.Ok(map[string]any{
        "message": "Dry run successful",
        "result": result,
    })
}

// Conditional commit
func (s *Service) BatchProcess(ctx *request.Context, items []Item) error {
    ctx.BeginTransaction("main-db")
    
    successCount := s.processItems(ctx, items)
    
    if successCount < len(items) * 0.8 {
        ctx.RollbackTransaction("main-db") // Below threshold
        return ctx.Api.Ok(map[string]any{"status": "rolled_back"})
    }
    
    ctx.CommitTransaction("main-db") // Above threshold
    return ctx.Api.Ok(map[string]any{"status": "committed"})
}
```

**Available Methods:**
- `ctx.RollbackTransaction(poolName)` - Force rollback
- `ctx.CommitTransaction(poolName)` - Force commit

**See:** [Manual Transaction Control Examples](../examples/manual-transaction-control.md)

### Using Without Request Context (Service Layer)

If you need transactions outside of HTTP handlers, use `serviceapi.BeginTransaction()`:

```go
import "github.com/primadi/lokstra/serviceapi"

func (s *UserService) DoWork(ctx context.Context) (err error) {
    // Begin transaction (same API, but for standard context.Context)
    ctx, finish := serviceapi.BeginTransaction(ctx, "main-db")
    defer finish(&err)
    
    // Database operations join the transaction
    s.repo1.Create(ctx, ...)
    s.repo2.Update(ctx, ...)
    
    return nil // Auto commit
}
```

**Note:** In HTTP handlers, prefer `ctx.BeginTransaction()` from `request.Context` as shown above.

### Transaction with Multiple Pools

Each pool name has its own transaction context:

```go
func (s *Service) Transfer(ctx *request.Context, amount float64) error {
    // Transaction for main database
    ctx.BeginTransaction("main-db")
    
    // Transaction for analytics database (separate)
    ctx.BeginTransaction("analytics-db")
    
    // Operations on main-db join main-db transaction
    s.mainRepo.Deduct(ctx, amount)
    
    // Operations on analytics-db join analytics-db transaction
    s.analyticsRepo.Log(ctx, "transfer", amount)
    
    return nil
}
```

### Ignoring Parent Transaction

Sometimes you need to execute operations outside of a transaction (e.g., audit logs that must commit immediately):

```go
import "github.com/primadi/lokstra/serviceapi"

func (s *Service) CreateWithAudit(ctx *request.Context, data *Data) error {
    ctx.BeginTransaction("main-db")
    
    // This joins the transaction
    if err := s.repo.Create(ctx, data); err != nil {
        return err
    }
    
    // This uses a separate connection (no transaction)
    isolatedCtx := serviceapi.WithoutTransaction(ctx)
    s.auditRepo.Log(isolatedCtx, "data_created") // Commits immediately
    
    return nil
}
```

## Summary

### Recommended Setup Flow

1. **Use lokstra_init** for initialization (handles everything correctly)
2. **Define pools** in YAML `dbpool-definitions:` section
3. **Inject DbPool** into services using `@Inject "pool-name"`
4. **Use transactions** via `ctx.BeginTransaction("pool-name")` in handlers

### Component Usage

| Component | Purpose | How to Get |
|-----------|---------|------------|
| **DbPoolManager** | Manages pools | Service: `"dbpool-manager"` |
| **DbPool** | Connection pool | Inject: `@Inject "pool-name"` or `lokstra_registry.GetService[DbPool]("pool-name")` |
| **DbConn** | Individual connection | `pool.Acquire(ctx)` |
| **Transaction** | Database transaction | `ctx.BeginTransaction("pool-name")` (recommended) |

## See Also

- [Lokstra Initialization](./09-lokstra-init.md) - **Recommended initialization approach**
- [Service Registration](./02-service/index.md) - Service setup
- [Dependency Injection](./07-inject-annotation.md) - Injection patterns
- [Configuration Management](./04-config/index.md) - YAML configuration
- [DbPool Manager API Reference](../03-api-reference/06-services/dbpool-manager.md) - API details
