# Lokstra Roadmap

> **The future of Lokstra - features, improvements, and vision**

Last Updated: October 2025

---

## 🎯 Current Status

**Version**: 2.x (dev2 branch)  
**Status**: Active Development  
**Focus**: Core stabilization + Essential features

### What's Working Now ✅
- ✅ 29 handler forms
- ✅ Service as router (convention-based routing)
- ✅ Multi-deployment (monolith ↔ microservices)
- ✅ Built-in lazy dependency injection
- ✅ YAML configuration with environment variables
- ✅ Middleware system (direct + by-name)
- ✅ Request/response helpers
- ✅ Route groups and scopes
- ✅ Path parameter extraction
- ✅ Query parameter binding
- ✅ JSON request/response
- ✅ FastHTTP engine support

---

## 🚀 Next Release (v2.1)

**Target**: Q4 2025  
**Theme**: Developer Experience + Production Essentials

### 1. 🎨 HTMX Support

**Goal**: Make building web applications as easy as REST APIs

#### Features
- [ ] Template rendering integration
  - [ ] `html/template` support
  - [ ] `templ` support (type-safe templates)
  - [ ] Auto content-type detection
- [ ] HTMX helpers
  - [ ] Response headers (HX-Trigger, HX-Redirect, etc.)
  - [ ] Request detection (HX-Request, HX-Target)
  - [ ] Partial rendering utilities
- [ ] Form handling
  - [ ] Form binding to structs
  - [ ] Validation error rendering
  - [ ] CSRF protection
- [ ] Examples
  - [ ] Todo app with HTMX
  - [ ] Dashboard with real-time updates
  - [ ] Form validation patterns

#### API Design
```go
// Handler returns templ component
r.GET("/users", func() templ.Component {
    users := userService.GetAll()
    return views.UserList(users)
})

// Partial update
r.POST("/users", func(req *CreateUserReq) templ.Component {
    user := userService.Create(req)
    return views.UserRow(user)  // Returns single row
})

// HTMX helpers
r.POST("/toggle", func(ctx *request.Context) (*response.Response, error) {
    return response.Success(data).
        WithHeader("HX-Trigger", "itemToggled").
        WithHeader("HX-Redirect", "/dashboard"), nil
})
```

---

### 2. 🛠️ CLI Tools

**Goal**: Speed up development workflow

#### Features
- [ ] Project scaffolding
  ```bash
  lokstra new my-api --template=rest-api
  lokstra new my-web --template=htmx-app
  lokstra new my-mono --template=monolith
  ```
  
- [ ] Code generation
  ```bash
  lokstra generate service user
  lokstra generate router api
  lokstra generate middleware auth
  lokstra generate handler users/create
  ```
  
- [ ] Development server
  ```bash
  lokstra dev --port 3000 --hot-reload
  ```
  
- [ ] Migration management
  ```bash
  lokstra migrate create add_users_table
  lokstra migrate up
  lokstra migrate down
  lokstra migrate status
  ```
  
- [ ] Testing utilities
  ```bash
  lokstra test --coverage
  lokstra test --watch
  ```

#### Templates
- **rest-api**: Classic REST API with database
- **htmx-app**: Web app with HTMX + templ
- **monolith**: Single deployment with multiple services
- **microservices**: Multi-service deployment ready
- **minimal**: Bare bones setup

---

### 3. 📦 Standard Middleware Library

**Goal**: Production-ready middleware out of the box

#### Authentication & Authorization
```go
// JWT Authentication
r.Use(middleware.JWT(middleware.JWTConfig{
    Secret:     "your-secret",
    SigningAlg: "HS256",
    Header:     "Authorization",
    Prefix:     "Bearer ",
}))

// OAuth2 Integration
r.Use(middleware.OAuth2(middleware.OAuth2Config{
    Provider:     "google",
    ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
    ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
    RedirectURL:  "http://localhost:8080/callback",
}))

// Basic Auth (simple use cases)
r.Use(middleware.BasicAuth(map[string]string{
    "admin": "hashed-password",
}))

// API Key
r.Use(middleware.APIKey(middleware.APIKeyConfig{
    Header: "X-API-Key",
    Validator: func(key string) bool {
        return db.ValidateAPIKey(key)
    },
}))
```

