# Health Check Service Example

This example demonstrates comprehensive usage of Lokstra's built-in health check service for robust application monitoring, Kubernetes integration, and operational visibility.

## What You'll Learn

- **Health Check Service Configuration**: Setting up comprehensive health monitoring
- **Custom Health Checkers**: Implementing health checks for various system components
- **Kubernetes Integration**: Readiness and liveness probes for container orchestration
- **Health Status Management**: Understanding healthy, degraded, and unhealthy states
- **Dependency Monitoring**: Checking external services and system resources
- **Operational Visibility**: Detailed health reporting and monitoring integration

## Key Features Demonstrated

### 1. Health Check Service Setup
```go
healthConfig := map[string]interface{}{
    "endpoint": "/health",
    "timeout":  "30s",
}
```

### 2. Custom Health Checkers
```go
// Database health check
healthSvc.RegisterCheck("database", func(ctx context.Context) serviceapi.HealthCheck {
    // Perform database connectivity check
    // Return detailed health status with timing and diagnostics
})

// Redis cache health check
healthSvc.RegisterCheck("redis", func(ctx context.Context) serviceapi.HealthCheck {
    // Check Redis connectivity and performance
    // Return status with cache metrics
})
```

### 3. Health Status Types
- **Healthy**: All systems operational and performing well
- **Degraded**: Functional but with performance issues or warnings
- **Unhealthy**: Critical failures affecting service availability

### 4. Kubernetes Integration
```go
// Readiness probe - checks critical dependencies
app.GET("/health/ready", readinessHandler)

// Liveness probe - checks application health
app.GET("/health/live", livenessHandler)
```

## Available Endpoints

- `GET /health` - Overall application health status
- `GET /health/ready` - Kubernetes readiness probe
- `GET /health/live` - Kubernetes liveness probe  
- `GET /health/detailed` - Comprehensive health report with system info
- `GET /health/summary` - Health status summary with counts
- `GET /health/check/:check` - Individual health check status

## Health Checks Implemented

### 1. Database Health Check
- **Purpose**: Verify database connectivity and performance
- **Checks**: Connection time, active connections, query responsiveness
- **Status Logic**: 
  - Healthy: Fast connection, normal load
  - Unhealthy: Connection timeout, database errors

### 2. Redis Cache Health Check
- **Purpose**: Monitor cache availability and performance
- **Checks**: Ping response time, memory usage, client connections
- **Status Logic**:
  - Healthy: Fast ping, normal memory usage
  - Degraded: Slow response, high memory usage
  - Unhealthy: Connection failures, unresponsive

### 3. External API Health Check
- **Purpose**: Verify external service dependencies
- **Checks**: HTTP response status, response time, rate limits
- **Status Logic**:
  - Healthy: 2xx responses, normal latency
  - Unhealthy: 5xx errors, timeouts

### 4. Disk Space Health Check
- **Purpose**: Monitor storage capacity
- **Checks**: Used space percentage, available free space
- **Status Logic**:
  - Healthy: < 80% usage
  - Degraded: 80-90% usage
  - Unhealthy: > 90% usage

### 5. Memory Usage Health Check
- **Purpose**: Track application memory consumption
- **Checks**: Memory usage percentage, garbage collection metrics
- **Status Logic**:
  - Healthy: < 85% usage
  - Degraded: 85-95% usage
  - Unhealthy: > 95% usage

## Testing the Example

### 1. Start the Server
```bash
cd cmd/examples/06_builtin/03_health_check_service
go run main.go
```

### 2. Test Health Endpoints
```bash
# Overall health status
curl http://localhost:8080/health

# Kubernetes probes
curl http://localhost:8080/health/ready
curl http://localhost:8080/health/live

# Detailed health report
curl http://localhost:8080/health/detailed

# Health summary
curl http://localhost:8080/health/summary
```

