---
layout: docs
title: 01 – Router Guide
description: Using Lokstra as a fast HTTP router (like Gin/Echo).
---

## Lokstra as a Router

If you only need **routing + middleware**, you can use Lokstra like Gin/Echo.

### Minimal Example

```go
r := lokstra.NewRouter("api")

r.GET("/ping", func() string {
    return "pong"
})

app := lokstra.NewApp("api", ":3000", r)
app.Run(30 * time.Second)
```

### Route Groups & Middleware

From `project_templates/01_router/01_router_only`:

```go
r := lokstra.NewRouter("demo_router")

r.Use(recovery.Middleware(recovery.DefaultConfig()))
r.Use(slow_request_logger.Middleware(&slow_request_logger.Config{
    Threshold: 100 * time.Millisecond,
}))

users := r.AddGroup("/users")
users.GET("", handleGetUsers)
users.POST("", handleCreateUser)
```

You get:

- Grouped routes (`/users`, `/roles`, etc.).
- Global and per-group middleware.
- Built-in middleware (recovery, logging, slow-request logger, gzip, cors).

### When to Use Router-Only Mode

Use Lokstra as a router when:

- You would normally pick **Gin / Echo / Chi**.  
- Your service is mostly HTTP handlers with simple dependencies.  
- You don't need DI, YAML config, or multi-topology deployment yet.

When your domain and deployment grow, you can move to Track 2 (Framework)
without changing how you think about handlers and middleware.


