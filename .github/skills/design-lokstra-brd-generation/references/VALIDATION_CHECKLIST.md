# Pre-Publication Validation Checklist

Comprehensive checklist untuk memastikan BRD quality sebelum publishing dan approval.

---

## Section 1: Content Completeness

### Executive Summary ✅
- [ ] Clear problem statement (in 2-3 sentences)
- [ ] Business objectives listed (2-5 items)
- [ ] Success metrics defined
- [ ] Stakeholders identified by role
- [ ] Timeline mentioned (launch target date)
- [ ] Budget range or justification mentioned

**Validation Rule:** All 6 items checked = Complete

---

### Scope ✅
- [ ] "In Scope" section lists concrete features (not vague)
- [ ] "Out of Scope" section lists what's NOT included
- [ ] "Future Scope" lists v2.0+ plans with target timeline
- [ ] No overlap between In/Out/Future scopes
- [ ] All features are testable/measurable

**Validation Rule:** All 5 items checked = Clear Scope

---

### Functional Requirements ✅
- [ ] Each requirement has ID (FR-001, FR-002, etc)
- [ ] Each requirement has user story ("As a X, I want Y, so Z")
- [ ] Each requirement has acceptance criteria (3+ concrete criteria)
- [ ] Acceptance criteria are testable (not vague)
- [ ] No duplicate requirements (search for similar words)
- [ ] Requirements map to scope features
- [ ] Total requirements realistic for scope (10-15 is typical for MVP)

**Validation Rule:** All 7 items checked = Good Requirements

**Common Issues to Flag:**
```
❌ "System should be user-friendly" → Too vague
✅ "User can register in < 2 minutes" → Measurable

❌ "Support multiple languages" → How many? Which ones?
✅ "Support English, Indonesian, Mandarin" → Specific

❌ Acceptance criteria missing → Ask developer to add
✅ 3+ acceptance criteria per requirement → Good
```

---

### Non-Functional Requirements ✅
- [ ] Performance targets specified (response time in ms, not "fast")
- [ ] Uptime/availability target (e.g., 99.5%)
- [ ] User capacity mentioned (concurrent users, or DAU/MAU)
- [ ] Security requirements specific (encryption type, auth method)
- [ ] All NFRs are measurable with units (ms, %, GB, etc)
- [ ] Scalability targets mentioned
- [ ] Compliance requirements documented

**Validation Rule:** At least 5 items checked

**Measurable vs Vague:**
```
❌ "Fast response time" → How fast?
✅ "< 500ms for 95th percentile" → Specific target

❌ "Secure" → What kind of security?
✅ "AES-256 encryption at rest, TLS 1.3 in transit" → Specific

❌ "Scalable" → To what scale?
✅ "Support 100,000 concurrent users" → Specific
```

---

## Section 2: Business Alignment

### Success Metrics ✅
- [ ] Each metric is measurable (has target number & unit)
- [ ] Each metric has baseline (current state)
- [ ] Each metric has timeline (by when?)
- [ ] At least 3-5 metrics defined
- [ ] Metrics map to business objectives
- [ ] Metrics have owners (who tracks this?)

**Validation Rule:** All 6 items checked

**Example of Good Metrics:**
```markdown
| Metric | Current | Target | Timeline | Owner |
|--------|---------|--------|----------|-------|
| Registration time | 30 min | 15 min | Go-live | FO Manager |
| Data accuracy | 85% | 100% | 2 weeks | Admin |
| API uptime | N/A | 99.5% | Ongoing | DevOps |
| User adoption | 0% | 80% | 1 month | Product Owner |
```

---

### Stakeholders & Approval ✅
- [ ] All stakeholders identified by role (Product Owner, Tech Lead, etc)
- [ ] Stakeholder list includes 3+ people (minimum)
- [ ] Contact info provided for each stakeholder
- [ ] Approval section ready (though not signed yet)
- [ ] Clear path to approval defined

**Validation Rule:** At least 5 items checked

**Red Flag:** Only 1-2 stakeholders → Ask: "Who else needs to approve?"

---

## Section 3: Technical Feasibility

### Technology Stack ✅
- [ ] Frontend framework specified (React, Vue, Angular, etc)
- [ ] Backend framework specified (Lokstra, Django, Node, etc)
- [ ] Database specified (PostgreSQL, MongoDB, etc)
- [ ] Deployment strategy mentioned (Cloud, On-premise, Docker, etc)
- [ ] Tech stack is appropriate for project scope
- [ ] No obvious mismatches (e.g., blockchain for simple CRUD app)

