# 05-config - YAML Configuration

This example demonstrates how to use YAML configuration files with Lokstra's configuration system.

## What You'll Learn

- ✅ The 4 main sections of Lokstra YAML config
- ✅ Registering middleware factories in code
- ✅ Registering routers in code (not in YAML)
- ✅ Loading configuration from YAML
- ✅ Using lokstra_registry to wire everything
- ✅ Applying named middleware to routes
- ✅ Environment variables in YAML config

## Files

- `config.yaml` - YAML configuration with 4 sections
- `main.go` - Application entry point
- `setup_routers.go` - Router registration
- `handlers.go` - HTTP handlers
- `test.http` - Test requests

## YAML Configuration Structure

Lokstra config has **4 main sections**:

### 1. `configs` - General Key-Value Configuration

Accessible via `lokstra_registry.GetConfig(name, default)`

```yaml
configs:
  - name: app-name
    value: "My App"
  - name: environment
    value: ${ENV:development}  # Environment variable with default
```

### 2. `services` - Lazy-Loaded Services

Services are created only when first accessed via `lokstra_registry.GetService(name)`

```yaml
services:
  - name: main_db      # Service identifier
    type: dbpool_pg    # Factory name (if empty, uses 'name')
    enable: true       # Optional, default is true
    config:
      host: ${DB_HOST:localhost}
      port: ${DB_PORT:5432}
```

**Important**: Service factories must be registered in code BEFORE loading config:

```go
lokstra_registry.RegisterServiceFactory("dbpool_pg", func(cfg map[string]any) any {
    return dbpool_pg.New(cfg)
})
```

### 3. `middlewares` - Named Middleware

Middleware can be applied via:
- `router.Use("middleware-name")`  // Global
- `router.GET(path, handler, "middleware-name")`  // Per-route

```yaml
middlewares:
  - name: cors                # Middleware identifier
    type: cors                # Factory name (if empty, uses 'name')
    enable: true              # Optional, default is true
    config:
      allow_origins:
        - "*"
```

**Important**: Middleware factories must be registered in code BEFORE loading config:

```go
lokstra_registry.RegisterMiddlewareFactory("cors", cors.MiddlewareFactory)
```

### 4. `servers` - Server Definitions

Defines servers with apps and routers. Only ONE server runs at a time.

```yaml
servers:
  - name: dev-server
    baseUrl: http://localhost
    deployment-id: development
    apps:
      - addr: ":8080"
        routers: [api-router]  # References router registered in code
```

**Important**: Routers are registered in CODE, not in YAML:

```go
router := router.New("api-router")
// ... define routes ...
lokstra_registry.RegisterRouter("api-router", router)
```

## How It Works

### Step-by-Step Flow

```go
// 1. Register factories BEFORE loading config
lokstra_registry.RegisterMiddlewareFactory("cors", cors.MiddlewareFactory)
lokstra_registry.RegisterServiceFactory("dbpool_pg", dbpool_pg.NewService)

// 2. Register routers in code
router := router.New("api-router")
router.GET("/users", handler, "cors")  // Use named middleware!
lokstra_registry.RegisterRouter("api-router", router)

// 3. Load YAML config
cfg := config.New()
config.LoadConfigFile("config.yaml", cfg)

// 4. Register config (processes all 4 sections)
lokstra_registry.RegisterConfig(cfg)

// 5. Select and start server
lokstra_registry.SetCurrentServerName("dev-server")
lokstra_registry.StartServer()
```

## Running the Example

```bash
cd cmd/learning/01-basics/05-config
go run .
```

The server will start on `http://localhost:8080`

## Testing Endpoints

Use the REST Client extension with `test.http`:

```http
### Health Check
GET http://localhost:8080/health

### List Users (with CORS + Request Logger middleware)
GET http://localhost:8080/users

### Create User
POST http://localhost:8080/users
Content-Type: application/json

{
  "name": "Alice",
  "email": "alice@example.com",
  "age": 28
}
```

## Key Concepts

### Named Middleware

Middleware defined in YAML can be used by name:

```go
// Apply globally to all routes
router.Use("cors", "request_logger")

// Apply per-route
router.GET("/users", handler, "cors", "request_logger")
router.POST("/users", handler, "cors", "recovery")
```

### Route Options

Routes can have additional options:

```go
router.GET("/admin/stats", handler,
    "cors",                                          // Named middleware
    route.WithDescriptionOption("Admin statistics"), // Description
    route.WithOverrideParentMwOption(true),         // Override parent middleware
)
```

Available options:
- `route.WithNameOption(name)` - Set route name
- `route.WithDescriptionOption(desc)` - Set description
- `route.WithOverrideParentMwOption(bool)` - Override parent middleware

### Environment Variables

YAML supports environment variables with defaults:

