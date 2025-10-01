# Service Inter-Communication Pattern

## 🤔 **Problem: Service Dependencies**

Ketika service perlu memanggil service lain, ada challenge besar:

### **Real-world Order Creation Flow:**
1. **Order Service** receives `POST /orders`
2. Must call **Product Service** → validate products & get prices
3. Must call **User Service** → validate user exists  
4. Must call **Payment Service** → process payment
5. Must call **Analytics Service** → track order event

### **Deployment Challenge:**
- **Monolith**: All services in same process → direct function calls
- **Microservices**: Services in different processes → HTTP calls

## 💡 **Solution Options**

### ❌ **Option 1: Reverse Proxy (Your Current Approach)**
```yaml
apps:
  - name: order-app
    reverse_proxies:
      - path: /internal/products
        target: http://product-service:8081
      - path: /internal/users  
        target: http://user-service:8083
```

**Problems:**
- Complex routing configuration
- Tight coupling between services
- Hard to manage service discovery
- Performance overhead (multiple HTTP hops)

### ✅ **Option 2: Service Client Registry (Better)**
```go
// Service clients auto-resolve based on deployment
productClient := lokstra_registry.GetServiceClient[ProductServiceClient]("product-service")
products := productClient.GetProducts(productIDs)
```

**Benefits:**
- **Monolith**: Direct function calls (same process)
- **Microservices**: HTTP calls (different processes)  
- **Transparent**: Business logic doesn't change
- **Configuration-driven**: Service URLs in YAML

### ✅ **Option 3: Service Interface Pattern (Best)**
```go
// Business logic uses interfaces
type ProductService interface {
    ValidateProducts(ids []string) ([]Product, error)
    GetProductPrices(ids []string) (map[string]float64, error)
}

// Implementation auto-selected based on deployment
productSvc := lokstra_registry.GetService[ProductService]("product-service")
```

## 🚀 **Recommended Implementation**

Let me show you the **Service Interface Pattern** - the cleanest solution!