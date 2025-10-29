# App & Server Deep Dive - Examples

Production-ready patterns for application lifecycle and server management.

## Examples

### ✅ 01 - Lifecycle Hooks
Application initialization and shutdown hooks.

**Topics**: Init hooks, startup, shutdown, cleanup  
**Port**: 3070

Demonstrates:
- Startup initialization
- Resource cleanup on shutdown
- Request tracking
- Uptime monitoring

[View Example →](./01-lifecycle-hooks/)

---

### ✅ 02 - Multiple Servers
Run multiple Lokstra servers concurrently.

**Topics**: Server grouping, port management, concurrent servers  
**Ports**: 3080 (API), 3081 (Admin), 3082 (Metrics)

Demonstrates:
- Multiple servers in one app
- Separation of concerns
- Independent service scaling
- Different ports per service

[View Example →](./02-multiple-servers/)

---

### ✅ 03 - Graceful Shutdown
Handle termination signals gracefully.

**Topics**: Request draining, cleanup, timeouts  
**Port**: 3090

Demonstrates:
- Signal handling (SIGTERM, SIGINT)
- Wait for active requests
- Shutdown timeout
- Resource cleanup

[View Example →](./03-graceful-shutdown/)

---

### ✅ 04 - Health Checks
Implement liveness and readiness probes.

**Topics**: Health endpoints, probes, dependency checks  
**Port**: 3100

Demonstrates:
- Basic health check
- Detailed health status
- Readiness probe (for load balancers)
- Liveness probe (for orchestrators)
- Dependency health tracking

[View Example →](./04-health-checks/)

---

### ✅ 05 - Production Monitoring
Metrics, monitoring, and observability.

**Topics**: Prometheus, metrics, observability  
**Port**: 3110

Demonstrates:
- Prometheus metrics format
- JSON metrics endpoint
- Request counting
- Latency tracking
- Error rate monitoring

[View Example →](./05-production-monitoring/)

---

## Running Examples

Each example is self-contained:
```
01-lifecycle-hooks/
├── main.go              # Working code
├── index.md             # Documentation
└── test.http            # HTTP tests
```

To run any example:
```bash
cd 01-lifecycle-hooks
go run main.go
```

Test with the included `test.http` file or curl:
```bash
curl http://localhost:3070/
```

---

## Quick Reference

| Example | Port(s) | Key Feature |
|---------|---------|-------------|
| Lifecycle Hooks | 3070 | Startup/shutdown hooks |
| Multiple Servers | 3080-3082 | Concurrent servers |
| Graceful Shutdown | 3090 | Clean termination |
| Health Checks | 3100 | Monitoring endpoints |
| Production Monitoring | 3110 | Metrics & observability |

**Status**: ✅ All examples complete and tested
