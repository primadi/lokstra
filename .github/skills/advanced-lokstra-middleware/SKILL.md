---
name: advanced-lokstra-middleware
description: Create custom middleware for request/response filtering, logging, authentication, and authorization. Use after framework setup to add cross-cutting concerns.
phase: advanced
order: 2
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.21"
---

# Advanced: Custom Middleware

## When to Use

Use this skill when:
- Adding authentication/authorization checks
- Implementing request logging and metrics
- Creating rate limiting or throttling
- Adding request/response modification
- Handling cross-cutting concerns globally

Prerequisites:
- ✅ Framework initialized (see: implementation-lokstra-init-framework)
- ✅ Basic endpoints implemented
- ✅ Understand request flow

## Core Concepts

### Middleware Type Definition

```go
// Middleware function signature (from core/request/handler.go)
type HandlerFunc func(c *Context) error
```

### Key Context Methods

| Method | Description |
|--------|-------------|
| `c.Next()` | Calls the next middleware/handler in chain |
| `c.R` | Access to `*http.Request` |
| `c.W` | Access to response writer wrapper |
| `c.Api` | API response helpers (Ok, Error, Unauthorized, etc.) |
| `c.Set(key, value)` | Store value in context |
| `c.Get(key)` | Retrieve value from context |
| `c.StatusCode()` | Get response status code |

---

## Middleware Architecture

### Request Flow

```
Request
  ↓
Middleware 1 (pre-processing)
  ↓
Middleware 2 (pre-processing)
  ↓
Middleware 3 (pre-processing)
  ↓
Handler
  ↓
Middleware 3 (post-processing)
  ↓
Middleware 2 (post-processing)
  ↓
Middleware 1 (post-processing)
  ↓
Response
```

### Middleware Execution Order (Recommended)

```yaml
# In config.yaml
middlewares:
  - type: recovery           # 1. Panic recovery (FIRST)
  - type: request_logger     # 2. Logging
  - type: slow_request_logger # 3. Slow request detection
  - type: cors               # 4. CORS headers
  - type: body_limit         # 5. Request size protection
  - type: gzip_compression   # 6. Response compression (LAST)
```

**Why this order?**
1. **Recovery first** - Catches panics from all other middleware
2. **Logging early** - Records all requests, even failed ones
3. **CORS early** - Handles preflight before authentication
4. **Body limit before parsing** - Prevents memory exhaustion
5. **Compression last** - Compresses final response

---

## Built-in Middleware

Lokstra provides 6 built-in middleware packages:

| Middleware | Package | Config Type |
|------------|---------|-----------|
| Recovery | `middleware/recovery` | `recovery` |
| Request Logger | `middleware/request_logger` | `request_logger` |
| Slow Request Logger | `middleware/slow_request_logger` | `slow_request_logger` |
| CORS | `middleware/cors` | `cors` |
| Body Limit | `middleware/body_limit` | `body_limit` |
| Gzip Compression | `middleware/gzipcompression` | `gzip_compression` |

### Register Built-in Middleware

```go
import (
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/middleware/cors"
	"github.com/primadi/lokstra/middleware/body_limit"
	"github.com/primadi/lokstra/middleware/gzipcompression"
)

func registerMiddleware() {
	recovery.Register()
	request_logger.Register()
	cors.Register()
	body_limit.Register()
	gzipcompression.Register()
}
```

---

## Creating Custom Middleware

### Standard Pattern (3 Components)

Every Lokstra middleware should have these 3 components:

1. **Config struct** - Configuration options
2. **Middleware function** - Returns `request.HandlerFunc`
3. **MiddlewareFactory function** - For YAML configuration support
4. **Register function** - Registers with lokstra_registry

### Template: Custom Middleware

File: `middleware/your_middleware/your_middleware.go`

