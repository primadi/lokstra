---
module: [MODULE_NAME]
version: [VERSION]
status: draft
based_on: BRD v[VERSION]
---

# [Module Name] - Requirements

## Module Information

| Field | Value |
|-------|-------|
| Module Name | [module_name] |
| Version | [VERSION] |
| Status | draft / approved / implemented |
| Based On | BRD v[VERSION] |
| Author | [AUTHOR] |
| Last Updated | [DATE] |

---

## 1. Overview

### 1.1 Purpose
[Describe the purpose of this module]

### 1.2 Scope
[Define what this module covers]

### 1.3 Bounded Context
[Define the bounded context in DDD terms]

---

## 2. Functional Requirements

### FR-[MOD]-001: [Requirement Title]

**Priority:** P0 (Must Have) | P1 (Should Have) | P2 (Nice to Have)

**User Story:**
> As a [user role], I want to [action], so that [benefit].

**Description:**
[Detailed description of the requirement]

**Inputs:**
- **[field_name]** (required/optional)
  - Type: [string / int / date / etc.]
  - Constraints: [min=3, max=100, etc.]
  - Description: [Description]

**Outputs:**
- **[field_name]**
  - Type: [type]
  - Description: [Description]

**Business Rules:**
1. [Business rule 1]
2. [Business rule 2]
3. [Business rule 3]

**Validation:**
- [Validation rule 1]
- [Validation rule 2]
- [Validation rule 3]

**Error Conditions:**
- [Error condition 1] → [Response]
- [Error condition 2] → [Response]

**Integration:**
- [External System 1]: [Integration description]
- [External System 2]: [Integration description]

**Acceptance Criteria:**
- [ ] [Criterion 1]
- [ ] [Criterion 2]
- [ ] [Criterion 3]

---

### FR-[MOD]-002: [Requirement Title]
[Follow same structure]

---

### FR-[MOD]-003: [Requirement Title]
[Follow same structure]

---

## 3. Data Model

### 3.1 Entities

#### [Entity Name 1]
**Description:** [Entity description]

**Attributes:**
- `id`: Primary key, unique identifier
- `[field1]`: [Description]
- `[field2]`: [Description]
- `created_at`: Record creation timestamp
- `updated_at`: Last update timestamp
- `deleted_at`: Soft delete timestamp (nullable)

**Relationships:**
- Has many [Entity 2] (1:N)
- Belongs to [Entity 3] (N:1)

**Constraints:**
- [Constraint 1]
- [Constraint 2]

---

#### [Entity Name 2]
[Follow same structure]

---

### 3.2 Value Objects
- **[ValueObject1]:** [Description]
- **[ValueObject2]:** [Description]

### 3.3 Enumerations
- **[Enum1]:** [values: value1, value2, value3]
- **[Enum2]:** [values: value1, value2, value3]

---

## 4. Business Logic

### 4.1 Business Rules

#### BR-[MOD]-001: [Rule Title]
**Description:** [Business rule description]

**Trigger:** [When this rule applies]

**Condition:** [Rule condition]

**Action:** [What happens when rule is triggered]

**Example:**
```
Given: [Precondition]
When: [Action]
Then: [Expected result]
```

---

#### BR-[MOD]-002: [Rule Title]
[Follow same structure]

---

### 4.2 Calculations

#### [Calculation Name 1]
**Formula:** [Mathematical formula or algorithm]

**Inputs:**
- [Input 1]
- [Input 2]

**Output:** [Calculated value]

**Example:**
```
Input: field1 = 100, field2 = 20
Calculation: result = field1 * field2 / 100
Output: result = 20
```

---

### 4.3 Workflows

#### Workflow 1: [Workflow Name]

```
[Actor] → Action 1 → System validates → 
  Success: Action 2 → Complete
  Failure: Show error → Retry
```

**Steps:**
1. [Step 1]
2. [Step 2]
3. [Step 3]