### 3. Test Individual Health Checks
```bash
# Database health
curl http://localhost:8080/health/check/database

# Redis health
curl http://localhost:8080/health/check/redis

# External API health
curl http://localhost:8080/health/check/external_api

# System resource checks
curl http://localhost:8080/health/check/disk_space
curl http://localhost:8080/health/check/memory
```

### 4. Monitor Health Over Time
```bash
# Watch overall health status
watch -n 5 'curl -s http://localhost:8080/health | jq .status'

# Monitor health summary
watch -n 10 'curl -s http://localhost:8080/health/summary | jq'

# Check for failures
for i in {1..20}; do 
  curl -s http://localhost:8080/health | jq '.status, .checks | keys'
  sleep 2
done
```

### 5. Test Status Code Responses
```bash
# Check HTTP status codes
curl -w "HTTP Status: %{http_code}\n" -s -o /dev/null http://localhost:8080/health
curl -w "HTTP Status: %{http_code}\n" -s -o /dev/null http://localhost:8080/health/ready
```

## Kubernetes Deployment

### Pod Specification
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: lokstra-app
spec:
  containers:
  - name: app
    image: lokstra-app:latest
    ports:
    - containerPort: 8080
    livenessProbe:
      httpGet:
        path: /health/live
        port: 8080
      initialDelaySeconds: 30
      periodSeconds: 10
      timeoutSeconds: 5
      failureThreshold: 3
    readinessProbe:
      httpGet:
        path: /health/ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
      timeoutSeconds: 3
      failureThreshold: 2
```

### Service Configuration
```yaml
apiVersion: v1
kind: Service
metadata:
  name: lokstra-service
spec:
  selector:
    app: lokstra-app
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Health Check Response Format

### Overall Health Response
```json
{
  "status": "healthy",
  "checks": {
    "database": {
      "name": "database",
      "status": "healthy",
      "message": "Database connection successful",
      "duration": "15ms",
      "checked_at": "2025-09-18T10:30:00Z",
      "details": {
        "connection_time": "10ms",
        "active_connections": 12,
        "max_connections": 100,
        "version": "PostgreSQL 15.2"
      }
    }
  },
  "duration": "45ms",
  "checked_at": "2025-09-18T10:30:00Z"
}
```

### Individual Check Response
```json
{
  "name": "redis",
  "status": "degraded",
  "message": "Redis responding slowly",
  "duration": "25ms",
  "checked_at": "2025-09-18T10:30:00Z",
  "details": {
    "ping_time": "20ms",
    "memory_usage": "85%",
    "connected_clients": 75,
    "warning": "high memory usage"
  }
}
```

## Key Concepts

### Health Check Design Principles
1. **Fast Execution**: Keep checks under 30 seconds
2. **Meaningful Messages**: Provide clear status descriptions
3. **Rich Details**: Include diagnostic information for troubleshooting
4. **Proper Timeouts**: Handle slow or unresponsive dependencies
5. **Error Handling**: Graceful degradation when checks fail

### Kubernetes Integration
1. **Liveness Probes**: Detect deadlocked applications (restart container)
2. **Readiness Probes**: Determine traffic readiness (remove from endpoints)
3. **Startup Probes**: Handle slow-starting applications
4. **Probe Configuration**: Set appropriate timeouts and thresholds

### Monitoring Integration
1. **Status Codes**: Use HTTP status codes for automated monitoring
2. **Metrics Export**: Export health metrics to monitoring systems
3. **Alerting**: Set up alerts based on health status changes
4. **Dashboards**: Visualize health trends and dependency status

### Best Practices
1. **Dependency Categorization**: Separate critical from non-critical checks
2. **Cascading Failures**: Prevent health check cascades
3. **Circuit Breakers**: Implement circuit breaker patterns for external services
4. **Historical Tracking**: Maintain health check history for trend analysis

This example provides a comprehensive foundation for implementing production-ready health monitoring in Lokstra applications with full Kubernetes integration and operational visibility.