---
name: lokstra-overview
description: Provides foundational understanding of Lokstra Framework architecture, design principles, annotation-based code generation, and decision-making guidelines for AI agents building Go web applications. Use when starting any Lokstra project or when agents need framework context.
phase: design
order: 1
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.18"
---

# Lokstra Framework - Overview & Philosophy

## When to use this skill

Use this skill when:
- Starting any new Lokstra project with AI agents
- Need to understand overall framework philosophy
- Understanding 2-phase workflow: Design → Implementation
- Determining module structure (DDD bounded contexts)
- Need high-level architecture guidance

## Framework Overview

Lokstra is a **Go web framework** with:
- Design-first development approach (BRD → Code)
- Annotation-based code generation
- Full dependency injection framework
- Modular architecture using DDD bounded contexts
- Auto-code generation for routes, services, dependencies

**Two Main Phases:**

### Phase 1: Design (Using Design Skills)
```
BRD → Module Requirements → API Spec → Database Schema
```
- Focus on planning, requirements, and specifications
- Human-readable documents for stakeholder approval
- No code written yet

### Phase 2: Implementation (Using Implementation Skills)
```
Initialize Framework → Config → Handlers → Services → Migrations → Tests
```
- Focus on code generation based on design phase
- Microskills for each implementation concern
- Annotation-based code (minimal boilerplate)

---

## Development Workflow

**Design Phase → Implementation Phase:**

```
1. lokstra-overview (framework fundamentals)
   ↓
2. lokstra-brd-generation (define business requirements)
   ↓
3. lokstra-module-requirements (break into modules)
   ↓
4. lokstra-api-specification (define API endpoints)
   ↓
5. lokstra-schema-design (define database schema)
   ↓
   SPECIFICATIONS READY
   ↓
6. implementation-lokstra-init-framework (initialize)
7. implementation-lokstra-yaml-config (config)
8. implementation-lokstra-create-handler (@Handler)
9. implementation-lokstra-create-service (@Service)
10. implementation-lokstra-create-migrations (migrations)
11. implementation-lokstra-generate-http-files (HTTP client)
    ↓
    CODE READY
    ↓
12. advanced-lokstra-tests (testing)
13. advanced-lokstra-middleware (custom middleware)
14. advanced-lokstra-validate-consistency (validation)
```

**Migrations note:** Apply migrations explicitly with `lokstra migration up` (not auto-run). If `migration.yaml` exists in a folder, it supplies `dbpool-name` / optional `schema-table` / optional `enabled: false`; otherwise use `-db` to pick the pool.

## Core Design Philosophy

### 1. Document-Driven Development

**Problem:** Code-first approaches lead to rework, inconsistencies, technical debt.

**Solution:** Design-first workflow (specs before code):
- Clear stakeholder alignment
- Built-in compliance documentation
- Eliminates "code-then-fix" cycles
- Estimated 10x productivity improvement

### 2. Bounded Context (DDD)

**Module Granularity:**
- ❌ Wrong: One `user` module handling auth + profile + notifications
- ✅ Right: Separate `auth`, `user_profile`, `notification` modules

**Benefits:** Independent scaling, clear responsibilities, easier testing

### 3. Project Structure Consistency

All Lokstra projects follow same structure (simple → enterprise):

```
myapp/
├── .github/skills/              # AI Agent instructions
├── configs/                     # YAML config files (auto-merged)
│   ├── config.yaml
│   ├── database-config.yaml
│   └── <module-name>-config.yaml
├── migrations/                  # Database migrations per module
│   ├── shared/
│   └── <module_name>/
├── docs/
│   ├── drafts/                  # BRD/requirements drafts
│   └── modules/                 # Published specs
└── modules/
    ├── shared/domain/           # Cross-module models
    └── <module_name>/
        ├── handler/             # @Handler (1 per entity)
        ├── repository/          # Data access (1 per entity)
        └── domain/              # Business logic & DTOs
```

**Rules:**
- Multi-file YAML config in `/configs/` - auto-merged
- Migrations by module in `/migrations/{module}/`
- 1 handler per entity (not `handlers.go` with 10+ entities)
- DTOs with validation tags in domain layer

---

## Annotations & Code Generation

**Why Annotations?**
- Reduce boilerplate code
- Type-safe at compile time
- Clear service dependencies
- Auto-generates routing, DI, service registration

**Core Annotations:**
- `@Handler` - Business service with routes
- `@Service` - Infrastructure service
- `@Route` - HTTP endpoint mapping
- `@Inject` - Dependency injection

**Code Details:** See implementation-phase skills for specific syntax

---

## Cross-Module Communication

**Pattern:** Via dependency injection (not direct calls)

**Shared Models:** `/modules/shared/domain/`

**Async Operations:** Event-driven for loose coupling

**Details:** See design-lokstra-module-requirements for integration points

---

## Configuration Management

**Multi-file YAML:**
- All files in `/configs/` auto-merged
- Environment variables supported (${VAR:default})
- Service definitions with dependencies
- Deployment configurations

**Details:** See implementation-lokstra-yaml-config skill

---

## Best Practices Summary

