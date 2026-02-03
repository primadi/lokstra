---
name: implementation-lokstra-init-framework
description: Initialize Lokstra Framework in main.go. Setup lokstra_init.BootstrapAndRun(), service registration, and configuration loading. Use after design-lokstra-overview to create working application entry point.
phase: implementation
order: 1
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.18"
references:
  - references/01-basic-main.go
  - references/02-with-options-main.go
  - references/03-multi-module-main.go
  - references/04-custom-router-main.go
  - references/05-initialization-options.md
  - references/06-startup-flow.md
  - references/07-troubleshooting.md
---

# Implementation: Lokstra Framework Initialization

## Overview

This skill teaches you how to initialize the Lokstra Framework in `main.go`. The framework uses a convention-over-configuration approach with automatic code generation from annotations. Unlike traditional Go web frameworks, you **don't manually register routes** - they're discovered from `@Handler` and `@Route` annotations.

## When to Use

Use this skill when:
- ✅ Starting a new Lokstra project and need to create `main.go`
- ✅ Setting up framework initialization with middleware and services
- ✅ Configuring database connections and migrations
- ✅ Setting up multi-module applications with dependency injection
- ✅ Configuring development vs production environments

Prerequisites:
- ✅ Project structure created (`configs/`, `modules/` folders)
- ✅ `config.yaml` file exists in `configs/` folder
- ✅ At least 1 module with `@Handler` annotation
- ✅ `go.mod` initialized with lokstra dependency

```bash
# Initialize Go module
go mod init myapp

# Add Lokstra dependency
go get github.com/primadi/lokstra
```


---

## Basic Setup (main.go)

### Minimal Configuration

The simplest working `main.go` for a Lokstra application:

```go
package main

import (
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

func main() {
	// 1. Register middleware (order matters!)
	recovery.Register()       // Panic recovery - always first
	request_logger.Register() // Request logging

	// 2. Register infrastructure services
	dbpool_pg.Register() // PostgreSQL connection pooling

	// 3. Bootstrap and run
	lokstra_init.BootstrapAndRun()
}
```

**What Happens:**

1. **Middleware Registration** - Registers middleware in the framework registry (order matters!)
2. **Service Registration** - Registers service factory functions for later instantiation
3. **BootstrapAndRun()** - Performs the following:
   - Scans for `@Handler` and `@Service` annotations
   - Generates route registration code (`zz_generated.lokstra.go`)
   - Loads `configs/*.yaml` files (auto-merged)
   - Creates services in dependency order
   - Creates `@Handler` instances with dependency injection
   - Mounts HTTP routes
   - Starts server on configured address

**Note:** You do NOT need to call `lokstra.Bootstrap()` manually. It's called internally by `BootstrapAndRun()`.

**Reference:** See [references/01-basic-main.go](references/01-basic-main.go) for complete example.

---

## Multi-Module Setup

For applications with multiple Domain-Driven Design (DDD) modules:

```go
package main

import (
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

func main() {
	recovery.Register()
	request_logger.Register()
	dbpool_pg.Register()

	lokstra_init.BootstrapAndRun()
}
```

**Important:** You do NOT need to manually import module packages! Module imports are **automatically generated** in `zz_lokstra_imports.go` during bootstrap.

**How It Works:**
1. `BootstrapAndRun()` scans all directories for `@Handler` and `@Service` annotations
2. Generates `zz_lokstra_imports.go` at project root with all required imports
3. Auto-restarts the application if code changes (dev mode only)
4. All modules are automatically discovered and registered

**Project Structure:**
```
myapp/
├── main.go                    # Your clean main file (no module imports!)
├── zz_lokstra_imports.go      # AUTO-GENERATED - contains all module imports
├── configs/
│   └── config.yaml
└── modules/
    ├── auth/
    │   └── application/
    │       ├── handler.go
    │       └── zz_generated.lokstra.go  # AUTO-GENERATED
    ├── user/
    │   └── application/
    │       ├── handler.go
    │       └── zz_generated.lokstra.go  # AUTO-GENERATED
    └── order/
        └── application/
            ├── handler.go
            └── zz_generated.lokstra.go  # AUTO-GENERATED
```

