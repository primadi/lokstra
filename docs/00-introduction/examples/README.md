# Lokstra Examples# Lokstra Examples



> 🎯 **Progressive learning path: Manual basics → Production patterns**> 🎯 **Progressive learning path: Manual basics → Production patterns**



Learn Lokstra step by step, from basic routing to production-ready middleware and architecture.Learn Lokstra step by step, from basic routing to production-ready middleware and architecture.



---



## 📚 Learning Path---



```## 📚 Learning Path

01-hello-world

    ↓ Learn: Router basics, simple handlers```

02-handler-forms01-hello-world

    ↓ Learn: 29 handler variations, request/response patterns    ↓ Learn: Router basics, simple handlers

03-crud-api02-handler-forms

    ↓ Learn: Services, dependency injection, manual routing    ↓ Learn: 29 handler variations, request/response patterns

04-multi-deployment03-crud-api

    ↓ Learn: Clean Architecture, auto-router, microservices    ↓ Learn: Services, dependency injection, manual routing

05-middleware04-multi-deployment

    ↓ Learn: Global/route middleware, auth, recovery, rate limiting    ↓ Learn: Clean Architecture, auto-router, microservices

```05-middleware

    ↓ Learn: Global/route middleware, auth, recovery, rate limiting

**Time investment**: ~6-8 hours to complete all examples  ```

**Outcome**: Ready to build production REST APIs with Lokstra

**Time investment**: ~6-8 hours to complete all examples  

---**Outcome**: Ready to build production REST APIs with Lokstra



## 📂 Examples---



### [01-hello-world](./01-hello-world/)## � Examples



**Your first Lokstra API**### [01-hello-world](./01-hello-world/)



- Simple router with GET handlers**Your first Lokstra API**

- Auto JSON responses

- Basic string and map returns- Simple router with GET handlers

- Auto JSON responses

```bash- Basic string and map returns

cd 01-hello-world && go run main.go

curl http://localhost:3000/```bash

```cd 01-hello-world && go run main.go

curl http://localhost:3000/

**Why manual?** Perfect for quick prototyping and learning basics!```



---**Why manual?** Perfect for quick prototyping and learning basics!



### [02-handler-forms](./02-handler-forms/)---



**Explore 29 handler variations**### [02-handler-forms](./02-handler-forms/)



- Request binding (JSON, path, query, header)**Explore 29 handler variations**

- Response forms (string, map, struct, error handling)

- Context access patterns- Request binding (JSON, path, query, header)

- Response forms (string, map, struct, error handling)

```bash- Context access patterns

cd 02-handler-forms && go run main.go

``````bash

cd 02-handler-forms && go run main.go

**Why manual?** Understanding handler flexibility is fundamental!```



---**Why manual?** Understanding handler flexibility is fundamental!



### [03-crud-api](./03-crud-api/)---



**Full CRUD with service pattern**### [03-crud-api](./03-crud-api/)



- Service-based architecture**Full CRUD with service pattern**

- Dependency injection

- Manual router registration- Service-based architecture

- Dependency injection

```bash- Manual router registration

cd 03-crud-api && go run main.go

curl http://localhost:3000/users```bash

```cd 03-crud-api && go run main.go

curl http://localhost:3000/users

**Features:**```

- ✅ Service factories

- ✅ Lazy dependency injection**Features:**

- ✅ Clean separation of concerns- ✅ Service factories

- ✅ Manual route registration (understand the foundation!)- ✅ Lazy dependency injection

- ✅ Clean separation of concerns

---- ✅ Manual route registration (understand the foundation!)



### [04-multi-deployment](./04-multi-deployment/)---



**One binary, multiple deployments**### [04-multi-deployment](./04-multi-deployment/)



- Monolith vs Microservices**One binary, multiple deployments**

- Service interface pattern (local vs remote)

- Cross-service communication- Monolith vs Microservices

- Service interface pattern (local vs remote)

```bash- Cross-service communication

# Run as monolith

go run . -server=monolith```bash

# Run as monolith

# Run as microservicesgo run . -server=monolith

go run . -server=user-service    # Terminal 1

go run . -server=order-service   # Terminal 2# Run as microservices

```go run . -server=user-service    # Terminal 1

go run . -server=order-service   # Terminal 2

**Key Learning:**```

- Manual router for each deployment

- Interface abstraction (UserService local vs remote)**Key Learning:**

