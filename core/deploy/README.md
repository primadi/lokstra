# Lokstra Deploy Package

**Status:** ✨ New implementation (replaces `core/config`)

## Overview

The `deploy` package provides a declarative way to configure and deploy Lokstra applications using YAML configuration with code-based factories.

## Design Philosophy

### 1. Global Registry Pattern

All component definitions (configs, middlewares, services, routers) are stored in a **global registry**. Deployments simply select which components to use and where to deploy them.

**Benefits:**
- ✅ **DRY** - Define once, use everywhere
- ✅ **Consistency** - Same service definition across all deployments
- ✅ **Reusability** - Mix and match components
- ✅ **Maintainability** - Change once, affects all

### 2. Code-First, YAML Second

**Factories are always registered in code** (type-safe, compile-time checked). **YAML is used for wiring** (which services, where to deploy).

```go
// Factories in code (MUST be registered first)
deploy.Global().RegisterServiceType("user-factory", userFactoryLocal, userFactoryRemote)

// Wiring in YAML (defines instances and deployments)
// services:
//   - name: user-service
//     type: user-factory
```

### 3. Two-Step Config Resolution

Config values are resolved in two steps:

**Step 1:** Resolve all `${...}` placeholders except `${@cfg:...}`
- `${ENV_VAR}` - environment variable
- `${ENV_VAR:default}` - with default value
- `${@resolver:key}` - custom resolver (consul, aws-ssm, k8s, etc.)
- `${@resolver:key:default}` - custom resolver with default

**Step 2:** Resolve `${@cfg:...}` placeholders using Step 1 results
- `${@cfg:CONFIG_NAME}` - reference to resolved config value

**Example:**
```yaml
configs:
  - name: DB_MAX_CONNS
    value: 20
  
  - name: DB_USER_URL
    value: ${DB_USER_URL:postgres://localhost/users}

services:
  - name: db-user
    type: dbpool_pg
    config:
      dsn: ${DB_USER_URL}              # Step 1: env var
      max-conns: ${@cfg:DB_MAX_CONNS}  # Step 2: config reference (preserves type!)
```

**Why 2 steps?**
- Step 1 values can come from environment, consul, AWS, etc.
- Step 2 allows configs to reference other configs
- `${@cfg:...}` preserves types (integers stay integers, not strings)

## YAML Structure

```yaml
# ===== GLOBAL DEFINITIONS =====
configs:          # Configuration values
middlewares:      # Middleware instances
services:         # Service instances
routers:          # Manual routers (referenced by name)
router-overrides: # Route customizations
service-routers:  # Auto-generated routers from services

# ===== DEPLOYMENTS =====
deployments:      # Deployment configurations
  - name: monolith
    config-overrides: # Override global configs per deployment
    servers:          # Servers in this deployment
      - apps:         # Apps on this server
          services:        # Which services to instantiate
          routers:         # Which manual routers to use
          service-routers: # Which service routers to create
          remote-services: # Remote service proxies
```

## Config Resolution Examples

### Example 1: Simple Environment Variable
```yaml
configs:
  - name: JWT_SECRET
    value: ${JWT_SECRET:dev-secret}
```
- Looks up `JWT_SECRET` env var
- Falls back to `"dev-secret"` if not found

### Example 2: Custom Resolver
```yaml
configs:
  - name: API_KEY
    value: ${@consul:config/api-key:fallback}
```
- Uses consul resolver
- Looks up `config/api-key` in Consul
- Falls back to `"fallback"` if not found

### Example 3: Config Reference (Preserves Type)
```yaml
configs:
  - name: DB_MAX_CONNS
    value: 20  # Integer

services:
  - name: db-user
    config:
      max-conns: ${@cfg:DB_MAX_CONNS}  # Resolved as integer 20, not string "20"
```

### Example 4: Multiple References
```yaml
configs:
  - name: DB_HOST
    value: ${DB_HOST:localhost}
  - name: DB_PORT
    value: ${DB_PORT:5432}
  - name: DB_NAME
    value: mydb

services:
  - name: db
    config:
      dsn: "postgres://${DB_HOST}:${DB_PORT}/${@cfg:DB_NAME}"
      # Resolves to: "postgres://localhost:5432/mydb"
```

## Service Dependencies

Services can depend on other services using `depends-on`:

### Format 1: Direct Mapping
```yaml
services:
  - name: user-service
    type: user-factory
    depends-on: [db-user, logger]
```
- Parameter names match service names
- Factory signature: `func(dbUser service.Cached[*DBPool], logger service.Cached[*Logger]) any`

### Format 2: Alias Mapping
```yaml
services:
  - name: order-service
    type: order-factory
    depends-on: ["dbOrder:db-order", "userSvc:user-service"]
```
- Format: `"paramName:serviceName"`
- Parameter name `dbOrder` maps to service `db-order`
- Factory signature: `func(dbOrder service.Cached[*DBPool], userSvc service.Cached[UserService]) any`

## Router Overrides

Customize auto-generated service routers:

```yaml
router-overrides:
  - name: user-public-api
    path-prefix: /api/v1
    middlewares: [cors]           # Router-level middleware
    hidden: [Delete, BulkDelete]  # Hide these methods
    routes:
      - name: Create
        path: /register
        middlewares: [rate-limit]  # Route-level middleware
      - name: Update
        enabled: false             # Alternative way to hide
```

**Features:**
- `path-prefix` - URL prefix for all routes
- `middlewares` - Applied to all routes in this router
- `hidden` - Array of method names to hide
- `routes[].enabled` - Explicitly enable/disable individual routes
- `routes[].middlewares` - Method-specific middleware

## Deployments

