# Lokstra Framework – Positioning Statement

## Purpose

Lokstra is a backend framework for Go developers who want a **structured, scalable, and developer-friendly way to build web services and APIs**. It is designed to reduce boilerplate, enforce clear conventions, and support both monolith and microservice architectures.

Lokstra is not a platform, not a service mesh, and not a low-level network proxy. It is a lightweight and practical framework that enhances the Go ecosystem without hiding its core principles.

---

## What Lokstra **Is**

| Domain                | Lokstra Provides                                               |
| --------------------- | -------------------------------------------------------------- |
| Backend structure     | Modular project layout, app/server separation                  |
| Routing               | Fast and flexible routing engine with middleware at all levels |
| Request Binding       | Smart struct tag binding (`path:"id"`, `query:"page"`, `body:"name"`) |
| Response System       | Structured JSON responses with method chaining                 |
| Service registration  | Type-safe dependency injection with factory pattern           |
| Configuration         | YAML schema validation with directory loading & ENV overrides |
| HTMX Integration      | First-class HTMX support for modern web applications          |
| Static Files          | Efficient static serving with SPA support                     |
| Developer Experience  | Auto-bind handlers, context helpers, minimal boilerplate      |
| Observability (basic) | Prometheus metrics, structured logging, health checks         |

---

## What Lokstra **Is Not**

| Not a...              | Explanation                                                    |
| --------------------- | -------------------------------------------------------------- |
| Service mesh          | Does not handle traffic routing between services (e.g., Istio) |
| APM or Tracing system | Does not inject distributed tracing (yet)                      |
| Operator framework    | No K8s controller/operator support at runtime (yet)            |
| RPC framework         | Not a gRPC or internal RPC tool (planned as `lokstra-call`)    |
| Web framework with UI | Lokstra is backend-only; frontend handled separately           |
| Deployment platform  | Deployment modes achieved via Docker strategies, not framework |

---

## Key Differentiators

What sets Lokstra apart from other Go web frameworks:

### Smart Request Binding
```go
type UserRequest struct {
    ID    string `path:"id"`           // From URL path
    Token string `header:"Authorization"` // From HTTP headers  
    Name  string `body:"name"`         // From request body
    Page  int    `query:"page"`        // From query parameters
}

// Auto-bind smart pattern
func getUserHandler(ctx *lokstra.Context, req *UserRequest) error {
    // Request automatically bound - use data directly
    return ctx.Ok(req)
}
```

### HTMX-First Development
- Built-in HTMX page serving with script injection
- HTMX-aware response helpers and context methods
- Seamless static file + dynamic content integration

### Configuration as Code
- YAML schema validation with IDE support
- Directory-based config loading and merging
- Environment variable overrides with type safety

### Type-Safe Service Container
```go
// Type-safe service retrieval
dbPool, err := lokstra.GetService[serviceapi.DbPool](regCtx, "db.main")
logger, err := lokstra.GetOrCreateService[serviceapi.Logger](regCtx, "logger", "info")
```

### Structured Response System
```go
// Method-chained responses with consistent structure
return ctx.Ok(data).
    WithMessage("Custom success message").
    WithHeader("X-Custom-Header", "value").
    WithResponseCode("CUSTOM_CODE")
```

---

## Use Case Fit

Lokstra is suitable for:

* Go developers building APIs, internal services, or microservices
* Teams that want convention and structure without excessive abstraction
* Modern web applications with HTMX for dynamic interfaces
* Applications requiring smart request binding and structured responses
* Projects using YAML-based configuration with schema validation
* Lightweight deployments (Docker, VPS, Kubernetes)
* Applications that grow over time from single app → multiple apps

Lokstra is **not ideal** for:

* Complex streaming workloads
* Graph-heavy systems needing auto schema
* Applications requiring real-time bidirectional communication
* Projects needing advanced distributed tracing out-of-the-box

---

## Summary

> Lokstra is the Rails-inspired framework for Go backend services. It combines Go's simplicity with structured conventions, smart request binding, HTMX integration, and type-safe dependency injection — without sacrificing control or performance.

Tagline: **Simple. Smart. Structured.**
