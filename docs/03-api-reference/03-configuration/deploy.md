# Deploy

> Deployment topology management and YAML-based deployment configuration

## Overview

The `deploy` package provides deployment topology management for Lokstra applications. It manages the global registry, handles YAML-based deployment configurations, and provides a 2-layer architecture for deployments (Deployment → Server → App).

## Import Path

```go
import (
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/core/deploy/schema"
    "github.com/primadi/lokstra/core/deploy/loader"
)
```

---

## Core Concepts

### 2-Layer Architecture

Lokstra uses a simplified 2-layer deployment model:

```
Deployment (Environment: prod, staging, dev)
  └─ Servers (Physical/Virtual servers)
       └─ Apps (HTTP listeners on ports)
            └─ Routers (Route handlers)
```

**Key Points:**
- **Deployment** - Environment grouping (production, staging, development)
- **Server** - Physical or virtual server instance (BaseURL, services)
- **App** - HTTP listener on a specific address (port/socket)
- **Router** - Route handler (manual or auto-generated from services)

**Services are at SERVER level** - shared across all apps in that server.

---

## Global Registry

### Global
Returns the singleton global registry instance.

**Signature:**
```go
func Global() *GlobalRegistry
```

**Example:**
```go
registry := deploy.Global()

// Register service type
registry.RegisterServiceType("user-service", localFactory, remoteFactory,
    deploy.WithResource("user", "users"))

// Define service instance
registry.DefineService(&schema.ServiceDef{
    Name: "user-svc",
    Type: "user-service",
})
```

---

### GlobalRegistry
Main registry for all global definitions and runtime instances.

**Type:**
```go
type GlobalRegistry struct {
    // Factories
    serviceFactories    map[string]*ServiceFactoryEntry
    middlewareFactories map[string]MiddlewareFactory
    
    // Definitions (YAML or code)
    configs         map[string]*schema.ConfigDef
    middlewares     map[string]*schema.MiddlewareDef
    services        map[string]*schema.ServiceDef
    routers         map[string]*schema.RouterDef
    routerOverrides map[string]*schema.RouterOverrideDef
    
    // Runtime instances
    routerInstances     sync.Map // map[string]router.Router
    serviceInstances    sync.Map // map[string]any
    middlewareInstances sync.Map // map[string]request.HandlerFunc
    
    // Lazy services
    lazyServiceFactories sync.Map // map[string]*LazyServiceEntry
    
    // Topology (2-Layer)
    deploymentTopologies sync.Map // map[deploymentName]*DeploymentTopology
    serverTopologies     sync.Map // map[compositeKey]*ServerTopology
}
```

---

## Registration Options

### Service Registration Options

#### WithResource
Specifies resource names for auto-router generation.

**Signature:**
```go
func WithResource(singular, plural string) RegisterServiceTypeOption
```

**Example:**
```go
deploy.WithResource("user", "users")
deploy.WithResource("person", "people")
```

---

#### WithConvention
Specifies routing convention (default: "rest").

**Signature:**
```go
func WithConvention(convention string) RegisterServiceTypeOption
```

**Example:**
```go
deploy.WithConvention("rest")
deploy.WithConvention("rpc")
```

---

#### WithDependencies
Declares service dependencies for automatic injection.

**Signature:**
```go
func WithDependencies(deps ...string) RegisterServiceTypeOption
```

**Example:**
```go
deploy.WithDependencies("db", "cache", "logger")
```

---

#### WithPathPrefix
Sets path prefix for all routes.

**Signature:**
```go
func WithPathPrefix(prefix string) RegisterServiceTypeOption
```

**Example:**
```go
deploy.WithPathPrefix("/api/v1")
deploy.WithPathPrefix("/api/v2")
```

---

#### WithMiddleware
Attaches middleware to all service routes.

**Signature:**
```go
func WithMiddleware(names ...string) RegisterServiceTypeOption
```

**Example:**
```go
deploy.WithMiddleware("auth", "logger", "rate-limiter")
```

---

#### WithRouteOverride
Customizes path for specific methods.

**Signature:**
```go
func WithRouteOverride(methodName, pathSpec string) RegisterServiceTypeOption
```

**Example:**
```go
deploy.WithRouteOverride("Login", "POST /auth/login")
deploy.WithRouteOverride("Logout", "POST /auth/logout")
```

---

#### WithHiddenMethods
Excludes methods from auto-router generation.

**Signature:**
```go
func WithHiddenMethods(methods ...string) RegisterServiceTypeOption
```

**Example:**
```go
deploy.WithHiddenMethods("InternalHelper", "validateUser")
```

---

### Middleware Registration Options

#### WithAllowOverride
Allows overriding existing middleware types.

**Signature:**
```go
func WithAllowOverride(allow bool) MiddlewareTypeOption
```

**Example:**
```go
deploy.Global().RegisterMiddlewareType("logger", loggerFactory,
    deploy.WithAllowOverride(true))
```