```go
package your_middleware

import (
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

// 1. Constants for registry type and param names
const YOUR_MIDDLEWARE_TYPE = "your_middleware"
const PARAMS_OPTION1 = "option1"
const PARAMS_OPTION2 = "option2"

// 2. Config struct with documentation
type Config struct {
	// Option1 is the first configuration option
	Option1 string
	
	// Option2 is the second configuration option
	Option2 int
}

// 3. DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Option1: "default_value",
		Option2: 100,
	}
}

// 4. Middleware function - returns request.HandlerFunc
func Middleware(cfg *Config) request.HandlerFunc {
	// Apply defaults if nil
	defConfig := DefaultConfig()
	if cfg == nil {
		cfg = defConfig
	}
	if cfg.Option1 == "" {
		cfg.Option1 = defConfig.Option1
	}
	if cfg.Option2 <= 0 {
		cfg.Option2 = defConfig.Option2
	}

	// Return the middleware handler
	return request.HandlerFunc(func(c *request.Context) error {
		// ========== PRE-PROCESSING ==========
		// Code here runs BEFORE the handler
		
		// Example: Log incoming request
		// logger.LogInfo("Request: %s %s", c.R.Method, c.R.URL.Path)
		
		// Example: Early return (reject request)
		// if someCondition {
		//     return c.Api.Unauthorized("Access denied")
		// }
		
		// ========== CALL NEXT ==========
		err := c.Next()
		
		// ========== POST-PROCESSING ==========
		// Code here runs AFTER the handler
		
		// Example: Log response status
		// logger.LogInfo("Response: %d", c.StatusCode())
		
		return err
	})
}

// 5. MiddlewareFactory - for YAML config support
func MiddlewareFactory(params map[string]any) request.HandlerFunc {
	defConfig := DefaultConfig()
	if params == nil {
		return Middleware(defConfig)
	}

	cfg := &Config{
		Option1: utils.GetValueFromMap(params, PARAMS_OPTION1, defConfig.Option1),
		Option2: utils.GetValueFromMap(params, PARAMS_OPTION2, defConfig.Option2),
	}
	return Middleware(cfg)
}

// 6. Register function - register with lokstra_registry
func Register() {
	lokstra_registry.RegisterMiddlewareFactory(YOUR_MIDDLEWARE_TYPE, MiddlewareFactory,
		lokstra_registry.AllowOverride(true))
}
```

---

## Example: Authentication Middleware

File: `middleware/auth/auth.go`

```go
package auth

import (
	"strings"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

const AUTH_TYPE = "auth"
const PARAMS_TOKEN_PREFIX = "token_prefix"
const PARAMS_HEADER_NAME = "header_name"

type Config struct {
	// TokenPrefix is the expected prefix (e.g., "Bearer")
	TokenPrefix string
	
	// HeaderName is the header to check (default: "Authorization")
	HeaderName string
	
	// ValidateFunc is a custom token validation function
	ValidateFunc func(token string) (map[string]any, error)
}

func DefaultConfig() *Config {
	return &Config{
		TokenPrefix: "Bearer",
		HeaderName:  "Authorization",
		ValidateFunc: nil,
	}
}

func Middleware(cfg *Config) request.HandlerFunc {
	defConfig := DefaultConfig()
	if cfg == nil {
		cfg = defConfig
	}
	if cfg.TokenPrefix == "" {
		cfg.TokenPrefix = defConfig.TokenPrefix
	}
	if cfg.HeaderName == "" {
		cfg.HeaderName = defConfig.HeaderName
	}

	return request.HandlerFunc(func(c *request.Context) error {
		// Get Authorization header
		authHeader := c.R.Header.Get(cfg.HeaderName)
		if authHeader == "" {
			logger.LogInfo("🔒 [auth] Missing %s header", cfg.HeaderName)
			return c.Api.Unauthorized("Missing authorization header")
		}

		// Check token prefix
		prefix := cfg.TokenPrefix + " "
		if len(authHeader) < len(prefix) || !strings.HasPrefix(authHeader, prefix) {
			logger.LogInfo("🔒 [auth] Invalid authorization format")
			return c.Api.Unauthorized("Invalid authorization format")
		}

		// Extract token
		token := authHeader[len(prefix):]
		if token == "" {
			return c.Api.Unauthorized("Empty token")
		}

		// Validate token (custom function or default)
		if cfg.ValidateFunc != nil {
			claims, err := cfg.ValidateFunc(token)
			if err != nil {
				logger.LogInfo("🔒 [auth] Invalid token: %v", err)
				return c.Api.Unauthorized("Invalid token")
			}
			
			// Store claims in context
			for key, value := range claims {
				c.Set(key, value)
			}
		}

		c.Set("authenticated", true)
		logger.LogInfo("✅ [auth] Authenticated")

		return c.Next()
	})
}

func MiddlewareFactory(params map[string]any) request.HandlerFunc {
	defConfig := DefaultConfig()
	if params == nil {
		return Middleware(defConfig)
	}

	cfg := &Config{
		TokenPrefix:  utils.GetValueFromMap(params, PARAMS_TOKEN_PREFIX, defConfig.TokenPrefix),
		HeaderName:   utils.GetValueFromMap(params, PARAMS_HEADER_NAME, defConfig.HeaderName),
		ValidateFunc: nil, // Cannot be set via params
	}
	return Middleware(cfg)
}

func Register() {
	lokstra_registry.RegisterMiddlewareFactory(AUTH_TYPE, MiddlewareFactory,
		lokstra_registry.AllowOverride(true))
}
```

