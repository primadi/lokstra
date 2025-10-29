# Logging Middleware Example

Demonstrates automatic request logging using Lokstra's built-in request_logger middleware.

## What You'll Learn

- How to add middleware globally to a router
- Automatic request/response logging
- See logs for successful and error responses
- Understand middleware execution flow

## Running

```bash
cd docs/01-essentials/03-middleware/examples/01-logging-middleware
go run main.go
```

## Testing

Use `test.http` or curl:

```bash
# Test various endpoints
curl http://localhost:3000/users
curl http://localhost:3000/products
curl -X POST http://localhost:3000/users
curl http://localhost:3000/error
```

## Expected Console Output

```
🚀 Logging Middleware Demo
📝 All requests will be logged automatically

Try these endpoints:
  GET  /users
  GET  /products
  POST /users
  GET  /error     (will log error)

Watch the console for request logs!

Server: http://localhost:3000

[request_logger] GET /users 200 OK (1.2ms)
[request_logger] GET /products 200 OK (0.8ms)
[request_logger] POST /users 200 OK (2.1ms)
[request_logger] GET /error 500 Internal Server Error (0.5ms)
```

## How It Works

```go
// Add middleware to router - applies to ALL routes
router.Use(request_logger.Middleware(nil))

// All these routes will be logged automatically
router.GET("/users", handler)
router.GET("/products", handler)
router.POST("/users", handler)
```

## Configuration Options

The `request_logger.Middleware()` accepts a config map:

```go
// Default (nil) - logs method, path, status, duration
router.Use(request_logger.Middleware(nil))

// Custom config (if supported)
router.Use(request_logger.Middleware(map[string]any{
    "show_body": false,
    "show_headers": true,
    "color": true,
}))
```

## Key Takeaways

- ✅ Middleware runs **before** your handler
- ✅ Global middleware applies to **all routes** in the router
- ✅ Logs include: method, path, status code, duration
- ✅ Works with both successful and error responses
- ✅ Zero code needed in handlers - automatic!
