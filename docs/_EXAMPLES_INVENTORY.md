# Examples Inventory & Migration Plan

This document tracks existing examples and their mapping to the new documentation structure.

---

## 📂 Existing Example Folders

### `/cmd_draft/examples/` (Main examples folder)
```
01-basic-router/
02-basic-app/
03-basic-server/
04-services/
05-yaml-monolith/
06-microservices/
07-hybrid/
08-lokstra-handler/
09-api-standard/
10-response-patterns/
11-two-layer-responses/
12-custom-formatters/
13-router-integration/
14-client-response-parsing/
15-custom-formatter-parsing/
16-service-usage/
17-auth-system/
18-bind-samples/
18-service-router/
19-flexible-handlers/
20-service-router-struct-based/
21-type-alias-context/
22-all-handler-forms/
23-service-local-remote/
25-single-binary-deployment/
26-remote-service-path-detection/
response-return-types/
return-type-any-only/
reverse-proxy-gateway/
test-port-error/
```

### `/cmd_draft/examples2/` (New organized examples)
```
A1_core_overview/
├── 01_basic_router/
└── 02_app_server/
```

---

## 🗺️ Migration Mapping

### 01-essentials/01-router/examples/

| New Example | Source | Status | Notes |
|-------------|--------|--------|-------|
| `01-basic-routes/` | `examples/01-basic-router/` | ✅ Exists | Review & simplify |
| `02-route-parameters/` | NEW | ❌ Create | Path params, query params |
| `03-route-groups/` | NEW | ❌ Create | API versioning example |
| `04-with-middleware/` | Part of `08-lokstra-handler/` | ⚠️ Extract | Focus on middleware application |
| `05-complete-api/` | NEW | ❌ Create | Mini CRUD API |

---

### 01-essentials/02-service/examples/

| New Example | Source | Status | Notes |
|-------------|--------|--------|-------|
| `01-simple-service/` | `examples/04-services/` | ✅ Exists | Review & simplify |
| `02-service-in-handler/` | `examples/16-service-usage/` | ✅ Exists | May need cleanup |
| `03-service-as-router/` | `examples/18-service-router/` | ✅ Exists | CRITICAL: Auto HTTP routing |
| `04-service-dependencies/` | NEW | ❌ Create | Service DI example |

**IMPORTANT**: Example `03-service-as-router` demonstrates the **killer feature** - must be in Essentials!

---

### 01-essentials/03-middleware/examples/

| New Example | Source | Status | Notes |
|-------------|--------|--------|-------|
| `01-logging/` | Part of examples | ⚠️ Extract | Simple logging middleware |
| `02-authentication/` | `examples/17-auth-system/` | ✅ Exists | Simplify for essentials |
| `03-cors/` | NEW | ❌ Create | CORS configuration |

---

### 01-essentials/04-configuration/examples/

| New Example | Source | Status | Notes |
|-------------|--------|--------|-------|
| `01-json-config/` | NEW | ❌ Create | Basic JSON config |
| `02-yaml-config/` | `examples/05-yaml-monolith/` | ✅ Exists | Simplify |
| `03-env-variables/` | NEW | ❌ Create | Environment-based config |

---

### 01-essentials/05-app-and-server/examples/

| New Example | Source | Status | Notes |
|-------------|--------|--------|-------|
| `01-basic-app/` | `examples/02-basic-app/` | ✅ Exists | Review |
| `02-graceful-shutdown/` | Part of `03-basic-server/` | ⚠️ Extract | Focus on shutdown |

---

### 01-essentials/06-putting-it-together/examples/

| New Example | Source | Status | Notes |
|-------------|--------|--------|-------|
| `todo-api/` | NEW | ❌ Create | Complete REST API with all concepts |

---

## 🚀 02-deep-dive/router/examples/

| New Example | Source | Status | Notes |
|-------------|--------|--------|-------|
| `all-handler-forms/` | `examples/22-all-handler-forms/` | ✅ Exists | Already complete! 29 forms |
| `router-lifecycle/` | NEW | ❌ Create | Build(), Walk(), PrintRoutes() |
| `prefix-matching/` | NEW | ❌ Create | ANYPrefix, GETPrefix, etc |
| `router-chaining/` | NEW | ❌ Create | SetNextChain example |
| `service-as-router-advanced/` | `examples/20-service-router-struct-based/` | ✅ Exists | Advanced patterns |
| `custom-engines/` | NEW | ❌ Create | Using different router engines |

---

## 🚀 02-deep-dive/service/examples/

| New Example | Source | Status | Notes |
|-------------|--------|--------|-------|
| `dependency-injection/` | NEW | ❌ Create | DI patterns |
| `service-lifecycle/` | NEW | ❌ Create | Eager vs lazy |
| `remote-services/` | `examples/23-service-local-remote/` | ✅ Exists | Local vs remote |
| `layered-architecture/` | NEW | ❌ Create | Clean architecture |
| `service-conventions/` | Part of `25-single-binary-deployment/` | ⚠️ Extract | Convention system |

---

