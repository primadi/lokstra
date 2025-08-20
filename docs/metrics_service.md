# Metrics Service Implementation

The Lokstra metrics service provides comprehensive metrics collection using Prometheus client library.

## Features

- **Multiple Metric Types**: Counters, Gauges, and Histograms
- **Automatic Label Management**: Dynamic label creation and management
- **Custom Buckets**: Configurable histogram buckets for performance metrics
- **Registry Management**: Isolated Prometheus registry for each service instance
- **HTTP Handler**: Built-in HTTP handler for metrics endpoint
- **Runtime Metrics**: Optional Go runtime and process metrics collection

## Interface Implementation

The service implements `serviceapi.Metrics` interface with three main methods:

```go
type Metrics interface {
    IncCounter(name string, labels Labels)
    ObserveHistogram(name string, value float64, labels Labels)
    SetGauge(name string, value float64, labels Labels)
}
```

## Configuration

### Basic Configuration

```yaml
services:
  - name: "metrics"
    type: "lokstra.metrics"
    config:
      enabled: true
      endpoint: "/metrics"
```

### Advanced Configuration

```yaml
services:
  - name: "metrics"
    type: "lokstra.metrics"
    config:
      enabled: true
      endpoint: "/metrics"
      namespace: "myapp"
      subsystem: "api"
      collect_interval: "15s"
      timeout: "10s"
      
      # Custom histogram buckets
      buckets: [0.001, 0.01, 0.1, 1.0, 10.0]
      
      # Constant labels for all metrics
      labels:
        service: "user_management"
        version: "1.0.0"
        
      # Runtime metrics
      include_go_metrics: true
      include_process_metrics: true
      
      # Separate HTTP server (optional)
      host: "localhost"
      port: 9090
```

## Configuration Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `enabled` | boolean | `true` | Enable/disable metrics collection |
| `endpoint` | string | `"/metrics"` | HTTP endpoint path for metrics |
| `namespace` | string | `""` | Prometheus namespace prefix |
| `subsystem` | string | `""` | Prometheus subsystem name |
| `buckets` | array | `prometheus.DefBuckets` | Histogram bucket boundaries |
| `labels` | object | `{}` | Constant labels for all metrics |
| `collect_interval` | string | `"15s"` | Collection interval (Go duration) |
| `host` | string | `"localhost"` | Host for separate metrics server |
| `port` | integer | `0` | Port for separate metrics server (0 = disabled) |
| `timeout` | string | `"10s"` | Operation timeout |
| `include_go_metrics` | boolean | `true` | Include Go runtime metrics |
| `include_process_metrics` | boolean | `true` | Include process metrics |

## Usage Examples

### Counter Metrics

```go
// Increment request counter
metrics.IncCounter("http_requests_total", serviceapi.Labels{
    "method": "GET",
    "status": "200",
    "endpoint": "/users",
})
```

### Gauge Metrics

```go
// Set current active connections
metrics.SetGauge("active_connections", 42, serviceapi.Labels{
    "pool": "database",
})
```

### Histogram Metrics

```go
// Record request duration
start := time.Now()
// ... process request ...
duration := time.Since(start).Seconds()

metrics.ObserveHistogram("http_request_duration_seconds", duration, serviceapi.Labels{
    "method": "POST",
    "endpoint": "/users",
})
```

## Metric Naming Convention

The service follows Prometheus naming conventions:

- **Namespace**: Optional prefix (e.g., `myapp_`)
- **Subsystem**: Optional subsystem (e.g., `http_`)
- **Metric Name**: Descriptive name (e.g., `requests_total`)
- **Full Name**: `{namespace}_{subsystem}_{name}` (e.g., `myapp_http_requests_total`)

## Built-in Metrics

When enabled, the service automatically includes:

### Go Runtime Metrics
- `go_memstats_*` - Memory statistics
- `go_goroutines` - Number of goroutines
- `go_gc_*` - Garbage collection metrics

### Process Metrics
- `process_cpu_seconds_total` - CPU usage
- `process_open_fds` - Open file descriptors
- `process_max_fds` - Maximum file descriptors
- `process_virtual_memory_bytes` - Virtual memory usage
- `process_resident_memory_bytes` - Resident memory usage

## HTTP Integration

### Using with HTTP Router

```go
// Get HTTP handler for metrics endpoint
handler := metricsService.GetHTTPHandler()

// Mount in HTTP router
router.Handle("GET", "/metrics", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    handler.ServeHTTP(w, r)
})
```

### Automatic Middleware Integration

The metrics service can be automatically integrated with HTTP middleware to collect request metrics:

```yaml
middleware:
  - name: "metrics"
    enabled: true
    config:
      collect_request_metrics: true
      collect_response_metrics: true
```

## Custom Metrics Registration

```go
// Register custom Prometheus collectors
customCollector := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "custom_operations_total",
        Help: "Total number of custom operations",
    },
    []string{"operation_type"},
)

err := metricsService.RegisterCustomMetrics(customCollector)
```

## Performance Considerations

1. **Label Cardinality**: Keep the number of unique label combinations low to avoid memory issues
2. **Metric Collection**: Disabled metrics have minimal overhead
3. **Histogram Buckets**: Choose buckets appropriate for your value distribution
4. **Collection Interval**: Balance between granularity and performance

## Best Practices

1. **Naming**: Use descriptive metric names following Prometheus conventions
2. **Labels**: Use labels for dimensions, avoid high-cardinality labels
3. **Units**: Include units in metric names (e.g., `_seconds`, `_bytes`)
4. **Help Text**: Provide clear help text for custom metrics
5. **Monitoring**: Monitor the metrics system itself for performance

## Integration with Lokstra Framework

The metrics service integrates seamlessly with other Lokstra services:

- **HTTP Router**: Automatic request/response metrics
- **Database Pool**: Connection pool metrics
- **Cache Service**: Hit/miss ratio metrics
- **Logger**: Error rate metrics
- **Service Registry**: Service health metrics

## Prometheus Integration

The service is fully compatible with Prometheus:

- Standard Prometheus exposition format
- Compatible with Prometheus server scraping
- Works with Grafana dashboards
- Supports Prometheus alerting rules

Example Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'lokstra-app'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```
