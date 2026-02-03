# Lokstra Skills Roadmap

**Purpose:** Plan skill restructuring for maximum efficiency and clarity.

## Current Status (Phase 1 - COMPLETED) + Phase 2 (COMPLETED) + Phase 3 (COMPLETED)

✅ **Phase 1 Skills (Completed - Jan 28, 2026):**
1. `design-lokstra-overview` - Framework fundamentals
2. `design-lokstra-brd-generation` - Business Requirements Document
3. `design-lokstra-module-requirements` - Module breakdown (DDD)
4. `design-lokstra-api-specification` - API endpoint specs
5. `design-lokstra-schema-design` - Database schema design

✅ **Phase 2 Skills (Completed - Jan 29, 2026):**
1. `implementation-lokstra-init-framework` - Initialize main.go
2. `implementation-lokstra-yaml-config` - Configuration management
3. `implementation-lokstra-create-handler` - @Handler endpoint creation
4. `implementation-lokstra-create-service` - @Service infrastructure
5. `implementation-lokstra-create-migrations` - Database migrations
6. `implementation-lokstra-generate-http-files` - HTTP test files

✅ **Phase 3 Skills (Completed - Jan 29, 2026):**
1. `advanced-lokstra-tests` - Unit & integration testing
2. `advanced-lokstra-middleware` - Custom middleware creation
3. `advanced-lokstra-validate-consistency` - Validation & consistency checks

**Naming Convention:**
- Folder: `design-lokstra-overview` (human-readable, easy CLI filtering)
- YAML name: `lokstra-overview` (clean agent discovery)
- Metadata: `phase: design|implementation|advanced`, `order: 1-N`

## Phase 2: Implementation Skills (COMPLETED)

**Status:** All 6 micro-skills created and fully documented (Jan 29, 2026)

**Benefits Realized:**
- ✅ Token Efficiency: ~500 tokens each vs ~2800 monolithic
- ✅ Parallel Work: Multiple agents can work on different skills
- ✅ Clear Intent: Focused skills for specific tasks
- ✅ Reusability: Each skill can be used independently
- ✅ Easier Maintenance: Update one skill = isolated changes

### Implementation Micro-Skills (Phase 2) - COMPLETED

| Folder | YAML name | Purpose | Output | Token Budget |
|--------|-----------|---------|--------|--------------|
| `implementation-lokstra-init-framework` | `lokstra-init-framework` | Initialize main.go, lokstra.Bootstrap() | main.go, go.mod | ~400 |
| `implementation-lokstra-yaml-config` | `lokstra-yaml-config` | Create configs/, multi-file YAML setup | config.yaml files | ~500 |
| `implementation-lokstra-create-handler` | `lokstra-create-handler` | Create @Handler with @Route | handler files | ~600 |
| `implementation-lokstra-create-service` | `lokstra-create-service` | Create @Service for infrastructure | service files | ~500 |
| `implementation-lokstra-create-migrations` | `lokstra-create-migrations` | Create UP/DOWN migration files | migration SQL files | ~400 |
| `implementation-lokstra-generate-http-files` | `lokstra-generate-http-files` | Generate .http client files | .http files | ~300 |

### Advanced Skills (Phase 3) - COMPLETED

| Folder | YAML name | Purpose | Output | Status |
|--------|-----------|---------|--------|--------|
| `advanced-lokstra-tests` | `lokstra-tests` | Unit & integration tests | test files | ✅ |
| `advanced-lokstra-middleware` | `lokstra-middleware` | Custom middleware creation | middleware files | ✅ |
| `advanced-lokstra-validate-consistency` | `lokstra-validate-consistency` | Check circular deps, validate config | validation report | ✅ |

## Skill Selection Logic

```
Developer needs to implement module:
│
├─ "I need to understand the framework"
│  └─ Use: design-lokstra-overview
│
├─ "I need to create BRD"
│  └─ Use: design-lokstra-brd-generation
│
├─ "I need to break BRD into modules"
│  └─ Use: design-lokstra-module-requirements
│
├─ "I need to design APIs"
│  └─ Use: design-lokstra-api-specification
│
├─ "I need to design database"
│  └─ Use: design-lokstra-schema-design
│
├─ "I need to init the project"
│  └─ Use: implementation-lokstra-init-framework
│
├─ "I need to set up configs"
│  └─ Use: implementation-lokstra-yaml-config
│
├─ "I need to create handlers"
│  └─ Use: implementation-lokstra-create-handler
│
├─ "I need to create services"
│  └─ Use: implementation-lokstra-create-service
│
├─ "I need to create migrations"
│  └─ Use: implementation-lokstra-create-migrations
│
├─ "I need .http files for testing"
│  └─ Use: implementation-lokstra-generate-http-files
│
├─ "I need unit tests"
│  └─ Use: advanced-lokstra-tests
│
├─ "I need custom middleware"
│  └─ Use: advanced-lokstra-middleware
│
└─ "Validate my code"
   └─ Use: advanced-lokstra-validate-consistency
```

## Benefits

✅ **Token Efficiency:** Load only ~500 tokens vs ~2800 tokens  
✅ **Parallel Work:** Multiple agents work on different skills  
✅ **Clear Intent:** "Create handler" is unambiguous  
✅ **Reusability:** Handler skill standalone useful  
✅ **Easier Maintenance:** Change handler creation = update 1 skill  

## Relationship to Existing Skills

```
design-lokstra-overview
    ↓
design-lokstra-brd-generation (BRD v1.0)
    ↓
design-lokstra-module-requirements (identify modules)
    ↓
design-lokstra-api-specification (define endpoints)
    ↓
design-lokstra-schema-design (define database)
    ↓
IMPLEMENTATION PHASE (using micro-skills):
    ├─ implementation-lokstra-init-framework (main.go)
    ├─ implementation-lokstra-yaml-config (configs/)
    ├─ implementation-lokstra-create-handler (@Handler)
    ├─ implementation-lokstra-create-service (@Service)
    ├─ implementation-lokstra-create-migrations (migrations/)
    └─ implementation-lokstra-generate-http-files (.http)
    ↓
ADVANCED PHASE:
    ├─ advanced-lokstra-tests (unit & integration)
    ├─ advanced-lokstra-middleware (custom middleware)
    └─ advanced-lokstra-validate-consistency (validation)
```

## Implementation Priority

**Status: ALL PHASES COMPLETE ✅**

1. **Phase 1: Design (COMPLETE)** 
   - 5 skills for business analysis and architecture

2. **Phase 2: Implementation (COMPLETE)**
   - 6 micro-skills for efficient code generation

3. **Phase 3: Advanced (COMPLETE)**
   - 3 skills for testing, middleware, and validation

## Notes

- Micro-skills replace SKILL 4-7 from `lokstra-code-implementation`
- `lokstra-code-advanced` handles SKILL 8-13 (tests, validation, deployment)
- `lokstra-implementation-flow` provides orchestration guide
- Each micro-skill can be used independently or as sequence
