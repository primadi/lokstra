# Lokstra Health Check Service Example

A comprehensive example demonstrating the Lokstra Health Check Service with YAML configuration and Kubernetes-ready health endpoints.

## 🚀 Quick Start

```bash
go run main.go
```

The application will start at `http://localhost:8080`

## 📊 Health Check Endpoints

### Basic Health Check
```bash
GET /health
```
**Purpose**: Primary health endpoint for load balancers and basic monitoring.
- Returns overall health status of all registered checks
- HTTP 200 for healthy/degraded, HTTP 503 for unhealthy
- Lightweight response suitable for frequent polling

**Response Example**:
```json
{
  "status": "healthy",
  "checks": {
    "application": {
      "name": "application",
      "status": "healthy",
      "message": "Application is running",
      "duration": "1.234ms",
      "checked_at": "2025-08-21T10:30:00Z"
    }
  },
  "duration": "5.678ms",
  "checked_at": "2025-08-21T10:30:00Z"
}
```

### Detailed Health Information
```bash
GET /health/detailed
```
**Purpose**: Comprehensive health information with full diagnostic details.
- Includes detailed information for each health check
- Contains error messages, metrics, and diagnostic data
- Suitable for debugging and administrative interfaces

**Use Cases**:
- Dashboard displays
- Administrative interfaces
- Troubleshooting and debugging
- Manual health investigations

### List Available Health Checks
```bash
GET /health/list
```
**Purpose**: Lists all registered health checks without executing them.
- Fast response showing available health check names
- Useful for discovery and documentation
- No actual health checking performed

### Individual Health Check
```bash
GET /health/check/:name
```
**Purpose**: Execute and return results for a specific health check.
- Targeted health checking for specific components
- Useful for isolating issues to specific services
- Returns HTTP 404 if health check doesn't exist

**Examples**:
```bash
curl http://localhost:8080/health/check/database
curl http://localhost:8080/health/check/memory
curl http://localhost:8080/health/check/external_api
```

## 🔧 Kubernetes Integration

### Liveness Probe
```bash
GET /health/liveness
```
**Purpose**: Determines if the application is alive and should be restarted.
- Always returns HTTP 200 if the application is running
- Kubernetes will restart the pod if this fails
- Should only fail if the application is completely broken

**Kubernetes Configuration**:
```yaml
livenessProbe:
  httpGet:
    path: /health/liveness
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### Readiness Probe
```bash
GET /health/readiness
```
**Purpose**: Determines if the application is ready to receive traffic.
- Returns HTTP 200 when ready, HTTP 503 when not ready
- Kubernetes will remove pod from service endpoints if this fails
- Should fail if dependencies are unavailable

**Kubernetes Configuration**:
```yaml
readinessProbe:
  httpGet:
    path: /health/readiness
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

### Complete Kubernetes Deployment Example
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lokstra-health-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: lokstra-health-app
  template:
    metadata:
      labels:
        app: lokstra-health-app
    spec:
      containers:
      - name: app
        image: lokstra-health-app:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health/liveness
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health/readiness
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        startupProbe:
          httpGet:
            path: /health/readiness
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 30
---
apiVersion: v1
kind: Service
metadata:
  name: lokstra-health-service
spec:
  selector:
    app: lokstra-health-app
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

## 📈 Prometheus Metrics Integration

### Metrics Endpoint
```bash
GET /health/metrics
```
**Purpose**: Exports health check metrics in Prometheus format.
- Compatible with Prometheus scraping
- Includes individual check status and duration metrics
- Suitable for monitoring and alerting

**Metrics Exported**:
- `health_status` - Overall health status (1=healthy, 0.5=degraded, 0=unhealthy)
- `health_checks_total` - Total number of health checks
- `health_check_status{name="..."}` - Individual check status
- `health_check_duration_seconds{name="..."}` - Check execution time

**Prometheus Configuration**:
```yaml
scrape_configs:
  - job_name: 'lokstra-health'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /health/metrics
    scrape_interval: 15s
```

**Example Alerting Rules**:
```yaml
groups:
- name: lokstra-health
  rules:
  - alert: ApplicationUnhealthy
    expr: health_status < 1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "Lokstra application is unhealthy"
      description: "Health status is {{ $value }} for 2 minutes"

  - alert: HealthCheckFailing
    expr: health_check_status < 1
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Health check {{ $labels.name }} is failing"
      description: "Check {{ $labels.name }} status is {{ $value }}"
```

## 🏥 Available Health Checks

### 1. Application Health Check
**Name**: `application`
**Purpose**: Basic application health and metadata
- Always healthy unless application is shutting down
- Provides application name, version, and uptime
- Good baseline check for application status

