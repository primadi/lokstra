# Interactive Mode Guide for BRD Generation

This guide helps AI agents choose and execute the right BRD generation mode based on project context.

---

## Mode Selection Matrix

### Mode 1: Detailed Interactive Mode

**Use when:**
- ✅ Complex project (10+ features, 4+ user roles)
- ✅ Multiple stakeholders (3+ approvers needed)
- ✅ Compliance/regulatory requirements (HIPAA, GDPR, etc)
- ✅ High risk or mission-critical system
- ✅ Budget-constrained (need clarity to avoid rework)
- ✅ Tight timeline (need detailed planning)
- ✅ First-time collaboration with stakeholder

**Characteristics:**
- **Duration:** 30-45 minutes Q&A
- **Questions:** 20-25 detailed questions
- **BRD Output:** 22+ sections, comprehensive
- **Revisions:** Expect 2-3 iterations before approval
- **Best for:** Enterprise, medical, financial, compliance-heavy projects

**Example Projects:**
- Clinic Management System (healthcare compliance)
- Banking app (regulatory, security)
- Insurance system (complex rules, compliance)
- E-government portal (legal requirements)

---

### Mode 2: Quick Proposal Mode

**Use when:**
- ✅ Simple MVP (5-8 features, 2-3 user roles)
- ✅ Single stakeholder or small team
- ✅ Flexible timeline (iterative approach OK)
- ✅ No hard compliance requirements
- ✅ Developer has clear vision
- ✅ Willing to revise in code (agile approach)
- ✅ Speed is priority over completeness

**Characteristics:**
- **Duration:** 10-15 minutes Q&A
- **Questions:** 5-7 core questions only
- **BRD Output:** 12-15 sections, focused
- **Revisions:** Expect 1-2 iterations
- **Best for:** Startup MVP, internal tools, learning projects

**Example Projects:**
- Todo app (simple)
- Internal tools (CRM, inventory)
- MVP prototype
- Learning/training project

---

## Core Questions (Required for ALL BRD)

Regardless of Mode 1 or Mode 2, agent **MUST** gather answers to these 7 core questions before proceeding to generate BRD:

| # | Question | Why Required | BRD Section |
|----|----------|--------------|-------------|
| Q1 | What problem does this solve? | Define business value & objectives | Executive Summary + Objectives |
| Q2 | Who are the primary users/stakeholders? | Identify scope & roles | Stakeholders + Scope |
| Q3 | How will you measure success? | Define KPIs & acceptance criteria | Success Metrics |
| Q4 | What's in scope for v1.0? | Define deliverables & timeline | Scope |
| Q5 | When do you need to launch? | Set realistic timeline | Timeline & Milestones |
| Q6 | What are budget/resource constraints? | Assess feasibility | Assumptions & Constraints |
| Q7 | What are known risks or blockers? | Plan mitigation | Risks & Mitigation |

**Agent Rule:** 
```
IF any core question unanswered:
    → STOP & ask developer
    → Don't proceed until all 7 answered
ELSE:
    → Proceed to complexity assessment
```

**Example - Checking Core Questions:**
```
Agent: "Let me confirm I understand your project:
✓ Q1: Problem = Managing patient appointments (avoid manual Excel)
✓ Q2: Users = Clinic staff (FO, Kasir, Dokter) + Patients
✓ Q3: Success = Eliminate 2 hours daily admin, reduce double-booking
✓ Q4: v1.0 = Patient registration, appointment booking, basic reporting
✓ Q5: Launch = 3 months (by April 2026)
✓ Q6: Budget = $50K, 2 developers
✓ Q7: Risk = Integration dengan SatuSehat API

Semua jelas? Lanjut ke complexity assessment?"
```

---

## Project Complexity Assessment

After core 7 questions answered, agent assesses project complexity to determine BRD scope:

### Simple MVP
**Characteristics:**
- Timeline: < 4 weeks
- Budget: < $20K
- Users: 1-2 main user roles
- Features: < 10 core features
- Stakeholders: 1-2 approvers
- Compliance: None or minimal
- Integrations: 0-1

