# YAML Configuration Example

This example demonstrates Lokstra's powerful YAML configuration system that allows you to define your entire application architecture declaratively.

## What This Example Shows

1. **Single File Loading** - Load configuration from a single YAML file
2. **Multi-File Loading** - Load and merge configuration from multiple YAML files
3. **Configuration Validation** - Automatic validation of configuration structure and references
4. **Selective Application** - Apply specific parts of configuration (routers, services, etc.)
5. **Complete Application** - Apply entire configuration for a server
6. **Default Values** - Demonstrate how default values work

## Running the Example

```bash
cd cmd/examples/13-yaml-config
go run main.go
```

## Key Features Demonstrated

### 1. Configuration Structure
- **Routers**: Define HTTP routes and middleware chains
- **Services**: Configure dependency injection services
- **Middlewares**: Define reusable middleware components  
- **Servers**: Configure server instances and applications

### 2. Multi-File Support
Configuration can be split across multiple files:
```
config/
├── routers.yaml     # Route definitions
├── services.yaml    # Service configurations
├── middlewares.yaml # Middleware definitions
└── servers.yaml     # Server configurations
```

### 3. Validation
Automatic validation ensures:
- No duplicate names within sections
- All references between sections are valid
- Required fields are present
- Configuration is internally consistent

### 4. Default Values
Smart defaults reduce configuration verbosity:
- Routes default to GET method
- Components default to enabled
- Engine types default to "default"
- Override flags default to false

### 5. Flexible Application
Apply configuration flexibly:
```go
// Apply specific components
config.ApplyRoutersConfig(&cfg, "api-router")
config.ApplyServicesConfig(&cfg, "database", "cache")

// Apply complete server configuration  
config.ApplyAllConfig(&cfg, "web-server")
```

## Configuration Examples

### Simple Router Configuration
```yaml
routers:
  - name: api-router
    use: [cors-mw]
    routes:
      - name: health
        path: /health
        handler: HealthHandler
      - name: users
        path: /users
        method: GET
        use: [auth-mw]
        handler: GetUsersHandler
```

### Service Configuration
```yaml
services:
  - name: database
    type: postgres
    config:
      dsn: "postgres://localhost/myapp"
      max_connections: 25
      
  - name: cache
    type: redis
    config:
      addr: "localhost:6379"
      db: 0
```

### Complete Server Configuration
```yaml
servers:
  - name: web-server
    description: "Main application server"
    services: [database, cache]
    apps:
      - name: api-app
        addr: ":8080"
        routers: [api-router]
        reverse-proxies:
          - path: /external
            target: "http://external-service:8080"
```

## Integration with Main Application

In your main.go:
```go
func main() {
    var cfg config.Config
    
    // Load configuration
    if err := config.LoadConfigDir("./config", &cfg); err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // Apply complete configuration
    if err := config.ApplyAllConfig(&cfg, "web-server"); err != nil {
        log.Fatal("Failed to apply config:", err)
    }
    
    // Server is now configured and ready to start
    fmt.Println("Server configured successfully!")
}
```

## Benefits

1. **Declarative**: Define what you want, not how to build it
2. **Maintainable**: Easy to modify without changing code
3. **Environment-Friendly**: Different configs for dev/staging/prod
4. **Validation**: Catch configuration errors early
5. **Flexible**: Mix and match components as needed

See `docs/yaml-config.md` for complete documentation.