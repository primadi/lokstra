# Health Check Service

## Overview

The Health Check service provides comprehensive health monitoring capabilities for Lokstra applications, specifically designed for Kubernetes deployments and monitoring systems. It supports multiple health check types, concurrent execution, and provides various output formats including JSON and Prometheus metrics.

## Features

- **Auto-Registration**: HTTP handlers automatically registered via module system
- **Multiple Health Status Types**: Healthy, Degraded, Unhealthy
- **Concurrent Health Checks**: All checks run in parallel for optimal performance
- **Kubernetes Integration**: Built-in liveness and readiness probe endpoints
- **Prometheus Metrics**: Export health metrics in Prometheus format
- **Extensible**: Easy to add custom health checks
- **Timeout Support**: Configurable timeouts for health check execution
- **Built-in Checkers**: Database, Redis, Memory, and Disk usage checkers
- **Zero Configuration**: Works out-of-the-box with defaults package

## Architecture

```
serviceapi/health_check.go     - Interface definitions and types only
services/health_check/
├── service.go                 - Main health service implementation
├── module.go                  - Module registration
├── handlers.go                - HTTP handlers for health endpoints
├── checkers.go                - Built-in health checker implementations
└── service_test.go           - Comprehensive tests
```

## Health Status Types

- **Healthy**: All systems operating normally
- **Degraded**: Service is running but with reduced performance
- **Unhealthy**: Service has critical issues

## Quick Start

### Simple Usage (Auto-Registration)

With the enhanced module system, health check endpoints are automatically registered:

```go
package main

import (
    "context"
    "log"
    
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/serviceapi"
    "github.com/primadi/lokstra/services/health_check"
)

func main() {
    // 1. Setup registration context (health module auto-registers service & handlers)
    regCtx := lokstra.NewGlobalRegistrationContext()
    
    // 2. Get auto-created health service
    healthService, _ := regCtx.GetService("health_check.default")
    health := healthService.(serviceapi.HealthService)
    
    // 3. Register health checks (handlers already auto-registered!)
    health.RegisterCheck("app", health_check.ApplicationHealthChecker("my-app"))
    health.RegisterCheck("memory", health_check.MemoryHealthChecker(512)) // 512MB limit
    
    // 4. Start app - all health endpoints automatically available!
    app := lokstra.NewApp(regCtx, "health-example", ":8080")
    
    log.Println("🏥 Health endpoints auto-registered:")
    log.Println("  - GET /health          - Main health check")
    log.Println("  - GET /health/liveness - Kubernetes liveness probe") 
    log.Println("  - GET /health/readiness - Kubernetes readiness probe")
    log.Println("  - GET /health/detailed - Detailed health information")
    log.Println("  - GET /health/list     - List all health checks")
    log.Println("  - GET /health/check/{name} - Individual check")
    log.Println("  - GET /health/metrics  - Prometheus metrics")
    
    app.Start()
}
```

### Manual Registration (Advanced)

For custom setups or when you need fine-grained control:

#### 1. Register the Service

```go
import "github.com/primadi/lokstra"

// Create registration context with defaults (includes health_check)
regCtx := lokstra.NewGlobalRegistrationContext()

// Create health service
healthService, err := regCtx.CreateService("health_check", "health", nil)
health := healthService.(serviceapi.HealthService)
```

#### 2. Register Health Checks

```go
import "github.com/primadi/lokstra/services/health_check"

// Register application health check
health.RegisterCheck("app", health_check.ApplicationHealthChecker("my-app"))

// Register database health check
if dbPool, ok := dbService.(serviceapi.DbPool); ok {
    health.RegisterCheck("database", health_check.DatabaseHealthChecker(dbPool))
}

// Register Redis health check
if redis, ok := redisService.(serviceapi.Redis); ok {
    health.RegisterCheck("redis", health_check.RedisHealthChecker(redis))
}

// Register memory health check
health.RegisterCheck("memory", health_check.MemoryHealthChecker(1024)) // 1GB limit

// Register disk health check
health.RegisterCheck("disk", health_check.DiskHealthChecker("/tmp", 80.0)) // 80% limit
```

