# Router Integration Demo

This example demonstrates **Router Integration** - a better approach than Service Integration for building scalable applications with Lokstra framework.

## Concept: Router Integration vs Service Integration

### ❌ Problems with Service Integration (Old Approach)
- Framework hardcoded specific service names ("product-service", "user-service")  
- Framework should be generic, not tied to specific business domains
- Services are internal abstractions, routers handle HTTP communication
- Unnecessary network overhead for local calls

### ✅ Router Integration (Better Approach)  
- Framework provides generic router communication mechanisms
- Applications define their own router names and URLs
- Local routers use **httptest** (zero network overhead)
- Remote routers use **HTTP** calls seamlessly
- Same business logic works for monolith or microservices

## How Router Integration Works

### 1. Router Registration
```go
// Business logic routers (same in all deployments)
productRouter := lokstra.NewRouter("product-api")
productRouter.GET("/products", getProductsHandler)
lokstra_registry.RegisterRouter("product-api", productRouter)

orderRouter := lokstra.NewRouter("order-api") 
orderRouter.POST("/orders", createOrderHandler)
lokstra_registry.RegisterRouter("order-api", orderRouter)
```

### 2. Router Communication
```go
// Get router client (automatically local or remote)
productClient := lokstra_registry.GetRouterClient("product-api")

// This call uses:
// - httptest.NewRequest() for local routers (zero overhead)
// - http.Client.Do() for remote routers (network call)
resp, err := productClient.GET("/products/123")
```

### 3. Deployment Configuration
```go
// Application registers its own router URLs (not framework)
switch deploymentType {
case "monolith":
    // All routers are local (httptest)
    
case "microservices":
    // Register remote router URLs
    lokstra_registry.RegisterRouterURL("product-api", "http://localhost:8081")
    lokstra_registry.RegisterRouterURL("order-api", "http://localhost:8082")
}
```

## Running the Demo

### Monolith Mode (All routers local - zero overhead)
```bash
go run main.go
# Uses config.yaml (deployment-type: monolith)
```

### Microservices Mode

#### Option 1: Gateway Server (All routers remote)
```bash
# Load microservices config
copy config-microservices.yaml config.yaml
go run main.go
```

#### Option 2: Individual Services

**Terminal 1 - Product Service:**
```bash
copy config-product-service.yaml config.yaml
go run main.go
# Runs on :8081, order-api calls go to :8082
```

**Terminal 2 - Order Service:**  
```bash
copy config-order-service.yaml config.yaml
go run main.go
# Runs on :8082, product-api calls go to :8081
```

## Testing Router Integration

### Test Product API
```bash
# Get all products
curl http://localhost:8080/api/v1/products

# Get specific product
curl http://localhost:8080/api/v1/products/1
```

### Test Cross-Router Communication
```bash
# Create order (order-api calls product-api automatically)
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "product_ids": ["1", "2"]
  }'
```

### Watch the Integration in Action

In **monolith mode**, you'll see:
```
🏠 Router 'product-api': Local (httptest)
🏠 Router 'order-api': Local (httptest)
```

In **microservices mode**, you'll see:
```
🌐 Router 'product-api': Remote at http://localhost:8081
🌐 Router 'order-api': Remote at http://localhost:8082
```

## Benefits of Router Integration

1. **Framework Genericity**: No hardcoded business domain names in framework
2. **Zero Local Overhead**: Local calls use httptest, not network
3. **Seamless Communication**: Same code works for local and remote routers  
4. **Application Control**: Apps define their own router names and URLs
5. **Deployment Flexibility**: Switch between monolith/microservices with config
6. **Better Architecture**: Routers handle HTTP, not internal services

## Configuration Files

- `config.yaml` - Monolith mode (all local)
- `config-microservices.yaml` - Gateway mode (all remote)  
- `config-product-service.yaml` - Product service instance
- `config-order-service.yaml` - Order service instance

## Key Framework Components

- `lokstra_registry.RouterClient` - Handles local/remote router calls
- `lokstra_registry.GetRouterClient()` - Creates router client
- `lokstra_registry.RegisterRouterURL()` - Registers remote router URLs  
- `lokstra_registry.AutoConfigureRouterIntegration()` - Auto-config from deployment-type

This approach makes Lokstra a truly generic framework that can build any type of scalable application without domain-specific assumptions.