# YAML Configuration Deep Dive

> **Master YAML-based configuration for scalable deployments**  
> **Time**: 60-75 minutes • **Level**: Advanced • **Prerequisites**: [Service](../02-service/), [Middleware](../03-middleware/)

---

## 🎯 What You'll Learn

- YAML configuration structure and hierarchy
- Deployment and server topology
- Service definitions and dependencies
- Router auto-generation and customization
- Middleware configuration patterns
- Handler configurations (reverse proxy, SPA, static files)
- Environment-specific overrides
- Multi-file configuration merging

---

## 📚 Topics

### 1. Configuration Structure
Understand the YAML schema:
- Root-level sections
- Deployment hierarchy
- Server and app topology
- Inline vs global definitions

### 2. Service Definitions
Define and configure services:
- Service types and factories
- Dependency injection via YAML
- Service configuration options
- Remote vs local services

### 3. Router Auto-Generation
Auto-generate routers from services:
- Convention-based routing (REST, RPC, GraphQL)
- Resource naming (singular/plural)
- Path prefixes and rewrites
- Route hiding and customization

### 4. Middleware Configuration
Configure middleware declaratively:
- Global middleware definitions
- Router-level middleware
- Route-level middleware
- Middleware parameters and config

### 5. Handler Configurations
Mount handlers at app level:
- Reverse proxy configuration
- SPA mounting
- Static file serving
- Path stripping and rewriting

### 6. Deployment Topologies
Design multi-deployment setups:
- Development, staging, production
- Service location registry
- Config overrides per deployment
- Server grouping strategies

### 7. Advanced Patterns
Master complex configurations:
- Multi-file merging
- Environment variable resolution
- Config inheritance
- Inline definitions and normalization

### 8. Database Pools
Configure named database pools:
- DSN vs component-based config
- Pool parameters (min/max connections)
- Schema configuration
- SSL mode and timeouts

---

## 📂 Configuration Schema

### Basic Structure

```yaml
# Global configuration values
configs:
  app:
    name: "MyApp"
    version: "1.0.0"
  database:
    dsn: "postgres://localhost/mydb"

# Named database pools
dbpool-definitions:
  main-db:
    dsn: "postgres://localhost:5432/mydb"
    schema: "public"
    min-conns: 2
    max-conns: 10

# Middleware definitions
middleware-definitions:
  cors:
    type: cors
    config:
      allowed-origins: ["*"]

# Service definitions
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [user-repository]
    router:
      convention: rest
      resource: user
      path-prefix: /api/v1

# Router definitions
router-definitions:
  user-service-router:
    convention: rest
    resource: user
    middlewares: [cors]

# External service definitions
external-service-definitions:
  payment-api:
    url: "https://api.payment.com"
    type: payment-client-factory

# Deployments
deployments:
  development:
    config-overrides:
      app:
        debug: true
    
    servers:
      api:
        base-url: "http://localhost"
        apps:
          - addr: ":8080"
            published-services: [user-service]
            
            # Handler configurations
            reverse-proxies:
              - prefix: "/api/v2"
                strip-prefix: true
                target: "http://backend:9000"
            
            mount-spa:
              - prefix: "/admin"
                dir: "./dist/admin"
            
            mount-static:
              - prefix: "/assets"
                dir: "./public/assets"
```

---

## 🔧 Key Features

### 1. Service Auto-Discovery
Services with `@EndpointService` annotation are automatically discovered and registered.

### 2. Router Auto-Generation
Routers are auto-generated from published services based on metadata.

### 3. Dependency Injection
Dependencies are resolved automatically from `depends-on` declarations.

### 4. Remote Service Resolution
Services published on other servers are automatically configured as remote.

### 5. Config Variable Resolution
Use `${config.key}` to reference config values anywhere.

### 6. Case-Insensitive Lookups
All names (services, routers, middleware) are case-insensitive.

