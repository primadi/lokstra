# Graceful Shutdown

Properly shut down server while completing active requests.

## Running

```bash
go run main.go
```

Server starts on `http://localhost:3090`

## Testing Graceful Shutdown

1. Start server
2. Make a request to `/slow` (takes 5 seconds)
3. Press `Ctrl+C` while request is processing
4. Watch logs - server waits for request to complete
5. Server shuts down gracefully

## Implementation

### Signal Handling
```go
shutdownSignal := make(chan os.Signal, 1)
signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

// Wait for signal
<-shutdownSignal
```

### Wait for Active Requests
```go
for {
    if activeRequests == 0 {
        log.Println("✅ All requests completed")
        return
    }
    log.Printf("⏳ Waiting... (active: %d)", activeRequests)
    time.Sleep(1 * time.Second)
}
```

### Timeout
```go
ctx, cancel := context.WithTimeout(
    context.Background(), 
    30 * time.Second,
)
defer cancel()
```

## Shutdown Flow

1. Receive shutdown signal (Ctrl+C or SIGTERM)
2. Stop accepting new connections
3. Wait for active requests to complete
4. Execute cleanup tasks
5. Exit process

## Best Practices

- Set appropriate timeout (30-60 seconds)
- Log shutdown progress
- Track active requests
- Clean up resources
- Notify monitoring systems