- Proxy pattern for remote calls- Manual router for each deployment

- Interface abstraction (UserService local vs remote)

---- Proxy pattern for remote calls



### [05-middleware](./05-middleware/) ⭐ **NEW!**---



**Global and route-specific middleware**### [05-middleware](./05-middleware/) ⭐ **NEW!**



This is where you learn production-ready request handling!**Global and route-specific middleware**



- ✅ Global middleware (applied to all routes)This is where you learn production-ready request handling!

- ✅ Route-specific middleware (per-endpoint auth)

- ✅ Custom middleware creation- ✅ Global middleware (applied to all routes)

- ✅ Built-in middleware (CORS, Recovery, Logger)- ✅ Route-specific middleware (per-endpoint auth)

- ✅ Middleware chaining and execution order- ✅ Custom middleware creation

- ✅ Built-in middleware (CORS, Recovery, Logger)

```bash- ✅ Middleware chaining and execution order

cd 05-middleware

go run main.go```bash

cd 05-middleware

# Test with different scenariosgo run main.go

curl http://localhost:3000/                           # Public

curl http://localhost:3000/protected -H "X-API-Key: secret-key-123"  # Auth required# Test with different scenarios

curl http://localhost:3000/api/admin/dashboard -H "X-API-Key: admin-key-456"  # Admin onlycurl http://localhost:3000/                           # Public

curl http://localhost:3000/panic                      # Recovery middlewarecurl http://localhost:3000/protected -H "X-API-Key: secret-key-123"  # Auth required

```curl http://localhost:3000/api/admin/dashboard -H "X-API-Key: admin-key-456"  # Admin only

curl http://localhost:3000/panic                      # Recovery middleware

**What you'll learn:**```

- ✅ **Global middleware**: Recovery, CORS, Logger, Rate Limiting

- ✅ **Auth middleware**: API key validation**What you'll learn:**

- ✅ **Role-based access**: Admin-only endpoints- ✅ **Global middleware**: Recovery, CORS, Logger, Rate Limiting

- ✅ **Custom middleware**: LoggingMiddleware, RateLimitMiddleware- ✅ **Auth middleware**: API key validation

- ✅ **Middleware chain**: Multiple middleware per route- ✅ **Role-based access**: Admin-only endpoints

- ✅ **Override parent**: Route with `WithOverrideParentMwOption(true)`- ✅ **Custom middleware**: LoggingMiddleware, RateLimitMiddleware

- ✅ **Middleware chain**: Multiple middleware per route

**Production patterns covered:**- ✅ **Override parent**: Route with `WithOverrideParentMwOption(true)`

- Panic recovery (graceful error handling)

- Request logging with timing**Production patterns covered:**

- Rate limiting per IP- Panic recovery (graceful error handling)

- Authentication & Authorization- Request logging with timing

- CORS for API access- Rate limiting per IP

- Authentication & Authorization

**Code size**: ~180 lines  - CORS for API access

**Endpoints**: 11 routes with various middleware combinations

**Code size**: ~180 lines  

**This is essential for production!** 🚀**Endpoints**: 11 routes with various middleware combinations



---**This is essential for production!** 🚀



## 🎯 Learning Objectives by Example---



| Example | Router | Handlers | Services | DI | Middleware | Clean Arch | Microservices |**Code size**: ~30 lines```

|---------|--------|----------|----------|----|-----------||---------------|---------------|

| **01** | ✅ Basic | ✅ Simple | ❌ | ❌ | ❌ | ❌ | ❌ |

| **02** | ✅ Routes | ✅ 29 forms | ❌ | ❌ | ❌ | ❌ | ❌ |

| **03** | ✅ Manual | ✅ Service | ✅ Yes | ✅ Yes | ❌ | ❌ | ❌ |---**Features:**

| **04** | ✅ Auto | ✅ Advanced | ✅ Layered | ✅ Interface | ❌ | ✅ Yes | ✅ Yes |

| **05** | ✅ Manual | ✅ Full | ✅ Yes | ✅ Yes | ✅ Production | ❌ | ❌ |- ✅ Service factories



---### [02 - Handler Forms](./02-handler-forms/)- ✅ Lazy dependency injection



## 🔄 Recommended Learning Strategy⏱️ **30 minutes** • 🎯 **Beginner**- ✅ Clean separation of concerns



### Week 1: Foundations (4-5 hours)- ✅ Manual route registration (understand the foundation!)

- **Day 1**: Example 01 (15min) + Example 02 (30min)

- **Day 2**: Example 03 (1 hour)**Explore Lokstra's 29 handler variations**

- **Day 3**: Example 05 - Middleware (1-2 hours)

- **Day 4**: Review and build small API with middleware---



**Goal**: Understand basics, middleware patterns, write first protected APIUnderstand handler flexibility:



### Week 2: Production Patterns (6-8 hours)- Request binding (JSON, path params, query, headers)### [04-multi-deployment](./04-multi-deployment/)

- **Day 1-2**: Example 04 (read, understand, run all modes)

- **Day 3-5**: Build your project using examples as template- Response types (string, struct, error handling)**One binary, multiple deployments**

- **Weekend**: Refine and deploy

- Context access patterns

**Goal**: Master production-ready architecture

- Different parameter combinations- Monolith vs Microservices

---

- Service interface pattern (local vs remote)

## 💡 Key Progression

```bash- Cross-service communication

