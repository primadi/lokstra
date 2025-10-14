# Deployment Patterns - Complete Comparison

This guide compares all three deployment strategies with the **same application** code.

## The Same Application, Three Ways

```
E-Commerce API
├── Product Service (product-api)
├── Order Service (order-api)
├── User Service (user-api)
└── Health Check (health-api)
```

**Key Point:** The application code stays the same. Only `config.yaml` changes!

## Comparison Table

| Feature | Monolith Single | Monolith Multi | Microservices |
|---------|----------------|----------------|---------------|
| **Processes** | 1 | 1 | 3+ (one per service) |
| **Ports** | 1 (:8080) | 2+ (:8080, :8081) | 1 per service |
| **Config Files** | 1 (config.yaml) | 1 (config.yaml) | 1 per service |
| **Deployment** | Single binary | Single binary | Multiple binaries |
| **Communication** | Local (zero network) | Local + HTTP localhost | HTTP over network |
| **Complexity** | ⭐ Simple | ⭐⭐ Moderate | ⭐⭐⭐ Complex |
| **Scalability** | Scale entire app | Scale entire app | Scale services independently |
| **Cost** | 💰 Lowest | 💰💰 Low | 💰💰💰 Higher |
| **Team Size** | 1-10 devs | 5-20 devs | 20+ devs |
| **Best For** | Startups, MVPs | Growing apps | Large enterprises |

## Pattern 1: Monolith Single Port

### Configuration

```yaml
servers:
  - name: monolith-single-server
    baseUrl: http://localhost
    deployment-id: monolith-single-port
    apps:
      - addr: ":8080"
        routers: [product-api, order-api, user-api, health-api]
```

### Architecture

```
┌─────────────────────────────────────────┐
│  Single Process (one binary)           │
│  ┌─────────────────────────────────┐   │
│  │  App (:8080)                     │   │
│  │  ├─ product-api                  │   │
│  │  ├─ order-api                    │   │
│  │  ├─ user-api                     │   │
│  │  └─ health-api                   │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

### Communication

```
order-api → product-api:  Local (httptest)  🚀 Zero network overhead
order-api → user-api:     Local (httptest)  🚀 Zero network overhead
```

### Pros & Cons

**✅ Pros:**
- Simplest deployment
- Zero network latency
- Easy local development
- Single process to monitor
- Lowest infrastructure cost

**❌ Cons:**
- All services scale together
- One failure affects all
- Single performance bottleneck
- Must redeploy everything for any change

**🎯 Use When:**
- Starting a new project
- Team < 10 developers
- Traffic < 1000 req/sec
- Budget constraints

### Deployment

```bash
# Single command
./ecommerce-app

# Or with Docker
docker run -p 8080:8080 ecommerce-app
```

## Pattern 2: Monolith Multi Port

### Configuration

```yaml
servers:
  - name: monolith-multi-server
    baseUrl: http://localhost
    deployment-id: monolith-multi-port
    apps:
      # Public API
      - addr: ":8080"
        routers: [product-api, health-api]
      
      # Internal API
      - addr: ":8081"
        routers: [order-api, user-api]
```

### Architecture

```
┌─────────────────────────────────────────┐
│  Single Process (one binary)           │
│  ┌─────────────────────────────────┐   │
│  │  App 1 (:8080) - Public         │   │
│  │  ├─ product-api                  │   │
│  │  └─ health-api                   │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │  App 2 (:8081) - Internal       │   │
│  │  ├─ order-api                    │   │
│  │  └─ user-api                     │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

### Communication

```
order-api → product-api:  HTTP localhost:8080  📡 Minimal overhead
order-api → user-api:     Local (same app)     🚀 Zero overhead
```

### Pros & Cons

**✅ Pros:**
- Logical API separation (public/internal)
- Different middleware per app
- Can expose only some APIs externally
- Still single deployment
- Can run multiple instances per app

**❌ Cons:**
- Slightly more complex than single-port
- Small network overhead between apps
- Still monolithic (all services in one binary)

**🎯 Use When:**
- Need different security policies
- Want to isolate public/internal APIs
- Preparing for microservices
- Team 10-20 developers
- Traffic 1000-10000 req/sec

### Deployment

```bash
# Single binary, multiple ports
./ecommerce-app

# Or with Docker - expose both ports
docker run -p 8080:8080 -p 8081:8081 ecommerce-app

# Or scale individual apps (run multiple processes)
./ecommerce-app &  # Process 1
./ecommerce-app &  # Process 2 (load balancer distributes)
```

## Pattern 3: Microservices

### Configuration (3 separate files)

**config-product-service.yaml:**
```yaml
servers:
  - name: product-service
    baseUrl: http://product-service
    deployment-id: microservice
    apps:
      - addr: ":8080"
        routers: [product-api, health-api]
```

**config-order-service.yaml:**
```yaml
servers:
  - name: order-service
    baseUrl: http://order-service
    deployment-id: microservice
    apps:
      - addr: ":8080"
        routers: [order-api, health-api]
```

**config-user-service.yaml:**
```yaml
servers:
  - name: user-service
    baseUrl: http://user-service
    deployment-id: microservice
    apps:
      - addr: ":8080"
        routers: [user-api, health-api]
```