**Reference:** See [references/03-multi-module-main.go](references/03-multi-module-main.go) for complete example.

---

## Advanced Configuration with Options

Customize framework behavior using initialization options:

```go
package main

import (
	"fmt"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/cors"
	"github.com/primadi/lokstra/middleware/gzipcompression"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/services/eventbus"
	"github.com/primadi/lokstra/services/kvstore"
)

func main() {
	// Register middleware
	recovery.Register()
	request_logger.Register()
	cors.Register([]string{"*"})
	gzipcompression.Register()

	// Register services
	dbpool_pg.Register()
	eventbus.Register()
	kvstore.Register()

	// Bootstrap with options
	if err := lokstra_init.BootstrapAndRun(
		// Set log level (default: Info)
		lokstra_init.WithLogLevel(logger.LogLevelDebug),

		// Enable database migrations (default: false)
		lokstra_init.WithDbMigrations(true, "migrations"),

		// Enable PostgreSQL distributed config sync (default: false)
		lokstra_init.WithPgSyncMap(true, "db_main"),

		// Custom server initialization hook
		lokstra_init.WithServerInitFunc(func() error {
			fmt.Println("╔════════════════════════════════════╗")
			fmt.Println("║   Server Initializing...          ║")
			fmt.Println("╚════════════════════════════════════╝")

			// Custom initialization logic here
			return nil
		}),
	); err != nil {
		panic("Failed to initialize: " + err.Error())
	}
}
```

**Available Options:**

| Option | Description | Default |
|--------|-------------|---------|
| `WithLogLevel(level)` | Set logging level | `logger.LogLevelInfo` |
| `WithDbMigrations(enable, folder)` | Auto-run migrations on startup | `false`, `"migrations"` |
| `WithPgSyncMap(enable, dbPool)` | Distributed config sync | `false`, `"db_main"` |
| `WithServerInitFunc(fn)` | Custom initialization hook | `nil` |
| `WithYAMLConfigPath(enable, paths...)` | Custom config paths | `true`, `"configs"` |
| `WithAnnotations(enable, paths...)` | Annotation scanning | `true`, current dir |
| `WithAutoRunServer(enable)` | Auto-start server | `true` |
| `WithPanicOnConfigError(enable)` | Panic on config error | `true` |

**Reference:** 
- [references/02-with-options-main.go](references/02-with-options-main.go) - Complete example with all options
- [references/05-initialization-options.md](references/05-initialization-options.md) - Detailed option documentation


---

## Configuration File Structure

Lokstra requires a `config.yaml` in the `configs/` folder. This defines services and deployment configurations.

### Minimal config.yaml

```yaml
# configs/config.yaml
service-definitions:
  # Define infrastructure services
  db_main:
    type: dbpool_pg
    config:
      dsn: ${DB_DSN:postgres://user:pass@localhost:5432/mydb}
      schema: ${DB_SCHEMA:public}

deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services:
          - user-handler  # Must match @Handler name="user-handler"
          - auth-handler  # Must match @Handler name="auth-handler"
```

**Key Sections:**

1. **service-definitions** - Infrastructure services (DB pools, caches, queues)
2. **deployments** - Server configurations per environment
3. **published-services** - List of `@Handler` names to expose in this server

**Note:** Multiple YAML files in `configs/` folder are automatically merged. Use this for:
- `configs/base.yaml` - Shared configuration
- `configs/development.yaml` - Dev overrides
- `configs/production.yaml` - Prod overrides

---

## Middleware Registration

### Registration Order Matters!

Middleware is applied in **reverse order** of registration:

```go
// Registration order
recovery.Register()        // 1. Registered first
request_logger.Register()  // 2. Registered second
cors.Register([]string{"*"}) // 3. Registered third
auth.Register()            // 4. Registered last

// Execution order for incoming requests:
// Request → auth → cors → request_logger → recovery → Handler
```

### Recommended Middleware Order

