# 02-Monolith Multi Port

Monolithic deployment with logical separation using multiple ports.

## Configuration Highlights

```yaml
servers:
  - name: monolith-multi-server
    deployment-id: monolith-multi-port
    apps:
      - addr: ":8080"  # Public APIs
        routers: [product-api, health-api]
      
      - addr: ":8081"  # Internal APIs
        routers: [order-api, user-api]
```

## Key Points

- **1 Process** - Single binary
- **2+ Ports** - Different APIs on different ports
- **Logical Separation** - Public vs Internal
- **Still Simple** - One deployment

## Benefits

✅ Separate public/internal APIs
✅ Different middleware per app
✅ Can scale apps independently (multiple instances)
✅ Still simple deployment

## Use When

- Need different security policies
- Want to isolate APIs
- Preparing for microservices
- Team 10-20 developers

## See Also

- `../04-deployment-patterns/comparison.md` - Complete comparison
