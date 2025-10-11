# Middleware Configuration Pattern

## Overview

Middlewares have their own configuration, separate from routers. This provides:
- **Clarity**: No ambiguity about which config belongs to which middleware
- **Reusability**: Same middleware can be used with different configs
- **Flexibility**: Create multiple instances of the same middleware type

## Configuration Structure

```yaml
middlewares:
  - name: logger                    # Unique name for this middleware instance
    type: request_logger            # Middleware type (factory)
    config:                         # Config specific to this middleware
      format: "json"
      log_level: "info"

routers:
  - name: user-api
    path-prefix: /api/v1
    middlewares: [logger]           # Reference middleware by name
```

## Pattern: Multiple Instances with Different Configs

### Example 1: Different Rate Limits

```yaml
middlewares:
  # Strict rate limit for public API
  - name: rate-limit-public
    type: rate_limit
    config:
      max_requests_per_minute: 10
      burst: 2
      
  # Relaxed rate limit for internal API
  - name: rate-limit-internal
    type: rate_limit
    config:
      max_requests_per_minute: 1000
      burst: 100
      
  # VIP rate limit
  - name: rate-limit-vip
    type: rate_limit
    config:
      max_requests_per_minute: 10000
      burst: 500

routers:
  - name: public-api
    path-prefix: /api/public
    middlewares: [rate-limit-public]
    
  - name: internal-api
    path-prefix: /api/internal
    middlewares: [rate-limit-internal]
    
  - name: vip-api
    path-prefix: /api/vip
    middlewares: [rate-limit-vip]
```

### Example 2: Different Logging Levels

```yaml
middlewares:
  # Debug logging for development
  - name: logger-debug
    type: request_logger
    config:
      log_level: "debug"
      log_request_body: true
      log_response_body: true
      
  # Info logging for production
  - name: logger-info
    type: request_logger
    config:
      log_level: "info"
      log_request_body: false
      log_response_body: false
      
  # Sensitive data logging (for payment)
  - name: logger-payment
    type: request_logger
    config:
      log_level: "debug"
      log_request_body: true
      log_response_body: true
      mask_fields: ["card_number", "cvv", "pin"]

routers:
  - name: payment-api
    middlewares: [logger-payment]
    
  - name: user-api
    middlewares: [logger-info]
```

### Example 3: Different Timeout Durations

```yaml
middlewares:
  # Short timeout for fast APIs
  - name: timeout-fast
    type: timeout
    config:
      duration: 5s
      
  # Medium timeout for standard APIs
  - name: timeout-standard
    type: timeout
    config:
      duration: 30s
      
  # Long timeout for analytics
  - name: timeout-analytics
    type: timeout
    config:
      duration: 120s

routers:
  - name: user-api
    middlewares: [timeout-fast]
    
  - name: order-api
    middlewares: [timeout-standard]
    
  - name: analytics-api
    middlewares: [timeout-analytics]
```

## Middleware Composition

### Stacking Multiple Middlewares

```yaml
middlewares:
  - name: logger
    type: request_logger
    config:
      format: "json"
      
  - name: cors
    type: cors
    config:
      allowed_origins: ["*"]
      allowed_methods: ["GET", "POST", "PUT", "DELETE"]
      
  - name: auth-jwt
    type: jwt_auth
    config:
      secret: "your-secret-key"
      algorithm: "HS256"
      
  - name: rate-limiter
    type: rate_limit
    config:
      max_requests_per_minute: 100

routers:
  - name: api-router
    path-prefix: /api/v1
    middlewares: [logger, cors, auth-jwt, rate-limiter]  # Applied in order
```

**Execution Order:**
1. `logger` - Logs request
2. `cors` - Handles CORS headers
3. `auth-jwt` - Validates JWT token
4. `rate-limiter` - Checks rate limit
5. **Handler** - Your route handler
6. `rate-limiter` (response)
7. `auth-jwt` (response)
8. `cors` (response)
9. `logger` (response) - Logs response

## Common Middleware Configurations

### 1. Request Logger

```yaml
middlewares:
  - name: logger
    type: request_logger
    config:
      format: "json"              # or "text"
      log_level: "info"           # debug, info, warn, error
      log_request_body: false
      log_response_body: false
      log_headers: true
      exclude_paths: ["/health", "/metrics"]
```

### 2. CORS

```yaml
middlewares:
  - name: cors
    type: cors
    config:
      allowed_origins: ["https://myapp.com"]
      allowed_methods: ["GET", "POST", "PUT", "DELETE"]
      allowed_headers: ["Content-Type", "Authorization"]
      expose_headers: ["X-Total-Count"]
      allow_credentials: true
      max_age: 3600
```

### 3. Rate Limiting

```yaml
middlewares:
  - name: rate-limiter
    type: rate_limit
    config:
      max_requests_per_minute: 60
      burst: 10                   # Allow bursts
      key_func: "ip"              # ip, user, api_key
      exclude_paths: ["/health"]
```

