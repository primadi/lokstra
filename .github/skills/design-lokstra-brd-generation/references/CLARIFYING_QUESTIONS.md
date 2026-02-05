# Clarifying Questions Template

Use these questions to gather comprehensive information from developer before generating BRD. Choose relevant questions based on project context.

---

## Section 1: Problem & Vision (Essential)

**Q1.1:** Apa masalah utama yang ingin diselesaikan? (Jelaskan current pain point)
- *Why:* Define problem statement & business justification
- *Good answer:* "Manual entry data pasien memakan waktu 30 menit, banyak error"
- *Bad answer:* "Kami perlu sistem yang bagus"

**Q1.2:** Siapa target pengguna? (List roles dengan estimasi jumlah)
- *Why:* Define scope & scale
- *Example:* "Dokter (5 org), Perawat (3), FO (2), Kasir (2)"

**Q1.3:** Apa business value/outcome yang diharapkan? (Jelaskan benefit)
- *Why:* Define success criteria
- *Example:* "Reduce registration time 30→15 min, 100% data accuracy"

**Q1.4:** Apakah ada competitor/reference system yang sudah ada?
- *Why:* Understand market context
- *Example:* "Similar to XYZ system, but with [differences]"

---

## Section 2: Scope & Features (Critical)

**Q2.1:** Apa MVP features untuk v1.0? (List top 3-5 harus-ada)
- *Why:* Define scope & priority
- *Example:* "Patient registration, E-Prescription, Billing"

**Q2.2:** Apa yang TIDAK masuk di v1.0? (List out-of-scope)
- *Why:* Prevent scope creep
- *Example:* "Insurance claims, Telemedicine, Mobile app"

**Q2.3:** Apa planned features untuk v2.0 & beyond?
- *Why:* Plan roadmap
- *Example:* "V2: Advanced reporting, Accounting GL, Mobile app"

**Q2.4:** Apakah ada external system yang perlu diintegrasikan?
- *Why:* Define integration complexity
- *Example:* "SatuSehat API, BPJS portal, Payment gateway (QRIS)"

**Q2.5:** Apakah ada compliance/regulatory requirement?
- *Why:* Define security & audit needs
- *Example:* "HIPAA, ISO 27001, SatuSehat standard"

---

## Section 3: Technical Context (Important)

**Q3.1:** Apa tech stack yang direncanakan? (Frontend, Backend, Database)
- *Why:* Define architecture
- *Example:* "React + Next.js, Go/Lokstra, PostgreSQL"

**Q3.2:** Berapa estimasi user capacity? (Concurrent users, data volume)
- *Why:* Define non-functional requirements
- *Example:* "100 concurrent users, 1M patient records"

**Q3.3:** Apa performance target? (Response time, uptime, etc)
- *Why:* Define SLA
- *Example:* "API < 500ms, 99.5% uptime, page load < 2sec"

**Q3.4:** Apakah ada specific security requirement?
- *Why:* Define encryption, auth, compliance
- *Example:* "End-to-end encryption, JWT auth, audit logging"

---

## Section 4: Timeline & Constraints (Operational)

**Q4.1:** Kapan go-live target? (Realistis dalam minggu/bulan)
- *Why:* Define timeline & planning
- *Example:* "12 minggu dari sekarang"

**Q4.2:** Berapa budget total? (Rough estimate OK)
- *Why:* Feasibility check
- *Example:* "$50K for MVP, $20K/bulan operational"

**Q4.3:** Berapa tim developer? (Estimasi headcount)
- *Why:* Assess resource constraint
- *Example:* "2 backend, 1 frontend, 1 QA"

**Q4.4:** Apakah ada dependency ke sistem/team lain?
- *Why:* Identify blockers
- *Example:* "Perlu approval SatuSehat dari Kemenkes (2-4 weeks)"

**Q4.5:** Apakah ada constraint khusus? (Legal, infrastructure, vendor)
- *Why:* Identify risk
- *Example:* "Must host on-premise (no cloud), legacy DB support"

---

## Section 5: Success & Metrics (Measurement)

**Q5.1:** Bagaimana mengukur project sukses? (Define KPIs)
- *Why:* Define acceptance criteria
- *Example:* "User adoption > 80%, data accuracy 100%, reduce complaints by 50%"

**Q5.2:** Apa acceptance criteria per feature? (Measurable)
- *Why:* Define done
- *Example:* "Patient registration < 5 min, E-prescription approval < 10 min"

**Q5.3:** Apakah ada pilot phase? (Before full rollout)
- *Why:* Risk mitigation
- *Example:* "Pilot 1 clinic untuk 2 weeks, then full rollout"

