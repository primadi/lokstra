---
name: lokstra-module-requirements
description: Generate module-specific requirements from approved BRD for Lokstra projects. Breaks down business requirements into bounded contexts (DDD modules) with detailed functional requirements, use cases, and acceptance criteria. Use after BRD approval.
license: MIT
metadata:
  author: lokstra-framework
  version: "1.0"
  framework: lokstra
compatibility: Designed for GitHub Copilot, Cursor, Claude Code
---

# Lokstra Module Requirements Generation

## When to use this skill

Use this skill when:
- BRD has been published with approval in `docs/modules/{project}/BRD-*.md`
- Need to identify bounded contexts (modules) from business requirements
- Breaking down business requirements into technical modules
- Defining module dependencies and interfaces
- Creating detailed functional requirements per module

## Prerequisites Check (CRITICAL)

**Before proceeding, verify:**
1. Published BRD exists: `docs/modules/{project}/BRD-{project}-v*.md` (not drafts/)
2. BRD has approval section with sign-offs

**If missing:** Stop, use lokstra-brd-generation first, get approval, publish to docs/modules/

## How to generate module requirements

### Step 1: Identify Modules (Bounded Contexts)

**Interactive Mode (Default):**

Agent analyzes approved BRD and proposes modules:

```
Based on the approved BRD, I identify these modules:

1. auth - Authentication & authorization (handles FR-001 to FR-005)
2. user-profile - User information management (FR-006 to FR-010)
3. product - Product catalog (FR-011 to FR-020)
4. order - Order management (FR-021 to FR-030)
5. payment - Payment processing (FR-031 to FR-035)
6. notification - Email/SMS notifications (FR-036 to FR-040)

Do you agree with this breakdown? Would you like to:
- Approve as-is
- Merge some modules (e.g., auth + user-profile)
- Split modules further (e.g., separate inventory from product)
- Add/remove modules
```

**Proposal Mode (If developer requests):**

Developer can ask agent to propose alternative module structures:

```
Developer: "Can you propose a microservices architecture?"
Developer: "What if we combine auth and user-profile?"
Developer: "Show me a modular monolith structure"
```

Agent provides reasoning for each proposal.

**Guiding Questions:**
- "Could this be a separate microservice?"
- "Does this have its own domain logic?"
- "Can this scale independently?"

**Example:** `auth`, `user-profile`, `product`, `order`, `payment`, `notification`

**Anti-pattern:** One big `user` module for auth + profile + notifications (combine when truly related only)

### Step 2: For Each Module, Generate Requirements

**Draft Phase:**
Save to: `docs/drafts/{module}/REQUIREMENTS-{module}-v{version}-draft.md`

**Published Phase:**
Save to: `docs/modules/{module}/REQUIREMENTS-{module}-v{version}.md`

**Versioning:** Same as BRD (v1.0-draft → v1.1-draft → v1.1 published)

Use template from [references/MODULE_REQUIREMENTS_TEMPLATE.md](references/MODULE_REQUIREMENTS_TEMPLATE.md)

**Required Sections:**
1. Module Overview - Bounded context & dependencies only
2. Functional Requirements - FR-{MODULE}-001, FR-{MODULE}-002, etc.
3. Domain Model - Entities, value objects
4. Use Cases - Step-by-step flows
5. Data Validation Rules
6. Error Handling
7. Security Requirements
8. Performance Requirements
9. Integration Points - Dependencies on other modules

### Step 3: Document Integration Points

**Integration Points** describe WHAT dependencies exist (business perspective).

Actual implementation (HOW to inject via `@Inject`) is handled during code generation phase.

**Rules:**
- No circular dependencies
- Shared domain models in `/modules/shared/domain/`
- All cross-module communication documented in "Integration Points" section

**In Requirements Document:**
List which modules this one depends on and why, under "9. Integration Points" section (not YAML config)

## Module Requirements Format

