# Framework Templates Comparison

## Overview

Lokstra provides **two framework templates** for building applications with domain-driven architecture:

1. **01_medium_system** - For medium-sized applications (2-10 entities)
2. **02_enterprise_modular** - For enterprise applications (10+ entities) with bounded contexts

---

## Quick Comparison

| Feature | Medium System | Enterprise Modular |
|---------|--------------|-------------------|
| **Entity Count** | 2-10 entities | 10+ entities |
| **Structure** | Flat (domain/, repository/, service/) | Modular (modules/{context}/) |
| **Architecture** | Clean Architecture | DDD with Bounded Contexts |
| **Configuration** | Single config.yaml | Per-module YAML in config/ |
| **Team Size** | Single team | Multiple teams |
| **Complexity** | Low | High |
| **Learning Curve** | Easy | Requires DDD knowledge |
| **Portability** | Moderate | High (copy module folder) |
| **Microservices** | Harder to split | Easy to split |
| **Code Organization** | By layer (all domains together) | By module (domain grouped) |

---

## Structure Comparison

### Medium System (Flat Structure)

```
01_medium_system/
├── domain/
│   ├── user/           # User domain
│   │   ├── entity.go
│   │   ├── contract.go
│   │   └── dto.go
│   └── order/          # Order domain
│       ├── entity.go
│       ├── contract.go
│       └── dto.go
├── repository/
│   ├── user_repository.go
│   └── order_repository.go
├── service/
│   ├── user_service.go
│   └── order_service.go
├── config.yaml         # Single config
├── main.go
└── register.go
```

**Organization**: By technical layer (all services together, all repos together)

---

### Enterprise Modular (DDD Structure)

```
02_enterprise_modular/
├── modules/
│   ├── user/           # User bounded context
│   │   ├── domain/
│   │   │   ├── entity.go
│   │   │   ├── service.go
│   │   │   └── dto.go
│   │   ├── application/
│   │   │   └── user_service.go
│   │   └── infrastructure/
│   │       └── repository/
│   │           └── user_repository.go
│   └── order/          # Order bounded context
│       ├── domain/
│       ├── application/
│       └── infrastructure/
├── config/
│   ├── user.yaml       # Per-module config
│   └── order.yaml
├── main.go
└── register.go
```

**Organization**: By business capability (everything for user module together)

---

## When to Use Each

### Use Medium System When:

✅ You have **2-10 entities**  
✅ **Single team** working on the project  
✅ **Simple to moderate** domain complexity  
✅ Need **quick development** without much overhead  
✅ **Monolith deployment** is sufficient  
✅ Domain relationships are **simple and clear**  

**Examples**:
- Blog platform (posts, comments, users)
- Inventory system (products, categories, suppliers)
- Booking system (bookings, customers, rooms)

---

### Use Enterprise Modular When:

✅ You have **10+ entities**  
✅ **Multiple teams** need to work independently  
✅ **Complex domain** requiring bounded contexts  
✅ Need **modularity** for future microservices  
✅ Different parts of system have **different deployment needs**  
✅ Domain has **multiple business contexts** with different terminology  

**Examples**:
- E-commerce platform (users, products, orders, payments, shipping, reviews, analytics)
- Banking system (accounts, transactions, loans, investments, compliance)
- Hospital management (patients, appointments, billing, pharmacy, lab results)

---

## Migration Path

### From Medium to Enterprise

When your medium system grows to 10+ entities:

1. **Create modules folder**:
   ```bash
   mkdir -p modules/user/{domain,application,infrastructure/repository}
   ```

2. **Move domain files**:
   ```bash
   # Move domain/user/* → modules/user/domain/
   mv domain/user/* modules/user/domain/
   ```

3. **Move service files**:
   ```bash
   # Move service/user_service.go → modules/user/application/
   mv service/user_service.go modules/user/application/
   ```

4. **Move repository files**:
   ```bash
   # Move repository/user_repository.go → modules/user/infrastructure/repository/
   mv repository/user_repository.go modules/user/infrastructure/repository/
   ```

5. **Split config**:
   ```bash
   mkdir config
   # Create config/user.yaml, config/order.yaml from config.yaml
   ```

6. **Update imports** in all files to reflect new structure

7. **Update register.go** with new import paths

---

## Code Examples

### Medium System - User Service

```go
// service/user_service.go
package service

import (
    "github.com/primadi/lokstra/project_templates/.../domain/user"
)

type UserServiceImpl struct {
    UserRepo *service.Cached[user.UserRepository]
}

func (s *UserServiceImpl) GetByID(p *user.GetUserParams) (*user.User, error) {
    return s.UserRepo.MustGet().GetByID(p.ID)
}
```

**Import**: `domain/user` (cross-layer import)

---

