# Scalable Deployment Example

This example demonstrates Lokstra's deployment flexibility - how the same codebase can be deployed as either a **monolith** or **microservices** just by changing configuration files.

## Project Structure

```
08_scalable_deployment/
├── main.go                           # Single binary entry point
├── go.mod                           # Go module
├── apps/                            # Application handlers
│   ├── user_api.go                  # User management API
│   ├── order_api.go                 # Order management API
│   └── dashboard_htmx.go            # HTMX admin dashboard
├── monolith-config.yaml             # Monolith deployment config
├── user-service-config.yaml         # User microservice config
├── order-service-config.yaml        # Order microservice config
├── dashboard-service-config.yaml    # Dashboard microservice config
├── Dockerfile                       # Multi-stage Docker build
├── docker-compose.monolith.yml      # Monolith deployment
├── docker-compose.microservices.yml # Microservices deployment
├── nginx.conf                       # Load balancer config
└── README.md                        # This file
```

## Applications

This example contains three applications:

1. **User API** (`/api/users`) - User management endpoints
2. **Order API** (`/api/orders`) - Order management endpoints  
3. **Dashboard** (`/dashboard`) - HTMX-powered admin interface

## Deployment Strategies

### 1. Monolith Deployment

**Single container running all applications on one port (8080)**

```bash
# Run monolith
docker-compose -f docker-compose.monolith.yml up --build

# Test endpoints
curl http://localhost:8080/api/users
curl http://localhost:8080/api/orders
curl http://localhost:8080/dashboard
```

**Advantages:**
- Simple deployment and management
- Lower resource usage
- Easier development and debugging
- Single point of monitoring

### 2. Microservices Deployment

**Separate containers for each application with load balancer**

```bash
# Run microservices
docker-compose -f docker-compose.microservices.yml up --build

# Test endpoints through load balancer
curl http://localhost/api/users      # Routes to user-service:8081
curl http://localhost/api/orders     # Routes to order-service:8082
curl http://localhost/dashboard      # Routes to dashboard-service:8083

# Or test services directly
curl http://localhost:8081/api/users
curl http://localhost:8082/api/orders
curl http://localhost:8083/dashboard
```

**Advantages:**
- Independent scaling of each service
- Technology diversity per service
- Fault isolation
- Independent deployments

## Key Lokstra Features Demonstrated

### 1. **One Binary, Multiple Deployments**
```go
// Same main.go for all deployment strategies
func main() {
    configFile := os.Getenv("CONFIG_FILE")
    if configFile == "" {
        configFile = "config.yaml"
    }
    
    app, err := lokstra.NewFromConfigFile(configFile)
    // ... rest of code
}
```

### 2. **Configuration-Driven Architecture**
- **Monolith Config**: All apps in one server
- **Microservice Config**: Each app in separate server with different ports

### 3. **Flexible App Registration**
```yaml
# Monolith: All apps enabled
apps:
  - name: "user-api"
    enabled: true
  - name: "order-api" 
    enabled: true
  - name: "dashboard"
    enabled: true

# Microservice: Only relevant app enabled
apps:
  - name: "user-api"
    enabled: true
```

### 4. **Middleware Inheritance**
- Global middleware applies to all apps
- App-specific middleware for fine-grained control

## Running the Examples

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for local development)

### Local Development
```bash
# Run monolith locally
CONFIG_FILE=monolith-config.yaml go run main.go

# Run individual services locally
CONFIG_FILE=user-service-config.yaml go run main.go &
CONFIG_FILE=order-service-config.yaml go run main.go &
CONFIG_FILE=dashboard-service-config.yaml go run main.go &
```

### Docker Deployment

#### Monolith
```bash
# Build and run monolith
docker-compose -f docker-compose.monolith.yml up --build -d

# Check logs
docker-compose -f docker-compose.monolith.yml logs -f

# Stop
docker-compose -f docker-compose.monolith.yml down
```

#### Microservices
```bash
# Build and run microservices
docker-compose -f docker-compose.microservices.yml up --build -d

# Check logs for specific service
docker-compose -f docker-compose.microservices.yml logs -f user-service

# Scale a specific service
docker-compose -f docker-compose.microservices.yml up --scale user-service=3 -d

# Stop
docker-compose -f docker-compose.microservices.yml down
```

## Testing the Applications

### User API
```bash
# Get all users
curl http://localhost:8080/api/users

# Get specific user
curl http://localhost:8080/api/users/123

# Create user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'
```

### Order API
```bash
# Get all orders
curl http://localhost:8080/api/orders

# Get specific order
curl http://localhost:8080/api/orders/456

# Create order
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "product": "Laptop", "amount": 999.99}'
```

### Dashboard (HTMX)
```bash
# Open in browser
open http://localhost:8080/dashboard

# Or with curl
curl http://localhost:8080/dashboard
curl http://localhost:8080/dashboard/users
curl http://localhost:8080/dashboard/orders
```

## Scaling Scenarios

### Scenario 1: High User Load
Scale only the user service:
```bash
docker-compose -f docker-compose.microservices.yml up --scale user-service=5 -d
```

### Scenario 2: Batch Order Processing
Scale only the order service:
```bash
docker-compose -f docker-compose.microservices.yml up --scale order-service=3 -d
```

### Scenario 3: Development Environment
Use monolith for simplicity:
```bash
docker-compose -f docker-compose.monolith.yml up -d
```

### Scenario 4: Production Environment
Use microservices with load balancer:
```bash
docker-compose -f docker-compose.microservices.yml up -d
```

## Health Checks

All configurations include health check endpoints:

```bash
# Monolith
curl http://localhost:8080/health

# Individual microservices
curl http://localhost:8081/health  # User service
curl http://localhost:8082/health  # Order service
curl http://localhost:8083/health  # Dashboard service

# Through load balancer
curl http://localhost/health
```

## Monitoring and Observability

Both deployment strategies support:
- Request logging middleware
- Health check endpoints
- Graceful shutdowns
- Container health checks

## Production Considerations

### Monolith Advantages
- Simpler deployment pipeline
- Lower infrastructure complexity
- Easier debugging and tracing
- Better for small to medium applications

### Microservices Advantages
- Independent scaling
- Technology diversity
- Fault isolation
- Team autonomy
- Better for large, complex applications

## Next Steps

1. Add database connections to each service
2. Implement inter-service communication
3. Add distributed tracing
4. Implement service discovery
5. Add API gateway for authentication
6. Configure monitoring and alerting

This example showcases Lokstra's core philosophy: **Build once, deploy anywhere** - the same codebase adapts to your deployment needs through configuration.