**Decision Points:**
- **If [condition]:** [Path A]
- **Else:** [Path B]

---

## 5. Integration Requirements

### 5.1 Internal Module Dependencies

| Module | Purpose | Data Exchanged |
|--------|---------|----------------|
| [Module 1] | [Purpose] | [Data] |
| [Module 2] | [Purpose] | [Data] |

### 5.2 External System Integration

#### [External System 1]
**Purpose:** [Integration purpose]

**Protocol:** [REST API / SOAP / GraphQL / Message Queue]

**Authentication:** [Auth method]

**Endpoints:**
- `[METHOD] /path` - [Description]

**Data Format:** [JSON / XML / etc.]

**Error Handling:** [Error handling approach]

---

## 6. Non-Functional Requirements

### 6.1 Performance
- API response time: [target]
- Database query time: [target]
- Concurrent users: [number]

### 6.2 Security
- Authentication: [method]
- Authorization: [model]
- Data encryption: [requirements]
- Audit logging: [requirements]

### 6.3 Data Retention
- Active records: [retention period]
- Deleted records: [retention period]
- Audit logs: [retention period]

### 6.4 Scalability
- Expected data volume: [estimate]
- Growth rate: [percentage per year]

---

## 7. User Interface Requirements

### 7.1 Screens/Views
- **[Screen 1]:** [Description and purpose]
- **[Screen 2]:** [Description and purpose]

### 7.2 User Interactions
- [Interaction 1]
- [Interaction 2]

### 7.3 Validation Messages
| Field | Validation | Error Message |
|-------|-----------|---------------|
| [field1] | [rule] | "[Message]" |
| [field2] | [rule] | "[Message]" |

---

## 8. Testing Requirements

### 8.1 Unit Tests
- [ ] Test [functionality 1]
- [ ] Test [functionality 2]
- [ ] Test validation rules
- [ ] Test business rules
- [ ] Test error handling

### 8.2 Integration Tests
- [ ] Test [integration 1]
- [ ] Test [integration 2]

### 8.3 Test Data
**Test Scenario 1:**
```
Input: [test data]
Expected Output: [expected result]
```

**Test Scenario 2:**
```
Input: [test data]
Expected Output: [expected result]
```

---

## 9. Migration Requirements

### 9.1 Data Migration
- **Source:** [Source system/database]
- **Volume:** [Estimated record count]
- **Mapping:** [Field mappings]
- **Validation:** [Validation rules]

### 9.2 Migration Steps
1. [Step 1]
2. [Step 2]
3. [Step 3]

---

## 10. Documentation

### 10.1 Required Documentation
- [ ] API Specification
- [ ] Database Schema
- [ ] User Manual
- [ ] Admin Guide
- [ ] Deployment Guide

---

## 11. Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| [Risk 1] | High/Medium/Low | High/Medium/Low | [Strategy] |
| [Risk 2] | High/Medium/Low | High/Medium/Low | [Strategy] |

---

## 12. Dependencies

### 12.1 Technical Dependencies
- [Dependency 1]: [Description]
- [Dependency 2]: [Description]

### 12.2 Business Dependencies
- [Dependency 1]: [Description]
- [Dependency 2]: [Description]

---

## 13. Acceptance Criteria

### Module-Level Acceptance
- [ ] All functional requirements implemented
- [ ] All business rules enforced
- [ ] All validations working
- [ ] All tests passing (> 80% coverage)
- [ ] API documentation complete
- [ ] Database migrations working
- [ ] Error handling implemented
- [ ] Logging implemented
- [ ] Performance targets met

---

## 14. Glossary

| Term | Definition |
|------|------------|
| [Term 1] | [Definition] |
| [Term 2] | [Definition] |

---

## 15. References

- BRD v[VERSION]
- [Other reference documents]

---

## 16. Change History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| [VERSION] | [DATE] | [AUTHOR] | [CHANGES] |

---

**End of Requirements Document**
