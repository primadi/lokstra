# Hello World Example

> **Your first Lokstra application - the simplest possible API**

Related to: [Quick Start Guide](../../quick-start.md)

---

## 📖 What This Example Shows

- ✅ Creating a router
- ✅ Registering simple GET routes
- ✅ Returning strings and maps (auto JSON)
- ✅ Running the app

---

## 🚀 Run the Example

```bash
# From this directory
go run main.go
```

Server will start on `http://localhost:3000`

---

## 🧪 Test the Endpoints

### Option 1: Using test.http (VS Code REST Client)

Open `test.http` in VS Code and click "Send Request" above each request.

### Option 2: Using curl

```bash
# Hello endpoint
curl http://localhost:3000/

# Ping endpoint
curl http://localhost:3000/ping

# Time endpoint (returns JSON)
curl http://localhost:3000/time
```

---

## 📝 Expected Results

**GET /**
```
{
  "status": "success",
  "data": "Hello, Lokstra!"
}
```

**GET /ping**
```
{
  "status": "success",
  "data": "pong"
}
```

**GET /time**
```json
{
  "status": "success",
  "data": {
    "datetime": "2025-10-15T02:22:27+07:00",
    "timestamp": 1760469747
  }
}
```

---

## 💡 Key Concepts

### 1. Router Creation
```go
r := lokstra.NewRouter("api")
```
Creates a new router named "api"

### 2. Simple Handler Forms
```go
// Form 1: Return string
r.GET("/", func() string {
    return "Hello, Lokstra!"
})

// Form 2: Return map (auto converts to JSON)
r.GET("/time", func() map[string]any {
    return map[string]any{
        "timestamp": time.Now().Unix(),
    }
})
```

### 3. Running the App
```go
app := lokstra.NewApp("hello", ":3000", r)
app.Run(30 * time.Second)  // 30s graceful shutdown timeout
```

---

## 🔍 What's Next?

Try modifying:
- Add more routes
- Return different data types
- Change the port

See more examples:
- [Handler Forms](../02-handler-forms/) - All 29 handler variations
- [CRUD API](../03-crud-api/) - Full REST API with services
- [Multi-Deployment](../04-multi-deployment/) - Monolith vs Microservices

---

**Questions?** Check the [Quick Start Guide](../../quick-start.md)