**BRD Sections (Minimal - 6 sections):**
1. Executive Summary
2. Business Objectives
3. Functional Requirements (core features only)
4. Scope & Out of Scope
5. Timeline & Milestones
6. Approval Section

**Agent approach:** Skip NFR, Tech Stack, advanced sections. Keep lean.

---

### Standard Project
**Characteristics:**
- Timeline: 2-3 months
- Budget: $20K-$100K
- Users: 3-5 user roles
- Features: 10-30 core features + nice-to-haves
- Stakeholders: 3-5 approvers
- Compliance: Some requirements (GDPR basic, security)
- Integrations: 1-3 systems

**BRD Sections (Standard - 12 sections):**
1. Executive Summary
2. Business Objectives
3. Stakeholders & Roles
4. Scope & Out of Scope
5. Functional Requirements (with acceptance criteria)
6. Non-Functional Requirements (performance, security)
7. Success Metrics & KPIs
8. Risks & Mitigation
9. Timeline & Milestones
10. Approval Section
11. Assumptions & Constraints
12. Integration Points

**Agent approach:** Include NFR & performance metrics. Standard coverage.

---

### Enterprise Project
**Characteristics:**
- Timeline: > 3 months
- Budget: > $100K
- Users: 5+ user roles with complex permissions
- Features: 30+ core + many customization points
- Stakeholders: 5+ approvers, multiple departments
- Compliance: HIPAA, GDPR, ISO, industry-specific
- Integrations: 3+ external systems
- Risk: Mission-critical, high impact

**BRD Sections (Comprehensive - 22 sections):**
All sections from BRD_TEMPLATE.md including:
1. Executive Summary
2. Business Objectives
3. Stakeholders & Roles
4. Scope & Out of Scope
5. Functional Requirements (detailed)
6. Non-Functional Requirements (comprehensive)
7. **Role-Based Access Control (RBAC)** ← Enterprise addition
8. **Technology Stack & Architecture** ← Enterprise addition
9. Success Metrics & KPIs
10. Risks & Mitigation
11. Timeline & Milestones
12. Approval Section
13. Assumptions & Constraints
14. Integration & API Requirements
15. Data Security & Compliance
16. Testing Strategy
17. Training & Onboarding
18. Support & Maintenance
19. Change Management
20. Disaster Recovery & Backup
21. Performance & Scalability
22. Budget & Resource Allocation

**Agent approach:** Full comprehensive BRD. Include RBAC matrix, tech stack rationale, security controls.

---

**Agent Assessment Logic:**
```
Calculate Complexity Score:
- Timeline: < 1 month = 1 pt, 1-3 months = 2 pts, > 3 months = 3 pts
- Budget: < $20K = 1 pt, $20-100K = 2 pts, > $100K = 3 pts
- Features: < 10 = 1 pt, 10-30 = 2 pts, > 30 = 3 pts
- Users: < 3 = 1 pt, 3-5 = 2 pts, > 5 = 3 pts
- Compliance: None = 1 pt, Some = 2 pts, Strict = 3 pts

Total Score:
- 5-7: Simple MVP
- 8-12: Standard Project
- 13+: Enterprise Project

Show to developer:
"Based on your inputs, this looks like [COMPLEXITY LEVEL] project.
This means BRD will have [X] sections and take [TIME] to prepare.
Agree?"
```

---

## Pre-Publish BRD Checklist

Before agent publishes BRD to `docs/modules/{project}/`, verify ALL items:

**Content Completeness:**
- [ ] All 7 core questions documented & visible in BRD
- [ ] Executive Summary is clear & concise (max 1 page)
- [ ] All business objectives are SMART (Specific, Measurable, Achievable, Relevant, Time-bound)
- [ ] All stakeholders listed with roles & responsibilities
- [ ] Scope clearly separated: In-Scope vs Out-of-Scope vs Future
- [ ] All FR have acceptance criteria (testable)
- [ ] All NFR have measurable targets (< 200ms, 99.9% uptime, etc)
- [ ] Success metrics defined with target values & timeline
- [ ] Risks identified with probability & impact assessment
- [ ] Mitigation plans for each risk
- [ ] Timeline has clear milestones with dates
- [ ] Approval section complete with sign-offs

