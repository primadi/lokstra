# 04-Config-Driven Deployment

A complete e-commerce API demonstrating config-driven architecture where only factories and handlers are in code - everything else is YAML.

## Quick Start

```bash
go run .
```

## What's Config-Driven?

**In Code (Factories & Handlers):**
- Service factories (how to create services)
- Router setup (HTTP endpoints and logic)

**In YAML (Everything Else):**
- Service instances (which services to create)
- Service configuration (DB host, cache port, email settings)
- Server topology (ports, routers)
- Environment-specific settings

## Architecture

```
┌─────────────────────────────────────────┐
│ main.go (Minimal - loads config)       │
└────────────┬────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│ config.yaml (Defines everything)       │
│ - Services (9 services across 3 layers)│
│ - Server (ports, routers)              │
│ - Environment variables                 │
└────────────┬────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│ Framework assembles application        │
│ - Creates services from factories      │
│ - Mounts routers                        │
│ - Starts HTTP server                    │
└─────────────────────────────────────────┘
```

## Service Layers

### Layer 1: Infrastructure (3 services)
- `db-service` - PostgreSQL connection
- `cache-service` - Redis cache
- `email-service` - SMTP email

### Layer 2: Repositories (3 services)
- `user-repository` → db-service
- `product-repository` → db-service, cache-service
- `order-repository` → db-service

### Layer 3: Domain Services (3 services)
- `user-service` → user-repository
- `product-service` → product-repository
- `order-service` → order-repository, product-service, user-service, email-service

## Complete Example Flow

### 1. Create Order Request

```http
POST /api/orders
{
  "user_id": "123",
  "product_id": "1",
  "quantity": 2
}
```

### 2. Service Orchestration

```
OrderService.CreateOrder()
├─ ProductService.GetProduct()
│  └─ ProductRepository.FindByID()
│     ├─ CacheService.Get() ← Try cache first
│     └─ DBService.Query()  ← Cache miss, query DB
│
├─ Calculate: subtotal, tax, total
├─ OrderRepository.Create()
│  └─ DBService.Execute() ← Insert order
│
├─ UserService.GetUser()
│  └─ UserRepository.FindByID()
│     └─ DBService.Query() ← Get user email
│
└─ EmailService.Send() ← Send confirmation
```

### 3. Result

- Order created in database
- User notified via email
- Product cached for next request
- Response returned to client

## Benefits

### 1. Environment-Specific Configuration

```yaml
# Development (config.yaml defaults)
services:
  - name: db-service
    config:
      host: ${DB_HOST:localhost}  # localhost for dev

# Production (environment variables)
export DB_HOST=prod-db.example.com
export DB_PASSWORD=secure-password
```

### 2. No Code Changes for Deployment

**Same code runs in:**
- Development (localhost services)
- Staging (staging servers)
- Production (production servers)

**Just change environment variables!**

### 3. Easy Testing

```bash
# Test environment
DB_HOST=test-db EMAIL_SERVICE=mock go run .

# Mock services for unit tests
old_registry.RegisterServiceFactory("email", MockEmailFactory)
```

### 4. Clear Separation

| Component | Location | Purpose |
|-----------|----------|---------|
| Business Logic | Code | How things work |
| Configuration | YAML | What to create |
| Secrets | Env Vars | Sensitive data |
| Deployment | YAML | Where things run |

## Configuration Examples

### Simple Service

```yaml
services:
  - name: email-service
    type: email
    config:
      smtp_host: ${SMTP_HOST:localhost}
      smtp_port: ${SMTP_PORT:587}
```

### Service with Dependencies

```yaml
services:
  - name: user-service
    type: user
    config:
      password_min_length: ${PASSWORD_MIN_LENGTH:8}
    # Dependencies resolved automatically in factory
```

### Complex Service

```yaml
services:
  - name: order-service
    type: order
    config:
      tax_rate: ${TAX_RATE:0.10}
      min_order_amount: ${MIN_ORDER_AMOUNT:10.00}
    # Depends on: order-repo, product-service, user-service, email-service
```

## Deployment Patterns

See `03-best-practices` for deployment patterns:

### Monolith Single Port
```yaml
servers:
  - name: monolith-server
    apps:
      - addr: ":8080"
        routers: [product-api, order-api, user-api]
```

### Monolith Multi Port
```yaml
servers:
  - name: monolith-server
    apps:
      - addr: ":8080"
        routers: [product-api]
      - addr: ":8081"
        routers: [order-api, user-api]
```

### Microservices
```yaml
# product-service config.yaml
servers:
  - name: product-service
    apps:
      - addr: ":8080"
        routers: [product-api]

# order-service config.yaml
servers:
  - name: order-service
    apps:
      - addr: ":8080"
        routers: [order-api]
```

## Best Practices

### ✅ DO

1. **Define services in YAML**
   ```yaml
   services:
     - name: db-service
       type: db
       config: {...}
   ```

2. **Use environment variables for secrets**
   ```yaml
   password: ${DB_PASSWORD}  # Never hardcode!
   ```

3. **Layer your services**
   - Infrastructure → Repositories → Domain

4. **Keep main.go minimal**
   ```go
   func main() {
       setupFactories()
       setupRouters()
       loadConfig()
       startServer()
   }
   ```

### ❌ DON'T

1. **Don't hardcode configuration**
   ```go
   // BAD!
   db := NewDBService("localhost", 5432, "mydb")
   
   // GOOD!
   // In YAML: config with env vars
   ```

2. **Don't mix config and code**
   ```go
   // BAD!
   if os.Getenv("ENV") == "production" {
       // production config
   } else {
       // dev config
   }
   
   // GOOD!
   // All in YAML with env vars
   ```

3. **Don't create services manually**
   ```go
   // BAD!
   var emailService = NewEmailService(...)
   
   // GOOD!
   email := services.GetEmail()  // From registry
   ```

## Next Steps

- **03-best-practices** - Deployment patterns (monolith vs microservices)

## Summary

This example demonstrates:
- ✅ Complete app from YAML
- ✅ 9 services across 3 layers
- ✅ Automatic dependency resolution
- ✅ Environment-specific configuration
- ✅ Clean separation of concerns
- ✅ Production-ready architecture