- ✅ Design first, code second
- ✅ Document-driven development
- ✅ DDD bounded contexts (separate modules)
- ✅ 1 file per entity
- ✅ Annotations over manual code
- ✅ Type-safe dependency injection
- ✅ Validation tags on DTOs
```

---

## Next Steps

After understanding the framework overview:

### DESIGN PHASE (Specifications)

1. **[design-lokstra-brd-generation](../design-lokstra-brd-generation/SKILL.md)**
   - Create Business Requirements Document
   - Define stakeholder requirements, user stories, acceptance criteria

2. **[design-lokstra-module-requirements](../design-lokstra-module-requirements/SKILL.md)**
   - Break BRD into modules (Bounded Contexts using DDD)
   - Define module responsibilities and integration points

3. **[design-lokstra-api-specification](../design-lokstra-api-specification/SKILL.md)**
   - Design API endpoints, request/response schemas
   - Create endpoint specs with validation rules and examples

4. **[design-lokstra-schema-design](../design-lokstra-schema-design/SKILL.md)**
   - Design database schema with tables, indexes, constraints
   - Create migration files (UP/DOWN SQL)

### IMPLEMENTATION PHASE (Code Generation)

5. **[implementation-lokstra-init-framework](../implementation-lokstra-init-framework/SKILL.md)**
   - Initialize main.go with lokstra.Bootstrap()
   - Register services and middleware

6. **[implementation-lokstra-yaml-config](../implementation-lokstra-yaml-config/SKILL.md)**
   - Create multi-file YAML configuration
   - Set up service definitions and deployments

7. **[implementation-lokstra-create-handler](../implementation-lokstra-create-handler/SKILL.md)**
   - Create @Handler annotated endpoints
   - Implement @Route decorators with request/response handling

8. **[implementation-lokstra-create-service](../implementation-lokstra-create-service/SKILL.md)**
   - Create @Service repository implementations
   - Implement data access layer

9. **[implementation-lokstra-create-migrations](../implementation-lokstra-create-migrations/SKILL.md)**
   - Create database migration files
   - Generate UP/DOWN SQL from schema design

10. **[implementation-lokstra-generate-http-files](../implementation-lokstra-generate-http-files/SKILL.md)**
    - Create .http client files for testing
    - Generate REST examples with environment variables

### ADVANCED PHASE (Testing & Validation)

11. **[advanced-lokstra-tests](../advanced-lokstra-tests/SKILL.md)**
    - Write unit tests for handlers and services
    - Create integration tests with test database

12. **[advanced-lokstra-middleware](../advanced-lokstra-middleware/SKILL.md)**
    - Implement custom middleware (auth, logging, rate limiting)
    - Add cross-cutting concerns

13. **[advanced-lokstra-validate-consistency](../advanced-lokstra-validate-consistency/SKILL.md)**
    - Validate circular dependencies
    - Check configuration completeness
    - Pre-deployment validation

---

## Key Principles

**1. Design First, Code Second**
- Always create specifications before implementation
- Clear requirements prevent rework

**2. Document-Driven Development**
- Use BRD → Requirements → Specs as single source of truth
- Generated code directly maps to specifications

**3. Bounded Contexts (DDD)**
- Each module is independent and self-contained
- Clear interfaces between modules
- Easier to test and deploy

**4. Type-Safe Dependency Injection**
- Compile-time verification of dependencies
- Configuration-based service selection
- No runtime surprises

**5. Annotation-Based Code Generation**
- Minimize boilerplate code
- Auto-generated routing and service registration
- Consistent code patterns

---

## When to Choose Lokstra

### ✅ Use Lokstra When

- Building production applications (2+ entities/modules)
- Team size: 2-50 developers
- Complexity: Medium to high (multiple services)
- Scale: Growing from MVP to enterprise
- Need: Strong type safety and structured DI

### ❌ Consider Alternatives When

- Single-file scripts or prototypes
- Learning Go fundamentals
- Very simple APIs (<5 endpoints)
- Extreme performance requirements (1M+ requests/sec)

---

## Framework Comparison

| Aspect | Lokstra | Echo | Gin | Spring Boot |
|--------|---------|------|-----|------------|
| **Language** | Go | Go | Go | Java |
| **DI** | Full Framework | Optional | Optional | Built-in |
| **Code Gen** | Annotations ✅ | Manual | Manual | Annotations |
| **Type Safety** | Excellent | Good | Good | Good |
| **Learning Curve** | Moderate | Easy | Easy | Steep |
| **Enterprise Ready** | ✅ | Good | Good | ✅ |

---

## Common Questions

**Q: Do I need to use all 14 skills?**  
A: For a typical project: Yes. Design phase → Implementation phase → Advanced phase. Optional: Skip advanced if no tests needed.

**Q: Can I skip the design phase?**  
A: Not recommended. Design phase prevents implementation errors. Estimated 10x productivity gain.

**Q: How long does a complete project take?**  
A: Small module: 4-6 hours. Medium module: 1-2 days. Large module: 3-5 days.

**Q: Do I need to use all 6 implementation skills?**  
A: Yes. They're sequential: init → config → handlers → services → migrations → tests.

**Q: Can multiple teams work in parallel?**  
A: Yes. After design phase, each team works on separate modules (bounded contexts).

---

## Summary

Lokstra Framework combines:
- ✅ **Strong typing** from Go
- ✅ **Elegant architecture** from Spring Boot/NestJS
- ✅ **Rapid development** from annotation-based code generation
- ✅ **Production readiness** from dependency injection and modular design

With the 14 AI skills, you can go from business requirement to production-ready application in days instead of weeks. 
