# Document Change Log

Track all document versions and changes in this file.

## Format

```markdown
## [VERSION] - DATE - STATUS

**Document:** [Document Name]
**Status:** draft | review | approved | implemented | deprecated
**Approved By:** [Name] (if approved)
**Approval Date:** [Date] (if approved)

### Added
- [New feature/section 1]
- [New feature/section 2]

### Changed
- [Changed item 1]: [Description]
- [Changed item 2]: [Description]

### Deprecated
- [Deprecated item]: [Reason]

### Removed
- [Removed item]: [Reason]

### Fixed
- [Fixed issue]: [Description]

### Reasons
- [Explanation of why changes were made]

---
```

## Example Entry

```markdown
## [1.2.0] - 2026-02-15 - APPROVED

**Document:** BRD - Clinic Management System
**Status:** approved
**Approved By:** Dr. Budi (Product Owner)
**Approval Date:** 2026-02-15

### Added
- FR-PAY-001: Payment gateway integration (Midtrans)
- FR-PAY-002: Payment confirmation workflow
- Security requirement: PCI-DSS compliance for payment processing
- Section 9.3: Payment Gateway Integration Requirements

### Changed
- FR-PAT-001: NIK field changed from optional to required
  - Old: NIK optional (validate if provided)
  - New: NIK required (validate with Dukcapil)
- Performance SLA: API response time updated
  - Old: < 500ms (95th percentile)
  - New: < 300ms (95th percentile)
- Section 12: Updated timeline to include payment integration phase

### Reasons
- Government regulation now mandates NIK validation for all patients
- User feedback indicated system performance was too slow
- Business expansion requires online payment capability

---

## [1.1.0] - 2026-02-05 - APPROVED

**Document:** BRD - Clinic Management System
**Status:** approved
**Approved By:** Dr. Budi (Product Owner)
**Approval Date:** 2026-02-05

### Added
- FR-ENC-002: Upload medical images (X-ray, lab results)
- Integration requirement: PACS (Picture Archiving and Communication System)
- File upload size limit: 10MB per file
- Supported formats: JPEG, PNG, PDF, DICOM

### Changed
- FR-ENC-001: Added vital signs validation ranges
  - Blood Pressure: 50-300 mmHg (systolic), 30-200 mmHg (diastolic)
  - Heart Rate: 40-200 bpm
  - Temperature: 30-45°C
  - SpO2: 0-100%

### Reasons
- Medical team requested ability to attach diagnostic images
- Vital signs validation ranges ensure data quality

---

## [1.0.0] - 2026-02-01 - APPROVED

**Document:** BRD - Clinic Management System
**Status:** approved (Initial Release)
**Approved By:** Dr. Budi (Product Owner), Dr. Siti (Medical Director), Andi (IT Manager)
**Approval Date:** 2026-02-01

### Initial Requirements
- Module: Patient Management
  - FR-PAT-001: Create Patient
  - FR-PAT-002: Search Patient
  - FR-PAT-003: Update Patient
  - FR-PAT-004: Delete Patient
- Module: Clinical Encounter
  - FR-ENC-001: Create Encounter (with vital signs, diagnosis, treatment)
- Integration: Satu Sehat (mandatory)
- Security: JWT authentication, RBAC authorization
- Performance: Support 100 concurrent users, API < 500ms

### Reasons
- Initial baseline requirements for MVP
- Minimum viable product for clinic digitalization

---
```

---

## Tips for Maintaining CHANGELOG

1. **Always update CHANGELOG** when creating new document version
2. **Be specific** about what changed and why
3. **Reference requirement IDs** (e.g., FR-PAT-001)
4. **Include approver information** for approved versions
5. **Document reasons** for changes (regulation, feedback, business needs)
6. **Use semantic versioning** consistently
7. **Link to related changes** in other documents if applicable

---

## Change Categories

### Added
New features, sections, requirements that didn't exist before.

### Changed
Modifications to existing content. Always include:
- What changed
- Old value → New value
- Why it changed

### Deprecated
Content that is no longer recommended but still present for backward compatibility.

### Removed
Content that was completely removed from the document.

### Fixed
Corrections to errors, typos, or incorrect information.

---

## Cross-Document Changes

When changes in one document affect others, document the impact:

```markdown
## [1.2.0] - 2026-02-15 - APPROVED

**Document:** BRD v1.2
**Impacts:**
- API_SPEC: Patient module needs NIK validation endpoint
- SCHEMA: Patient table requires NIK column (NOT NULL)
- REQUIREMENTS: Update patient module validation rules

### Added
- FR-PAT-005: NIK validation via Dukcapil API

### Changed
- FR-PAT-001: NIK now required (was optional)

---
```

---

**Last Updated:** [DATE]
