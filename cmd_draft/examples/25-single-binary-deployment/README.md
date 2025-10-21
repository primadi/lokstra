# Example 25: Single Binary, Multiple Deployment Scenarios

## Overview

This example demonstrates the **correct pattern** for Lokstra deployment:
- **1 binary** (single main.go)
- **3 config files** (monolith, multiport, microservices)
- **Runtime server selection** via command-line flag
- **Auto-generated routers** from services

## Key Concepts

### 1. Service Definition ≠ Router

```yaml
services:
  - name: user-service
    type: user_service
    config:
      storage: memory
```

**This only defines the service.** Router is NOT automatically generated here.

### 2. Router Auto-Generated from apps.services

```yaml
servers:
  - name: auth-server
    apps:
      - services: [user-service, auth-service]  # 👈 Router generated HERE!
```

When `user-service` is listed in `apps.services[]`, framework:
1. Creates service instance
2. **Auto-generates HTTP router** from service interface
3. Mounts router to app

### 3. Service Dependencies

```yaml
services:
  - name: order-service
    type: order_service
    depends-on: [user-service, payment-service]  # 👈 Lazy dependencies
```

In factory:
```go
func CreateOrderService(cfg map[string]any) any {
    // ✅ Gets lazy service references from depends-on
    userSvc := GetLazyService[UserService](cfg, "user-service")
    paymentSvc := GetLazyService[PaymentService](cfg, "payment-service")
    
    return &OrderService{
        userSvc:    userSvc,
        paymentSvc: paymentSvc,
    }
}
```

### 4. Service Config Usage

```yaml
services:
  - name: order-service
    config:
      storage: memory      # 👈 Must be used!
      max_items: 100
```

```go
func CreateOrderService(cfg map[string]any) any {
    storage := utils.GetValueFromMap(cfg, "storage", "memory")  // ✅ USE IT!
    maxItems := utils.GetValueFromMap(cfg, "max_items", 10)     // ✅ USE IT!
    
    return &OrderService{
        storage:  storage,
        maxItems: maxItems,
    }
}
```

## Three Deployment Scenarios

### Scenario 1: Monolith (All-in-One)

**File:** `config-monolith.yaml`

```yaml
servers:
  - name: monolith
    base-url: http://localhost:8080
    apps:
      - addr: ":8080"
        services: [user-service, auth-service, order-service, 
                   cart-service, payment-service, invoice-service]
```

**All services** in one process, one port.

**Run:**
```bash
go run main.go -config config-monolith.yaml -server monolith
```

**Result:**
- All services LOCAL
- No remote calls
- Fastest (in-process)
- Port 8080: all APIs

### Scenario 2: Monolith Multiport (Logical Separation)

**File:** `config-multiport.yaml`

```yaml
servers:
  - name: auth-server
    apps:
      - addr: ":8081"
        services: [user-service, auth-service]
        
  - name: business-server
    apps:
      - addr: ":8082"
        services: [order-service, cart-service, payment-service, invoice-service]
```

**Same process**, but listening on multiple ports for logical separation.

**Run:**
```bash
go run main.go -config config-multiport.yaml -server all
```

**Result:**
- All services LOCAL (same process)
- Multiple ports (8081, 8082)
- No remote calls (in-process)
- Logical separation for monitoring/routing

### Scenario 3: Microservices (True Distribution)

**File:** `config-microservices.yaml`

```yaml
servers:
  - name: auth-server
    base-url: http://localhost:8081
    apps:
      - addr: ":8081"
        services: [user-service, auth-service]
        
  - name: order-server
    base-url: http://localhost:8082
    apps:
      - addr: ":8082"
        services: [order-service, cart-service]
        
  - name: payment-server
    base-url: http://localhost:8083
    apps:
      - addr: ":8083"
        services: [payment-service, invoice-service]
```

**Run 3 separate processes:**

```bash
# Terminal 1
go run main.go -config config-microservices.yaml -server auth-server

# Terminal 2
go run main.go -config config-microservices.yaml -server order-server

# Terminal 3
go run main.go -config config-microservices.yaml -server payment-server
```