### Enterprise Modular - User Service

```go
// modules/user/application/user_service.go
package application

import (
    "github.com/primadi/lokstra/project_templates/.../modules/user/domain"
)

type UserServiceImpl struct {
    UserRepo *service.Cached[domain.UserRepository]
}

func (s *UserServiceImpl) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
    return s.UserRepo.MustGet().GetByID(p.ID)
}
```

**Import**: `modules/user/domain` (stays within module)

---

## Configuration Differences

### Medium System - Single Config

```yaml
# config.yaml
deployments:
  - name: api-server
    type: server
    port: 3000
    
    services:
      - name: user-service
        factory: user-service-factory
        endpoints:
          - path: /api/users/{id}
            method: GET
            handler: GetByID
      
      - name: order-service
        factory: order-service-factory
        endpoints:
          - path: /api/orders/{id}
            method: GET
            handler: GetByID
```

**All services in one file**

---

### Enterprise Modular - Per-Module Config

**config/user.yaml**:
```yaml
deployments:
  - name: api-server
    type: server
    port: 3000
    
    services:
      - name: user-service
        factory: user-service-factory
        endpoints:
          - path: /api/users/{id}
            method: GET
            handler: GetByID
```

**config/order.yaml**:
```yaml
deployments:
  - name: api-server  # Same name = merges
    type: server
    port: 3000
    
    services:
      - name: order-service
        factory: order-service-factory
        endpoints:
          - path: /api/orders/{id}
            method: GET
            handler: GetByID
```

**Lokstra automatically merges** configs with same deployment name!

---

## Benefits Comparison

### Medium System Benefits

✅ **Simplicity**: Flat structure, easy to navigate  
✅ **Fast Development**: Less ceremony, direct coding  
✅ **Easy Learning**: Standard Go project structure  
✅ **Single Config**: One file to manage deployment  

### Enterprise Modular Benefits

✅ **Modularity**: Clear module boundaries  
✅ **Team Scalability**: Each team owns a module  
✅ **Portability**: Copy module = portable unit  
✅ **Microservices Ready**: Easy to split later  
✅ **Domain Clarity**: Bounded contexts make complex domains manageable  
✅ **Independent Evolution**: Change one module without affecting others  

---

## Decision Matrix

Use this to decide which template:

| Question | Medium | Enterprise |
|----------|--------|------------|
| Do you have < 10 entities? | ✅ Yes | ❌ No |
| Single team working on project? | ✅ Yes | ❌ No |
| Simple domain relationships? | ✅ Yes | ❌ No |
| Need quick development? | ✅ Yes | ⚠️ Maybe |
| Planning microservices? | ⚠️ Maybe | ✅ Yes |
| Complex business contexts? | ❌ No | ✅ Yes |
| Need module portability? | ⚠️ Maybe | ✅ Yes |
| Team working independently? | ❌ No | ✅ Yes |

**Legend**: ✅ Strongly fits | ⚠️ Can work | ❌ Not ideal

---

## Real-World Examples

### Medium System Example: Blog Platform

**Entities**: Post, Comment, User, Category, Tag (5 entities)

```
domain/
  ├── post/
  ├── comment/
  ├── user/
  ├── category/
  └── tag/
```

**Why Medium**: 
- 5 entities (< 10)
- Simple relationships
- Single team
- Quick development needed

---

### Enterprise Example: E-Commerce Platform

**Modules**: User, Product, Order, Payment, Shipping, Review, Analytics, Notification (8+ modules)

```
modules/
  ├── user/        # Auth, profile
  ├── product/     # Catalog, inventory
  ├── order/       # Order management
  ├── payment/     # Payment processing
  ├── shipping/    # Logistics
  ├── review/      # Product reviews
  ├── analytics/   # Reporting
  └── notification/ # Email, SMS
```

**Why Enterprise**:
- 8+ bounded contexts
- 15+ total entities
- Multiple teams
- Complex business rules
- Future microservices

---

## Summary

### Choose Medium System If:
- **Small to medium** complexity (2-10 entities)
- **Single team**
- Need **fast development**
- **Monolith** is sufficient

### Choose Enterprise Modular If:
- **Large** complexity (10+ entities)
- **Multiple teams**
- Need **modularity** and **scalability**
- Planning **microservices**
- Complex **business contexts**

**Remember**: You can always migrate from Medium to Enterprise as your system grows!

---

## Next Steps

1. **Start with Medium** if you're unsure
2. **Monitor growth**: When you hit 10+ entities, consider migrating
3. **Evaluate team size**: Multiple teams = Enterprise
4. **Check complexity**: Complex domains = Enterprise
5. **Plan deployment**: Microservices later = Enterprise now

Both templates are production-ready and follow best practices. Choose based on your current and future needs!
