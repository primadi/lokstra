# 03-Microservices

Full microservices architecture with independent services.

## Configuration Highlights

Each service has its own `config-{service}-service.yaml`:

**Product Service:**
```yaml
servers:
  - name: product-service
    baseUrl: http://product-service
    deployment-id: microservice
    apps:
      - addr: ":8080"
        routers: [product-api, health-api]
```

**Order Service:**
```yaml
servers:
  - name: order-service
    baseUrl: http://order-service
    deployment-id: microservice
    apps:
      - addr: ":8080"
        routers: [order-api, health-api]
```

**User Service:**
```yaml
servers:
  - name: user-service
    baseUrl: http://user-service
    deployment-id: microservice
    apps:
      - addr: ":8080"
        routers: [user-api, health-api]
```

## Key Points

- **3+ Processes** - One per service
- **Independent Deployment** - Deploy services separately
- **Independent Scaling** - Scale services independently
- **Network Communication** - HTTP between services

## Benefits

✅ Independent deployment
✅ Independent scaling
✅ Team autonomy
✅ Fault isolation
✅ Technology diversity

## Challenges

❌ Complex infrastructure (Kubernetes)
❌ Network latency
❌ Distributed tracing needed
❌ Higher operational overhead

## Use When

- Large team (20+ developers)
- High traffic requiring independent scaling
- Need independent deployments
- Have DevOps team

## See Also

- `../04-deployment-patterns/comparison.md` - Complete comparison