### Example 01 → Router Basics

```gocd 02-handler-forms

r := lokstra.NewRouter("api")

r.GET("/ping", func() string { return "pong" })go run main.go```bash

```

# Check test.http for all variations# Run as monolith

### Example 02 → Handler Flexibility

```go```go run . -server=monolith

r.GET("/users/{id}", func(p *GetUserParams) (*User, error) {

    return db.GetUser(p.ID)

})

```**What you'll learn:**# Run as microservices



### Example 03 → Service Pattern- ✅ Path parameters: `{id}` → `struct { ID int \`path:"id"\` }`go run . -server=user-service    # Terminal 1

```go

type UserService struct {- ✅ JSON bodies: `func(user *User) error`go run . -server=order-service   # Terminal 2

    DB *service.Cached[*Database]

}- ✅ Query params: `struct { Page int \`query:"page"\` }````



r.GET("/users", func() ([]*User, error) {- ✅ Error handling: Return `(data, error)`

    return userService.List()

})- ✅ Full control: `*request.Context` access**Key Learning:**

```

- Manual router for each deployment

### Example 04 → Auto-Router + Clean Architecture

```go**Handler forms covered**: 29 variations  - Interface abstraction (UserService local vs remote)

// Just define the service interface and implementation

// Routes auto-generated from metadata!**Code size**: ~150 lines- Proxy pattern for remote calls

// GetByID() → GET /users/{id}

// List()    → GET /users

```

**Key takeaway**: Lokstra adapts to YOUR code style!---

### Example 05 → Production Middleware

```go

// Global middleware

r.Use(RecoveryMiddleware)---### [05-auto-router-proxy](./05-auto-router-proxy/) ⭐ **NEW!**

r.Use(CORSMiddleware)

r.Use(LoggerMiddleware)**Automatic router generation + Convention-based proxy**



// Route-specific auth### [03 - CRUD API](./03-crud-api/)

r.GET("/protected", ProtectedHandler, AuthMiddleware)

r.GET("/admin", AdminHandler, AuthMiddleware, AdminOnlyMiddleware)⏱️ **1 hour** • 🎯 **Intermediate**This is where automation begins!

```



---

**Full CRUD with service pattern**- ✅ `autogen.NewFromService()` - Auto-generate REST routes

## 🚀 Running Examples

- ✅ `proxy.Service` - Convention-based remote calls

```bash

# Navigate to any exampleBuild a real API with proper architecture:- ✅ Zero boilerplate routing

cd 01-hello-world  # or 02, 03, 04, 05

- Service layer for business logic

# Run it

go run main.go- Dependency injection```bash



# Test it (use test.http or curl from README)- In-memory database# Terminal 1: User service with auto-router

curl http://localhost:3000/

```- RESTful conventionsgo run . -mode=server



**For multi-server examples:**- Error handling



Example 04:# Terminal 2: Order service with proxy

```bash

cd 04-multi-deployment```bashgo run . -mode=client



# Option 1: Monolithcd 03-crud-api```

go run . -server=monolith

go run main.go

# Option 2: Microservices (2 terminals)

go run . -server=user-service     # Terminal 1curl http://localhost:3000/users**What's automated:**

go run . -server=order-service    # Terminal 2

``````- Router generation from service methods



---- URL construction from conventions



## 📚 Next Steps**What you'll learn:**- HTTP calls with `proxy.Call()`



After completing these examples:- ✅ **Service pattern**: Separate business logic from HTTP



- **Deep Dive**: [01-essentials](../../01-essentials/README.md)- ✅ **Dependency Injection**: `service.Cached[T]` for lazy loading**Comparison with Example 04:**

- **API Reference**: [03-api-reference](../../03-api-reference/README.md) (coming soon)

- **Advanced Topics**: [02-deep-dive](../../02-deep-dive/README.md) (coming soon)- ✅ **Factory pattern**: Register services with `lokstra_registry`- Example 04: Manual `r.GET("/users", ...)` for each endpoint



---- ✅ **Manual routing**: Explicit route registration- Example 05: `autogen.NewFromService()` generates all routes automatically



**Start here**: → [01-hello-world](./01-hello-world/) 🚀- ✅ **CRUD operations**: Complete Create/Read/Update/Delete


---

**Architecture**:

```## � Learning Progression

main.go → Router → Handlers → UserService → Database

```### Phase 1: Manual Foundation (Examples 01-04)

**Learn the fundamentals** - How things work under the hood

**Code size**: ~200 lines  

**Endpoints**: 5 (List, GetByID, Create, Update, Delete)| Example | Focus | Why Manual? |

|---------|-------|-------------|

**Why manual routing?** Learn the foundation before automation!| **01** | Basic routing | Understand router creation |

| **02** | Handler forms | Learn request/response patterns |

---| **03** | Services & DI | Grasp service architecture |

| **04** | Deployments | Master interface abstraction |

### [04 - Multi-Deployment](./04-multi-deployment/) ⭐

⏱️ **2-3 hours** • 🎯 **Advanced****Benefits of learning manual first:**

- ✅ Deep understanding of Lokstra internals

**Production-ready architecture with Clean Architecture**- ✅ Better debugging when things go wrong

- ✅ Flexibility for custom scenarios

The complete package - everything you need for real applications:- ✅ Appreciation for what automation provides

- **Clean Architecture** (contract/model/service/repository layers)

- **Auto-router generation** from service metadata### Phase 2: Automation (Example 05+)

- **Convention-based routing** (REST)**Leverage the framework** - Let Lokstra do the heavy lifting

- **Microservices support** (local vs remote services)

- **Single binary, multiple deployments**| Example | Automation | Benefit |

- **Interface-based DI** for testability|---------|------------|---------|

| **05** | Auto-router + Proxy | Zero boilerplate routing |

```bash

cd 04-multi-deployment**When automation makes sense:**

- ✅ Service-based architectures (5+ endpoints per service)

# Option 1: Monolith- ✅ RESTful conventions (standard CRUD patterns)

go run . -server "monolith.api-server"- ✅ Microservices (multiple services to wire up)

- ✅ Consistency requirements (same patterns everywhere)

# Option 2: Microservices

go run . -server "microservice.user-server"    # Terminal 1**When to stay manual:**

go run . -server "microservice.order-server"   # Terminal 2- ❌ Quick prototypes (< 10 endpoints total)

```- ❌ Non-standard routing (custom URL patterns)

- ❌ Fine-grained control needed

**What you'll learn:**- ❌ Learning/debugging

- ✅ **Clean Architecture** layers

- ✅ **Auto-router**: Zero boilerplate routing---

- ✅ **Convention-based proxy**: Remote calls without manual HTTP

- ✅ **Metadata-driven routing**: Single source of truth## 🔄 Manual vs Auto-Router Comparison

- ✅ **Deployment flexibility**: Monolith ↔ Microservices

### Example 04 (Manual Router)

**Code size**: ~600 lines  ```go

**Deployments**: 2 modes  // For each service method, manually register route

r.GET("/users", userHandler.list)

**This is the production pattern!** 🚀r.GET("/users/{id}", userHandler.get)

r.POST("/users", userHandler.create)

[See full documentation](./04-multi-deployment/README.md)r.PUT("/users/{id}", userHandler.update)



---// Repeat for order service...

r.GET("/orders/{id}", orderHandler.get)

## 🎯 Learning Objectives by Exampler.GET("/users/{user_id}/orders", orderHandler.getUserOrders)

```

| Example | Router | Handlers | Services | DI | Auto-Router | Clean Arch | Microservices |

|---------|--------|----------|----------|----|-----------||---------------|---------------|**Lines of code:** ~20 per service (routing only)

| **01** | ✅ Basic | ✅ Simple | ❌ | ❌ | ❌ | ❌ | ❌ |

| **02** | ✅ Routes | ✅ 29 forms | ❌ | ❌ | ❌ | ❌ | ❌ |### Example 05 (Auto-Router)

| **03** | ✅ Manual | ✅ Service | ✅ Yes | ✅ Yes | ❌ | ❌ | ❌ |```go

| **04** | ✅ Auto | ✅ Advanced | ✅ Layered | ✅ Interface | ✅ Yes | ✅ Yes | ✅ Yes |// Define convention once

conversionRule := autogen.ConversionRule{

---    Convention:     convention.REST,

    Resource:       "user",

## 🔄 Recommended Learning Strategy    ResourcePlural: "users",

}

### Week 1: Foundations (4-5 hours)

- **Day 1**: Example 01 (15min) + Example 02 (30min)// Generate all routes automatically

- **Day 2**: Example 03 (1 hour)router := autogen.NewFromService(userService, conversionRule, routerOverride)

- **Day 3**: Review and build small API```



**Goal**: Understand basics, write first API**Lines of code:** ~8 per service (routing only)  

**Savings:** 60-70% less boilerplate!

### Week 2: Production Patterns (6-8 hours)

- **Day 1-2**: Example 04 (read, understand, run all modes)---

- **Day 3-5**: Build your project using example 04 as template

- **Weekend**: Refine and deploy## 🎯 Choosing the Right Approach



**Goal**: Master production-ready architecture### Use Manual Router When:

- ✅ Learning Lokstra fundamentals

---- ✅ Building small APIs (< 10 endpoints)

- ✅ Need custom route patterns

## 💡 Key Progression- ✅ Prototyping quickly

- ✅ Non-service based handlers (standalone functions)

### Example 01 → Router Basics

```go### Use Auto-Router When:

r := lokstra.NewRouter("api")- ✅ Service-based architecture

r.GET("/ping", func() string { return "pong" })- ✅ RESTful conventions

```- ✅ Large APIs (10+ endpoints per service)

- ✅ Microservices

### Example 02 → Handler Flexibility- ✅ Consistency is critical

```go- ✅ Rapid development

r.GET("/users/{id}", func(p *GetUserParams) (*User, error) {

    return db.GetUser(p.ID)**Pro tip:** Start manual (Examples 01-04), then adopt auto-router (Example 05) when complexity justifies it!

})

```---



### Example 03 → Service Pattern## 🚀 Running Examples

```go

type UserService struct {```bash

    DB *service.Cached[*Database]# Navigate to any example

}cd 01-hello-world  # or 02, 03, 04, 05



