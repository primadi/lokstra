# README: Lokstra Skills for GitHub Copilot

This directory contains modular skill definitions for AI agents (GitHub Copilot, Cursor, etc.) to generate production-ready code using the Lokstra Framework.

---

## Overview

**Lokstra Framework** is a Go web framework with two modes:
1. **Router Mode** - Lightweight HTTP routing (like Echo, Gin, Chi)
2. **Framework Mode** - Full dependency injection framework (like NestJS, Spring Boot)

**Key Feature:** Annotation-based code generation (`@Handler`, `@Service`, `@Route`, `@Inject`) eliminates boilerplate while maintaining type safety.

---

## Skills Structure

### Foundation (Read First)
- **[00-lokstra-overview.md](00-lokstra-overview.md)** - Framework philosophy, architecture, design patterns

### Document-Driven Development Workflow
- **[01-document-workflow.md](01-document-workflow.md)** - SKILL 0: Create Business Requirements Document (BRD)
- **[02-module-requirements.md](02-module-requirements.md)** - SKILL 1: Generate Module Requirements
- **[03-api-spec.md](03-api-spec.md)** - SKILL 2: Generate API Specification (OpenAPI)
- **[04-schema.md](04-schema.md)** - SKILL 3: Generate Database Schema

### Implementation Skills
- **[05-implementation.md](05-implementation.md)** - SKILL 4-7: Module structure, migrations, domain models, repositories
- **[06-implementation-advanced.md](06-implementation-advanced.md)** - SKILL 8-13: Handlers, config, tests, consistency checks

---

## Quick Start for AI Agents

### 1. Starting a New Project

**User Request:** *"Create a new e-commerce order management system"*

**AI Agent Workflow:**
```
1. Read: 00-lokstra-overview.md (understand framework)
2. Execute: SKILL 0 (01-document-workflow.md)
   → Ask clarifying questions
   → Generate BRD (docs/BRD.md)
   → Wait for user approval

3. Execute: SKILL 1 (02-module-requirements.md)
   → Identify modules (bounded contexts)
   → Generate module requirements
   → Wait for approval

4. Execute: SKILL 2 (03-api-spec.md)
   → Generate API spec for each module
   → Wait for approval

5. Execute: SKILL 3 (04-schema.md)
   → Generate database schema
   → Create migration files
   → Wait for approval

6. Execute: SKILL 4-13 (05-implementation.md, 06-implementation-advanced.md)
   → Implement module (domain, repository, handler)
   → Generate config.yaml
   → Generate tests
   → Run consistency check
```

**Output:**
```
myapp/
├── .github/
│   ├── skills/                    # These skill files
│   └── copilot-instructions.md
├── config.yaml
├── docs/
│   ├── BRD.md                     # Business requirements
│   ├── modules/
│   │   └── order/
│   │       ├── REQUIREMENTS.md
│   │       ├── API_SPEC.md
│   │       └── SCHEMA.md
│   └── templates/                  # Document templates
├── migrations/
│   └── order/
│       └── 001_create_orders_table.sql
└── modules/
    └── order/
        ├── handler/
        │   ├── order_handler.go
        │   └── zz_generated.lokstra.go  # Auto-generated
        ├── repository/
        │   └── order_repository.go
        └── domain/
            ├── order.go
            └── order_service.go
```

---

## Usage Patterns

### Pattern 1: Complete New Project
**User:** *"Build a task management system with users, projects, and tasks"*

**AI Agent:**
1. Read `00-lokstra-overview.md`
2. Execute SKILL 0 → Generate BRD
3. Execute SKILL 1 → Identify 3 modules: `auth`, `project`, `task`
4. For each module:
   - SKILL 2 → API Spec
   - SKILL 3 → Schema
   - SKILL 4-13 → Implementation

---

### Pattern 2: Add Module to Existing Project
**User:** *"Add payment processing module"*

**AI Agent:**
1. Skip BRD (project exists)
2. Execute SKILL 1 → Generate `docs/modules/payment/REQUIREMENTS.md`
3. Execute SKILL 2 → Generate API spec
4. Execute SKILL 3 → Generate schema + migrations
5. Execute SKILL 4-13 → Implement module

---

### Pattern 3: Modify Existing Module
**User:** *"Add order cancellation feature"*