**Requirements Quality:**
- [ ] No vague requirements ("system should be fast")
- [ ] All acceptance criteria are measurable
- [ ] No impossible requirements ("100ms response time for 1M users")
- [ ] Requirements are realistic for timeline & budget
- [ ] NFR match chosen technology stack

**Business Alignment:**
- [ ] Stakeholders clearly identified & agreed
- [ ] Success metrics can actually be measured
- [ ] Budget realistic for scope & timeline
- [ ] Timeline achievable with available resources

**Compliance & Risk:**
- [ ] Compliance requirements documented (GDPR, HIPAA, etc)
- [ ] Security concerns addressed
- [ ] Data privacy implications considered
- [ ] Dependencies on external systems identified

**Final Gate Questions:**
```
Agent MUST answer YES to all before publishing:

1. Are all 7 core questions answered? (Yes/No)
   If No → Go back to Phase 2-7, ask missing questions

2. Is scope realistic for timeline & budget? (Yes/No)
   If No → Ask developer: "Need to reduce scope or extend timeline?"

3. Are all acceptance criteria measurable? (Yes/No)
   If No → Go back to FRs, make them testable

4. Have all stakeholders reviewed & approved? (Yes/No)
   If No → Ask: "Need stakeholder review before publish?"

5. Is BRD complete & ready for next phase (Module Requirements)? (Yes/No)
   If No → Identify missing parts, ask developer
```

**If ANY gate is No:**
```
Agent: "BRD not ready for publish yet. Missing: [X]

Options:
1. Fix now (should take 10-15 min)
2. Keep as draft & fix later
3. Cancel & restart

What works best?"
```

**If ALL gates are Yes:**
```
Agent: "BRD ready to publish! 

Next steps:
1. Save to: docs/modules/{project}/BRD-{project}-v1.0.md
2. Add approval sign-offs
3. Can proceed to 'lokstra-module-requirements' skill for next phase"
```

---

## Mode 1: Detailed Interactive Workflow

### Phase 1: Kickoff & Mode Selection

**Agent greeting:**
```
"Saya akan membantu generate BRD yang comprehensive.

Ada 2 mode:

MODE 1 (Detailed) - 30-45 min Q&A
✅ Best untuk: Complex project, compliance, multiple stakeholders
❌ Time: Lebih lama tapi hasil lebih detail

MODE 2 (Quick) - 10-15 min Q&A
✅ Best untuk: MVP, simple project, fast iteration
❌ Trade-off: Less comprehensive

Mana yang cocok untuk project anda?"
```

### Phase 2: Problem & Vision (5-6 minutes)

Agent asks Q1.1-Q1.4 from CLARIFYING_QUESTIONS.md

**Agent approach:**
```
Q1.1: "Apa masalah utama yang ingin diselesaikan?"
[Developer answers]

Q1.2: "Siapa target pengguna? Berapa jumlahnya?"
[Developer answers]

Q1.3: "Apa business value yang diharapkan?"
[Developer answers]

Q1.4: "Ada competitor/reference system?"
[Developer answers]
```

**Agent validates:** All Q1 answered clearly & specifically

---

### Phase 3: Scope & Features (8-10 minutes)

Agent asks Q2.1-Q2.5

**Agent approach:**
```
Q2.1: "Apa MVP features untuk v1.0?"
[Developer lists top 5 features]

Q2.2: "Apa yang TIDAK masuk v1.0?"
[Developer clarifies out-of-scope]

Q2.3: "Planned features v2.0 & beyond?"
[Developer outlines roadmap]

Q2.4: "Ada external system yang perlu diintegrasikan?"
[Developer lists integrations]

Q2.5: "Ada compliance/regulatory requirement?"
[Developer lists requirements]
```

**Agent validates:** Scope is realistic, no conflicts

---

### Phase 4: Technical Context (8-10 minutes)

Agent asks Q3.1-Q3.4