```go
// 1. ALWAYS FIRST - Catches panics in all downstream middleware
recovery.Register()

// 2. Logging - Captures all requests
request_logger.Register()

// 3. CORS - Must be before auth (OPTIONS preflight)
cors.Register([]string{"*"})

// 4. Compression
gzipcompression.Register()

// 5. Authentication/Authorization
auth_middleware.Register()

// 6. Custom business middleware
rate_limiter.Register()
tenant_resolver.Register()
```

**Why this order?**
- **Recovery first** - Catches panics from all middleware
- **Logging early** - Captures all request details
- **CORS before auth** - OPTIONS preflight doesn't need auth
- **Auth before business logic** - Protect endpoints

---

## Service Registration Pattern

### Built-in Services

```go
import (
	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/services/eventbus"
	"github.com/primadi/lokstra/services/kvstore"
	"github.com/primadi/lokstra/services/metrics_prometheus"
)

func main() {
	// Database connection pooling
	dbpool_pg.Register()
	
	// In-memory event bus
	eventbus.Register()
	
	// Key-value store (in-memory or Redis)
	kvstore.Register()
	
	// Prometheus metrics
	metrics_prometheus.Register()

	lokstra_init.BootstrapAndRun()
}
```

### Custom Service Registration

For custom `@Service` annotations, register the service factory:

```go
// modules/user/infrastructure/register.go
import "github.com/primadi/lokstra/lokstra_registry"

func RegisterUserServices() {
	lokstra_registry.RegisterServiceFactory("user-repository", NewUserRepository)
}

// main.go
import "myapp/modules/user/infrastructure"

func main() {
	// Register custom services
	infrastructure.RegisterUserServices()
	
	lokstra_init.BootstrapAndRun()
}
```

**Note:** Most custom services use `@Service` annotations and don't require manual registration.

---

## Environment Detection

Lokstra automatically detects the runtime environment:

| Mode | Detection | Behavior |
|------|-----------|----------|
| **Production** | Compiled binary | No code generation, optimized |
| **Development** | `go run .` | Auto code generation, hot reload |
| **Debug** | `go run . -debug` or Delve | Auto code generation, debugger support |

### Force Code Generation

Use `--generate-only` flag to generate code without starting server:

```bash
# Generate code and exit
go run . --generate-only

# Verify generated files
ls modules/user/application/zz_generated.lokstra.go
ls zz_lokstra_imports.go  # Contains auto-generated module imports
```

**Generated Files:**
- `zz_generated.lokstra.go` - Per-module route registration code
- `zz_lokstra_imports.go` - Auto-generated module imports (project root)

**Use Cases:**
- CI/CD pipelines
- Pre-commit hooks
- Debugging annotation issues
- Manual code generation

---

## Advanced: Custom Router Registration

For complex deployments with multiple routers, use `WithServerInitFunc`:

```go
package main

import (
	"fmt"

	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

func main() {
	recovery.Register()
	dbpool_pg.Register()

	if err := lokstra_init.BootstrapAndRun(
		lokstra_init.WithServerInitFunc(func() error {
			fmt.Println("Registering custom routers...")

			// Register router instances
			registerRouters()

			// Register middleware types
			registerMiddlewareTypes()

			return nil
		}),
	); err != nil {
		panic("Failed to initialize: " + err.Error())
	}
}

func registerRouters() {
	// Create and register custom routers
	apiRouter := lokstra.NewRouter("api-router")
	adminRouter := lokstra.NewRouter("admin-router")

	lokstra_registry.RegisterRouter("api-router", apiRouter)
	lokstra_registry.RegisterRouter("admin-router", adminRouter)
}

func registerMiddlewareTypes() {
	// Register custom middleware factories
	lokstra_registry.RegisterMiddlewareType("auth", NewAuthMiddleware)
	lokstra_registry.RegisterMiddlewareType("rate-limit", NewRateLimitMiddleware)
}
```

**Use Cases for ServerInitFunc:**
- Multi-router deployments
- Custom middleware type registration
- Dynamic service configuration
- Environment-specific initialization
- Pre-startup validation

**Reference:** See [references/04-custom-router-main.go](references/04-custom-router-main.go) for complete example.

---

## Startup Flow Deep Dive

Understanding the complete bootstrap sequence:

