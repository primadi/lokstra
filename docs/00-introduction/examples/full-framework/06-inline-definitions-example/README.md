# Inline Definitions Example

**Runnable demonstration of the 3-Level Inline Definitions Architecture in Lokstra**

This example provides a complete, working project that demonstrates how inline definitions work at three levels: Global, Deployment, and Server.

---

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Architecture Overview](#architecture-overview)
- [Project Structure](#project-structure)
- [Configuration Explained](#configuration-explained)
- [Normalization Process](#normalization-process)
- [Testing the Example](#testing-the-example)
- [Use Cases](#use-cases)

---

## 🚀 Quick Start

### Prerequisites
- Go 1.23 or higher
- VS Code with REST Client extension (optional, for test.http)

### Run the Example

```bash
# From the example directory
cd docs/00-introduction/examples/full-framework/06-inline-definitions-example

# Download dependencies
go mod download

# Run the development server
go run .
```

Server will start on `http://localhost:3000`

### Test the APIs

Open `test.http` in VS Code and click **"Send Request"**:

- `GET /api/users` - User service (from deployment-level inline)
- `GET /api/products` - Product service (from server-level inline)
- `GET /api/orders` - Order service (from global definitions)

Or use curl:
```bash
curl http://localhost:3000/api/users
curl http://localhost:3000/api/products
curl http://localhost:3000/api/orders
```

---

## 🏗 Architecture Overview

Lokstra supports inline definitions at three levels with automatic normalization:

### 1. Global Level
Shared across **all deployments** (development, production, etc.)

```yaml
# Accessible everywhere
middleware-definitions:
  cors-global: 
    type: cors
    config: { ... }

service-definitions:
  order-service:
    type: order-service-type
```

### 2. Deployment Level
Shared across **all servers** in a specific deployment

```yaml
deployments:
  development:
    # Inline definitions here
    inline-middleware-definitions:
      dev-logger:              # → normalized to: development.dev-logger
        type: logging
        config: { prefix: "DEV" }
    
    inline-service-definitions:
      user-service:            # → normalized to: development.user-service
        type: user-service-type
```

### 3. Server Level
Specific to **one server** only

```yaml
deployments:
  development:
    servers:
      dev-server:
        # Inline definitions here
        inline-service-definitions:
          product-service:     # → normalized to: development.dev-server.product-service
            type: product-service-type
```

### Normalization Strategy

When a server runs, inline definitions are normalized with prefixes:

- **Deployment-level**: `{deployment}.{name}`
- **Server-level**: `{deployment}.{server}.{name}`

This prevents naming conflicts across deployments/servers.

### Priority Resolution

When multiple definitions have the same name:

1. **Server-level** (most specific) - wins
2. **Deployment-level** - used if not found at server level
3. **Global** (most general) - used if not found at deployment/server level

---

## 📁 Project Structure

```
06-inline-definitions-example/
├── main.go              # Application entry point
├── register.go          # Service type and middleware registrations
├── config.yaml          # Configuration with inline definitions
├── test.http            # API tests (for VS Code REST Client)
├── go.mod               # Go module configuration
├── .gitignore           # Git ignore file
└── README.md            # This file
```

### Key Files

**main.go**: Simple entry point that loads config and runs server
```go
lokstra_registry.LoadAndBuild([]string{"config.yaml"})
lokstra_registry.RunServer("development.dev-server", 30*time.Second)
```

**register.go**: Registers service factories and middleware
- Mock implementations of UserService, ProductService, OrderService
- Logging, Auth, and CORS middleware

**config.yaml**: Demonstrates all 3 levels of inline definitions
- Global: order-service, cors-global
- Deployment: user-service, dev-logger
- Server: product-service, api-logger

---

## ⚙️ Configuration Explained

See `config.yaml` for the complete configuration. Here's how it's organized:

### Global Definitions

```yaml
# Available to ALL deployments
middleware-definitions:
  cors-global:
    type: cors
    config: { origin: "*" }

service-definitions:
  order-service:
    type: order-service-type
```

### Deployment-Level Inline

```yaml
deployments:
  development:
    # Shared by all servers in 'development'
    inline-middleware-definitions:
      dev-logger:
        type: logging
        config: { prefix: "DEV" }
    
    inline-service-definitions:
      user-service:
        type: user-service-type
        middlewares: ["dev-logger"]  # References inline middleware
```

**After normalization**:
- `dev-logger` → `development.dev-logger`
- `user-service` → `development.user-service`

### Server-Level Inline

```yaml
deployments:
  development:
    servers:
      dev-server:
        # Specific to dev-server only
        inline-middleware-definitions:
          api-logger:
            type: logging
            config: { prefix: "API" }
        
        inline-service-definitions:
          product-service:
            type: product-service-type
            middlewares: ["api-logger"]  # References server-level inline
```

**After normalization**:
- `api-logger` → `development.dev-server.api-logger`
- `product-service` → `development.dev-server.product-service`

### Reference Normalization

All references in `depends-on` and `middlewares` are automatically updated:

```yaml
inline-service-definitions:
  user-service:
    middlewares: ["dev-logger"]     # Short name (relative reference)
    # After normalization becomes: ["development.dev-logger"]
```

---

## 🔄 Normalization Process

### When It Happens

Normalization is **lazy** - it only happens when you run a specific server:

```go
lokstra_registry.RunServer("development.dev-server", timeout)
```

### What Happens

1. **Extract deployment and server names**: `development`, `dev-server`

2. **Move inline definitions to global** with normalized names:
   - Deployment-level `user-service` → Global `development.user-service`
   - Server-level `product-service` → Global `development.dev-server.product-service`

3. **Build renaming map**:
   ```go
   {
     "dev-logger": "development.dev-logger",
     "user-service": "development.user-service",
     "api-logger": "development.dev-server.api-logger",
     "product-service": "development.dev-server.product-service",
   }
   ```

4. **Update all references** in service/router/external-service definitions:
   - `depends-on: ["user-service"]` → `depends-on: ["development.user-service"]`
   - `middlewares: ["dev-logger"]` → `middlewares: ["development.dev-logger"]`

5. **Register normalized definitions** to GlobalRegistry

6. **Lazy service registration** proceeds with normalized names

### Console Output

When you run the server, you'll see:
```
📝 Normalized and registered inline definitions for server development.dev-server
```

---

## 🧪 Testing the Example

### 1. Run Development Server

```bash
go run .
```

Expected output:
```
📝 Normalized and registered inline definitions for server development.dev-server
┌─────────────────────────────────────────────┐
│ Server: dev-server                          │
├─────────────────────────────────────────────┤
│ Deployment: development                     │
│ Base URL: http://localhost:3000             │
│                                             │
│ Services (server-level):                    │
│   • development.user-service                │
│   • development.dev-server.product-service  │
│   • order-service                           │
...
🚀 Server 'dev-server' is running on :3000
```

### 2. Test APIs with test.http

Open `test.http` in VS Code and click "Send Request" on each test.

### 3. Test with curl

```bash
# User service (from deployment-level inline)
curl http://localhost:3000/api/users
# Returns: [{"id":1,"name":"Alice"},{"id":2,"name":"Bob"},{"id":3,"name":"Charlie"}]

# Product service (from server-level inline)
curl http://localhost:3000/api/products
# Returns: [{"id":1,"name":"Laptop","price":1200},...]

# Order service (from global definitions)
curl http://localhost:3000/api/orders
# Returns: [{"id":1,"user_id":1,"total":1200,"status":"completed"},...]
```

### 4. Check Middleware Execution

Watch the console output - you'll see middleware logging:
```
[DEV] Request received      ← from deployment-level dev-logger
[API] Request received      ← from server-level api-logger
[CORS] Setting CORS headers ← from global cors-global
```

---

## 📚 Use Cases

### 1. Environment-Specific Configuration

**Problem**: Different settings for dev, staging, production

**Solution**: Use deployment-level inline definitions

```yaml
deployments:
  development:
    inline-service-definitions:
      cache:
        config: { db: 0, ttl: 60 }   # Short TTL for dev
  
  production:
    inline-service-definitions:
      cache:
        config: { db: 1, ttl: 3600 } # Long TTL for prod
```

### 2. Microservices per Server

**Problem**: Each microservice needs isolated configuration

**Solution**: Use server-level inline definitions

```yaml
deployments:
  production:
    servers:
      user-api:
        inline-service-definitions:
          user-db: { connection: "postgres://user-db..." }
      
      order-api:
        inline-service-definitions:
          order-db: { connection: "postgres://order-db..." }
```

### 3. Shared Infrastructure

**Problem**: Multiple servers share common services (auth, cache)

**Solution**: Use deployment-level inline definitions

```yaml
deployments:
  production:
    inline-service-definitions:
      redis-cache: { ... }   # Shared cache
      jwt-auth: { ... }      # Shared auth
    
    servers:
      api-1: { ... }        # Uses shared cache & auth
      api-2: { ... }        # Uses shared cache & auth
```

### 4. Override Pattern

**Problem**: Need server-specific override of deployment default

**Solution**: Use same name at both levels (server wins)

```yaml
deployments:
  production:
    inline-service-definitions:
      logger: { level: "info" }    # Default
    
    servers:
      debug-server:
        inline-service-definitions:
          logger: { level: "debug" }  # Override for this server
```

---

## 🎓 Key Benefits

### 1. No Naming Conflicts

Same name can be used at different scopes:

```yaml
deployments:
  dev:
    inline-service-definitions:
      cache: { db: 0 }              # dev.cache
    servers:
      api:
        inline-service-definitions:
          cache: { db: 1 }          # dev.api.cache (different!)
```

### 2. Clear Ownership

Definitions are located where they're used - easy to find and maintain.

### 3. Lazy Normalization

Only normalizes definitions for the server that's running - efficient.

### 4. Automatic Reference Resolution

No manual name management - framework handles it:

```yaml
inline-service-definitions:
  product-service:
    depends-on: ["cache"]          # Short name
    # Framework resolves to: development.dev-server.cache
```

### 5. Easy Testing

Different servers can run with different configs in the same YAML file.

---

## 🔧 Implementation Details

### Code Location

- **Normalization**: `core/deploy/loader/builder.go::NormalizeInlineDefinitionsForServer()`
- **Registration**: `core/deploy/loader/builder.go::RegisterDefinitionsToRegistry()`
- **Integration**: `lokstra_registry/deployment.go::RunCurrentServer()`

### Process Flow

```
LoadAndBuild(config.yaml)
  ├─ Load and parse YAML
  ├─ Store original config in GlobalRegistry
  ├─ Register GLOBAL definitions only
  └─ Build deployment topologies

RunServer("development.dev-server", timeout)
  ├─ Extract deployment="development", server="dev-server"
  ├─ NormalizeInlineDefinitionsForServer(config, deployment, server)
  │   ├─ Move inline → global with normalized names
  │   ├─ Build renaming map
  │   └─ Update all references
  ├─ RegisterDefinitionsToRegistry(registry, config)
  │   └─ Register normalized definitions
  └─ Start server with lazy service resolution
```

---

## 📝 Best Practices

1. **Use Global for Shared Services**: Production databases, external APIs
2. **Use Deployment for Environment Settings**: Dev/staging/prod differences
3. **Use Server for Isolation**: Microservice-specific configuration
4. **Prefer Short Names in References**: Let framework normalize them
5. **Document Inline Definitions**: Add comments explaining scope

---

## 🐛 Troubleshooting

### Service not found

Check the normalization output - service names are prefixed:
```
✗ "user-service" not found
✓ "development.user-service" found
```

### Middleware not applied

Ensure middleware name matches after normalization:
```yaml
middlewares: ["dev-logger"]  # Will be normalized to "development.dev-logger"
```

### Wrong definition used

Check priority: Server > Deployment > Global
- If both deployment and server define same name, server wins

---

## 📚 Learn More

- [Lokstra Documentation](https://primadi.github.io/lokstra/)
- [Configuration Guide](../../../../01-router-guide/)
- [Framework Guide](../../../../02-framework-guide/)

---

## 📄 License

This example is part of the Lokstra framework. See LICENSE file in project root.


## Architecture Overview

Lokstra supports inline definitions at three levels with automatic normalization:

1. **Global Level** - Shared across all deployments
2. **Deployment Level** - Shared across all servers in a deployment
3. **Server Level** - Specific to individual servers

### Normalization Strategy

Inline definitions are lazily normalized when a server runs:

- **Deployment-level**: `{deployment}.{name}`
- **Server-level**: `{deployment}.{server}.{name}`

### Priority Resolution

When resolving references, the system uses this priority:

1. **Server-level** (most specific)
2. **Deployment-level**  
3. **Global** (most general)

## Example Structure

```yaml
# Global definitions
middleware-definitions:
  cors-global: { ... }

deployments:
  dev:
    # Deployment-level (shared across all servers in 'dev')
    middleware-definitions:
      auth-jwt: { ... }          # → normalized to: dev.auth-jwt
    
    service-definitions:
      cache: { ... }             # → normalized to: dev.cache
    
    servers:
      api-server:
        # Server-level (specific to this server only)
        service-definitions:
          user-service: { ... }  # → normalized to: dev.api-server.user-service
        
        middleware-definitions:
          rate-limit: { ... }    # → normalized to: dev.api-server.rate-limit
```

## Use Cases

### 1. Development Environment
- **Deployment-level**: Shared JWT auth, shared Redis cache
- **Server-level**: Server-specific services and configurations
- **Benefit**: Easy to share common settings while allowing customization

### 2. Production Microservices
- **Deployment-level**: Production JWT auth, production Redis
- **Server-level**: Each microservice has its own database and services
- **External services**: Reference other microservices in the deployment

### 3. Hybrid Deployment
- Mix of monolith and microservices
- **Deployment-level**: Common infrastructure (auth, cache)
- **Server-level**: Different architectures per server

## Key Benefits

1. **No Naming Conflicts**: Same name can be used at different scopes
   ```yaml
   deployments:
     dev:
       service-definitions:
         cache: { db: 0 }        # dev.cache
       servers:
         api-server:
           service-definitions:
             cache: { db: 1 }    # dev.api-server.cache (overrides deployment-level)
   ```

2. **Clear Ownership**: Definitions are located where they're used

3. **Lazy Normalization**: Only normalizes definitions for the server that's running

4. **Flexible References**: Can reference any level
   ```yaml
   depends-on:
     - cors-global              # Global
     - auth-jwt                 # Deployment-level (dev.auth-jwt)
     - user-service             # Server-level (dev.api-server.user-service)
   ```

## Running the Example

```bash
# This example requires implementing the factory types referenced
# It demonstrates the configuration structure only

# The normalization happens automatically when you run:
lokstra_registry.RunServer("dev.api-server", 30*time.Second)

# You'll see logs like:
# 📝 Normalized 3 inline definition(s) for server dev.api-server
```

## Implementation Notes

- Normalization is performed in `loader.NormalizeInlineDefinitionsForServer()`
- Called just before `registerLazyServicesForServer()` in `RunCurrentServer()`
- Only normalizes for the specific deployment and server being run
- Original config is preserved in `GlobalRegistry.deployConfig`