**Agent approach:**
```
Q3.1: "Apa tech stack yang direncanakan?"
[Developer specifies Frontend, Backend, Database]

Q3.2: "Estimasi user capacity?"
[Developer gives concurrent users, data volume]

Q3.3: "Performance target?"
[Developer gives response time, uptime, etc]

Q3.4: "Specific security requirement?"
[Developer lists security needs]
```

**Agent validates:** Tech stack is appropriate for scope

---

### Phase 5: Timeline & Constraints (5-7 minutes)

Agent asks Q4.1-Q4.5

**Agent approach:**
```
Q4.1: "Kapan go-live target?"
[Developer gives timeline in weeks/months]

Q4.2: "Estimasi budget total?"
[Developer gives budget range]

Q4.3: "Berapa team developer?"
[Developer gives headcount]

Q4.4: "Ada dependency ke sistem/team lain?"
[Developer lists dependencies]

Q4.5: "Ada constraint khusus?"
[Developer lists constraints]
```

**Agent validates:** Timeline & budget realistic for scope

---

### Phase 6: Success Metrics (3-5 minutes)

Agent asks Q5.1-Q5.4

**Agent approach:**
```
Q5.1: "Bagaimana mengukur project sukses?"
[Developer defines KPIs/success criteria]

Q5.2: "Acceptance criteria per feature?"
[Developer gives measurable criteria]

Q5.3: "Ada pilot phase?"
[Developer explains rollout strategy]

Q5.4: "Siapa stakeholders yang perlu approval?"
[Developer lists stakeholders & roles]
```

**Agent validates:** Success metrics are measurable

---

### Phase 7: Additional Context (2-3 minutes, Optional)

Agent asks Q6.1-Q6.4 if relevant

**Agent approach:**
```
Q6.1: "Ada existing documentation?"
Q6.2: "Architecture preference?"
Q6.3: "Data migration requirement?"
Q6.4: "Organizational constraints?"
```

---

### Phase 8: Generate BRD Document

**Agent generates:**

```
BRD Structure (22 sections):
1. Executive Summary
2. Business Objectives
3. Stakeholders
4. Scope
5. Business Processes (with Mermaid diagrams)
6. Functional Requirements
7. Non-Functional Requirements
8. Data Requirements
9. Integration Requirements
10. User Interface Requirements
11. Role-Based Access Control (RBAC)
12. Technology Stack & Architecture
13. Reporting & Analytics
14. Training & Support
15. Assumptions & Constraints
16. Risks & Mitigation
17. Timeline & Milestones
18. Budget Estimate
19. Glossary
20. References
21. Approval Section (Empty, waiting for approval)
22. Document History
```

**Save location:** `docs/drafts/{project}/BRD-{project}-v1.0-draft.md`

**Output sample:**
```markdown
# Business Requirements Document
## [Project Name]

**Version:** 1.0.0 (draft)
**Status:** 🔄 In Review
**Created:** 2026-01-30
**Last Updated:** 2026-01-30

[22 detailed sections based on Q&A answers]

## Approval Status
[Empty - waiting for developer review]
```

---

### Phase 9: Review & Revision Cycle

**Agent presents BRD to developer:**

```
"BRD v1.0 sudah jadi! 📄 

Berikut highlight:
- Features: [FR-001 to FR-010]
- NFR: [< 500ms, 99.5% uptime, AES-256]
- Timeline: [12 weeks in 4 sprints]
- Budget: [$50,000]
- Stakeholders: [Product Owner, Tech Lead, CEO]

Feedback?"
```

**Developer can:**
1. ✅ "Bagus, siap approve"
2. 🔄 "Bagian [section] perlu diubah..."
3. ❌ "Terlalu detail, bisa lebih ringkas?"

**If revision needed:**

Agent updates BRD → Increment version v1.0 → v1.1

```
docs/drafts/{project}/BRD-{project}-v1.1-draft.md

Changes:
- Section 5: Updated feature list
- Section 15: Added constraint X
- Section 17: Revised timeline
```

**Repeat until:** Developer says "Ready to approve"

