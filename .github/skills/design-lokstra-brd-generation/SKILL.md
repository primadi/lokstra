---
name: lokstra-brd-generation
description: Generate Business Requirements Document (BRD) for new Lokstra projects. Use when starting a new project or adding major features to establish clear, stakeholder-approved requirements before implementation. Asks clarifying questions and produces versioned BRD.
phase: design
order: 2
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.18"
---

# Lokstra BRD Generation

## When to use this skill

Use this skill when:
- Starting a new Lokstra project
- Adding major features to existing project
- Need stakeholder-approved requirements before coding
- Transitioning from business idea to technical implementation
- Documenting business objectives and success metrics

## Quick Mode Selection

Choose the right mode based on project complexity:

| Criteria | Mode 1: Detailed | Mode 2: Quick |
|----------|------------------|---------------|
| **Features** | 10+ features | 5-8 features |
| **User Roles** | 4+ roles | 2-3 roles |
| **Stakeholders** | 3+ approvers | Single/small team |
| **Compliance** | Required (HIPAA, GDPR) | Not required |
| **Duration** | 30-45 min Q&A | 10-15 min Q&A |
| **BRD Output** | 22+ sections | 12-15 sections |
| **Best For** | Enterprise, healthcare, finance | MVP, internal tools, learning |

**Decision Rule:**
```
IF (features >= 10 OR compliance_required OR stakeholders >= 3):
    → Use Mode 1: Detailed Interactive
ELSE:
    → Use Mode 2: Quick Proposal
```

See [references/INTERACTIVE_MODE_GUIDE.md](references/INTERACTIVE_MODE_GUIDE.md) for complete guide.

## Document Versioning & Workflow

**Draft Phase:**
- Save to: `docs/drafts/{project}/BRD-{project}-v{version}-draft.md`
- Can be revised freely (v1.0-draft, v1.1-draft, v1.2-draft)
- Used for internal development discussion

**Published Phase:**
- Save to: `docs/modules/{project}/BRD-{project}-v{version}.md`
- Contains approval section with sign-offs
- Used for stakeholder approval and as reference for next phase
- Each version saved with unique filename (never overwrite)
- Can proceed to Module Requirements even with revision (e.g., v1.1)

**Example Flow:**
```
1. Generate: docs/drafts/ecommerce/BRD-ecommerce-v1.0-draft.md
2. Review & revise: docs/drafts/ecommerce/BRD-ecommerce-v1.1-draft.md
3. Publish: docs/modules/ecommerce/BRD-ecommerce-v1.1.md (with approval)
4. Major changes: docs/drafts/ecommerce/BRD-ecommerce-v2.0-draft.md
```

## How to generate a BRD

### Step 1: Gather Information (Interactive Mode)

Agent asks developer about:

- **Problem & Users:** What problem? Who uses it? Business value?
- **Scope:** v1.0 features? Out of scope? Integrations?
- **Constraints:** Timeline? Budget? Compliance (GDPR)? Tech stack?
- **Success:** How measure success? KPIs? Definition of done?

**Proposal Mode:** If developer requests, agent can propose based on brief description and let developer adjust.

### Step 2: Generate BRD Document

Use the template from [references/BRD_TEMPLATE.md](references/BRD_TEMPLATE.md)

**Required Sections:**
1. Executive Summary - One-page overview
2. Business Objectives - Why we're building this
3. Stakeholders - Who's involved
4. Scope - In-scope vs out-of-scope
5. Functional Requirements - What system must do
6. Non-Functional Requirements - Performance, security
7. Success Metrics - KPIs, acceptance criteria
8. Risks & Mitigation
9. Timeline & Milestones
10. Approval Section (Required for published version)

**Approval Section Format:**
```markdown
## Approval Status

**Document Version:** 1.1  
**Status:** ✅ Approved / ⏳ Pending Review / 🔄 Revision Required

| Role | Name | Date | Signature/Notes |
|------|------|------|----------------|
| Project Owner | John Doe | 2026-01-28 | Approved |
| Tech Lead | Jane Smith | 2026-01-28 | Approved with minor notes |
| Stakeholder | Alice Wong | 2026-01-30 | Pending |

**Approval Notes:**
- Minor UI adjustments needed (noted in FR-015)
- Can proceed to Module Requirements phase
- Re-approval required if scope changes
```

**Version Control Format:**
```yaml
version: 1.1.0
status: draft | published | approved
created: 2026-01-28
last-updated: 2026-01-28
approved-by: [John Doe, Jane Smith]
approved-date: 2026-01-28
```

### Step 3: Save and Version

**Save Draft:**
```
docs/drafts/{project}/BRD-{project}-v{version}-draft.md
```

**Publish After Approval:**
```
docs/modules/{project}/BRD-{project}-v{version}.md
```

**Versioning Rules:**
- Minor changes (clarifications, typos): v1.0 → v1.1
- Major changes (scope, features): v1.x → v2.0
- Each version is separate file (never overwrite)
- Draft versions can be freely revised

