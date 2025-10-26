# Remote Service Auto-Router Integration

## Overview

`RemoteService` sekarang secara otomatis membaca konfigurasi `auto-router` dari service definition di YAML config. Ini memastikan bahwa client-side remote calls menggunakan convention, path-prefix, resource-name, dan route overrides yang sama dengan server-side router.

## Key Features

### 1. **Auto-Router Configuration Support**

RemoteService membaca dan menerapkan konfigurasi dari `auto-router`:

```yaml
services:
  - name: auth-service
    type: auth_service
    auto-router:
      convention: "rest"              # Convention: rest, rpc, kebab-case
      path-prefix: "/api/v1"          # Path prefix
      resource-name: "auth"           # Resource name (singular)
      plural-resource-name: "auths"   # Resource name (plural, optional)
      routes:                         # Route overrides
        - name: Login
          method: POST
          path: "/login"
        - name: Logout
          method: POST
          path: "/logout"
        - name: ValidateToken
          method: POST
          path: "/validate-token"
```

### 2. **Automatic Configuration Propagation**

When a remote service is registered, the `auto-router` configuration is automatically passed to the service factory:

```go
// In lokstra_registry/config.go
remoteConfig := map[string]any{
    "router":              svc.Name,
    "convention":          svc.AutoRouter.Convention,
    "path-prefix":         svc.AutoRouter.PathPrefix,
    "resource-name":       svc.AutoRouter.ResourceName,
    "plural-resource-name": svc.AutoRouter.PluralResourceName,
    "routes":              convertedRoutes,
}
```

### 3. **Convention-Based Path Generation**

RemoteService supports multiple conventions:

#### REST Convention (Default)
```go
// Method: GetUser
// Path params: ["dep", "id"]
// Result: GET /users/{dep}/{id}

// Method: CreateUser
// Path param: ["dep"]
// Result: POST /users/{dep}

// Method: ListUsers
// Path param: ["dep"]
// Result: GET /users/{dep}
```

#### Kebab-Case Convention
```go
// Method: ValidateToken
// Result: POST /auth/validate-token

// Method: ProcessPayment
// Result: POST /payment/process-payment
```

#### RPC Convention
```go
// Method: ValidateToken
// Result: POST /auth/validate_token

// Method: ProcessPayment
// Result: POST /payment/process_payment
```

### 4. **Path Parameter Detection**

RemoteService automatically detects path parameters from struct tags:

```go
type GetUserRequest struct {
    DepartmentID string `path:"dep"`  // Detected as path param
    UserID       string `path:"id"`   // Detected as path param
}

// Automatically generates path: /users/{dep}/{id}
```

### 5. **Route Override Support**

Route overrides from YAML config are automatically applied:

```yaml
routes:
  - name: Login           # Method name
    method: POST          # HTTP method override
    path: "/login"        # Path override
```

The remote service will use these overrides instead of convention-based paths.

## Implementation Flow

### Server-Side (config.go)

```go
// registerService in lokstra_registry/config.go
if location.IsLocal {
    // LOCAL SERVICE
    RegisterLazyService(svc.Name, svcType, cfg, AllowOverride(true))
} else {
    // REMOTE SERVICE - Pass auto-router config
    remoteConfig := map[string]any{
        "router":              svc.Name,
        "convention":          svc.AutoRouter.Convention,
        "path-prefix":         svc.AutoRouter.PathPrefix,
        "resource-name":       svc.AutoRouter.ResourceName,
        "plural-resource-name": svc.AutoRouter.PluralResourceName,
        "routes":              routeOverrides,
    }
    RegisterLazyService(svc.Name, svcType, remoteConfig, AllowOverride(true))
}
```

### Service Factory (service.go)

