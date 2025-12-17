---
layout: docs
title: Lokstra Router vs Gin / Echo / Chi
description: How Lokstra compares when used as a pure HTTP router.
---

## Same Problem Space

When you use Lokstra in **Track 1 (Router)** mode, it competes with:

- Gin  
- Echo  
- Chi  

The goal is the same: define routes, run middleware, handle requests.

## What Feels Familiar

- Register routes with methods: `GET`, `POST`, `PUT`, `DELETE`, etc.  
- Use groups/prefixes for APIs (e.g. `/api/v1/users`).  
- Attach global or per‑group middleware.

Example (from `01_router_only`):

```go
r := lokstra.NewRouter("demo_router")

r.Use(recovery.Middleware(recovery.DefaultConfig()))

users := r.AddGroup("/users")
users.GET("", handleGetUsers)
users.POST("", handleCreateUser)
```

## What Lokstra Adds

- **Rich handler signatures**  
  - Handlers can return values, errors, or use `*request.Context`.  
  - Lokstra handles JSON encoding and error mapping for you.

- **Type‑safe binding**  
  - Struct tags (`path:"id"`, `json:"name"`, `query:"status"`, `validate:"required"`)  
  - Automatic binding from path/query/body + validation.

- **Smooth upgrade path**  
  - Start as a pure router.  
  - Later, introduce services, YAML config, and annotations (Track 2) without replacing the router.

If you only need an HTTP router today, you can adopt Lokstra like Gin/Echo,
and keep the option open to grow into a full framework later.


