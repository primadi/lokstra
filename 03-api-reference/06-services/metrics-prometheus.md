# Metrics (Prometheus)

The `metrics_prometheus` service provides Prometheus-based metrics collection for monitoring application performance, usage patterns, and health.

## Table of Contents

- [Overview](#overview)
- [Configuration](#configuration)
- [Registration](#registration)
- [Metric Types](#metric-types)
- [Usage](#usage)
- [HTTP Integration](#http-integration)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

**Service Type:** `metrics_prometheus`

**Interface:** `serviceapi.Metrics`

**Key Features:**

```
✓ Counter Metrics       - Track cumulative values
✓ Histogram Metrics     - Track distributions (latency, sizes)
✓ Gauge Metrics         - Track current values
✓ Dynamic Labels        - Flexible metric dimensions
✓ Auto-Registration     - Metrics created on first use
✓ Thread-Safe           - Concurrent access safe
```

## Configuration

### Config Struct

```go
type Config struct {
    Namespace string `json:"namespace" yaml:"namespace"`  // Metric namespace prefix
    Subsystem string `json:"subsystem" yaml:"subsystem"`  // Metric subsystem prefix
}
```

**Metric Naming:**
```
{namespace}_{subsystem}_{metric_name}

Example:
myapp_api_requests_total
└───┘ └─┘ └─────────────┘
namespace subsystem name
```

### YAML Configuration

**Basic Configuration:**

```yaml
services:
  metrics:
    type: metrics_prometheus
    config:
      namespace: myapp
      subsystem: api
```

**Multiple Metric Services:**

```yaml
services:
  # API metrics
  api_metrics:
    type: metrics_prometheus
    config:
      namespace: myapp
      subsystem: api
      
  # Database metrics
  db_metrics:
    type: metrics_prometheus
    config:
      namespace: myapp
      subsystem: database
      
  # Worker metrics
  worker_metrics:
    type: metrics_prometheus
    config:
      namespace: myapp
      subsystem: worker
```

**Environment-Based Configuration:**

```yaml
services:
  metrics:
    type: metrics_prometheus
    config:
      namespace: ${APP_NAME:myapp}
      subsystem: ${SERVICE_NAME:api}
```

### Programmatic Configuration

```go
import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
    "github.com/primadi/lokstra/services/metrics_prometheus"
)

// Register service
metrics_prometheus.Register()

// Create metrics service
metrics := lokstra_registry.NewService[serviceapi.Metrics](
    "metrics", "metrics_prometheus",
    map[string]any{
        "namespace": "myapp",
        "subsystem": "api",
    },
)
```

## Registration

### Basic Registration

```go
import "github.com/primadi/lokstra/services/metrics_prometheus"

func init() {
    metrics_prometheus.Register()
}
```

### Bulk Registration

```go
import "github.com/primadi/lokstra/services"

func main() {
    // Registers all services including metrics_prometheus
    services.RegisterAllServices()
    
    // Or register only core services
    services.RegisterCoreServices()
}
```

## Metric Types

### Interface Definition

```go
type Metrics interface {
    // Increment a counter by 1
    IncCounter(name string, labels Labels)
    
    // Record a histogram observation
    ObserveHistogram(name string, value float64, labels Labels)
    
    // Set a gauge to a specific value
    SetGauge(name string, value float64, labels Labels)
}

type Labels = map[string]string
```

### Counter Metrics

**Counters** track cumulative values that only increase (never decrease).

**Use Cases:**
- Request counts
- Error counts
- Task completions
- Events processed

**Example:**

```go
// Increment counter
metrics.IncCounter("requests_total", serviceapi.Labels{
    "method": "GET",
    "path":   "/api/users",
    "status": "200",
})

// Result: myapp_api_requests_total{method="GET",path="/api/users",status="200"} 1
```

### Histogram Metrics

**Histograms** track distributions of values (automatically creates sum, count, and buckets).

**Use Cases:**
- Request duration
- Response sizes
- Processing time
- Database query time

**Example:**

```go
// Record request duration
duration := time.Since(start).Seconds()
metrics.ObserveHistogram("request_duration_seconds", duration, serviceapi.Labels{
    "method": "GET",
    "path":   "/api/users",
})

// Result: Multiple metrics created:
// myapp_api_request_duration_seconds_bucket{method="GET",path="/api/users",le="0.005"} 0
// myapp_api_request_duration_seconds_bucket{method="GET",path="/api/users",le="0.01"} 1
// myapp_api_request_duration_seconds_sum{method="GET",path="/api/users"} 0.123
// myapp_api_request_duration_seconds_count{method="GET",path="/api/users"} 1
```

### Gauge Metrics

**Gauges** track current values that can increase or decrease.

**Use Cases:**
- Active connections
- Queue length
- Memory usage
- Temperature readings

**Example:**

```go
// Set current active connections
metrics.SetGauge("active_connections", float64(connectionCount), serviceapi.Labels{
    "server": "api-1",
})

// Result: myapp_api_active_connections{server="api-1"} 42
```

## Usage

### Basic Usage

```go
import (
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

// Get metrics service
metrics := lokstra_registry.GetService[serviceapi.Metrics]("metrics")

// Track request
metrics.IncCounter("requests_total", serviceapi.Labels{
    "endpoint": "/users",
    "method":   "GET",
})

// Track duration
start := time.Now()
// ... do work ...
duration := time.Since(start).Seconds()
metrics.ObserveHistogram("request_duration_seconds", duration, serviceapi.Labels{
    "endpoint": "/users",
})

// Track active goroutines
metrics.SetGauge("goroutines", float64(runtime.NumGoroutine()), serviceapi.Labels{})
```

### Labels

Labels add dimensions to metrics for filtering and aggregation:

```go
// Good: Meaningful labels
metrics.IncCounter("requests_total", serviceapi.Labels{
    "method":   "GET",           // HTTP method
    "path":     "/api/users",    // Request path
    "status":   "200",           // Response status
    "tenant":   "acme-corp",     // Multi-tenant ID
})

// Bad: High cardinality labels
metrics.IncCounter("requests_total", serviceapi.Labels{
    "user_id": "user-12345",     // BAD: Too many unique values
    "timestamp": time.Now().String(),  // BAD: Infinite cardinality
})
```

**Label Best Practices:**
- Use low cardinality (< 100 unique values per label)
- Avoid user IDs, timestamps, or random values
- Use consistent label names across metrics

### Request Tracking

```go
func trackRequest(next http.Handler) http.Handler {
    metrics := lokstra_registry.GetService[serviceapi.Metrics]("metrics")
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Track request
        metrics.IncCounter("requests_total", serviceapi.Labels{
            "method": r.Method,
            "path":   r.URL.Path,
        })
        
        // Process request
        next.ServeHTTP(w, r)
        
        // Track duration
        duration := time.Since(start).Seconds()
        metrics.ObserveHistogram("request_duration_seconds", duration, serviceapi.Labels{
            "method": r.Method,
            "path":   r.URL.Path,
        })
    })
}
```

### Error Tracking

```go
func trackErrors(err error, operation string) {
    metrics := lokstra_registry.GetService[serviceapi.Metrics]("metrics")
    
    if err != nil {
        metrics.IncCounter("errors_total", serviceapi.Labels{
            "operation": operation,
            "type":      errorType(err),
        })
    }
}

func errorType(err error) string {
    switch {
    case errors.Is(err, context.Canceled):
        return "canceled"
    case errors.Is(err, context.DeadlineExceeded):
        return "timeout"
    default:
        return "internal"
    }
}
```

## HTTP Integration

### Exposing Metrics Endpoint

```go
import (
    "net/http"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupMetricsEndpoint() {
    // Get metrics service (type assertion to access Registry())
    metricsService := lokstra_registry.GetService[any]("metrics")
    
    // Type assert to get Prometheus registry
    promService := metricsService.(*metrics_prometheus.metricsPrometheus)
    registry := promService.Registry()
    
    // Create HTTP handler
    handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
    
    // Register endpoint
    http.Handle("/metrics", handler)
    
    log.Println("Metrics available at http://localhost:8080/metrics")
}
```

### Metrics Response Format

```
# HELP myapp_api_requests_total Total number of HTTP requests
# TYPE myapp_api_requests_total counter
myapp_api_requests_total{method="GET",path="/api/users",status="200"} 1523

# HELP myapp_api_request_duration_seconds HTTP request duration in seconds
# TYPE myapp_api_request_duration_seconds histogram
myapp_api_request_duration_seconds_bucket{method="GET",path="/api/users",le="0.005"} 145
myapp_api_request_duration_seconds_bucket{method="GET",path="/api/users",le="0.01"} 456
myapp_api_request_duration_seconds_bucket{method="GET",path="/api/users",le="0.025"} 1234
myapp_api_request_duration_seconds_sum{method="GET",path="/api/users"} 12.34
myapp_api_request_duration_seconds_count{method="GET",path="/api/users"} 1523

# HELP myapp_api_active_connections Current number of active connections
# TYPE myapp_api_active_connections gauge
myapp_api_active_connections{server="api-1"} 42
```

## Best Practices

### Metric Naming

```go
✓ DO: Use descriptive names with units
"request_duration_seconds"
"response_size_bytes"
"queue_length_total"

✗ DON'T: Use unclear names
"request_time"      // What unit?
"size"              // Size of what?

✓ DO: Use consistent suffixes
"_total"    for counters
"_seconds"  for durations
"_bytes"    for sizes

✗ DON'T: Mix naming conventions
"requests_count"
"total_errors"
```

### Label Usage

```go
✓ DO: Use low cardinality labels
labels := serviceapi.Labels{
    "method":   "GET",        // ~10 values
    "status":   "200",        // ~20 values
    "endpoint": "/api/users", // ~50 values
}

✗ DON'T: Use high cardinality labels
labels := serviceapi.Labels{
    "user_id":   userID,      // BAD: Thousands of users
    "request_id": requestID,  // BAD: Every request unique
    "timestamp":  timestamp,  // BAD: Infinite values
}

✓ DO: Normalize path labels
path := normalizePath(r.URL.Path)  // "/api/users/123" -> "/api/users/:id"
labels := serviceapi.Labels{"path": path}

✗ DON'T: Use raw dynamic paths
labels := serviceapi.Labels{"path": r.URL.Path}  // BAD: /api/users/1, /api/users/2, ...
```

### Metric Selection

```go
✓ DO: Choose appropriate metric types

// Counter - cumulative values
metrics.IncCounter("requests_total", labels)
metrics.IncCounter("errors_total", labels)

// Histogram - distributions
metrics.ObserveHistogram("request_duration_seconds", duration, labels)
metrics.ObserveHistogram("response_size_bytes", float64(size), labels)

// Gauge - current state
metrics.SetGauge("active_connections", float64(count), labels)
metrics.SetGauge("queue_length", float64(len(queue)), labels)

✗ DON'T: Misuse metric types
// BAD: Using gauge for cumulative count
metrics.SetGauge("request_count", float64(count), labels)

// BAD: Using counter for current state
metrics.IncCounter("active_connections", labels)
```

### Performance

```go
✓ DO: Reuse label maps when possible
var labels = serviceapi.Labels{
    "service": "api",
    "version": "1.0",
}
metrics.IncCounter("requests_total", labels)

✗ DON'T: Create labels unnecessarily
for i := 0; i < 1000; i++ {
    metrics.IncCounter("requests", serviceapi.Labels{  // Creates map 1000 times
        "endpoint": "/api/users",
    })
}

✓ DO: Aggregate before recording
totalDuration := 0.0
for _, d := range durations {
    totalDuration += d
}
metrics.ObserveHistogram("batch_duration", totalDuration, labels)

✗ DON'T: Record every single value unnecessarily
for _, d := range durations {
    metrics.ObserveHistogram("duration", d, labels)  // May be too granular
}
```

## Examples

### Complete Request Tracking Middleware

```go
package middleware

import (
    "net/http"
    "strconv"
    "time"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

type responseWriter struct {
    http.ResponseWriter
    status int
    size   int
}

func (rw *responseWriter) WriteHeader(status int) {
    rw.status = status
    rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
    size, err := rw.ResponseWriter.Write(b)
    rw.size += size
    return size, err
}

func RequestMetrics(next http.Handler) http.Handler {
    metrics := lokstra_registry.GetService[serviceapi.Metrics]("metrics")
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap response writer
        rw := &responseWriter{ResponseWriter: w, status: 200}
        
        // Process request
        next.ServeHTTP(rw, r)
        
        // Calculate duration
        duration := time.Since(start).Seconds()
        
        // Normalize path (replace IDs with placeholders)
        path := normalizePath(r.URL.Path)
        
        // Track request count
        metrics.IncCounter("requests_total", serviceapi.Labels{
            "method": r.Method,
            "path":   path,
            "status": strconv.Itoa(rw.status),
        })
        
        // Track request duration
        metrics.ObserveHistogram("request_duration_seconds", duration, serviceapi.Labels{
            "method": r.Method,
            "path":   path,
        })
        
        // Track response size
        metrics.ObserveHistogram("response_size_bytes", float64(rw.size), serviceapi.Labels{
            "method": r.Method,
            "path":   path,
        })
        
        // Track active requests (gauge)
        // Note: This would need a counter incremented at start and decremented here
    })
}

func normalizePath(path string) string {
    // Replace numeric IDs with :id placeholder
    // /api/users/123 -> /api/users/:id
    // Implementation depends on your routing strategy
    return path
}
```

### Database Query Metrics

```go
package repository

import (
    "context"
    "time"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

type MetricsRepository struct {
    db      serviceapi.DbPool
    metrics serviceapi.Metrics
}

func NewMetricsRepository() *MetricsRepository {
    return &MetricsRepository{
        db:      lokstra_registry.GetService[serviceapi.DbPool]("main_db"),
        metrics: lokstra_registry.GetService[serviceapi.Metrics]("db_metrics"),
    }
}

func (r *MetricsRepository) Query(ctx context.Context, query string, args ...any) ([]any, error) {
    start := time.Now()
    
    // Execute query
    conn, err := r.db.Acquire(ctx, "public")
    if err != nil {
        r.trackError("acquire", err)
        return nil, err
    }
    defer conn.Release()
    
    rows, err := conn.SelectManyRowMap(ctx, query, args...)
    duration := time.Since(start).Seconds()
    
    // Track query
    labels := serviceapi.Labels{
        "operation": "select",
    }
    
    r.metrics.IncCounter("queries_total", labels)
    r.metrics.ObserveHistogram("query_duration_seconds", duration, labels)
    
    if err != nil {
        r.trackError("query", err)
        return nil, err
    }
    
    return rows, nil
}

func (r *MetricsRepository) trackError(operation string, err error) {
    r.metrics.IncCounter("errors_total", serviceapi.Labels{
        "operation": operation,
        "type":      getErrorType(err),
    })
}

func getErrorType(err error) string {
    // Classify error types
    if err == nil {
        return "none"
    }
    // Add error classification logic
    return "unknown"
}
```

### Background Worker Metrics

```go
package worker

import (
    "context"
    "time"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

type Worker struct {
    metrics serviceapi.Metrics
    queue   chan Task
}

func NewWorker() *Worker {
    return &Worker{
        metrics: lokstra_registry.GetService[serviceapi.Metrics]("worker_metrics"),
        queue:   make(chan Task, 100),
    }
}

func (w *Worker) Start(ctx context.Context) {
    // Track active workers
    w.metrics.SetGauge("active_workers", 1, serviceapi.Labels{
        "worker_id": "worker-1",
    })
    defer w.metrics.SetGauge("active_workers", 0, serviceapi.Labels{
        "worker_id": "worker-1",
    })
    
    for {
        select {
        case <-ctx.Done():
            return
        case task := <-w.queue:
            w.processTask(task)
        }
    }
}

func (w *Worker) processTask(task Task) {
    start := time.Now()
    
    // Track queue length
    w.metrics.SetGauge("queue_length", float64(len(w.queue)), serviceapi.Labels{})
    
    // Process task
    err := task.Execute()
    duration := time.Since(start).Seconds()
    
    // Track task completion
    labels := serviceapi.Labels{
        "task_type": task.Type(),
    }
    
    w.metrics.IncCounter("tasks_processed_total", labels)
    w.metrics.ObserveHistogram("task_duration_seconds", duration, labels)
    
    if err != nil {
        w.metrics.IncCounter("task_errors_total", labels)
    }
}
```

### System Metrics Collector

```go
package metrics

import (
    "runtime"
    "time"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi"
)

type SystemMetrics struct {
    metrics serviceapi.Metrics
    ticker  *time.Ticker
}

func NewSystemMetrics() *SystemMetrics {
    return &SystemMetrics{
        metrics: lokstra_registry.GetService[serviceapi.Metrics]("metrics"),
        ticker:  time.NewTicker(10 * time.Second),
    }
}

func (s *SystemMetrics) Start() {
    go func() {
        for range s.ticker.C {
            s.collect()
        }
    }()
}

func (s *SystemMetrics) Stop() {
    s.ticker.Stop()
}

func (s *SystemMetrics) collect() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    labels := serviceapi.Labels{}
    
    // Goroutines
    s.metrics.SetGauge("goroutines", float64(runtime.NumGoroutine()), labels)
    
    // Memory
    s.metrics.SetGauge("memory_alloc_bytes", float64(m.Alloc), labels)
    s.metrics.SetGauge("memory_sys_bytes", float64(m.Sys), labels)
    s.metrics.SetGauge("memory_heap_inuse_bytes", float64(m.HeapInuse), labels)
    
    // GC
    s.metrics.SetGauge("gc_pause_seconds", float64(m.PauseNs[(m.NumGC+255)%256])/1e9, labels)
    s.metrics.IncCounter("gc_runs_total", labels)
}
```

## Related Documentation

- [Services Overview](README.md) - Service architecture and patterns
- [Request Logger Middleware](../05-middleware/request-logger.md) - Request logging
- [Slow Request Logger](../05-middleware/slow-request-logger.md) - Performance monitoring
- [Configuration](../03-configuration/config.md) - YAML configuration

---

**Next:** [Redis Service](redis.md) - Direct Redis client access