---

## Example: Rate Limiting Middleware

File: `middleware/ratelimit/ratelimit.go`

```go
package ratelimit

import (
	"sync"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

const RATELIMIT_TYPE = "ratelimit"
const PARAMS_REQUESTS_PER_SECOND = "requests_per_second"
const PARAMS_BURST = "burst"

type Config struct {
	// RequestsPerSecond is the max requests per second per client
	RequestsPerSecond int
	
	// Burst allows temporary burst over the limit
	Burst int
}

func DefaultConfig() *Config {
	return &Config{
		RequestsPerSecond: 100,
		Burst:             10,
	}
}

type clientLimiter struct {
	tokens    int
	lastReset time.Time
}

type RateLimiter struct {
	config  *Config
	mu      sync.Mutex
	clients map[string]*clientLimiter
}

func Middleware(cfg *Config) request.HandlerFunc {
	defConfig := DefaultConfig()
	if cfg == nil {
		cfg = defConfig
	}
	if cfg.RequestsPerSecond <= 0 {
		cfg.RequestsPerSecond = defConfig.RequestsPerSecond
	}
	if cfg.Burst <= 0 {
		cfg.Burst = defConfig.Burst
	}

	limiter := &RateLimiter{
		config:  cfg,
		clients: make(map[string]*clientLimiter),
	}

	return request.HandlerFunc(func(c *request.Context) error {
		clientIP := c.R.RemoteAddr

		limiter.mu.Lock()
		client, exists := limiter.clients[clientIP]
		if !exists {
			client = &clientLimiter{
				tokens:    cfg.RequestsPerSecond + cfg.Burst,
				lastReset: time.Now(),
			}
			limiter.clients[clientIP] = client
		}

		// Reset tokens if second elapsed
		now := time.Now()
		if now.Sub(client.lastReset) >= time.Second {
			client.tokens = cfg.RequestsPerSecond + cfg.Burst
			client.lastReset = now
		}

		if client.tokens <= 0 {
			limiter.mu.Unlock()
			return c.Api.Error(429, "RATE_LIMITED", "Too many requests")
		}

		client.tokens--
		limiter.mu.Unlock()

		return c.Next()
	})
}

func MiddlewareFactory(params map[string]any) request.HandlerFunc {
	defConfig := DefaultConfig()
	if params == nil {
		return Middleware(defConfig)
	}

	cfg := &Config{
		RequestsPerSecond: utils.GetValueFromMap(params, PARAMS_REQUESTS_PER_SECOND, defConfig.RequestsPerSecond),
		Burst:             utils.GetValueFromMap(params, PARAMS_BURST, defConfig.Burst),
	}
	return Middleware(cfg)
}

func Register() {
	lokstra_registry.RegisterMiddlewareFactory(RATELIMIT_TYPE, MiddlewareFactory,
		lokstra_registry.AllowOverride(true))
}
```

---

## Example: Request ID Middleware

