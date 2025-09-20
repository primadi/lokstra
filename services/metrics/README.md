# Lokstra Metrics Service

Lokstra Metrics Service provides Prometheus-based metrics collection and exposure for monitoring your applications.

## Features

- **Counter Metrics**: Track incrementing values like request counts, errors
- **Gauge Metrics**: Track current values like active connections, memory usage
- **Histogram Metrics**: Track distributions like request duration, response sizes
- **Custom Labels**: Add dimensional labels to metrics for better filtering
- **HTTP Endpoint**: Expose metrics via `/metrics` endpoint for Prometheus scraping
- **Graceful Shutdown**: Clean resource management

## Configuration

Configure the metrics service in your `lokstra.yaml`:

```yaml
serviceConfig:
  lokstra.metrics:
    enabled: true
    host: "0.0.0.0"
    port: 8080
    timeout: 30s
    # Custom default labels applied to all metrics
    defaultLabels:
      environment: "production"
      service: "my-app"
    # Default histogram buckets for response time metrics
    histogramBuckets: [0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0]
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | boolean | `true` | Enable/disable metrics collection |
| `host` | string | `"0.0.0.0"` | Host to bind HTTP server |
| `port` | int | `8080` | Port for metrics endpoint |
| `timeout` | string | `"30s"` | HTTP server timeout |
| `defaultLabels` | map | `{}` | Labels applied to all metrics |
| `histogramBuckets` | []float64 | Prometheus defaults | Custom histogram buckets |

## Usage

### Basic Usage

```go
package main

import (
    "github.com/primadi/lokstra/serviceapi"
    "github.com/primadi/lokstra/services/metrics"
)

func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    // Get metrics service
    metricsService := lokstra.GetService[serviceapi.Metrics](regCtx, "metrics")
    
    // Increment a counter
    metricsService.IncCounter("http_requests_total", map[string]string{
        "method": "GET",
        "endpoint": "/api/users",
    })
    
    // Set a gauge value
    metricsService.SetGauge("active_connections", 42, map[string]string{
        "server": "web-01",
    })
    
    // Record histogram observation
    metricsService.ObserveHistogram("request_duration_seconds", 0.25, map[string]string{
        "method": "POST",
        "status": "200",
    })
}
```

### Counter Metrics

Counters only go up and are ideal for tracking cumulative values:

```go
// Track requests
metricsService.IncCounter("http_requests_total", map[string]string{
    "method": "GET",
    "status": "200",
})

// Track errors
metricsService.IncCounter("errors_total", map[string]string{
    "type": "validation",
    "severity": "warning",
})
```

### Gauge Metrics

Gauges can go up and down, perfect for current state values:

```go
// Active connections
metricsService.SetGauge("active_connections", float64(connectionCount), nil)

// Memory usage in bytes
metricsService.SetGauge("memory_usage_bytes", float64(memUsage), map[string]string{
    "type": "heap",
})

// Queue length
metricsService.SetGauge("queue_length", float64(queueSize), map[string]string{
    "queue": "processing",
})
```

### Histogram Metrics

Histograms track distributions and provide percentiles:

```go
// Request duration
start := time.Now()
// ... process request ...
duration := time.Since(start).Seconds()
metricsService.ObserveHistogram("request_duration_seconds", duration, map[string]string{
    "endpoint": "/api/users",
    "method": "GET",
})

// Response size
metricsService.ObserveHistogram("response_size_bytes", float64(responseSize), map[string]string{
    "content_type": "application/json",
})
```

### HTTP Middleware Integration

```go
func metricsMiddleware(metricsService serviceapi.Metrics) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Track request
            metricsService.IncCounter("http_requests_total", map[string]string{
                "method": r.Method,
                "endpoint": r.URL.Path,
            })
            
            next.ServeHTTP(w, r)
            
            // Track duration
            duration := time.Since(start).Seconds()
            metricsService.ObserveHistogram("request_duration_seconds", duration, map[string]string{
                "method": r.Method,
                "endpoint": r.URL.Path,
            })
        })
    }
}
```

## Prometheus Integration

### Scraping Configuration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'lokstra-app'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Example Queries

```promql
# Request rate
rate(http_requests_total[5m])

# 95th percentile response time
histogram_quantile(0.95, rate(request_duration_seconds_bucket[5m]))

# Error rate
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])

# Active connections
active_connections
```

## Best Practices

### 1. Use Consistent Label Names
```go
// Good - consistent labels
metricsService.IncCounter("http_requests_total", map[string]string{
    "method": "GET",
    "endpoint": "/api/users",
    "status": "200",
})

// Avoid - inconsistent labels
metricsService.IncCounter("http_requests_total", map[string]string{
    "http_method": "GET",  // Different from above
    "path": "/api/users",  // Different from above
})
```

### 2. Avoid High Cardinality Labels
```go
// Good - low cardinality
metricsService.IncCounter("requests_total", map[string]string{
    "method": "GET",
    "status": "200",
    "endpoint": "/api/users", // Fixed set of endpoints
})

// Avoid - high cardinality
metricsService.IncCounter("requests_total", map[string]string{
    "user_id": "12345",    // Too many possible values
    "request_id": "abc123", // Unique per request
})
```

### 3. Use Appropriate Metric Types
```go
// Counters for cumulative values
metricsService.IncCounter("bytes_processed_total", labels)

// Gauges for current state
metricsService.SetGauge("active_connections", value, labels)

// Histograms for distributions
metricsService.ObserveHistogram("request_duration_seconds", duration, labels)
```

### 4. Meaningful Metric Names
Follow Prometheus naming conventions:
- Use `_total` suffix for counters
- Use `_seconds` for time durations
- Use `_bytes` for byte measurements
- Use snake_case for metric names

## Error Handling

The metrics service handles errors gracefully:

```go
// Service continues to work even if metric operations fail
metricsService.IncCounter("invalid_metric_name!", labels) // Logs error, continues

// Check if service is enabled
if summary := metricsService.GetMetricsSummary(); summary["enabled"] == "true" {
    // Metrics are actively being collected
}
```

## Monitoring the Metrics Service

Monitor the metrics service itself:

```go
// Track metric operations
metricsService.IncCounter("metrics_operations_total", map[string]string{
    "operation": "inc_counter",
    "status": "success",
})

// Track service health
metricsService.SetGauge("metrics_service_up", 1, nil)
```

## Performance Considerations

- Metrics collection has minimal overhead
- Label cardinality affects memory usage
- Consider sampling for high-frequency metrics
- Use histogram buckets appropriate for your data

## Integration Examples

See the `examples/` directory for complete integration examples:
- Basic web application with request metrics
- Database connection pool monitoring
- Custom business metrics
- Multi-service metrics aggregation