```
┌─────────────────────────────────────┐
│ 1. main() execution starts          │
└─────────────────────────────────────┘
                ↓
┌─────────────────────────────────────┐
│ 2. Middleware Registration          │
│    - recovery.Register()            │
│    - request_logger.Register()      │
│    - Custom middleware              │
└─────────────────────────────────────┘
                ↓
┌─────────────────────────────────────┐
│ 3. Service Registration             │
│    - dbpool_pg.Register()           │
│    - eventbus.Register()            │
│    - Custom services                │
└─────────────────────────────────────┘
                ↓
┌─────────────────────────────────────┐
│ 4. lokstra_init.BootstrapAndRun()  │
└─────────────────────────────────────┘
                ↓
      ┌─────────────────────────┐
      │ 4.1 Environment Detection│
      │     prod/dev/debug mode  │
      └─────────────────────────┘
                ↓
      ┌─────────────────────────┐
      │ 4.2 Code Generation     │
      │     Scan @Handler       │
      │     Generate routes     │
      │     Generate imports    │
      │     (zz_lokstra_imports)│
      │     Auto-restart if ↑   │
      └─────────────────────────┘
                ↓
      ┌─────────────────────────┐
      │ 4.3 Config Loading      │
      │     Load configs/*.yaml │
      │     Merge YAML files    │
      │     Substitute ${ENV}   │
      └─────────────────────────┘
                ↓
      ┌─────────────────────────┐
      │ 4.4 DB Migrations       │
      │     Run .up.sql files   │
      └─────────────────────────┘
                ↓
      ┌─────────────────────────┐
      │ 4.5 Service Creation    │
      │     Resolve deps graph  │
      │     Inject dependencies │
      └─────────────────────────┘
                ↓
      ┌─────────────────────────┐
      │ 4.6 Handler Creation    │
      │     Create @Handler     │
      │     Inject @Service     │
      └─────────────────────────┘
                ↓
      ┌─────────────────────────┐
      │ 4.7 ServerInitFunc      │
      │     Custom init hook    │
      └─────────────────────────┘
                ↓
      ┌─────────────────────────┐
      │ 4.8 Route Mounting      │
      │     Mount @Route        │
      │     Apply middleware    │
      └─────────────────────────┘
                ↓
      ┌─────────────────────────┐
      │ 4.9 Server Start        │
      │     Bind address        │
      │     Start HTTP listener │
      └─────────────────────────┘
                ↓
┌─────────────────────────────────────┐
│ 5. Server Running ✅                │
│    Ready to accept requests         │
└─────────────────────────────────────┘
```

**Key Points:**

1. **Code Generation** (dev/debug only) - Scans annotations and generates route registration code
2. **Auto-restart** - If code changed, restarts with `go run` or debugger
3. **Dependency Resolution** - Services created in topological order based on dependencies
4. **Fail-Fast** - Panics on configuration errors by default (configurable)

**Reference:** See [references/06-startup-flow.md](references/06-startup-flow.md) for detailed flow documentation.

---

## Common Issues and Solutions

### Issue: "no published-services found"

**Cause:** No `@Handler` annotations discovered.

**Solutions:**
1. Verify `@Handler` annotation syntax:
   ```go
   // @Handler name="user-handler", prefix="/api/users"
   ```
2. Ensure handler files are in correct location:
   ```
   modules/user/application/handler.go
   ```
3. Force regeneration:
   ```bash
   go run . --generate-only
   ```
4. Check generated imports file:
   ```bash
   cat zz_lokstra_imports.go
   # Should contain auto-generated imports for all modules
   ```

### Issue: Service Not Found Error

**Cause:** Service injected but never registered or defined in config.

**Solutions:**
1. Add service to `configs/config.yaml`:
   ```yaml
   service-definitions:
     user-repo:
       type: user_repository
       config: {}
   ```
2. Verify `@Service` annotation exists
3. Check `published-services` list includes the handler

### Issue: Port Already in Use

**Solutions:**
1. Change port in `configs/config.yaml`:
   ```yaml
   deployments:
     development:
       servers:
         api:
           addr: ":8081"
   ```
