# Health Checks

Implement comprehensive health check endpoints for monitoring and orchestration.

## Running

```bash
go run main.go
```

Server starts on `http://localhost:3100`

## Health Check Types

### Basic Health (`/health`)
Simple up/down status:
```json
{"status": "healthy"}
```

### Detailed Health (`/health/detailed`)
Comprehensive system status:
```json
{
  "status": "healthy",
  "checks": {
    "database": {"status": "healthy", "latency": "2ms"},
    "cache": {"status": "healthy", "latency": "1ms"},
    "disk": {"status": "healthy", "usage": "45%"},
    "memory": {"status": "healthy", "usage": "512MB"}
  }
}
```

### Readiness (`/readiness`)
Ready to serve traffic?
```json
{
  "ready": true,
  "services": {
    "database": true,
    "cache": true
  }
}
```

### Liveness (`/liveness`)
Is process alive?
```json
{
  "alive": true,
  "uptime": 3600
}
```

## Use Cases

### Load Balancers
Use `/health` or `/readiness`:
- Remove unhealthy instances from pool
- Route traffic only to healthy nodes

### Kubernetes
Use readiness and liveness probes:
```yaml
readinessProbe:
  httpGet:
    path: /readiness
    port: 3100
livenessProbe:
  httpGet:
    path: /liveness
    port: 3100
```

### Monitoring Systems
Use `/health/detailed`:
- Track dependency health
- Alert on degraded services
- Monitor system resources

## Testing Failures

Simulate issues with toggle endpoints:
- `/toggle-health` - Toggle overall health
- `/toggle-db` - Simulate database issue

Watch how health checks respond to failures.
