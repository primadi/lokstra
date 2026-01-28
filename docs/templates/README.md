# Document Templates - README

This directory contains standardized templates for Lokstra project documentation.

## Available Templates

### 1. BRD_TEMPLATE.md
**Business Requirements Document**

Complete template for documenting business requirements, including:
- Executive summary & objectives
- Stakeholder information
- Business processes & workflows
- Functional & non-functional requirements
- Integration requirements
- Data requirements
- Risks & mitigation

**Use when:** Starting a new project or major feature

---

### 2. MODULE_REQUIREMENTS_TEMPLATE.md
**Module-Level Requirements**

Template for module-specific requirements extracted from BRD:
- Functional requirements per module
- Data models & entities
- Business rules & workflows
- Integration points
- Testing requirements

**Use when:** Breaking down BRD into module specifications

---

### 3. API_SPEC_TEMPLATE.md
**API Specification**

Comprehensive API documentation template:
- Endpoint definitions (CRUD + custom)
- Request/response formats
- Validation rules
- Error responses
- Data models
- Authentication & authorization
- Rate limiting

**Use when:** Designing module APIs before implementation

---

### 4. SCHEMA_TEMPLATE.md
**Database Schema**

Database schema documentation:
- Table definitions with constraints
- Indexes & performance optimization
- Relationships (1:N, M:N)
- Triggers & functions
- Migration files
- Data integrity rules
- Security & access control

**Use when:** Designing database schema for a module

---

## Template Usage Workflow

### Recommended Workflow:

```
Step 1: Business Requirements
├─> Use: BRD_TEMPLATE.md
├─> Output: docs/business/BRD.draft.md
└─> Review & Approve → BRD.v1.0.md

Step 2: Module Requirements
├─> Use: MODULE_REQUIREMENTS_TEMPLATE.md
├─> Output: docs/modules/<module>/REQUIREMENTS.md
└─> For each module identified in BRD

Step 3: API Specification
├─> Use: API_SPEC_TEMPLATE.md
├─> Output: docs/modules/<module>/API_SPEC.md
└─> Define all endpoints before coding

Step 4: Database Schema
├─> Use: SCHEMA_TEMPLATE.md
├─> Output: docs/modules/<module>/SCHEMA.md
└─> Design database structure

Step 5: Implementation
├─> Generate code from approved specs
└─> Use LOKSTRA-SKILLS.md for guidance
```

---

## Customization

### Customize for Your Industry

Templates can be customized for specific industries:

**Healthcare:**
- Add HIPAA compliance sections
- Include patient privacy requirements
- Add medical coding standards (ICD-10, SNOMED)

**Finance:**
- Add PCI-DSS requirements
- Include financial regulations
- Add transaction integrity rules

**E-commerce:**
- Add payment gateway requirements
- Include inventory management
- Add shipping/logistics specs

### Create Custom Templates

Save custom templates in:
```
docs/templates/industries/
├── healthcare/
│   ├── satu_sehat_brd.md
│   └── patient_module_requirements.md
├── finance/
│   └── payment_module_api_spec.md
└── ecommerce/
    └── product_schema.md
```

---

## Version Control

### Document Versioning

All documents use semantic versioning:

**Format:** `MAJOR.MINOR.PATCH`
- **MAJOR:** Breaking changes (scope change, major features)
- **MINOR:** New features (backward compatible)
- **PATCH:** Bug fixes, clarifications

**Status Values:**
- `draft` - Work in progress
- `review` - Under stakeholder review
- `approved` - Approved and locked
- `implemented` - Code implemented
- `deprecated` - No longer valid

**File Naming:**
```
docs/business/
├── BRD.md             # Current/latest (symlink or copy)
├── BRD.v1.0.md       # Approved version 1.0
├── BRD.v1.1.md       # Approved version 1.1
├── BRD.draft.md      # Working draft
└── CHANGELOG.md      # All changes tracked
```

---

## Best Practices

### 1. Design Before Code
- ✅ Write specifications first
- ✅ Get approval before implementation
- ✅ Lock approved versions
- ❌ Don't code without specs

### 2. Keep Documentation Updated
- Update specs when requirements change
- Sync code with approved specs
- Document all changes in CHANGELOG.md

### 3. Use Consistent Format
- Follow template structure
- Use same terminology across docs
- Maintain cross-references

### 4. Version Everything
- Track all document versions
- Never modify approved versions
- Create new draft for changes

### 5. Review Regularly
- Review specs before major releases
- Check spec-code consistency
- Update deprecated sections

---

## AI Agent Integration

These templates are designed to work with AI agents using **LOKSTRA-SKILLS.md**.

### AI Agent Capabilities:

1. **Generate from Templates**
   - AI can auto-populate templates
   - Extract from existing documents
   - Interview mode for requirements gathering

2. **Validate Consistency**
   - Check spec vs code
   - Validate cross-references
   - Ensure completeness

3. **Auto-Update**
   - Sync docs with code changes
   - Update CHANGELOGs
   - Generate migration docs

### Usage with AI:

```
# Generate BRD from scratch
User: "Create BRD for clinic system using BRD_TEMPLATE"
AI: Generate docs/business/BRD.draft.md

# Generate module spec from BRD
User: "Generate patient module requirements from BRD v1.0"
AI: Generate docs/modules/patient/REQUIREMENTS.md using MODULE_REQUIREMENTS_TEMPLATE

# Generate API spec
User: "Generate API spec for patient module"
AI: Generate docs/modules/patient/API_SPEC.md using API_SPEC_TEMPLATE

# Validate consistency
User: "Check if patient module code matches API spec"
AI: Compare implementation vs API_SPEC.md
```

---

## Examples

See `examples/` directory for filled-out templates:
- `examples/clinic_brd.md` - Complete healthcare BRD
- `examples/patient_module_requirements.md` - Patient module requirements
- `examples/patient_api_spec.md` - Patient API specification
- `examples/patient_schema.md` - Patient database schema

---

## Support

For questions or suggestions:
- Framework Documentation: https://primadi.github.io/lokstra/
- AI Agent Guide: https://primadi.github.io/lokstra/LOKSTRA-SKILLS
- Issues: https://github.com/primadi/lokstra/issues

---

**Last Updated:** January 27, 2026
