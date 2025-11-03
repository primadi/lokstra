# Production Configuration Patterns

Best practices and patterns for production deployments.

## Running

```bash
# Development mode
APP_ENV=development go run main.go

# Production mode
APP_ENV=production go run main.go production
```

Server starts on `http://localhost:3060`

## Production Patterns

### Health Checks
```go
func HealthHandler() map[string]any {
    return map[string]any{
        "status": "ok",
        "timestamp": time.Now(),
    }
}
```

### Metrics
```go
func MetricsHandler() map[string]any {
    return map[string]any{
        "requests_total": counter,
        "errors_total": errors,
    }
}
```

### Graceful Shutdown
```go
shutdownTimeout := 60 * time.Second
app.Run(shutdownTimeout)
```

## Best Practices

- ✅ Health check endpoints for load balancers
- ✅ Metrics endpoints for monitoring
- ✅ Graceful shutdown (longer timeout in prod)
- ✅ Structured logging
- ✅ Configuration validation
- ✅ Environment-based settings
- ✅ Request timeouts
- ✅ Connection pooling
