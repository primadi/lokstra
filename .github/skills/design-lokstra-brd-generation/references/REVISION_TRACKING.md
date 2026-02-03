# Revision Tracking & Version Control

System untuk track changes antar BRD versions dan manage versioning dengan jelas.

---

## Versioning Scheme

### Version Format

```
BRD-{project}-v{major}.{minor}-{stage}.md

Example:
- BRD-clinic-v1.0-draft.md        (v1.0, draft stage)
- BRD-clinic-v1.1-draft.md        (v1.1, draft stage - revised)
- BRD-clinic-v1.1.md              (v1.1, published/approved)
- BRD-clinic-v2.0-draft.md        (v2.0, major revision)
```

### Versioning Rules

**Minor Version (v1.0 → v1.1):**
- Clarifications & corrections
- Additional details within same scope
- Wording/formatting improvements
- Non-breaking changes to requirements

**Major Version (v1.x → v2.0):**
- Significant scope changes
- New major features added/removed
- Architecture changes
- Compliance/regulatory changes

---

## Draft vs Published

### Draft Stage (docs/drafts/)

**Location:** `docs/drafts/{project}/BRD-{project}-v1.0-draft.md`

**Characteristics:**
- ✏️ **Editable:** Can be freely revised
- 🔄 **Iterative:** Multiple revisions expected
- 📝 **Informal:** Approval section not needed yet
- 🗑️ **Temporary:** Can be deleted when final version published

**Typical Draft Lifecycle:**
```
v1.0-draft → Review by dev
          ↓ Needs changes
v1.1-draft → Review by dev
          ↓ Needs changes
v1.2-draft → Ready to publish
          ↓ Developer approval
Publish as v1.2.md
```

**File naming for drafts:**
```
docs/drafts/clinic/BRD-clinic-v1.0-draft.md
docs/drafts/clinic/BRD-clinic-v1.1-draft.md
docs/drafts/clinic/BRD-clinic-v1.2-draft.md
```

### Published Stage (docs/modules/)

**Location:** `docs/modules/{project}/BRD-{project}-v1.2.md`

**Characteristics:**
- 🔒 **Read-only reference:** Should not be edited directly
- ✅ **Approved:** Has approval section with sign-offs
- 📋 **Official:** Serves as official requirement document
- 🔗 **Referenced:** Linked from Module Requirements, etc

**Published file naming:**
```
docs/modules/clinic/BRD-clinic-v1.0.md   (if v1.0 was approved)
docs/modules/clinic/BRD-clinic-v1.1.md   (published v1.1)
docs/modules/clinic/BRD-clinic-v1.2.md   (published v1.2)
```

**Never overwrite published files** - Each version is separate history

---

## Change Tracking Template

### Track Changes in Draft

**Add to draft before publishing:**

```markdown
## Change Log (This Draft)

### From v1.0 to v1.1

**Changed:**
- Section 4 (Scope): Added "Dashboard Antrean" to MVP
- Section 6 (FR-005): Updated requirement from "View-only" to "Edit"
- Section 15 (Timeline): Extended from 10 to 12 weeks due to SatuSehat API complexity

**Added:**
- Section 10 (RBAC): New roles "Manager" and "Auditor"
- Section 9 (Integration): SatuSehat API details

**Removed:**
- Section 14: Training section simplified (moved to Phase 2)

**Reason for Changes:**
- Stakeholder feedback requested dashboard visibility
- SatuSehat integration more complex than estimated
- Additional roles identified during stakeholder review

---
```

### What NOT to Track

❌ Don't track:
- Typo fixes (unless many)
- Formatting changes
- Grammar improvements
- Moved paragraphs (unless section reordering)

✅ Do track:
- Requirement changes
- Scope changes
- New/removed features
- Timeline changes
- Budget changes
- RBAC changes

---

## Agent Instructions for Version Management

### When Agent Generates BRD

```
1️⃣ Save to docs/drafts/{project}/BRD-{project}-v1.0-draft.md
2️⃣ Do NOT add approval section yet
3️⃣ Add CHANGELOG section (initially empty)
4️⃣ Set status to "🔄 In Review"
```

### When Developer Requests Changes

```
1️⃣ Don't overwrite v1.0-draft.md
2️⃣ Create new file: v1.1-draft.md
3️⃣ Update CHANGELOG section
4️⃣ Copy entire content from v1.0, then edit
5️⃣ Document what changed (Section, Old → New)
```

### When Developer Says Ready to Publish

```
1️⃣ Ask: "Should I publish as v1.1 or v1.0?"
   (If no changes, can publish v1.0)

2️⃣ Add approval section:
   - Status: "⏳ Pending Approval"
   - Stakeholder table (empty signatures)
   
3️⃣ Save to docs/modules/{project}/BRD-{project}-v1.1.md

4️⃣ Keep draft file for reference (don't delete)

5️⃣ Summary message:
   "✅ Published to docs/modules/clinic/BRD-clinic-v1.1.md
    Send to stakeholders for sign-off"
```