r.GET("/users", func() ([]*User, error) {# Run it

    return userService.List()go run main.go

})

```# Test it (use test.http or curl from README)

curl http://localhost:3000/

### Example 04 → Auto-Router + Clean Architecture```

```go

// Just define the service interface and implementation**For multi-server examples:**

// Routes auto-generated from metadata!

// GetByID() → GET /users/{id}Example 04 (Manual):

// List()    → GET /users```bash

```cd 04-multi-deployment



---# Option 1: Monolith

go run . -server=monolith

## 🚀 Quick Start

# Option 2: Microservices (2 terminals)

```bashgo run . -server=user-service     # Terminal 1

# Clone and navigatego run . -server=order-service    # Terminal 2

cd docs/00-introduction/examples```



# Try each exampleExample 05 (Auto-Router):

cd 01-hello-world && go run main.go```bash

cd ../02-handler-forms && go run main.gocd 05-auto-router-proxy

cd ../03-crud-api && go run main.go

cd ../04-multi-deployment && go run . -server "monolith.api-server"# Terminal 1: Server

```go run . -mode=server



---# Terminal 2: Client

go run . -mode=client

## 📚 Next Steps```



After completing these examples:---



- **Deep Dive**: [01-essentials](../../01-essentials/README.md)## 🎓 Recommended Learning Path

- **API Reference**: [03-api-reference](../../03-api-reference/README.md) (coming soon)

- **Advanced Topics**: [02-deep-dive](../../02-deep-dive/README.md) (coming soon)### Week 1: Foundations

1. **Day 1-2:** Example 01 - Hello World

---2. **Day 3-4:** Example 02 - Handler Forms

3. **Day 5-7:** Example 03 - CRUD API

**Start here**: → [01-hello-world](./01-hello-world/) 🚀

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
