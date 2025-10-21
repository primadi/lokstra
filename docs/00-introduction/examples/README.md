# Lokstra Examples

> 🎯 **Progressive learning path: Manual basics → Auto-router patterns**

Learn Lokstra step by step, from manual router creation to automated service-to-router generation.

---

## 📂 Examples

### [01-hello-world](./01-hello-world/)
**Your first Lokstra API**

- Simple router with GET handlers
- Auto JSON responses
- Basic string and map returns

```bash
cd 01-hello-world && go run main.go
curl http://localhost:3000/
```

**Why manual?** Perfect for quick prototyping and learning basics!

---

### [02-handler-forms](./02-handler-forms/)
**Explore 29 handler variations**

- Request binding (JSON, path, query, header)
- Response forms (string, map, struct, error handling)
- Context access patterns

```bash
cd 02-handler-forms && go run main.go
```

**Why manual?** Understanding handler flexibility is fundamental!

---

### [03-crud-api](./03-crud-api/)
**Full CRUD with service pattern**

- Service-based architecture
- Dependency injection
- Manual router registration

```bash
cd 03-crud-api && go run main.go
curl http://localhost:3000/users
```

**Features:**
- ✅ Service factories
- ✅ Lazy dependency injection
- ✅ Clean separation of concerns
- ✅ Manual route registration (understand the foundation!)

---

### [04-multi-deployment](./04-multi-deployment/)
**One binary, multiple deployments**

- Monolith vs Microservices
- Service interface pattern (local vs remote)
- Cross-service communication

```bash
# Run as monolith
go run . -server=monolith

# Run as microservices
go run . -server=user-service    # Terminal 1
go run . -server=order-service   # Terminal 2
```

**Key Learning:**
- Manual router for each deployment
- Interface abstraction (UserService local vs remote)
- Proxy pattern for remote calls

---

### [05-auto-router-proxy](./05-auto-router-proxy/) ⭐ **NEW!**
**Automatic router generation + Convention-based proxy**

This is where automation begins!

- ✅ `autogen.NewFromService()` - Auto-generate REST routes
- ✅ `proxy.Service` - Convention-based remote calls
- ✅ Zero boilerplate routing

```bash
# Terminal 1: User service with auto-router
go run . -mode=server

# Terminal 2: Order service with proxy
go run . -mode=client
```

**What's automated:**
- Router generation from service methods
- URL construction from conventions
- HTTP calls with `proxy.Call()`

**Comparison with Example 04:**
- Example 04: Manual `r.GET("/users", ...)` for each endpoint
- Example 05: `autogen.NewFromService()` generates all routes automatically

---

## � Learning Progression

### Phase 1: Manual Foundation (Examples 01-04)
**Learn the fundamentals** - How things work under the hood

| Example | Focus | Why Manual? |
|---------|-------|-------------|
| **01** | Basic routing | Understand router creation |
| **02** | Handler forms | Learn request/response patterns |
| **03** | Services & DI | Grasp service architecture |
| **04** | Deployments | Master interface abstraction |

**Benefits of learning manual first:**
- ✅ Deep understanding of Lokstra internals
- ✅ Better debugging when things go wrong
- ✅ Flexibility for custom scenarios
- ✅ Appreciation for what automation provides

### Phase 2: Automation (Example 05+)
**Leverage the framework** - Let Lokstra do the heavy lifting

| Example | Automation | Benefit |
|---------|------------|---------|
| **05** | Auto-router + Proxy | Zero boilerplate routing |

**When automation makes sense:**
- ✅ Service-based architectures (5+ endpoints per service)
- ✅ RESTful conventions (standard CRUD patterns)
- ✅ Microservices (multiple services to wire up)
- ✅ Consistency requirements (same patterns everywhere)

**When to stay manual:**
- ❌ Quick prototypes (< 10 endpoints total)
- ❌ Non-standard routing (custom URL patterns)
- ❌ Fine-grained control needed
- ❌ Learning/debugging

---

## 🔄 Manual vs Auto-Router Comparison

### Example 04 (Manual Router)
```go
// For each service method, manually register route
r.GET("/users", userHandler.list)
r.GET("/users/{id}", userHandler.get)
r.POST("/users", userHandler.create)
r.PUT("/users/{id}", userHandler.update)

// Repeat for order service...
r.GET("/orders/{id}", orderHandler.get)
r.GET("/users/{user_id}/orders", orderHandler.getUserOrders)
```

**Lines of code:** ~20 per service (routing only)

### Example 05 (Auto-Router)
```go
// Define convention once
conversionRule := autogen.ConversionRule{
    Convention:     convention.REST,
    Resource:       "user",
    ResourcePlural: "users",
}

// Generate all routes automatically
router := autogen.NewFromService(userService, conversionRule, routerOverride)
```

**Lines of code:** ~8 per service (routing only)  
**Savings:** 60-70% less boilerplate!

---

## 🎯 Choosing the Right Approach

### Use Manual Router When:
- ✅ Learning Lokstra fundamentals
- ✅ Building small APIs (< 10 endpoints)
- ✅ Need custom route patterns
- ✅ Prototyping quickly
- ✅ Non-service based handlers (standalone functions)

### Use Auto-Router When:
- ✅ Service-based architecture
- ✅ RESTful conventions
- ✅ Large APIs (10+ endpoints per service)
- ✅ Microservices
- ✅ Consistency is critical
- ✅ Rapid development

**Pro tip:** Start manual (Examples 01-04), then adopt auto-router (Example 05) when complexity justifies it!

---

## 🚀 Running Examples

```bash
# Navigate to any example
cd 01-hello-world  # or 02, 03, 04, 05

# Run it
go run main.go

# Test it (use test.http or curl from README)
curl http://localhost:3000/
```

**For multi-server examples:**

Example 04 (Manual):
```bash
cd 04-multi-deployment

# Option 1: Monolith
go run . -server=monolith

# Option 2: Microservices (2 terminals)
go run . -server=user-service     # Terminal 1
go run . -server=order-service    # Terminal 2
```

Example 05 (Auto-Router):
```bash
cd 05-auto-router-proxy

# Terminal 1: Server
go run . -mode=server

# Terminal 2: Client
go run . -mode=client
```

---

## 🎓 Recommended Learning Path

### Week 1: Foundations
1. **Day 1-2:** Example 01 - Hello World
2. **Day 3-4:** Example 02 - Handler Forms
3. **Day 5-7:** Example 03 - CRUD API

**Goal:** Understand manual router, handlers, and services

### Week 2: Advanced Patterns
1. **Day 1-3:** Example 04 - Multi-deployment (deep dive!)
2. **Day 4-5:** Example 05 - Auto-Router & Proxy
3. **Day 6-7:** Build your own project using learned patterns

**Goal:** Master deployment patterns and automation

---

## 💡 Key Takeaways

1. **Manual First**: Examples 01-04 teach fundamentals - don't skip them!
2. **Auto-Router is Optional**: Use it when complexity justifies automation
3. **Both Approaches Valid**: Manual for control, auto for consistency
4. **Progressive Enhancement**: Start simple, add automation as needed

---