### When Major Changes Needed

```
If scope changes significantly:

1️⃣ Create new major version: v2.0-draft.md
2️⃣ Copy from v1.1 (last approved version)
3️⃣ Update all changed sections
4️⃣ Document in CHANGELOG:
   "Version 2.0: Major scope expansion
    - Added 8 new features (FR-011 to FR-018)
    - Changed architecture to microservices
    - Extended timeline to 6 months
    Reason: User feedback, market requirements"
```

---

## Revision Summary Report

### Agent Should Provide After Each Revision

**Template for agent to show developer:**

```markdown
## Revision Summary: v1.0 → v1.1

### Statistics
- Sections modified: 3
- Requirements changed: 2
- Requirements added: 3
- Requirements removed: 1
- Timeline impact: +2 weeks

### Detailed Changes

| Section | Change | Details |
|---------|--------|---------|
| Scope | Added feature | "Dashboard Antrean" to MVP |
| FR-005 | Updated | "View-only" → "Edit capability" |
| FR-010 | Added | "User activity logging" (new) |
| Timeline | Extended | 10 → 12 weeks (SatuSehat complexity) |
| RBAC | Added roles | "Manager" & "Auditor" |

### Impact Assessment
- 🟢 Low risk: Scope increase manageable
- 🟡 Medium risk: Timeline extension needs team confirmation
- 🔴 High risk: None

### Next Steps
1. Review changes above
2. Approve or request further changes
3. Ready to publish when satisfied
```

---

## Managing Multiple Stakeholder Feedback

**Scenario:** Stakeholder 1 wants feature X, Stakeholder 2 wants feature Y

**Agent should:**

```
1️⃣ Create v1.1-draft with Stakeholder 1 feedback
2️⃣ Present to Stakeholder 2
3️⃣ If Stakeholder 2 wants different changes:
   → Create v1.2-draft (incorporate both)
   → Document: "Incorporated feedback from [S1] & [S2]"
4️⃣ If conflicting requirements:
   → Flag to Product Owner
   → Document: "Conflicting requirements: [X] vs [Y], PO decision needed"
   → Don't publish until resolved
```

---

## File Organization Example

```
docs/
├── drafts/
│   └── clinic/
│       ├── BRD-clinic-v1.0-draft.md     ← Initial draft
│       ├── BRD-clinic-v1.1-draft.md     ← After revision 1
│       ├── BRD-clinic-v1.2-draft.md     ← After revision 2
│       └── BRD-clinic-v1.3-draft.md     ← Final draft before publish
│
└── modules/
    └── clinic/
        ├── BRD-clinic-v1.3.md           ← Published & approved
        ├── MODULE_REQUIREMENTS-clinic-v1.0.md
        ├── API_SPEC-clinic-v1.0.md
        └── SCHEMA-clinic-v1.0.md
```

---

## Publishing Workflow

### Pre-Publish Checklist

Before moving draft to published:

```markdown
## Pre-Publish Validation

- [ ] All requirements have clear acceptance criteria
- [ ] All non-functional requirements are measurable
- [ ] Success metrics defined & quantifiable
- [ ] Timeline is realistic for scope & team size
- [ ] Budget estimate provided
- [ ] All stakeholders identified
- [ ] No conflicting requirements
- [ ] Compliance requirements documented
- [ ] Risks & mitigation identified
- [ ] CHANGELOG completed
- [ ] No typos/formatting issues
```

If any ❌, don't publish - revise first.

### Approval Section Template

**Add when publishing:**

```markdown
---

## Approval & Sign-off

**Document Status:** ✅ Approved

**Version:** 1.3  
**Published Date:** 2026-01-30  
**Approval Status:** ⏳ Pending Stakeholder Review

### Approvers

| Role | Name | Department | Approval Date | Signature |
|------|------|-----------|--------------|-----------|
| Product Owner | John Doe | Business | | [ ] |
| Tech Lead | Jane Smith | Engineering | | [ ] |
| Compliance Officer | Alice Wong | Legal | | [ ] |
| Finance Lead | Bob Johnson | Finance | | [ ] |

### Approval Notes

- Tech Lead: "Architecture looks good, minor security notes in email"
- Product Owner: "Features align with business strategy"

### Change Control

If changes needed after publishing:
1. Create new draft (v1.4-draft)
2. Document reason for change
3. Get approval for v1.4
4. Publish as v1.4 (keep v1.3 as history)

---
```

---

## Agent Commands Summary

```
# Generate initial BRD
→ Save to docs/drafts/{project}/v1.0-draft.md

# Update after feedback
→ Save to docs/drafts/{project}/v1.1-draft.md
→ Document changes in CHANGELOG
→ Show revision summary to developer

# Ready to publish
→ Add approval section
→ Copy to docs/modules/{project}/v1.1.md
→ Keep draft for reference

# Major revision needed
→ Create docs/drafts/{project}/v2.0-draft.md
→ Document why v2.0 instead of v1.x
```

---

**End of Revision Tracking Guide**