```go
// GetRemoteService in lokstra_registry/service.go
func GetRemoteService(cfg map[string]any) *api_client.RemoteService {
    routerName := utils.GetValueFromMap(cfg, "router", "")
    convention := utils.GetValueFromMap(cfg, "convention", "rest")
    pathPrefix := utils.GetValueFromMap(cfg, "path-prefix", "/")
    resourceName := utils.GetValueFromMap(cfg, "resource-name", "")
    pluralResourceName := utils.GetValueFromMap(cfg, "plural-resource-name", "")
    
    clientRouter := GetClientRouter(routerName)
    remoteService := api_client.NewRemoteService(clientRouter, pathPrefix)
    
    // Apply configuration
    remoteService.WithConvention(convention)
    remoteService.WithResourceName(resourceName)
    remoteService.WithPluralResourceName(pluralResourceName)
    
    // Apply route overrides
    if routes, ok := cfg["routes"].([]any); ok {
        for _, routeRaw := range routes {
            if routeMap, ok := routeRaw.(map[string]any); ok {
                name := utils.GetValueFromMap(routeMap, "name", "")
                method := utils.GetValueFromMap(routeMap, "method", "")
                path := utils.GetValueFromMap(routeMap, "path", "")
                
                if name != "" && path != "" {
                    remoteService.WithRouteOverride(name, path)
                }
                if name != "" && method != "" {
                    remoteService.WithMethodOverride(name, method)
                }
            }
        }
    }
    
    return remoteService
}
```

### Remote Service Client (client_remote_service.go)

```go
// RemoteService struct
type RemoteService struct {
    client             *ClientRouter
    basePath           string            
    convention         string            // e.g., "rest", "rpc"
    resourceName       string            
    pluralResourceName string            
    routeOverrides     map[string]string // methodName -> custom path
    methodOverrides    map[string]string // methodName -> HTTP method
}

// methodToHTTP - Main path generation logic
func (c *RemoteService) methodToHTTP(methodName string, req any) (httpMethod string, path string) {
    // 1. Extract path params from struct tags
    pathParams, _, _ := c.extractStructMetadata(req)
    
    // 2. Check for configured route override
    if overridePath, exists := c.routeOverrides[methodName]; exists {
        path = overridePath
    }
    
    // 3. Check for configured method override
    if overrideMethod, exists := c.methodOverrides[methodName]; exists {
        httpMethod = overrideMethod
    }
    
    // 4. Build path based on convention if not overridden
    if path == "" {
        switch c.convention {
        case "rest":
            path = c.buildRESTPath(methodName, pathParams)
        case "kebab-case":
            path = c.buildKebabPath(methodName)
        case "rpc":
            path = c.buildRPCPath(methodName)
        }
    }
    
    return httpMethod, path
}
```

## Example Configuration

### Complete YAML Configuration

```yaml
services:
  - name: user-service
    type: user_service
    auto-router:
      convention: "rest"
      path-prefix: "/api/v1"
      resource-name: "user"
      plural-resource-name: "users"
      
  - name: auth-service
    type: auth_service
    auto-router:
      convention: "rest"
      path-prefix: "/api/v1"
      resource-name: "auth"
      routes:
        - name: Login
          method: POST
          path: "/login"
        - name: ValidateToken
          method: POST
          path: "/validate-token"
    depends-on: [user-service]
    
  - name: payment-service
    type: payment_service
    auto-router:
      convention: "rpc"
      path-prefix: "/api/v1"
      routes:
        - name: ProcessPayment
          method: POST
          path: "/payments"
```

### Service Factory Example

```go
// User Service Remote Factory
func CreateUserServiceRemote(cfg map[string]any) any {
    return &userServiceRemote{
        client: lokstra_registry.GetRemoteService(cfg),
        // GetRemoteService automatically reads:
        // - convention: "rest"
        // - path-prefix: "/api/v1"
        // - resource-name: "user"
        // - plural-resource-name: "users"
    }
}

// Auth Service Remote Factory
func CreateAuthServiceRemote(cfg map[string]any) any {
    return &authServiceRemote{
        client: lokstra_registry.GetRemoteService(cfg),
        // GetRemoteService automatically reads:
        // - convention: "rest"
        // - path-prefix: "/api/v1"
        // - resource-name: "auth"
        // - routes: [Login, ValidateToken] with overrides
    }
}
```

