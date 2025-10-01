# Scalable E-Commerce Application

Contoh aplikasi e-commerce real yang menunjukkan **skalabilitas lokstra framework** melalui konfigurasi YAML. DevOps bisa mengatur deployment sebagai **monolith** atau **microservices** tanpa mengubah kode!

## 🎯 **Key Features**

### 🏗️ **Deployment Flexibility**
- **Monolith**: Single server, semua services dalam satu aplikasi
- **Microservices**: Multiple servers, setiap service terpisah
- **Same Code Base**: Tidak perlu mengubah business logic

### 🔧 **Configuration-Driven**
- Service configuration melalui YAML
- Environment variable support  
- Runtime configuration changes
- Middleware composition via config

### 📊 **Business Domains**
- **Products**: Catalog management
- **Orders**: Transaction processing  
- **Users**: User management & auth
- **Payments**: Secure payment processing
- **Analytics**: Reporting & insights

## 🚀 **Quick Start**

### 1. **Monolith Deployment**
```bash
# Single server with all services
export DEPLOYMENT_TYPE=monolith
go run main.go
```

**Result**: Single server on port 8080
- `GET /api/products` - Product catalog
- `POST /api/orders` - Order creation  
- `GET /api/users` - User management
- `POST /api/payments` - Payment processing
- `GET /api/analytics` - Analytics data

### 2. **Microservices Deployment**

#### Start Product Service:
```bash
export DEPLOYMENT_TYPE=microservices
export SERVER_NAME=product-service
export PRODUCT_PORT=8081
go run main.go
```

#### Start Order Service:
```bash  
export DEPLOYMENT_TYPE=microservices
export SERVER_NAME=order-service
export ORDER_PORT=8082
go run main.go
```

#### Start User Service:
```bash
export DEPLOYMENT_TYPE=microservices  
export SERVER_NAME=user-service
export USER_PORT=8083
go run main.go
```

#### Start Payment Service:
```bash
export DEPLOYMENT_TYPE=microservices
export SERVER_NAME=payment-service
export PAYMENT_PORT=8084
go run main.go
```

#### Start Analytics Service:
```bash
export DEPLOYMENT_TYPE=microservices
export SERVER_NAME=analytics-service  
export ANALYTICS_PORT=8085
go run main.go
```

**Result**: 5 independent services
- Product Service: `GET localhost:8081/products`
- Order Service: `POST localhost:8082/orders`
- User Service: `GET localhost:8083/users`  
- Payment Service: `POST localhost:8084/payments`
- Analytics Service: `GET localhost:8085/analytics`

## 📁 **Configuration Files**

### 🏢 **Monolith Config**
- **File**: `config-monolith.yaml`
- **Server**: Single `monolith-server` with all APIs
- **Services**: Shared database, redis, email
- **Performance**: High throughput, shared resources

### 🔄 **Microservices Configs**
- **Files**: `config-microservices-*.yaml`
- **Servers**: Independent per service
- **Services**: Isolated databases per domain  
- **Performance**: Scalable, fault-tolerant

## 🛠️ **Configuration Highlights**

### **Monolith Configuration**
```yaml
servers:
  - name: monolith-server
    listen: "8080"
    services: [database, redis, email]
    apps:
      - name: ecommerce-app
        path_prefix: "/api"
        routers: [product-api, order-api, user-api, payment-api, analytics-api]
```

### **Microservices Configuration**  
```yaml
# Product Service
servers:
  - name: product-service
    listen: "8081"
    services: [database, redis]
    apps:
      - name: product-app
        routers: [product-api]
```

## 🔧 **Middleware Configuration**

### **Per-Service Customization**
- **Product**: Caching (600s TTL) + Rate limiting (500 RPS)
- **Order**: Auth + No caching (transactional data) + Rate limiting (200 RPS)  
- **User**: Auth + Caching (900s TTL) + Rate limiting (300 RPS)
- **Payment**: Strict auth + Short caching (30s TTL) + Low rate limit (50 RPS)
- **Analytics**: Heavy caching (1800s TTL) + Rate limiting (100 RPS)

### **Environment-Specific Configs**
```yaml
configs:
  - name: database-url
    value: "${DATABASE_URL:postgres://localhost:5432/ecommerce}"
  - name: cors-origins  
    value: "${CORS_ORIGINS:*}"
  - name: payment-gateway
    value: "${PAYMENT_GATEWAY:stripe}"
```

## 🎬 **Demo Scenarios**

### **Scenario 1: Development** (Monolith)
```bash
export DEPLOYMENT_TYPE=monolith
export DATABASE_URL="postgres://localhost:5432/dev_ecommerce"
export CORS_ORIGINS="http://localhost:3000"
go run main.go
```

### **Scenario 2: Production** (Microservices)
```bash
# Terminal 1: Product Service
export DEPLOYMENT_TYPE=microservices
export SERVER_NAME=product-service
export PRODUCT_DB_URL="postgres://prod-db:5432/products"
export PRODUCT_REDIS_URL="redis://prod-redis:6379/0"
go run main.go

# Terminal 2: Order Service  
export DEPLOYMENT_TYPE=microservices
export SERVER_NAME=order-service
export ORDER_DB_URL="postgres://prod-db:5432/orders"
export ORDER_REDIS_URL="redis://prod-redis:6379/1"  
go run main.go

# ... repeat for other services
```

## ✅ **Benefits Demonstrated**

### 🎯 **DevOps Flexibility**
- **Same codebase** untuk development dan production
- **Environment-specific** configuration
- **Gradual migration** dari monolith ke microservices

### 🚀 **Performance Optimization**
- **Service-specific** middleware tuning
- **Isolated** database connections per service
- **Custom** caching strategies per domain

### 🔒 **Security**
- **Per-service** rate limiting
- **Environment-specific** CORS policies  
- **Graduated** authentication requirements

### 📈 **Scalability**
- **Independent** service scaling
- **Domain-specific** resource allocation
- **Zero-downtime** service updates

## 🏆 **Production Ready Features**

- ✅ **Multi-environment** configuration support
- ✅ **Graceful** service isolation  
- ✅ **Performance** tuning per service
- ✅ **Security** policies per domain
- ✅ **Monitoring** and logging integration
- ✅ **Zero-code** deployment strategy changes

---

## 🎉 **Conclusion**

Aplikasi ini membuktikan bahwa dengan **lokstra framework** dan **konfigurasi YAML**, tim development bisa fokus pada business logic, sementara DevOps team punya **full control** atas deployment architecture tanpa perlu mengubah satu baris kode pun! 🚀