**Result:**
- Services distributed across processes
- Cross-server HTTP calls
- True microservices architecture

## How It Works

### 1. Single main.go

```go
func main() {
    configFile := flag.String("config", "config-monolith.yaml", "Config file")
    serverName := flag.String("server", "", "Server name to run")
    flag.Parse()
    
    // Load config
    cfg := config.New()
    config.LoadConfigFile(*configFile, cfg)
    
    // Set server
    if *serverName == "all" {
        // Run all servers (multiport mode)
        old_registry.RunAllServers()
    } else {
        // Run specific server
        old_registry.SetCurrentServerName(*serverName)
        old_registry.RegisterConfig(cfg)
        old_registry.StartServer()
    }
}
```

### 2. Service Auto-Router Generation

When service is listed in `apps.services[]`:

```go
// Framework automatically generates:
router := router.New(serviceName)

// For each method in service interface:
// - POST /users (CreateUser)
// - GET /users/{id} (GetUser)
// - GET /users (ListUsers)

router.POST("/users", func(ctx *request.Context) error {
    var req CreateUserRequest
    ctx.Req.BindBody(&req)
    
    svc := GetService[UserService](serviceName)
    result, err := svc.CreateUser(ctx, req.Name, req.Email)
    // ...
})
```

### 3. Service Location Inference

```
Config: order-server runs order-service, cart-service

Framework builds:
serviceLocationMap = {
    "user-service": {server: "auth-server", baseURL: "http://localhost:8081", isLocal: false}
    "order-service": {server: "order-server", baseURL: "http://localhost:8082", isLocal: true}
    "cart-service": {server: "order-server", baseURL: "http://localhost:8082", isLocal: true}
    "payment-service": {server: "payment-server", baseURL: "http://localhost:8083", isLocal: false}
}

When order-server starts:
- order-service, cart-service → LOCAL factory (in-process)
- user-service, payment-service → REMOTE factory (HTTP client)
```

## File Structure

```
25-single-binary-deployment/
├── main.go                          # ✅ ONLY 1 MAIN.GO
├── config-monolith.yaml             # Scenario 1: All-in-one
├── config-multiport.yaml            # Scenario 2: Logical separation
├── config-microservices.yaml        # Scenario 3: True distribution
├── services/
│   ├── user_service.go             # Business logic only
│   ├── order_service.go            # Uses lazy dependencies
│   └── payment_service.go          # Uses config values
└── test-api.http                    # Test all scenarios
```

## Benefits

### 1. Development: Monolith
- Fast startup
- Easy debugging
- Single process
- No network overhead

### 2. Staging: Multiport
- Test port separation
- Same codebase
- Prepare for microservices
- Monitor per logical component

### 3. Production: Microservices
- True scalability
- Independent deployment
- Fault isolation
- Same binary, different config!

## Configuration Comparison

| Feature | Monolith | Multiport | Microservices |
|---------|----------|-----------|---------------|
| Processes | 1 | 1 | 3 |
| Ports | 1 (8080) | 3 (8081-8083) | 3 (8081-8083) |
| Service Calls | In-process | In-process | HTTP |
| Deployment | Single deploy | Single deploy | Independent |
| Scaling | Vertical | Vertical | Horizontal |
| Debugging | Easy | Easy | Complex |
| Latency | Lowest | Lowest | Higher |
| Fault Isolation | None | Logical | Physical |

## Next Steps

1. ✅ Understand the 3 deployment scenarios
2. ✅ Test monolith mode first (simplest)
3. ✅ Try multiport to see logical separation
4. ✅ Run microservices to see cross-server calls
5. ✅ Notice: **Same code, just different config!**

## Key Takeaways

1. **Service definition** = business logic only
2. **Router generation** = happens when service listed in apps.services
3. **Service config** = must be used in factories
4. **Dependencies** = defined in depends-on, accessed via lazy services
5. **1 binary** = all deployment modes from single codebase
6. **Runtime selection** = -server flag determines which services are local
