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
    // services.RegisterAuthServices()    // Only auth-related services
}
```

## Available Services

### Core Infrastructure Services

| Service | Type | Contract | Description |
|---------|------|----------|-------------|
| **Redis** | `redis` | `serviceapi.Redis` | Redis client wrapper |
| **KvStore** | `kvstore_redis` | `serviceapi.KvStore` | Key-value store with Redis backend |
| **Metrics** | `metrics_prometheus` | `serviceapi.Metrics` | Prometheus metrics collection |
| **DbPool** | `dbpool_pg` | `serviceapi.DbPool` | PostgreSQL connection pool |

### Authentication Services

| Service | Type | Contract | Description |
|---------|------|----------|-------------|
| **Session** | `auth_session_redis` | `auth.Session` | Session management with Redis |
| **TokenIssuer** | `auth_token_jwt` | `auth.TokenIssuer` | JWT token generation and verification |
| **UserRepository** | `auth_user_repo_pg` | `auth.UserRepository` | User CRUD with PostgreSQL |
| **Flow (Password)** | `auth_flow_password` | `auth.Flow` | Username/password authentication |
| **Flow (OTP)** | `auth_flow_otp` | `auth.Flow` | One-time password authentication |
| **Auth Service** | `auth_service` | `auth.Service` | Main authentication orchestrator |
| **Validator** | `auth_validator` | `auth.Validator` | Token validation for middleware |

## Service Pattern

All services follow this standard pattern:

```go
package myservice

import (
    "github.com/primadi/lokstra/common/utils"
    "github.com/primadi/lokstra/lokstra_registry"
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
    lokstra_registry.RegisterServiceFactory(SERVICE_TYPE, ServiceFactory,
        lokstra_registry.AllowOverride(true))
}
```

## Usage Examples

### 1. Creating a Database Pool

```go
import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/services/dbpool_pg"
)

dbpool_pg.Register()

dbPool := lokstra_registry.NewService[any](
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
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/services/kvstore_redis"
)

kvstore_redis.Register()

kvStore := lokstra_registry.NewService[any](
    "my_cache", "kvstore_redis",
    map[string]any{
        "addr":   "localhost:6379",
        "prefix": "myapp",
    },
)
```

### 3. Setting Up Authentication

```go
import (
    "context"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi/auth"
    "github.com/primadi/lokstra/services"
)

// Register all auth services
services.RegisterAuthServices()

// Configure services
lokstra_registry.NewService[any]("my_db", "dbpool_pg", map[string]any{...})
lokstra_registry.NewService[any]("my_token_issuer", "auth_token_jwt", map[string]any{...})
lokstra_registry.NewService[any]("my_session", "auth_session_redis", map[string]any{...})
lokstra_registry.NewService[any]("my_user_repo", "auth_user_repo_pg", map[string]any{...})
lokstra_registry.NewService[any]("my_password_flow", "auth_flow_password", map[string]any{...})

// Create main auth service
authSvc := lokstra_registry.NewService[auth.Service](
    "my_auth", "auth_service",
    map[string]any{
        "token_issuer_service_name": "my_token_issuer",
        "session_service_name":      "my_session",
        "flow_service_names": map[string]string{
            "password": "my_password_flow",
        },
    },
)

// Use the auth service
resp, err := authSvc.Login(context.Background(), auth.LoginRequest{
    Flow: "password",
    Payload: map[string]any{
        "tenant_id": "tenant-123",
        "username":  "user@example.com",
        "password":  "password123",
    },
})
```

### 4. Using Metrics

```go
import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
    "github.com/primadi/lokstra/services/metrics_prometheus"
)

metrics_prometheus.Register()

metrics := lokstra_registry.NewService[serviceapi.Metrics](
    "my_metrics", "metrics_prometheus",
    map[string]any{
        "namespace": "myapp",
        "subsystem": "api",
    },
)

// Track metrics
metrics.IncCounter("requests_total", serviceapi.Labels{
    "method": "GET",
    "path":   "/api/users",
})

metrics.ObserveHistogram("request_duration_seconds", 0.123, serviceapi.Labels{
    "method": "GET",
    "path":   "/api/users",
})
```

## Service Dependencies

Some services depend on other services. Here's the dependency graph:

```
auth_service
├── auth_token_jwt (TokenIssuer)
├── auth_session_redis (Session)
│   └── redis
└── auth_flow_* (Flows)
    ├── auth_user_repo_pg (UserRepository)
    │   └── dbpool_pg (DbPool)
    └── kvstore_redis (for OTP flow)
        └── redis

auth_validator
├── auth_token_jwt (TokenIssuer)
└── auth_user_repo_pg (optional, UserRepository)
    └── dbpool_pg (DbPool)
```

When creating dependent services, make sure to register and create the dependency services first.

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
      
  my_auth:
    type: auth_service
    config:
      token_issuer_service_name: my_token_issuer
      session_service_name: my_session
      flow_service_names:
        password: my_password_flow
        otp: my_otp_flow
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
