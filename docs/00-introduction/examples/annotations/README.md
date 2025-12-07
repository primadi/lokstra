# Lokstra Annotations Examples

This folder contains complete examples of using Lokstra annotations for dependency injection and code generation.

## Files

### 1. `service_example.go` - @Service Examples

Pure services without HTTP endpoints, demonstrating:

**AuthService:**
- ✅ Required dependencies: `@Inject "user-repository"`, `@Inject "cache-service"`
- ✅ String config: `@InjectCfgValue "auth.jwt-secret"`
- ✅ Duration config: `@InjectCfgValue key="auth.token-expiry", default="24h"`
- ✅ Int config: `@InjectCfgValue key="auth.max-attempts", default=5`
- ✅ Bool config: `@InjectCfgValue key="auth.debug-mode", default=false`

**NotificationService:**
- ✅ String config: `@InjectCfgValue "smtp.host"`
- ✅ Int config with default: `@InjectCfgValue key="smtp.port", default=587`
- ✅ String config with default: `@InjectCfgValue key="smtp.from-email", default="noreply@example.com"`
- ✅ Bool config with default: `@InjectCfgValue key="notification.enabled", default=true`

### 2. `router_service_with_config_example.go` - @RouterService Example

HTTP service with routes, demonstrating:

**UserAPIService:**
- ✅ HTTP routes: `@Route "GET /{id}"`, `@Route "POST /"`, etc.
- ✅ Per-route middleware: `middlewares=["auth"]`, `middlewares=["auth", "admin"]`
- ✅ Required dependencies: `@Inject "user-repository"`, `@Inject "cache-service"`
- ✅ Multiple config injections for API settings, rate limiting, pagination
- ✅ Bool, int, duration, string configs

### 3. `init_example.go` - Init() Method Pattern

Service with post-initialization setup:

**CacheManager:**
- ✅ Config injection: `@InjectCfgValue key="cache.max-size", default=1000`
- ✅ `Init() error` method called after dependency injection
- ✅ Internal state initialization (maps, slices)
- ✅ Configuration validation
- ✅ Pre-loading data

### 4. `router_with_init_example.go` - RouterService with Init()

HTTP service with initialization:

**ProductAPIService:**
- ✅ Dependency injection: `@Inject "product-repository"`
- ✅ Config injection: `@InjectCfgValue key="api.products.max-items", default=100`
- ✅ `Init() error` method for cache setup
- ✅ HTTP routes with internal state

### 5. `config.example.yaml` - Configuration

Example configuration file showing:
- Config values for all injected configurations
- Service definitions (dependencies)
- Deployment configuration for development and production

## How to Use

### 1. Copy Examples to Your Project

```bash
cp service_example.go your-project/application/
cp router_service_with_config_example.go your-project/application/
cp config.example.yaml your-project/config.yaml
```

### 2. Generate Code

```bash
# Automatic (recommended)
go run .  # lokstra.Bootstrap() auto-generates

# Manual
lokstra autogen ./application

# Force rebuild
go run . --generate-only
```

### 3. Generated Code

After running code generation, you'll get `zz_generated.lokstra.go` with:

**For @Service:**
```go
func RegisterAuthService() {
    lokstra_registry.RegisterLazyService("auth-service", func(deps map[string]any, cfg map[string]any) any {
        svc := &AuthService{
            Cache:       deps["cache-service"].(CacheService),
            UserRepo:    deps["user-repository"].(UserRepository),
            DebugMode:   cfg["auth.debug-mode"].(bool),
            JwtSecret:   cfg["auth.jwt-secret"].(string),
            MaxAttempts: cfg["auth.max-attempts"].(int),
            TokenExpiry: cfg["auth.token-expiry"].(time.Duration),
        }
        
        return svc
    }, map[string]any{
        "depends-on": []string{ "cache-service", "user-repository", },
        "auth.debug-mode": lokstra_registry.GetConfig("auth.debug-mode", false),
        "auth.jwt-secret": lokstra_registry.GetConfig("auth.jwt-secret", ""),
        "auth.max-attempts": lokstra_registry.GetConfig("auth.max-attempts", 5),
        "auth.token-expiry": lokstra_registry.GetConfig("auth.token-expiry", 24*time.Hour),
    })
}
```