### 2. Memory Health Check  
**Name**: `memory`
**Purpose**: Monitor memory usage against configurable thresholds
- **Healthy**: < 70% of threshold
- **Degraded**: 70-90% of threshold  
- **Unhealthy**: > 90% of threshold
- Uses Go runtime memory statistics

### 3. Disk Health Check
**Name**: `disk`
**Purpose**: Monitor disk space usage
- Configurable path and threshold percentage
- Cross-platform implementation (Windows/Unix)
- Critical for preventing disk space issues

### 4. Database Health Check
**Name**: `database` (Simulated)
**Purpose**: Database connectivity and performance
- Connection testing
- Response time monitoring
- Connection pool status

### 5. External API Health Check
**Name**: `external_api` (Simulated)
**Purpose**: External service dependencies
- API endpoint availability
- Response time monitoring
- Retry logic and failure detection

### 6. Business Logic Health Check
**Name**: `business_logic`
**Purpose**: Custom business rule validation
- Data consistency checks
- Business process monitoring
- Request rate monitoring

### 7. Periodic Tasks Health Check
**Name**: `periodic_tasks`
**Purpose**: Background job monitoring
- Job scheduler health
- Queue size monitoring
- Failed job tracking

## 🧪 Testing and Simulation

### Simulate Error Conditions
```bash
# Activate error simulation
curl -X POST http://localhost:8080/api/simulate-error

# Check degraded health
curl http://localhost:8080/health

# Check specific failing components
curl http://localhost:8080/health/detailed

# Recover from errors
curl -X POST http://localhost:8080/api/recover
```

### Application Status
```bash
# Application information
curl http://localhost:8080/

# API status
curl http://localhost:8080/api/status
```

## ⚙️ Configuration

The application uses YAML configuration (`health-check.yaml`):

```yaml
# Server configuration
server:
  name: health-check-example
  global_setting:
    log_level: info

# Health service configuration
services:
  - type: health_check
    name: health-service
    config:
      enabled: true

# Application routes
apps:
  - name: health-app
    address: :8080
    routes:
      - method: GET
        path: /health
        handler: health.check
      # ... additional routes
```

## 🏗️ Architecture

### Health Status Levels
1. **Healthy** - All systems operating normally
2. **Degraded** - Some issues detected but service still functional
3. **Unhealthy** - Critical issues requiring immediate attention

### Response Codes
- **HTTP 200** - Healthy or degraded (service operational)
- **HTTP 503** - Unhealthy (service unavailable)
- **HTTP 404** - Health check not found
- **HTTP 500** - Internal server error

### Best Practices for Kubernetes

1. **Liveness Probe**: Should only fail if application needs restart
   - Use simple checks (basic connectivity)
   - Avoid dependency checks
   - Set appropriate timeouts

2. **Readiness Probe**: Should fail if application can't serve traffic
   - Include dependency checks
   - Check database connectivity
   - Validate external service availability

3. **Startup Probe**: For slow-starting applications
   - More lenient failure threshold
   - Prevents premature restarts during startup

4. **Monitoring Integration**:
   - Use `/health/metrics` for Prometheus
   - Set up alerting on health status changes
   - Monitor individual component health

## 🔍 Troubleshooting

### Common Issues

1. **503 Service Unavailable**
   - Check `/health/detailed` for specific failures
   - Verify external dependencies
   - Check resource usage (memory, disk)

2. **Degraded Performance**
   - Monitor `/health/metrics` for trends
   - Check individual component response times
   - Investigate resource constraints

3. **Kubernetes Pod Restarts**
   - Review liveness probe configuration
   - Check application logs
   - Verify probe timeouts and thresholds

### Debug Commands
```bash
# Get detailed health information
curl -s http://localhost:8080/health/detailed | jq

# Check specific component
curl -s http://localhost:8080/health/check/database | jq

# Monitor metrics
curl http://localhost:8080/health/metrics

# List all available checks
curl -s http://localhost:8080/health/list | jq
```

## 🎯 Key Features

- ✅ **Zero Configuration**: Auto-registration with sensible defaults
- ✅ **Kubernetes Ready**: Native liveness and readiness probes
- ✅ **Prometheus Integration**: Built-in metrics export
- ✅ **Error Simulation**: Test degraded and unhealthy conditions
- ✅ **Multiple Status Types**: Healthy, degraded, and unhealthy states
- ✅ **Cross-Platform**: Windows and Unix compatibility
- ✅ **Production Ready**: Real system monitoring capabilities
- ✅ **YAML Configuration**: Declarative setup and routing
