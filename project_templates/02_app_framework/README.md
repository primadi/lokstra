# Lokstra App Framework Templates

**Production-Ready Templates for Domain-Driven Applications**

This folder contains framework templates that demonstrate how to build **production-ready applications** using Lokstra with domain-driven architecture patterns.

---

## 📂 Available Templates

### 1. [01_medium_system](./01_medium_system/)

**For medium-sized applications (2-10 entities)**

```
domain/ + repository/ + service/ + config.yaml
```

- ✅ Clean Architecture with flat structure
- ✅ Domain-driven design principles
- ✅ Single config file
- ✅ Perfect for single team
- ✅ Quick development

**Use when**: You have 2-10 entities, single team, monolith deployment

---

### 2. [02_enterprise_modular](./02_enterprise_modular/)

**For enterprise applications (10+ entities)**

```
modules/{context}/{domain,application,infrastructure}/ + config/{module}.yaml
```

- ✅ DDD with Bounded Contexts
- ✅ Modular architecture per business capability
- ✅ Per-module configuration files
- ✅ Multi-team scalability
- ✅ Microservices-ready

**Use when**: You have 10+ entities, multiple teams, need modularity

---

## 🎯 Which Template Should I Use?

### Quick Decision Guide

| Your Situation | Recommended Template |
|---------------|---------------------|
| 2-10 entities, single team | **01_medium_system** |
| 10+ entities, multiple teams | **02_enterprise_modular** |
| Simple domain, quick start | **01_medium_system** |
| Complex domain, bounded contexts | **02_enterprise_modular** |
| Monolith only | **01_medium_system** |
| Future microservices | **02_enterprise_modular** |
| Learning DDD | **01_medium_system** (start here) |
| Production enterprise app | **02_enterprise_modular** |

**Not sure?** → Start with **01_medium_system**, migrate to **02_enterprise_modular** when needed

---

## 📊 Template Comparison

See [TEMPLATES_COMPARISON.md](./TEMPLATES_COMPARISON.md) for detailed comparison including:
- Structure differences
- When to use each
- Migration path
- Code examples
- Real-world scenarios

---

## 🏗 Architecture Overview

Both templates follow **Clean Architecture** principles with domain-driven design:

### Medium System (Flat)

```
Layers organized by technical concern:

┌─────────────────────────────────┐
│         domain/                 │  ← Entities & Contracts
│    (user/, order/)              │
├─────────────────────────────────┤
│         service/                │  ← Business Logic
│    (user_service, order_service)│
├─────────────────────────────────┤
│         repository/             │  ← Data Access
│    (user_repo, order_repo)      │
└─────────────────────────────────┘
```

**Good for**: When all domains are closely related

---

### Enterprise Modular (DDD)

```
Modules organized by business capability:

┌──────────────────┐  ┌──────────────────┐
│   modules/user/  │  │  modules/order/  │
│  ┌────────────┐  │  │  ┌────────────┐  │
│  │  domain    │  │  │  │  domain    │  │
│  ├────────────┤  │  │  ├────────────┤  │
│  │application │  │  │  │application │  │
│  ├────────────┤  │  │  ├────────────┤  │
│  │infrastructure│ │  │  │infrastructure│ │
│  └────────────┘  │  │  └────────────┘  │
└──────────────────┘  └──────────────────┘
  User Context          Order Context
```

**Good for**: When domains need clear boundaries

---

## 🚀 Quick Start

### Option 1: Medium System

```bash
cd lokstra-project-root
go run ./project_templates/02_app_framework/01_medium_system
```

Server starts on `http://localhost:3000`

**Test APIs**:
- Open `01_medium_system/test.http` in VS Code
- Click "Send Request" to test endpoints

---

### Option 2: Enterprise Modular

```bash
cd lokstra-project-root
go run ./project_templates/02_app_framework/02_enterprise_modular
```

Server starts on `http://localhost:3000`

