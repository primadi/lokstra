# Runnable Reverse Proxy Test

Complete working example demonstrating Lokstra's reverse proxy configuration with 3 applications.

## Architecture

```
┌─────────────────────────────────────────────────┐
│          API Gateway (App3)                     │
│          Port: 8080                             │
│  ┌──────────────────────────────────────────┐   │
│  │  Reverse Proxy Router                    │   │
│  │  - /users/*    → localhost:9090          │   │
│  │  - /products/* → localhost:9091          │   │
│  └──────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
           │                    │
           ▼                    ▼
    ┌─────────────┐      ┌─────────────┐
    │   App1      │      │   App2      │
    │ Port: 9090  │      │ Port: 9091  │
    │             │      │             │
    │ User        │      │ Product     │
    │ Service     │      │ Service     │
    └─────────────┘      └─────────────┘
```

## Services

### App1 (Backend - Users) - Port 9090
REST API for user management:
- `GET /users` - List all users
- `GET /users/:id` - Get user by ID
- `POST /users` - Create user
- `PUT /users/:id` - Update user
- `DELETE /users/:id` - Delete user

### App2 (Backend - Products) - Port 9091
REST API for product management:
- `GET /products` - List all products
- `GET /products/:id` - Get product by ID
- `POST /products` - Create product
- `PUT /products/:id` - Update product
- `DELETE /products/:id` - Delete product

### App3 (API Gateway) - Port 8080
Reverse proxy that routes:
- `/users/*` → `http://localhost:9090/*` (strip prefix)
- `/products/*` → `http://localhost:9091/*` (strip prefix)

## Running

### 1. Start the application

```bash
cd cmd_draft/examples/reverse-proxy-gateway/test-runnable
go run main.go
```

### 2. Verify services are running

You should see:
```
====================================
🚀 Starting Lokstra Reverse Proxy Test
====================================

📋 Services Running:
  • App1 (Backend Users):    http://localhost:9090
  • App2 (Backend Products): http://localhost:9091
  • App3 (API Gateway):      http://localhost:8080

🔀 Reverse Proxy Routing:
  • /users/*    → http://localhost:9090/*
  • /products/* → http://localhost:9091/*

✅ All services started. Use test.http to test endpoints.
====================================
```

### 3. Test with HTTP requests

#### Using VS Code REST Client
Open `test.http` and click "Send Request" above any `###` line.

#### Using curl

**Direct access to backends:**
```bash
# App1 - Users
curl http://localhost:9090/users

# App2 - Products
curl http://localhost:9091/products
```

**Through API Gateway (reverse proxy):**
```bash
# Get users through gateway (proxied to App1)
curl http://localhost:8080/users

# Get products through gateway (proxied to App2)
curl http://localhost:8080/products

# Get specific user
curl http://localhost:8080/users/1

# Get specific product
curl http://localhost:8080/products/101

# Create user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"David","email":"david@example.com"}'

# Create product
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Monitor","price":350,"stock":80}'
```

## Configuration

All configuration is in `config.yaml`:

```yaml
servers:
  # Backend Server 1
  - name: backend-app1
    apps:
      - addr: ":9090"
        services:
          - user-service

  # Backend Server 2
  - name: backend-app2
    apps:
      - addr: ":9091"
        services:
          - product-service

  # API Gateway with Reverse Proxy
  - name: api-gateway
    apps:
      - addr: ":8080"
        reverse-proxies:
          - prefix: /users
            strip-prefix: true
            target: http://localhost:9090
          - prefix: /products
            strip-prefix: true
            target: http://localhost:9091
```

## Testing Checklist

Use `test.http` to verify:

- ✅ Direct access to App1 (port 9090) works
- ✅ Direct access to App2 (port 9091) works
- ✅ Gateway routes `/users/*` to App1
- ✅ Gateway routes `/products/*` to App2
- ✅ Strip prefix works correctly
- ✅ All HTTP methods (GET, POST, PUT, DELETE) work
- ✅ 404 for invalid routes
- ✅ Response includes "source" field showing origin

## Expected Results

### Get Users Through Gateway
```bash
curl http://localhost:8080/users
```

Response:
```json
{
  "users": [
    {"id": 1, "name": "Alice", "email": "alice@example.com"},
    {"id": 2, "name": "Bob", "email": "bob@example.com"},
    {"id": 3, "name": "Charlie", "email": "charlie@example.com"}
  ],
  "source": "App1 (port 9090)"
}
```

### Get Products Through Gateway
```bash
curl http://localhost:8080/products
```

Response:
```json
{
  "products": [
    {"id": 101, "name": "Laptop", "price": 1200, "stock": 50},
    {"id": 102, "name": "Mouse", "price": 25, "stock": 200},
    {"id": 103, "name": "Keyboard", "price": 75, "stock": 150}
  ],
  "source": "App2 (port 9091)"
}
```

## Verification

The `"source"` field in responses helps verify which backend processed the request:
- `"source": "App1 (port 9090)"` - Processed by User Service
- `"source": "App2 (port 9091)"` - Processed by Product Service

## Strip Prefix Behavior

With `strip-prefix: true`:

| Client Request | Gateway Receives | Backend Receives |
|----------------|------------------|------------------|
| `GET /users` | `GET /users` | `GET /` |
| `GET /users/1` | `GET /users/1` | `GET /1` |
| `GET /products/101` | `GET /products/101` | `GET /101` |

The backend services don't need to know about the `/users` or `/products` prefix.

## Key Features Demonstrated

✅ **Config-driven** - Everything configured via YAML  
✅ **Multiple backends** - Route to different services  
✅ **Strip prefix** - Clean backend routing  
✅ **REST conventions** - Auto-generated routes from services  
✅ **Hybrid mode** - Can combine with regular routers  
✅ **Production-ready** - Uses `httputil.ReverseProxy`

## Troubleshooting

### Port already in use
If you get "address already in use" errors:
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/Mac
lsof -ti:8080 | xargs kill -9
```

### Services not starting
Check the logs for:
- Configuration errors
- Missing service registrations
- Port conflicts

### Gateway not routing
Verify:
1. Backend services are running (check ports 9090, 9091)
2. Gateway started after backends (500ms delay in code)
3. Config file is correct

## Files

- `config.yaml` - Complete configuration
- `main.go` - Runnable application
- `test.http` - HTTP test requests
- `README.md` - This file

## Next Steps

Try modifying:
1. Add more backend services
2. Change strip-prefix behavior
3. Add custom routes alongside proxies
4. Test with real microservices
5. Add middleware to gateway

---

**Happy Testing! 🚀**
