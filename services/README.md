# Lokstra Services

This directory contains built-in service implementations for the Lokstra framework. All services follow a standard pattern with three main functions: `Service()`, `ServiceFactory()`, and `Register()`.

## Quick Start

```go
package main

import (
    "github.com/primadi/lokstra/services"
)

func main() {
    // Register all built-in services
    services.RegisterAllServices()
    
    // Or register selectively:
    // services.RegisterCoreServices()    // Only Redis, KvStore, Metrics, DbPool
}

> **Note:** Authentication services have been moved to [github.com/primadi/lokstra-auth](https://github.com/primadi/lokstra-auth)
```

## Available Services

### Core Infrastructure Services

| Service | Type | Contract | Description |
|---------|------|----------|-------------|
| **Redis** | `redis` | `serviceapi.Redis` | Redis client wrapper |
| **KvStore** | `kvstore_redis` | `serviceapi.KvStore` | Key-value store with Redis backend |
| **Metrics** | `metrics_prometheus` | `serviceapi.Metrics` | Prometheus metrics collection |
| **DbPool** | `dbpool_pg` | `serviceapi.DbPool` | PostgreSQL connection pool |

> **Note:** Authentication services (Session, TokenIssuer, UserRepository, Auth Flows, etc.) have been moved to [github.com/primadi/lokstra-auth](https://github.com/primadi/lokstra-auth)

## Service Pattern

All services follow this standard pattern:

```go
package myservice

import (
    "github.com/primadi/lokstra/common/utils"
    "github.com/primadi/lokstra/old_registry"
)

const SERVICE_TYPE = "myservice"

type Config struct {
    // Configuration fields
}

type myService struct {
    cfg *Config
    // other dependencies
}

func Service(cfg *Config) *myService {
    return &myService{cfg: cfg}
}

func ServiceFactory(params map[string]any) any {
    cfg := &Config{
        // Extract config from params using utils.GetValueFromMap
    }
    return Service(cfg)
}

func Register() {
    old_registry.RegisterServiceFactory(SERVICE_TYPE, ServiceFactory,
        old_registry.AllowOverride(true))
}
```

## Usage Examples

### 1. Creating a Database Pool

```go
import (
    "github.com/primadi/lokstra/old_registry"
    "github.com/primadi/lokstra/services/dbpool_pg"
)

dbpool_pg.Register()

dbPool := old_registry.NewService[any](
    "main_db", "dbpool_pg",
    map[string]any{
        "host":     "localhost",
        "port":     5432,
        "database": "myapp",
        "username": "postgres",
        "password": "password",
    },
)
```

### 2. Creating a KvStore

```go
import (
    "github.com/primadi/lokstra/old_registry"
    "github.com/primadi/lokstra/services/kvstore_redis"
)

kvstore_redis.Register()

kvStore := old_registry.NewService[any](
    "my_cache", "kvstore_redis",
    map[string]any{
        "addr":   "localhost:6379",
        "prefix": "myapp",
    },
)
```

> **Note:** For authentication examples, see [github.com/primadi/lokstra-auth](https://github.com/primadi/lokstra-auth)

## Configuration via YAML

All services can be configured via YAML files using the Lokstra configuration system:

```yaml
services:
  main_db:
    type: dbpool_pg
    config:
      host: localhost
      port: 5432
      database: myapp
      username: postgres
      password: ${DB_PASSWORD} # Environment variable
```

## Testing

Each service implementation includes its own test file. Run tests with:

```bash
go test ./services/...
```

## Documentation

For detailed documentation on each service, see:
- [Service Implementations Guide](../docs/service-implementations.md)
- [API Standard](../docs/api-standard.md)
- [Architecture](../docs/architecture.md)

## Adding New Services

To add a new service implementation:

1. Create a new directory under `/services`
2. Implement the service following the standard pattern
3. Create `module.go` with `Service()`, `ServiceFactory()`, and `Register()` functions
4. Add tests
5. Update `register_all.go` to include the new service
6. Update this README

## License

See the LICENSE file in the project root.
