# How to Develop with Lokstra Framework

**Welcome to AI-Driven Development with Lokstra!**

This project is set up for **design-first development** using AI agents (GitHub Copilot, Cursor, etc.).

## 🏗️ Architecture Overview

**This template supports multi-tenant SaaS applications** with a shared database model:
- Single database shared across all tenants
- Tenant isolation via `tenant_id` column
- Support for different database connections per tenant (via dbpool_pg service)
- Flexible deployment: single tenant or multi-tenant

**Single Tenant Mode:** If you only need one tenant, you can skip the `tenant-registration` module and simplify the `dbpool_pg` configuration to use a single connection.

---

## 🚀 Quick Start

### Step 1: Install Lokstra CLI & Create Project

**1. Install Lokstra CLI:**
```bash
go install github.com/primadi/lokstra/cmd/lokstra@latest
```

**2. Create new Lokstra project:**
```bash
lokstra new clinic-app
```

When prompted, select: **`03_ai_driven/01_starter`** template

This creates a project structure with AI-driven development setup.

**3. Enter the project directory:**
```bash
cd clinic-app
```

---

### Step 2: Setup Project Dependencies & Skills

**Run these commands in the project directory:**

```bash
# Download latest AI skills
lokstra update-skills

# Tidy Go dependencies
go mod tidy
```

After this, your project is ready with:
- ✅ 14 AI skills loaded (design, implementation, advanced phases)
- ✅ Go dependencies configured
- ✅ Project structure ready

---

### Step 3: Create Database & Update Configuration

**1. Create your database:**
```bash
# PostgreSQL example
createdb clinic_app
```

**2. Update `config.yaml` with database connections:**
```yaml
service-definitions:
   db_master:  # Master database service for migrations
      type: dbpool_pg
      config:
         dsn: postgres://postgres:password@localhost:5432/clinic_app
         schema: master
   db_main:    # Main database service for application
      type: dbpool_pg
      config:
         dsn: postgres://postgres:password@localhost:5432/clinic_app
         schema: main

   # Optional: Additional tenant connections for multi-tenant per-DB deployment
   # db_tenant_01:
   #    type: dbpool_pg
   #    config:
   #       dsn: postgres://tenant1_user:password@tenant1.example.com:5432/tenant1_db
   #       schema: main
```

**Notes:**
- `db_master`: Used for running migrations (setup & updates)
- `db_main`: Used by application at runtime
- Can point to same database (recommended for single tenant) or different databases (for multi-tenant with separate DBs per tenant)
- For single tenant: Use same credentials for both `db_master` and `db_main`
- For multi-tenant with shared DB: Use same connection for both, add `tenant_id` to schemas

**3. Verify configuration:**
```bash
go run . --generate-only
```

---

### Step 4: Understand Your Business Needs

Before writing any code, clearly define:
- What problem are you solving?
- Who are the users?
- What are the main features?
- What are the success metrics?

---

### Step 5: Generate Business Requirements Document (BRD)

**Tell your AI agent:**

> "I want to build [describe your app]. Help me create a comprehensive Business Requirements Document."

**AI will:**
1. Ask clarifying questions about your business needs
2. Generate `docs/BRD.md` with:
   - Executive summary
   - Business objectives & metrics
   - Stakeholders
   - Functional requirements
   - Non-functional requirements
   - Scope & constraints

**Reference:** [.github/skills/design-lokstra-brd-generation/](../.github/skills/design-lokstra-brd-generation/)

---

### ✅ Review & Approve BRD Checklist

**Before proceeding to Step 6, review the generated BRD:**

| Review Item | Check | Questions to Ask |
|---|---|---|
| **Business Goals** | Clear? | Are success metrics measurable? Do goals align with business strategy? |
| **User Needs** | Complete? | Have all user types been considered? Are pain points addressed? |
| **Features** | Realistic? | Are all requested features necessary? Are priorities clear? |
| **Scope** | Defined? | What's included? What's excluded? Are boundaries clear? |
| **Constraints** | Documented? | Budget? Timeline? Technology? Legal/compliance requirements? |
| **Success Metrics** | Measurable? | How will you know the project succeeded? |