**Test APIs**:
- Open `02_enterprise_modular/test.http` in VS Code
- Click "Send Request" to test endpoints

---

## 📚 What You'll Learn

### Medium System Teaches:

1. **Clean Architecture**: Separation of concerns
2. **Domain Layer**: Entities and contracts
3. **Service Layer**: Business logic implementation
4. **Repository Pattern**: Data access abstraction
5. **Dependency Injection**: Factory-based DI
6. **Config-Driven Deployment**: Single YAML configuration

---

### Enterprise Modular Teaches:

All of the above, plus:

7. **Bounded Contexts**: Module isolation
8. **Domain-Driven Design**: Strategic patterns
9. **Modular Configuration**: Per-module YAML files
10. **Team Scalability**: Multi-team architecture
11. **Microservices Patterns**: Service decomposition
12. **Module Portability**: Copy-paste modules

---

## 🎓 Learning Path

### Beginner (Start Here)

1. **Complete Router Templates** first:
   - `project_templates/01_router/01_router_only/`
   - `project_templates/01_router/02_single_app/`
   - `project_templates/01_router/03_multi_app/`

2. **Then try Medium System**:
   - `project_templates/02_app_framework/01_medium_system/`

### Intermediate

3. **Study Enterprise Modular**:
   - `project_templates/02_app_framework/02_enterprise_modular/`

4. **Read comparison**:
   - `TEMPLATES_COMPARISON.md`

### Advanced

5. **Build your own** using these templates as reference
6. **Experiment** with microservices deployment
7. **Customize** for your specific domain

---

## 📖 Documentation

### Per-Template Documentation

Each template has comprehensive README:

- **[01_medium_system/README.md](./01_medium_system/README.md)**
  - When to use
  - Project structure
  - Getting started
  - Adding domains
  - API documentation

- **[02_enterprise_modular/README.md](./02_enterprise_modular/README.md)**
  - When to use
  - DDD concepts
  - Module structure
  - Adding modules
  - Team scalability
  - Deployment strategies

### Comparison Guide

- **[TEMPLATES_COMPARISON.md](./TEMPLATES_COMPARISON.md)**
  - Side-by-side comparison
  - Decision matrix
  - Migration path
  - Real-world examples

---

## 🔧 Common Features

Both templates include:

- ✅ **In-Memory Storage**: No database setup needed
- ✅ **Seed Data**: Pre-populated test data
- ✅ **REST API**: Full CRUD operations
- ✅ **Validation**: Request validation with struct tags
- ✅ **Error Handling**: Proper error responses
- ✅ **Logging**: Request/response logging
- ✅ **test.http**: API testing with VS Code REST Client
- ✅ **README**: Comprehensive documentation
- ✅ **Production-Ready**: Follows best practices

---

## 💡 Key Differences

| Feature | Medium | Enterprise |
|---------|---------|-----------|
| **Structure** | Flat by layer | Nested by module |
| **Configuration** | Single file | Per-module files |
| **Import Paths** | Cross-layer | Within-module |
| **Team Model** | Single team | Multiple teams |
| **Portability** | Moderate | High |
| **Complexity** | Low | High |

---

## 🎯 Real-World Use Cases

### Medium System Examples:

- **Blog Platform**: Posts, comments, users, categories
- **Inventory System**: Products, suppliers, stock
- **Booking System**: Bookings, customers, rooms
- **Task Management**: Projects, tasks, users
- **CMS**: Pages, media, users, settings

### Enterprise Modular Examples:

- **E-Commerce**: User, product, order, payment, shipping, review, analytics
- **Banking**: Account, transaction, loan, investment, compliance
- **Healthcare**: Patient, appointment, billing, pharmacy, lab
- **ERP**: Sales, inventory, procurement, HR, finance
- **SaaS Platform**: Auth, tenant, billing, analytics, support

---

## 🚢 Deployment Options

### Medium System

