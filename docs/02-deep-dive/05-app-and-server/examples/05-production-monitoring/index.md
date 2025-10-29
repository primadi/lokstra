# Production Monitoring

Implement metrics and monitoring for production environments.

## Running

```bash
go run main.go
```

Server starts on `http://localhost:3110`

## Metrics Endpoints

### Prometheus Format (`/metrics`)
Standard Prometheus exposition format:
```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total 1234

# HELP http_errors_total Total number of HTTP errors
# TYPE http_errors_total counter
http_errors_total 5

# HELP http_request_duration_avg Average request duration
# TYPE http_request_duration_avg gauge
http_request_duration_avg 25.5
```

### JSON Format (`/stats`)
Easy-to-read JSON metrics:
```json
{
  "uptime_seconds": 3600,
  "requests_total": 1234,
  "errors_total": 5,
  "requests_active": 3,
  "avg_latency_ms": 25.5
}
```

## Metrics Tracked

### Request Metrics
- Total requests
- Active requests
- Error count
- Request latency

### System Metrics
- Uptime
- Memory usage (simulated)
- Goroutines (simulated)

### Performance Metrics
- Average latency
- Request duration
- Throughput

## Integration

### Prometheus
Scrape `/metrics` endpoint:
```yaml
scrape_configs:
  - job_name: 'lokstra-app'
    static_configs:
      - targets: ['localhost:3110']
    metrics_path: '/metrics'
```

### Grafana
Create dashboards using metrics:
- Request rate over time
- Error rate over time
- Average latency
- Active requests

### Alerting
Set up alerts based on thresholds:
- Error rate > 5%
- Average latency > 100ms
- No requests in 5 minutes

## Best Practices

- Use atomic operations for counters
- Track both successes and failures
- Include timestamps
- Export in standard formats
- Keep metrics lightweight
- Aggregate data appropriately