### Remote Service Usage

```go
type userServiceRemote struct {
    client *api_client.RemoteService
}

// GetUser - Automatic path generation
func (s *userServiceRemote) GetUser(ctx *request.Context, req *GetUserRequest) (*User, error) {
    // Convention: rest
    // Resource: user
    // Method: GetUser
    // Path params from struct tags: ["dep", "id"]
    // → Generates: GET /api/v1/users/{dep}/{id}
    return api_client.CallRemoteService[*User](s.client, "GetUser", ctx, req)
}

type authServiceRemote struct {
    client *api_client.RemoteService
}

// Login - Uses route override
func (s *authServiceRemote) Login(ctx *request.Context, req *LoginRequest) (*LoginResponse, error) {
    // Override path: "/login"
    // → Uses: POST /api/v1/login (from route override)
    return api_client.CallRemoteService[*LoginResponse](s.client, "Login", ctx, req)
}

// ValidateToken - Uses route override
func (s *authServiceRemote) ValidateToken(ctx *request.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
    // Override path: "/validate-token"
    // → Uses: POST /api/v1/validate-token (from route override)
    return api_client.CallRemoteService[*ValidateTokenResponse](s.client, "ValidateToken", ctx, req)
}
```

## Benefits

1. **✅ Consistency**: Client and server use the same routing configuration
2. **✅ Maintainability**: Single source of truth in YAML config
3. **✅ Flexibility**: Support for multiple conventions and route overrides
4. **✅ Type Safety**: Path parameters detected from struct tags
5. **✅ Less Boilerplate**: No need to manually construct paths
6. **✅ Convention Over Configuration**: Sensible defaults with override capability

## Configuration Priority

The system uses the following priority for path generation:

1. **Route Override** (highest priority)
   - From `auto-router.routes[].path` in YAML
   
2. **Convention-Based Generation**
   - Using `auto-router.convention` + path parameters from struct tags
   
3. **Method Override**
   - From `auto-router.routes[].method` in YAML
   
4. **Default Inference**
   - From method name prefix (Get*, Create*, Update*, Delete*)

## Testing

To test the auto-router integration:

```bash
# Run the example
cd cmd/examples/25-single-binary-deployment
go run . -server server-01

# Test endpoints
curl -X POST http://localhost:8081/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"secret"}'

curl -X GET http://localhost:8081/api/v1/users/engineering/123

curl -X POST http://localhost:8081/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user_1","items":[...]}'
```

## Related Files

- `api_client/client_remote_service.go` - RemoteService implementation
- `lokstra_registry/service.go` - GetRemoteService helper
- `lokstra_registry/config.go` - Configuration loading and service registration
- `core/config/config.go` - AutoRouter config structure
- `core/router/convention.go` - Convention-based path generation

## Migration Guide

### Before (Manual Configuration)

```go
func CreateAuthServiceRemote(cfg map[string]any) any {
    routerName := utils.GetValueFromMap(cfg, "router", "auth-service")
    pathPrefix := utils.GetValueFromMap(cfg, "path-prefix", "/api/v1")
    
    clientRouter := lokstra_registry.GetClientRouter(routerName)
    client := api_client.NewRemoteService(clientRouter, pathPrefix)
    
    // Manual path construction in each method
    return &authServiceRemote{client: client}
}
```

### After (Automatic Configuration)

```go
func CreateAuthServiceRemote(cfg map[string]any) any {
    // Automatically reads auto-router config from YAML
    return &authServiceRemote{
        client: lokstra_registry.GetRemoteService(cfg),
    }
}
```

The YAML config now controls all routing behavior!
