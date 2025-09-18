# Example HTTP Requests

This directory contains example HTTP requests that demonstrate all the binding approaches in the Lokstra framework. You can use these with curl, Postman, or any HTTP client.

## 🚀 Quick Test All Endpoints

Run this script to test all endpoints:

```bash
#!/bin/bash

# Start server first: go run main.go

echo "=== Testing Lokstra Request Binding Examples ==="
echo "🏥 Health Check"
curl -s http://localhost:8080/health | jq '.'
echo -e "\n"

echo "📝 Manual Binding (GET with query params and headers)"
curl -s "http://localhost:8080/users/user123?page=2&limit=10&tags=web&tags=api&active=true" \
  -H "Authorization: Bearer token123" \
  -H "User-Agent: TestClient/1.0" | jq '.'
echo -e "\n"

echo "🤖 Smart Binding (POST with all parameter types)"
curl -s -X POST "http://localhost:8080/users/user456/smart?page=1&limit=5&tags=premium&active=false" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer smarttoken" \
  -H "User-Agent: SmartClient/2.0" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com", 
    "age": 30,
    "preferences": {
      "theme": "dark",
      "language": "en",
      "notifications": true
    }
  }' | jq '.'
echo -e "\n"

echo "🗺️ BindBodySmart to Map (Dynamic JSON)"
curl -s -X POST http://localhost:8080/users/create-map \
  -H "Content-Type: application/json" \
  -d '{
    "dynamic_field_1": "value1",
    "nested_object": {
      "key": "value",
      "number": 42
    },
    "array_field": ["item1", "item2"],
    "boolean_field": true
  }' | jq '.'
echo -e "\n"

echo "❌ BindAllSmart to Map (Limitation Demo)"
curl -s -X POST http://localhost:8080/users/user789/all-map \
  -H "Content-Type: application/json" \
  -d '{"name": "Charlie", "age": 35}' | jq '.'
echo -e "\n"

echo "🔀 Hybrid Binding (Recommended approach)"
curl -s -X POST "http://localhost:8080/users/hybrid123/hybrid?page=3&limit=20&tags=vip&tags=beta" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer hybridtoken" \
  -d '{
    "profile": {
      "firstName": "David",
      "lastName": "Smith",
      "avatar": "https://example.com/avatar.jpg"
    },
    "settings": {
      "notifications": {
        "email": true,
        "sms": false,
        "push": true
      },
      "privacy": {
        "showEmail": false,
        "showProfile": true
      }
    },
    "metadata": {
      "source": "api",
      "version": "2.0"
    }
  }' | jq '.'
echo -e "\n"

echo "🔍 Complex Query Parameters"
curl -s "http://localhost:8080/search?q=lokstra&filter=type:web&filter=lang:go&sort=name&page=1&limit=10&opt[format]=json&opt[include]=docs&date=2023-01-01&date=2023-12-31" | jq '.'
echo -e "\n"

echo "📝 Form Data Binding"
curl -s -X POST http://localhost:8080/users/create-map \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "name=FormUser&email=form@example.com&age=28" | jq '.'

echo "✅ All tests completed!"
```

Save this as `test_all.sh` and run:
```bash
chmod +x test_all.sh
./test_all.sh
```

## 📋 Individual Endpoint Examples

### 1. Health Check
```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "request-binding-examples"
}
```

### 2. Manual Binding (GET)
```bash
curl "http://localhost:8080/users/user123?page=2&limit=10&tags=web&tags=api&active=true" \
  -H "Authorization: Bearer token123" \
  -H "User-Agent: TestClient/1.0"
```

**Query Parameters:**
- `page=2` - Integer page number
- `limit=10` - Integer limit
- `tags=web&tags=api` - Array of strings
- `active=true` - Boolean value

**Headers:**
- `Authorization: Bearer token123`
- `User-Agent: TestClient/1.0`

**Expected Response:**
```json
{
  "method": "manual_binding",
  "data": {
    "id": "user123",
    "page": 2,
    "limit": 10,
    "tags": ["web", "api"],
    "active": true,
    "authorization": "Bearer token123",
    "user_agent": "TestClient/1.0"
  },
  "message": "Successfully bound using manual step-by-step approach"
}
```

### 3. Smart Binding (POST)
```bash
curl -X POST "http://localhost:8080/users/user456/smart?page=1&limit=5&tags=premium&active=false" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer smarttoken" \
  -H "User-Agent: SmartClient/2.0" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30,
    "preferences": {
      "theme": "dark",
      "language": "en",
      "notifications": true
    }
  }'
```

**Combines all parameter types:**
- Path: `user456`
- Query: `page=1&limit=5&tags=premium&active=false`
- Headers: `Authorization`, `User-Agent`
- Body: JSON with nested objects

### 4. BindBodySmart to Map
```bash
curl -X POST http://localhost:8080/users/create-map \
  -H "Content-Type: application/json" \
  -d '{
    "user_profile": {
      "name": "Alice Johnson",
      "email": "alice@example.com",
      "department": "Engineering"
    },
    "permissions": ["read", "write", "admin"],
    "settings": {
      "theme": "light",
      "language": "en-US",
      "timezone": "America/New_York"
    },
    "metadata": {
      "created_by": "system",
      "version": 1,
      "active": true
    }
  }'
```