```yaml
configs:
  - name: db-host
    value: ${DB_HOST:localhost}  # Use DB_HOST env var, default to "localhost"
  
  - name: db-port
    value: ${DB_PORT:5432}       # Number default
```

Set environment variables:

```bash
# Windows PowerShell
$env:DB_HOST="mydb.example.com"
$env:DB_PORT="5433"

# Linux/Mac
export DB_HOST=mydb.example.com
export DB_PORT=5433
```

## Scalable Deployment

The same binary can run in different modes using different server configurations:

```yaml
servers:
  # Development - single port
  - name: dev-server
    deployment-id: development
    apps:
      - addr: ":8080"
        routers: [api-router, admin-router]

  # Production - separate services
  - name: api-service
    deployment-id: production
    apps:
      - addr: ":8080"
        routers: [api-router]

  - name: admin-service  
    deployment-id: production
    apps:
      - addr: ":8081"
        routers: [admin-router]
```

Select server at runtime:

```go
serverName := os.Getenv("SERVER_NAME")
if serverName == "" {
    serverName = "dev-server"
}
lokstra_registry.SetCurrentServerName(serverName)
lokstra_registry.StartServer()
```

## Comparison: Manual vs YAML Config

### Manual Configuration (Previous Examples)

```go
// Everything in code
router := lokstra.NewRouter("api")
router.Use(cors.Middleware([]string{"*"}))
router.Use(request_logger.Middleware(nil))

app := lokstra.NewApp("demo", ":8080", router)
app.Run(30 * time.Second)
```

**Pros:**
- Type-safe
- No YAML parsing
- Simple for learning

**Cons:**
- Configuration hardcoded
- Need recompilation for changes
- Not scalable for multiple environments

### YAML Configuration (This Example)

```yaml
# config.yaml
middlewares:
  - name: cors
    type: cors
    config:
      allow_origins: ["*"]

servers:
  - name: dev-server
    apps:
      - addr: ":8080"
        routers: [api-router]
```

```go
// main.go - minimal code
lokstra_registry.RegisterMiddlewareFactory("cors", cors.MiddlewareFactory)
setupRouters()
cfg := config.New()
config.LoadConfigFile("config.yaml", cfg)
lokstra_registry.RegisterConfig(cfg)
lokstra_registry.SetCurrentServerName("dev-server")
lokstra_registry.StartServer()
```

**Pros:**
- Configuration separated from code
- No recompilation for config changes
- Multiple environments (dev/staging/prod)
- Scalable (monolith → microservices)
- Environment variables support

**Cons:**
- Runtime errors if config invalid
- Need to register factories in code

## When to Use YAML Config?

✅ **Use YAML Config when:**
- Multiple deployment environments
- Configuration changes frequently
- Microservices architecture
- Team needs to modify config without code changes
- CI/CD pipelines

❌ **Use Manual Config when:**
- Simple single-server app
- Maximum type safety needed
- Configuration rarely changes
- Quick prototyping/learning

## Advanced Example

For a complete scalable deployment example (monolith → microservices), see:

📁 `/cmd/examples/13-router-integration`

That example demonstrates:
- Single binary running 3 different deployment scenarios
- Automatic service discovery
- Zero-config inter-service communication
- Scaling from monolith to microservices without code changes

## Next Steps

- **02-services**: Deep dive into service patterns (factory, lazy, registry, contracts)
- **03-middleware**: Middleware patterns (basic, factory, chaining, parent override)
- **04-yaml-config**: Advanced YAML patterns for scalable deployment
- **05-advanced**: API client, response patterns, error handling

## See Also

- [YAML Configuration System](../../../../docs/yaml-configuration-system.md)
- [Service Registry Pattern](../../../../docs/lokstra-registry.md)
- [Configuration Strategies](../../../../docs/configuration-strategies.md)
- [JSON Schema Implementation](../../../../docs/json-schema-implementation.md)
- [Example: 13-router-integration](../../../examples/13-router-integration/)


## Files

- `config.yaml` - Application configuration (server, services, middleware)
- `main.go` - Application entry point with service/middleware registration
- `handlers.go` - HTTP handlers that use registered services
- `test.http` - Test requests for all endpoints

## Configuration Structure

```yaml
server:
  name: "config-demo"
  port: ":8080"
  shutdown_timeout: 30

services:
  - name: "main_db"
    type: "dbpool_pg"
    config:
      host: "localhost"
      port: 5432
      database: "lokstra_demo"
      user: "postgres"
      password: "postgres"
      max_connections: 20
      min_connections: 2

middleware:
  - name: "cors"
    type: "cors"
    config:
      allowed_origins:
        - "*"
      allowed_methods:
        - "GET"
        - "POST"
        - "PUT"
        - "DELETE"
      allowed_headers:
        - "Content-Type"
        - "Authorization"
```

## Key Concepts

### 1. Service Registration

Services must be registered in code before loading YAML:

```go
// Register service factory
lokstra_registry.RegisterServiceFactory("dbpool_pg", func(name string, cfg map[string]any) (any, error) {
    // Create and return service instance
    return dbpool_pg.NewService(name, cfg)
})
```

### 2. Middleware Registration

Middleware must be registered in code:

```go
// Register middleware factory
lokstra_registry.RegisterMiddlewareFactory("cors", func(cfg map[string]any) (lokstra.MiddlewareFunc, error) {
    // Create and return middleware
    return cors.New(cfg), nil
})
```

### 3. Loading Configuration

```go
// Load config from YAML
config, err := lokstra_registry.LoadConfigFromFile("config.yaml")
if err != nil {
    panic(err)
}

// Run server with config
lokstra_registry.RunServer(config)
```

## Running the Example

### Step 1: Set up PostgreSQL (Optional)

If you want to test database connection:

```bash
# Using Docker
docker run --name postgres-demo -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:15

# Create database
docker exec -it postgres-demo psql -U postgres -c "CREATE DATABASE lokstra_demo;"
```

Or edit `config.yaml` to skip database service.

### Step 2: Run the server

```bash
cd cmd/learning/01-basics/05-config
go run .
```

### Step 3: Test endpoints

Use the REST Client extension with `test.http` or curl:

```bash
# Health check
curl http://localhost:8080/health

# Get users (uses database service)
curl http://localhost:8080/api/users

# Create user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

## YAML Schema Validation

The configuration file supports YAML Language Server for validation and autocomplete.

Add to `config.yaml`:

```yaml
# yaml-language-server: $schema=../../lokstra.json
```

This provides:
- ✅ Autocomplete for services and middleware
- ✅ Validation of configuration structure
- ✅ Inline documentation
- ✅ Error detection

## Configuration Best Practices

### 1. Environment Variables

Use environment variables for sensitive data:

```yaml
services:
  - name: "main_db"
    type: "dbpool_pg"
    config:
      host: "${DB_HOST:localhost}"
      port: ${DB_PORT:5432}
      password: "${DB_PASSWORD}"
```

### 2. Separate Configs

Different configs for different environments:

```
config.dev.yaml
config.staging.yaml
config.production.yaml
```

Load based on environment:

```go
env := os.Getenv("ENV")
if env == "" {
    env = "dev"
}

config, err := lokstra_registry.LoadConfigFromFile(fmt.Sprintf("config.%s.yaml", env))
```

### 3. Config Validation

Validate configuration after loading:

```go
config, err := lokstra_registry.LoadConfigFromFile("config.yaml")
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// Validate required services
if !config.HasService("main_db") {
    log.Fatal("main_db service is required")
}
```

## Comparison: Manual vs YAML Config

### Manual Configuration (Previous Examples)

```go
// main.go
router := lokstra.NewRouter("demo")
router.Use(cors.New())

app := lokstra.NewApp("demo", ":8080", router)

// Register services manually
dbService := dbpool_pg.New(dbConfig)
app.RegisterService("main_db", dbService)

app.Run()
```

**Pros:**
- Full control
- Type-safe
- Good for simple apps

**Cons:**
- Hardcoded configuration
- Need recompilation for changes
- Not scalable for multiple environments

### YAML Configuration (This Example)

```yaml
# config.yaml
server:
  port: ":8080"

services:
  - name: "main_db"
    type: "dbpool_pg"
    config: {...}
```

```go
// main.go
lokstra_registry.RegisterServiceFactory("dbpool_pg", dbpool_pg.NewService)
config, _ := lokstra_registry.LoadConfigFromFile("config.yaml")
lokstra_registry.RunServer(config)
```

**Pros:**
- Configuration separated from code
- Easy to change without recompiling
- Multiple environments support
- Scalable for microservices
- Schema validation

**Cons:**
- Less type-safe (runtime errors)
- Need to register factories

## When to Use YAML Config?

✅ **Use YAML Config when:**
- Multiple deployment environments (dev/staging/prod)
- Configuration needs to change without recompiling
- Microservices architecture
- Team collaboration (config changes without code changes)
- CI/CD pipelines

❌ **Use Manual Config when:**
- Simple single-server app
- Need maximum type safety
- Configuration rarely changes
- Learning/prototyping

## Next Steps

- **02-services**: Deep dive into service patterns (factory, lazy, contracts)
- **03-middleware**: Middleware patterns (basic, factory, chaining)
- **04-yaml-config**: Advanced YAML patterns for microservices
- **05-advanced**: API client, response patterns, error handling

## See Also

- [YAML Configuration System](../../../../docs/yaml-configuration-system.md)
- [Service Registry Pattern](../../../../docs/lokstra-registry.md)
- [Configuration Strategies](../../../../docs/configuration-strategies.md)
- [JSON Schema Implementation](../../../../docs/json-schema-implementation.md)
