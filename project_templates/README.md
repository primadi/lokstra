# Lokstra Project Templates

**Production-Ready Templates for Building Applications with Lokstra**

This folder contains a comprehensive collection of templates demonstrating different patterns and architectures for building applications with the Lokstra framework.

---

## 📂 Template Categories

### [01_router](./01_router/) - Router & Server Patterns

**Basic building blocks for Lokstra applications**

Learn routing, middleware, multi-app servers, and deployment basics.

| Template | Description | Best For |
|----------|-------------|----------|
| **01_router_only** | Pure router with CRUD operations | Understanding Lokstra routing |
| **02_single_app** | App wrapper with graceful shutdown | Single application servers |
| **03_multi_app** | Multiple apps on different ports | Admin/API separation |

**Start here if**: You're new to Lokstra and want to understand the basics.

---

### [02_app_framework](./02_app_framework/) - Domain-Driven Architecture

**Production-ready application frameworks**

Learn domain-driven design, clean architecture, and enterprise patterns.

| Template | Description | Best For |
|----------|-------------|----------|
| **01_medium_system** | Flat domain-driven structure | 2-10 entities, single team |
| **02_enterprise_modular** | DDD with bounded contexts | 10+ entities, multiple teams |

**Start here if**: You understand Lokstra basics and need production architecture.

---

## 🎯 Which Template Should I Use?

### Quick Decision Tree

```
1. Are you learning Lokstra basics?
   YES → Start with 01_router/01_router_only
   NO  → Go to question 2

2. Do you need domain-driven architecture?
   NO  → Use 01_router templates (simple routing)
   YES → Go to question 3

3. How many entities do you have?
   < 10 entities  → Use 02_app_framework/01_medium_system
   10+ entities   → Use 02_app_framework/02_enterprise_modular
```

---

## 📊 Template Comparison Matrix

| Feature | Router Only | Single App | Multi App | Medium System | Enterprise Modular |
|---------|------------|-----------|-----------|---------------|-------------------|
| **Complexity** | ⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Use Case** | Learning | Production | Multi-server | Domain apps | Enterprise |
| **Architecture** | None | Basic | Multi-app | Clean Arch | DDD |
| **Entity Count** | 1 | 1-5 | 1-5 | 2-10 | 10+ |
| **Team Size** | 1 | 1 | 1-2 | 1 | Multiple |
| **Middleware** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Domain Layer** | ❌ | ❌ | ❌ | ✅ | ✅ |
| **Bounded Contexts** | ❌ | ❌ | ❌ | ❌ | ✅ |
| **Microservices Ready** | ❌ | ❌ | ⚠️ | ⚠️ | ✅ |

**Legend**: ⭐ = Complexity level | ✅ = Supported | ❌ = Not included | ⚠️ = Possible but not optimized

---

## 🚀 Getting Started

### 1. Choose Your Path

#### Path A: Learning Lokstra (Router Templates)

**Goal**: Understand Lokstra fundamentals

**Progression**:
1. `01_router/01_router_only` - Learn routing and CRUD
2. `01_router/02_single_app` - Learn app wrapper
3. `01_router/03_multi_app` - Learn multi-app servers

**Time**: 1-2 hours

---

#### Path B: Building Production Apps (Framework Templates)

**Goal**: Build production-ready applications

**Prerequisites**: 
- ✅ Understand Go basics
- ✅ Completed router templates OR understand Lokstra routing
- ✅ Understand REST APIs

**Choose**:
- **Small/Medium project** → `02_app_framework/01_medium_system`
- **Large project** → `02_app_framework/02_enterprise_modular`

**Time**: 2-4 hours

---

### 2. Run a Template

```bash
# Navigate to project root
cd /path/to/lokstra

# Run a template
go run ./project_templates/{category}/{template}

# Example: Run medium system
go run ./project_templates/02_app_framework/01_medium_system
```

---

### 3. Test the APIs

Each template includes a `test.http` file:

1. Open `test.http` in VS Code
2. Install REST Client extension (if not installed)
3. Click "Send Request" above each API call
4. See results inline

---

## 📚 Learning Resources

### Template Documentation

Each template has comprehensive README:

- **Router Templates**:
  - [01_router_only/README.md](./01_router/01_router_only/README.md)
  - [02_single_app/README.md](./01_router/02_single_app/README.md)
  - [03_multi_app/README.md](./01_router/03_multi_app/README.md)

