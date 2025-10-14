# 01-Monolith Single Port

The simplest deployment pattern - all services in one process on one port.

## Configuration Highlights

```yaml
servers:
  - name: monolith-single-server
    deployment-id: monolith-single-port
    apps:
      - addr: ":8080"
        routers: [product-api, order-api, user-api, health-api]
```

## Key Points

- **1 Process** - Single binary
- **1 Port** - All APIs on :8080
- **Zero Network Overhead** - All inter-service calls are local
- **Simplest Deployment** - One command to run

## Benefits

✅ Simplest to develop and deploy
✅ Zero network latency between services
✅ Lowest infrastructure cost
✅ Easy to debug (single process)

## Use When

- Starting a new project
- Team < 10 developers
- Budget constraints
- Simple requirements

## See Also

- `../04-deployment-patterns/comparison.md` - Complete comparison