#### Metrics & Monitoring
```go
// Prometheus metrics
r.Use(middleware.Prometheus(middleware.PrometheusConfig{
    Subsystem: "api",
    Path:      "/metrics",
}))

// OpenTelemetry tracing
r.Use(middleware.OpenTelemetry(middleware.OTelConfig{
    ServiceName: "my-api",
    Endpoint:    "otel-collector:4317",
}))

// Custom metrics
r.Use(middleware.Metrics(func(ctx *request.Context, duration time.Duration) {
    metrics.RecordRequest(
        ctx.R.Method,
        ctx.R.URL.Path,
        ctx.ResponseStatus,
        duration,
    )
}))
```

#### Rate Limiting
```go
// In-memory rate limiter
r.Use(middleware.RateLimit(middleware.RateLimitConfig{
    Requests: 100,
    Window:   time.Minute,
    KeyFunc: func(ctx *request.Context) string {
        return ctx.R.RemoteAddr  // By IP
    },
}))

// Redis-backed rate limiter
r.Use(middleware.RateLimitRedis(middleware.RateLimitRedisConfig{
    Redis:    redisClient,
    Requests: 1000,
    Window:   time.Hour,
    KeyFunc: func(ctx *request.Context) string {
        return ctx.Get("user_id").(string)  // By user
    },
}))
```

#### Security
```go
// CSRF Protection
r.Use(middleware.CSRF(middleware.CSRFConfig{
    TokenLength: 32,
    CookieName:  "_csrf",
    HeaderName:  "X-CSRF-Token",
}))

// Security Headers
r.Use(middleware.SecureHeaders(middleware.SecureHeadersConfig{
    ContentSecurityPolicy: "default-src 'self'",
    XFrameOptions:         "DENY",
    XContentTypeOptions:   "nosniff",
}))

// Request ID
r.Use(middleware.RequestID(middleware.RequestIDConfig{
    Header:    "X-Request-ID",
    Generator: uuid.New,
}))
```

---

### 4. 📦 Standard Service Library

**Goal**: Common service patterns ready to use

#### Health Checks
```go
health := lokstra_registry.GetService[*service.Health]("health")

// Add custom checks
health.AddCheck("database", func() error {
    return db.Ping()
})

health.AddCheck("redis", func() error {
    return redis.Ping()
})

health.AddCheck("external-api", func() error {
    resp, err := http.Get("https://api.example.com/health")
    if err != nil || resp.StatusCode != 200 {
        return fmt.Errorf("external API down")
    }
    return nil
})

// Auto-register /health endpoint
r.GET("/health", health.Handler())
```

#### Metrics Service
```go
metrics := lokstra_registry.GetService[*service.Metrics]("metrics")

// Record metrics
metrics.Counter("requests_total", labels).Inc()
metrics.Histogram("request_duration", labels).Observe(duration)
metrics.Gauge("active_connections").Set(count)

// Auto-register /metrics endpoint
r.GET("/metrics", metrics.Handler())
```

#### Tracing Service
```go
tracer := lokstra_registry.GetService[*service.Tracing]("tracing")

// Manual spans
span := tracer.StartSpan(ctx, "user.create")
defer span.End()

span.SetAttribute("user.id", user.ID)
span.SetAttribute("user.email", user.Email)

// Automatic instrumentation
r.Use(tracer.Middleware())  // Auto-traces all requests
```

---

## 🔮 Future Releases

### v2.2 - Advanced Features (Q1 2026)

#### Plugin System
```go
// Load plugins
lokstra.LoadPlugin("auth-plugin", authPlugin)
lokstra.LoadPlugin("custom-logger", loggerPlugin)

// Plugins can:
// - Register services
// - Add middleware
// - Extend routers
// - Hook into lifecycle
```

#### Admin Dashboard
- Built-in API explorer
- Request/response inspector
- Performance metrics
- Log viewer
- Configuration editor
- Service registry viewer

#### API Documentation
```go
// Auto-generate OpenAPI/Swagger
r.GET("/users", getUsers).
    Doc("Get all users").
    Response(200, []User{}).
    Response(500, Error{})

// Generate docs
swagger := lokstra.GenerateOpenAPI(r)
r.GET("/swagger", swagger.Handler())
```

---

### v2.3 - Real-time & GraphQL (Q2 2026)

#### WebSocket Support
```go
r.WebSocket("/ws", func(conn *websocket.Conn) {
    for {
        msg, err := conn.ReadMessage()
        if err != nil {
            break
        }
        conn.WriteMessage(process(msg))
    }
})
```

