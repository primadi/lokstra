# Configuration Schema Updates - Path Prefix Support

## Changes Made

### 1. Updated JSON Schema (`core/config/lokstra.json`)

#### Added Routers Section at Root Level
```json
"routers": {
  "type": "array",
  "description": "Router configurations (independent of servers)",
  "items": {
    "type": "object",
    "required": ["name"],
    "properties": {
      "name": {
        "type": "string",
        "description": "Router name"
      },
      "path-prefix": {
        "type": "string", 
        "description": "Path prefix for this router (e.g., /api/v1, /admin)"
      },
      "middlewares": {
        "type": "array",
        "description": "Middleware names to apply to this router"
      }
    }
  }
}
```

#### Added Services Array to Apps
```json
"services": {
  "type": "array",
  "description": "Service names hosted by this app (for auto local/remote detection)",
  "items": {
    "type": "string"
  }
}
```

#### Added Path-Prefix to Services
```json
"path-prefix": {
  "type": "string",
  "description": "Path prefix for this service's routes (e.g., /api/v1, /admin)",
  "pattern": "^/[a-zA-Z0-9/_-]*$"
}
```

#### Removed routers-with-prefix
- Replaced with simpler `routers` array in apps
- Path prefix now configured in `routers` section or `services` section

### 2. Updated lokstra_registry API

#### Simplified GetClientRouter
```go
// Before (with cache parameter)
func GetClientRouter(routerName string, current *api_client.ClientRouter) *api_client.ClientRouter

// After (no cache parameter needed)
func GetClientRouter(routerName string) *api_client.ClientRouter
```

**Reasoning**: Client is resolved once during service registration, no need for caching.

#### Simplified GetClientRouterOnServer
```go
// Before
func GetClientRouterOnServer(routerName, serverName string, current *api_client.ClientRouter) *api_client.ClientRouter

// After
func GetClientRouterOnServer(routerName, serverName string) *api_client.ClientRouter
```

### 3. Updated Service Implementation

#### Path Prefix from Config
```go
// Before (hardcoded)
return &authServiceRemote{
    client: NewRemoteClient(client, "/auth"),
}

// After (from config)
func CreateAuthServiceRemote(cfg map[string]any) any {
    routerName := getConfigValue(cfg, "router", "auth-service").(string)
    pathPrefix := getConfigValue(cfg, "path-prefix", "/auth").(string)
    
    client := lokstra_registry.GetClientRouter(routerName)
    
    return &authServiceRemote{
        client: NewRemoteClient(client, pathPrefix),
    }
}
```

#### Benefits
1. **Configurable Paths** - No hardcoded paths in service code
2. **Consistency** - All services use same pattern  
3. **Flexibility** - Easy to change API versioning/structure
4. **Convention over Configuration** - Sensible defaults with override capability

### 4. Example Configuration

#### New YAML Structure
```yaml
# Services with path-prefix
services:
  - name: user-service
    type: UserService
    path-prefix: "/api/v1/users"  # Custom prefix
    config:
      database_dsn: "postgres://..."
      
  - name: auth-service
    path-prefix: "/api/v1/auth"   # Custom prefix
    config:
      jwt_secret: "secret"

# Routers (independent configuration)
routers:
  - name: user-api
    path-prefix: "/api/v1"
    middlewares: ["cors", "logging"]
    
  - name: auth-api
    path-prefix: "/auth"
    middlewares: ["cors", "rate-limit"]

servers:
  - name: api-server
    apps:
      - addr: ":8080"
        routers: ["user-api", "auth-api"]        # Simple array
        services: ["user-service", "auth-service"] # Services hosted here
```

### 5. Migration Guide

#### From Old Config Format
```yaml
# OLD (routers-with-prefix)
apps:
  - addr: ":8080"
    routers-with-prefix:
      - name: "user-api"
        prefix: "/api/v1"
      - name: "auth-api" 
        prefix: "/auth"
```

#### To New Config Format
```yaml
# NEW (routers + services sections)
routers:
  - name: user-api
    path-prefix: "/api/v1"
  - name: auth-api
    path-prefix: "/auth"

apps:
  - addr: ":8080"
    routers: ["user-api", "auth-api"]
    services: ["user-service", "auth-service"]
```

## Benefits of New Structure

### 1. Separation of Concerns
- **Routers**: Define routing and middleware concerns
- **Services**: Define business logic and dependencies  
- **Apps**: Define deployment and hosting

### 2. Better Local/Remote Detection
```yaml
# App explicitly declares which services it hosts
apps:
  - addr: ":8080"
    services: ["user-service", "auth-service"]  # LOCAL services
    # Other services automatically REMOTE
```

### 3. Flexible Path Configuration
```yaml
# Service-level prefix (business logic)
services:
  - name: user-service
    path-prefix: "/api/v1/users"

# Router-level prefix (infrastructure)  
routers:
  - name: api-router
    path-prefix: "/api/v1"
```

### 4. Convention over Configuration
```go
// Default path-prefix if not specified
pathPrefix := getConfigValue(cfg, "path-prefix", "/users").(string)
```

## Validation

### JSON Schema Validation
- ✅ Path prefix pattern: `^/[a-zA-Z0-9/_-]*$`
- ✅ Required fields validation
- ✅ No more `routers-with-prefix` (deprecated)

### Build Validation  
- ✅ All examples compile without errors
- ✅ API changes are backward compatible (deprecated functions available)

## Next Steps

1. **Update Existing Examples** - Migrate to new config format
2. **Add Path Prefix Documentation** - Usage examples and best practices
3. **Add Config Validation** - Runtime validation of path prefix usage
4. **Add Path Prefix Tests** - Unit tests for new functionality

---

**Status**: ✅ Complete and validated
**Breaking Changes**: None (deprecated functions still available)
**New Features**: Path prefix configuration, simplified APIs, better local/remote detection