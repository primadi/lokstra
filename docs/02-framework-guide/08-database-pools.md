---
title: Database Pools
layout: default
parent: Framework Guide
nav_order: 8
---

# Database Pools

Lokstra provides built-in support for database connection pooling with automatic configuration from YAML files.

## Setup Database Pools

### 1. Define DB Pools in Config

**config.yaml:**
```yaml
named-db-pools:
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

### 2. Explicit Setup (Recommended)

```go
func main() {
    lokstra.Bootstrap()
    
    // 1. Load config only
    if err := lokstra_registry.LoadConfig("config.yaml"); err != nil {
        log.Fatal(err)
    }
    
    // 2. Setup DB pools explicitly (if needed)
    if err := lokstra.SetupNamedDbPools(); err != nil {
        log.Fatal(err)
    }
    
    // 3. Run server
    lokstra_registry.InitAndRunServer()
}
```

### 3. Auto Setup (Legacy - Backward Compatible)

```go
func main() {
    lokstra.Bootstrap()
    
    // Auto-loads config + setup DB pools + run server
    lokstra_registry.RunServerFromConfig("config.yaml")
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
named-db-pools:
  mydb:
    dsn: "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
```

### Option 2: Component-Based (Recommended)

```yaml
named-db-pools:
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
lokstra.SetupNamedDbPools()
```

❌ **Bad:**
```go
// Auto-setup couples config loading with infrastructure
lokstra_registry.RunServerFromConfig("config.yaml")
```

### 2. Use Named Pools for Different Purposes

```yaml
named-db-pools:
  transactional-db:  # For OLTP workloads
    max_conns: 10
    
  analytics-db:      # For OLAP workloads
    max_conns: 50
    
  cache-db:          # For caching
    max_conns: 5
```

### 3. Environment-Specific Configuration

```yaml
named-db-pools:
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

## See Also

- [Service Registration](./03-services.md)
- [Dependency Injection](./04-dependency-injection.md)
- [Configuration Management](./05-configuration.md)
