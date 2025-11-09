---
layout: docs
title: Router-Only Examples
---

# Track 1: Router-Only Examples

> **Use Lokstra as a flexible HTTP router (like Echo, Gin, Chi)**  
> **Time**: 2-3 hours • **Level**: Beginner

---

## 📚 What You'll Learn

This track focuses on using Lokstra as a **pure HTTP router** without dependency injection or framework features:

- ✅ Basic routing and handler registration
- ✅ 29 different handler signature variations
- ✅ Middleware patterns (global, per-route, groups)
- ✅ Quick API prototyping

**No DI, no config files, no services** - just routing!

---

## 🎯 Learning Path

### [01 - Hello World](./01-hello-world/) ⏱️ 15 min

Your first Lokstra API in 10 lines of code.

```go
r := lokstra.NewRouter("api")
r.GET("/", func() string {
    return "Hello, Lokstra!"
})
app := lokstra.NewApp("hello", ":3000", r)
app.Run(30 * time.Second)
```

**What you'll learn:**
- Router creation
- Simple handlers
- Starting the app

---

### [02 - Handler Forms](./02-handler-forms/) ⏱️ 30 min

Explore Lokstra's flexible handler signatures.

```go
// Simple return
r.GET("/ping", func() string { return "pong" })

// With error
r.GET("/users", func() ([]User, error) {
    return db.GetUsers()
})

// Request binding
r.POST("/users", func(req *CreateUserRequest) (*User, error) {
    return db.Create(req)
})

// Full control
r.GET("/custom", func(ctx *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.Json(data)
    return resp, nil
})
```

**What you'll learn:**
- 29 handler signature variations
- When to use each form
- Request parameter binding
- Response patterns

---

### [03 - Middleware](./03-middleware/) ⏱️ 45 min

Master middleware for cross-cutting concerns.

```go
// Global middleware
r.Use(RecoveryMiddleware, LoggerMiddleware)

// Per-route middleware
r.GET("/public", publicHandler)
r.GET("/protected", protectedHandler, AuthMiddleware)

// Group middleware
admin := r.AddGroup("/admin")
admin.Use(AuthMiddleware, AdminMiddleware)
admin.GET("/users", listUsers)
```

**What you'll learn:**
- Global middleware (all routes)
- Per-route middleware (specific endpoints)
- Group middleware (API versioning)
- Custom middleware creation
- Built-in middleware (CORS, Recovery, Logger)

---

## 🚀 Running Examples

```bash
# Navigate to any example
cd 01-hello-world  # or 02, 03

# Run it
go run main.go

# Test it
curl http://localhost:3000/
```

---

## 📊 Skills Progression

```
Example 01:  Router Basics
    → Create routers, register simple handlers

Example 02:  Handler Flexibility
    → 29 handler forms, request/response patterns

Example 03:  Middleware
    → Global, per-route, groups, custom middleware
```

---

## 🎯 After This Track

### Want to Continue with Router?
- **[Router Guide](../../../01-router-guide/)** - Deep dive into routing features
- **[API Reference - Router](../../../03-api-reference/01-core-packages/router)** - Complete router API

### Ready for More?
- **[Framework Track](../full-framework/)** - Learn DI, services, auto-router
- **[Framework Guide](../../../02-framework-guide/)** - Enterprise patterns

---

## 💡 When to Use Router-Only

**Good for:**
- ✅ Quick prototypes and MVPs
- ✅ Simple REST APIs
- ✅ Learning HTTP routing
- ✅ Microservices without DI needs
- ✅ Teams familiar with Echo/Gin/Chi

**Upgrade to Framework when:**
- Need dependency injection
- Want auto-generated routers
- Building microservices
- Need configuration-driven deployment

---

**Ready to start?** → [01 - Hello World](./01-hello-world/)