**How to Review:**

1. **Read the entire BRD** - Understand the full picture
2. **Check for completeness** - All sections filled in?
3. **Validate accuracy** - Does it match your business vision?
4. **Identify gaps** - Missing features or requirements?
5. **Confirm priorities** - Are high-value features clear?

**How to Provide Feedback:**

If changes needed, tell your AI agent:

> "Review the BRD: [specific feedback]. Update docs/BRD.md section '[section-name]' with: [changes]."

**Example:**
> "The payment methods section needs expansion. Add: credit card, PayPal, bank transfer with descriptions."

**How to Approve:**

1. **Gather stakeholders** - Product manager, business owner, technical lead
2. **Review together** - Walk through each section
3. **Collect feedback** - Ask clarifying questions
4. **Make adjustments** - Tell AI agent your changes
5. **Final sign-off** - All stakeholders approve

**Example approval message:**
> "AI, the BRD review is complete. All stakeholders approve the document. We can proceed to Step 6."

**When to Approve:**

✅ Proceed to Step 6 when:
- All stakeholders agree on business goals
- Feature set is clear and prioritized
- Scope is well-defined
- Timeline and budget are realistic
- Success metrics are measurable

---

---

### Step 6: Generate Module Requirements

**Tell your AI agent:**

> "Based on the BRD, help me identify bounded contexts and generate module requirements for the [module-name] module."

**AI will:**
1. Analyze BRD to identify modules (bounded contexts)
2. Generate `docs/modules/{module-name}/REQUIREMENTS.md` for each module
3. Define module dependencies

**Example modules:**
- Auth (authentication & authorization)
- Product (product catalog)
- Order (order management)
- Payment (payment processing)

**Reference:** [.github/skills/design-lokstra-module-requirements/](../.github/skills/design-lokstra-module-requirements/)

**Multi-Tenant Modules:** Include `tenant-registration` module for tenant management, or skip if single-tenant

---

### ✅ Review & Approve Module Requirements Checklist

**Before proceeding to Step 7, review each module's requirements:**

| Review Item | Check | Questions to Ask |
|---|---|---|
| **Module Scope** | Clear? | What is this module responsible for? |
| **Features** | Complete? | All required features included? Any missing from BRD? |
| **Use Cases** | Realistic? | Can users achieve their goals? Are flows logical? |
| **Acceptance Criteria** | Measurable? | How will you test if module works? |
| **Dependencies** | Clear? | Which modules does this depend on? |
| **Data Entities** | Identified? | What data does this module manage? |

**How to Review:**

1. **Check bounded contexts** - Is each module cohesive? (Related features together)
2. **Validate feature mapping** - Are all BRD features assigned to modules?
3. **Review dependencies** - Do they make sense? Are there circular deps?
4. **Confirm use cases** - Are all user journeys covered?
5. **Verify acceptance criteria** - Are they testable?

**How to Provide Feedback:**

> "Update [module-name] requirements: Add/remove feature '[feature]' because [reason]."

**Example:**
> "Update Product module requirements: Add inventory tracking feature because the BRD mentions stock management."

**How to Approve:**

1. **Technical review** - Tech lead reviews module boundaries
2. **Dependencies check** - Architect verifies no circular dependencies
3. **Feature mapping** - Confirm all BRD features are covered
4. **Acceptance criteria** - QA lead verifies they can test it
5. **Team sign-off** - Dev team agrees on design

**Example approval message:**
> "AI, all module requirements are reviewed and approved. Module boundaries are clear, no circular dependencies found. Proceed to Step 7."

**When to Approve:**