## 🚀 02-deep-dive/configuration/examples/

| New Example | Source | Status | Notes |
|-------------|--------|--------|-------|
| `multi-deployment/` | `examples/25-single-binary-deployment/` | ✅ Exists | **KILLER FEATURE!** |
| `config-merge/` | NEW | ❌ Create | Multiple file merging |
| `custom-resolvers/` | NEW | ❌ Create | Custom config resolvers |
| `config-validation/` | NEW | ❌ Create | Schema validation |

**CRITICAL**: Example `25-single-binary-deployment` is the showcase for multi-deployment!

---

## 🚀 02-deep-dive/middleware/examples/

| New Example | Source | Status | Notes |
|-------------|--------|--------|-------|
| `custom-middleware/` | NEW | ❌ Create | Creating middleware |
| `rate-limiting/` | NEW | ❌ Create | Rate limit example |
| `request-transformation/` | NEW | ❌ Create | Transform requests |
| `middleware-factory/` | NEW | ❌ Create | Factory pattern |

---

## 📊 05-examples/ (Complete Applications)

| New Example | Source | Status | Notes |
|-------------|--------|--------|-------|
| `blog-api/` | NEW | ❌ Create | CRUD + auth + pagination |
| `e-commerce/` | NEW | ❌ Create | Complex business logic |
| `microservices/` | `examples/06-microservices/` | ✅ Exists | Review & enhance |
| `api-gateway/` | `examples/reverse-proxy-gateway/` | ✅ Exists | Reverse proxy pattern |
| `single-binary/` | `examples/25-single-binary-deployment/` | ✅ Exists | Multi-deployment showcase |

---

## 📋 Statistics

### Existing Examples Available:
- ✅ **Ready to use**: 12 examples
- ⚠️ **Need modification**: 6 examples
- ❌ **Need to create**: 25 examples

### Priority Examples to Create:

**HIGH PRIORITY** (Essentials):
1. `01-essentials/01-router/examples/02-route-parameters/`
2. `01-essentials/01-router/examples/03-route-groups/`
3. `01-essentials/01-router/examples/05-complete-api/`
4. `01-essentials/02-service/examples/04-service-dependencies/`
5. `01-essentials/04-configuration/examples/01-json-config/`
6. `01-essentials/04-configuration/examples/03-env-variables/`
7. `01-essentials/06-putting-it-together/examples/todo-api/` ← **MOST IMPORTANT**

**MEDIUM PRIORITY** (Deep Dive):
- Router lifecycle examples
- Service DI patterns
- Custom middleware examples

**LOW PRIORITY** (Nice to have):
- Complete applications (blog, e-commerce)
- Advanced configuration patterns

---

## 🎯 Action Items

### Phase 1: Review Existing Examples
- [ ] Review `22-all-handler-forms` - use as is or simplify?
- [ ] Review `25-single-binary-deployment` - essential showcase
- [ ] Review `18-service-router` - move to Essentials
- [ ] Review `23-service-local-remote` - for Deep Dive
- [ ] Extract middleware examples from existing code

### Phase 2: Create Essential Examples
- [ ] Router: Path parameters demo
- [ ] Router: Route groups demo
- [ ] Router: Complete mini API
- [ ] Service: Dependencies demo
- [ ] Configuration: Basic examples
- [ ] **Todo API - complete project** ← Top priority!

### Phase 3: Create Deep Dive Examples
- [ ] All advanced router patterns
- [ ] DI patterns
- [ ] Custom middleware
- [ ] Advanced configuration

### Phase 4: Complete Applications
- [ ] Blog API
- [ ] E-commerce backend
- [ ] Refine microservices example

---

## 📝 Notes for Migration

### Keep in Mind:
1. **Simplify for Essentials** - Focus on one concept per example
2. **Enhance for Deep Dive** - Show advanced patterns
3. **Preserve Working Examples** - Don't break existing code
4. **Add READMEs** - Every example needs clear documentation
5. **Test All Examples** - Ensure they actually run

### Naming Convention:
```
examples/
├── 01-basic-feature/       # Number + descriptive name
│   ├── main.go
│   ├── README.md          # What it demonstrates
│   └── test.http          # Optional: test requests
```

### README Template for Examples:
```markdown
# Example Name

## What This Demonstrates
- Concept 1
- Concept 2

## Running
\`\`\`bash
go run main.go
\`\`\`

## Testing
\`\`\`bash
curl http://localhost:3000/endpoint
\`\`\`

## Key Points
- Important takeaway 1
- Important takeaway 2

## Related
- Link to concept documentation
- Link to other examples
```

---

## 🎯 Success Criteria

An example is "done" when:
- ✅ Code runs without errors
- ✅ Demonstrates ONE clear concept
- ✅ Has README explaining what/why/how
- ✅ Has test commands (curl or test.http)
- ✅ Is referenced in relevant documentation
- ✅ Uses best practices
- ✅ Is simple enough for target audience (Essential vs Deep Dive)

---

**Next Steps**: Start with Phase 1 - Review existing examples and identify what can be reused directly.