**Validation Rule:** All 5 items checked

**Tech Stack Validation:**
```
Project: Clinic Management System
Frontend: React + Next.js ✅ (Web app, good choice)
Backend: Go + Lokstra ✅ (Good for APIs)
Database: PostgreSQL ✅ (Relational, HIPAA-ready)
Status: ✅ APPROPRIATE
```

---

### Integrations ✅
- [ ] All external integrations listed (APIs, third-party services)
- [ ] For each integration: protocol specified (REST, SOAP, GraphQL, etc)
- [ ] For each integration: authentication method specified
- [ ] Integration complexity assessed (simple vs complex)
- [ ] Contingency plans for integration failures

**Validation Rule:** All 5 items checked

**Integration Example:**
```markdown
### Integration: SatuSehat API
- Protocol: REST (FHIR R4)
- Authentication: OAuth 2.0
- Sync Frequency: Real-time
- Complexity: Medium
- Contingency: Offline mode, queue for sync when online
Status: ✅ WELL-DEFINED
```

---

### Risks & Mitigation ✅
- [ ] At least 3 risks identified
- [ ] Each risk has probability (High/Medium/Low)
- [ ] Each risk has impact assessment
- [ ] Each risk has mitigation strategy
- [ ] Owner assigned for risk monitoring
- [ ] Contingency plans for high-risk items

**Validation Rule:** All 6 items checked

**Risk Example:**
```markdown
| Risk | Probability | Impact | Mitigation | Owner |
|------|-------------|--------|-----------|-------|
| SatuSehat API outage | Medium | High | Build offline mode | Backend Lead |
| Tight timeline | High | High | Prioritize MVP features | PM |
| User adoption | Medium | Medium | Extensive training | HR |
```

---

## Section 4: Timeline & Resources

### Timeline Feasibility ✅
- [ ] Total duration realistic (not "50 features in 2 weeks")
- [ ] Broken down into phases/sprints
- [ ] Milestones defined with target dates
- [ ] Includes buffer time for UAT & deployment
- [ ] Timeline aligns with resource capacity
- [ ] Critical path identified

**Validation Rule:** All 6 items checked

**Timeline Assessment:**
```
Project: Clinic System
Features: 15 core features
Team: 2 backend, 1 frontend, 1 QA
Timeline: 12 weeks (12 features/week = unrealistic)

🔴 RED FLAG: Reduce scope or add team members
Recommendation: Focus on 8 MVP features in 8 weeks, then v2
```

---

### Budget & Resources ✅
- [ ] Budget estimate provided (even if rough)
- [ ] Team size specified (headcount)
- [ ] Role breakdown provided (Backend, Frontend, QA, etc)
- [ ] Resource allocation realistic
- [ ] Budget covers all phases (dev + test + deploy + training)
- [ ] Contingency budget (10-20%) included

**Validation Rule:** All 6 items checked

**Budget Example:**
```markdown
| Item | Cost | Notes |
|------|------|-------|
| Development (8 sprints × 2 devs) | $40,000 | Backend + Frontend |
| QA & Testing | $8,000 | UAT + Load testing |
| Infrastructure | $5,000 | AWS + SatuSehat integration |
| Training | $3,000 | User training, documentation |
| **Total** | **$56,000** | Includes 10% contingency |
```

---

## Section 5: Compliance & Security

### Compliance Requirements ✅
- [ ] Regulatory requirements identified (HIPAA, GDPR, ISO 27001, etc)
- [ ] Compliance requirements integrated into NFRs
- [ ] Audit logging requirements specified
- [ ] Data retention policies documented
- [ ] Privacy impact assessment considered
- [ ] Compliance verification plan included

**Validation Rule:** Applicable for project type (medical/finance = all checked)

**Compliance Example (Healthcare):**
```markdown
### Compliance Requirements
- **Standard:** HIPAA (US), SatuSehat (Indonesia)
- **Audit Logging:** All access to patient data logged
- **Encryption:** AES-256 at rest, TLS 1.3 in transit
- **Data Retention:** Patient records retained 7 years minimum
- **Access Control:** Role-based access with consent management
```

---

### Security Requirements ✅
- [ ] Authentication method specified (JWT, OAuth, etc)
- [ ] Authorization model specified (RBAC, ABAC)
- [ ] Data encryption specified (at rest & in transit)
- [ ] API security measures (rate limiting, WAF, etc)
- [ ] Vulnerability scanning plan
- [ ] Incident response plan outlined