✅ Proceed to Step 7 when:
- All BRD features are assigned to modules
- Module boundaries are clear
- Dependencies are documented
- Use cases are complete
- Team agrees on module design

---

---

### Step 7: Generate API Specifications

**Tell your AI agent:**

> "Generate comprehensive API specifications for the [module-name] module with endpoints, schemas, validation rules, and examples."

**AI will:**
1. Generate `docs/modules/{module-name}/API_SPEC.md`
2. Define all endpoints with:
   - HTTP method & path
   - Request/response schemas
   - Validation rules
   - Error responses
   - Examples

**Reference:** [.github/skills/design-lokstra-api-specification/](../.github/skills/design-lokstra-api-specification/)

---

### ✅ Review & Approve API Specs Checklist

**Before proceeding to Step 8, review the API specification:**

| Review Item | Check | Questions to Ask |
|---|---|---|
| **Endpoints** | Complete? | Are all module features accessible via API? |
| **HTTP Methods** | Correct? | Is RESTful pattern followed? |
| **Paths** | Logical? | Are URLs intuitive and consistent? |
| **Request Schema** | Clear? | What parameters are required? Type/format correct? |
| **Response Schema** | Documented? | What data is returned? Error responses defined? |
| **Validation Rules** | Specified? | Min/max lengths? Required fields? Formats (email, date)? |
| **Error Handling** | Complete? | All error codes documented? Messages clear? |
| **Security** | Considered? | Authentication required? Authorization checks? |

**How to Review:**

1. **Test the examples** - Can you call the API with provided examples?
2. **Check completeness** - Are all module features covered?
3. **Verify consistency** - Do endpoints follow same patterns?
4. **Review errors** - Are error codes clear and helpful?
5. **Confirm data types** - Request/response schemas match business logic?

**How to Provide Feedback:**

> "Update API spec for [module]: [specific change] because [reason]."

**Example:**
> "Update the GET /products endpoint: Add pagination with 'limit' and 'offset' parameters because the BRD mentions handling product catalogs with thousands of items."

**How to Approve:**

1. **Frontend review** - Frontend team confirms endpoints work for their needs
2. **Schema validation** - Check if response schemas match database design
3. **Test examples** - Actually test the API examples provided
4. **Error scenarios** - Confirm error codes are correct
5. **API consistency** - Verify all endpoints follow same patterns

**Example approval message:**
> "API spec reviewed and approved. Frontend team tested all examples, error handling is clear, schemas are consistent. Approved for Step 8."

**When to Approve:**

✅ Proceed to Step 8 when:
- All module features are represented as endpoints
- Request/response schemas are clear
- Validation rules are documented
- Error handling is comprehensive
- Frontend/client team confirms the spec works for them

---

---

### Step 8: Generate Database Schema

**Tell your AI agent:**

> "Design the database schema for the [module-name] module. Include tables, indexes, constraints, and migration files."

**AI will:**
1. Generate `docs/modules/{module-name}/SCHEMA.md`
2. Define tables, indexes, constraints
3. Create migration files in `migrations/{module-name}/`

**Reference:** [.github/skills/design-lokstra-schema-design/](../.github/skills/design-lokstra-schema-design/)

**Multi-Tenant Consideration:** Add `tenant_id` columns to all tables for tenant isolation (if multi-tenant)

---

### ✅ Review & Approve Database Schema Checklist

**Before proceeding to Step 9, review the database schema:**

| Review Item | Check | Questions to Ask |
|---|---|---|
| **Tables** | Complete? | Does schema match API response schemas? |
| **Columns** | Correct? | Data types appropriate? Nullability correct? |
| **Primary Keys** | Defined? | Every table has a PK? |
| **Foreign Keys** | Logical? | Relationships between tables clear? |
| **Indexes** | Optimized? | Performance-critical columns indexed? |
| **Constraints** | Enforced? | Unique constraints? Check constraints? |
| **Tenant Isolation** | Implemented? | tenant_id in multi-tenant tables? |

