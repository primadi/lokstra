---
layout: default
title: "@Service Annotation"
parent: Framework Guide
nav_order: 6
---

# @Service Annotation

## Overview

The `@Service` annotation is used to register **pure service classes** (non-HTTP services) with automatic dependency injection and configuration injection. This is ideal for:

- Business logic services
- Infrastructure services (email, SMS, etc.)
- Helper/utility services
- Background workers
- Any service that doesn't expose HTTP endpoints

**Key Differences:**
- `@Handler` → HTTP handlers (controllers) with routes
- `@Service` → Pure services without HTTP endpoints

## Basic Syntax

```go
// @Service name="service-name"
type MyService struct {
    // Fields with dependency injection
}
```

## Features

### 1. Service Registration

**Positional argument:**
```go
// @Service "auth-service"
type AuthService struct {
    // ...
}
```

**Named argument:**
```go
// @Service name="auth-service"
type AuthService struct {
    // ...
}
```

### 2. Dependency Injection with @Inject

**Basic dependency:**
```go
// @Service name="auth-service"
type AuthService struct {
    // @Inject "user-repository"
    UserRepo UserRepository
    
    // @Inject service="token-service"
    TokenSvc TokenService
}
```

**Optional dependency:**
```go
// @Service name="auth-service"
type AuthService struct {
    // @Inject service="cache-service", optional=true
    Cache CacheService  // nil if cache-service not found
}
```

**Generated code:**
```go
func RegisterAuthService() {
    lokstra_registry.RegisterLazyService("auth-service", func(deps map[string]any, cfg map[string]any) any {
        return &AuthService{
            UserRepo: lokstra_registry.GetService[UserRepository]("user-repository"),
            TokenSvc: lokstra_registry.GetService[TokenService]("token-service"),
            Cache: func() CacheService {
                if svc, ok := deps["cache-service"]; ok {
                    return svc.(CacheService)
                }
                return nil
            }(),
        }
    }, nil)
}
```

### 3. Configuration Injection with @Inject "cfg:..."

**Basic config:**
```go
// @Service name="auth-service"
type AuthService struct {
    // @Inject "cfg:auth.jwt-secret"
    JwtSecret string
    
    // @Inject "cfg:auth.token-expiry"
    TokenExpiry time.Duration
}
```

**With default values:**
```go
// @Service name="email-service"
type EmailService struct {
    // @Inject "cfg:smtp.host", "localhost"
    SMTPHost string
    
    // @Inject "cfg:smtp.port", "587"
    SMTPPort int
    
    // @Inject "cfg:smtp.enabled", "true"
    Enabled bool
}
```

**Generated code:**
```go
func RegisterEmailService() {
    lokstra_registry.RegisterLazyService("email-service", func(deps map[string]any, cfg map[string]any) any {
        return &EmailService{
            SMTPHost:  lokstra_registry.GetConfig("smtp.host", "localhost"),
            SMTPPort:  lokstra_registry.GetConfigInt("smtp.port", 587),
            Enabled:   lokstra_registry.GetConfigBool("smtp.enabled", true),
        }
    }, nil)
}
```

**Supported types (auto-detected):**
- `string` → `GetConfig`
- `int`, `int8`, `int16`, `int32`, `int64`, `uint*` → `GetConfigInt`
- `bool` → `GetConfigBool`
- `float32`, `float64` → `GetConfigFloat`
- `time.Duration` → `GetConfigDuration`

## Complete Example

### Service Definition

```go
package application

import (
    "time"
    "myapp/domain"
)

// @Service name="auth-service"
type AuthService struct {
    // Required dependency
    // @Inject "user-repository"
    UserRepo domain.UserRepository
    
    // Optional dependency (nil if not found)
    // @Inject service="cache-service", optional=true
    Cache domain.CacheService
    
    // Required config (no default)
    // @Inject "cfg:auth.jwt-secret"
    JwtSecret string
    
    // Config with defaults
    // @Inject "cfg:auth.token-expiry", "24h"
    TokenExpiry time.Duration
    
    // @Inject "cfg:auth.max-attempts", "5"
    MaxAttempts int
    
    // @Inject "cfg:auth.debug-mode", "false"
    DebugMode bool
}

func (s *AuthService) Login(email, password string) (string, error) {
    // Check cache if available
    if s.Cache != nil {
        if cachedUser, err := s.Cache.Get("user:" + email); err == nil {
            // Use cached user
        }
    }
    
    user, err := s.UserRepo.GetByEmail(email)
    if err != nil {
        return "", err
    }
    
    // Verify password, generate token, etc.
    token := s.generateToken(user.ID, s.TokenExpiry)
    
    if s.DebugMode {
        println("Login successful:", email)
    }
    
    return token, nil
}
```

### Configuration (config.yaml)

