# Lifecycle Hooks

Handle application startup and shutdown events.

## Running

```bash
go run main.go
```

Server starts on `http://localhost:3070`

Press `Ctrl+C` to trigger shutdown hook.

## Hooks Implemented

### Startup Hook
```go
func OnStartup() {
    log.Println("🚀 Application starting...")
    startTime = time.Now()
    // Initialize resources
}
```

Called when:
- Application starts
- Before accepting connections

### Shutdown Hook
```go
func OnShutdown() {
    uptime := time.Since(startTime)
    log.Printf("⏱️ Uptime: %v", uptime)
    // Cleanup resources
}
```

Called when:
- Application receives shutdown signal
- After all requests complete

## Use Cases

- Initialize database connections
- Load configuration
- Start background workers
- Log application statistics
- Cleanup resources
- Close connections
- Save state

## Pattern

```go
func main() {
    OnStartup()           // Initialize
    defer OnShutdown()    // Cleanup
    
    app.Run(timeout)      // Run app
}
```