**For @Service with Init():**
```go
func RegisterCacheManager() {
    lokstra_registry.RegisterLazyService("cache-manager", func(deps map[string]any, cfg map[string]any) any {
        svc := &CacheManager{
            MaxSize:    cfg["cache.max-size"].(int),
            TTLSeconds: cfg["cache.ttl-seconds"].(int),
        }
        
        // Call Init() for post-initialization
        if err := svc.Init(); err != nil {
            panic("failed to initialize cache-manager: " + err.Error())
        }
        
        return svc
    }, map[string]any{
        "cache.max-size": lokstra_registry.GetConfig("cache.max-size", 1000),
        "cache.ttl-seconds": lokstra_registry.GetConfig("cache.ttl-seconds", 300),
    })
}
```

**For @RouterService:**
```go
func UserAPIServiceFactory(deps map[string]any, config map[string]any) any {
    svc := &UserAPIService{
        Cache:            deps["cache-service"].(CacheService),
        UserRepo:         deps["user-repository"].(UserRepository),
        RateLimitEnabled: config["api.rate-limit.enabled"].(bool),
        MaxRequests:      config["api.rate-limit.max-requests"].(int),
        // ... all other fields
    }
    
    return svc
}

func RegisterUserAPIService() {
    lokstra_registry.RegisterRouterServiceType("user-api-service-factory",
        UserAPIServiceFactory,
        UserAPIServiceRemoteFactory,
        &deploy.ServiceTypeConfig{
            PathPrefix: "/api/v1/users",
            Middlewares: []string{"recovery", "request-logger"},
            RouteOverrides: map[string]deploy.RouteConfig{
                "GetByID": {Path: "GET /{id}"},
                "Create":  {Path: "POST /", Middlewares: []string{"auth"}},
                // ... all routes
            },
        },
    )
    
    lokstra_registry.RegisterLazyService("user-api-service",
        "user-api-service-factory",
        map[string]any{
            "depends-on": []string{ "cache-service", "user-repository", },
            "api.rate-limit.enabled": lokstra_registry.GetConfig("api.rate-limit.enabled", true),
            // ... all config values
        })
}
```

## Annotation Reference

### @Service
```go
// @Service name="service-name"
type ServiceName struct { }
```
- Used for pure services (no HTTP)
- Auto-generates `RegisterLazyService`
- All dependencies are **mandatory** (panic if not found)

### @RouterService
```go
// @RouterService name="service-name", prefix="/api/path", middlewares=["mw1"]
type ServiceName struct { }
```
- Used for HTTP services
- Auto-generates routes, factory, and registration
- All dependencies are **mandatory** (panic if not found)

### @Inject
```go
// Inject service dependency
// @Inject "service-name"
Field ServiceType
```
- Injects service dependencies
- Works with both `@Service` and `@RouterService`
- **All dependencies are mandatory** - framework panics if not found
- Generates: `deps["service-name"].(ServiceType)`

### @InjectCfgValue
```go
// Required config
// @InjectCfgValue "config.key"
Field string

// With default (unquoted for non-string types)
// @InjectCfgValue key="config.key", default=100
Field int

// @InjectCfgValue key="config.key", default=true
Field bool

// @InjectCfgValue key="config.key", default="24h"
Field time.Duration
```
- Injects configuration from `config.yaml`
- Works with both `@Service` and `@RouterService`
- Auto-detects type: `string`, `int`, `bool`, `float64`, `time.Duration`
- Default values are type-specific (no quotes for int/bool/float)

### @Route
```go
// @Route "GET /path/{id}"
func (s *Service) Method(p *Params) (*Result, error)

// With per-route middleware
// @Route "POST /path", middlewares=["auth", "admin"]
func (s *Service) Method(p *Params) (*Result, error)
```
- Only for `@RouterService`
- Defines HTTP endpoints

### Init() Method (Optional)