### Monolith Deployment
```yaml
deployments:
  - name: monolith
    servers:
      - name: main-server
        apps:
          - port: 3000
            services: [db-user, user-service, order-service]
            service-routers:
              - name: user-service-public
```
- All services in one process
- Single server, single app

### Microservices Deployment
```yaml
deployments:
  - name: microservices
    servers:
      - name: user-server
        apps:
          - port: 3001
            services: [user-service]
            service-routers:
              - name: user-service-admin
      
      - name: order-server
        apps:
          - port: 3002
            services: [order-service]
            remote-services:
              - name: user-service
                url: http://localhost:3001
                service-router: user-service-admin
            service-routers:
              - name: order-service-public
```
- Services split across servers
- `remote-services` creates HTTP proxy clients

### Config Overrides Per Deployment
```yaml
deployments:
  - name: dev
    config-overrides:
      LOG_LEVEL: debug
      DB_MAX_CONNS: 5
    servers: [...]
  
  - name: prod
    config-overrides:
      LOG_LEVEL: error
      DB_MAX_CONNS: 100
      JWT_SECRET: ${@aws-ssm:/prod/jwt-secret}
    servers: [...]
```
- Override global configs per deployment
- Use different resolvers per environment

## Usage

### 1. Register Factories (Code)
```go
import "github.com/primadi/lokstra/core/deploy"

func init() {
    // Register service factories
    deploy.Global().RegisterServiceType("user-factory", 
        userFactoryLocal, 
        userFactoryRemote)
    
    deploy.Global().RegisterServiceType("order-factory", 
        orderFactoryLocal, 
        orderFactoryRemote)
    
    // Register middleware factories
    deploy.Global().RegisterMiddlewareType("jwt-auth", jwtAuthFactory)
    deploy.Global().RegisterMiddlewareType("cors-middleware", corsFactory)
    
    // Register custom resolvers (optional)
    deploy.Global().RegisterResolver(consulResolver)
}
```

### 2. Load YAML Configuration
```go
func main() {
    // Load YAML and build deployment
    dep, err := deploy.LoadYAML("deployment.yaml", "monolith")
    if err != nil {
        log.Fatal(err)
    }
    
    // Run deployment
    if err := dep.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### 3. Programmatic Usage (No YAML)
```go
func main() {
    // Create deployment
    dep := deploy.New("monolith")
    
    // Define services (equivalent to YAML)
    dep.DefineService(&schema.ServiceDef{
        Name: "user-service",
        Type: "user-factory",
        DependsOn: []string{"db-user", "logger"},
    })
    
    // Add to deployment
    dep.UseServices("user-service")
    
    // Run
    dep.Run()
}
```

## Custom Resolvers

Implement the `Resolver` interface:

```go
type MyResolver struct{}

func (r *MyResolver) Name() string {
    return "myresolver"
}

func (r *MyResolver) Resolve(key string) (string, bool) {
    // Custom logic to resolve key
    value, err := fetchFromSomewhere(key)
    if err != nil {
        return "", false
    }
    return value, true
}

// Register
deploy.Global().RegisterResolver(&MyResolver{})
```

**Use in YAML:**
```yaml
configs:
  - name: SECRET
    value: ${@myresolver:path/to/secret:default}
```

## Testing

```go
func TestDeployment(t *testing.T) {
    // Create isolated registry for testing
    registry := deploy.NewGlobalRegistry()
    
    // Register test factories
    registry.RegisterServiceType("test-service", testFactory, nil)
    
    // Define test config
    registry.DefineConfig(&schema.ConfigDef{
        Name: "TEST_VALUE",
        Value: "test",
    })
    
    // Test resolution
    if err := registry.ResolveConfigs(); err != nil {
        t.Fatal(err)
    }
    
    value, ok := registry.GetResolvedConfig("TEST_VALUE")
    if !ok || value != "test" {
        t.Errorf("expected 'test', got %v", value)
    }
}
```

## Migration from `core/config`

Old implementation (`core/config`) is renamed to `core/config_old` for reference.

**Key differences:**

| Feature | Old (`config_old`) | New (`deploy`) |
|---------|-------------------|----------------|
| Registry | Deployment-scoped | Global registry |
| Config resolution | Single-step | Two-step (env → @cfg) |
| YAML structure | Deployment-first | Global definitions → Deployments |
| Type preservation | Strings only | Preserves types with @cfg |
| Reusability | Duplicate definitions | DRY (define once) |

## Examples

See:
- `example.yaml` - Complete YAML configuration
- `resolver/resolver_test.go` - Config resolution tests
- `docs/00-introduction/examples/04-multi-deployment/` - Migration examples

## Architecture

```
GlobalRegistry (Singleton)
├── Resolver Registry
│   ├── env (default)
│   ├── consul
│   ├── aws-ssm
│   └── k8s
├── Factories (code)
│   ├── ServiceFactories (local + remote)
│   └── MiddlewareFactories
└── Definitions (YAML/code)
    ├── Configs
    ├── Middlewares
    ├── Services
    ├── Routers
    ├── RouterOverrides
    └── ServiceRouters

Deployment
├── Config Overrides
└── Servers
    └── Apps
        ├── Services (instances)
        ├── Routers (manual)
        ├── ServiceRouters (auto)
        └── RemoteServices (proxies)
```

## Best Practices

1. ✅ **Define factories in code** - Type safety, compile-time checks
2. ✅ **Use @cfg for config references** - Preserves types
3. ✅ **Use global definitions** - DRY, reusable
4. ✅ **Override configs per deployment** - Environment-specific values
5. ✅ **Use depends-on with aliases** - Clear parameter mapping
6. ✅ **Use router-overrides** - Fine-grained control
7. ✅ **Test with isolated registries** - Clean test state
