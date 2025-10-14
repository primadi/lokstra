# Router Configuration Improvements

## Problem Statement

The old configuration pattern had routers with prefix defined in two places:
1. In code: `router.SetPathPrefix("/api/v1")`
2. In config: `routers-with-prefix` with custom prefixes per server

This caused issues:
- **Inconsistency**: Same router could have different prefixes on different servers
- **Duplication**: Path prefix logic in both code and config
- **Complexity**: Two ways to define routers (`routers` vs `routers-with-prefix`)
- **Hard to maintain**: Changing API version requires code changes

## Solution

### New Configuration Structure

#### 1. Router Definition (Top Level)

```yaml
routers:
  - name: user-api
    path-prefix: /api/v1              # ✅ Prepended to code prefix (additive)
    middlewares: [logger, cors]       # ✅ Apply middlewares from config
```

**How Path Prefix Works (Additive):**
```
Code prefix:   /users
Config prefix: /api/v1
Final path:    /api/v1/users  ← Combined!
```

**Benefits:**
- Config prefix is prepended (not replaced)
- Code defines semantic path (/users, /products)
- Config defines API versioning/namespace (/api/v1, /api/v2)
- Easy to change API version without touching code
- Middlewares can be applied from config (with their own config)

#### 2. Server/App Configuration

```yaml
servers:
  - name: auth-server
    apps:
      - name: auth-app
        addr: ":8081"
        services: [user-service, auth-service]
        routers: [user-api, auth-api]    # ✅ Just reference by name
```

**Benefits:**
- Simpler - just list router names
- No need for `routers-with-prefix`
- Path prefix comes from router definition

#### 3. Middleware Configuration (Separate)

```yaml
middlewares:
  - name: logger
    type: request_logger
    config:                          # ✅ Config belongs to middleware
      format: "json"
      log_level: "info"
      
  - name: rate-limiter
    type: rate_limit
    config:                          # ✅ Each middleware has its own config
      max_requests_per_minute: 100
      burst: 10
      
  - name: timeout
    type: timeout
    config:                          # ✅ Clear and unambiguous
      duration: 30s

routers:
  - name: api-router
    path-prefix: /api/v1
    middlewares: [logger, rate-limiter, timeout]  # ✅ Reference by name
```

**Why This is Better:**
- **No ambiguity** - Each middleware config is clear
- **Reusability** - Same middleware can be used by multiple routers
- **Flexibility** - Can create multiple middleware instances with different configs
- **Clarity** - Config belongs to middleware, not router

### Code Changes

#### Router Implementation (Before)

```go
func SetupUserRouter() router.Router {
	r := router.New("user-api")
	r.SetPathPrefix("/api/v1")           // ❌ Hardcoded in code
	
	r.GET("/users/{id}", handler)
	return r
}
```

#### Router Implementation (After - Additive Pattern)

```go
func SetupUserRouter() router.Router {
	r := router.New("user-api")
	r.SetPathPrefix("/users")            // ✅ Semantic path in code
	// Config adds: /api/v1
	// Final path: /api/v1/users
	
	r.GET("/{id}", handler)              // Route: /{id}
	return r                              // Full path: /api/v1/users/{id}
}
```

**Why This is Better:**
- **Separation of concerns**: Code defines resource paths, config defines API structure
- **Flexibility**: Easy to test router without API prefix
- **Versioning**: Change from `/api/v1` to `/api/v2` in config only
- **Consistency**: Same router code works across different API versions

## Implementation Details

### Config Structure

```go
// core/config/config.go

type Config struct {
	Configs     []*GeneralConfig `yaml:"configs,omitempty"`
	Services    ServicesConfig   `yaml:"services,omitempty"`
	Middlewares []*Middleware    `yaml:"middlewares,omitempty"`
	Routers     []*Router        `yaml:"routers,omitempty"`    // ✅ NEW
	Servers     []*Server        `yaml:"servers,omitempty"`
}

// ✅ NEW: Router configuration
type Router struct {
	Name        string         `yaml:"name"`
	PathPrefix  string         `yaml:"path-prefix,omitempty"`
	Middlewares []string       `yaml:"middlewares,omitempty"`
	Config      map[string]any `yaml:"config,omitempty"`
}

type App struct {
	Name         string   `yaml:"name"`
	Addr         string   `yaml:"addr"`
	Services     []string `yaml:"services,omitempty"`
	Routers      []string `yaml:"routers,omitempty"`
	// ❌ REMOVED: RoutersWithPrefix []*RouterWithPrefix
}
```

### Registry Processing

```go
// lokstra_registry/config.go

func RegisterConfig(c *config.Config) {
	// ... service registration ...
	
	// Process routers and store their config
	routerConfigs := make(map[string]*config.Router)
	for _, routerCfg := range c.Routers {
		routerConfigs[routerCfg.Name] = routerCfg
	}
	
	// Apply servers and apps
	for _, srvConfig := range c.Servers {
		for _, appConfig := range srvConfig.Apps {
			// Get routers and apply config
			for _, routerName := range appConfig.Routers {
				r := GetRouter(routerName)
				
				// Apply router config if available
				if routerCfg, exists := routerConfigs[routerName]; exists {
					r = r.Clone()  // Clone to avoid side effects
					
					// ✅ COMBINE path prefixes (additive)
					if routerCfg.PathPrefix != "" {
						existingPrefix := r.PathPrefix()      // From code: /users
						combinedPrefix := routerCfg.PathPrefix + existingPrefix
						r.SetPathPrefix(combinedPrefix)       // Result: /api/v1/users
					}
					
					// TODO: Apply middlewares
					// for _, mwName := range routerCfg.Middlewares {
					//     middleware := GetMiddleware(mwName)
					//     r.Use(middleware)
					// }
				}
				
				routers = append(routers, r)
			}
		}
	}
}
```