- **Framework Templates**:
  - [01_medium_system/README.md](./02_app_framework/01_medium_system/README.md)
  - [02_enterprise_modular/README.md](./02_app_framework/02_enterprise_modular/README.md)
  - [TEMPLATES_COMPARISON.md](./02_app_framework/TEMPLATES_COMPARISON.md)

### Category Overviews

- [01_router/README.md](./01_router/README.md) - Router patterns overview
- [02_app_framework/README.md](./02_app_framework/README.md) - Framework patterns overview

---

## 🎓 Learning Path

### Beginner

**Goal**: Understand Lokstra basics

1. **01_router_only**: Learn routing, handlers, CRUD
2. **02_single_app**: Learn app lifecycle, graceful shutdown
3. **03_multi_app**: Learn multi-app deployment

**What you'll learn**:
- ✅ Route registration
- ✅ Handler functions
- ✅ Request/Response handling
- ✅ Middleware usage
- ✅ Auto-binding with struct tags
- ✅ Multi-app servers

---

### Intermediate

**Goal**: Build domain-driven applications

4. **01_medium_system**: Learn clean architecture
   - Domain layer (entities, contracts)
   - Service layer (business logic)
   - Repository pattern (data access)
   - Config-driven deployment

**What you'll learn**:
- ✅ Clean architecture
- ✅ Domain-driven design basics
- ✅ Dependency injection
- ✅ Factory pattern
- ✅ Repository pattern

---

### Advanced

**Goal**: Master enterprise architecture

5. **02_enterprise_modular**: Learn DDD with bounded contexts
   - Module structure (bounded contexts)
   - Per-module configuration
   - Team scalability patterns
   - Microservices readiness

**What you'll learn**:
- ✅ Bounded contexts
- ✅ Domain-driven design (strategic)
- ✅ Modular architecture
- ✅ Multi-team patterns
- ✅ Microservices decomposition

---

## 🏗 Architecture Patterns

### Router Patterns (01_router)

**Focus**: HTTP layer, routing, middleware

```
Router → Handlers → Business Logic (inline)
```

**When to use**:
- Simple APIs
- Learning Lokstra
- Microservices (single responsibility)
- No complex domain logic

---

### Medium System (02_app_framework/01_medium_system)

**Focus**: Clean separation of concerns

```
Domain ← Service ← Repository
  ↓
HTTP Layer (auto-generated by Lokstra)
```

**When to use**:
- 2-10 entities
- Single team
- Moderate complexity
- Monolith deployment

---

### Enterprise Modular (02_app_framework/02_enterprise_modular)

**Focus**: Bounded contexts, modularity

```
modules/
  user/     → Domain ← Application ← Infrastructure
  order/    → Domain ← Application ← Infrastructure
  product/  → Domain ← Application ← Infrastructure
```

**When to use**:
- 10+ entities
- Multiple teams
- Complex domain
- Future microservices

---

## 💡 Common Patterns

### Pattern 1: CRUD Operations

**All templates demonstrate**:
- ✅ GET (single & list)
- ✅ POST (create)
- ✅ PUT (update)
- ✅ DELETE (delete)

**Learn in**: `01_router/01_router_only`

---

### Pattern 2: Middleware

**All templates demonstrate**:
- ✅ Recovery middleware
- ✅ Request logging
- ✅ Custom middleware

**Learn in**: `01_router/01_router_only`

---

### Pattern 3: Auto-Binding

**All templates demonstrate**:
- ✅ Path parameters (`path:"id"`)
- ✅ Query parameters (`query:"status"`)
- ✅ JSON body (`json:"name"`)
- ✅ Validation (`validate:"required,email"`)

**Learn in**: `01_router/01_router_only`

---

### Pattern 4: Multi-App Deployment

**Templates**: `01_router/03_multi_app`

Learn how to:
- Run multiple apps in one process
- Share middleware and components
- Separate admin from public APIs
- Deploy on different ports

---

### Pattern 5: Domain-Driven Design

**Templates**: `02_app_framework/*`

Learn how to:
- Separate business logic from infrastructure
- Use repository pattern
- Implement dependency injection
- Structure for scalability

---

### Pattern 6: Bounded Contexts

**Templates**: `02_app_framework/02_enterprise_modular`

Learn how to:
- Organize by business capability
- Maintain module independence
- Scale teams independently
- Prepare for microservices

---

## 🔧 Template Features

### All Templates Include

- ✅ **Working code**: Compiles and runs
- ✅ **In-memory storage**: No setup needed
- ✅ **Seed data**: Pre-populated test data
- ✅ **test.http**: API testing
- ✅ **README**: Comprehensive docs
- ✅ **.gitignore**: Standard Go ignores
- ✅ **Comments**: Explain key concepts

