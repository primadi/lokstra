# How to Develop with Lokstra Framework

**Welcome to AI-Driven Development with Lokstra!**

This project is set up for **design-first development** using AI agents (GitHub Copilot, Cursor, etc.).

---

## 🚀 Quick Start

### Step 1: Understand Your Business Needs

Before writing any code, clearly define:
- What problem are you solving?
- Who are the users?
- What are the main features?
- What are the success metrics?

---

### Step 2: Generate Business Requirements Document (BRD)

**Tell your AI agent:**

> "I want to build [describe your app]. Help me create a BRD using SKILL 0."

**AI will:**
1. Ask clarifying questions about your business needs
2. Generate `docs/BRD.md` with:
   - Executive summary
   - Business objectives & metrics
   - Stakeholders
   - Functional requirements
   - Non-functional requirements
   - Scope & constraints

**Template:** [templates/BRD_TEMPLATE.md](templates/BRD_TEMPLATE.md)

**Reference:** [.github/skills/01-document-workflow.md](../.github/skills/01-document-workflow.md)

**⚠️ Review and approve BRD before proceeding!**

---

### Step 3: Generate Module Requirements

**Tell your AI agent:**

> "Based on the BRD, help me identify bounded contexts and generate module requirements using SKILL 1."

**AI will:**
1. Analyze BRD to identify modules (bounded contexts)
2. Generate `docs/modules/{module-name}/REQUIREMENTS.md` for each module
3. Define module dependencies

**Example modules:**
- Auth (authentication & authorization)
- Product (product catalog)
- Order (order management)
- Payment (payment processing)

**Template:** [templates/MODULE_REQUIREMENTS_TEMPLATE.md](templates/MODULE_REQUIREMENTS_TEMPLATE.md)

**Reference:** [.github/skills/02-module-requirements.md](../.github/skills/02-module-requirements.md)

**⚠️ Review and approve module requirements!**

---

### Step 4: Generate API Specifications

**Tell your AI agent:**

> "Generate API specifications for [module-name] using SKILL 2."

**AI will:**
1. Generate `docs/modules/{module-name}/API_SPEC.md`
2. Define all endpoints with:
   - HTTP method & path
   - Request/response schemas
   - Validation rules
   - Error responses
   - Examples

**Template:** [templates/API_SPEC_TEMPLATE.md](templates/API_SPEC_TEMPLATE.md)

**Reference:** [.github/skills/03-api-spec.md](../.github/skills/03-api-spec.md)

**⚠️ Review and approve API specs!**

---

### Step 5: Generate Database Schema

**Tell your AI agent:**

> "Generate database schema for [module-name] using SKILL 3."

**AI will:**
1. Generate `docs/modules/{module-name}/SCHEMA.md`
2. Define tables, indexes, constraints
3. Create migration files in `migrations/{module-name}/`

**Template:** [templates/SCHEMA_TEMPLATE.md](templates/SCHEMA_TEMPLATE.md)

**Reference:** [.github/skills/04-schema.md](../.github/skills/04-schema.md)

**⚠️ Review and approve schema!**

---

### Step 6: Generate Code

**After all documents are approved, tell your AI agent:**

> "Generate code for [module-name] using SKILL 4-13."

**AI will:**
1. Create module folder structure in `modules/{module-name}/`
2. Generate database migrations
3. Generate domain models
4. Generate repositories
5. Generate handlers with `@Handler` annotations
6. Update `config.yaml`
7. Generate tests

**Reference:** 
- [.github/skills/05-implementation.md](../.github/skills/05-implementation.md)
- [.github/skills/06-implementation-advanced.md](../.github/skills/06-implementation-advanced.md)

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
1. Discuss needs with AI
   ↓
2. AI generates BRD (SKILL 0)
   → Review & Approve
   ↓
3. AI generates Module Requirements (SKILL 1)
   → Review & Approve
   ↓
4. AI generates API Specs (SKILL 2)
   → Review & Approve
   ↓
5. AI generates Database Schema (SKILL 3)
   → Review & Approve
   ↓
6. AI generates Code (SKILL 4-13)
   → Review & Test
   ↓
7. Run & Deploy!
```

---

## 🔧 Commands

```bash
# Run the application
go run .

# Generate code from annotations (automatic on startup)
go run . --generate-only

# Run tests
go test ./...

# Run specific module tests
go test ./modules/auth/...

# Run with custom config
go run . --config=config.yaml
```

---

## 📚 Resources

- **Skills Guide:** [.github/skills/README.md](../.github/skills/README.md)
- **Lokstra Documentation:** https://primadi.github.io/lokstra/
- **Quick Reference:** https://primadi.github.io/lokstra/QUICK-REFERENCE
- **AI Agent Guide:** https://primadi.github.io/lokstra/AI-AGENT-GUIDE

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

### First Module (Example: Auth)

1. **Generate docs:**
   ```
   AI: Create BRD
   AI: Create auth module requirements
   AI: Create auth API spec
   AI: Create auth database schema
   ```

2. **Generate code:**
   ```
   AI: Generate auth module code
   ```

3. **Test:**
   ```bash
   go test ./modules/auth/...
   go run .
   # Test endpoints manually or with tests
   ```

### Subsequent Modules (Example: Product)

1. **Generate docs:**
   ```
   AI: Create product module requirements
   AI: Create product API spec
   AI: Create product database schema
   ```

2. **Generate code:**
   ```
   AI: Generate product module code
   ```

3. **Test & integrate:**
   ```bash
   go test ./modules/product/...
   go run .
   ```

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
- [.github/skills/06-implementation-advanced.md](../.github/skills/06-implementation-advanced.md) (SKILL 10)
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
1. Read [.github/skills/00-lokstra-overview.md](../.github/skills/00-lokstra-overview.md)
2. Follow this guide step by step
3. Start with simple module (e.g., health check)

### For Experienced Developers
1. Review [.github/skills/README.md](../.github/skills/README.md)
2. Check implementation skills (SKILL 4-13)
3. Customize patterns as needed

---

**Happy Coding! 🚀**
