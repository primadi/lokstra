# Lifecycle Hooks & Middleware

> **Master middleware execution order and lifecycle**

This example demonstrates how middleware creates before/after hooks around handlers.

## Middleware Execution Flow

```
Request
  ↓
[Global Middleware]
  ↓
[Route Middleware 1] → Before
  ↓
[Route Middleware 2] → Before
  ↓
[Handler] → Main logic
  ↓
[Route Middleware 2] → After
  ↓
[Route Middleware 1] → After
  ↓
Response
```

## Key Concept: ctx.Next()

```go
func middleware(ctx *request.Context) error {
    // Code BEFORE ctx.Next() runs BEFORE handler
    fmt.Println("Before handler")
    
    // Call next middleware/handler
    err := ctx.Next()
    
    // Code AFTER ctx.Next() runs AFTER handler
    fmt.Println("After handler")
    
    return err
}
```

## Patterns

### Pattern 1: Before Hook
```go
func beforeMiddleware(ctx *request.Context) error {
    fmt.Println("Setup work")
    ctx.Set("start_time", time.Now())
    return ctx.Next()
}
```

### Pattern 2: After Hook
```go
func afterMiddleware(ctx *request.Context) error {
    err := ctx.Next() // Execute handler first
    
    // Cleanup or post-processing
    duration := time.Since(ctx.Get("start_time"))
    fmt.Printf("Duration: %v\n", duration)
    
    return err
}
```

### Pattern 3: Around Hook (Before + After)
```go
func loggingMiddleware(ctx *request.Context) error {
    start := time.Now()
    fmt.Println("→ Request started")
    
    err := ctx.Next()
    
    fmt.Printf("← Request finished in %v\n", time.Since(start))
    return err
}
```

### Pattern 4: Short-circuit
```go
func authMiddleware(ctx *request.Context) error {
    if !isAuthorized(ctx) {
        // Don't call ctx.Next() - stop here
        return ctx.Api.Unauthorized("Not authorized")
    }
    return ctx.Next() // Continue to handler
}
```

## Middleware Order

### Per-Route Middleware
```go
router.GET("/path", handler, 
    middleware1,  // Executes 1st (before)
    middleware2,  // Executes 2nd (before)
    middleware3,  // Executes 3rd (before)
)
// Handler executes
// middleware3 after
// middleware2 after
// middleware1 after
```

### Global + Route Middleware
```go
router.Use(globalMiddleware) // Runs first

router.GET("/path", handler, routeMiddleware) // Runs second
```

### Group Middleware
```go
group := router.AddGroup("/api")
group.Use(groupMiddleware) // Applies to all routes in group

group.GET("/users", handler) // groupMiddleware → handler
```

## Use Cases

| Pattern | Use Case | Example |
|---------|----------|---------|
| **Before** | Setup, validation | Request ID, auth check |
| **After** | Cleanup, logging | Duration tracking, audit log |
| **Around** | Timing, tracing | Performance monitoring |
| **Short-circuit** | Auth, rate limit | Stop if unauthorized |

## Running

```bash
go run main.go

# Watch console output to see middleware execution order
```

## Key Takeaways

✅ Middleware wraps handlers with before/after logic  
✅ `ctx.Next()` executes the next middleware/handler  
✅ Code before `ctx.Next()` = before hook  
✅ Code after `ctx.Next()` = after hook  
✅ Middleware executes in order, then unwinds in reverse  
✅ Short-circuit by NOT calling `ctx.Next()`