**Publishing Workflow:**
1. Generate BRD in drafts/ folder
2. Developer reviews → revisions if needed
3. Add approval section to draft
4. Ask developer: "Ready to publish to docs/modules/?"
5. Copy to docs/modules/ with approval section
6. Keep draft for reference

**Agent Instructions:**
- ALWAYS start in drafts/ folder
- NEVER overwrite published documents
- ASK before publishing to docs/modules/
- Ensure approval section complete before publishing

### Step 4: Validation

Before proceeding to module requirements:

- [ ] All stakeholders identified
- [ ] Clear, measurable success metrics defined
- [ ] Functional requirements have acceptance criteria
- [ ] Non-functional requirements are measurable (< 200ms, 99.9% uptime)
- [ ] Risks documented with mitigation plans
- [ ] Timeline is realistic
- [ ] Status = `approved` with sign-off

## Requirements Format

### Functional Requirements (FR)

```markdown
### FR-001: [Requirement Name]
**Priority:** High/Medium/Low  
**User Story:** As a [role], I want to [action] so that [benefit].

**Acceptance Criteria:**
- [Specific, measurable criterion 1]
- [Specific, measurable criterion 2]
mode
**Business Rules:**
- [Rule 1]
- [Rule 2]
```

### Non-Functional Requirements (NFR)

```markdown
### NFR-001: Performance
- API response time: < 200ms (p95)
- Support 100,000 concurrent users
- Database query time: < 50ms (p99)
```

## Example BRD Structure

```markdown
# Business Requirements Document (BRD)
## [Project Name]

**Version:** 1.0.0  
**Status:** approved  
**Last Updated:** 2026-01-28  
**Approved By:** [Names]

---

## 1. Executive Summary
[Brief overview of project, business impact, key goals]

## 2. Business Objectives
- Primary Objective 1: [Measurable goal]
- Primary Objective 2: [Measurable goal]

## 3. Stakeholders
| Role | Name | Responsibilities | Contact |
|------|------|------------------|---------|
| ... | ... | ... | ... |

## 4. Scope
### In Scope (v1.0)
- Feature 1
- Feature 2

### Out of Scope
- Feature X (v2.0)
- Feature Y (future)

## 5. Functional Requirements
[FR-001, FR-002, etc.]

## 6. Non-Functional Requirements
[NFR-001, NFR-002, etc.]

## 7. Success Metrics
| Metric | Current | Target | Timeline |
|--------|---------|--------|----------|
| ... | ... | ... | ... |

## 8. Risks & Mitigation
| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| ... | ... | ... | ... |

## 9. Timeline
- Milestone 1: [Date]
- Milestone 2: [Date]

## 10. Approval
[Sign-off section]
```

## Common Mistakes to Avoid

❌ **Don't:**
- Skip stakeholder identification
- Write vague requirements ("system should be fast")
- Ignore non-functional requirements
- Start coding before approval
- Use technical jargon in business objectives

✅ **Do:**
- Use measurable criteria (< 200ms, 99.9% uptime)
- Include acceptance criteria for each requirement
- Version and track changes
- Get explicit approval before implementation
- Write for business stakeholders, not just developers

## Next Steps

After BRD is published with approval section:
1. Use `lokstra-module-requirements` skill to break down into modules
2. Agent will verify BRD approval before proceeding
3. Each module will have its own detailed requirements document
4. Proceed to API specification and schema design

**Note:** Can proceed with approved revision (e.g., v1.1), doesn't have to be final version

## Reference Files

**For Getting Started:**
- [BRD_TEMPLATE.md](references/BRD_TEMPLATE.md) - 22-section comprehensive template with RBAC & Tech Stack
- [KLINIK_BRD_EXAMPLE.md](references/KLINIK_BRD_EXAMPLE.md) - Real-world clinic system example with 3 Mermaid diagrams

**For Agents (Workflow & Questions):**
- [CLARIFYING_QUESTIONS.md](references/CLARIFYING_QUESTIONS.md) - 20+ structured questions organized by category (Problem, Scope, Technical, Timeline, Metrics, Additional)
- [INTERACTIVE_MODE_GUIDE.md](references/INTERACTIVE_MODE_GUIDE.md) - Complete guide for Mode 1 (detailed, 30-45 min) vs Mode 2 (quick, 10-15 min) with Phase-by-phase breakdown

**For Diagrams & Visualization:**
- [MERMAID_DIAGRAMS.md](references/MERMAID_DIAGRAMS.md) - 9 reusable diagram templates (Context, User Flow, ERD, Sequence, Gantt, State, Decision Tree, Architecture, RBAC Matrix)

**For Versioning & Quality:**
- [REVISION_TRACKING.md](references/REVISION_TRACKING.md) - Version control system (draft vs published, versioning scheme, change tracking)
- [VALIDATION_CHECKLIST.md](references/VALIDATION_CHECKLIST.md) - Pre-publication quality gates (95+ validation items, auto-validation with red/yellow/green flags)
- [CHANGELOG_TEMPLATE.md](references/CHANGELOG_TEMPLATE.md) - Track BRD version history and changes