2. Kill process using port:
   ```bash
   # Windows
   netstat -ano | findstr :8080
   taskkill /PID <PID> /F
   
   # Linux/Mac
   lsof -ti:8080 | xargs kill -9
   ```

### Issue: Database Connection Failed

**Solutions:**
1. Check DSN in `configs/config.yaml`:
   ```yaml
   service-definitions:
     db_main:
       config:
         dsn: "postgres://user:pass@localhost:5432/mydb"
   ```
2. Use environment variables:
   ```yaml
   dsn: ${DB_DSN:postgres://localhost:5432/mydb}
   ```
3. Verify database is running:
   ```bash
   pg_isready -h localhost -p 5432
   ```

**Reference:** See [references/07-troubleshooting.md](references/07-troubleshooting.md) for comprehensive troubleshooting guide.

---

## Testing Your Setup

### 1. Verify Framework Initialization

```bash
# Start the application
go run .

# Expected output:
# [Lokstra] Environment detected: DEV
# [Lokstra] Starting server on :8080
# ✅ Server ready - registered routes:
#   GET  /api/users
#   POST /api/users
```

### 2. Test Health Endpoint

```bash
# If you have a health endpoint
curl http://localhost:8080/health

# Expected: {"status":"ok"}
```

### 3. Force Code Generation

```bash
# Generate code without starting server
go run . --generate-only

# Verify generated files exist
ls modules/*/application/zz_generated.lokstra.go
ls zz_lokstra_imports.go
```

### 4. Enable Debug Logging

```go
lokstra_init.BootstrapAndRun(
    lokstra_init.WithLogLevel(logger.LogLevelDebug),
)
```

---

## Complete Example: Enterprise Application

Here's a complete production-ready `main.go`:

```go
package main

import (
	"fmt"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/cors"
	"github.com/primadi/lokstra/middleware/gzipcompression"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/middleware/slow_request_logger"
	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/services/eventbus"
	"github.com/primadi/lokstra/services/kvstore"

	// Module imports are AUTO-GENERATED in zz_lokstra_imports.go
	// No need to manually import modules!
)

func main() {
	// Middleware (order matters!)
	recovery.Register()
	request_logger.Register()
	slow_request_logger.Register()
	cors.Register([]string{"https://app.example.com", "https://admin.example.com"})
	gzipcompression.Register()

	// Infrastructure services
	dbpool_pg.Register()
	eventbus.Register()
	kvstore.Register()

	// Bootstrap with enterprise options
	if err := lokstra_init.BootstrapAndRun(
		// Production log level
		lokstra_init.WithLogLevel(logger.LogLevelInfo),

		// Enable database migrations
		lokstra_init.WithDbMigrations(true, "migrations"),

		// Enable distributed config sync
		lokstra_init.WithPgSyncMap(true, "db_main"),

		// Custom initialization
		lokstra_init.WithServerInitFunc(func() error {
			printBanner()
			return nil
		}),
	); err != nil {
		panic("❌ Failed to start application: " + err.Error())
	}
}

func printBanner() {
	fmt.Println("")
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║   Enterprise Application v1.0.0        ║")
	fmt.Println("║   Powered by Lokstra Framework         ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println("")
}
```

**Corresponding configs/config.yaml:**

```yaml
service-definitions:
  # Main database pool
  db_main:
    type: dbpool_pg
    config:
      dsn: ${DB_DSN:postgres://user:pass@localhost:5432/myapp}
      schema: ${DB_SCHEMA:public}
      max_connections: ${DB_MAX_CONN:25}
      min_connections: ${DB_MIN_CONN:5}

  # Event bus for async operations
  eventbus_main:
    type: eventbus

  # Key-value store for caching
  kvstore_main:
    type: kvstore
    config:
      type: ${KVSTORE_TYPE:memory}  # or redis
      redis_url: ${REDIS_URL}

deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services:
          - auth-handler
          - user-handler
          - product-handler
          - order-handler
          - payment-handler
          - notification-handler

  production:
    servers:
      api:
        addr: ":8080"
        published-services:
          - auth-handler
          - user-handler
          - product-handler
          - order-handler
          - payment-handler
          - notification-handler
```

