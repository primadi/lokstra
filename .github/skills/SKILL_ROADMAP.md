# Lokstra Skills Roadmap

**Purpose:** Plan skill restructuring for maximum efficiency and clarity.

## Current Status (Phase 1 - COMPLETED)

✅ **Phase 1 Skills (Completed & Renamed with design- prefix):**
1. `design-lokstra-overview` - Framework fundamentals
2. `design-lokstra-brd-generation` - Business Requirements Document
3. `design-lokstra-module-requirements` - Module breakdown (DDD)
4. `design-lokstra-api-specification` - API endpoint specs
5. `design-lokstra-schema-design` - Database schema design

**Naming Convention:**
- Folder: `design-lokstra-overview` (human-readable, easy CLI filtering)
- YAML name: `lokstra-overview` (clean agent discovery)
- Metadata: `phase: design`, `order: 1-5`

## Phase 2: Split Implementation Skills

**Problem:** `lokstra-code-implementation` (~2800 tokens) + `lokstra-code-advanced` (~2200 tokens) = Too large, mixed concerns.

**Solution:** Break into focused micro-skills (~500-800 tokens each) with `implementation-` prefix.

### Implementation Micro-Skills (Phase 2)

| Folder | YAML name | Purpose | Output | Token Budget |
|--------|-----------|---------|--------|--------------|
| `implementation-lokstra-init-framework` | `lokstra-init-framework` | Initialize main.go, lokstra.Bootstrap() | main.go, go.mod | ~400 |
| `implementation-lokstra-yaml-config` | `lokstra-yaml-config` | Create configs/, multi-file YAML setup | config.yaml files | ~500 |
| `implementation-lokstra-create-handler` | `lokstra-create-handler` | Create @Handler with @Route | handler files | ~600 |
| `implementation-lokstra-create-service` | `lokstra-create-service` | Create @Service for infrastructure | service files | ~500 |
| `implementation-lokstra-create-migrations` | `lokstra-create-migrations` | Create UP/DOWN migration files | migration SQL files | ~400 |
| `implementation-lokstra-generate-http-files` | `lokstra-generate-http-files` | Generate .http client files | .http files | ~300 |

### Advanced Skills (Phase 3)

| Folder | YAML name | Purpose | Output | Token Budget |
|--------|-----------|---------|--------|--------------|
| `advanced-lokstra-tests` | `lokstra-tests` | Unit & integration tests | test files | ~800 |
| `advanced-lokstra-middleware` | `lokstra-middleware` | Custom middleware creation | middleware files | ~500 |
| `advanced-lokstra-validate-consistency` | `lokstra-validate-consistency` | Check circular deps, validate config | validation report | ~400 |

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

1. **High Priority (Start next):**
   - `lokstra-init-framework` - Foundational
   - `lokstra-yaml-config` - Required early
   - `lokstra-create-handler` - Most common use case
   - `lokstra-create-service` - High frequency

2. **Medium Priority:**
   - `lokstra-create-migrations` - Data layer
   - `lokstra-generate-http-files` - Testing/validation
   - `lokstra-tests` - Quality

3. **Lower Priority:**
   - `lokstra-middleware` - Advanced feature
   - `lokstra-validate-consistency` - Final check

## Notes

- Micro-skills replace SKILL 4-7 from `lokstra-code-implementation`
- `lokstra-code-advanced` handles SKILL 8-13 (tests, validation, deployment)
- `lokstra-implementation-flow` provides orchestration guide
- Each micro-skill can be used independently or as sequence
