# 04-Scalable Deployment

Learn how to deploy your Lokstra application in various scalability patterns - from monolith to microservices - **with the same codebase**.

## What You'll Learn

1. **Deployment Patterns**
   - Monolith Single Port (simplest)
   - Monolith Multi Port (isolation)
   - Microservices (distributed)
   - Hybrid (private + public)

2. **Router Integration**
   - Local router calls (zero network overhead)
   - Remote router calls (HTTP)
   - Automatic discovery based on deployment

3. **Service Communication**
   - Cross-router communication
   - Same code for all deployment patterns
   - No code changes needed

## Quick Start

### Scenario 1: Monolith Single Port (Development)

```bash
# All services on one port - simplest deployment
go run . monolith-single

# Server runs on :8080
# All router calls are local (zero overhead)
```

### Scenario 2: Monolith Multi Port (Staging)

```bash
# Services on different ports - better isolation
go run . monolith-multi

# Product API: :8081
# Order API: :8082
# Router calls use HTTP to localhost
```

### Scenario 3: Microservices (Production)

```bash
# Terminal 1 - Product Service
go run . product-service

# Terminal 2 - Order Service  
go run . order-service

# Terminal 3 - API Gateway (optional)
go run . gateway
```

### Scenario 4: Hybrid Deployment (Public + Private)

```bash
# Public API on :8080 (exposed to internet)
# Internal APIs on :8081, :8082 (private network)
go run . hybrid
```

## Architecture

```
┌─────────────────────────────────────────┐
│           Code (Same for All)           │
│                                         │
│  ProductService → OrderService          │
│       ↓               ↓                 │
│  product-api      order-api             │
│                                         │
└────────────┬────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│      Deployment Configuration           │
│         (Different YAML)                │
│                                         │
│  • monolith-single.yaml                 │
│  • monolith-multi.yaml                  │
│  • microservices.yaml                   │
│  • hybrid.yaml                          │
└─────────────────────────────────────────┘
```

## How Router Integration Works

### 1. Register Routers (Code)

```go
// Product Router
productRouter := lokstra.NewRouter("product-api")
productRouter.GET("/products", handlers.GetProducts)
productRouter.GET("/products/{id}", handlers.GetProduct)
old_registry.RegisterRouter("product-api", productRouter)

// Order Router
orderRouter := lokstra.NewRouter("order-api")
orderRouter.POST("/orders", handlers.CreateOrder)
orderRouter.GET("/orders/{id}", handlers.GetOrder)
old_registry.RegisterRouter("order-api", orderRouter)
```

### 2. Cross-Router Communication

```go
// In CreateOrder handler (order-api)
func CreateOrder(c *lokstra.RequestContext) error {
    // Get product info from product-api
    // This automatically uses:
    // - Local call (httptest) if same server
    // - HTTP call if different server
    productClient := services.GetProductClient()
    resp, err := productClient.GET("/products/" + productID)
    
    // Business logic...
    return c.Api.Created(order)
}
```

### 3. Deployment Config (YAML)

**Monolith Single Port:**
```yaml
servers:
  - name: monolith-server
    deployment-id: dev
    apps:
      - addr: ":8080"
        routers: [product-api, order-api, health-api]
```

**Microservices:**
```yaml
servers:
  - name: product-service
    baseUrl: http://localhost
    deployment-id: prod
    apps:
      - addr: ":8081"
        routers: [product-api, health-api]
        
  - name: order-service
    baseUrl: http://localhost
    deployment-id: prod
    apps:
      - addr: ":8082"
        routers: [order-api, health-api]
```

## Deployment Patterns Explained

### Pattern 1: Monolith Single Port

**Use Case:** Development, small projects

**Benefits:**
- ✅ Simplest to develop and test
- ✅ Zero network overhead (all local)
- ✅ Single process, easy debugging
- ✅ No port management needed

**Drawbacks:**
- ❌ All services share resources
- ❌ Can't scale individual services
- ❌ Restart affects everything

**Config:** `monolith-single.yaml`

### Pattern 2: Monolith Multi Port

**Use Case:** Staging, medium projects

**Benefits:**
- ✅ Better resource isolation per service
- ✅ Can monitor each service separately
- ✅ Easier load testing individual services
- ✅ Still runs on same server

**Drawbacks:**
- ❌ Network calls between services (localhost)
- ❌ Port management needed
- ❌ Can't scale across machines

**Config:** `monolith-multi.yaml`

### Pattern 3: Microservices

**Use Case:** Production, large projects

**Benefits:**
- ✅ Independent scaling per service
- ✅ Independent deployment
- ✅ Technology diversity possible
- ✅ Fault isolation

**Drawbacks:**
- ❌ Network latency between services
- ❌ Distributed system complexity
- ❌ Requires orchestration (k8s, docker-compose)

**Config:** `microservices.yaml` + separate configs per service

### Pattern 4: Hybrid (Public + Private)

**Use Case:** Security-sensitive production

**Benefits:**
- ✅ Public APIs isolated from internal
- ✅ Private services not exposed
- ✅ Better security posture
- ✅ Can scale public/private independently

**Drawbacks:**
- ❌ Network configuration complexity
- ❌ Requires firewall rules
- ❌ More deployment complexity

**Config:** `hybrid.yaml`

## Testing

All patterns use the same test file!

```bash
# Start server (any pattern)
go run . monolith-single

# In another terminal, use test.http
```

**test.http:**
```http
### Get Products
GET http://localhost:8080/api/products

### Create Order (triggers cross-router call to product-api)
POST http://localhost:8080/api/orders
Content-Type: application/json

{
  "user_id": "user123",
  "product_ids": ["1", "2"],
  "total_amount": 299.98
}
```

## Performance Comparison

| Pattern | Local Call Overhead | Network Call Overhead | Scalability |
|---------|---------------------|----------------------|-------------|
| Monolith Single | 0 ns (direct) | N/A | Low |
| Monolith Multi | ~50 µs (localhost) | ~100 KB/s | Medium |
| Microservices | ~1-5 ms (network) | ~10-100 KB/s | High |
| Hybrid | Mixed | Mixed | High |

## Migration Path

```
Development → Staging → Production

Step 1: Develop with monolith-single.yaml
  ↓
Step 2: Test with monolith-multi.yaml
  ↓
Step 3: Deploy with microservices.yaml
  ↓
Step 4: Optimize with hybrid.yaml
```

**No code changes needed at any step!**

## Key Takeaways

1. **Same Code, Different Deployments**
   - Write business logic once
   - Deploy in any pattern
   - Configuration-driven scaling

2. **Router Integration is Key**
   - Framework doesn't know about your domains
   - You define router names and URLs
   - Automatic local/remote switching

3. **Start Simple, Scale Later**
   - Begin with monolith single port
   - Add ports when needed
   - Split to microservices when required
   - No refactoring needed

4. **Performance Awareness**
   - Local calls are free
   - Localhost calls are cheap
   - Network calls have cost
   - Design accordingly

## Next Steps

- **Optimize:** Add caching between services
- **Monitor:** Add metrics to router calls
- **Resilience:** Add retry/circuit breaker patterns
- **Security:** Add authentication between services

## Comparison with 03-service-dependencies

| Feature | 03-service-dependencies | 04-scalable-deployment |
|---------|------------------------|------------------------|
| Focus | Service lazy loading | Deployment patterns |
| Scope | Single server | Multi-server |
| Communication | Direct service calls | Router integration |
| Scalability | Limited | Full |
| Use Case | Learning internals | Production deployment |