---

#### WithAllowOverrideForName
Allows overriding existing middleware names.

**Signature:**
```go
func WithAllowOverrideForName(allow bool) MiddlewareNameOption
```

---

### Lazy Service Registration Options

#### WithRegistrationMode
Sets registration mode for lazy services.

**Signature:**
```go
func WithRegistrationMode(mode LazyServiceMode) LazyServiceOption
```

**Modes:**
```go
const (
    LazyServicePanic    LazyServiceMode = iota // Panic if exists (default)
    LazyServiceSkip                            // Skip if exists (idempotent)
    LazyServiceOverride                        // Override existing
)
```

**Example:**
```go
deploy.WithRegistrationMode(deploy.LazyServiceSkip)
deploy.WithRegistrationMode(deploy.LazyServiceOverride)
```

---

## Loader Functions

### LoadConfig
Loads deployment configuration from YAML file(s).

**Signature:**
```go
func LoadConfig(paths ...string) (*schema.DeployConfig, error)
```

**Parameters:**
- `paths` - One or more YAML file paths

**Returns:**
- `*schema.DeployConfig` - Merged configuration
- `error` - Error if loading or validation fails

**Example:**
```go
// Single file
config, err := loader.LoadConfig("config/deployment.yaml")
if err != nil {
    log.Fatal(err)
}

// Multiple files (merged)
config, err := loader.LoadConfig(
    "config/base.yaml",
    "config/services.yaml",
    "config/production.yaml",
)
if err != nil {
    log.Fatal(err)
}
```

**Features:**
- ✅ Multi-file merging
- ✅ JSON schema validation
- ✅ Unknown field detection
- ✅ Dependency validation

---

### ValidateConfig
Validates deployment configuration against JSON schema.

**Signature:**
```go
func ValidateConfig(config *schema.DeployConfig) error
```

**Example:**
```go
config, _ := loader.LoadConfig("config/app.yaml")
if err := loader.ValidateConfig(config); err != nil {
    log.Fatal(err)
}
```

---

## Topology Management

### DeploymentTopology
Deployment-level configuration.

**Type:**
```go
type DeploymentTopology struct {
    Name            string
    ConfigOverrides map[string]any
    Servers         map[string]*ServerTopology
}
```

**Example:**
```yaml
deployments:
  production:
    config-overrides:
      log.level: INFO
      db.pool_size: 100
    servers:
      api-server:
        # ...
```

---

### ServerTopology
Server-level topology (services shared across apps).

**Type:**
```go
type ServerTopology struct {
    Name           string
    DeploymentName string
    BaseURL        string
    Services       []string          // Server-level services (shared)
    RemoteServices map[string]string // serviceName -> remoteBaseURL
    Apps           []*AppTopology
}
```

**Example:**
```yaml
servers:
  api-server:
    base-url: http://api.example.com
    apps:
      - addr: ":8080"
        routers:
          - user-router
```

---

### AppTopology
App-level topology (HTTP listener).

**Type:**
```go
type AppTopology struct {
    Addr    string
    Routers []string
}
```

---

## Complete Examples

### Basic Deployment
```yaml
# config/deployment.yaml
configs:
  app.name: "MyApp"
  app.version: "1.0.0"

service-definitions:
  user-service:
    type: user-service-factory
    depends-on:
      - db-service
  
  db-service:
    type: postgres-factory
    config:
      dsn: "postgresql://localhost/myapp"

routers:
  user-router:
    service: user-service
    convention: rest

deployments:
  production:
    servers:
      api-server:
        base-url: http://api.example.com
        apps:
          - addr: ":8080"
            routers:
              - user-router
```

**Load and Use:**
```go
package main

import (
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/core/deploy/loader"
)

func main() {
    // Load config
    config, err := loader.LoadConfig("config/deployment.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Register definitions to global registry
    for name, def := range config.ServiceDefinitions {
        deploy.Global().DefineService(def)
    }
    
    for name, def := range config.Routers {
        deploy.Global().DefineRouter(name, def)
    }
    
    // Build and run deployment
    // ... (framework handles this automatically)
}
```

---

### Multi-File Deployment
```yaml
# config/01-base.yaml
configs:
  app.name: "MyApp"

service-definitions:
  db-service:
    type: postgres-factory

# config/02-services.yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on:
      - db-service
  
  order-service:
    type: order-service-factory
    depends-on:
      - db-service
      - user-service

# config/03-production.yaml
configs:
  db.dsn: "${DATABASE_URL}"
  log.level: "INFO"

deployments:
  production:
    servers:
      api-server:
        base-url: https://api.example.com
        apps:
          - addr: ":443"
            routers:
              - user-router
              - order-router
```

**Load:**
```go
config, err := loader.LoadConfig(
    "config/01-base.yaml",
    "config/02-services.yaml",
    "config/03-production.yaml",
)
```

---