---

### Phase 10: Approval & Publishing

**Agent asks:**
```
"Siap publish ke docs/modules/ untuk approval stakeholders?"
```

**If yes:**

Agent:
1. ✅ Adds approval section template
2. ✅ Copies to `docs/modules/{project}/BRD-{project}-v1.1.md`
3. ✅ Lists stakeholders who need to sign-off

**Approval section template:**
```markdown
## Approval Status

**Document Version:** 1.1  
**Status:** ⏳ Pending Approval

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Product Owner | [Name] | [Date] | [ ] |
| Tech Lead | [Name] | [Date] | [ ] |
| Stakeholder 1 | [Name] | [Date] | [ ] |

**Next Step:** Send to stakeholders for sign-off
```

---

## Mode 2: Quick Proposal Workflow

### Simplified Q&A (10-15 minutes)

Agent asks only core questions:

```
Q1.1: "Apa masalah & target user?"
Dev: [Brief answer - 1-2 sentences]

Q2.1: "Apa MVP features?"
Dev: [List 5-8 features]

Q4.1: "Timeline target?"
Dev: [Give weeks/months]

Q4.2: "Budget estimate?"
Dev: [Give range]

Q5.1: "How measure success?"
Dev: [Give 2-3 KPIs]

Q2.4: "Any integrations?"
Dev: [Yes/No, if yes describe]
```

### Quick BRD Generation

Agent generates condensed BRD (12-15 sections):

```
1. Executive Summary (concise)
2. Business Objectives
3. Scope (In/Out)
4. Functional Requirements (top features only)
5. Success Metrics
6. Timeline
7. RBAC (simple)
8. Tech Stack
9. Risks
10. Approval (empty)
```

**Save location:** `docs/drafts/{project}/BRD-{project}-v1.0-draft.md`

### Quick Revision Cycle

```
Dev: "Looks good but add [requirement]"
Agent: Update → v1.1-draft

Dev: "Perfect, publish"
Agent: Publish → docs/modules/
```

---

## When to Escalate to Mode 1

If developer chose Mode 2 but answers reveal complexity:

```
Agent: "Melihat dari jawaban Anda:
- 15+ features (kompleks)
- 5 user roles berbeda
- 3 external integration
- Compliance requirement HIPAA

Recommend upgrade ke Mode 1 untuk lebih comprehensive BRD?
(Akan add 15-20 min di Q&A)"
```

---

## Key Decision Points for Agent

**Decision 1: Mode Selection**
```
if complexity > 10 features AND stakeholders > 2:
    → Use Mode 1
else if complexity < 8 features AND timeline_flexible:
    → Use Mode 2
else:
    → Ask developer: "Which mode fits better?"
```

**Decision 2: Revision Needed?**
```
if feedback_count >= 3:
    → Consider suggesting requirement refinement
    → Offer to regroup requirements

if timeline_unrealistic:
    → Alert: "12 weeks for 50 features may be tight, need adjustment?"
    
if budget_insufficient:
    → Alert: "Team + timeline may not fit budget, prioritize features?"
```

**Decision 3: Publish or Keep Draft?**
```
if all_stakeholders_identified AND 
   all_requirements_measurable AND
   success_metrics_defined:
    → Ready to publish
else:
    → Ask developer: "Missing [X], need to revise?"
```

---

## Common Pitfalls & Solutions

| Pitfall | How Agent Should Handle |
|---------|------------------------|
| Vague requirements ("system should be fast") | Ask to clarify: "Seberapa cepat? < 200ms?" |
| Unrealistic timeline ("50 features in 4 weeks") | Alert: "Risk tinggi, recommend prioritize top 10?" |
| Missing stakeholders | Ask: "Siapa yang perlu approve? Jangan lupa..." |
| No compliance considered | Ask: "Ada compliance req? (GDPR, HIPAA, etc?)" |
| Unclear success metrics | Ask: "Gimana measure success? KPI apa?" |
| Tech stack mismatch | Warn: "Go + React might be overkill untuk MVP simple" |

---

**End of Interactive Mode Guide**