**Q5.4:** Siapa stakeholders yang perlu approval?
- *Why:* Define approval workflow
- *Example:* "Product Owner, Tech Lead, CEO, Department Heads"

---

## Section 6: Additional Context (Optional)

**Q6.1:** Apakah ada existing documentation/requirement?
- *Why:* Reuse existing work
- *Example:* "Spreadsheet of features, wireframes in Figma"

**Q6.2:** Apakah ada non-functional preference? (e.g., monolith vs microservices)
- *Why:* Architecture decision
- *Example:* "Prefer monolith for simplicity, modular architecture"

**Q6.3:** Apakah ada data migration requirement dari system lama?
- *Why:* Plan data layer
- *Example:* "Migrate 10K patient records dari Excel/old DB"

**Q6.4:** Apakah ada organizational constraints?
- *Why:* Plan change management
- *Example:* "Company prefers vendor X, needs vendor support"

---

## Questioning Strategy

### Interactive Mode Selection

**Choose Mode 1 (Detailed - Ask All Questions) when:**
- ✅ Large/complex project (> 20 features)
- ✅ Multiple stakeholders
- ✅ Compliance/regulatory requirements
- ✅ Tight timeline (need clarity)
- ✅ Budget-constrained (prevent rework)

**Choose Mode 2 (Quick - Ask 5-6 Core Questions) when:**
- ✅ MVP/small project (< 10 features)
- ✅ Single stakeholder/team
- ✅ Flexible timeline
- ✅ Dev has clear vision
- ✅ Willing to iterate

### Core Questions (Minimum)

If developer wants quick mode, ask at least:
1. Q1.1 - Problem statement
2. Q2.1 - MVP features
3. Q4.1 - Timeline
4. Q4.2 - Budget
5. Q5.1 - Success metrics
6. Q2.4 - External integrations (if any)

---

## Follow-up Technique

**When developer answers vague:**

❌ **Don't accept:** "Sistem harus cepat dan aman"

✅ **Ask to clarify:** 
- "Cepat itu berapa detik? (< 200ms? 1 second?)"
- "Aman dalam hal apa? (Data encryption? Access control? Audit logs?)"

**Record in BRD as Non-Functional Requirement:**
```markdown
### NFR-001: Performance
- API response time: < 200ms (p95)
- Page load time: < 2 seconds

### NFR-002: Security
- Data encryption: AES-256 at rest
- Authentication: OAuth 2.0 / JWT
- Audit logging: All changes logged
```

---

## Validation After Q&A

Before generating BRD, validate:

- [ ] All Q1 answers provided (Problem & Vision)
- [ ] Q2.1 answered (MVP features clear)
- [ ] Q4.1 & Q4.2 answered (Timeline & Budget realistic)
- [ ] Q5.1 answered (Success metrics defined)
- [ ] No major conflicting requirements
- [ ] Scope is realistic for timeline & budget

**If any missing:**
✅ Agent: "Belum terjawab: [Q], ini penting untuk BRD. Bisa dijawab?"

---

## Answer Quality Checklist

Good answers have these characteristics:

✅ **Specific** - "10 concurrent users" not "many users"  
✅ **Measurable** - "< 500ms response time" not "fast"  
✅ **Realistic** - "4 month timeline" not "2 weeks for 50 features"  
✅ **Justified** - "why this target?" not just "because"  
✅ **Complete** - All Q&A answered, no assumptions made

---

## Example: Mode 1 vs Mode 2

### Mode 1 (Full Q&A) - Clinic Project

```
Agent: "Pakai Mode 1 (detailed) atau Mode 2 (quick)?"
Dev: "Mode 1, biar comprehensive"

Agent: [Ask Q1.1-Q6.4]
Dev: [Answers all questions - 30-45 min]

Agent: Generates comprehensive BRD (22 sections) → v1.0-draft
Dev: Review, revise → v1.1-draft
Dev: "Ready to approve"
Agent: Publish → docs/modules/clinic/BRD-clinic-v1.1.md
```

### Mode 2 (Quick Q&A) - Simple MVP

```
Agent: "Pakai Mode 1 (detailed) atau Mode 2 (quick)?"
Dev: "Mode 2, urgent MVP"

Agent: [Ask Q1.1, Q2.1, Q4.1, Q4.2, Q5.1]
Dev: [Answers 5 questions - 10-15 min]

Agent: Generates focused BRD (12 key sections) → v1.0-draft
Dev: "Looks good, publish"
Agent: Publish → docs/modules/project/BRD-project-v1.0.md
```

---

**End of Clarifying Questions Template**