**How to Review:**

1. **Map to API specs** - Do tables match response schemas?
2. **Check relationships** - Are foreign keys correct?
3. **Verify data integrity** - Are constraints sufficient?
4. **Review performance** - Are indexes on right columns?
5. **Confirm isolation** - Multi-tenant: Is tenant_id everywhere needed?

**Example Check:**
```sql
-- Good: tenant_id included in multi-tenant table
CREATE TABLE products (
   id UUID PRIMARY KEY,
   tenant_id UUID NOT NULL,
   name VARCHAR(255) NOT NULL,
   created_at TIMESTAMP DEFAULT NOW(),
   FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
CREATE INDEX idx_products_tenant ON products(tenant_id);
```

**How to Provide Feedback:**

> "Update [module] schema: Add [table/column/index] because [reason]."

**Example:**
> "Update Product schema: Add created_by_user_id column because the API spec requires tracking who created each product."

**How to Approve:**

1. **DBA review** - Database administrator checks performance & indexes
2. **API mapping** - Verify schema matches API response structure
3. **Data integrity** - Confirm constraints are sufficient
4. **Multi-tenant check** - Verify tenant_id isolation (if applicable)
5. **Migration feasibility** - Check if migrations are reversible

**Example approval message:**
> "Database schema reviewed and approved by DBA. All relationships properly defined, indexes are optimized, tenant isolation is correct. Ready for migrations in Step 9."

**When to Approve:**

✅ Proceed to Step 9 when:
- Schema matches API response structure
- All relationships are defined
- Indexes are on performance-critical columns
- Constraints enforce business rules
- Multi-tenant isolation is clear (if applicable)
- DBA has reviewed for performance concerns

---

---

### Step 9: Create Database Migrations

**Tell your AI agent:**

> "Create database migrations for [module-name] based on the schema."

**AI will:**
1. Generate `migrations/{module-number}_{module-name}/` directory
2. Create `migration.yaml` specifying which database connection to use
3. Create `.up.sql` files for schema creation
4. Create `.down.sql` files for rollback

**Reference:** [.github/skills/implementation-lokstra-create-migrations/](../.github/skills/implementation-lokstra-create-migrations/)

**Run migrations manually:**
```bash
# For all migrations
lokstra migration up

# For specific module migrations
lokstra migration up -dir migrations/01_auth

# Check status
lokstra migration status
```

**migration.yaml example:**
```yaml
dbpool-name: db_main
schema-table: schema_migrations
enabled: true
description: Auth module database
```

**Note:** Migrations do NOT run automatically. Developer explicitly controls when to run them using the `lokstra migration` command. If a folder has `enabled: false`, `up/down` will be skipped for that folder.

---

### Step 10: Generate Code

**After all documents and migrations are ready, tell your AI agent:**

> "Generate complete implementation for [module-name]."

**AI will generate in order:**
1. **Framework Setup** (First time only)
   - Reference: [.github/skills/implementation-lokstra-init-framework/](../.github/skills/implementation-lokstra-init-framework/)
   - Updates `main.go` with bootstrap and `lokstra_init` configuration
   
2. **Configuration** (Config YAML)
   - Reference: [.github/skills/implementation-lokstra-yaml-config/](../.github/skills/implementation-lokstra-yaml-config/)
   
3. **Handlers** (HTTP Endpoints)
   - Reference: [.github/skills/implementation-lokstra-create-handler/](../.github/skills/implementation-lokstra-create-handler/)
   
4. **Services** (Repository/Data Access)
   - Reference: [.github/skills/implementation-lokstra-create-service/](../.github/skills/implementation-lokstra-create-service/)
   
5. **HTTP Test Files** (For API testing)
   - Reference: [.github/skills/implementation-lokstra-generate-http-files/](../.github/skills/implementation-lokstra-generate-http-files/)

