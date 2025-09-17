# Metrics Service Example

This example demonstrates comprehensive usage of Lokstra's built-in metrics service for application monitoring, performance tracking, and business intelligence using Prometheus metrics.

## What You'll Learn

- **Metrics Service Configuration**: Setting up Prometheus-based metrics collection
- **Counter Metrics**: Tracking events that only increase (requests, orders, errors)
- **Histogram Metrics**: Measuring distributions (response times, sizes, durations)
- **Gauge Metrics**: Monitoring current state values (active users, system resources)
- **Labels and Dimensions**: Adding context for filtering and aggregation
- **Business Metrics**: Tracking KPIs and business-critical measurements
- **Performance Monitoring**: System health and application performance metrics

## Key Features Demonstrated

### 1. Metrics Service Setup
```go
metricsConfig := map[string]interface{}{
    "enabled":                 true,
    "endpoint":               "/metrics",
    "namespace":              "lokstra_demo",
    "subsystem":              "api",
    "collect_interval":       "10s",
    "include_go_metrics":     true,
    "include_process_metrics": true,
    "buckets":                []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
}
```

### 2. Counter Metrics (Event Tracking)
```go
// Track API requests
metrics.IncCounter("api_requests_total", serviceapi.Labels{
    "endpoint": "/",
    "method":   "GET",
    "handler":  "home",
})

// Track business events
metrics.IncCounter("orders_processed_total", serviceapi.Labels{
    "order_type": orderType,
    "region":     region,
})
```

### 3. Histogram Metrics (Duration/Size Measurements)
```go
// Measure response times
duration := time.Since(start).Seconds()
metrics.ObserveHistogram("api_request_duration_seconds", duration, serviceapi.Labels{
    "endpoint": "/",
    "method":   "GET",
})

// Track business values
metrics.ObserveHistogram("order_value_dollars", amount, serviceapi.Labels{
    "order_type": orderType,
    "region":     region,
})
```

### 4. Gauge Metrics (Current State)
```go
// Track current system state
metrics.SetGauge("active_users_current", float64(activeUsers), serviceapi.Labels{
    "user_type": userType,
})

// Monitor system resources
metrics.SetGauge("system_cpu_usage_percent", cpuUsage, nil)
metrics.SetGauge("system_memory_usage_percent", memoryUsage, nil)
```

## Available Endpoints

- `GET /` - Home page with metrics overview
- `GET /metrics` - Prometheus metrics endpoint
- `GET /api/users` - User operations with detailed metrics
- `GET /api/orders` - Order processing with business metrics
- `GET /api/analytics` - Analytics queries with performance metrics
- `GET /health` - Health check with system metrics

## Testing the Example

### 1. Start the Server
```bash
cd cmd/examples/06_builtin/02_metrics_service
go run main.go
```

### 2. View Prometheus Metrics
```bash
# View all metrics
curl http://localhost:8080/metrics

# Filter specific metrics
curl http://localhost:8080/metrics | grep lokstra_demo
```

### 3. Generate Metrics Data
```bash
# Basic requests
curl http://localhost:8080/
curl http://localhost:8080/health

# User operations
curl "http://localhost:8080/api/users?operation=create&type=premium"
curl "http://localhost:8080/api/users?operation=view&type=regular"
curl "http://localhost:8080/api/users?operation=update&type=admin"

# Order processing
curl "http://localhost:8080/api/orders?type=subscription&region=eu-west"
curl "http://localhost:8080/api/orders?type=product&region=us-east"
curl "http://localhost:8080/api/orders?type=service&region=asia-pacific"

# Analytics queries
curl "http://localhost:8080/api/analytics?query=revenue&source=database"
curl "http://localhost:8080/api/analytics?query=dashboard&source=cache"
```

### 4. Load Testing for Metrics
```bash
# Generate API request metrics
for i in {1..20}; do curl "http://localhost:8080/api/users?operation=view"; done

# Generate order metrics
for i in {1..10}; do curl "http://localhost:8080/api/orders?type=product&region=us-east"; done

# Generate analytics metrics
for i in {1..5}; do curl "http://localhost:8080/api/analytics?query=dashboard"; done
```

### 5. Monitor Metrics in Real-time
```bash
# Watch metrics changes
watch -n 2 'curl -s http://localhost:8080/metrics | grep lokstra_demo'

# Monitor specific metrics
curl -s http://localhost:8080/metrics | grep -E "(requests_total|duration_seconds|active_users)"
```

## Metrics Categories

### Application Metrics
- `api_requests_total` - HTTP request counter with endpoint/method labels
- `api_request_duration_seconds` - Request response time histogram
- `user_operations_total` - User operation counter with operation/type labels
- `user_operation_duration_seconds` - User operation duration histogram

### Business Metrics
- `orders_processed_total` - Order processing counter with type/region labels
- `order_value_dollars` - Order value histogram for revenue tracking
- `business_revenue_current` - Current revenue gauge
- `customer_satisfaction_score` - Customer satisfaction gauge
- `transaction_volume_current` - Current transaction volume gauge

### System Metrics
- `active_users_current` - Currently active users gauge
- `orders_processing_current` - Orders currently being processed
- `system_cpu_usage_percent` - CPU usage percentage
- `system_memory_usage_percent` - Memory usage percentage
- `system_disk_usage_percent` - Disk usage percentage
- `database_connections_active` - Active database connections
- `cache_hit_ratio_total` - Cache hit ratio percentage

### Performance Metrics
- `analytics_queries_total` - Analytics query counter
- `analytics_computation_seconds` - Analytics computation time histogram
- `analytics_cache_hit_ratio` - Analytics cache hit ratio gauge
- `application_error_rate` - Application error rate gauge
- `message_queue_depth` - Message queue depth gauge

## Prometheus Configuration

### Sample prometheus.yml
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'lokstra-demo'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s
```

### Grafana Dashboard Queries
```promql
# Request rate
rate(lokstra_demo_api_requests_total[5m])

# Average response time
rate(lokstra_demo_api_request_duration_seconds_sum[5m]) / 
rate(lokstra_demo_api_request_duration_seconds_count[5m])

# 95th percentile response time
histogram_quantile(0.95, rate(lokstra_demo_api_request_duration_seconds_bucket[5m]))

# Current active users
lokstra_demo_active_users_current

# Order processing rate
rate(lokstra_demo_orders_processed_total[5m])

# System resource usage
lokstra_demo_system_cpu_usage_percent
lokstra_demo_system_memory_usage_percent
```

## Key Concepts

### Metric Types
1. **Counters**: Monotonically increasing values (requests, errors, events)
2. **Histograms**: Distribution of values with buckets (latencies, sizes)
3. **Gauges**: Current state values that can go up/down (active connections, CPU usage)

### Labels and Cardinality
- Use labels to add dimensions for filtering and aggregation
- Avoid high-cardinality labels (user IDs, timestamps)
- Keep label values consistent and meaningful
- Monitor total metric cardinality for performance

### Best Practices
- Use descriptive metric names with units
- Include application/service prefix (namespace)
- Pre-define buckets for histograms based on expected values
- Track both technical and business metrics
- Monitor the monitoring system itself

This example provides a comprehensive foundation for implementing production-ready monitoring in Lokstra applications using industry-standard Prometheus metrics.