### What's NOT Included (by design)

Templates are starting points. Production needs:

- ❌ Real database connections
- ❌ Authentication/Authorization
- ❌ Rate limiting
- ❌ Distributed tracing
- ❌ CI/CD configs
- ❌ Docker/K8s configs

**Why?** These are environment/business-specific decisions.

---

## 🎯 Real-World Scenarios

### Scenario 1: Simple REST API

**Requirements**:
- Single entity (users)
- Basic CRUD
- No complex business logic

**Template**: `01_router/01_router_only`

---

### Scenario 2: Admin Dashboard + Public API

**Requirements**:
- Two separate apps
- Different ports
- Shared middleware

**Template**: `01_router/03_multi_app`

---

### Scenario 3: Blog Platform

**Requirements**:
- Posts, comments, users, categories
- Moderate business logic
- Single team

**Template**: `02_app_framework/01_medium_system`

---

### Scenario 4: E-Commerce Platform

**Requirements**:
- Users, products, orders, payments, shipping
- Multiple teams
- Future microservices

**Template**: `02_app_framework/02_enterprise_modular`

---

## 🚢 Deployment

### Local Development

```bash
# Run directly
go run ./project_templates/{category}/{template}

# Build binary
go build -o myapp ./project_templates/{category}/{template}
./myapp
```

---

### Docker

```dockerfile
# Example Dockerfile
FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./project_templates/02_app_framework/01_medium_system

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/project_templates/02_app_framework/01_medium_system/config.yaml .
CMD ["./server"]
```

---

### Kubernetes

Deploy framework templates as microservices using per-module configs.

See [02_enterprise_modular/README.md](./02_app_framework/02_enterprise_modular/README.md) for microservices deployment patterns.

---

## 🔄 Migration Paths

### From Router to Framework

**When**: Domain logic becomes complex

**Steps**:
1. Create `domain/` folder
2. Extract entities from handlers
3. Create service layer
4. Create repository layer
5. Update handlers to use services

---

### From Medium to Enterprise

**When**: Growing past 10 entities, need teams

**Steps**:
1. Create `modules/` structure
2. Group related domains into modules
3. Split config into per-module YAMLs
4. Update imports
5. Assign modules to teams

See [TEMPLATES_COMPARISON.md](./02_app_framework/TEMPLATES_COMPARISON.md) for detailed migration guide.

---

## 📝 Best Practices

### General

- ✅ Start with simpler templates
- ✅ Understand each pattern before moving to next
- ✅ Use `test.http` to experiment
- ✅ Read all comments in code
- ✅ Customize templates for your needs

### Router Templates

- ✅ Keep handlers focused and small
- ✅ Use middleware for cross-cutting concerns
- ✅ Validate input with struct tags
- ✅ Return proper HTTP status codes

### Framework Templates

- ✅ Keep domain pure (no framework deps)
- ✅ Use interfaces for dependencies
- ✅ Implement repository pattern for data access
- ✅ Follow dependency injection pattern
- ✅ Organize by domain/module

---

## 🛠 Prerequisites

- **Go 1.23+**
- **VS Code** (recommended) with REST Client extension
- **Lokstra** framework (this project)
- **Understanding of**:
  - Go programming
  - HTTP/REST basics
  - (For framework templates) Architecture patterns

---

## 📞 Support

- **Documentation**: [Lokstra Docs](https://primadi.github.io/lokstra/)
- **Issues**: [GitHub Issues](https://github.com/primadi/lokstra/issues)
- **Templates**: This folder
- **Examples**: See individual template READMEs

---

## 📄 License

These templates are part of the Lokstra framework. See LICENSE file in project root.

---

## 🎉 Next Steps

1. **Choose a template** using the decision guide above
2. **Read the template's README**
3. **Run the template** and test the APIs
4. **Study the code** and comments
5. **Modify for your needs**
6. **Build something awesome!**

Happy coding with Lokstra! 🚀

---

## 📑 Quick Links

### Router Templates
- [Router Overview](./01_router/README.md)
- [Router Only](./01_router/01_router_only/README.md)
- [Single App](./01_router/02_single_app/README.md)
- [Multi App](./01_router/03_multi_app/README.md)

### Framework Templates
- [Framework Overview](./02_app_framework/README.md)
- [Medium System](./02_app_framework/01_medium_system/README.md)
- [Enterprise Modular](./02_app_framework/02_enterprise_modular/README.md)
- [Templates Comparison](./02_app_framework/TEMPLATES_COMPARISON.md)