**Advanced (Optional):**
- Middleware: [.github/skills/advanced-lokstra-middleware/](../.github/skills/advanced-lokstra-middleware/)
- Tests: [.github/skills/advanced-lokstra-tests/](../.github/skills/advanced-lokstra-tests/)
- Validation: [.github/skills/advanced-lokstra-validate-consistency/](../.github/skills/advanced-lokstra-validate-consistency/)

---

## 📁 Project Structure (After Implementation)

```
myapp/
├── .github/
│   ├── skills/                    # AI agent skills (already here!)
│   └── copilot-instructions.md    # Copilot configuration
│
├── docs/
│   ├── BRD.md                     # Business requirements
│   ├── templates/                 # Document templates
│   └── modules/
│       ├── auth/
│       │   ├── REQUIREMENTS.md
│       │   ├── API_SPEC.md
│       │   └── SCHEMA.md
│       └── product/
│           ├── REQUIREMENTS.md
│           ├── API_SPEC.md
│           └── SCHEMA.md
│
├── modules/
│   ├── auth/
│   │   ├── domain/
│   │   │   ├── user.go
│   │   │   └── auth_dto.go
│   │   ├── handler/
│   │   │   ├── auth_handler.go
│   │   │   └── zz_generated.lokstra.go
│   │   └── repository/
│   │       └── user_repository.go
│   └── product/
│       ├── domain/
│       ├── handler/
│       └── repository/
│
├── migrations/
│   ├── auth/
│   │   └── 001_create_users_table.sql
│   └── product/
│       └── 001_create_products_table.sql
│
├── config.yaml                    # Application configuration
├── main.go                        # Bootstrap code
├── go.mod                         # Go module
└── README.md                      # Project readme
```

---

## 🎯 Development Workflow Summary

```
1. Install Lokstra CLI
   ↓
2. Create new project (lokstra new clinic-app)
   ↓
3. Setup skills & dependencies
   - lokstra update-skills
   - go mod tidy
   ↓
4. Create database & update config.yaml
   - createdb clinic_app
   - Update service-definitions (db_master, db_main)
   ↓
5. Discuss needs with AI → AI generates BRD
   → Review & Approve
   ↓
6. AI generates Module Requirements
   → Review & Approve
   ↓
7. AI generates API Specifications
   → Review & Approve
   ↓
8. AI generates Database Schema
   → Review & Approve
   ↓
9. AI generates Database Migrations
   → Developer runs: lokstra migration up
   ↓
10. AI generates Implementation Code
    - Services (Repositories)
    - Handlers (HTTP Endpoints)
    - Configuration updates
    - HTTP test files
    ↓
11. Test & Review
    → go test ./modules/...
    → go run .
    ↓
12. Deploy!
```

**Multi-Tenant:** Repeat steps 5-10 for each module. Remember to add `tenant_id` isolation in schemas.

---

## 🔧 Commands

```bash
# Run the application
go run .

# Generate code from annotations only (without running server)
go run . --generate-only

# Run all tests
go test ./...

# Run specific module tests
go test ./modules/auth/...
```

### Database Migrations

Use the `lokstra migration` command to manage database migrations:

```bash
# Run pending migrations
lokstra migration up

# Run migrations for specific database
lokstra migration up -db main-db

# Run migrations from specific folder
lokstra migration up -dir migrations/auth

# Check migration status
lokstra migration status

# Rollback last migration
lokstra migration down

# Rollback multiple migrations
lokstra migration down -steps 3
```

**Migration Structure:**

Migrations can be organized in two ways:

**Single Database (All Migrations in One Folder):**
```
migrations/
├── 001_create_users.up.sql
├── 001_create_users.down.sql
├── 002_create_products.up.sql
└── 002_create_products.down.sql
```

