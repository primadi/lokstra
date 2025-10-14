# Router Integration Example

This example demonstrates Lokstra's Router Integration capabilities, showing how business logic routers can be deployed in different architectural patterns using YAML configuration.

## Deployment Types

This example supports three different deployment patterns:

### 1. Monolith Single Port (`monolith-single-port`)
- All routers (product-api, order-api) run in one server on one port
- Server: `monolith-single-port-server` on `:8081`
- Both Product API and Order API accessible from the same port

### 2. Monolith Multi Port (`monolith-multi-port`) 
- All routers run in one server but on different ports
- Server: `monolith-multi-port-server`
  - Product API on `:8081`
  - Order API on `:8082`

### 3. Microservices (`microservice`)
- Each router runs as a separate service on different ports
- `product-service` on `:8082` (product-api router)
- `order-service` on `:8083` (order-api router)

## Configuration

The deployment configuration is defined in `config.yaml` with the following structure:

```yaml
servers:
  - name: monolith-single-port-server
    baseUrl: http://localhost
    deployment-id: monolith-single-port
    apps: 
      - addr: ":8081"
        routers: [product-api, order-api]

  - name: monolith-multi-port-server
    baseUrl: http://localhost
    deployment-id: monolith-multi-port
    apps:
      - addr: ":8081"
        routers: [product-api]
      - addr: ":8082"
        routers: [order-api]

  - name: product-service
    baseUrl: http://localhost
    deployment-id: microservice
    apps:
      - addr: ":8082"
        routers: [product-api]
        
  - name: order-service
    baseUrl: http://localhost
    deployment-id: microservice
    apps:
      - addr: ":8083"
        routers: [order-api]
```

## Running the Example

### Method 1: Using Helper Scripts

#### Windows (PowerShell)
```powershell
# Interactive mode - will prompt for deployment type
.\run.ps1

# Direct deployment
.\run.ps1 monolith-single-port
.\run.ps1 monolith-multi-port
.\run.ps1 microservice

# Run specific microservice
.\run.ps1 microservice -ServerName product-service
```

#### Unix/Linux/Mac (Bash)
```bash
# Make script executable first
chmod +x run.sh

# Interactive mode
./run.sh

# Direct deployment  
./run.sh monolith-single-port
./run.sh monolith-multi-port
./run.sh microservice

# Run specific microservice
./run.sh microservice product-service
```

### Method 2: Using Environment Variables

```bash
# Run all servers in monolith-single-port deployment
export DEPLOYMENT_ID=monolith-single-port
export SERVER_NAME=all
go run .

# Run all servers in microservice deployment
export DEPLOYMENT_ID=microservice
export SERVER_NAME=all
go run .

# Run only product-service in microservice deployment
export DEPLOYMENT_ID=microservice  
export SERVER_NAME=product-service
go run .
```

### Method 3: Interactive Mode

```bash
# Will prompt you to choose deployment type
go run .
```

## API Endpoints

### Product API Router (`product-api`)
- `GET /ping` - Health check for product API
- `GET /products` - List all products
- `GET /products/{id}` - Get specific product by ID

### Order API Router (`order-api`)
- `GET /ping` - Health check for order API  
- `GET /orders` - List all orders
- `POST /orders` - Create new order (requires product IDs)

### Health Check (Auto-generated for each server)
- `GET /health` - Server health status with deployment information

## Testing Different Deployments

### Monolith Single Port (`:8081`)
```bash
curl http://localhost:8081/health
curl http://localhost:8081/products
curl http://localhost:8081/orders
```

### Monolith Multi Port
```bash
# Product API on :8081
curl http://localhost:8081/health  
curl http://localhost:8081/products

# Order API on :8082
curl http://localhost:8082/health
curl http://localhost:8082/orders
```

### Microservices
```bash
# Product Service on :8082
curl http://localhost:8082/health
curl http://localhost:8082/products

# Order Service on :8083  
curl http://localhost:8083/health
curl http://localhost:8083/orders
```

## Example API Usage

### Create an Order
```bash
# First get available products
curl http://localhost:8081/products  # (adjust port based on deployment)

# Create order with product IDs
curl -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "product_ids": ["1", "2"]
  }'
```

## Key Features Demonstrated

1. **Flexible Deployment**: Same business logic code can be deployed in different architectural patterns
2. **Configuration-Driven**: Deployment strategy controlled by YAML configuration
3. **Service Discovery**: Router integration handles communication between services
4. **Health Monitoring**: Each deployment provides health endpoints with deployment information
5. **Development Flexibility**: Easy switching between monolith and microservice architectures

## Files

- `main.go` - Main application with deployment logic
- `config.yaml` - Deployment configuration 
- `setup_routers.go` - Business logic router definitions
- `handlers.go` - API handler implementations
- `run.ps1` - PowerShell helper script
- `run.sh` - Bash helper script
- `test-request.http` - HTTP test requests