### 7. Inline Definitions
Define services/routers inline at deployment or server level for scoped usage.

### 8. Handler Mounting
Mount reverse proxies, SPAs, and static files directly in YAML.

---

## 📖 Configuration Sections

### Global Configs
```yaml
configs:
  app:
    name: "MyApp"
    environment: "production"
  database:
    host: "localhost"
    port: 5432
```

Reference: `${app.name}`, `${database.host}`

### Named DB Pools
```yaml
dbpool-definitions:
  main-db:
    host: "localhost"
    port: 5432
    database: "mydb"
    username: "user"
    password: "pass"
    schema: "public"
    min-conns: 2
    max-conns: 10
    max-idle-time: "30m"
    max-lifetime: "1h"
    sslmode: "disable"
```

Or use DSN directly:
```yaml
dbpool-definitions:
  main-db:
    dsn: "postgres://user:pass@localhost:5432/mydb"
    schema: "public"
```

### Service Definitions
```yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on:
      - userRepo:user-repository
      - logger:app-logger
    config:
      cache-enabled: true
      cache-ttl: "5m"
    router:
      convention: rest
      resource: user
      path-prefix: /api/v1
      middlewares: [auth, rate-limit]
```

### Router Definitions
```yaml
router-definitions:
  user-service-router:
    path-prefix: /api/v1
    middlewares: [cors, auth]
    custom:
      - name: GetByEmail
        method: GET
        path: /by-email/{email}
        middlewares: [rate-limit]
```

### Middleware Definitions
```yaml
middleware-definitions:
  cors:
    type: cors
    config:
      allowed-origins: ["https://app.com"]
      allowed-methods: [GET, POST, PUT, DELETE]
  
  rate-limit:
    type: rate-limiter
    config:
      requests-per-second: 100
      burst: 50
```

### Deployments
```yaml
deployments:
  production:
    config-overrides:
      app:
        debug: false
    
    servers:
      api-server:
        base-url: "https://api.myapp.com"
        
        # Inline definitions (scoped to this server)
        middleware-definitions:
          custom-auth:
            type: jwt-auth
            config:
              secret: "${JWT_SECRET}"
        
        apps:
          - addr: ":8080"
            routers: [user-service-router]
            published-services: [order-service]
```

---

## 🚀 Handler Configurations

### Reverse Proxy
```yaml
apps:
  - addr: ":8080"
    reverse-proxies:
      # Simple proxy with prefix stripping
      - prefix: "/api"
        strip-prefix: true
        target: "http://backend:9000"
      
      # Proxy with path rewriting
      - prefix: "/graphql"
        target: "http://graphql-server:4000"
        rewrite:
          from: "^/graphql"
          to: "/api/graphql"
```

**Use Cases:**
- API Gateway pattern
- Backend-for-Frontend (BFF)
- Legacy system integration
- Microservice routing

### SPA Mounting
```yaml
apps:
  - addr: ":8080"
    mount-spa:
      # Admin dashboard
      - prefix: "/admin"
        dir: "./dist/admin-spa"
      
      # Main app at root
      - prefix: "/"
        dir: "./dist/main-app"
```

**Behavior:**
- Routes without file extension → serve `index.html`
- Static files (`.js`, `.css`, `.png`) → serve directly
- 404 for missing files

### Static File Serving
```yaml
apps:
  - addr: ":8080"
    mount-static:
      # Public assets
      - prefix: "/assets"
        dir: "./public/assets"
      
      # Download files
      - prefix: "/downloads"
        dir: "./public/downloads"
```

**Behavior:**
- Paths without extension → append `/index.html`
- Static files → serve directly
- 404 for missing files

---

## 🔄 Configuration Loading

### Single File
```go
lokstra_registry.RunServerFromConfig("config.yaml")
```

### Multiple Files (Merged)
```go
lokstra_registry.RunServerFromConfig(
    "config/base.yaml",
    "config/services.yaml",
    "config/production.yaml",
)
```