**AI Agent:**
1. Read existing docs: `docs/modules/order/REQUIREMENTS.md`, `API_SPEC.md`, `SCHEMA.md`
2. Update docs:
   - Add FR-ORDER-006 to requirements
   - Add `DELETE /api/orders/{id}` to API spec
   - Update schema if needed
3. Update implementation:
   - Add `Cancel` method to handler
   - Add `Delete` method to repository
4. Execute SKILL 13 → Consistency check

---

## Document Templates

Located in: [../../docs/templates/](../../docs/templates/)

- **BRD_TEMPLATE.md** - Business Requirements Document
- **MODULE_REQUIREMENTS_TEMPLATE.md** - Module-specific requirements
- **API_SPEC_TEMPLATE.md** - API endpoint specifications
- **SCHEMA_TEMPLATE.md** - Database schema documentation
- **CHANGELOG_TEMPLATE.md** - Document version tracking

---

## Key Principles for AI Agents

### 1. Always Design Before Code
```
❌ Wrong: Generate handler → Realize missing validation → Fix
✅ Right: BRD → Requirements → API Spec → Schema → Implementation
```

### 2. One File Per Entity
```
❌ Wrong: handlers/handlers.go (5 entities)
✅ Right: 
    handler/user_handler.go
    handler/order_handler.go
    handler/product_handler.go
```

### 3. Bounded Context (Module Granularity)
```
❌ Wrong: modules/user/ (auth + profile + notifications)
✅ Right: 
    modules/auth/
    modules/user_profile/
    modules/notification/
```

### 4. Use Annotations (Not Manual Registration)
```
❌ Wrong: Manual lokstra_registry.RegisterServiceType(...)
✅ Right: 
    // @Handler name="order-handler", prefix="/api/orders"
    // @Route "POST /"
```

### 5. Cross-Module Communication via Repositories
```
❌ Wrong: orderHandler → paymentHandler (direct call)
✅ Right: orderHandler → paymentRepository (injection)
```

---

## Validation & Quality

### Consistency Check (SKILL 13)

After implementation, validate:

- [ ] **Requirements → API:** All FRs have endpoints
- [ ] **API → Handler:** All endpoints implemented with @Route
- [ ] **Schema → Repository:** All tables have CRUD methods
- [ ] **Validation:** Request DTOs have correct `validate` tags
- [ ] **Tests:** Unit coverage > 80%, integration tests pass
- [ ] **Config:** All services registered in config.yaml

---

## AI Agent Decision Tree

```
User Request?
│
├─ New Project?
│  ├─ Yes → SKILL 0 (BRD)
│  │       → SKILL 1 (Module Requirements)
│  │       → For each module: SKILL 2-13
│  │
│  └─ No → Existing project
│          │
│          ├─ New Module?
│          │  └─ Yes → SKILL 1-13
│          │
│          └─ Modify Existing?
│              └─ Update docs → Update code → SKILL 13
│
└─ Simple Script/Prototype?
   └─ Use Router Mode (skip DI/annotations)
```

---

## Common Imports

```go
import "github.com/primadi/lokstra"
import "github.com/primadi/lokstra/core/request"
import "github.com/primadi/lokstra/core/service"
import "github.com/primadi/lokstra/lokstra_registry"
import "github.com/primadi/lokstra/middleware/recovery"
import "github.com/primadi/lokstra/middleware/cors"
```

---

## Resources

- **Full Framework Docs:** https://primadi.github.io/lokstra/
- **AI Agent Guide:** [AI-AGENT-GUIDE.md](https://primadi.github.io/lokstra/AI-AGENT-GUIDE)
- **Quick Reference:** [QUICK-REFERENCE.md](https://primadi.github.io/lokstra/QUICK-REFERENCE)
- **Examples:** https://primadi.github.io/lokstra/00-introduction/examples/
- **Project Templates:** [../../project_templates/](../../project_templates/)

---

## Auto-Loading for New Projects

These skills are automatically copied to new projects created via:

```bash
lokstra new myapp
```

New projects will have:
```
myapp/
├── .github/
│   ├── skills/           # This directory (copied)
│   └── copilot-instructions.md
└── docs/
    └── templates/        # Document templates (copied)
```

This ensures AI agents can generate code correctly in user projects.

---

## Support

For issues or questions:
- **GitHub:** https://github.com/primadi/lokstra
- **Documentation:** https://primadi.github.io/lokstra/
- **Discord:** (link in repo README)

---

**Start Here:** [00-lokstra-overview.md](00-lokstra-overview.md)
