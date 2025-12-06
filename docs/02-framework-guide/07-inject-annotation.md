---
layout: default
title: "@Inject Annotation"
parent: Framework Guide
nav_order: 7
---

# @Inject Annotation

## Overview

The `@Inject` annotation marks struct fields for automatic dependency injection. It works with both `@Service` and `@RouterService` annotations.

## Basic Syntax

```go
// @Inject "service-name"
FieldName ServiceType

// or with named parameter
// @Inject service="service-name"
FieldName ServiceType

// Optional dependency
// @Inject service="service-name", optional=true
FieldName ServiceType
```

## Supported Formats

### 1. Positional Arguments

```go
// @Inject "user-repository"
UserRepo UserRepository

// @Inject "user-repository", false  (required)
UserRepo UserRepository

// @Inject "cache-service", true  (optional)
Cache CacheService
```

### 2. Named Arguments

```go
// @Inject service="user-repository"
UserRepo UserRepository

// @Inject service="cache-service", optional=true
Cache CacheService

// @Inject service="user-service", optional=false
UserSvc UserService
```

## Required Dependencies (Default)

By default, all dependencies are **required**. If the service is not found, the application will panic at startup.

```go
// @Service name="order-service"
type OrderService struct {
    // @Inject "user-repository"
    UserRepo UserRepository  // REQUIRED - panic if not found
    
    // @Inject service="product-repository"
    ProductRepo ProductRepository  // REQUIRED
}
```

**Generated code:**
```go
func RegisterOrderService() {
    lokstra_registry.RegisterLazyService("order-service", func(deps map[string]any, cfg map[string]any) any {
        return &OrderService{
            UserRepo:    lokstra_registry.GetService[UserRepository]("user-repository"),
            ProductRepo: lokstra_registry.GetService[ProductRepository]("product-repository"),
        }
    }, nil)
}
```

## Optional Dependencies

Mark dependencies as optional when they're nice-to-have but not critical:

```go
// @Service name="user-service"
type UserService struct {
    // Required
    // @Inject "user-repository"
    UserRepo UserRepository
    
    // Optional - gracefully degrades if not available
    // @Inject service="cache-service", optional=true
    Cache CacheService
    
    // Optional - feature is disabled if not available
    // @Inject service="analytics-service", optional=true
    Analytics AnalyticsService
}
```

**Generated code:**
```go
func RegisterUserService() {
    lokstra_registry.RegisterLazyService("user-service", func(deps map[string]any, cfg map[string]any) any {
        return &UserService{
            UserRepo: lokstra_registry.GetService[UserRepository]("user-repository"),
            Cache: func() CacheService {
                if svc, ok := deps["cache-service"]; ok {
                    return svc.(CacheService)
                }
                return nil
            }(),
            Analytics: func() AnalyticsService {
                if svc, ok := deps["analytics-service"]; ok {
                    return svc.(AnalyticsService)
                }
                return nil
            }(),
        }
    }, nil)
}
```

**Usage in service:**
```go
func (s *UserService) GetUser(id string) (*User, error) {
    // Check cache if available
    if s.Cache != nil {
        if user, err := s.Cache.Get(id); err == nil {
            return user.(*User), nil
        }
    }
    
    user, err := s.UserRepo.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    // Track analytics if available
    if s.Analytics != nil {
        s.Analytics.Track("user.get", id)
    }
    
    return user, nil
}
```

## Examples

### Basic Dependency Injection

```go
// @Service name="auth-service"
type AuthService struct {
    // @Inject "user-repository"
    UserRepo UserRepository
    
    // @Inject "token-service"
    TokenSvc TokenService
}

func (s *AuthService) Login(email, password string) (string, error) {
    user, err := s.UserRepo.GetByEmail(email)
    if err != nil {
        return "", err
    }
    
    token, err := s.TokenSvc.Generate(user.ID)
    return token, err
}
```

### Optional Dependencies for Fallback Behavior

```go
// @Service name="notification-service"
type NotificationService struct {
    // Primary notification channel (required)
    // @Inject "email-service"
    Email EmailService
    
    // Backup channel (optional)
    // @Inject service="sms-service", optional=true
    SMS SMSService
    
    // Monitoring (optional)
    // @Inject service="metrics-service", optional=true
    Metrics MetricsService
}

func (s *NotificationService) Send(userID, message string) error {
    // Try email first
    err := s.Email.Send(userID, message)
    
    // If email fails and SMS available, try SMS
    if err != nil && s.SMS != nil {
        err = s.SMS.Send(userID, message)
    }
    
    // Track metrics if available
    if s.Metrics != nil {
        s.Metrics.IncrementCounter("notifications.sent")
    }
    
    return err
}
```