```markdown
# Module Requirements: [Module Name]
## [Project Name]

**Version:** 1.0.0  
**Status:** draft  
**BRD Reference:** BRD v1.0.0  
**Last Updated:** 2026-01-28  
**Module Owner:** [Name/Team]

---

## 1. Module Overview

**Module Name:** [Module Name]  
**Purpose:** [One sentence what this module does]

**Bounded Context:**
- What domain concepts this module owns
- What it does NOT own (belongs to other modules)

**Dependencies:**
- List which modules this one depends on (e.g., "Needs auth module for user verification")
- Document the business/functional reason, not technical implementation

**Dependent Modules:**
- List which modules depend on this one (informational)

---

## 2. Functional Requirements

### FR-[MODULE]-001: [Requirement Name]
**BRD Reference:** FR-XXX  
**Priority:** High/Medium/Low  
**User Story:** As a [role], I want to [action] so that [benefit].

**Acceptance Criteria:**
- [Criterion 1]
- [Criterion 2]

**Business Rules:**
- [Rule 1]
- [Rule 2]

**API Endpoint:**
- Method: POST
- Path: `/api/[resource]`
- Authentication: Required
- Authorization: [Roles]

---

## 3. Domain Model

### Entities

#### [Entity Name]
**Attributes:**
- `id`: UUID - Unique identifier
- `[field]`: [Type] - [Description]
- `created_at`: Timestamp
- `updated_at`: Timestamp

**Relationships:**
- Belongs to: [Entity]
- Has many: [Entity]

**Business Rules:**
- [Rule 1]

---

## 4. Use Cases

### UC-[MODULE]-001: [Use Case Name]
**Actor:** [User role]  
**Goal:** [What actor wants]

**Preconditions:**
- [Condition 1]

**Main Flow:**
1. [Step 1]
2. [Step 2]
3. [Step 3]

**Alternative Flows:**
- [Alt flow]

**Postconditions:**
- [Result]

---

## 5. Validation Rules

| Field | Rules | Error Message |
|-------|-------|---------------|
| name  | Required, Min 3, Max 50 | "Name must be 3-50 characters" |
| email | Required, Email format | "Invalid email format" |

---

## 6. Error Handling

| Code | HTTP Status | Description | User Message |
|------|-------------|-------------|--------------|
| [MODULE]_001 | 400 | [Error] | "[Message]" |

---

## 7. Security Requirements

**Authentication:** JWT tokens  
**Authorization:**
- Endpoint 1: [Roles]
- Endpoint 2: [Roles]

**Data Protection:**
- [What data needs encryption]

---

## 8. Performance Requirements

- List operations: < 100ms (p95)
- Create operations: < 200ms (p95)
- Database queries: < 50ms (p99)

---

## 9. Integration Points

### Dependencies
| Module | Purpose | Data Exchanged |
|--------|---------|----------------|
| auth | Verify tokens | User ID, roles |

### Provides To
| Module | Purpose | Data Exchanged |
|--------|---------|----------------|
| order | Product details | Product info |

---

## Technical Implementation Note

**Dependency Injection:** When you proceed to code generation:
- Document "Integration Points" in requirements
- Technical implementation via `@Inject` is handled in `lokstra-code-implementation` skill
- Agent will create proper service dependencies based on integration points documented here
```

## Validation Checklist

Before proceeding to API specification:

- [ ] All modules identified with clear boundaries
- [ ] No circular dependencies
- [ ] Each FR has acceptance criteria
- [ ] Domain entities defined
- [ ] Use cases documented
- [ ] Validation rules specified
- [ ] Performance requirements measurable
- [ ] Module owner assigned

## Next Steps

After module requirements approved:
1. Use `lokstra-api-specification` skill for each module
2. Use `lokstra-schema-design` skill for database design
3. Then proceed to code generation

## Resources

- **Template:** [references/MODULE_REQUIREMENTS_TEMPLATE.md](references/MODULE_REQUIREMENTS_TEMPLATE.md)
- **DDD Guide:** See Lokstra documentation on bounded contexts