### 4. Authentication

```yaml
middlewares:
  - name: jwt-auth
    type: jwt_auth
    config:
      secret: "your-secret-key"
      algorithm: "HS256"          # or RS256, ES256
      token_lookup: "header:Authorization"
      token_prefix: "Bearer "
      skip_paths: ["/login", "/register"]
```

### 5. Timeout

```yaml
middlewares:
  - name: timeout
    type: timeout
    config:
      duration: 30s
      message: "Request timeout"
```

### 6. Body Limit

```yaml
middlewares:
  - name: body-limit
    type: body_limit
    config:
      max_size: 1048576           # 1MB in bytes
      error_message: "Request body too large"
```

### 7. Compression

```yaml
middlewares:
  - name: gzip
    type: gzip_compression
    config:
      level: 5                    # 1-9, balance speed vs compression
      min_size: 1024              # Don't compress small responses
```

## Environment-Specific Middleware

### Development

```yaml
middlewares:
  - name: logger-dev
    type: request_logger
    config:
      log_level: "debug"
      log_request_body: true
      log_response_body: true
      pretty_print: true
      
  - name: cors-dev
    type: cors
    config:
      allowed_origins: ["*"]      # Permissive for development
```

### Production

```yaml
middlewares:
  - name: logger-prod
    type: request_logger
    config:
      log_level: "info"
      log_request_body: false
      log_response_body: false
      structured_logging: true
      
  - name: cors-prod
    type: cors
    config:
      allowed_origins: ["https://myapp.com"]  # Strict for production
```

## Best Practices

### 1. Name Middlewares Clearly

❌ **Bad:**
```yaml
middlewares:
  - name: mw1
    type: rate_limit
  - name: mw2
    type: rate_limit
```

✅ **Good:**
```yaml
middlewares:
  - name: rate-limit-public
    type: rate_limit
  - name: rate-limit-internal
    type: rate_limit
```

### 2. Group Related Middlewares

```yaml
# Authentication middlewares
middlewares:
  - name: auth-jwt
    type: jwt_auth
    
  - name: auth-api-key
    type: api_key_auth
    
  - name: auth-oauth
    type: oauth_auth

# Rate limiting middlewares
  - name: rate-limit-strict
    type: rate_limit
    
  - name: rate-limit-relaxed
    type: rate_limit
```

### 3. Use Descriptive Configs

```yaml
middlewares:
  - name: rate-limit-payment
    type: rate_limit
    config:
      max_requests_per_minute: 30    # Lower for security-sensitive endpoints
      burst: 5
      comment: "Strict limit for payment processing"
```

### 4. Reuse Common Middlewares

```yaml
middlewares:
  - name: logger
    type: request_logger
    config:
      format: "json"

routers:
  - name: user-api
    middlewares: [logger]        # Reuse
    
  - name: order-api
    middlewares: [logger]        # Reuse
    
  - name: payment-api
    middlewares: [logger]        # Reuse
```

## Advanced Patterns

### Conditional Middleware (Future)

```yaml
middlewares:
  - name: auth
    type: jwt_auth
    config:
      skip_if:
        path_starts_with: ["/public"]
        method: ["OPTIONS"]
```

### Middleware Chains (Future)

```yaml
middleware-chains:
  - name: standard-chain
    middlewares: [logger, cors, rate-limiter]
    
  - name: secure-chain
    middlewares: [logger, cors, auth, rate-limiter]

routers:
  - name: public-api
    middleware-chain: standard-chain
    
  - name: private-api
    middleware-chain: secure-chain
```

## Summary

### Key Principles

1. **Middleware config belongs to middleware** - Not router, not service
2. **One middleware type, many instances** - Different configs for different use cases
3. **Clear naming** - Name reflects purpose (rate-limit-public vs rate-limit-internal)
4. **Reusability** - Same middleware can be referenced by multiple routers
5. **Composition** - Stack multiple middlewares for complex behavior

### Configuration Flow

```
1. Define middleware instances with their configs
   ↓
2. Reference middlewares in router definitions
   ↓
3. Framework applies middlewares in order
   ↓
4. Middlewares use their individual configs
```

### No Ambiguity

❌ **Wrong (ambiguous):**
```yaml
routers:
  - name: api
    middlewares: [rate-limiter]
    config:
      max_requests: 100        # Which middleware is this for?
```

✅ **Right (clear):**
```yaml
middlewares:
  - name: rate-limiter
    type: rate_limit
    config:
      max_requests_per_minute: 100  # Clear: belongs to rate-limiter

routers:
  - name: api
    middlewares: [rate-limiter]     # Just reference
```

## Related Documentation

- [Configuration Strategies](./configuration-strategies.md)
- [Router Configuration Improvements](./router-config-improvements.md)
- [Middleware Development Guide](./middleware-development.md) (TODO)