### Architecture

```
┌─────────────────────┐   ┌─────────────────────┐   ┌─────────────────────┐
│ Product Service     │   │ Order Service       │   │ User Service        │
│ (product-service    │   │ (order-service      │   │ (user-service       │
│  :8080)             │   │  :8080)             │   │  :8080)             │
│ ├─ product-api      │   │ ├─ order-api        │   │ ├─ user-api         │
│ └─ health-api       │   │ └─ health-api       │   │ └─ health-api       │
│ ├─ DB: products     │   │ ├─ DB: orders       │   │ ├─ DB: users        │
│ └─ Cache            │   │ └─ Cache            │   │ └─ Cache            │
└─────────────────────┘   └─────────────────────┘   └─────────────────────┘
         ▲                         ▲                          ▲
         └─────────────────────────┴──────────────────────────┘
                         HTTP Network Calls
```

### Communication

```
order-api → product-api:  HTTP http://product-service:8080  📡 Network call
order-api → user-api:     HTTP http://user-service:8080     📡 Network call
```

### Pros & Cons

**✅ Pros:**
- Independent deployment (deploy product without touching order)
- Independent scaling (scale order service 10x, product 2x)
- Technology diversity (product in Go, order in Java)
- Team autonomy (different teams own different services)
- Fault isolation (product failure doesn't affect order)
- Database per service (different schemas)

**❌ Cons:**
- Complex infrastructure (Kubernetes, service mesh)
- Network latency (HTTP calls between services)
- Distributed tracing needed
- More operational overhead
- Eventual consistency challenges
- Higher cost (more servers, more monitoring)

**🎯 Use When:**
- Large team (20+ developers)
- Need independent scaling
- Different services have different SLAs
- High availability critical
- Traffic > 10000 req/sec
- Multiple teams/products

### Deployment

```bash
# Kubernetes deployment
kubectl apply -f product-service.yaml
kubectl apply -f order-service.yaml
kubectl apply -f user-service.yaml

# Or Docker Compose
docker-compose up -d
```

## Decision Matrix

| Requirement | Recommended Pattern |
|-------------|---------------------|
| Just starting | **Monolith Single** |
| < 10 developers | **Monolith Single** |
| 10-20 developers | **Monolith Multi** |
| 20+ developers | **Microservices** |
| Public/Internal APIs | **Monolith Multi** or **Microservices** |
| Different scaling needs | **Microservices** |
| Budget constraints | **Monolith Single/Multi** |
| High availability critical | **Microservices** |
| Simple deployment | **Monolith Single/Multi** |
| Independent deployments | **Microservices** |

## Migration Path

Most teams follow this path:

```
1. Start
   └─> Monolith Single Port
        (Simple, fast to develop)
        
2. Growth (10+ devs)
   └─> Monolith Multi Port
        (Logical separation, still simple)
        
3. Scale (20+ devs, high traffic)
   └─> Microservices
        (Full independence, complexity justified)
```

**🚨 Warning:** Don't start with microservices unless you have:
- 20+ developers
- DevOps team
- Kubernetes experience
- Budget for infrastructure

## Configuration Comparison

### Same Application, Different Configs

All three patterns use the **same code** and **same services**, just different YAML:

```go
// This code is IDENTICAL in all three patterns
func setupServices() {
    lokstra_registry.RegisterServiceFactory("product", ProductServiceFactory)
    lokstra_registry.RegisterServiceFactory("order", OrderServiceFactory)
    lokstra_registry.RegisterServiceFactory("user", UserServiceFactory)
}

func setupRouters() {
    lokstra_registry.RegisterRouter("product-api", createProductRouter())
    lokstra_registry.RegisterRouter("order-api", createOrderRouter())
    lokstra_registry.RegisterRouter("user-api", createUserRouter())
}
```

**Only config.yaml changes!**

## Environment Variables

### Monolith (Single/Multi)

```bash
# All services in one process
DB_HOST=localhost
DB_PORT=5432
CACHE_HOST=localhost
CACHE_PORT=6379
```

### Microservices

```bash
# Product Service
PRODUCT_SERVICE_URL=http://product-service
DB_HOST=product-db
DB_PORT=5432

# Order Service
ORDER_SERVICE_URL=http://order-service
PRODUCT_SERVICE_URL=http://product-service  # For inter-service calls
USER_SERVICE_URL=http://user-service
DB_HOST=order-db
DB_PORT=5432

# User Service
USER_SERVICE_URL=http://user-service
DB_HOST=user-db
DB_PORT=5432
```

## Summary

| Pattern | Complexity | Cost | Scalability | Best For |
|---------|-----------|------|-------------|----------|
| **Monolith Single** | ⭐ | 💰 | ⭐⭐ | Startups, MVPs |
| **Monolith Multi** | ⭐⭐ | 💰💰 | ⭐⭐⭐ | Growing apps |
| **Microservices** | ⭐⭐⭐ | 💰💰💰 | ⭐⭐⭐⭐⭐ | Large enterprises |

**Golden Rule:** Start simple, evolve as needed. Most teams never need microservices.

## Next Steps

- Try each pattern with the same application code
- Measure performance differences
- Understand trade-offs before committing to microservices