#### 3. Set Up HTTP Endpoints (Manual)

```go
// Register health endpoints
regCtx.RegisterHandler("health", func(ctx *lokstra.Context) error {
    result := health.CheckHealthWithTimeout(30 * time.Second)
    if result.Status == serviceapi.HealthStatusUnhealthy {
        ctx.StatusCode = 503
    }
    return ctx.Ok(result)
})

regCtx.RegisterHandler("health.liveness", func(ctx *lokstra.Context) error {
    return ctx.Ok(map[string]any{
        "status":     "healthy",
        "service":    "lokstra",
        "checked_at": time.Now(),
    })
})

regCtx.RegisterHandler("health.readiness", func(ctx *lokstra.Context) error {
    isReady := health.IsHealthy(context.Background())
    response := map[string]any{
        "status": "ready",
        "ready":  isReady,
    }
    if !isReady {
        ctx.StatusCode = 503
        response["status"] = "not_ready"
    }
    return ctx.Ok(response)
})
```

## API Reference

### Core Interface

```go
type HealthService interface {
    service.Service
    
    // Register/unregister health checks
    RegisterCheck(name string, checker HealthChecker)
    UnregisterCheck(name string)
    
    // Execute health checks
    CheckHealth(ctx context.Context) HealthResult
    CheckHealthWithTimeout(timeout time.Duration) HealthResult
    IsHealthy(ctx context.Context) bool
    
    // Individual check operations
    GetCheck(ctx context.Context, name string) (HealthCheck, bool)
    ListChecks() []string
}
```

### Health Check Structure

```go
type HealthCheck struct {
    Name        string            `json:"name"`
    Status      HealthStatus      `json:"status"`
    Message     string            `json:"message,omitempty"`
    Details     map[string]any    `json:"details,omitempty"`
    Duration    time.Duration     `json:"duration"`
    CheckedAt   time.Time         `json:"checked_at"`
    Error       string            `json:"error,omitempty"`
}
```

### Health Result Structure

```go
type HealthResult struct {
    Status    HealthStatus            `json:"status"`
    Checks    map[string]HealthCheck  `json:"checks"`
    Duration  time.Duration           `json:"duration"`
    CheckedAt time.Time               `json:"checked_at"`
}
```

## HTTP Endpoints

### Basic Health Check
- **GET** `/health` - Main health check endpoint
- **Response**: Complete health status with all checks
- **Status Codes**: 200 (healthy/degraded), 503 (unhealthy)

### Kubernetes Probes
- **GET** `/health/liveness` - Kubernetes liveness probe
- **GET** `/health/readiness` - Kubernetes readiness probe
- **Status Codes**: 200 (ready), 503 (not ready)

### Detailed Information
- **GET** `/health/detailed` - Detailed health information with summary
- **GET** `/health/list` - List all registered health checks
- **GET** `/health/check/{name}` - Get specific health check result

### Monitoring Integration
- **GET** `/health/metrics` - Prometheus metrics format

## Built-in Health Checkers

All built-in checkers are available in the `health_check` package:

### Database Health Checker
```go
import "github.com/primadi/lokstra/services/health_check"

checker := health_check.DatabaseHealthChecker(dbPool)
health.RegisterCheck("database", checker)
```

### Redis Health Checker
```go
checker := health_check.RedisHealthChecker(redis)
health.RegisterCheck("redis", checker)
```

### Memory Health Checker
```go
checker := health_check.MemoryHealthChecker(1024) // 1GB limit
health.RegisterCheck("memory", checker)
```

### Disk Health Checker
```go
checker := health_check.DiskHealthChecker("/tmp", 80.0) // 80% limit
health.RegisterCheck("disk", checker)
```