## Migration Guide

### Step 1: Update Config File

**Before:**
```yaml
servers:
  - name: auth-server
    apps:
      - routers: [health-api]
        routers-with-prefix:
          - name: user-api
            prefix: /api/v1/users
          - name: auth-api
            prefix: /api/v1/auth
```

**After:**
```yaml
# Add router definitions at top level
routers:
  - name: user-api
    path-prefix: /api/v1/users
    middlewares: [logger]
  
  - name: auth-api
    path-prefix: /api/v1/auth
    middlewares: [logger, cors]

servers:
  - name: auth-server
    apps:
      - routers: [health-api, user-api, auth-api]  # ✅ Simpler!
```

### Step 2: Update Router Code

**Before:**
```go
func SetupUserRouter() router.Router {
	r := router.New("user-api")
	r.SetPathPrefix("/api/v1")       // ❌ Remove this
	
	r.GET("/users/{id}", handler)
	return r
}
```

**After (Additive Pattern):**
```go
func SetupUserRouter() router.Router {
	r := router.New("user-api")
	r.SetPathPrefix("/users")        // ✅ Semantic path
	// Config will add: /api/v1
	
	r.GET("/{id}", handler)          // Route path
	return r                          // Full: /api/v1/users/{id}
}
```

**Explanation:**
- Code defines `/users` (what resource)
- Config defines `/api/v1` (API version/namespace)
- Framework combines them: `/api/v1/users`

### Step 3: Test

The framework will now:
1. Read router config from `routers` section
2. Clone router when used in app
3. **Combine** config prefix + code prefix (additive)
4. Result: `/api/v1` (config) + `/users` (code) = `/api/v1/users`

## Example Configurations

### Microservices Example

```yaml
configs:
  - name: api_version
    value: "v1"

# Middleware configurations (separate, reusable)
middlewares:
  - name: logger
    type: request_logger
    config:
      format: "json"
      log_level: "info"
  
  - name: rate-limit-public
    type: rate_limit
    config:
      max_requests_per_minute: 10    # Strict for public
  
  - name: rate-limit-internal
    type: rate_limit
    config:
      max_requests_per_minute: 1000  # Relaxed for internal

# Router definitions reference middlewares
routers:
  - name: user-api
    path-prefix: /api/v1
    middlewares: [logger, rate-limit-internal]  # Uses internal rate limit
  
  - name: public-api
    path-prefix: /api/v1
    middlewares: [logger, rate-limit-public]    # Uses public rate limit

servers:
  - name: auth-server
    base-url: http://localhost:8081
    apps:
      - addr: ":8081"
        services: [user-service]
        routers: [user-api]          # ✅ Consistent prefix
  
  - name: public-server
    base-url: http://localhost:8082
    apps:
      - addr: ":8082"
        services: [public-service]
        routers: [public-api]        # ✅ Different middleware config
```

### API Versioning

Easy to support multiple API versions:

```yaml
middlewares:
  - name: logger
    type: request_logger
  
  - name: new-auth
    type: jwt_auth
    config:
      version: 2
      algorithm: "RS256"

routers:
  # V1 API
  - name: user-api-v1
    path-prefix: /api/v1
    middlewares: [logger]
  
  # V2 API (with breaking changes)
  - name: user-api-v2
    path-prefix: /api/v2
    middlewares: [logger, new-auth]  # Different middleware

servers:
  - name: api-server
    apps:
      - addr: ":8080"
        routers: [user-api-v1, user-api-v2]  # Support both versions
```

### Environment-Specific Prefixes

Development:
```yaml
routers:
  - name: user-api
    path-prefix: /dev/api/users    # ✅ Easy to change per environment
```

Production:
```yaml
routers:
  - name: user-api
    path-prefix: /api/users
```

## Benefits Summary

### For Configuration
- ✅ **Single source of truth** - Path prefix defined once
- ✅ **Consistency** - Same prefix across all servers
- ✅ **Flexibility** - Easy to change API versioning
- ✅ **Clarity** - No more `routers-with-prefix` complexity

### For Code
- ✅ **Simpler** - No hardcoded prefixes
- ✅ **Reusable** - Same router code across projects
- ✅ **Testable** - Can test router without prefix
- ✅ **Maintainable** - API structure in config, not scattered in code

### For Operations
- ✅ **Easy deployment** - Just change config for different environments
- ✅ **API versioning** - Add new versions without code changes
- ✅ **Monitoring** - Middleware configuration centralized
- ✅ **Documentation** - Config file is self-documenting

## Future Enhancements

### Middleware Application (TODO)
```yaml
routers:
  - name: user-api
    path-prefix: /api/v1/users
    middlewares: [logger, cors, auth]  # ⏳ TODO: Apply these automatically
```

### Route-Level Config (Future)
```yaml
routers:
  - name: user-api
    path-prefix: /api/v1/users
    middlewares: [logger]
    routes:                          # 💡 Future: Define routes in config
      - path: /{id}
        method: GET
        handler: GetUserHandler
        middlewares: [auth]
```

### Dynamic Routing (Future)
```yaml
routers:
  - name: dynamic-api
    path-prefix: /api/${api_version}/users  # 💡 Future: Template variables
```

## Related Documentation

- [Configuration Strategies](./configuration-strategies.md)
- [Example 24: Microservices Deployment](../cmd/examples/24-microservices-deployment/README.md)
- [YAML Configuration System](./yaml-configuration-system.md)