### Feature Flags with Optional Services

```go
// @Service name="payment-service"
type PaymentService struct {
    // Core payment processor (required)
    // @Inject "payment-gateway"
    Gateway PaymentGateway
    
    // Fraud detection (optional - enable if available)
    // @Inject service="fraud-detector", optional=true
    FraudDetector FraudDetector
    
    // Loyalty points (optional - feature flag)
    // @Inject service="loyalty-service", optional=true
    Loyalty LoyaltyService
}

func (s *PaymentService) ProcessPayment(amount float64, userID string) error {
    // Check fraud if detector is available
    if s.FraudDetector != nil {
        if isFraud, _ := s.FraudDetector.Check(userID, amount); isFraud {
            return errors.New("transaction flagged as fraudulent")
        }
    }
    
    err := s.Gateway.Charge(amount)
    if err != nil {
        return err
    }
    
    // Award loyalty points if service is available
    if s.Loyalty != nil {
        s.Loyalty.AwardPoints(userID, amount*0.01)
    }
    
    return nil
}
```

## With @RouterService

```go
// @RouterService name="user-service", prefix="/api/users"
type UserServiceImpl struct {
    // @Inject "user-repository"
    UserRepo domain.UserRepository
    
    // @Inject service="cache-service", optional=true
    Cache CacheService
}

// @Route "GET /{id}"
func (s *UserServiceImpl) GetByID(p *GetUserRequest) (*User, error) {
    if s.Cache != nil {
        // Use cache if available
    }
    return s.UserRepo.GetByID(p.ID)
}
```

## Best Practices

### 1. Use Optional for Degradable Features

✅ **Good:**
```go
// @Service name="user-service"
type UserService struct {
    // Core functionality - required
    // @Inject "user-repository"
    Repo UserRepository
    
    // Performance optimization - optional
    // @Inject service="cache", optional=true
    Cache CacheService
    
    // Observability - optional
    // @Inject service="metrics", optional=true
    Metrics MetricsService
}
```

### 2. Always Check Optional Dependencies

✅ **Good:**
```go
func (s *Service) DoWork() {
    if s.Cache != nil {
        // Use cache
    }
    // Continue without cache
}
```

❌ **Bad:**
```go
func (s *Service) DoWork() {
    s.Cache.Get("key")  // Panic if Cache is nil!
}
```

### 3. Document Why Dependencies Are Optional

```go
// @Service name="order-service"
type OrderService struct {
    // @Inject "order-repository"
    OrderRepo OrderRepository
    
    // Optional: Email notifications can fail without breaking orders
    // @Inject service="email-service", optional=true
    EmailSvc EmailService
    
    // Optional: Payment processing falls back to manual processing
    // @Inject service="payment-gateway", optional=true
    PaymentGW PaymentGateway
}
```

### 4. Use Required for Critical Dependencies

❌ **Bad:**
```go
// @Service name="auth-service"
type AuthService struct {
    // DON'T make critical services optional!
    // @Inject service="user-repository", optional=true
    UserRepo UserRepository  // Auth can't work without users!
}
```

✅ **Good:**
```go
// @Service name="auth-service"
type AuthService struct {
    // Critical - required
    // @Inject "user-repository"
    UserRepo UserRepository
    
    // Nice-to-have - optional
    // @Inject service="rate-limiter", optional=true
    RateLimiter RateLimiter
}
```

## Dependency Resolution Order

When services depend on each other, Lokstra automatically resolves them in the correct order:

```go
// @Service name="a-service"
type ServiceA struct {
    // @Inject "b-service"
    B ServiceB
}

// @Service name="b-service"
type ServiceB struct {
    // @Inject "c-service"
    C ServiceC
}

// @Service name="c-service"
type ServiceC struct {
    // No dependencies
}
```

**Resolution order:** C → B → A

## Circular Dependency Detection

Lokstra detects circular dependencies at startup:

```go
// ❌ This will fail at startup
// @Service name="a"
type A struct {
    // @Inject "b"
    B B
}

// @Service name="b"
type B struct {
    // @Inject "a"  // Circular!
    A A
}
```

**Error:** `circular dependency detected: a → b → a`

## See Also

- [@Service](06-service-annotation.md) - Service registration
- [@RouterService](05-router-service-annotation.md) - HTTP services
- [@InjectCfg](08-inject-cfg-annotation.md) - Configuration injection
- [Service Registry](09-service-registry.md) - Manual service registration
