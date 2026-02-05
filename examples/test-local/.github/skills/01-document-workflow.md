# SKILL 0: Create Business Requirements Document (BRD)

**When to use:** Starting a new project or adding major features.

**Purpose:** Establish clear, versioned, stakeholder-approved business requirements before any code is written.

---

## Workflow

```
User Request → Ask Clarifying Questions → Generate BRD → Save to docs/BRD.md
```

### Step 1: Clarifying Questions

Before generating BRD, ask:

1. **Business Context:**
   - What problem does this solve?
   - Who are the primary users?
   - What's the business value/ROI?

2. **Scope:**
   - What features are in scope (v1.0)?
   - What's explicitly out of scope?
   - Are there integrations with existing systems?

3. **Constraints:**
   - Timeline/deadlines?
   - Budget limitations?
   - Compliance requirements (GDPR, HIPAA, etc.)?
   - Technical constraints (specific tech stack, infrastructure)?

4. **Success Criteria:**
   - How do we measure success?
   - What are the KPIs?
   - What does "done" look like?

### Step 2: Generate BRD

Use template: [docs/templates/BRD_TEMPLATE.md](../../docs/templates/BRD_TEMPLATE.md)

**Key Sections:**

1. **Executive Summary** - One-page overview
2. **Business Objectives** - Why we're building this
3. **Stakeholders** - Who's involved (sponsors, users, developers)
4. **Scope** - In-scope vs out-of-scope
5. **Functional Requirements** - What the system must do
6. **Non-Functional Requirements** - Performance, security, scalability
7. **Integrations** - External systems/APIs
8. **Constraints & Assumptions**
9. **Success Metrics** - KPIs, acceptance criteria
10. **Risks** - Technical, business, timeline risks

**Version Control:**
```yaml
version: 1.0.0
status: draft | review | approved | implemented
last-updated: 2026-01-27
approved-by: [John Doe, Jane Smith]
```

**Example BRD Output:**

```markdown
# Business Requirements Document (BRD)
## E-Commerce Order Management System

**Version:** 1.0.0  
**Status:** draft  
**Last Updated:** 2026-01-27  
**Document Owner:** Product Team  

---

## 1. Executive Summary

Build a scalable order management system to handle 10,000+ daily orders...

## 2. Business Objectives

- Increase order processing efficiency by 50%
- Reduce order fulfillment errors to < 1%
- Support 100,000 concurrent users
- Enable real-time inventory sync

## 3. Stakeholders

| Role              | Name         | Responsibilities           |
|-------------------|--------------|----------------------------|
| Executive Sponsor | John Doe     | Budget approval, strategy  |
| Product Owner     | Jane Smith   | Requirements, priorities   |
| Tech Lead         | Bob Johnson  | Architecture, implementation|
| End Users         | Customers    | Place orders, track status |

## 4. Scope

### In Scope (v1.0)
- Order creation, listing, tracking
- Inventory management
- Payment processing (Stripe integration)
- Email notifications

### Out of Scope
- Mobile app (v2.0)
- Multi-currency support (v2.0)
- Advanced analytics dashboard (future)

## 5. Functional Requirements

### FR-001: Order Creation
**Priority:** High  
**Description:** Users must be able to create orders with multiple items.

**Acceptance Criteria:**
- User can add 1-100 items per order
- System validates inventory availability
- Order total calculated correctly (items + tax + shipping)
- Order confirmation sent via email

### FR-002: Order Tracking
**Priority:** High  
**Description:** Users can track order status in real-time.

**Acceptance Criteria:**
- Status updates: pending → processing → shipped → delivered
- Estimated delivery date displayed
- Push notifications on status change

[... continue for all functional requirements ...]

## 6. Non-Functional Requirements

### NFR-001: Performance
- API response time: < 200ms (p95)
- Support 100,000 concurrent users
- Database query time: < 50ms (p99)

### NFR-002: Security
- TLS 1.3 for all API endpoints
- PCI-DSS compliant payment processing
- Role-based access control (RBAC)

### NFR-003: Scalability
- Horizontal scaling to handle 10x traffic spikes
- Database sharding for > 10M orders

[... continue for all NFRs ...]

## 7. Integrations

| System      | Purpose              | Protocol | SLA        |
|-------------|----------------------|----------|------------|
| Stripe      | Payment processing   | REST API | 99.9%      |
| SendGrid    | Email notifications  | REST API | 99.95%     |
| Inventory   | Stock sync           | gRPC     | 99.99%     |

## 8. Constraints & Assumptions

**Constraints:**
- Go 1.22+ required (company standard)
- PostgreSQL 15+ (existing infrastructure)
- Deploy on AWS EKS

**Assumptions:**
- Inventory service API is stable
- Payment gateway uptime > 99.9%
- Users have modern browsers (Chrome, Firefox, Safari)

## 9. Success Metrics

| Metric                     | Target      | Measure Method       |
|----------------------------|-------------|----------------------|
| Order processing time      | < 2 seconds | Application logs     |
| API uptime                 | 99.9%       | Health check monitor |
| User satisfaction (NPS)    | > 8/10      | Post-order survey    |
| Order error rate           | < 1%        | Error tracking       |

## 10. Risks

| Risk                       | Probability | Impact | Mitigation                    |
|----------------------------|-------------|--------|-------------------------------|
| Inventory API downtime     | Medium      | High   | Implement caching + fallback  |
| Payment gateway rate limit | Low         | High   | Queue + retry mechanism       |
| Database scaling issues    | Medium      | High   | Implement read replicas early |

## 11. Timeline

- **Week 1-2:** BRD approval, module requirements
- **Week 3:** API spec, schema design
- **Week 4-8:** Implementation
- **Week 9:** Testing, bug fixes
- **Week 10:** Production deployment

## 12. Approval

| Name         | Role          | Signature | Date       |
|--------------|---------------|-----------|------------|
| John Doe     | Exec Sponsor  |           |            |
| Jane Smith   | Product Owner |           |            |
| Bob Johnson  | Tech Lead     |           |            |

---

## Document History

| Version | Date       | Author      | Changes              |
|---------|------------|-------------|----------------------|
| 1.0.0   | 2026-01-27 | Jane Smith  | Initial draft        |
```

### Step 3: Save & Version

```bash
# Save to project root
docs/BRD.md

# Version control
git add docs/BRD.md
git commit -m "docs: add BRD v1.0.0 (draft)"
```

**Immutability Rule:**
- Once status = `approved`, BRD is **immutable**
- Changes require new version (e.g., 1.1.0) and re-approval

---

## Validation Checklist

Before moving to next step (Module Requirements):

- [ ] All stakeholders identified
- [ ] Clear success metrics defined
- [ ] Functional requirements have acceptance criteria
- [ ] Non-functional requirements are measurable
- [ ] Risks documented with mitigation plans
- [ ] Timeline is realistic
- [ ] Status = `approved` (sign-off obtained)

---

## Next Step

Once BRD is approved, proceed to:
- **SKILL 1:** [02-module-requirements.md](02-module-requirements.md) - Generate module-specific requirements

---

## Common Mistakes to Avoid

❌ **Don't:**
- Skip stakeholder identification
- Write vague requirements ("system should be fast")
- Ignore non-functional requirements
- Start coding before approval

✅ **Do:**
- Use measurable criteria (< 200ms, 99.9% uptime)
- Include acceptance criteria for each requirement
- Version and track changes
- Get explicit approval before implementation

---

**Template:** [docs/templates/BRD_TEMPLATE.md](../../docs/templates/BRD_TEMPLATE.md)  
**Example:** See above for complete BRD example.
