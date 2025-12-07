---
layout: default
title: AI Agent Guide - Lokstra Framework
description: Comprehensive guide for AI agents to understand and use Lokstra Framework effectively
---

# AI Agent Guide - Lokstra Framework

**Version:** 1.0.0  
**Last Updated:** November 12, 2025  
**Target:** AI Agents (GitHub Copilot, Claude, ChatGPT, etc.)

This guide helps AI agents understand how to assist programmers using the **Lokstra Framework** (`github.com/primadi/lokstra`). It provides structured patterns, complete code examples, and best practices for building Go web applications.

---

## Table of Contents

1. [Framework Overview](#framework-overview)
2. [Core Concepts](#core-concepts)
3. [Quick Start Patterns](#quick-start-patterns)
4. [Router Patterns](#router-patterns)
5. [Service Patterns](#service-patterns)
6. [Configuration YAML](#configuration-yaml)
7. [Annotation System](#annotation-system)
8. [Middleware Usage](#middleware-usage)
9. [Dependency Injection](#dependency-injection)
10. [Project Structure Templates](#project-structure-templates)
11. [Common Patterns & Idioms](#common-patterns--idioms)
12. [Troubleshooting](#troubleshooting)

---

## Framework Overview

### What is Lokstra?

Lokstra is a **versatile Go web framework** with two usage modes:

1. **Router Mode** (Like Echo, Gin, Chi)
   - Simple HTTP routing
   - Flexible handler signatures (29+ forms)
   - Middleware support
   - No DI, no config files required

2. **Framework Mode** (Like NestJS, Spring Boot)
   - Lazy dependency injection (type-safe)
   - Auto-generated REST routers
   - YAML-driven configuration
   - Multi-deployment support (monolith → microservices)

### Key Design Principles

- **Type-safe DI** with direct type assertions
- **Service-level lazy loading** (services created on first access)
- **Eager dependency resolution** (dependencies loaded when service created)
- **Zero reflection** in hot path
- **Configuration-driven** deployment (optional YAML)
- **Flexible handler signatures** (supports many parameter combinations)

---

## Core Concepts

### 1. Router

HTTP request routing with flexible handler signatures.

```go
import "github.com/primadi/lokstra"

r := lokstra.NewRouter("api")

// Various handler signatures supported
r.GET("/hello", func() string {
    return "Hello, World!"
})

r.GET("/user/{id}", func(id string) (map[string]any, error) {
    return map[string]any{"id": id, "name": "John"}, nil
})

r.POST("/users", func(ctx *request.Context, params *CreateUserParams) error {
    // params automatically validated
    return ctx.Api.Ok(params)
})
```

### 2. App

Application wrapper with graceful shutdown.

```go
app := lokstra.NewApp("myapp", ":8080", router)
app.Run(30 * time.Second) // 30s graceful shutdown timeout
```

### 3. Service

Business logic component with dependency injection.

```go
type UserService struct {
    UserRepo UserRepository
}

func (s *UserService) GetByID(id string) (*User, error) {
    return s.UserRepo.GetByID(id)
}
```

### 4. Repository

Data access layer (interface pattern).

```go
type UserRepository interface {
    GetByID(id string) (*User, error)
    List() ([]*User, error)
    Create(user *User) (*User, error)
    Update(user *User) (*User, error)
    Delete(id string) error
}
```

### 5. Domain Models

Data structures with validation tags.

```go
type User struct {
    ID    string `json:"id"`
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=18,max=120"`
}
```

---

## Quick Start Patterns

### Pattern 1: Simple Router (No DI, No Config)

**Use Case:** Learning, simple APIs, minimal setup

```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/middleware/cors"
    "github.com/primadi/lokstra/middleware/recovery"
    "time"
)

func main() {
    r := lokstra.NewRouter("api")
    
    // Middleware
    r.Use(recovery.Middleware(nil))
    r.Use(cors.Middleware([]string{"*"}))
    
    // Routes
    r.GET("/", func() string {
        return "Hello, Lokstra!"
    })
    
    r.GET("/ping", func() string {
        return "pong"
    })
    
    r.GET("/users", func() []map[string]any {
        return []map[string]any{
            {"id": "1", "name": "Alice"},
            {"id": "2", "name": "Bob"},
        }
    })
    
    // Run
    app := lokstra.NewApp("simple", ":8080", r)
    app.Run(30 * time.Second)
}
```

### Pattern 2: With Annotation-Based Services (Recommended)

**Use Case:** Production apps, team projects, scalable architecture

**File: `main.go`**
```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/lokstra_registry"
    
    // Import packages with @RouterService annotations
    _ "myapp/modules/user/application"
    _ "myapp/modules/order/application"
)

func main() {
    // Auto-generates code when @RouterService changes detected
    lokstra.Bootstrap()
    
    deploy.SetLogLevelFromEnv() // LOKSTRA_LOG_LEVEL=debug
    
    // Services auto-registered via annotations!
    lokstra_registry.RunServerFromConfig()
}
```

**File: `modules/user/application/user_service.go`**
```go
package application

import (
    "myapp/modules/user/domain"
    "myapp/modules/user/infrastructure"
)

// @RouterService name="user-service", prefix="/api/users"
type UserServiceImpl struct {
    // @Inject "user-repository"
    UserRepo domain.UserRepository
}

// @Route "GET /"
func (s *UserServiceImpl) List(p *domain.ListUsersParams) ([]*domain.User, error) {
    return s.UserRepo.List()
}

// @Route "GET /{id}"
func (s *UserServiceImpl) GetByID(p *domain.GetUserParams) (*domain.User, error) {
    return s.UserRepo.GetByID(p.ID)
}

// @Route "POST /"
func (s *UserServiceImpl) Create(p *domain.CreateUserParams) (*domain.User, error) {
    u := &domain.User{
        Name:  p.Name,
        Email: p.Email,
    }
    return s.UserRepo.Create(u)
}

// @Route "PUT /{id}"
func (s *UserServiceImpl) Update(p *domain.UpdateUserParams) (*domain.User, error) {
    u := &domain.User{
        ID:    p.ID,
        Name:  p.Name,
        Email: p.Email,
    }
    return s.UserRepo.Update(u)
}

// @Route "DELETE /{id}"
func (s *UserServiceImpl) Delete(p *domain.DeleteUserParams) error {
    return s.UserRepo.Delete(p.ID)
}
```

**Generate code (automatic with Bootstrap):**
```go
func main() {
    lokstra.Bootstrap() // Auto-generates on every run when changes detected
    // ...
}
```

**Manual generation (optional):**
```bash
# Before build/deploy to ensure latest code
lokstra autogen .

# Or force rebuild all
go run . --generate-only
```

**File: `config.yaml`**
```yaml
service-definitions:
  user-repository:
    type: user-repository-factory
    config:
      dsn: "memory://users"
  
  user-service:
    # Type auto-registered via @RouterService annotation
    depends-on:
      - user-repository

deployments:
  development:
    servers:
      api:
        base-url: "http://localhost:8080"
        addr: ":8080"
        published-services:
          - user-service
```

### Pattern 3: Manual Service Registration (Legacy/Advanced)

**Use Case:** Complex custom factories, special initialization logic

**File: `main.go`**
```go
package main

import (
    "github.com/primadi/lokstra/core/deploy"
    "github.com/primadi/lokstra/lokstra_registry"
)

func main() {
    deploy.SetLogLevelFromEnv() // LOKSTRA_LOG_LEVEL=debug
    
    registerServiceTypes()
    registerMiddlewareTypes()
    
    lokstra_registry.RunServerFromConfig()
}
```

**File: `register.go`**
```go
package main

import (
    "github.com/primadi/lokstra/lokstra_registry"
    "myapp/service"
    "myapp/repository"
)

func registerServiceTypes() {
    // Register repository
    lokstra_registry.RegisterServiceType(
        "user-repository-factory",
        repository.NewUserRepositoryFactory,
        nil,
    )
    
    // Register service
    lokstra_registry.RegisterServiceType(
        "user-service-factory",
        service.UserServiceFactory,
        service.UserServiceRemoteFactory,
    )
}

func registerMiddlewareTypes() {
    // Built-in middleware auto-registered
    // Custom middleware registration here
}
```

**File: `config.yaml`**
```yaml
service-definitions:
  user-repository:
    type: user-repository-factory
    config:
      dsn: "memory://users"
  
  user-service:
    type: user-service-factory
    depends-on:
      - user-repository

deployments:
  development:
    servers:
      api:
        base-url: "http://localhost:8080"
        addr: ":8080"
        published-services:
          - user-service
```

---

## Router Patterns

### Handler Signature Variations

Lokstra supports **29+ handler signatures**. Here are the most common:

```go
// 1. Simple return
r.GET("/hello", func() string {
    return "Hello"
})

// 2. Return with error
r.GET("/user/{id}", func(id string) (string, error) {
    if id == "" {
        return "", errors.New("id required")
    }
    return "User: " + id, nil
})

// 3. Struct response
r.GET("/user/{id}", func(id string) (*User, error) {
    return &User{ID: id, Name: "John"}, nil
})

// 4. Context access
r.GET("/header", func(ctx *request.Context) error {
    token := ctx.Req.Header.Get("Authorization")
    return ctx.Api.Ok(map[string]string{"token": token})
})

// 5. Path parameters
r.GET("/posts/{postId}/comments/{commentId}", 
    func(postId, commentId string) (string, error) {
        return fmt.Sprintf("Post: %s, Comment: %s", postId, commentId), nil
    })

// 6. Request body binding with validation
type CreateUserParams struct {
    Name  string `json:"name" validate:"required,min=3"`
    Email string `json:"email" validate:"required,email"`
}

r.POST("/users", func(ctx *request.Context, params *CreateUserParams) error {
    // params already validated!
    return ctx.Api.Ok(params)
})

// 7. Multiple parameters
r.PUT("/users/{id}", func(ctx *request.Context, id string, params *UpdateUserParams) error {
    params.ID = id
    user := updateUser(params)
    return ctx.Api.Ok(user)
})

// 8. Query parameters
type SearchParams struct {
    Query string `query:"q" validate:"required"`
    Page  int    `query:"page" validate:"min=1"`
    Limit int    `query:"limit" validate:"min=1,max=100"`
}

r.GET("/search", func(params *SearchParams) ([]Result, error) {
    return searchResults(params.Query, params.Page, params.Limit), nil
})
```

### Router Groups and Versioning

```go
r := lokstra.NewRouter("api")

// Group routes
v1 := r.Group("/v1")
v1.GET("/users", listUsersV1)
v1.POST("/users", createUserV1)

v2 := r.Group("/v2")
v2.GET("/users", listUsersV2)
v2.POST("/users", createUserV2)

// Nested groups
api := r.Group("/api")
admin := api.Group("/admin")
admin.GET("/stats", getStats)
```

### Middleware Application

```go
import (
    "github.com/primadi/lokstra/middleware/recovery"
    "github.com/primadi/lokstra/middleware/request_logger"
    "github.com/primadi/lokstra/middleware/cors"
)

r := lokstra.NewRouter("api")

// Global middleware (all routes)
r.Use(recovery.Middleware(nil))
r.Use(request_logger.Middleware(nil))

// Group middleware
authorized := r.Group("/admin")
authorized.Use(authMiddleware)
authorized.GET("/users", listUsers)

// Route-specific middleware
r.GET("/public", publicHandler)
r.GET("/private", privateHandler, authMiddleware, logMiddleware)
```

---

## Service Patterns

### Service Factory Pattern

**Service implementation:**

```go
package service

import (
    "github.com/primadi/lokstra/core/service"
    "myapp/domain/user"
)

type UserServiceImpl struct {
    UserRepo user.UserRepository
}

var _ user.UserService = (*UserServiceImpl)(nil)

func (s *UserServiceImpl) GetByID(p *user.GetUserParams) (*user.User, error) {
    return s.UserRepo.GetByID(p.ID)
}

func (s *UserServiceImpl) List(p *user.ListUsersParams) ([]*user.User, error) {
    return s.UserRepo.List()
}

func (s *UserServiceImpl) Create(p *user.CreateUserParams) (*user.User, error) {
    u := &user.User{
        Name:  p.Name,
        Email: p.Email,
    }
    return s.UserRepo.Create(u)
}

// Factory function for local deployment
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserServiceImpl{
        UserRepo: deps["user-repository"].(user.UserRepository),
    }
}

// Factory function for remote deployment (microservices)
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
    proxyService, _ := config["remote"].(*proxy.Service)
    return NewUserServiceRemote(proxyService)
}
```

### Repository Factory Pattern

```go
package repository

import "myapp/domain/user"

type UserRepositoryInMemory struct {
    users map[string]*user.User
}

func (r *UserRepositoryInMemory) GetByID(id string) (*user.User, error) {
    if u, ok := r.users[id]; ok {
        return u, nil
    }
    return nil, errors.New("user not found")
}

func (r *UserRepositoryInMemory) List() ([]*user.User, error) {
    result := make([]*user.User, 0, len(r.users))
    for _, u := range r.users {
        result = append(result, u)
    }
    return result, nil
}

// Factory function
func NewUserRepositoryFactory(deps map[string]any, config map[string]any) any {
    return &UserRepositoryInMemory{
        users: make(map[string]*user.User),
    }
}
```

### Domain Interface Pattern

**File: `domain/user/repository.go`**
```go
package user

type UserRepository interface {
    GetByID(id string) (*User, error)
    List() ([]*User, error)
    Create(user *User) (*User, error)
    Update(user *User) (*User, error)
    Delete(id string) error
}
```

**File: `domain/user/service.go`**
```go
package user

type UserService interface {
    GetByID(p *GetUserParams) (*User, error)
    List(p *ListUsersParams) ([]*User, error)
    Create(p *CreateUserParams) (*User, error)
    Update(p *UpdateUserParams) (*User, error)
    Delete(p *DeleteUserParams) error
}
```

**File: `domain/user/models.go`**
```go
package user

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type GetUserParams struct {
    ID string `path:"id" validate:"required"`
}

type ListUsersParams struct {
    Page  int `query:"page" validate:"min=1"`
    Limit int `query:"limit" validate:"min=1,max=100"`
}

type CreateUserParams struct {
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
}

type UpdateUserParams struct {
    ID    string `path:"id" validate:"required"`
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
}

type DeleteUserParams struct {
    ID string `path:"id" validate:"required"`
}
```

---

## Configuration YAML

### Complete Schema Reference

```yaml
# YAML Schema: https://primadi.github.io/lokstra/schema/lokstra.schema.json

# Middleware definitions (optional, built-in middleware auto-registered)
middleware-definitions:
  recovery:
    type: recovery
    config:
      enable_stack_trace: false
      enable_logging: true
  
  request-logger:
    type: request-logger
    config:
      prefix: "API"
      enable_colors: true
      skip_paths: ["/health", "/metrics"]
  
  cors:
    type: cors
    config:
      allow_origins: ["*"]
  
  body-limit:
    type: body-limit
    config:
      max_size: 10485760  # 10MB
      skip_on_path: ["/upload/**"]

# Service definitions
service-definitions:
  # Repository layer
  user-repository:
    type: user-repository-factory
    config:
      dsn: "postgres://localhost/mydb"
  
  # Service layer
  user-service:
    type: user-service-factory
    depends-on:
      - user-repository
    config:
      cache_ttl: 300
    router:
      path-prefix: /api/v1
      middlewares:
        - recovery
        - request-logger
      hidden:  # Hide specific methods from auto-router
        - InternalMethod
        - PrivateHelper

# Router definitions (advanced path rewriting)
router-definitions:
  user-router:
    path-prefix: /api
    path-rewrites:
      - pattern: "^/api/v1/(.*)$"
        replacement: "/api/v2/$1"
    middlewares:
      - cors
      - body-limit
    hidden:
      - DeprecatedMethod

# Deployment configurations
deployments:
  # Development environment (monolith)
  development:
    servers:
      api-server:
        base-url: "http://localhost:8080"
        addr: ":8080"
        published-services:
          - user-service
          - order-service
          - payment-service
  
  # Staging environment
  staging:
    servers:
      api-server:
        base-url: "https://staging-api.example.com"
        addr: ":8080"
        published-services:
          - user-service
          - order-service
  
  # Production microservices
  production:
    servers:
      user-api:
        base-url: "https://user-api.example.com"
        addr: ":8001"
        published-services:
          - user-service
      
      order-api:
        base-url: "https://order-api.example.com"
        addr: ":8002"
        published-services:
          - order-service
      
      payment-api:
        base-url: "https://payment-api.example.com"
        addr: ":8003"
        published-services:
          - payment-service
```

### Environment-Specific Config

**Development:**
```yaml
deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services: [user-service, order-service]
```

**Production (Microservices):**
```yaml
deployments:
  production:
    servers:
      user-api:
        addr: ":8001"
        base-url: "https://user.example.com"
        published-services: [user-service]
      
      order-api:
        addr: ":8002"
        base-url: "https://order.example.com"
        published-services: [order-service]
```

**Run specific deployment:**
```bash
LOKSTRA_DEPLOYMENT=production go run .
```

---

## Annotation System

### @RouterService Annotation (HTTP Controllers)

Generate REST routers automatically from service methods. Use for services that expose HTTP endpoints.

```go
package application

import (
    "github.com/primadi/lokstra/core/service"
    "myapp/domain"
)

// @RouterService name="user-service", prefix="/api", middlewares=["recovery", "request-logger"]
type UserServiceImpl struct {
    // @Inject "user-repository"
    UserRepo domain.UserRepository
    
    // @Inject service="cache-service", optional=true
    Cache domain.CacheService
}

// @Route "GET /users/{id}"
func (s *UserServiceImpl) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
    return s.UserRepo.GetByID(p.ID)
}

// @Route "GET /users"
func (s *UserServiceImpl) List(p *domain.ListUsersRequest) ([]*domain.User, error) {
    return s.UserRepo.List()
}

// @Route "POST /users"
func (s *UserServiceImpl) Create(p *domain.CreateUserRequest) (*domain.User, error) {
    u := &domain.User{
        Name:   p.Name,
        Email:  p.Email,
    }
    return s.UserRepo.Create(u)
}

// @Route "PUT /users/{id}", middlewares=["auth", "admin"]
func (s *UserServiceImpl) Update(p *domain.UpdateUserRequest) (*domain.User, error) {
    u := &domain.User{
        ID:    p.ID,
        Name:  p.Name,
        Email: p.Email,
    }
    return s.UserRepo.Update(u)
}

// @Route "DELETE /users/{id}", middlewares=["auth", "admin"]
func (s *UserServiceImpl) Delete(p *domain.DeleteUserRequest) error {
    return s.UserRepo.Delete(p.ID)
}

func Register() {
    // Package auto-loaded by code generation
}
```

### @Service Annotation (Pure Services)

For services without HTTP endpoints (business logic, utilities, infrastructure):

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
    
    // Configuration injection (type-safe)
    // @InjectCfgValue "auth.jwt-secret"
    JwtSecret string
    
    // @InjectCfgValue key="auth.token-expiry", default="24h"
    TokenExpiry time.Duration
    
    // @InjectCfgValue key="auth.max-attempts", default="5"
    MaxAttempts int
    
    // @InjectCfgValue key="auth.debug-mode", default="false"
    DebugMode bool
}

func (s *AuthService) Login(email, password string) (string, error) {
    // Check cache if available
    if s.Cache != nil {
        // Use cache
    }
    
    user, err := s.UserRepo.GetByEmail(email)
    if err != nil {
        return "", err
    }
    
    token := s.generateToken(user.ID, s.TokenExpiry)
    
    if s.DebugMode {
        println("Login successful:", email)
    }
    
    return token, nil
}
```

**Config (config.yaml):**
```yaml
configs:
  auth:
    jwt-secret: "your-secret-key"
    token-expiry: "48h"
    max-attempts: 3
    debug-mode: false
```

### Generate Code from Annotations

**Recommended: Automatic with Bootstrap**
```go
func main() {
    lokstra.Bootstrap() // Auto-generates when @Service/@RouterService changes detected
    // App code...
}
```

**Manual generation (before build/deploy):**
```bash
# From project root
lokstra autogen .

# Or from specific folder
lokstra autogen ./modules/user/application

# Force rebuild all (useful before deployment)
go run . --generate-only

# Generated file: zz_generated.lokstra.go
```

### Annotation Reference

| Annotation | Purpose | Example |
|------------|---------|---------|
| `@RouterService` | HTTP service with routes | `@RouterService name="user-service", prefix="/api"` |
| `@Service` | Pure service (no HTTP) | `@Service name="auth-service"` |
| `@Inject` | Dependency injection | `@Inject "user-repository"` or `@Inject service="cache", optional=true` |
| `@InjectCfgValue` | Config injection | `@InjectCfgValue "jwt-secret"` or `@InjectCfgValue key="timeout", default="30s"` |
| `@Route` | HTTP route mapping | `@Route "GET /users/{id}"` |

**@RouterService Parameters:**
- `name`: Service name (required)
- `prefix`: URL prefix (optional, default: "/")
  - **Supports variables**: `prefix="${api-prefix}"` resolves from config
- `middlewares`: Middleware list (optional)

**@Service Parameters:**
- `name`: Service name (required, positional or named)

**@Inject Parameters:**
- `service`: Service name (required, positional or named)
- `optional`: Boolean, default `false` - set to `true` for optional dependencies

**@InjectCfgValue Parameters:**
- `key`: Config key (required, positional or named)
- `default`: Default value (optional)
- Type auto-detected: `string`, `int`, `bool`, `float64`, `time.Duration`

**@Route Parameters:**
- HTTP method + path pattern
- Supports path parameters: `{id}`, `{userId}`, etc.
- **Supports variables in path**: `"GET ${api-version}/users/{id}"` resolves from config
- Optional `middlewares` parameter for per-route middleware: `middlewares=["mw1", "mw2"]`

**Variable Resolution:**
Variables in `${key}` or `${key:default}` format are resolved from config registry:

```yaml
# config.yaml
configs:
  - name: api-prefix
    value: /api/v1  # Can also use ${ENV_VAR:default}
  - name: api-version
    value: v2
```

```go
// Usage in annotations
// @RouterService name="user-service", prefix="${api-prefix}"
// Resolves to: prefix="/api/v1"

// @Route "GET ${api-version}/users/{id}"
// Resolves to: "GET v2/users/{id}"

// @Route "POST /${api-version:v1}/users"
// With default: if api-version not in config, uses "v1"
```

**Example with per-route middleware:**
```go
// @Route "GET /users/{id}"  // No middleware
func (s *UserService) GetByID(p *GetUserParams) (*User, error) { ... }

// @Route "POST /users", middlewares=["auth"]  // Requires auth
func (s *UserService) Create(p *CreateUserParams) (*User, error) { ... }

// @Route "DELETE /users/{id}", middlewares=["auth", "admin"]  // Requires auth + admin
func (s *UserService) Delete(p *DeleteUserParams) error { ... }
```

---

## Middleware Usage

### Built-in Middleware

```go
import (
    "github.com/primadi/lokstra/middleware/recovery"
    "github.com/primadi/lokstra/middleware/request_logger"
    "github.com/primadi/lokstra/middleware/slow_request_logger"
    "github.com/primadi/lokstra/middleware/cors"
    "github.com/primadi/lokstra/middleware/body_limit"
    "github.com/primadi/lokstra/middleware/gzipcompression"
)

r := lokstra.NewRouter("api")

// 1. Recovery - catch panics
r.Use(recovery.Middleware(&recovery.Config{
    EnableStackTrace: false, // Disable in production
    EnableLogging: true,
}))

// 2. Request Logger
r.Use(request_logger.Middleware(&request_logger.Config{
    Prefix: "API",
    EnableColors: true,
    SkipPaths: []string{"/health"},
}))

// 3. Slow Request Logger
r.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 500 * time.Millisecond,
    EnableColors: true,
}))

// 4. CORS
r.Use(cors.Middleware([]string{"*"}))

// 5. Body Limit
r.Use(body_limit.Middleware(&body_limit.Config{
    MaxSize: 10 * 1024 * 1024, // 10MB
    SkipOnPath: []string{"/upload/**"},
}))

// 6. Gzip Compression
r.Use(gzipcompression.Middleware(&gzipcompression.Config{
    MinSize: 1024,
    CompressionLevel: gzip.BestSpeed,
}))
```

### Custom Middleware

```go
func CustomAuthMiddleware(cfg *AuthConfig) request.HandlerFunc {
    return request.HandlerFunc(func(ctx *request.Context) error {
        // Pre-processing
        token := ctx.Req.Header.Get("Authorization")
        
        if !validateToken(token) {
            return ctx.Api.Unauthorized("Invalid token")
        }
        
        // Call next handler
        err := ctx.Next()
        
        // Post-processing (optional)
        
        return err
    })
}

// Usage
r.Use(CustomAuthMiddleware(&AuthConfig{
    Secret: "my-secret",
}))
```

### Middleware in YAML

```yaml
middleware-definitions:
  my-auth:
    type: custom-auth
    config:
      secret: "my-secret"
      skip_paths: ["/login", "/register"]

service-definitions:
  user-service:
    router:
      middlewares:
        - my-auth
```

---

## Dependency Injection

### Lazy Loading Pattern

**Service-level lazy loading** (services created on first access):

```go
import "github.com/primadi/lokstra/core/service"

// Define lazy reference (global variable)
var userService = service.LazyLoad[*UserService]("user-service")
var dbPool = service.LazyLoad[*DBPool]("database")

func handler() {
    // First call loads service (thread-safe via sync.Once)
    users := userService.MustGet().GetAll()
    
    // Subsequent calls return cached instance (zero overhead)
    user := userService.MustGet().GetByID("123")
}
```

### Dependency Injection Pattern

**Important**: Dependencies are **always resolved eagerly** when the service is created.

```go
type UserServiceImpl struct {
    // Direct dependency injection (resolved when service created)
    UserRepo user.UserRepository
}

func (s *UserServiceImpl) GetByID(id string) (*user.User, error) {
    // Direct access - dependency already loaded during service creation
    return s.UserRepo.GetByID(id)
}

func UserServiceFactory(deps map[string]any, config map[string]any) any {
    return &UserServiceImpl{
        // Direct type assertion - dependency already resolved
        UserRepo: deps["user-repository"].(user.UserRepository),
    }
}
```

### Service Registration

```go
import "github.com/primadi/lokstra/lokstra_registry"

func registerServiceTypes() {
    // Register with factory functions
    lokstra_registry.RegisterServiceType(
        "user-service-factory",       // Type name (matches config.yaml)
        UserServiceFactory,            // Local factory (monolith)
        UserServiceRemoteFactory,      // Remote factory (microservices)
    )
}
```

---

## Project Structure Templates

### Template 1: Router Only (Simple)

```
myapp/
├── main.go
├── handlers.go
├── middleware.go
└── go.mod
```

**main.go:**
```go
package main

import "github.com/primadi/lokstra"

func main() {
    r := lokstra.NewRouter("api")
    r.GET("/users", listUsers)
    r.POST("/users", createUser)
    
    app := lokstra.NewApp("simple", ":8080", r)
    app.Run(30 * time.Second)
}
```

### Template 2: Medium System (DDD)

```
myapp/
├── main.go
├── register.go
├── config.yaml
├── domain/
│   ├── user/
│   │   ├── models.go
│   │   ├── repository.go
│   │   └── service.go
│   └── order/
│       ├── models.go
│       ├── repository.go
│       └── service.go
├── repository/
│   ├── user_repository.go
│   └── order_repository.go
└── service/
    ├── user_service.go
    └── order_service.go
```

### Template 3: Enterprise Modular (Bounded Contexts)

```
myapp/
├── main.go
├── register.go
├── config.yaml
├── shared/
│   ├── errors/
│   └── types/
└── modules/
    ├── user/
    │   ├── domain/
    │   │   ├── models.go
    │   │   ├── repository.go
    │   │   └── service.go
    │   ├── application/
    │   │   └── user_service.go
    │   └── infrastructure/
    │       └── user_repository.go
    └── order/
        ├── domain/
        ├── application/
        └── infrastructure/
```

### Template 4: Enterprise with Annotations

```
myapp/
├── main.go
├── register.go
├── config.yaml
└── modules/
    └── user/
        ├── domain/
        │   ├── models.go
        │   ├── repository.go
        │   └── service.go
        ├── application/
        │   ├── user_service.go           # Contains @RouterService, @Route
        │   └── zz_generated.lokstra.go  # Auto-generated
        └── infrastructure/
            └── user_repository.go
```

**Generate code:**
```bash
lokstra autogen ./modules/user/application
```

---

## Common Patterns & Idioms

### Pattern: Request/Response DTOs

```go
// Request DTO with validation
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=18,max=120"`
}

// Response DTO
type UserResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

// Handler
r.POST("/users", func(ctx *request.Context, req *CreateUserRequest) error {
    user := createUser(req)
    resp := &UserResponse{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        CreatedAt: user.CreatedAt,
    }
    return ctx.Api.Created(resp)
})
```

### Pattern: Error Handling

```go
import "github.com/primadi/lokstra/core/request"

func handler(ctx *request.Context) error {
    // Not found
    return ctx.Api.NotFound("User not found")
    
    // Bad request
    return ctx.Api.BadRequest("Invalid parameters")
    
    // Unauthorized
    return ctx.Api.Unauthorized("Invalid token")
    
    // Forbidden
    return ctx.Api.Forbidden("Access denied")
    
    // Internal error
    return ctx.Api.InternalServerError("Database connection failed")
    
    // Custom error
    return ctx.Api.ErrorWithCode(422, "Validation failed", validationErrors)
}
```

### Pattern: Context Response Helpers

```go
func handler(ctx *request.Context) error {
    // Success responses
    ctx.Api.Ok(data)                    // 200 OK
    ctx.Api.Created(data)               // 201 Created
    ctx.Api.NoContent()                 // 204 No Content
    
    // Error responses
    ctx.Api.BadRequest(message)         // 400
    ctx.Api.Unauthorized(message)       // 401
    ctx.Api.Forbidden(message)          // 403
    ctx.Api.NotFound(message)           // 404
    ctx.Api.InternalServerError(msg)    // 500
    
    // Custom response
    ctx.Api.ErrorWithCode(code, msg, data)
}
```

### Pattern: Database Integration

```go
import "github.com/primadi/lokstra/services/dbpool_pg"

// Register database pool
func registerServiceTypes() {
    lokstra_registry.RegisterServiceType(
        "database",
        dbpool_pg.Factory,
        nil,
    )
}
```

**config.yaml:**
```yaml
service-definitions:
  database:
    type: database
    config:
      dsn: "postgres://user:pass@localhost/mydb?sslmode=disable"
      max_open_conns: 25
      max_idle_conns: 5
```

**Usage in repository:**
```go
type UserRepository struct {
    DB *pgxpool.Pool
}

func (r *UserRepository) GetByID(id string) (*User, error) {
    var user User
    err := r.DB.QueryRow(context.Background(),
        "SELECT id, name, email FROM users WHERE id = $1", id,
    ).Scan(&user.ID, &user.Name, &user.Email)
    
    return &user, err
}
```

### Pattern: Multiple Deployments

**config.yaml:**
```yaml
deployments:
  # Development: All services in one server
  development:
    servers:
      monolith:
        addr: ":8080"
        published-services: [user-service, order-service, payment-service]
  
  # Production: Separate microservices
  production:
    servers:
      user-api:
        addr: ":8001"
        base-url: "https://user-api.example.com"
        published-services: [user-service]
      
      order-api:
        addr: ":8002"
        base-url: "https://order-api.example.com"
        published-services: [order-service]
```

**Run:**
```bash
# Development
LOKSTRA_DEPLOYMENT=development go run .

# Production (multiple processes)
LOKSTRA_DEPLOYMENT=production LOKSTRA_SERVER=user-api go run .
LOKSTRA_DEPLOYMENT=production LOKSTRA_SERVER=order-api go run .
```

---

## Troubleshooting

### Common Issues

#### 1. Service Not Found

**Error:**
```
panic: service 'user-service' not found in registry
```

**Solution:**
- **For business services**: Use `@RouterService` annotation + `lokstra autogen .`
- **For infrastructure services**: Check service registered: `lokstra_registry.RegisterServiceType("user-service-factory", ...)`
- Check config.yaml: Service name must match factory type
- Check annotation-generated file: `zz_generated.lokstra.go` exists

#### 2. Import Cycle

**Error:**
```
import cycle not allowed
```

**Solution:**
- Use domain interfaces in separate package
- Repository/Service should depend on domain, not vice versa
- Structure: `domain/` → `repository/` → `service/`

#### 3. Handler Signature Not Recognized

**Error:**
```
unsupported handler signature
```

**Solution:**
- Check handler returns error or supported type
- Use `*request.Context` for context access
- Use struct pointers for request parameters: `*CreateUserParams`

#### 4. Validation Not Working

**Error:**
```
validation tags ignored
```

**Solution:**
- Ensure struct has `validate` tags: `validate:"required,email"`
- Use pointer to struct: `func handler(params *CreateUserParams)`
- Import validator: `github.com/primadi/lokstra/common/validator`

#### 5. Annotation Code Not Generated

**Error:**
```
zz_generated.lokstra.go not created
```

**Solution:**
```bash
# Run code generator
lokstra autogen ./path/to/service

# Ensure annotations are correct
# @RouterService name="service-name", prefix="/api"
# @Route "GET /users/{id}"
```

#### 6. Middleware Not Applied

**Solution:**
- Check middleware order: `recovery` should be first
- Check middleware registered: Built-in middleware auto-registered
- Verify YAML config if using config-driven middleware

---

## Quick Reference

### CLI Commands

```bash
# Install CLI
go install github.com/primadi/lokstra/cmd/lokstra@latest

# Create new project
lokstra new myapp
lokstra new myapp -template 02_app_framework/01_medium_system

# Generate code from annotations (manual)
lokstra autogen .                      # Current directory
lokstra autogen ./modules/user/app     # Specific folder

# Force regenerate all (before deployment)
go run . --generate-only               # Uses lokstra.Bootstrap()

# List templates
lokstra new --help

# Note: Use lokstra.Bootstrap() in main() for automatic generation
```

### Import Paths

```go
// Core
import "github.com/primadi/lokstra"
import "github.com/primadi/lokstra/core/request"
import "github.com/primadi/lokstra/core/service"
import "github.com/primadi/lokstra/core/deploy"

// Registry
import "github.com/primadi/lokstra/lokstra_registry"

// Middleware
import "github.com/primadi/lokstra/middleware/recovery"
import "github.com/primadi/lokstra/middleware/request_logger"
import "github.com/primadi/lokstra/middleware/cors"

// Services
import "github.com/primadi/lokstra/services/dbpool_pg"
import "github.com/primadi/lokstra/services/redis"
```

### Environment Variables

```bash
# Deployment selection
LOKSTRA_DEPLOYMENT=production

# Server selection (multi-server deployment)
LOKSTRA_SERVER=api-server

# Log level
LOKSTRA_LOG_LEVEL=debug  # silent, error, warn, info, debug

# Config file path
LOKSTRA_CONFIG=./config.yaml
```

---

## Resources

- **Documentation:** https://primadi.github.io/lokstra/
- **GitHub:** https://github.com/primadi/lokstra
- **Examples:** https://primadi.github.io/lokstra/00-introduction/examples/
- **Templates:** https://github.com/primadi/lokstra/tree/dev2/project_templates
- **Schema:** https://primadi.github.io/lokstra/schema/lokstra.schema.json

---

## AI Agent Best Practices

### When Helping Programmers:

1. **Ask about project scale:**
   - Small/learning → Suggest router-only mode
   - Medium/production → Suggest framework mode with annotations
   - Enterprise → Suggest modular architecture with annotations

2. **Suggest appropriate template:**
   - `01_router/01_router_only` for learning
   - `02_app_framework/01_medium_system` for production apps (2-10 entities)
   - `02_app_framework/03_enterprise_router_service` for enterprise with annotations

3. **Always provide complete code:**
   - Include imports
   - Include error handling
   - Include validation tags
   - Include config.yaml if using framework mode
   - **Use `@RouterService` annotations for business services**

4. **Follow project structure:**
   - `domain/` for interfaces and models
   - `infrastructure/` for data access (repositories)
   - `application/` for business logic (services with `@RouterService`)
   - `main.go` for bootstrap (import annotation packages)

5. **Use type-safe patterns:**
   - Direct type assertions in factory functions
   - `service.LazyLoad[T]` for lazy service loading
   - **Prefer annotations over manual registration for business services**

6. **Recommend annotation workflow:**
   - Define service with `@RouterService` annotation
   - Add routes with `@Route` annotation
   - Run `lokstra autogen .` to generate code
   - Manual registration only for infrastructure/custom factories

---

**End of AI Agent Guide**

For more details, refer to the [complete documentation](https://primadi.github.io/lokstra/).
