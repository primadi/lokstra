# YAML Configuration Example

This example demonstrates how to run a Lokstra application using YAML configuration files, showing how to configure servers, services, middleware, and routes declaratively.

## What You'll Learn

- **YAML Configuration Structure**: Understanding the complete configuration schema
- **Service Configuration**: Defining services with configurations in YAML
- **Middleware Setup**: Configuring middleware chains through YAML
- **Route Definitions**: Declaring HTTP routes and handlers in configuration
- **Environment Variables**: Using environment variable substitution
- **Production Patterns**: Best practices for config-driven applications

## Key Features Demonstrated

### 1. Server Configuration
```yaml
server:
  name: "YAML Configuration Demo Server"
  global_setting:
    log_level: "${LOG_LEVEL:info}"
    debug_mode: "${DEBUG_MODE:false}"
```

### 2. Service Definitions
```yaml
services:
  - name: "app-logger"
    type: "logger"
    config:
      level: "info"
      format: "json"
      output: "stdout"
```

### 3. Middleware Configuration
```yaml
middleware:
  - name: "timing"
    enabled: true
  - name: "recovery"
    enabled: true
  - name: "cors"
    enabled: true
    config:
      allow_origins: ["*"]
      allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
      allow_headers: ["*"]
```

### 4. Route Declarations
```yaml
routes:
  - method: "GET"
    path: "/"
    handler: "home"
    
  - method: "POST"
    path: "/api/data"
    handler: "data_handler"
    middleware:
      - name: "validate_content"
        enabled: true
```

## Environment Variables

The configuration supports environment variable substitution:

- `LOG_LEVEL`: Log level (default: "info")
- `DEBUG_MODE`: Debug mode flag (default: "false")
- `API_HOST`: Server host (default: "localhost")
- `API_PORT`: Server port (default: "8080")

## Available Endpoints

- `GET /` - Home endpoint with configuration info
- `GET /health` - Health check endpoint
- `GET /config-info` - Detailed configuration information
- `POST /api/data` - Data processing with validation middleware
- `GET /api/protected/profile` - Protected endpoint requiring API key

## Testing the Example

### 1. Start the Server
```bash
cd cmd/examples/02_core_concepts/06_yaml_configuration
go run main.go
```

### 2. Test Public Endpoints
```bash
# Home page
curl http://localhost:8080/

# Health check
curl http://localhost:8080/health

# Configuration info
curl http://localhost:8080/config-info
```

### 3. Test Protected Endpoint
```bash
# Without API key (should fail)
curl http://localhost:8080/api/protected/profile

# With valid API key (should succeed)
curl -H "X-API-Key: yaml-config-key-123" \
  http://localhost:8080/api/protected/profile
```

### 4. Test Data Endpoint
```bash
# Valid request with content-type
curl -X POST http://localhost:8080/api/data \
  -H "Content-Type: application/json" \
  -d '{"type":"user","content":"Test data from YAML config"}'

# Invalid request without content-type (should fail)
curl -X POST http://localhost:8080/api/data \
  -d '{"type":"user","content":"Test"}'
```

### 5. Test with Environment Variables
```bash
# Run with debug logging
LOG_LEVEL=debug go run main.go

# Run on different port
API_PORT=9090 go run main.go

# Run with custom host
API_HOST=0.0.0.0 API_PORT=8080 go run main.go
```

## Key Concepts

### Configuration-Driven Development
- **Declarative Setup**: Define infrastructure and behavior in YAML
- **Environment Flexibility**: Support multiple deployment environments
- **Centralized Management**: Single source of truth for application config
- **Version Control**: Configuration changes tracked with code

### Service Management
- **Service Definitions**: Declare services with types and configurations
- **Dependency Injection**: Services automatically injected where needed
- **Lifecycle Management**: Services managed by the framework
- **Type Safety**: Compile-time service type checking

### Middleware Orchestration
- **Chain Configuration**: Define middleware execution order
- **Conditional Enabling**: Enable/disable middleware per environment
- **Route-Specific**: Apply middleware to specific routes only
- **Parameter Configuration**: Pass configuration to middleware instances

### Route Management
- **Handler Mapping**: Map route paths to handler functions
- **Method Specification**: Define HTTP methods for each route
- **Middleware Integration**: Apply middleware at route level
- **Parameter Validation**: Automatic request parameter validation

## Production Benefits

### Deployment Advantages
- **No Code Changes**: Modify behavior through configuration
- **Environment Separation**: Different configs for dev/staging/production
- **Easy Rollbacks**: Revert configuration changes quickly
- **Infrastructure as Code**: Configuration versioned and reviewed

### Operational Benefits
- **Clear Dependencies**: Service dependencies explicit in config
- **Troubleshooting**: Configuration state visible and documented
- **Monitoring**: Built-in health checks and metrics endpoints
- **Scalability**: Easy to add/remove services and middleware

### Development Benefits
- **Reduced Complexity**: Less boilerplate code for setup
- **Testing**: Easy to test with different configurations
- **Documentation**: Configuration serves as documentation
- **Consistency**: Standardized application structure

This example showcases how YAML configuration enables robust, maintainable, and flexible Lokstra applications perfect for production environments.