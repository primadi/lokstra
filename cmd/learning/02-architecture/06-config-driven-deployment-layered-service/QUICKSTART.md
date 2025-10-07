# Quick Start Guide

## Installation

```bash
cd cmd/learning/02-architecture/06-config-driven-deployment-layered-service
```

## Run Application

### Option 1: Layered Services (Recommended - New Pattern)
```bash
go run . layered
```

**Output:**
```
📚 Lokstra Learning: 06-Layered Services Comparison
=====================================================

🎯 Mode: layered
📄 Config: config-layered.yaml

📋 Service Mode: LAYERED (grouped by layer)
   Total layers: 4
   - infrastructure: 3 services
   - repository: 3 services
   - domain: 2 services
   - orchestration: 1 services
✅ Validation passed!

📊 Layered Services Pattern:
   ✅ Type-safe with Lazy[T]
   ✅ ~3 lines per dependency (80% less!)
   ✅ Explicit depends-on
   ✅ Automatic validation
   ✅ Architecture visible in config
   ✅ Auto-caching with sync.Once

🚀 Starting server...
```

### Option 2: Simple Services (Backward Compatible)
```bash
go run . simple
```

**Output:**
```
🎯 Mode: simple
📄 Config: config-simple.yaml

📋 Service Mode: SIMPLE (flat array)
   Total services: 9

📊 Simple Services Pattern:
   ✅ Backward compatible
   ✅ Familiar pattern
   ❌ ~15 lines boilerplate per dependency
   ❌ Manual lazy loading + caching
   ❌ Dependencies hidden in code
   ❌ No validation

🚀 Starting server...
```

## Test API Endpoints

Open `test.http` in VS Code with REST Client extension, or use curl:

### Health Check
```bash
curl http://localhost:8080/health
```

### List Products
```bash
curl http://localhost:8080/api/products
```

### Get Product
```bash
curl http://localhost:8080/api/products/1
```

### Create Order
```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123",
    "product_id": "1",
    "quantity": 2
  }'
```

**Response (Same for both modes!):**
```json
{
  "success": true,
  "message": "Order created successfully",
  "data": {
    "id": "order-1234567890",
    "user_id": "123",
    "product_id": "1",
    "quantity": 2,
    "subtotal": 2599.98,
    "tax": 259.998,
    "total": 2859.98
  }
}
```

## Console Output Comparison

### When Creating Order

**Both modes show:**
```
💾 [DB] Query: SELECT * FROM products WHERE id = 1
🗄️  [Cache] GET product:1
💾 [DB] Execute: INSERT INTO orders ... (args: [123 1 2 2859.98])
💾 [DB] Query: SELECT * FROM users WHERE id = 123
📧 [Email] To: user123@example.com, Subject: Order Confirmation
```

The difference is ONLY in the configuration and code structure, not in runtime behavior!

## Key Features Demonstrated

### 1. Architecture Layers

**Layered Config Shows:**
- 📦 **Layer 1: Infrastructure** - db, cache, email (3 services)
- 📚 **Layer 2: Repository** - user-repo, product-repo, order-repo (3 services)
- 🎯 **Layer 3: Domain** - user-service, product-service (2 services)
- 🔄 **Layer 4: Orchestration** - order-service (1 service)

**Simple Config:** Flat array (no visibility)

### 2. Dependency Validation

**Layered Config:**
```
✅ Validation passed!
```

Checks:
- All `depends-on` services exist
- Dependencies only reference previous layers
- All `depends-on` are used in config
- All config service refs are in `depends-on`

**Simple Config:** No validation

### 3. Code Reduction

**OrderService Factory:**
- Simple mode: **~60 lines**
- Layered mode: **~15 lines**
- **Reduction: 75%**

### 4. Type Safety

**Simple Mode:**
```go
func (s *OrderService) getRepo() *OrderRepository {
    s.repoCache = lokstra_registry.GetService(s.repoServiceName, s.repoCache)
    return s.repoCache  // Manual cast
}
```

**Layered Mode:**
```go
repo := s.repo.Get()  // Type-safe, no casts!
```

## Files Overview

| File | Purpose |
|------|---------|
| `README.md` | Complete documentation |
| `COMPARISON.md` | Side-by-side code comparison |
| `QUICKSTART.md` | This file - quick reference |
| `config-simple.yaml` | Simple services config (backward compatible) |
| `config-layered.yaml` | Layered services config (new pattern) |
| `main.go` | Application with both patterns |
| `test.http` | HTTP API tests |

## Next Steps

1. **Read `README.md`** - Full documentation with examples
2. **Read `COMPARISON.md`** - Detailed code comparison
3. **Try both modes** - See the difference yourself!
4. **Examine configs** - Compare simple vs layered YAML
5. **Look at factories** - See Lazy[T] pattern in action

## Migration Guide

### Step 1: Keep using simple services
Your existing configs work unchanged!

### Step 2: Experiment with layered services
Create a new config file with layers and `depends-on`.

### Step 3: Update factories gradually
Convert to `Lazy[T]` pattern one service at a time.

### Step 4: Enjoy the benefits
- 75% less boilerplate
- Type-safe dependencies
- Automatic validation
- Clear architecture

## Summary

Both patterns produce **IDENTICAL** API behavior:
- ✅ Same endpoints
- ✅ Same responses
- ✅ Same console output
- ✅ Same performance

The difference is in:
- 🎨 Configuration structure
- 📝 Code clarity
- ✅ Type safety
- 🛡️ Validation
- 📊 Architecture visibility

**Try both and see which you prefer!**