```yaml
# config.yaml - Single deployment
deployments:
  - name: api-server
    type: server
    port: 3000
    services:
      - name: user-service
      - name: order-service
```

**Deployment**: Single process, all services together

---

### Enterprise Modular

```yaml
# config/user.yaml
deployments:
  - name: api-server
    port: 3000
    services:
      - name: user-service

# config/order.yaml
deployments:
  - name: api-server  # Same name = merge
    port: 3000
    services:
      - name: order-service
```

**Flexible Deployment**:
- **Monolith**: `go run . config/` (all modules)
- **Microservices**: `go run . config/user.yaml` (per module)
- **Hybrid**: Group related modules

---

## 🔄 Migration Guide

### From Medium to Enterprise

When to migrate:
- ✅ Growing from 10 to 20+ entities
- ✅ Need to split teams
- ✅ Planning microservices
- ✅ Domain complexity increasing

**Steps**:
1. Create `modules/` folder structure
2. Move `domain/{entity}/` → `modules/{entity}/domain/`
3. Move `service/{entity}_service.go` → `modules/{entity}/application/`
4. Move `repository/{entity}_repository.go` → `modules/{entity}/infrastructure/repository/`
5. Split `config.yaml` → `config/{entity}.yaml` per module
6. Update imports in all files
7. Update `register.go` with new paths

See [TEMPLATES_COMPARISON.md](./TEMPLATES_COMPARISON.md) for detailed migration guide.

---

## 🛠 Prerequisites

- **Go 1.23+**
- **VS Code** (recommended) with REST Client extension
- **Lokstra** framework (in parent directory)
- Understanding of:
  - Go programming
  - REST APIs
  - Basic architecture patterns

---

## 📝 Template Structure

### What's Included

Each template contains:

```
{template}/
├── domain/          or  modules/
├── service/         or  application/
├── repository/      or  infrastructure/
├── config.yaml      or  config/*.yaml
├── main.go             ← Entry point
├── register.go         ← Service registration
├── test.http           ← API tests
├── README.md           ← Documentation
└── .gitignore          ← Git ignore rules
```

### What's NOT Included (by design)

These templates are **starting points**. Production apps need:

- ❌ Real database (templates use in-memory)
- ❌ Authentication/Authorization
- ❌ Rate limiting
- ❌ Caching layer
- ❌ Monitoring/Metrics
- ❌ CI/CD configuration
- ❌ Docker/Kubernetes configs
- ❌ API documentation (Swagger/OpenAPI)

**Why?** These are environment-specific and business-specific decisions you should make based on your needs.

---

## 🎨 Customization

### Extending Templates

Both templates are designed to be extended:

1. **Add domains/modules**: Follow existing patterns
2. **Add middleware**: Register in `register.go`
3. **Add validation**: Use struct tags
4. **Add persistence**: Swap in-memory repos with DB repos
5. **Add authentication**: Add auth middleware
6. **Add API docs**: Generate from code

### Example: Add PostgreSQL

Replace in-memory repository:

```go
// Before (in-memory)
func NewUserRepositoryMemory() domain.UserRepository {
    return &UserRepositoryMemory{...}
}

// After (PostgreSQL)
func NewUserRepositoryPostgres(db *sql.DB) domain.UserRepository {
    return &UserRepositoryPostgres{db: db}
}
```

Update `register.go` to inject database connection.

---

## 📞 Support

- **Documentation**: [Lokstra Docs](https://primadi.github.io/lokstra/)
- **Issues**: [GitHub Issues](https://github.com/primadi/lokstra/issues)
- **Examples**: See other templates in `project_templates/`

---

## 📄 License

These templates are part of the Lokstra framework. See LICENSE file in project root.

---

## 🎉 Get Started

1. **Choose your template** (use decision guide above)
2. **Read the template's README**
3. **Run the example**
4. **Study the code**
5. **Customize for your needs**

Happy coding with Lokstra! 🚀