File: `middleware/request_id/request_id.go`

```go
package request_id

import (
	"github.com/google/uuid"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

const REQUEST_ID_TYPE = "request_id"
const PARAMS_HEADER_NAME = "header_name"
const PARAMS_CONTEXT_KEY = "context_key"

type Config struct {
	// HeaderName is the response header name for request ID
	HeaderName string
	
	// ContextKey is the key to store request ID in context
	ContextKey string
}

func DefaultConfig() *Config {
	return &Config{
		HeaderName: "X-Request-ID",
		ContextKey: "request_id",
	}
}

func Middleware(cfg *Config) request.HandlerFunc {
	defConfig := DefaultConfig()
	if cfg == nil {
		cfg = defConfig
	}
	if cfg.HeaderName == "" {
		cfg.HeaderName = defConfig.HeaderName
	}
	if cfg.ContextKey == "" {
		cfg.ContextKey = defConfig.ContextKey
	}

	return request.HandlerFunc(func(c *request.Context) error {
		// Check if request already has an ID
		requestID := c.R.Header.Get(cfg.HeaderName)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context
		c.Set(cfg.ContextKey, requestID)

		// Set response header
		c.W.Header().Set(cfg.HeaderName, requestID)

		return c.Next()
	})
}

func MiddlewareFactory(params map[string]any) request.HandlerFunc {
	defConfig := DefaultConfig()
	if params == nil {
		return Middleware(defConfig)
	}

	cfg := &Config{
		HeaderName: utils.GetValueFromMap(params, PARAMS_HEADER_NAME, defConfig.HeaderName),
		ContextKey: utils.GetValueFromMap(params, PARAMS_CONTEXT_KEY, defConfig.ContextKey),
	}
	return Middleware(cfg)
}

func Register() {
	lokstra_registry.RegisterMiddlewareFactory(REQUEST_ID_TYPE, MiddlewareFactory,
		lokstra_registry.AllowOverride(true))
}
```

---

## Registering Custom Middleware

### Option 1: Using Register() Function

```go
package main

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/cors"
	"github.com/primadi/lokstra/middleware/request_logger"
	"myapp/middleware/auth"
	"myapp/middleware/ratelimit"

	_ "myapp/modules/user/application"
)

func main() {
	lokstra.Bootstrap()

	// Register built-in middleware
	recovery.Register()
	request_logger.Register()
	cors.Register()

	// Register custom middleware
	auth.Register()
	ratelimit.Register()

	lokstra_init.BootstrapAndRun()
}
```

### Option 2: Inline Registration with Factory

```go
package main

import (
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

func registerMiddleware() {
	// Register inline middleware with factory function
	lokstra_registry.RegisterMiddlewareFactory("request-logger", func(config map[string]any) request.HandlerFunc {
		return func(ctx *request.Context) error {
			logger.LogInfo("→ %s %s", ctx.R.Method, ctx.R.URL.Path)
			err := ctx.Next()
			logger.LogInfo("← %s %s (status: %d)", ctx.R.Method, ctx.R.URL.Path, ctx.StatusCode())
			return err
		}
	})

	// Register auth middleware
	lokstra_registry.RegisterMiddlewareFactory("simple-auth", func(config map[string]any) request.HandlerFunc {
		tokenPrefix := "Bearer"
		if prefix, ok := config["token_prefix"].(string); ok {
			tokenPrefix = prefix
		}

		return func(ctx *request.Context) error {
			authHeader := ctx.R.Header.Get("Authorization")
			if authHeader == "" {
				return ctx.Api.Unauthorized("Missing authorization header")
			}

			expectedPrefix := tokenPrefix + " "
			if len(authHeader) < len(expectedPrefix) {
				return ctx.Api.Unauthorized("Invalid authorization format")
			}

			token := authHeader[len(expectedPrefix):]
			ctx.Set("token", token)
			ctx.Set("authenticated", true)

			return ctx.Next()
		}
	})
}
```

---

## YAML Configuration

### Global Middleware

File: `configs/middlewares.yaml`