### Multi-Environment Deployment
```yaml
# config/base.yaml
service-definitions:
  user-service:
    type: user-service-factory

routers:
  user-router:
    service: user-service
    convention: rest

# config/development.yaml
deployments:
  development:
    config-overrides:
      log.level: DEBUG
    servers:
      dev-server:
        base-url: http://localhost:8080
        apps:
          - addr: ":8080"
            routers:
              - user-router

# config/production.yaml
deployments:
  production:
    config-overrides:
      log.level: INFO
    servers:
      api-server-1:
        base-url: https://api-1.example.com
        apps:
          - addr: ":443"
            routers:
              - user-router
      
      api-server-2:
        base-url: https://api-2.example.com
        apps:
          - addr: ":443"
            routers:
              - user-router
```

---

### External Service Integration
```yaml
service-definitions:
  user-service:
    type: user-service-factory

routers:
  user-router:
    service: user-service

external-service-definitions:
  payment-service:
    url: https://payment-api.example.com
    type: payment-service-factory  # Auto-creates wrapper
    resource: payment
    resource-plural: payments
    convention: rest

deployments:
  production:
    servers:
      api-server:
        base-url: https://api.example.com
        apps:
          - addr: ":443"
            routers:
              - user-router
              - payment-router  # Auto-generated from external service
```

---

### Router Overrides
```yaml
service-definitions:
  user-service:
    type: user-service-factory

routers:
  user-router:
    service: user-service
    convention: rest
    overrides: user-router-overrides

router-overrides:
  user-router-overrides:
    path-prefix: /api/v1
    middlewares:
      - auth
      - logger
    hidden:
      - InternalHelper
    custom:
      - name: Login
        method: POST
        path: /auth/login
      - name: Logout
        method: POST
        path: /auth/logout

deployments:
  production:
    servers:
      api-server:
        base-url: https://api.example.com
        apps:
          - addr: ":443"
            routers:
              - user-router
```

---

### Published Services (Shorthand)
```yaml
service-definitions:
  user-service:
    type: user-service-factory
  order-service:
    type: order-service-factory

deployments:
  production:
    servers:
      api-server:
        base-url: https://api.example.com
        # Shorthand: automatically creates routers
        addr: ":443"
        published-services:
          - user-service
          - order-service
```

**Equivalent to:**
```yaml
routers:
  user-service-router:
    service: user-service
  order-service-router:
    service: order-service

deployments:
  production:
    servers:
      api-server:
        base-url: https://api.example.com
        apps:
          - addr: ":443"
            routers:
              - user-service-router
              - order-service-router
```

---

## Best Practices

### 1. Use Multi-File Configuration
```yaml
# ✅ Good: Separate concerns
config/
  ├── 01-base.yaml       # Base config
  ├── 02-services.yaml   # Service definitions
  ├── 03-routers.yaml    # Router definitions
  └── 04-production.yaml # Environment-specific

# 🚫 Avoid: Everything in one file
config/
  └── monolith.yaml      # 500+ lines
```

---

### 2. Use Config Overrides per Environment
```yaml
# ✅ Good: Environment-specific overrides
deployments:
  production:
    config-overrides:
      log.level: INFO
      db.pool_size: 100
  development:
    config-overrides:
      log.level: DEBUG
      db.pool_size: 10

# 🚫 Avoid: Hardcoded values
service-definitions:
  db-service:
    config:
      pool_size: 10  # Same for all environments
```

---

### 3. Validate Configuration Early
```go
// ✅ Good: Validate on load
config, err := loader.LoadConfig("config/app.yaml")
if err != nil {
    log.Fatalf("Config validation failed: %v", err)
}

// 🚫 Avoid: No validation
config := loadYAML("config/app.yaml") // May have errors
```

---

### 4. Use External Services for Third-Party APIs
```yaml
# ✅ Good: External service definitions
external-service-definitions:
  stripe-api:
    url: https://api.stripe.com
    type: stripe-client-factory

# 🚫 Avoid: Mixing with local services
service-definitions:
  stripe-api:  # This is external, not local!
    type: stripe-client-factory
```

---

### 5. Document Dependencies
```yaml
# ✅ Good: Clear dependencies
service-definitions:
  order-service:
    type: order-service-factory
    depends-on:
      - user-service
      - payment-service
      - inventory-service

# 🚫 Avoid: Hidden dependencies
service-definitions:
  order-service:
    type: order-service-factory
    # Dependencies not declared
```

---

## See Also

- **[Config](./config.md)** - Configuration management
- **[Schema](./schema.md)** - YAML schema definitions
- **[lokstra_registry](../02-registry/lokstra_registry.md)** - Registry API
- **[Service Registration](../02-registry/service-registration.md)** - Service patterns

---

## Related Guides

- **[Deployment Essentials](../../01-essentials/05-deployment/)** - Deployment basics
- **[Multi-Environment Setup](../../04-guides/multi-environment/)** - Environment strategies
- **[Microservices Architecture](../../04-guides/microservices/)** - Distributed deployment