```yaml
configs:
  auth:
    jwt-secret: "your-secret-key-here"
    token-expiry: "48h"
    max-attempts: 3
    debug-mode: false
  smtp:
    host: "smtp.gmail.com"
    port: 587
    enabled: true

service-definitions:
  user-repository:
    type: user-repository-factory
    
  cache-service:
    type: redis-cache-factory
    
  auth-service:
    # Auto-registered via @Service annotation
    # Dependencies: user-repository, cache-service (optional)
```

### Generated Code

Running `lokstra.Bootstrap()` generates:

```go
// zz_generated.lokstra.go
package application

import (
    "github.com/primadi/lokstra/lokstra_registry"
    "myapp/domain"
)

func init() {
    RegisterAuthService()
}

// RegisterAuthService registers the auth-service with the registry
// Auto-generated from annotations:
//   - @Service name="auth-service"
//   - @Inject annotations
func RegisterAuthService() {
    lokstra_registry.RegisterLazyService("auth-service", func(deps map[string]any, cfg map[string]any) any {
        return &AuthService{
            UserRepo: lokstra_registry.GetService[domain.UserRepository]("user-repository"),
            Cache: func() domain.CacheService {
                if svc, ok := deps["cache-service"]; ok {
                    return svc.(domain.CacheService)
                }
                return nil
            }(),
            JwtSecret:   lokstra_registry.GetConfig("auth.jwt-secret", ""),
            TokenExpiry: lokstra_registry.GetConfigDuration("auth.token-expiry", 24*time.Hour),
            MaxAttempts: lokstra_registry.GetConfigInt("auth.max-attempts", 5),
            DebugMode:   lokstra_registry.GetConfigBool("auth.debug-mode", false),
        }
    }, nil)
}
```

## Main Application

```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/lokstra_registry"
    
    // Import packages with @Service annotations
    _ "myapp/application"
    _ "myapp/infrastructure"
)

func main() {
    // Auto-generates code when @Service changes detected
    lokstra.Bootstrap()
    
    // Start server from config
    lokstra_registry.RunServerFromConfig()
}
```

## Best Practices

### 1. Use @Service for Pure Services

✅ **Good:**
```go
// @Service name="email-service"
type EmailService struct {
    // Pure service, no HTTP endpoints
}

// @Service name="payment-processor"
type PaymentProcessor struct {
    // Business logic only
}
```

❌ **Bad:**
```go
// Don't use @Service for HTTP controllers
// Use @Handler instead
```

### 2. Separate Configuration Concerns

✅ **Good:**
```go
// @Service name="sms-service"
type SMSService struct {
    // @Inject "cfg:sms.api-key"
    APIKey string
    
    // @Inject "cfg:sms.endpoint", "https://api.sms.com"
    Endpoint string
}
```

### 3. Use Optional for Nice-to-Have Dependencies

✅ **Good:**
```go
// @Service name="user-service"
type UserService struct {
    // Required
    // @Inject "user-repository"
    Repo UserRepository
    
    // Optional - degrade gracefully
    // @Inject service="cache", optional=true
    Cache CacheService
    
    // Optional - feature flag
    // @Inject service="analytics", optional=true
    Analytics AnalyticsService
}

func (s *UserService) GetUser(id string) (*User, error) {
    // Check cache if available
    if s.Cache != nil {
        if user, err := s.Cache.Get(id); err == nil {
            return user.(*User), nil
        }
    }
    
    user, err := s.Repo.GetByID(id)
    
    // Track analytics if available
    if s.Analytics != nil {
        s.Analytics.Track("user.get", id)
    }
    
    return user, err
}
```

### 4. Type-Safe Config Injection

✅ **Good:**
```go
// @Service name="config-service"
type ConfigService struct {
    // @Inject "cfg:server.port", "8080"
    Port int  // Auto-uses GetConfigInt
    
    // @Inject "cfg:cache.ttl", "5m"
    CacheTTL time.Duration  // Auto-uses GetConfigDuration
    
    // @Inject "cfg:debug", "false"
    Debug bool  // Auto-uses GetConfigBool
}
```

## Comparison: @Service vs @Handler

| Feature | @Service | @Handler |
|---------|----------|----------------|
| HTTP Routes | ❌ No | ✅ Yes (@Route) |
| Dependency Injection | ✅ @Inject | ✅ @Inject |
| Config Injection | ✅ @Inject "cfg:..." | ✅ @Inject "cfg:..." |
| Optional Dependencies | ✅ Yes | ✅ Yes |
| Use Case | Business logic, utilities | HTTP controllers |
| Generated Code | `RegisterLazyService` | `RegisterRouterServiceType` |

## Code Generation

### Manual Generation

```bash
lokstra autogen .
```

### Automatic Generation (Recommended)

```go
func main() {
    lokstra.Bootstrap()  // Auto-generates when changes detected
    // ...
}
```

### Force Rebuild

```bash
go run . --generate-only
```

## See Also

- [@Handler](05-router-service-annotation.md) - For HTTP endpoints
- [@Inject](07-inject-annotation.md) - Dependency injection details
- [@Inject "cfg:..."](08-inject-cfg-annotation.md) - Configuration injection
- [Service Registry](09-service-registry.md) - Manual service registration
