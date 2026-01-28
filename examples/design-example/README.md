# Design-First Development Example

**Purpose:** Demonstrate complete document-driven development workflow using Lokstra Framework.

**Case Study:** E-Commerce Order Management System

---

## Overview

This example shows the complete design-first workflow:

```
BRD → Module Requirements → API Specifications → Database Schema → Implementation
```

**Business Case:** Build an order management system to handle 10,000+ daily orders with real-time inventory sync and payment processing.

---

## Document Workflow

### Phase 1: Business Requirements (SKILL 0)
**File:** [docs/BRD.md](docs/BRD.md)

Defines:
- Business objectives and success metrics
- Stakeholders and their roles
- Functional and non-functional requirements
- Scope boundaries
- Risks and constraints

**Status:** ✅ Complete

---

### Phase 2: Module Requirements (SKILL 1)
**Module Breakdown (Bounded Contexts):**

1. **Auth Module** - [docs/modules/auth/REQUIREMENTS.md](docs/modules/auth/REQUIREMENTS.md)
   - User registration and login
   - JWT token management
   - Password reset
   - Role-based access control

2. **Product Module** - [docs/modules/product/REQUIREMENTS.md](docs/modules/product/REQUIREMENTS.md)
   - Product catalog management
   - Categories and search
   - Inventory tracking

3. **Order Module** - [docs/modules/order/REQUIREMENTS.md](docs/modules/order/REQUIREMENTS.md)
   - Order creation and tracking
   - Status management
   - Order history

**Status:** ✅ Complete

---

### Phase 3: API Specifications (SKILL 2)

For each module:
- **Auth:** [docs/modules/auth/API_SPEC.md](docs/modules/auth/API_SPEC.md)
- **Product:** [docs/modules/product/API_SPEC.md](docs/modules/product/API_SPEC.md)
- **Order:** [docs/modules/order/API_SPEC.md](docs/modules/order/API_SPEC.md)

Each spec includes:
- Complete endpoint definitions (HTTP method, path, auth)
- Request/response schemas
- Validation rules
- Error responses (400, 401, 403, 404, 500)
- Examples

**Status:** ✅ Complete

---

### Phase 4: Database Schema (SKILL 3)

For each module:
- **Auth:** [docs/modules/auth/SCHEMA.md](docs/modules/auth/SCHEMA.md)
- **Product:** [docs/modules/product/SCHEMA.md](docs/modules/product/SCHEMA.md)
- **Order:** [docs/modules/order/SCHEMA.md](docs/modules/order/SCHEMA.md)

Each schema includes:
- Table definitions with constraints
- Indexes for performance
- Foreign key relationships
- Triggers and stored procedures
- Migration files

**Status:** ✅ Complete

---

### Phase 5: Implementation (SKILL 4-13)

**Status:** ⏳ Ready for code generation

Once all documents are approved, AI agents can generate:
1. Folder structure (`modules/auth/`, `modules/product/`, `modules/order/`)
2. Database migrations
3. Domain models and DTOs
4. Repositories (data access layer)
5. Handlers (HTTP endpoints with `@Handler` annotations)
6. Configuration (`config.yaml`)
7. Unit and integration tests

---

## Key Insights from This Example

### 1. Design-First Prevents Rework
- All requirements documented upfront
- Stakeholder approval before coding
- No "build-then-fix" cycles

### 2. Bounded Contexts Keep Code Maintainable
- Auth, Product, Order = separate modules
- Each can be developed independently
- Clear ownership and boundaries

### 3. Specifications Enable Consistency
- API spec ensures all endpoints follow same patterns
- Schema ensures data integrity
- Tests can be written from specs before implementation

### 4. Documentation is Compliance
- BRD = Business alignment
- Module requirements = Functional specs
- API/Schema docs = Technical specs
- All versioned and immutable once approved

---

## Using This Example with AI Agents

### To Generate Code from These Specs:

1. **Read the skills:**
   ```
   .github/skills/00-lokstra-overview.md (framework overview)
   .github/skills/05-implementation.md (SKILL 4-7)
   .github/skills/06-implementation-advanced.md (SKILL 8-13)
   ```

2. **Follow the specs:**
   - Generate domain models from SCHEMA.md
   - Generate repositories from REQUIREMENTS.md
   - Generate handlers from API_SPEC.md
   - Generate config from module dependencies

3. **Run code generation:**
   ```bash
   # Create new project with this design
   lokstra new ecommerce-system -template 02_app_framework/04_design_first_example
   
   # Or apply to existing project
   cd myproject
   lokstra autogen
   ```

---

## Benefits of This Approach

### For Development Teams:
- ✅ Clear requirements before coding
- ✅ Reduced miscommunication
- ✅ Parallel development (backend/frontend)
- ✅ Test-driven development from specs

### For Stakeholders:
- ✅ Visibility into what's being built
- ✅ Early feedback opportunity
- ✅ Compliance documentation
- ✅ Version control for requirements

### For AI Agents:
- ✅ Complete context for code generation
- ✅ Validation rules defined upfront
- ✅ Consistent patterns across modules
- ✅ Testable requirements

---

## Next Steps

1. Review all documents in `docs/` folder
2. Get stakeholder approval on BRD and requirements
3. Use AI agents to generate implementation
4. Run tests to verify specs are met
5. Deploy with confidence

---

**See Also:**
- [Lokstra AI Skills Guide](../../.github/skills/README.md)
- [Document Templates](../../docs/templates/)
- [Framework Documentation](https://primadi.github.io/lokstra/)