**Multi-Database (Migrations Per Module/Database):**
```
migrations/
├── 01_main-db/
│   ├── migration.yaml              # Required
│   ├── 001_create_users.up.sql
│   └── 001_create_users.down.sql
├── 02_tenant-db/
│   ├── migration.yaml              # Required
│   ├── 001_create_tenants.up.sql
│   └── 001_create_tenants.down.sql
```

**migration.yaml example:**
```yaml
dbpool-name: main-db              # From config.yaml service-definitions
schema-table: schema_migrations
enabled: true
description: Main application database
```

**migration.yaml fields:**
- `dbpool-name` (recommended) - Database pool from `config.yaml` `service-definitions`
- `schema-table` (optional) - Migration tracking table (default: `schema_migrations`)
- `enabled` (optional) - If `false`, `up`/`down` will be skipped for that folder
- `description` (optional) - Documentation only

**Common Flags:**
- `-dir <path>` - Migrations directory (default: `migrations`)
- `-db <name>` - Database pool name (used when `migration.yaml` is missing or doesn't specify `dbpool-name`; default: `db_main`)
- `-config <file>` - Config file path (default: `config.yaml`)
- `-steps <n>` - Number of migrations to rollback

### Environment Variables

Override config values using environment variables:

```bash
# Override database connection
DB_DSN=postgres://user:pass@prod-host:5432/mydb go run .

# Override multiple values
DB_DSN=postgres://... AUTH_SECRET=my-secret go run .

# Windows PowerShell
$env:DB_DSN="postgres://user:pass@prod-host:5432/mydb"; go run .
```

**In config.yaml, use `${ENV_VAR:default}` syntax:**

```yaml
service-definitions:
   db_main:
      type: dbpool_pg
      config:
         dsn: ${DB_DSN:postgres://localhost:5432/mydb}
         schema: ${DB_SCHEMA:main}
```

**Note:** Config values support multiple resolution sources:
- `${ENV_VAR}` - Environment variable
- `${ENV_VAR:default}` - With default value
- Static values in YAML

---

## 📚 Resources

**Project Documentation:**
- BRD: `docs/BRD.md` (business requirements)
- Module Requirements: `docs/modules/{module}/REQUIREMENTS.md`
- API Specs: `docs/modules/{module}/API_SPEC.md`
- Database Schema: `docs/modules/{module}/SCHEMA.md`

**Lokstra Framework:**
- Framework Overview: [.github/skills/design-lokstra-overview/](../.github/skills/design-lokstra-overview/)
- Online Docs: https://primadi.github.io/lokstra/
- Quick Reference: https://primadi.github.io/lokstra/QUICK-REFERENCE

**Database:**
- Migrations: `migrations/{module}/` (auto-generated .up.sql, .down.sql)
- Configuration: `config.yaml` (dbpool_pg service settings)

---

## 💡 Best Practices

### 1. Always Approve Docs Before Generating Code
- Prevents rework
- Ensures stakeholder alignment
- Maintains consistency

### 2. One Module at a Time
- Easier to review and test
- Clearer progress tracking
- Isolates issues

### 3. Use @Handler Annotations
- Minimal boilerplate for business services
- Type-safe dependency injection
- Auto-generates routing code

### 4. Test After Each Module
- Catch issues early
- Validate business logic
- Ensure API compliance

### 5. Version Your Documents
- Track requirements changes
- Maintain audit trail
- Enable rollback if needed

---

## 🎨 Module Structure Pattern

Each module follows clean architecture:

```go
modules/{module-name}/
├── domain/           # Business entities and DTOs
│   ├── entity.go     # Domain model (User, Product, Order)
│   └── dto.go        # Request/Response contracts
│
├── handler/          # HTTP handlers (application layer)
│   ├── handler.go    # With @Handler annotation
│   └── zz_generated.lokstra.go  # Auto-generated
│
└── repository/       # Data access layer
    └── repository.go # Database operations
```

---

## 🔄 Typical Development Cycle

### Setup Phase (One-time)

```bash
# 1. Install Lokstra CLI
go install github.com/primadi/lokstra/cmd/lokstra@latest

# 2. Create project
lokstra new clinic-app
cd clinic-app

# 3. Setup skills & dependencies
lokstra update-skills
go mod tidy

# 4. Create database
createdb clinic_app

# 5. Update config.yaml with db_master and db_main connections
vim config.yaml

# 6. Verify setup
go run . --generate-only
```

### First Module (Example: Auth)

1. **Generate docs:**
   ```
   AI: Create BRD for clinic system
   AI: Create auth module requirements
   AI: Create auth API spec
   AI: Create auth database schema
   ```

2. **Generate migrations & code:**
   ```
   AI: Create auth migrations
   lokstra migration up         # Run migrations explicitly
   AI: Generate auth services
   AI: Generate auth handlers
   AI: Generate auth HTTP test files
   ```

3. **Test:**
   ```bash
   go test ./modules/auth/...
   go run .                    # Start application
   # Test endpoints using .http files
   ```

### Subsequent Modules (Example: Patient)

1. **Generate docs:**
   ```
   AI: Create patient module requirements
   AI: Create patient API spec
   AI: Create patient database schema
   ```

2. **Generate migrations & code:**
   ```
   AI: Create patient migrations
   lokstra migration up         # Run migrations explicitly
   AI: Generate patient services
   AI: Generate patient handlers
   AI: Generate patient HTTP test files
   ```

3. **Test & integrate:**
   ```bash
   go test ./modules/patient/...
   go run .                    # Verify integration
   ```

### Multi-Tenant Considerations

**For multi-tenant deployment:**
- Include `tenant_id` in all table schemas
- Configure `dbpool_pg` with different connections per tenant (optional)
- Implement tenant isolation at data access layer

**For single-tenant deployment:**
- Skip `tenant-registration` module
- Use single database connection for both `db_master` and `db_main`
- No need for `tenant_id` in schemas

---

## ❓ Common Questions

### Q: Can I skip the documentation phase?

**A:** Not recommended. Documents serve as:
- Requirements validation
- Stakeholder communication
- Implementation contracts
- Testing specifications

### Q: Can I modify generated code?

**A:** Yes! But:
- Avoid modifying `zz_generated.lokstra.go` files (regenerated automatically)
- Modify handlers, domain models, repositories as needed
- Update docs if requirements change

### Q: How do I add custom middleware?

**A:** See examples in:
- [.github/skills/advanced-lokstra-middleware/](../.github/skills/advanced-lokstra-middleware/)
- Per-route: Add `middlewares=["auth", "admin"]` to `@Route` annotation

### Q: How do I handle module dependencies?

**A:** 
- Define in `REQUIREMENTS.md` (Dependencies section)
- AI will generate proper `@Inject` annotations
- Configure in `config.yaml` (service-definitions)

---

## 🆘 Need Help?

**Ask your AI agent:**

| Question | AI Will |
|----------|---------|
| "What's next in the workflow?" | Guide you to the next step |
| "Show me an example of [document type]" | Provide template or example |
| "Help me review this [document]" | Check completeness and quality |
| "Generate code for [module] step by step" | Walk through implementation |
| "Fix this error: [error message]" | Diagnose and suggest fixes |

**Your AI agent has all the skills loaded from `.github/skills/`!**

---

## 🎓 Learning Resources

### For Beginners
1. Read [.github/skills/design-lokstra-overview/](../.github/skills/design-lokstra-overview/) - Framework fundamentals
2. Follow this HOW_TO_DEVELOP.md guide step by step
3. Start with a simple module (e.g., health check)

### For Experienced Developers
1. Review all skills in [.github/skills/](../.github/skills/)
2. Check implementation skills for code patterns
3. Use advanced skills for testing, validation, and middleware
4. Customize patterns as needed for your domain

---

**Happy Coding! 🚀**
