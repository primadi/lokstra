# Example 07 - Named Database Pools

Demonstrates configuring multiple named database pools with different settings.

## What's Demonstrated

- ✅ Multiple database pool configurations
- ✅ DSN vs component-based configuration
- ✅ Pool sizing (min/max connections)
- ✅ Connection timeouts and lifetimes
- ✅ SSL mode configuration
- ✅ Schema-specific pools
- ✅ Service dependencies on specific pools

## Named Database Pools

### Configuration Options

#### Option 1: Component-Based
```yaml
named-db-pools:
  main-db:
    host: "localhost"
    port: 5432
    database: "myapp"
    username: "postgres"
    password: "${DB_PASSWORD}"
    schema: "public"
    min-conns: 2
    max-conns: 10
    max-idle-time: "30m"
    max-lifetime: "1h"
    sslmode: "disable"
```

#### Option 2: DSN-Based
```yaml
named-db-pools:
  analytics-db:
    dsn: "postgres://user:pass@host:5432/db?sslmode=require"
    schema: "analytics"
    min-conns: 1
    max-conns: 5
```

## Pool Parameters

### Connection Sizing

**min-conns** (default: 2)
- Minimum connections kept alive
- Best practice: 2-4 for most apps

**max-conns** (default: 10)
- Maximum concurrent connections
- Best practice: 10-20 for web apps

```yaml
min-conns: 2   # Always keep 2 connections ready
max-conns: 10  # Allow up to 10 concurrent connections
```

### Connection Lifecycle

**max-idle-time** (default: 30m)
- How long idle connections stay alive
- Prevents stale connections

**max-lifetime** (default: 1h)
- Maximum lifetime of any connection
- Forces periodic connection refresh

```yaml
max-idle-time: "30m"  # Close idle connections after 30 minutes
max-lifetime: "1h"    # Refresh all connections every hour
```

### SSL Configuration

**sslmode** options:
- `disable` - No SSL (development only)
- `allow` - Try SSL, fallback to non-SSL
- `prefer` - Try SSL first (default)
- `require` - Require SSL
- `verify-ca` - Require SSL + verify CA
- `verify-full` - Require SSL + verify CA + hostname

```yaml
sslmode: "require"  # Production: always require SSL
```

## Use Cases

### 1. Main Application Database
```yaml
main-db:
  host: "localhost"
  database: "myapp"
  min-conns: 2
  max-conns: 10
  max-idle-time: "30m"
```

**Best for:**
- Primary transactional database
- High traffic, frequent queries
- CRUD operations

### 2. Analytics Database (Read-Only)
```yaml
analytics-db:
  dsn: "postgres://readonly:pass@analytics:5432/analytics"
  min-conns: 1
  max-conns: 5
  max-idle-time: "10m"
```

**Best for:**
- Read-only queries
- Long-running analytics queries
- Lower connection count (fewer writes)

### 3. Reporting Database
```yaml
reporting-db:
  host: "reporting-server"
  database: "reports"
  min-conns: 1
  max-conns: 3
  max-lifetime: "2h"
```

**Best for:**
- Scheduled report generation
- Infrequent access
- Minimal connection pool

## Service Dependencies

### Single Pool Dependency
```yaml
service-definitions:
  user-repository:
    type: user-repository-factory
    depends-on: [main-db]
```

**Factory signature:**
```go
func UserRepositoryFactory(deps map[string]any, config map[string]any) any {
    pool := deps["main-db"].(*pgxpool.Pool)
    return &UserRepository{pool: pool}
}
```

### Multiple Pool Dependencies
```yaml
service-definitions:
  report-generator:
    type: report-generator-factory
    depends-on:
      - mainDb:main-db
      - reportDb:reporting-db
```

**Factory signature:**
```go
func ReportGeneratorFactory(deps map[string]any, config map[string]any) any {
    mainPool := deps["mainDb"].(*pgxpool.Pool)
    reportPool := deps["reportDb"].(*pgxpool.Pool)
    
    return &ReportGenerator{
        mainPool:   mainPool,
        reportPool: reportPool,
    }
}
```

## Best Practices

### 1. Pool Sizing Guidelines

**Web API servers:**
```yaml
min-conns: 2
max-conns: 10
```

**Background workers:**
```yaml
min-conns: 1
max-conns: 5
```

**High-traffic services:**
```yaml
min-conns: 5
max-conns: 20
```

### 2. Connection Timeouts

**Interactive applications:**
```yaml
max-idle-time: "30m"
max-lifetime: "1h"
```

**Batch processing:**
```yaml
max-idle-time: "10m"
max-lifetime: "2h"
```

### 3. SSL Configuration

**Development:**
```yaml
sslmode: "disable"
```

**Production:**
```yaml
sslmode: "require"  # or verify-full
```

### 4. Environment Variables

```yaml
named-db-pools:
  main-db:
    host: "${DB_HOST}"
    port: 5432
    database: "${DB_NAME}"
    username: "${DB_USER}"
    password: "${DB_PASSWORD}"
```

## Monitoring

### Pool Health Metrics

Track these metrics in production:
- Active connections
- Idle connections
- Wait time for connections
- Connection errors
- Query duration

### Optimization Tips

**Problem: Connection pool exhausted**
```yaml
# Increase max-conns
max-conns: 20  # from 10
```

**Problem: Too many idle connections**
```yaml
# Decrease max-idle-time
max-idle-time: "15m"  # from 30m
```

**Problem: Stale connections**
```yaml
# Decrease max-lifetime
max-lifetime: "30m"  # from 1h
```

## Run

```bash
# Set environment variables
export DB_PASSWORD="secret123"
export ANALYTICS_PASSWORD="analytics456"
export REPORTING_PASSWORD="reporting789"

# Run application
go run main.go
```

## Summary

Named database pools allow you to:
- ✅ Configure multiple databases with different settings
- ✅ Optimize connection pooling per use case
- ✅ Use DSN or component-based configuration
- ✅ Inject specific pools into services
- ✅ Monitor and tune pool performance
- ✅ Keep secrets in environment variables
