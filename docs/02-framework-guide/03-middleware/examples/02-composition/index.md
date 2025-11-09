# Middleware Composition

> # Middleware Composition Example

Learn how to compose multiple middleware together, create conditional middleware, and build reusable middleware chains.

## Running the Example

```bash
go run main.go
```

Server starts on `http://localhost:3001`

## Key Patterns

### Chain Pattern

Combine multiple middleware into one:

```go
func Chain(middlewares ...func(*request.Context) error) func(*request.Context) error {
    return func(c *request.Context) error {
        for _, middleware := range middlewares {
            if err := middleware(c); err != nil {
                return err
            }
        }
        return c.Next()
    }
}

// Usage
router.GET("/admin", handler, Chain(AuthMiddleware, AdminMiddleware))
```

### Conditional Middleware

Execute middleware only when condition is met:

```go
func When(condition func(*request.Context) bool, middleware func(*request.Context) error) func(*request.Context) error {
    return func(c *request.Context) error {
        if condition(c) {
            return middleware(c)
        }
        return c.Next()
    }
}

// Usage
router.GET("/api/data", handler, When(func(c *request.Context) bool {
    return strings.HasPrefix(c.R.URL.Path, "/api/")
}, CORSMiddleware))
```

### Unless Pattern

Skip middleware when condition is met:

```go
func Unless(condition func(*request.Context) bool, middleware func(*request.Context) error) func(*request.Context) error {
    return func(c *request.Context) error {
        if !condition(c) {
            return middleware(c)
        }
        return c.Next()
    }
}

// Usage - skip logging for health checks
router.GET("/health", handler, Unless(func(c *request.Context) bool {
    return c.R.URL.Path == "/health"
}, LoggerMiddleware))
```

## Test Endpoints

- `/public` - No extra middleware
- `/user` - Requires authentication
- `/admin` - Requires auth + admin role
- `/api/data` - Conditional CORS middleware
- `/health` - Skips logging
- `/api/protected` - Composed middleware stack - This example is being prepared.

## Topics Covered

Chaining, conditional middleware

## Placeholder

This example is being prepared. Check back soon for:
- Working code examples
- Comprehensive documentation
- Test files
- Best practices guide

---

**Status**: 📝 In Development
