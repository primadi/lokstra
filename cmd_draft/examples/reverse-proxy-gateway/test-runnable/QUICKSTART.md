# Quick Start - Reverse Proxy Test

## 🚀 Start the Application

```bash
cd cmd_draft/examples/reverse-proxy-gateway/test-runnable
go run main.go
```

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

## 🧪 Test with curl

### Via Gateway (Reverse Proxy)

```bash
# Get all users (proxied to App1:9090)
curl http://localhost:8080/users

# Get user by ID
curl http://localhost:8080/users/1

# Get all products (proxied to App2:9091)
curl http://localhost:8080/products

# Get product by ID
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

### Direct to Backends (No Proxy)

```bash
# Direct to App1
curl http://localhost:9090/users

# Direct to App2
curl http://localhost:9091/products
```

## 📝 Test with VS Code REST Client

1. Open `test.http` in VS Code
2. Click "Send Request" above any `###` marker
3. View response in the side panel

## ✅ Expected Responses

All responses include a `"source"` field to verify routing:

**Via Gateway → App1:**
```json
{
  "users": [...],
  "source": "App1 (port 9090)"
}
```

**Via Gateway → App2:**
```json
{
  "products": [...],
  "source": "App2 (port 9091)"
}
```

## 📊 Architecture

```
Client Request
     ↓
API Gateway (8080)
     ├─ /users/*    → App1 (9090)
     └─ /products/* → App2 (9091)
```

## 🛑 Stop the Application

Press `Ctrl+C` in the terminal running the application.

---

For detailed information, see `README.md`
