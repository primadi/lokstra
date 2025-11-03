# Error Recovery Example

Demonstrates panic recovery and error handling in middleware.

## Running

```bash
go run main.go
```

Server starts on `http://localhost:3003`

## Patterns

### Recovery Middleware

```go
func RecoveryMiddleware(c *request.Context) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("PANIC: %v", r)
        }
    }()
    return c.Next()
}
```

### Error Logging

```go
func ErrorLoggingMiddleware(c *request.Context) error {
    err := c.Next()
    if err != nil {
        log.Printf("ERROR: %v", err)
    }
    return err
}
```