```yaml
middlewares:
  # Built-in middleware
  - type: recovery
    params:
      enable_stack_trace: false
      enable_logging: true
  
  - type: request_logger
    params:
      enable_colors: true
      skip_paths: ["/health", "/metrics"]
  
  - type: slow_request_logger
    params:
      threshold: 500  # milliseconds
      enable_colors: true
  
  - type: cors
    params:
      allow_origins: ["*"]
  
  - type: body_limit
    params:
      max_size: 10485760  # 10MB
      skip_on_path: ["/upload/**"]
  
  - type: gzip_compression
    params:
      min_size: 1024
      compression_level: -1
  
  # Custom middleware
  - type: auth
    params:
      token_prefix: Bearer
      header_name: Authorization
  
  - type: ratelimit
    params:
      requests_per_second: 100
      burst: 10
```

---

## Per-Route Middleware

### Using @Route Decorators

```go
// @Handler name="user-handler", prefix="/api/users"
type UserHandler struct {
	// ...
}

// Only authenticated users
// @Route "GET /profile", middlewares=["auth"]
func (h *UserHandler) GetProfile(ctx *request.Context) error {
	userID := ctx.Get("user_id")
	// ...
}

// Multiple middleware: auth + admin check
// @Route "DELETE /{id}", middlewares=["auth", "admin"]
func (h *UserHandler) Delete(ctx *request.Context, id string) error {
	// ...
}

// Public endpoint (no middleware)
// @Route "POST /login"
func (h *UserHandler) Login(req *LoginRequest) (*TokenResponse, error) {
	// ...
}
```

---

## Best Practices

### 1. Use Standard Pattern

```go
✅ Always include:
   - Config struct with documented fields
   - DefaultConfig() function
   - Middleware(cfg *Config) function
   - MiddlewareFactory(params map[string]any) function
   - Register() function

❌ Don't:
   - Use struct methods for middleware
   - Use router.HandlerFunc (use request.HandlerFunc)
   - Forget to call c.Next() for passthrough
```

### 2. Middleware Order Matters

```yaml
# Correct order in YAML:
middlewares:
  - type: recovery           # 1. First (catches panics from all)
  - type: request_logger     # 2. Logs all requests
  - type: cors               # 3. Handle preflight early
  - type: body_limit         # 4. Protect before parsing
  - type: auth               # 5. Authentication
  - type: gzip_compression   # 6. Last (compress final response)
```

### 3. Fail Fast

```go
// ✅ Good: Return immediately on error
func Middleware(cfg *Config) request.HandlerFunc {
	return request.HandlerFunc(func(c *request.Context) error {
		if !isValid(c) {
			return c.Api.Unauthorized("Access denied")  // Stop here
		}
		return c.Next()  // Continue only if valid
	})
}

// ❌ Bad: Continue even on error
func Middleware(cfg *Config) request.HandlerFunc {
	return request.HandlerFunc(func(c *request.Context) error {
		if !isValid(c) {
			c.Api.Unauthorized("Access denied")  // Response sent but...
		}
		return c.Next()  // ...still continues to handler!
	})
}
```

### 4. Use Context for Request Data

```go
// ✅ Good: Use c.Set/c.Get for request-scoped data
func (c *request.Context) error {
	c.Set("user_id", "123")
	c.Set("role", "admin")
	return c.Next()
}

// In handler:
userID := ctx.Get("user_id").(string)

// ❌ Bad: Global variables for request data
var currentUserID string  // Race condition!
```

### 5. Thread-Safe Shared State

```go
// ✅ Good: Use mutex for shared state
type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientLimiter
}

func (rl *RateLimiter) check(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	// Safe access
}

// ❌ Bad: Unprotected shared state
var clients = make(map[string]int)  // Race condition!
```

---

## Testing Middleware

### Unit Test Pattern