```go
// @Service name="my-service"
type MyService struct {
    // @InjectCfgValue key="max-size", default=100
    MaxSize int
    
    // Internal state (not injected)
    cache map[string]any
}

// Called automatically after dependency injection
func (s *MyService) Init() error {
    // Initialize internal state
    s.cache = make(map[string]any, s.MaxSize)
    
    // Validate configuration
    if s.MaxSize <= 0 {
        return fmt.Errorf("max size must be positive")
    }
    
    // Pre-load data, setup connections, etc.
    log.Println("Service initialized")
    return nil
}
```

**Init() Requirements:**
- Method name must be exactly `Init`
- Signature: `func (s *Service) Init() error`
- No parameters
- Returns `error`
- Called after all dependencies and configs are injected
- If returns error, service creation panics

**Use Init() for:**
- ✅ Initialize maps, slices, channels
- ✅ Validate injected configuration
- ✅ Pre-load data or cache
- ✅ Setup complex internal state
- ✅ Establish connections (after config available)

## Testing

1. **Start the application:**
   ```bash
   go run .
   ```

2. **Test endpoints:**
   ```bash
   # Get user
   curl http://localhost:8080/api/v1/users/123
   
   # List users
   curl http://localhost:8080/api/v1/users
   
   # Create user (requires auth)
   curl -X POST http://localhost:8080/api/v1/users \
        -H "Content-Type: application/json" \
        -d '{"name":"John","email":"john@example.com"}'
   ```

## Key Features Demonstrated

✅ **Dependency Injection**: Mandatory service dependencies (panic if missing)  
✅ **Configuration Injection**: Type-safe config from YAML  
✅ **Auto Type Detection**: Config types inferred from field type  
✅ **Default Values**: Sensible defaults when config missing  
✅ **Init() Method**: Post-initialization setup and validation  
✅ **HTTP Routes**: Automatic REST API generation  
✅ **Per-Route Middleware**: Fine-grained access control  
✅ **Service Separation**: `@Service` for logic, `@RouterService` for HTTP  

## Design Patterns

### 1. Simple Service (Config Only)
```go
// @Service name="email-service"
type EmailService struct {
    // @InjectCfgValue "smtp.host"
    SMTPHost string
}
```

### 2. Service with Dependencies
```go
// @Service name="auth-service"
type AuthService struct {
    // @Inject "user-repository"
    UserRepo UserRepository
    
    // @InjectCfgValue "auth.jwt-secret"
    JwtSecret string
}
```

### 3. Service with Init()
```go
// @Service name="cache-manager"
type CacheManager struct {
    // @InjectCfgValue key="cache.max-size", default=1000
    MaxSize int
    
    cache map[string]any
}

func (c *CacheManager) Init() error {
    c.cache = make(map[string]any, c.MaxSize)
    return nil
}
```

### 4. HTTP Service
```go
// @RouterService name="user-api", prefix="/api/users"
type UserAPIService struct {
    // @Inject "user-repository"
    UserRepo UserRepository
}

// @Route "GET /{id}"
func (s *UserAPIService) Get(p *GetParams) (*User, error) { }
```

## Next Steps

