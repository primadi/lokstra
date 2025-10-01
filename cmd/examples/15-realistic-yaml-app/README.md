# Realistic YAML Configuration Example

This example demonstrates a realistic application setup using Lokstra's YAML configuration system with code-first component registration.

## Project Structure

```
15-realistic-yaml-app/
├── main.go                    # Main application with embedded fallback config
├── app-config/                # Multi-file production config
│   ├── base.yaml             # Shared middlewares and services
│   ├── routers.yaml          # API and admin route configurations
│   └── servers.yaml          # Server and app bindings
├── config-dev/               # Development environment config
│   └── development.yaml      # Dev-specific overrides
└── README.md                # This file
```

## How It Works

### 1. Code-First Registration
The application first registers all components in code:
- Middleware factories (logger, auth, cors)
- Service factories (database, cache)
- Routers with base routes
- Servers with apps

### 2. YAML Configuration Loading
Configuration is loaded from:
- **Directory**: Multiple YAML files merged together
- **Single File**: All config in one file
- **Embedded Fallback**: Hardcoded config if no files found

### 3. Component Modification
YAML config modifies existing registry entries:
- Add middleware to routers
- Add new routes to existing routers
- Configure services with environment-specific settings
- Bind services to servers

## Running the Example

### With Multi-file Configuration
```bash
cd cmd/examples/15-realistic-yaml-app
go run main.go ./app-config
```

### With Single File Configuration
```bash
go run main.go ./config-dev/development.yaml
```

### With Embedded Configuration (no config files)
```bash
go run main.go
```

### With Non-existent Path
```bash
go run main.go ./does-not-exist
# Falls back to embedded config
```

## Expected Output

```
🚀 Lokstra Realistic YAML Config Demo
=====================================

1. Setting up factories...
✅ Factories registered

2. Setting up routers...
✅ API router registered
✅ Admin router registered

3. Setting up servers...
✅ Main server registered

4. Loading configuration from: ./app-config
📊 Connecting to database: localhost/lokstra_app
🔄 Connecting to cache: localhost:6379
✅ Configuration applied successfully

🎯 Starting configured server...
✅ Server ready to start
```

## Configuration Features

### Environment Variables
```yaml
services:
  - name: main-db
    type: database
    config:
      host: "${DB_HOST:localhost}"        # With default
      password: "${DB_PASSWORD}"          # Required
```

### Route Modifications
```yaml
routers:
  - name: api
    use: [api-logger, api-cors]  # Add middleware
    routes:
      - name: health-check
        path: ""                 # Empty = use existing path from code
        method: GET
      
      - name: new-endpoint
        path: /api/new           # Add new route
        method: POST
        
      - name: disabled-route
        path: /api/dangerous
        method: DELETE
        enable: false            # Disable route
```

### Multi-Environment Support
- `app-config/`: Production configuration with security
- `config-dev/`: Development configuration with debug features
- Embedded: Fallback configuration for quick testing

### Service Factories
Services are created via factory functions:
```go
lokstra_registry.RegisterServiceFactory("database", func(config map[string]any) any {
    host := config["host"].(string)
    db := config["database"].(string)
    // Create actual database connection
    return connectionInstance
})
```

### Middleware Factories
Middleware created with custom configuration:
```go
lokstra_registry.RegisterMiddlewareFactory("logger", func(config map[string]any) request.HandlerFunc {
    level := config["level"].(string)
    return func(c *request.Context) error {
        // Custom logging logic
        return nil
    }
})
```

## Error Handling

The application will panic if:
- Router name not found in registry
- Route name not found in router
- Middleware factory not registered
- Service factory not registered
- Server name not found in registry

This fail-fast approach ensures configuration consistency at startup.

## Development vs Production

### Development Config (`config-dev/`)
- Debug logging enabled
- CORS allows all origins
- Authentication disabled for admin routes
- Dangerous operations enabled
- Different ports for dev servers

### Production Config (`app-config/`)
- Info-level logging
- Restricted CORS origins
- Authentication required
- Dangerous operations disabled
- Standard production ports

## Integration Points

This example shows how to:
1. **Register components in code** (business logic)
2. **Configure behavior via YAML** (deployment config)
3. **Support multiple environments** (dev/staging/prod)
4. **Handle missing configurations** (embedded fallback)
5. **Use environment variables** (12-factor app compliance)
6. **Validate configurations** (fail-fast on startup)