**Dynamic JSON structure - any valid JSON will work**

### 5. Form Data Example
```bash
curl -X POST http://localhost:8080/users/create-map \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "name=FormUser&email=form@example.com&age=28&active=true&tags=test&tags=form"
```

### 6. Complex Query Parameters
```bash
curl "http://localhost:8080/search?q=lokstra%20framework&filter=type:web&filter=lang:go&filter=category:backend&sort=relevance&page=1&limit=20&opt[format]=json&opt[include]=docs&opt[highlight]=true&date=2023-01-01&date=2023-12-31"
```

**Complex query features:**
- URL encoding: `%20` for spaces
- Multiple filters: `filter=type:web&filter=lang:go`
- Nested parameters: `opt[format]=json`
- Date ranges: `date=start&date=end`

### 7. Error Testing

**Invalid JSON:**
```bash
curl -X POST http://localhost:8080/users/create-map \
  -H "Content-Type: application/json" \
  -d '{invalid json here'
```

**Missing Path Parameter:**
```bash
curl http://localhost:8080/users/
```

**Invalid Query Parameter Type:**
```bash
curl "http://localhost:8080/users/test123?page=not_a_number"
```

## 🧪 Testing with Different Tools

### Using HTTPie
```bash
# Install: pip install httpie

# GET with query params
http GET localhost:8080/users/user123 page==2 limit==10 tags==web tags==api active==true Authorization:"Bearer token123"

# POST with JSON
http POST localhost:8080/users/user456/smart page==1 name="John Doe" email=john@example.com age:=30

# Form data
http --form POST localhost:8080/users/create-map name=FormUser email=form@example.com age=28
```

### Using Postman

**Import Collection:**
Create a Postman collection with these requests:

```json
{
  "info": {
    "name": "Lokstra Binding Examples",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "http://localhost:8080/health",
          "protocol": "http",
          "host": ["localhost"],
          "port": "8080",
          "path": ["health"]
        }
      }
    },
    {
      "name": "Manual Binding",
      "request": {
        "method": "GET",
        "header": [
          {
            "key": "Authorization",
            "value": "Bearer token123"
          },
          {
            "key": "User-Agent",
            "value": "TestClient/1.0"
          }
        ],
        "url": {
          "raw": "http://localhost:8080/users/user123?page=2&limit=10&tags=web&tags=api&active=true",
          "protocol": "http",
          "host": ["localhost"],
          "port": "8080",
          "path": ["users", "user123"],
          "query": [
            {"key": "page", "value": "2"},
            {"key": "limit", "value": "10"},
            {"key": "tags", "value": "web"},
            {"key": "tags", "value": "api"},
            {"key": "active", "value": "true"}
          ]
        }
      }
    }
  ]
}
```

### Using VS Code REST Client

Create a file `requests.http`:

```http
### Health Check
GET http://localhost:8080/health

### Manual Binding
GET http://localhost:8080/users/user123?page=2&limit=10&tags=web&tags=api&active=true
Authorization: Bearer token123
User-Agent: TestClient/1.0

### Smart Binding
POST http://localhost:8080/users/user456/smart?page=1&limit=5&tags=premium&active=false
Content-Type: application/json
Authorization: Bearer smarttoken
User-Agent: SmartClient/2.0

{
  "name": "John Doe",
  "email": "john@example.com",
  "age": 30,
  "preferences": {
    "theme": "dark",
    "language": "en",
    "notifications": true
  }
}

### BindBodySmart to Map
POST http://localhost:8080/users/create-map
Content-Type: application/json

{
  "dynamic_field": "value",
  "nested": {
    "key": "value"
  },
  "array": ["item1", "item2"]
}

### Form Data
POST http://localhost:8080/users/create-map
Content-Type: application/x-www-form-urlencoded

name=FormUser&email=form@example.com&age=28
```

## 📊 Load Testing

Use these examples for load testing with tools like `ab`, `wrk`, or `hey`:

```bash
# Apache Bench
ab -n 1000 -c 10 http://localhost:8080/health

# wrk
wrk -t12 -c400 -d30s http://localhost:8080/health

# hey
hey -n 1000 -c 50 http://localhost:8080/health
```

**POST request with body:**
```bash
# Create a JSON file for POST testing
echo '{"name":"LoadTest","email":"load@test.com","age":25}' > test_payload.json

# Test with wrk
wrk -t4 -c100 -d30s -s post.lua --latency http://localhost:8080/users/create-map
```

**wrk POST script (post.lua):**
```lua
wrk.method = "POST"
wrk.body   = '{"name":"LoadTest","email":"load@test.com","age":25}'
wrk.headers["Content-Type"] = "application/json"
```

## 🔧 Environment Variables

You can override default settings:

```bash
# Different port
PORT=9090 go run main.go

# Different host
HOST=0.0.0.0 go run main.go

# Debug mode
DEBUG=true go run main.go
```

---

**Note**: Make sure the server is running (`go run main.go`) before executing these requests!