---

## Next Steps

After setting up `main.go`, proceed with these skills in order:

1. **Configuration** → [implementation-lokstra-yaml-config](../implementation-lokstra-yaml-config/SKILL.md)
   - Setup `configs/config.yaml`
   - Configure service definitions
   - Define deployment environments

2. **Database Schema** → [design-lokstra-schema-design](../design-lokstra-schema-design/SKILL.md)
   - Design PostgreSQL schema
   - Create migration files

3. **Services** → [implementation-lokstra-create-service](../implementation-lokstra-create-service/SKILL.md)
   - Create `@Service` annotated repositories
   - Implement data access layer

4. **Handlers** → [implementation-lokstra-create-handler](../implementation-lokstra-create-handler/SKILL.md)
   - Create `@Handler` annotated HTTP endpoints
   - Implement business logic

5. **Testing** → [advanced-lokstra-tests](../advanced-lokstra-tests/SKILL.md)
   - Unit tests
   - Integration tests
   - Mocks

---

## Quick Reference

### Essential Imports

```go
// Framework initialization
"github.com/primadi/lokstra/lokstra_init"

// Common middleware
"github.com/primadi/lokstra/middleware/recovery"
"github.com/primadi/lokstra/middleware/request_logger"
"github.com/primadi/lokstra/middleware/cors"

// Common services
"github.com/primadi/lokstra/services/dbpool_pg"
"github.com/primadi/lokstra/services/eventbus"

// Logging
"github.com/primadi/lokstra/common/logger"
```

### Common Patterns

```go
// Basic setup
recovery.Register()
lokstra_init.BootstrapAndRun()

// With database
dbpool_pg.Register()
lokstra_init.BootstrapAndRun(
    lokstra_init.WithDbMigrations(true, "migrations"),
)

// With custom init
lokstra_init.BootstrapAndRun(
    lokstra_init.WithServerInitFunc(func() error {
        // Custom initialization
        return nil
    }),
)

// Development mode
lokstra_init.BootstrapAndRun(
    lokstra_init.WithLogLevel(logger.LogLevelDebug),
)
```

---

## Resources

**Reference Files:**
- [01-basic-main.go](references/01-basic-main.go) - Minimal setup
- [02-with-options-main.go](references/02-with-options-main.go) - All options
- [03-multi-module-main.go](references/03-multi-module-main.go) - Multiple modules
- [04-custom-router-main.go](references/04-custom-router-main.go) - Custom routers
- [05-initialization-options.md](references/05-initialization-options.md) - Option reference
- [06-startup-flow.md](references/06-startup-flow.md) - Detailed flow
- [07-troubleshooting.md](references/07-troubleshooting.md) - Common issues

**Related Skills:**
- [design-lokstra-overview](../design-lokstra-overview/SKILL.md) - Framework fundamentals
- [implementation-lokstra-yaml-config](../implementation-lokstra-yaml-config/SKILL.md) - Configuration
- [implementation-lokstra-create-handler](../implementation-lokstra-create-handler/SKILL.md) - HTTP handlers
- [implementation-lokstra-create-service](../implementation-lokstra-create-service/SKILL.md) - Services
- [advanced-lokstra-middleware](../advanced-lokstra-middleware/SKILL.md) - Custom middleware

---

## Summary

✅ **Framework Initialization** - Use `lokstra_init.BootstrapAndRun()`
✅ **NO Manual Module Imports** - Auto-generated in `zz_lokstra_imports.go`
✅ **Middleware Registration** - Register in correct order (recovery first)
✅ **Service Registration** - Register infrastructure services
✅ **Configuration** - Setup `configs/config.yaml` with service definitions
✅ **Options** - Customize with `WithLogLevel`, `WithDbMigrations`, etc.
✅ **Environment Detection** - Auto-detects prod/dev/debug mode
✅ **Code Generation** - Automatic in dev mode, manual with `--generate-only`

The Lokstra Framework handles route registration, dependency injection, module discovery, and service lifecycle automatically. Focus on writing `@Handler` and `@Service` annotated code - the framework does the rest!