### Application Health Checker
```go
checker := health_check.ApplicationHealthChecker("my-app")
health.RegisterCheck("application", checker)
```

## Custom Health Checkers

Create custom health checkers by implementing the `HealthChecker` function:

```go
import "github.com/primadi/lokstra/services/health_check"

// Using the CustomHealthChecker helper
customChecker := health_check.CustomHealthChecker("my_service", func(ctx context.Context) (bool, string, map[string]any) {
    // Perform your health check logic here
    isHealthy := checkYourService()
    message := "Service is healthy"
    details := map[string]any{
        "version": "1.0.0",
        "uptime": time.Now().Sub(startTime).String(),
    }
    
    if !isHealthy {
        message = "Service has issues"
        details["error"] = "Connection failed"
    }
    
    return isHealthy, message, details
})

health.RegisterCheck("my_service", customChecker)

// Or implement directly
directChecker := func(ctx context.Context) serviceapi.HealthCheck {
    start := time.Now()
    
    // Perform your health check logic here
    isHealthy := checkYourService()
    
    status := serviceapi.HealthStatusHealthy
    message := "Service is healthy"
    
    if !isHealthy {
        status = serviceapi.HealthStatusUnhealthy
        message = "Service has issues"
    }
    
    return serviceapi.HealthCheck{
        Name:      "custom_service",
        Status:    status,
        Message:   message,
        CheckedAt: start,
        Duration:  time.Since(start),
    }
}

health.RegisterCheck("custom_service", directChecker)
```

## Kubernetes Configuration

### Deployment Example
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lokstra-app
spec:
  template:
    spec:
      containers:
      - name: app
        image: lokstra-app:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health/liveness
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/readiness
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Service Monitor for Prometheus
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: lokstra-health
spec:
  selector:
    matchLabels:
      app: lokstra-app
  endpoints:
  - port: http
    path: /health/metrics
    interval: 30s
```

## Response Examples

### Health Check Response
```json
{
  "status": "healthy",
  "checks": {
    "database": {
      "name": "database",
      "status": "healthy",
      "message": "Database connection is healthy",
      "duration": 5000000,
      "checked_at": "2025-08-21T10:30:00Z"
    },
    "redis": {
      "name": "redis",
      "status": "healthy",
      "message": "Redis connection is healthy",
      "duration": 2000000,
      "checked_at": "2025-08-21T10:30:00Z"
    }
  },
  "duration": 15000000,
  "checked_at": "2025-08-21T10:30:00Z"
}
```

### Prometheus Metrics Response
```
lokstra_health_status 1
lokstra_health_check_duration_seconds 0.015
lokstra_health_checks_total 2
lokstra_health_check_status{name="database"} 1
lokstra_health_check_duration_seconds{name="database"} 0.005
lokstra_health_check_status{name="redis"} 1
lokstra_health_check_duration_seconds{name="redis"} 0.002
```

## Testing

Run the health check tests:

```bash
go test ./services/health_check -v
```

## Example Implementation

See the complete example in `cmd/examples/health_check/main.go` which demonstrates:
- Service setup and registration
- Multiple health check types
- HTTP endpoint configuration
- Integration with database and Redis services

## Best Practices

1. **Keep checks lightweight**: Health checks should be fast and not impact performance
2. **Use appropriate timeouts**: Set reasonable timeouts to prevent hanging
3. **Monitor dependencies separately**: Use separate checks for each external dependency
4. **Implement graceful degradation**: Use degraded status for non-critical issues
5. **Include meaningful messages**: Provide clear error messages for debugging
6. **Use structured logging**: Log health check results for monitoring

## Integration with Monitoring

The health check service integrates seamlessly with:
- **Kubernetes**: Liveness and readiness probes
- **Prometheus**: Metrics collection
- **Grafana**: Visualization dashboards
- **AlertManager**: Alert notifications

This service provides a robust foundation for application health monitoring in cloud-native environments.