#### Server-Sent Events (SSE)
```go
r.GET("/events", func(ctx *request.Context) error {
    stream := sse.NewStream(ctx.W)
    
    for event := range events {
        stream.Send(event)
    }
    
    return nil
})
```

#### GraphQL Support
```go
schema := graphql.NewSchema(...)

r.POST("/graphql", func(req *graphql.Request) (*graphql.Response, error) {
    return schema.Execute(req)
})
```

---

### v3.0 - Breaking Changes & Modernization (Q4 2026)

**Focus**: Lessons learned, API improvements, Go 1.24+ features

Potential changes:
- Refined handler signatures
- Improved error handling patterns
- Enhanced generics usage
- Streamlined configuration
- Better performance optimizations

---

## 🎯 Long-term Vision

### Core Principles (Unchanging)
1. **Developer Experience First** - Easy to learn, productive to use
2. **Flexible by Design** - Multiple ways to solve problems
3. **Convention over Configuration** - Smart defaults, configure when needed
4. **Production Ready** - Battle-tested patterns
5. **Go Idiomatic** - Feels natural to Go developers

### Goals
- 🎯 **Top 5 Go web framework** by 2027
- 🎯 **10,000+ GitHub stars** by 2027
- 🎯 **100+ production deployments** by 2026
- 🎯 **Active community** with regular contributions
- 🎯 **Comprehensive ecosystem** of plugins and tools

---

## 🤝 How to Contribute

### Immediate Needs
- 📝 Documentation improvements
- 🧪 More example applications
- 🐛 Bug reports and fixes
- 💡 Feature suggestions
- 🎨 Logo and branding

### Getting Involved
1. Check [GitHub Issues](https://github.com/primadi/lokstra/issues)
2. Join discussions
3. Submit PRs
4. Write blog posts
5. Share your projects

---

## 📊 Progress Tracking

### Milestones

| Release | Target | Status | Features |
|---------|--------|--------|----------|
| v2.0 | ✅ Done | Released | Core framework, 29 handlers, service-as-router |
| v2.1 | Q4 2025 | 🟡 In Progress | HTMX, CLI tools, Standard middleware |
| v2.2 | Q1 2026 | 📅 Planned | Plugins, Admin dashboard, API docs |
| v2.3 | Q2 2026 | 📅 Planned | WebSocket, SSE, GraphQL |
| v3.0 | Q4 2026 | 💭 Concept | API refinement, modernization |

### Feature Status

#### Next Release (v2.1)
- 🟡 HTMX Support (30% complete)
  - ✅ Research and design
  - 🔄 Template integration
  - ⏳ Helper functions
  - ⏳ Examples
  
- 🟡 CLI Tools (20% complete)
  - ✅ Project structure
  - 🔄 Scaffolding templates
  - ⏳ Code generation
  - ⏳ Hot reload
  
- 🟡 Standard Middleware (40% complete)
  - ✅ Logging (done)
  - ✅ CORS (done)
  - 🔄 JWT auth
  - ⏳ OAuth2
  - ⏳ Metrics
  - ⏳ Rate limiting
  - ⏳ Security headers
  
- 🟡 Standard Services (10% complete)
  - ⏳ Health checks
  - ⏳ Metrics
  - ⏳ Tracing

Legend:
- ✅ Done
- 🔄 In Progress
- ⏳ Not Started
- 🟡 Partial
- 📅 Planned
- 💭 Concept

---

## 📝 Release Process

### Version Strategy
- **Major (v3.0)**: Breaking changes
- **Minor (v2.1)**: New features, backward compatible
- **Patch (v2.0.1)**: Bug fixes only

### Release Checklist
- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Migration guide (if breaking changes)
- [ ] Examples updated
- [ ] Performance benchmarks
- [ ] Security review
- [ ] Community announcement

---

## 💬 Feedback & Discussion

We want to hear from you!

- 🐛 **Found a bug?** [Open an issue](https://github.com/primadi/lokstra/issues/new?template=bug_report)
- 💡 **Have an idea?** [Open a feature request](https://github.com/primadi/lokstra/issues/new?template=feature_request)
- 💬 **Want to discuss?** [Start a discussion](https://github.com/primadi/lokstra/discussions)
- 📧 **Need help?** [Join our community](https://github.com/primadi/lokstra/discussions/categories/q-a)

---

**Last Updated**: October 2025  
**Maintained by**: Lokstra Core Team

👉 Back to [Documentation Home](index)
