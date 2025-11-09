# Multiple Servers

Run multiple Lokstra servers concurrently in the same application.

## Running

```bash
go run main.go
```

Three servers will start:
- **API Server**: `http://localhost:3080` - Business API endpoints
- **Admin Server**: `http://localhost:3081` - Admin/management endpoints
- **Metrics Server**: `http://localhost:3082` - Monitoring endpoints

## Architecture

```go
func main() {
    var wg sync.WaitGroup
    
    // Start multiple servers
    wg.Add(1)
    go func() {
        defer wg.Done()
        apiApp.Run(0)
    }()
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        adminApp.Run(0)
    }()
    
    wg.Wait()
}
```

## Servers

### API Server (3080)
Business logic endpoints:
- `/api/users` - User management
- `/api/orders` - Order processing

### Admin Server (3081)
Administrative endpoints:
- `/admin/stats` - System statistics
- `/admin/config` - Configuration management

### Metrics Server (3082)
Monitoring endpoints:
- `/metrics` - Application metrics
- `/health` - Health checks

## Use Cases

- **Separation of concerns**: Different ports for different audiences
- **Security**: Admin endpoints on internal port only
- **Scaling**: Scale each service independently
- **Load balancing**: Route traffic based on purpose
- **Monitoring**: Dedicated metrics endpoint