1. Read the [Full Documentation](https://primadi.github.io/lokstra/)
2. See [AI Agent Guide](../../AI-AGENT-GUIDE.md) for best practices
3. Check [Quick Reference](../../QUICK-REFERENCE.md) for common patterns
4. Explore [Full Framework Examples](../full-framework/) for larger projects

Example configuration file showing:
- Config values for all injected configurations
- Service definitions (dependencies)
- Deployment configuration for development and production

## How to Use

### 1. Copy Examples to Your Project

```bash
cp service_example.go your-project/application/
cp router_service_with_config_example.go your-project/application/
cp config.example.yaml your-project/config.yaml
```

### 2. Generate Code

```bash
# Automatic (recommended)
go run .  # lokstra.Bootstrap() auto-generates

# Manual
lokstra autogen ./application

# Force rebuild
go run . --generate-only
```

### 3. Generated Code

After running code generation, you'll get `zz_generated.lokstra.go` with:

**For @Service:**
```go
func RegisterAuthService() {
    lokstra_registry.RegisterLazyService("auth-service", func(deps map[string]any, cfg map[string]any) any {
        return &AuthService{
            UserRepo:    lokstra_registry.GetService[UserRepository]("user-repository"),
            Cache:       // Optional - nil if not found
            JwtSecret:   lokstra_registry.GetConfig("auth.jwt-secret", ""),
            TokenExpiry: lokstra_registry.GetConfigDuration("auth.token-expiry", 24*time.Hour),
            MaxAttempts: lokstra_registry.GetConfigInt("auth.max-attempts", 5),
            DebugMode:   lokstra_registry.GetConfigBool("auth.debug-mode", false),
        }
    }, nil)
}
```

**For @RouterService:**
```go
func UserAPIServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserAPIService{
        UserRepo:         deps["user-repository"].(domain.UserRepository),
        Cache:            // Optional - nil if not found
        RateLimitEnabled: lokstra_registry.GetConfigBool("api.rate-limit.enabled", true),
        MaxRequests:      lokstra_registry.GetConfigInt("api.rate-limit.max-requests", 100),
        // ... all other configs
    }
}

func RegisterUserAPIService() {
    lokstra_registry.RegisterRouterServiceType("user-api-service-factory",
        UserAPIServiceFactory,
        UserAPIServiceRemoteFactory,
        &deploy.ServiceTypeConfig{
            PathPrefix: "/api/v1/users",
            Middlewares: []string{"recovery", "request-logger"},
            RouteOverrides: map[string]deploy.RouteConfig{
                "GetByID": {Path: "GET /{id}"},
                "Create":  {Path: "POST /", Middlewares: []string{"auth"}},
                // ... all routes
            },
        },
    )
}
```

## Annotation Reference

### @Service
```go
// @Service name="service-name"
type ServiceName struct { }
```
- Used for pure services (no HTTP)
- Auto-generates `RegisterLazyService`

### @RouterService
```go
// @RouterService name="service-name", prefix="/api/path", middlewares=["mw1"]
type ServiceName struct { }
```
- Used for HTTP services
- Auto-generates routes, factory, and registration

### @Inject
```go
// Required
// @Inject "service-name"
Field ServiceType

// Optional
// @Inject service="service-name", optional=true
Field ServiceType  // nil if not found
```
- Injects service dependencies
- Works with both `@Service` and `@RouterService`

### @InjectCfgValue
```go
// Required
// @InjectCfgValue "config.key"
Field string

// With default
// @InjectCfgValue key="config.key", default="value"
Field string
```
- Injects configuration from `config.yaml`
- Works with both `@Service` and `@RouterService`
- Auto-detects type: `string`, `int`, `bool`, `float64`, `time.Duration`

### @Route
```go
// @Route "GET /path/{id}"
func (s *Service) Method(p *Params) (*Result, error)

// With per-route middleware
// @Route "POST /path", middlewares=["auth", "admin"]
func (s *Service) Method(p *Params) (*Result, error)
```
- Only for `@RouterService`
- Defines HTTP endpoints

## Testing

1. **Start the application:**
   ```bash
   go run .
   ```

2. **Test endpoints:**
   ```bash
   # Get user
   curl http://localhost:8080/api/v1/users/123
   
   # List users
   curl http://localhost:8080/api/v1/users
   
   # Create user (requires auth)
   curl -X POST http://localhost:8080/api/v1/users \
        -H "Content-Type: application/json" \
        -d '{"name":"John","email":"john@example.com"}'
   ```

## Key Features Demonstrated

✅ **Dependency Injection**: Required and optional service dependencies  
✅ **Configuration Injection**: Type-safe config from YAML  
✅ **Optional Dependencies**: Graceful degradation (e.g., cache)  
✅ **Auto Type Detection**: Config types inferred from field type  
✅ **Default Values**: Sensible defaults when config missing  
✅ **HTTP Routes**: Automatic REST API generation  
✅ **Per-Route Middleware**: Fine-grained access control  
✅ **Service Separation**: `@Service` for logic, `@RouterService` for HTTP  

## Next Steps

1. Read the [Full Documentation](https://primadi.github.io/lokstra/)
2. See [AI Agent Guide](../../AI-AGENT-GUIDE.md) for best practices
3. Check [Quick Reference](../../QUICK-REFERENCE.md) for common patterns
4. Explore [Full Framework Examples](../full-framework/) for larger projects
