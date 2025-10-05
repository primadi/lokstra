# 02-app: Combining Routers into an App

## What You'll Learn
- Create multiple independent routers
- Combine routers into a single App
- Run an App with graceful shutdown
- How routers are chained together

## Key Concepts

### App
An **App** combines multiple routers and runs them on the same address:
- Takes multiple routers as input
- Chains routers together internally
- Handles graceful shutdown
- Coordinates request handling across all routers

### Router Chaining
When multiple routers are passed to an App:
- Requests are tried against each router in order
- First matching route handles the request
- Routers are independent - each defines its own full paths

## Running the Example

```bash
cd cmd/learning/01-basics/02-app
go run main.go
```

## Testing

```bash
# Users endpoints
curl http://localhost:8080/users
curl http://localhost:8080/users/123
curl -X POST http://localhost:8080/users

# Products endpoints
curl http://localhost:8080/products
curl http://localhost:8080/products/456

# Admin endpoints
curl http://localhost:8080/admin/stats
curl http://localhost:8080/admin/health
```

## Key Differences from 01-router

| Concept | 01-router | 02-app |
|---------|-----------|--------|
| **Single Router** | ✅ One router with all routes | ❌ Multiple routers |
| **Multiple Routers** | ❌ N/A | ✅ Separate routers per domain |
| **Start Method** | `http.ListenAndServe` | `app.Run()` with graceful shutdown |
| **Organization** | All routes in one place | Routes grouped by domain |

## When to Use App vs Router

**Use Router directly** when:
- Simple applications with few routes
- Quick prototypes or demos
- Single domain/responsibility

**Use App with multiple Routers** when:
- Larger applications with multiple domains
- Want to organize routes by feature/module
- Need graceful shutdown support
- May want to deploy routers separately later

## What's Next?
- **03-server**: Learn about Servers that can run multiple Apps
- **05-config**: Learn how to configure Apps via YAML