### From Folder
```go
lokstra_registry.RunServerFromConfigFolder("config")
```

**Merge Strategy:**
- Later files override earlier ones
- Arrays are replaced (not merged)
- Maps are deep merged

---

## 🎯 Best Practices

### 1. Organize by Concern
```
config/
├── base.yaml          # Global configs
├── database.yaml      # DB pools
├── services.yaml      # Service definitions
├── middleware.yaml    # Middleware definitions
└── deployments/
    ├── dev.yaml
    ├── staging.yaml
    └── prod.yaml
```

### 2. Use Environment Variables
```yaml
configs:
  jwt:
    secret: "${JWT_SECRET}"
  database:
    password: "${DB_PASSWORD}"
```

### 3. Leverage Inline Definitions
```yaml
deployments:
  production:
    servers:
      api:
        # Server-scoped middleware
        middleware-definitions:
          prod-auth:
            type: jwt-auth
            config:
              secret: "${PROD_JWT_SECRET}"
```

### 4. Convention Over Configuration
```yaml
# Minimal config - uses conventions
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [user-repository]
    router: {}  # Auto-generates from metadata
```

### 5. Explicit When Needed
```yaml
# Explicit config - full control
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [user-repository]
    router:
      convention: rest
      resource: user
      path-prefix: /api/v1
      middlewares: [cors, auth]
      hidden: [Delete]
      custom:
        - name: GetByEmail
          path: /by-email/{email}
```

---

## 📂 Examples

All examples are in the `examples/` folder:

### [01 - Basic Configuration](examples/01-basic-config/)
Simple single-file configuration.

### [02 - Multi-File Merging](examples/02-multi-file/)
Split configuration across multiple files.

### [03 - Service Dependencies](examples/03-service-deps/)
Complex dependency injection.

### [04 - Router Customization](examples/04-router-custom/)
Advanced router configuration.

### [05 - Multi-Deployment](examples/05-multi-deployment/)
Development, staging, production setups.

### [06 - Handler Configurations](examples/06-handlers/)
Reverse proxy, SPA, and static file mounting.

### [07 - Database Pools](examples/07-db-pools/)
Named database pool configuration.

### [08 - Environment Variables](examples/08-env-vars/)
Environment-based configuration.

---

## 🚀 Quick Start

```bash
# Run any example
cd docs/02-framework-guide/04-config/examples/01-basic-config
go run main.go

# Test with provided test.http
```

---

## 📖 Prerequisites

Before diving in, make sure you understand:
- [Service basics](../02-service/)
- [Middleware basics](../03-middleware/)
- [App & Server structure](../05-app-and-server/)

---

## 🎯 Learning Path

1. **Understand structure** → Learn YAML schema
2. **Define services** → Configure dependency injection
3. **Generate routers** → Auto-generate from services
4. **Configure middleware** → Apply at different levels
5. **Mount handlers** → Add proxies, SPAs, static files
6. **Design deployments** → Multi-environment setups
7. **Merge configs** → Split across files
8. **Use env vars** → Environment-based config

---

## 💡 Key Takeaways

After completing this section:
- ✅ You'll design scalable YAML configurations
- ✅ You'll auto-generate routers from services
- ✅ You'll configure dependency injection declaratively
- ✅ You'll mount handlers (proxies, SPAs, static)
- ✅ You'll manage multi-deployment setups
- ✅ You'll split configs across files
- ✅ You'll use environment variables effectively
- ✅ You'll leverage conventions with explicit overrides

---

## 🔗 Related Topics

- [Service Deep Dive](../02-service/) - Service architecture patterns
- [Middleware Deep Dive](../03-middleware/) - Custom middleware creation
- [App & Server Deep Dive](../05-app-and-server/) - Production deployment

---

**Next**: [App & Server Deep Dive](../05-app-and-server/) →