```go
package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/stretchr/testify/assert"
	"myapp/middleware/auth"
)

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	// Create middleware
	mw := auth.Middleware(&auth.Config{
		TokenPrefix: "Bearer",
		HeaderName:  "Authorization",
	})

	// Create test context
	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()
	handlers := []request.HandlerFunc{mw}
	ctx := request.NewContext(rec, req, handlers)

	// Execute middleware
	err := mw(ctx)

	// Assert unauthorized response
	assert.Error(t, err)
	// Or check response status
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Track if next handler was called
	nextCalled := false
	nextHandler := request.HandlerFunc(func(c *request.Context) error {
		nextCalled = true
		return nil
	})

	// Create middleware
	mw := auth.Middleware(&auth.Config{
		TokenPrefix: "Bearer",
		HeaderName:  "Authorization",
	})

	// Create test context with valid token
	req := httptest.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Authorization", "Bearer valid-token-123")
	rec := httptest.NewRecorder()
	handlers := []request.HandlerFunc{mw, nextHandler}
	ctx := request.NewContext(rec, req, handlers)

	// Execute middleware
	err := mw(ctx)

	// Assert next was called
	assert.NoError(t, err)
	assert.True(t, nextCalled)
	assert.Equal(t, true, ctx.Get("authenticated"))
}
```

### Integration Test with Router

```go
func TestMiddlewareIntegration(t *testing.T) {
	// Setup router with middleware
	router := lokstra.NewRouter()
	
	router.Use(
		recovery.Middleware(&recovery.Config{}),
		auth.Middleware(&auth.Config{}),
	)
	
	router.GET("/protected", func(c *request.Context) error {
		return c.Api.Ok(map[string]any{
			"user_id": c.Get("user_id"),
		})
	})

	// Test without auth
	req1 := httptest.NewRequest("GET", "/protected", nil)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	assert.Equal(t, 401, rec1.Code)

	// Test with auth
	req2 := httptest.NewRequest("GET", "/protected", nil)
	req2.Header.Set("Authorization", "Bearer test-token")
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	assert.Equal(t, 200, rec2.Code)
}
```

---

## Common Patterns

### Conditional Middleware

```go
func Middleware(cfg *Config) request.HandlerFunc {
	return request.HandlerFunc(func(c *request.Context) error {
		// Skip for certain paths
		if shouldSkip(c.R.URL.Path, cfg.SkipPaths) {
			return c.Next()
		}

		// Apply middleware logic
		// ...

		return c.Next()
	})
}

func shouldSkip(path string, skipPaths []string) bool {
	for _, skip := range skipPaths {
		if path == skip || strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}
```

### Timing Middleware

```go
func Middleware(cfg *Config) request.HandlerFunc {
	return request.HandlerFunc(func(c *request.Context) error {
		start := time.Now()

		// Call next
		err := c.Next()

		// Log duration after handler completes
		duration := time.Since(start)
		logger.LogInfo("[%s] %s - %v", c.R.Method, c.R.URL.Path, duration)

		return err
	})
}
```

### Response Modification

```go
func Middleware(cfg *Config) request.HandlerFunc {
	return request.HandlerFunc(func(c *request.Context) error {
		// Set headers before handler
		c.W.Header().Set("X-Custom-Header", "value")

		// Call handler
		err := c.Next()

		// Add trailing headers (if supported)
		c.W.Header().Set("X-Response-Time", time.Since(start).String())

		return err
	})
}
```

---

## Checklist: Creating Middleware

- [ ] Create package in `middleware/your_middleware/`
- [ ] Define `Config` struct with documented fields
- [ ] Create `DefaultConfig()` function
- [ ] Implement `Middleware(cfg *Config) request.HandlerFunc`
- [ ] Use `request.HandlerFunc` (not `router.HandlerFunc`)
- [ ] Call `c.Next()` for passthrough
- [ ] Create `MiddlewareFactory(params map[string]any)` for YAML
- [ ] Create `Register()` function for registry
- [ ] Add constants for type and param names
- [ ] Write unit tests
- [ ] Add documentation to README

---

## Related Skills

- [implementation-lokstra-init-framework](../implementation-lokstra-init-framework/SKILL.md) - Framework setup
- [implementation-lokstra-create-handler](../implementation-lokstra-create-handler/SKILL.md) - Handler creation
- [implementation-lokstra-yaml-config](../implementation-lokstra-yaml-config/SKILL.md) - YAML configuration