**Validation Rule:** All 6 items checked

---

## Section 6: Quality & Clarity

### Writing Quality ✅
- [ ] No jargon-heavy language (business stakeholder can understand)
- [ ] No spelling/grammar errors
- [ ] Consistent formatting (headings, tables, lists)
- [ ] Diagrams are clear (if included)
- [ ] All abbreviations defined (first use: "SLM (Glossary of Transactions)")
- [ ] Document is scannable (headings, bold, tables)

**Validation Rule:** All 6 items checked

---

### Completeness Check ✅
- [ ] All required sections present (at minimum: 1, 2, 4, 5, 6, 7, 8, 9)
- [ ] No "TBD" or placeholder text remaining
- [ ] All references complete (links work, if any)
- [ ] Glossary section complete (if technical terms used)
- [ ] Version control info filled in (v1.0, date, author)

**Validation Rule:** All 5 items checked

---

## Section 7: Final Gate Questions

**Before publishing, agent asks developer:**

### Gate 1: Scope Confirmation
```
Q: "Apakah features di Section 4 sudah final?"
Expected: "Yes" or specific changes needed
If No: → Revise scope first, don't publish
```

### Gate 2: Timeline Confirmation
```
Q: "Apakah timeline 12 weeks realistis untuk team 3 orang?"
Expected: "Yes" or "Need to adjust"
If No: → Discuss trade-offs (scope or timeline)
```

### Gate 3: Budget Confirmation
```
Q: "Apakah budget $50K sudah approved?"
Expected: "Yes" or "Waiting for approval"
If No: → Don't proceed until approved
```

### Gate 4: Stakeholder Readiness
```
Q: "Sudah siap untuk stakeholder sign-off?"
Expected: "Yes" → Proceed to publish
Expected: "No, perlu [X]" → Revise
```

### Gate 5: Quality Check
```
Q: "Sudah di-review dan tidak ada typo/error?"
Expected: "Yes" → Good to publish
Expected: "No" → Give time to proofread
```

---

## Auto-Validation by Agent

Agent should automatically check and flag:

### 🔴 RED FLAGS (Stop publishing)
```
- Requirements are vague (contain "should", "nice", "flexible")
- No success metrics defined
- Timeline unrealistic (< 2 weeks for 20+ features)
- Stakeholders < 2 people
- Budget not mentioned
- No approval section
→ Agent action: "Found issues, can't publish yet"
```

### 🟡 YELLOW FLAGS (Warn but allow)
```
- Scope larger than typical MVP
- High complexity integrations
- Tight timeline with risks
- Limited team size
→ Agent action: "Warning: [Issue] - proceed anyway?"
```

### 🟢 GREEN (Safe to publish)
```
- Scope clear & realistic
- Stakeholders identified
- Budget specified
- Timeline feasible
- Success metrics defined
- All sections complete
→ Agent action: "✅ Ready to publish"
```

---

## Checklist Automation

**Agent should display:**

```markdown
# Pre-Publication Validation Report

## Completeness: ✅ 95% (47/49 items)

### Sections Status:
- ✅ Executive Summary (6/6)
- ✅ Scope (5/5)
- ✅ Functional Requirements (7/7)
- ✅ Non-Functional Requirements (6/7) ← NFR-003 needs units
- ✅ Success Metrics (6/6)
- ✅ Stakeholders (5/5)
- ✅ Technology Stack (5/5)
- ✅ Integrations (5/5)
- ✅ Timeline (6/6)
- ✅ Budget (6/6)

### Issues Found:
1. 🟡 NFR-003 (Scalability): "Support many concurrent users" → Needs specific number
2. 🟡 Risk section: Only 2 risks identified (recommend 3+)

### Recommendation:
⚠️ Can publish, but recommend fixing 2 items first
→ Ready in 5 minutes if developer clarifies above
```

---

## Publishing Decision Tree

```
All items checked?
├─ NO  → Show missing items, ask to complete
└─ YES → Check red flags?
         ├─ RED FLAGS found? 
         │  ├─ YES → Block publishing, show issues
         │  └─ NO  → Check yellow flags?
         │           ├─ YES → Warn developer, ask confirmation
         │           └─ NO  → ✅ READY TO PUBLISH
         │                   → Ask: "Publish as v1.0 or higher?"
         │                   → Add approval section
         │                   → Save to docs/modules/
         │                   → Notify stakeholders
```

---

**End of Validation Checklist**
