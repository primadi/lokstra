# Lokstra Framework - AI Agent Skills Guide

**Version:** 2.0.0  
**Last Updated:** January 27, 2026  
**Target Audience:** AI Agents (GitHub Copilot, Claude, ChatGPT, Cursor, etc.)  
**Framework:** Lokstra Framework (github.com/primadi/lokstra)

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Philosophy: Document-Driven Development](#philosophy-document-driven-development)
3. [Project Structure Standards](#project-structure-standards)
4. [Document Templates](#document-templates)
5. [Skills Reference](#skills-reference)
6. [Workflow Examples](#workflow-examples)
7. [Best Practices](#best-practices)

---

## Overview

This guide provides **actionable skills** for AI Agents to help developers build production-ready applications using the **Lokstra Framework**. 

### Key Principles

1. **📐 Design First, Code Second** - Always create specifications before implementation
2. **📝 Single Source of Truth** - Approved documentation drives all code generation
3. **✅ Consistency Enforcement** - AI validates that code matches specifications
4. **🔄 Version Control** - All documents are versioned and tracked
5. **🚀 Productivity** - Generate complete modules (handler + repo + tests + migrations) from specs

### Skills Overview

| Phase | Skills | Purpose |
|-------|--------|---------|
| **Phase 0: Design** | SKILL 0-3 | BRD creation, module design, consistency check |
| **Phase 1: Setup** | SKILL 4-5 | Project creation, module scaffolding |
| **Phase 2: Implementation** | SKILL 6-10 | Code generation from specs |
| **Phase 3: Testing** | SKILL 11-12 | Test generation, validation |
| **Phase 4: Maintenance** | SKILL 13-14 | Documentation sync, updates |

---

## Philosophy: Document-Driven Development

### Traditional Flow (Chaotic) ❌
```
User: "Buat endpoint create patient"
  ↓
AI: *generates random code*
  ↓
User: "Eh, validasinya salah"
  ↓
AI: *fixes code* (inconsistent with tests)
  ↓
User: "Test juga salah"
  ↓
Result: Wasted time, inconsistent code
```

### Lokstra Flow (Structured) ✅
```
Phase 0: DESIGN
User: "Buat design patient module"
  ↓
AI: Generate docs/modules/patient/
    - REQUIREMENTS.md
    - API_SPEC.md
    - SCHEMA.md
  ↓
User: Review & approve specs
  ↓

Phase 1: IMPLEMENT
User: "Generate implementation dari API spec"
  ↓
AI: Read approved specs
  ↓
AI: Generate complete module:
    ✅ handler + @Route annotations
    ✅ repository + @Service
    ✅ domain models + validation
    ✅ migrations (from SCHEMA)
    ✅ unit tests (from API_SPEC)
    ✅ .http files (from API_SPEC)
  ↓

Phase 2: VALIDATE
AI: Auto-check consistency:
    - API_SPEC ↔ handler routes
    - SCHEMA ↔ migrations
    - Validation rules ↔ domain structs
  ↓
Result: Production-ready, consistent, tested code
```

---

## Project Structure Standards

### Standard Lokstra Project Structure

```
myproject/
├── docs/                           # 📐 Design documents (versioned)
│   ├── business/
│   │   ├── BRD.md                 # Current approved BRD
│   │   ├── BRD.v1.0.md           # Version 1.0 (immutable)
│   │   ├── BRD.v1.1.md           # Version 1.1 (immutable)
│   │   ├── BRD.draft.md          # Working draft
│   │   └── CHANGELOG.md
│   ├── architecture/
│   │   ├── SYSTEM_DESIGN.md
│   │   ├── ERD.md
│   │   └── CHANGELOG.md
│   ├── api/
│   │   ├── API_CONVENTIONS.md
│   │   └── API_OVERVIEW.md
│   └── modules/
│       └── patient/
│           ├── REQUIREMENTS.md    # Module requirements (versioned)
│           ├── API_SPEC.md        # API endpoints (versioned)
│           ├── SCHEMA.md          # Database schema (versioned)
│           └── CHANGELOG.md
│
├── modules/                        # 💻 Implementation
│   ├── shared/                    # Cross-module types
│   │   └── domain/
│   │       └── patient_ref.go
│   └── patient/
│       ├── handler/
│       │   ├── patient_handler.go       # @Handler with @Route
│       │   ├── patient_handler_test.go
│       │   └── patient.http             # Manual test file
│       ├── repository/
│       │   ├── contract.go              # Repository interfaces
│       │   └── postgres/
│       │       ├── patient_repository.go     # @Service
│       │       └── patient_repository_test.go
│       └── domain/
│           ├── patient.go               # Entity + related types
│           ├── patient_dto.go           # Request/Response DTOs
│           └── service.go               # Service interfaces
│
├── migrations/                     # 🗄️ Database migrations
│   └── patient/
│       ├── 001_create_patients.up.sql
│       └── 001_create_patients.down.sql
│
├── config/
│   └── config.yaml                # Service configuration
│
├── main.go                        # Application entry point
├── go.mod
├── go.sum
├── .gitignore
└── README.md
```

---

## Document Templates

All document templates are located in: `docs/templates/`

### Document Versioning

**Format:** `MAJOR.MINOR.PATCH`
- **MAJOR:** Breaking changes (scope change)
- **MINOR:** New features (backward compatible)
- **PATCH:** Bug fixes, clarifications

**Status Values:**
- `draft` - Work in progress
- `review` - Under stakeholder review
- `approved` - Approved and locked
- `implemented` - Code implemented
- `deprecated` - No longer valid

### Available Templates

1. **BRD Template** - `docs/templates/BRD_TEMPLATE.md`
2. **Module Requirements** - `docs/templates/MODULE_REQUIREMENTS_TEMPLATE.md`
3. **API Specification** - `docs/templates/API_SPEC_TEMPLATE.md`
4. **Database Schema** - `docs/templates/SCHEMA_TEMPLATE.md`

---

## Skills Reference

### Phase 0: Design & Documentation

---

#### SKILL 0: Create BRD from Input

**Purpose:** Generate Business Requirements Document from various input methods

**Input Methods:**

##### Method 1: Simple Prompt
```
User: "Buat BRD untuk backend klinik dengan Lokstra framework, 
       sesuai Satu Sehat, fitur: patient management, encounter, 
       e-prescription"

AI Actions:
1. Analyze requirements from prompt
2. Load BRD template
3. Expand to functional requirements
4. Add compliance requirements (Satu Sehat)
5. Add Lokstra technical constraints
6. Generate docs/business/BRD.draft.md
```

##### Method 2: Interview Mode
```
AI: Start interview mode with questions:
    1. "Nama project dan tujuan bisnis?"
    2. "Target users dan jumlahnya?"
    3. "Fitur utama yang dibutuhkan?"
    4. "Integrasi eksternal?"
    5. "Performance requirements?"
    
User: Answers each question

AI Actions:
1. Collect all answers
2. Generate comprehensive BRD
3. Output: docs/business/BRD.draft.md
```

##### Method 3: Import Existing Document
```
User: Upload existing_brd.docx

AI Actions:
1. Parse document (extract text)
2. Identify sections:
   - Business objectives
   - Functional requirements
   - Non-functional requirements
   - Integration requirements
3. Convert to Lokstra BRD format
4. Output: docs/business/BRD.draft.md (for review)
```

##### Method 4: Industry Template
```
User: "Buat BRD untuk clinic system based on Satu Sehat template"

AI Actions:
1. Load template: docs/templates/industries/healthcare/satu_sehat_brd.md
2. Customize with user input
3. Output: docs/business/BRD.draft.md
```

**Output:**
- `docs/business/BRD.draft.md` (status: draft)
- `docs/business/CHANGELOG.md` (initial entry)

**Validation:**
- All required sections present
- Functional requirements have unique IDs (FR-XXX-NNN)
- Version metadata complete

---

#### SKILL 0.1: Approve BRD Version

**Purpose:** Lock BRD draft as approved version

**Input:**
```
User: "Approve BRD draft as v1.0"
User provides: approved_by, approval_date (optional, defaults to today)
```

**AI Actions:**
1. Read `docs/business/BRD.draft.md`
2. Update metadata:
   - `status: draft` → `approved`
   - `version: 1.0`
   - `approved_by: [user input]`
   - `approved_date: [current date]`
3. Save as `docs/business/BRD.v1.0.md` (immutable)
4. Update `docs/business/BRD.md` → copy of v1.0
5. Add entry to `docs/business/CHANGELOG.md`
6. Delete or archive `BRD.draft.md`

**Output:**
- `docs/business/BRD.v1.0.md` (locked)
- `docs/business/BRD.md` (latest approved)
- `docs/business/CHANGELOG.md` (updated)

---

#### SKILL 0.2: Create BRD Change Request

**Purpose:** Create new version of BRD with changes

**Input:**
```
User: "Create BRD v1.1 draft with appointment scheduling feature"
```

**AI Actions:**
1. Copy current approved version → `BRD.draft.md`
2. Update version to `1.1` (status: draft)
3. Add new requirements
4. Add CHANGELOG entry (draft changes)

**Output:**
- `docs/business/BRD.draft.md` (v1.1 draft)

---

#### SKILL 1: Generate Module Requirements from BRD

**Purpose:** Extract module-specific requirements from approved BRD

**Input:**
```
User: "Generate module requirements dari BRD v1.0"
```

**AI Actions:**
1. Read `docs/business/BRD.v1.0.md`
2. Identify modules from functional requirements:
   - FR-PAT-* → patient module
   - FR-ENC-* → clinical module
   - FR-PRE-* → prescription module
3. For each module, generate:
   - `docs/modules/<module>/REQUIREMENTS.md`
   - `docs/modules/<module>/API_SPEC.md`
   - `docs/modules/<module>/SCHEMA.md`
   - `docs/modules/<module>/CHANGELOG.md`

**Output Structure:**

**docs/modules/patient/REQUIREMENTS.md:**
```markdown
---
module: patient
version: 1.0
status: approved
based_on: BRD v1.0
---

# Patient Module - Requirements

## Functional Requirements

### FR-PAT-001: Create Patient
[Copy from BRD with full details]

### FR-PAT-002: Search Patient
[Copy from BRD with full details]

## Data Model
- Patient entity
- Patient address (1:N)
- Emergency contact (1:N)
```

---

#### SKILL 2: Generate API Specification from Requirements

**Purpose:** Create detailed API spec from requirements

**Input:**
```
User: "Generate API spec for patient module"
```

**AI Actions:**
1. Read `docs/modules/patient/REQUIREMENTS.md`
2. For each functional requirement:
   - Extract CRUD operations
   - Define HTTP method + endpoint
   - Define request/response format
   - Extract validation rules
   - Define error responses
3. Generate `docs/modules/patient/API_SPEC.md`

**Output Example:**

```markdown
---
module: patient
version: 1.0
status: approved
---

# Patient Module - API Specification

## Base URL
`/api/patients`

## Authentication
All endpoints require JWT token in `Authorization: Bearer <token>` header

---

## Endpoints

### 1. Create Patient
**Endpoint:** `POST /api/patients`
**Authorization:** Required (role: admin, nurse)

**Request Body:**
```json
{
  "name": "John Doe",
  "mrn": "PAT-20260127-001",
  "dob": "1990-01-15",
  "gender": "male",
  "nik": "3174012345678901",
  "phone": "+628123456789",
  "email": "john@example.com",
  "address": {
    "street": "Jl. Sudirman 123",
    "city": "Jakarta",
    "province": "DKI Jakarta",
    "postal_code": "12190"
  }
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid-123",
  "mrn": "PAT-20260127-001",
  "name": "John Doe",
  "created_at": "2026-01-27T10:00:00Z"
}
```

**Validation Rules:**
- `name`: required, min 3, max 100, letters + spaces only
- `mrn`: required, unique, format: `PAT-YYYYMMDD-XXX`
- `dob`: required, date, must be past
- `gender`: required, enum: male|female|other
- `nik`: optional, exactly 16 digits
- `phone`: required, E.164 format
- `email`: optional, valid email

**Business Rules:**
- MRN must be unique across all patients
- NIK validated with Dukcapil API (if provided)
- Age must be >= 0
- Cannot create patient with future DOB

**Error Responses:**
- `400` - Validation error
  ```json
  {
    "error": "validation_error",
    "message": "Invalid input",
    "details": [
      {
        "field": "email",
        "error": "invalid email format"
      }
    ]
  }
  ```
- `409` - MRN already exists
- `500` - Server error

**Integration:**
- Sync to Satu Sehat within 1 minute
- Validate NIK with Dukcapil (optional)

---

### 2. Get Patient by ID
**Endpoint:** `GET /api/patients/{id}`
[... similar detail]
```

---

#### SKILL 3: Generate Database Schema from Requirements

**Purpose:** Create database schema specification

**Input:**
```
User: "Generate database schema for patient module"
```

**AI Actions:**
1. Read `docs/modules/patient/REQUIREMENTS.md`
2. Extract entities and attributes
3. Define tables, columns, types, constraints
4. Define relationships
5. Define indexes
6. Generate `docs/modules/patient/SCHEMA.md`

**Output Example:**

```markdown
---
module: patient
version: 1.0
status: approved
---

# Patient Module - Database Schema

## Tables

### patients
Primary entity for patient records

```sql
CREATE TABLE patients (
    id VARCHAR(50) PRIMARY KEY,
    mrn VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    dob DATE NOT NULL,
    gender VARCHAR(10) NOT NULL,
    nik VARCHAR(16) UNIQUE,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_gender CHECK (gender IN ('male', 'female', 'other')),
    CONSTRAINT chk_status CHECK (status IN ('active', 'inactive', 'deceased')),
    CONSTRAINT chk_dob CHECK (dob <= CURRENT_DATE)
);

CREATE UNIQUE INDEX idx_patients_mrn ON patients(mrn);
CREATE INDEX idx_patients_nik ON patients(nik) WHERE nik IS NOT NULL;
CREATE INDEX idx_patients_status ON patients(status);
CREATE INDEX idx_patients_created ON patients(created_at);
```

**Columns:**
- `id`: Primary key, UUID format
- `mrn`: Medical Record Number, unique identifier
- `name`: Patient full name
- `dob`: Date of birth
- `gender`: male|female|other
- `nik`: National ID (optional)
- `phone`: Contact number
- `email`: Email address (optional)
- `status`: active|inactive|deceased
- `created_at`: Record creation timestamp
- `updated_at`: Last update timestamp
- `deleted_at`: Soft delete timestamp

**Indexes:**
- Primary key on `id`
- Unique index on `mrn` (for fast MRN lookup)
- Partial index on `nik` (only non-null values)
- Index on `status` (for filtering)
- Index on `created_at` (for sorting/pagination)

---

### patient_addresses
Related addresses for patient

```sql
CREATE TABLE patient_addresses (
    id SERIAL PRIMARY KEY,
    patient_id VARCHAR(50) NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    street TEXT NOT NULL,
    city VARCHAR(100),
    province VARCHAR(100),
    postal_code VARCHAR(10),
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT chk_address_type CHECK (type IN ('home', 'work', 'billing'))
);

CREATE INDEX idx_addresses_patient ON patient_addresses(patient_id);
CREATE INDEX idx_addresses_primary ON patient_addresses(patient_id, is_primary);
```

## Relationships

```
patients (1) ─────< (*) patient_addresses
```

## Data Constraints

- One patient can have multiple addresses
- Only one address can be marked as primary per patient
- Soft delete on patients (deleted_at timestamp)
- Cascade delete on addresses when patient deleted
```

---

#### SKILL 4: Consistency Check - Documents vs Code

**Purpose:** Validate that implementation matches approved specifications

**Input:**
```
User: "Check consistency patient module"
```

**AI Actions:**
1. Read approved docs:
   - `docs/modules/patient/API_SPEC.md`
   - `docs/modules/patient/SCHEMA.md`
2. Read implementation:
   - `modules/patient/handler/patient_handler.go`
   - `modules/patient/domain/patient.go`
   - `migrations/patient/*.sql`
3. Compare:
   - API endpoints ↔ @Route annotations
   - Request DTOs ↔ validation tags
   - Business rules ↔ handler logic
   - Schema tables ↔ migration SQL
   - Schema columns ↔ domain struct fields
4. Generate report

**Output Example:**

```markdown
# Consistency Check Report - Patient Module
Generated: 2026-01-27 10:30:00

## Summary
✅ 8 checks passed
❌ 2 checks failed
⚠️  1 warning

---

## API Endpoints

### ✅ POST /api/patients
- Spec: `POST /api/patients`
- Code: `@Route "POST /"` in patient_handler.go:15
- Status: ✅ Implemented

### ✅ GET /api/patients/{id}
- Spec: `GET /api/patients/{id}`
- Code: `@Route "GET /{id}"` in patient_handler.go:45
- Status: ✅ Implemented

### ❌ PUT /api/patients/{id}
- Spec: `PUT /api/patients/{id}` (API_SPEC.md line 120)
- Code: NOT FOUND
- Status: ❌ Missing implementation

---

## Validation Rules

### ✅ Patient.Name
- Spec: required, min 3, max 100
- Code: `validate:"required,min=3,max=100"`
- Status: ✅ Match

### ❌ Patient.Email
- Spec: optional, valid email (API_SPEC.md line 85)
- Code: `validate:"required,email"` (domain/patient.go:12)
- Status: ❌ Mismatch - Spec says optional, code has required

---

## Database Schema

### ✅ Table: patients
- Spec: 12 columns defined
- Migration: 12 columns created (001_patients.up.sql)
- Status: ✅ Match

### ⚠️  Index: idx_patients_email
- Spec: NOT defined
- Migration: Created in 001_patients.up.sql:25
- Status: ⚠️  Extra index (not in spec)

---

## Recommendations

1. ❌ **CRITICAL:** Implement missing endpoint PUT /api/patients/{id}
2. ❌ **CRITICAL:** Fix Patient.Email validation (should be optional)
3. ⚠️  **MINOR:** Document idx_patients_email index in SCHEMA.md or remove

---

## Next Steps

```bash
# Fix email validation
# Update: modules/patient/domain/patient.go:12
validate:"omitempty,email"  # Add omitempty

# Implement update endpoint
# Add to: modules/patient/handler/patient_handler.go
# @Route "PUT /{id}"
func (h *PatientHandler) Update(req *UpdatePatientRequest) error { ... }
```
```

---

### Phase 1: Project & Module Setup

---

#### SKILL 5: Create Lokstra Project

**Purpose:** Initialize new Lokstra project with complete structure

**Input:**
```
User: "Create Lokstra project: clinic-management"
Options:
  - project_name: string (required)
  - module_name: string (default: "core")
  - database: "postgres" | "mysql" | "none" (default: "postgres")
  - with_example: boolean (default: true)
```

**AI Actions:**
1. Create project structure
2. Generate `go.mod`
3. Create `main.go` with bootstrap
4. Create `config/config.yaml`
5. Create documentation structure
6. Create example module (if with_example=true)
7. Create `.gitignore`
8. Create `README.md`

**Generated Structure:**

```
clinic-management/
├── docs/
│   ├── business/
│   │   ├── BRD.draft.md (template)
│   │   └── CHANGELOG.md
│   ├── architecture/
│   │   ├── SYSTEM_DESIGN.md
│   │   └── ERD.md
│   ├── api/
│   │   └── API_CONVENTIONS.md
│   └── modules/
│       └── example/
│           ├── REQUIREMENTS.md
│           ├── API_SPEC.md
│           └── SCHEMA.md
│
├── modules/
│   ├── shared/
│   │   └── domain/
│   │       └── types.go
│   └── example/                     # Working example module
│       ├── handler/
│       │   ├── example_handler.go   # @Handler with routes
│       │   ├── example_handler_test.go
│       │   └── example.http
│       ├── repository/
│       │   ├── contract.go
│       │   └── postgres/
│       │       ├── example_repository.go  # @Service
│       │       └── example_repository_test.go
│       └── domain/
│           ├── example.go           # Entity
│           └── example_dto.go       # DTOs
│
├── migrations/
│   └── example/
│       └── 001_init.up.sql
│
├── config/
│   └── config.yaml
│
├── main.go
├── go.mod
├── go.sum
├── .gitignore
└── README.md
```

**Generated Files:**

**main.go:**
```go
package main

import (
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/services/dbpool_pg"
)

func main() {
	// Register middleware
	recovery.Register()
	request_logger.Register()
	
	// Register database service
	dbpool_pg.Register()

	// Bootstrap and run
	lokstra_init.BootstrapAndRun(
		// lokstra_init.WithLogLevel(logger.LogLevelDebug),
		// lokstra_init.WithDbMigrations(true, "migrations"),
	)
}
```

**config/config.yaml:**
```yaml
# yaml-language-server: $schema=https://primadi.github.io/lokstra/schema/lokstra.schema.json

configs:
  database:
    dsn: ${DB_DSN:postgres://postgres:password@localhost:5432/clinic_db}
    schema: ${DB_SCHEMA:public}
  
  repository:
    example-repository: postgres-example-repository

service-definitions:
  db_main:
    type: dbpool_pg
    config:
      dsn: ${GLOBAL_DB_DSN:postgres://postgres:password@localhost:5432/clinic_db}
      schema: ${GLOBAL_DB_SCHEMA:public}

servers:
  api-server:
    base-url: "http://localhost"
    addr: ":3000"
    published-services: [example-handler]
```

**modules/example/handler/example_handler.go:**
```go
package handler

import (
	"clinic-management/modules/example/domain"
	"clinic-management/modules/example/repository"
)

// @Handler name="example-handler", prefix="/api/examples", middlewares=["recovery", "request_logger"]
type ExampleHandler struct {
	// @Inject "@repository.example-repository"
	ExampleRepo repository.ExampleRepository
}

// @Route "GET /"
func (h *ExampleHandler) List() ([]*domain.Example, error) {
	return h.ExampleRepo.List()
}

// @Route "GET /{id}"
func (h *ExampleHandler) GetByID(id string) (*domain.Example, error) {
	return h.ExampleRepo.GetByID(id)
}

// @Route "POST /"
func (h *ExampleHandler) Create(req *domain.CreateExampleRequest) (*domain.Example, error) {
	example := &domain.Example{
		Name:        req.Name,
		Description: req.Description,
	}
	return h.ExampleRepo.Create(example)
}

// @Route "PUT /{id}"
func (h *ExampleHandler) Update(id string, req *domain.UpdateExampleRequest) (*domain.Example, error) {
	example := &domain.Example{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}
	return h.ExampleRepo.Update(example)
}

// @Route "DELETE /{id}"
func (h *ExampleHandler) Delete(id string) error {
	return h.ExampleRepo.Delete(id)
}
```

**README.md:**
```markdown
# Clinic Management System

Lokstra-powered clinic management backend.

## Quick Start

```bash
# Install dependencies
go mod download

# Run migrations (optional)
# Update main.go: lokstra_init.WithDbMigrations(true, "migrations")

# Run server
go run .

# Server running at http://localhost:3000
```

## API Endpoints

Example endpoints:
- `GET /api/examples` - List all examples
- `GET /api/examples/{id}` - Get example by ID
- `POST /api/examples` - Create example
- `PUT /api/examples/{id}` - Update example
- `DELETE /api/examples/{id}` - Delete example

Test with: `modules/example/handler/example.http`

## Project Structure

- `docs/` - Design documents (BRD, API specs, schemas)
- `modules/` - Business modules
- `migrations/` - Database migrations
- `config/` - Configuration files

## Documentation

See `docs/` folder for complete documentation.
```

**Output:**
- Complete working project
- Example module demonstrates Lokstra patterns
- Ready to run: `go run .`
- Developer can rename `example` → actual module name

---

#### SKILL 6: Add Module from Specification

**Purpose:** Generate complete module from approved documentation

**Input:**
```
User: "Generate patient module from spec"

Prerequisites:
- docs/modules/patient/API_SPEC.md exists
- docs/modules/patient/SCHEMA.md exists
```

**AI Actions:**
1. Read `docs/modules/patient/API_SPEC.md`
2. Read `docs/modules/patient/SCHEMA.md`
3. Generate complete module structure
4. Generate all files from spec
5. Update `config/config.yaml`
6. Generate migrations from SCHEMA

**Generated Files:**

```
modules/patient/
├── handler/
│   ├── patient_handler.go          # Generated from API_SPEC endpoints
│   ├── patient_handler_test.go     # Generated test cases
│   └── patient.http                # Generated HTTP test file
├── repository/
│   ├── contract.go                 # Generated interfaces
│   └── postgres/
│       ├── patient_repository.go   # Generated @Service implementation
│       └── patient_repository_test.go
└── domain/
    ├── patient.go                  # Generated entity from SCHEMA
    ├── patient_dto.go              # Generated DTOs from API_SPEC
    └── service.go                  # Generated interfaces

migrations/patient/
├── 001_create_patients.up.sql      # Generated from SCHEMA
└── 001_create_patients.down.sql
```

**Generation Logic:**

1. **Domain Entity (from SCHEMA.md):**
```go
// Generated from: docs/modules/patient/SCHEMA.md

package domain

import "time"

type Patient struct {
	ID        string     `json:"id" db:"id"`
	MRN       string     `json:"mrn" db:"mrn"`
	Name      string     `json:"name" db:"name" validate:"required,min=3,max=100"`
	DOB       time.Time  `json:"dob" db:"dob" validate:"required"`
	Gender    string     `json:"gender" db:"gender" validate:"required,oneof=male female other"`
	NIK       *string    `json:"nik,omitempty" db:"nik" validate:"omitempty,len=16"`
	Phone     string     `json:"phone" db:"phone" validate:"required,e164"`
	Email     *string    `json:"email,omitempty" db:"email" validate:"omitempty,email"`
	Status    string     `json:"status" db:"status" validate:"required,oneof=active inactive deceased"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
```

2. **DTOs (from API_SPEC.md):**
```go
// Generated from: docs/modules/patient/API_SPEC.md

package domain

type CreatePatientRequest struct {
	Name    string  `json:"name" validate:"required,min=3,max=100"`
	MRN     string  `json:"mrn" validate:"required"`
	DOB     string  `json:"dob" validate:"required"`
	Gender  string  `json:"gender" validate:"required,oneof=male female other"`
	NIK     *string `json:"nik,omitempty" validate:"omitempty,len=16"`
	Phone   string  `json:"phone" validate:"required,e164"`
	Email   *string `json:"email,omitempty" validate:"omitempty,email"`
	Address Address `json:"address" validate:"required"`
}

type UpdatePatientRequest struct {
	Name  string  `json:"name" validate:"required,min=3,max=100"`
	Phone string  `json:"phone" validate:"required,e164"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

type Address struct {
	Street     string `json:"street" validate:"required"`
	City       string `json:"city" validate:"required"`
	Province   string `json:"province" validate:"required"`
	PostalCode string `json:"postal_code" validate:"required"`
}
```

3. **Handler (from API_SPEC.md):**
```go
// Generated from: docs/modules/patient/API_SPEC.md

package handler

import (
	"myproject/modules/patient/domain"
	"myproject/modules/patient/repository"
)

// @Handler name="patient-handler", prefix="/api/patients", middlewares=["recovery", "request_logger"]
type PatientHandler struct {
	// @Inject "@repository.patient-repository"
	PatientRepo repository.PatientRepository
}

// @Route "POST /"
// Authorization: admin, nurse
func (h *PatientHandler) Create(req *domain.CreatePatientRequest) (*domain.Patient, error) {
	// TODO: Validate NIK with Dukcapil
	// TODO: Sync to Satu Sehat
	
	patient := &domain.Patient{
		Name:   req.Name,
		MRN:    req.MRN,
		Gender: req.Gender,
		Phone:  req.Phone,
		Status: "active",
	}
	
	return h.PatientRepo.Create(patient)
}

// @Route "GET /{id}"
func (h *PatientHandler) GetByID(id string) (*domain.Patient, error) {
	return h.PatientRepo.GetByID(id)
}

// @Route "PUT /{id}"
func (h *PatientHandler) Update(id string, req *domain.UpdatePatientRequest) (*domain.Patient, error) {
	patient, err := h.PatientRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	patient.Name = req.Name
	patient.Phone = req.Phone
	patient.Email = req.Email
	
	return h.PatientRepo.Update(patient)
}

// @Route "DELETE /{id}"
func (h *PatientHandler) Delete(id string) error {
	return h.PatientRepo.Delete(id)
}
```

4. **Repository Interface:**
```go
// Generated from: docs/modules/patient/SCHEMA.md + API_SPEC.md

package repository

import (
	"context"
	"myproject/modules/patient/domain"
)

type PatientRepository interface {
	Create(ctx context.Context, patient *domain.Patient) (*domain.Patient, error)
	GetByID(ctx context.Context, id string) (*domain.Patient, error)
	GetByMRN(ctx context.Context, mrn string) (*domain.Patient, error)
	Update(ctx context.Context, patient *domain.Patient) (*domain.Patient, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*domain.Patient, error)
}
```

5. **Repository Implementation:**
```go
// Generated from: docs/modules/patient/SCHEMA.md

package postgres

import (
	"context"
	"database/sql"
	"myproject/modules/patient/domain"
	"myproject/modules/patient/repository"
	"github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-patient-repository"
type PostgresPatientRepository struct {
	// @Inject "db_main"
	dbPool serviceapi.DbPool
}

var _ repository.PatientRepository = (*PostgresPatientRepository)(nil)

func (r *PostgresPatientRepository) Create(ctx context.Context, patient *domain.Patient) (*domain.Patient, error) {
	query := `
		INSERT INTO patients (id, mrn, name, dob, gender, nik, phone, email, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	
	err := r.dbPool.QueryRow(ctx, query,
		patient.ID, patient.MRN, patient.Name, patient.DOB, patient.Gender,
		patient.NIK, patient.Phone, patient.Email, patient.Status,
	).Scan(&patient.ID, &patient.CreatedAt, &patient.UpdatedAt)
	
	return patient, err
}

func (r *PostgresPatientRepository) GetByID(ctx context.Context, id string) (*domain.Patient, error) {
	query := `
		SELECT id, mrn, name, dob, gender, nik, phone, email, status, created_at, updated_at, deleted_at
		FROM patients
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	patient := &domain.Patient{}
	err := r.dbPool.QueryRow(ctx, query, id).Scan(
		&patient.ID, &patient.MRN, &patient.Name, &patient.DOB, &patient.Gender,
		&patient.NIK, &patient.Phone, &patient.Email, &patient.Status,
		&patient.CreatedAt, &patient.UpdatedAt, &patient.DeletedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return patient, err
}

// ... other methods
```

6. **Migration (from SCHEMA.md):**
```sql
-- Generated from: docs/modules/patient/SCHEMA.md
-- Migration: Create patients table
-- Version: 1.0

CREATE TABLE IF NOT EXISTS patients (
    id VARCHAR(50) PRIMARY KEY,
    mrn VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    dob DATE NOT NULL,
    gender VARCHAR(10) NOT NULL,
    nik VARCHAR(16) UNIQUE,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_gender CHECK (gender IN ('male', 'female', 'other')),
    CONSTRAINT chk_status CHECK (status IN ('active', 'inactive', 'deceased')),
    CONSTRAINT chk_dob CHECK (dob <= CURRENT_DATE)
);

CREATE UNIQUE INDEX idx_patients_mrn ON patients(mrn);
CREATE INDEX idx_patients_nik ON patients(nik) WHERE nik IS NOT NULL;
CREATE INDEX idx_patients_status ON patients(status);
CREATE INDEX idx_patients_created ON patients(created_at);

COMMENT ON TABLE patients IS 'Patient records with demographics';
COMMENT ON COLUMN patients.mrn IS 'Medical Record Number - unique identifier';
COMMENT ON COLUMN patients.nik IS 'National ID (NIK) - optional';
```

7. **.http file (from API_SPEC.md):**
```http
### Variables
@baseUrl = http://localhost:3000
@apiPath = /api/patients

### Health Check
GET {{baseUrl}}/health

### Create Patient
POST {{baseUrl}}{{apiPath}}
Content-Type: application/json

{
  "name": "John Doe",
  "mrn": "PAT-20260127-001",
  "dob": "1990-01-15",
  "gender": "male",
  "nik": "3174012345678901",
  "phone": "+628123456789",
  "email": "john@example.com",
  "address": {
    "street": "Jl. Sudirman 123",
    "city": "Jakarta",
    "province": "DKI Jakarta",
    "postal_code": "12190"
  }
}

### Get Patient by ID
GET {{baseUrl}}{{apiPath}}/PAT-20260127-001

### Update Patient
PUT {{baseUrl}}{{apiPath}}/PAT-20260127-001
Content-Type: application/json

{
  "name": "John Doe Updated",
  "phone": "+628123456789",
  "email": "john.doe@example.com"
}

### Delete Patient
DELETE {{baseUrl}}{{apiPath}}/PAT-20260127-001
```

8. **Test File:**
```go
// Generated from: docs/modules/patient/API_SPEC.md

package handler

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestPatientHandler_Create(t *testing.T) {
	// TODO: Setup mock repository
	// TODO: Create handler with mock
	// TODO: Test Create method with valid input
	// TODO: Test validation errors
	// TODO: Test duplicate MRN error
	t.Skip("Implement this test")
}

func TestPatientHandler_GetByID(t *testing.T) {
	// TODO: Test successful retrieval
	// TODO: Test patient not found
	t.Skip("Implement this test")
}

func TestPatientHandler_Update(t *testing.T) {
	// TODO: Test successful update
	// TODO: Test validation errors
	t.Skip("Implement this test")
}

func TestPatientHandler_Delete(t *testing.T) {
	// TODO: Test successful deletion (soft delete)
	// TODO: Test delete non-existent patient
	t.Skip("Implement this test")
}
```

9. **Update config/config.yaml:**
```yaml
configs:
  repository:
    patient-repository: postgres-patient-repository  # Added

service-definitions:
  # ... existing services ...

servers:
  api-server:
    published-services: [example-handler, patient-handler]  # Added patient-handler
```

**Output:**
- Complete working module
- All files generated from specs
- Tests scaffolded (ready to implement)
- HTTP file ready for manual testing
- Config updated

---

### Phase 2: Individual Component Generation

---

#### SKILL 7: Add Handler to Existing Module

**Purpose:** Add new handler/endpoint to existing module

**Input:**
```
User: "Add search patient endpoint to patient module"

Specification:
- Endpoint: GET /api/patients/search
- Query params: mrn, name, nik
- Response: Array of patients
```

**AI Actions:**
1. Update `docs/modules/patient/API_SPEC.md` (add endpoint spec)
2. Add method to `modules/patient/handler/patient_handler.go`
3. Add method to `modules/patient/repository/contract.go`
4. Implement in `modules/patient/repository/postgres/patient_repository.go`
5. Add test case to `patient_handler_test.go`
6. Add HTTP request to `patient.http`

**Generated Code:**

**API_SPEC.md (updated section):**
```markdown
### 5. Search Patients
**Endpoint:** `GET /api/patients/search`
**Authorization:** Required

**Query Parameters:**
- `mrn` (optional): Medical Record Number
- `name` (optional): Patient name (partial match)
- `nik` (optional): National ID

**Example:**
```
GET /api/patients/search?name=john
```

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid-123",
      "mrn": "PAT-001",
      "name": "John Doe",
      "phone": "+628123456789"
    }
  ],
  "count": 1
}
```
```

**Handler (added method):**
```go
// @Route "GET /search"
func (h *PatientHandler) Search(mrn, name, nik *string) ([]*domain.Patient, error) {
	return h.PatientRepo.Search(mrn, name, nik)
}
```

**Repository Interface (added method):**
```go
type PatientRepository interface {
	// ... existing methods ...
	Search(ctx context.Context, mrn, name, nik *string) ([]*domain.Patient, error)
}
```

**Repository Implementation:**
```go
func (r *PostgresPatientRepository) Search(ctx context.Context, mrn, name, nik *string) ([]*domain.Patient, error) {
	query := `
		SELECT id, mrn, name, dob, gender, phone, email, status, created_at, updated_at
		FROM patients
		WHERE deleted_at IS NULL
	`
	
	var args []interface{}
	argIndex := 1
	
	if mrn != nil {
		query += fmt.Sprintf(" AND mrn = $%d", argIndex)
		args = append(args, *mrn)
		argIndex++
	}
	
	if name != nil {
		query += fmt.Sprintf(" AND name ILIKE $%d", argIndex)
		args = append(args, "%"+*name+"%")
		argIndex++
	}
	
	if nik != nil {
		query += fmt.Sprintf(" AND nik = $%d", argIndex)
		args = append(args, *nik)
	}
	
	rows, err := r.dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var patients []*domain.Patient
	for rows.Next() {
		patient := &domain.Patient{}
		if err := rows.Scan(
			&patient.ID, &patient.MRN, &patient.Name, &patient.DOB,
			&patient.Gender, &patient.Phone, &patient.Email, &patient.Status,
			&patient.CreatedAt, &patient.UpdatedAt,
		); err != nil {
			return nil, err
		}
		patients = append(patients, patient)
	}
	
	return patients, nil
}
```

**HTTP file (added):**
```http
### Search Patient by Name
GET {{baseUrl}}{{apiPath}}/search?name=john

### Search Patient by MRN
GET {{baseUrl}}{{apiPath}}/search?mrn=PAT-001

### Search Patient by NIK
GET {{baseUrl}}{{apiPath}}/search?nik=3174012345678901
```

**Test (added):**
```go
func TestPatientHandler_Search(t *testing.T) {
	// TODO: Test search by MRN
	// TODO: Test search by name (partial match)
	// TODO: Test search by NIK
	// TODO: Test no results
	t.Skip("Implement this test")
}
```

---

#### SKILL 8: Add Repository Method

**Purpose:** Add new data access method to repository

**Input:**
```
User: "Add GetByEmail method to patient repository"
```

**AI Actions:**
1. Add method to `repository/contract.go`
2. Implement in `repository/postgres/patient_repository.go`
3. Add test case

**Generated Code:**

```go
// contract.go
type PatientRepository interface {
	// ... existing methods ...
	GetByEmail(ctx context.Context, email string) (*domain.Patient, error)
}

// postgres/patient_repository.go
func (r *PostgresPatientRepository) GetByEmail(ctx context.Context, email string) (*domain.Patient, error) {
	query := `
		SELECT id, mrn, name, dob, gender, nik, phone, email, status, created_at, updated_at
		FROM patients
		WHERE email = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	
	patient := &domain.Patient{}
	err := r.dbPool.QueryRow(ctx, query, email).Scan(
		&patient.ID, &patient.MRN, &patient.Name, &patient.DOB, &patient.Gender,
		&patient.NIK, &patient.Phone, &patient.Email, &patient.Status,
		&patient.CreatedAt, &patient.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return patient, err
}
```

---

#### SKILL 9: Add Entity/DTO

**Purpose:** Add new domain model or DTO

**Input:**
```
User: "Add EmergencyContact entity to patient module"

Specification:
- Belongs to Patient (1:N relationship)
- Fields: name, relationship, phone, is_primary
```

**AI Actions:**
1. Update `docs/modules/patient/SCHEMA.md`
2. Create `domain/emergency_contact.go`
3. Create migration file
4. Update repository interface
5. Update repository implementation

**Generated Files:**

**SCHEMA.md (added section):**
```markdown
### patient_emergency_contacts
Emergency contact information for patients

```sql
CREATE TABLE patient_emergency_contacts (
    id SERIAL PRIMARY KEY,
    patient_id VARCHAR(50) NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    relationship VARCHAR(50) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT chk_relationship CHECK (relationship IN ('spouse', 'parent', 'child', 'sibling', 'friend', 'other'))
);

CREATE INDEX idx_emergency_contacts_patient ON patient_emergency_contacts(patient_id);
```
```

**domain/emergency_contact.go:**
```go
package domain

import "time"

type EmergencyContact struct {
	ID           int       `json:"id" db:"id"`
	PatientID    string    `json:"patient_id" db:"patient_id"`
	Name         string    `json:"name" db:"name" validate:"required,min=3,max=100"`
	Relationship string    `json:"relationship" db:"relationship" validate:"required,oneof=spouse parent child sibling friend other"`
	Phone        string    `json:"phone" db:"phone" validate:"required,e164"`
	IsPrimary    bool      `json:"is_primary" db:"is_primary"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type CreateEmergencyContactRequest struct {
	Name         string `json:"name" validate:"required,min=3,max=100"`
	Relationship string `json:"relationship" validate:"required,oneof=spouse parent child sibling friend other"`
	Phone        string `json:"phone" validate:"required,e164"`
	IsPrimary    bool   `json:"is_primary"`
}
```

**migrations/patient/002_emergency_contacts.up.sql:**
```sql
-- Migration: Add emergency contacts table
-- Version: 1.0

CREATE TABLE IF NOT EXISTS patient_emergency_contacts (
    id SERIAL PRIMARY KEY,
    patient_id VARCHAR(50) NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    relationship VARCHAR(50) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT chk_relationship CHECK (relationship IN ('spouse', 'parent', 'child', 'sibling', 'friend', 'other'))
);

CREATE INDEX idx_emergency_contacts_patient ON patient_emergency_contacts(patient_id);

COMMENT ON TABLE patient_emergency_contacts IS 'Emergency contact information for patients';
```

---

#### SKILL 10: Update Configuration

**Purpose:** Update config.yaml with new services/settings

**Input:**
```
User: "Add Redis cache service to config"
```

**AI Actions:**
1. Read current `config/config.yaml`
2. Add service definition
3. Update dependencies if needed

**Generated Config Section:**
```yaml
service-definitions:
  # ... existing services ...
  
  redis_cache:
    type: redis
    config:
      addr: ${REDIS_ADDR:localhost:6379}
      password: ${REDIS_PASSWORD:}
      db: ${REDIS_DB:0}
      max_retries: 3
      pool_size: 10
```

---

### Phase 3: Testing & Validation

---

#### SKILL 11: Generate Unit Tests from Specification

**Purpose:** Generate comprehensive unit tests based on API spec

**Input:**
```
User: "Generate tests for patient handler"
```

**AI Actions:**
1. Read `docs/modules/patient/API_SPEC.md`
2. For each endpoint:
   - Generate success test case
   - Generate validation error test cases
   - Generate business rule violation test cases
3. Generate test helpers
4. Generate mock repository

**Generated Test File:**

```go
package handler

import (
	"context"
	"errors"
	"testing"
	"time"
	
	"myproject/modules/patient/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPatientRepository is a mock implementation of PatientRepository
type MockPatientRepository struct {
	mock.Mock
}

func (m *MockPatientRepository) Create(ctx context.Context, patient *domain.Patient) (*domain.Patient, error) {
	args := m.Called(ctx, patient)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Patient), args.Error(1)
}

func (m *MockPatientRepository) GetByID(ctx context.Context, id string) (*domain.Patient, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Patient), args.Error(1)
}

// ... other mock methods ...

// Test: Create Patient - Success
func TestPatientHandler_Create_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockPatientRepository)
	handler := &PatientHandler{
		PatientRepo: mockRepo,
	}
	
	req := &domain.CreatePatientRequest{
		Name:   "John Doe",
		MRN:    "PAT-001",
		DOB:    "1990-01-15",
		Gender: "male",
		Phone:  "+628123456789",
	}
	
	expectedPatient := &domain.Patient{
		ID:     "uuid-123",
		Name:   req.Name,
		MRN:    req.MRN,
		Gender: req.Gender,
		Phone:  req.Phone,
		Status: "active",
	}
	
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *domain.Patient) bool {
		return p.Name == req.Name && p.MRN == req.MRN
	})).Return(expectedPatient, nil)
	
	// Act
	result, err := handler.Create(req)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedPatient.ID, result.ID)
	assert.Equal(t, expectedPatient.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

// Test: Create Patient - Validation Error (Name too short)
func TestPatientHandler_Create_ValidationError_NameTooShort(t *testing.T) {
	// Arrange
	mockRepo := new(MockPatientRepository)
	handler := &PatientHandler{
		PatientRepo: mockRepo,
	}
	
	req := &domain.CreatePatientRequest{
		Name:   "Jo", // Too short (min 3)
		MRN:    "PAT-001",
		DOB:    "1990-01-15",
		Gender: "male",
		Phone:  "+628123456789",
	}
	
	// Act
	result, err := handler.Create(req)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "name")
	mockRepo.AssertNotCalled(t, "Create")
}

// Test: Create Patient - Duplicate MRN
func TestPatientHandler_Create_DuplicateMRN(t *testing.T) {
	// Arrange
	mockRepo := new(MockPatientRepository)
	handler := &PatientHandler{
		PatientRepo: mockRepo,
	}
	
	req := &domain.CreatePatientRequest{
		Name:   "John Doe",
		MRN:    "PAT-001",
		DOB:    "1990-01-15",
		Gender: "male",
		Phone:  "+628123456789",
	}
	
	mockRepo.On("Create", mock.Anything, mock.Anything).
		Return(nil, errors.New("duplicate key value violates unique constraint"))
	
	// Act
	result, err := handler.Create(req)
	
	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "duplicate")
	mockRepo.AssertExpectations(t)
}

// Test: GetByID - Success
func TestPatientHandler_GetByID_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockPatientRepository)
	handler := &PatientHandler{
		PatientRepo: mockRepo,
	}
	
	patientID := "uuid-123"
	expectedPatient := &domain.Patient{
		ID:     patientID,
		Name:   "John Doe",
		MRN:    "PAT-001",
		Status: "active",
	}
	
	mockRepo.On("GetByID", mock.Anything, patientID).Return(expectedPatient, nil)
	
	// Act
	result, err := handler.GetByID(patientID)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedPatient.ID, result.ID)
	mockRepo.AssertExpectations(t)
}

// Test: GetByID - Not Found
func TestPatientHandler_GetByID_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockPatientRepository)
	handler := &PatientHandler{
		PatientRepo: mockRepo,
	}
	
	patientID := "nonexistent-id"
	mockRepo.On("GetByID", mock.Anything, patientID).Return(nil, nil)
	
	// Act
	result, err := handler.GetByID(patientID)
	
	// Assert
	assert.NoError(t, err) // nil return, not error
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// Test: Update - Success
func TestPatientHandler_Update_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockPatientRepository)
	handler := &PatientHandler{
		PatientRepo: mockRepo,
	}
	
	patientID := "uuid-123"
	existingPatient := &domain.Patient{
		ID:     patientID,
		Name:   "John Doe",
		Phone:  "+628123456789",
		Status: "active",
	}
	
	req := &domain.UpdatePatientRequest{
		Name:  "John Doe Updated",
		Phone: "+628987654321",
	}
	
	mockRepo.On("GetByID", mock.Anything, patientID).Return(existingPatient, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Patient) bool {
		return p.ID == patientID && p.Name == req.Name && p.Phone == req.Phone
	})).Return(existingPatient, nil)
	
	// Act
	result, err := handler.Update(patientID, req)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}

// Test: Delete - Success (Soft Delete)
func TestPatientHandler_Delete_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockPatientRepository)
	handler := &PatientHandler{
		PatientRepo: mockRepo,
	}
	
	patientID := "uuid-123"
	mockRepo.On("Delete", mock.Anything, patientID).Return(nil)
	
	// Act
	err := handler.Delete(patientID)
	
	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Helper function for creating test patients
func createTestPatient(id, name, mrn string) *domain.Patient {
	return &domain.Patient{
		ID:        id,
		Name:      name,
		MRN:       mrn,
		Gender:    "male",
		Phone:     "+628123456789",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
```

**Repository Tests:**
```go
package postgres

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: These are integration tests that require a test database

func TestPostgresPatientRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	
	// TODO: Setup test database connection
	// TODO: Run migrations
	// TODO: Create repository instance
	// TODO: Test Create method
	// TODO: Cleanup
	
	t.Skip("Setup test database for integration tests")
}

func TestPostgresPatientRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	
	t.Skip("Setup test database for integration tests")
}

// Add more integration tests...
```

---

#### SKILL 12: Generate HTTP Test Files

**Purpose:** Generate .http file for manual API testing

**Input:**
```
User: "Generate .http file for patient module"
```

**AI Actions:**
1. Read `docs/modules/patient/API_SPEC.md`
2. For each endpoint, generate HTTP request
3. Include example payloads from spec
4. Include error case examples

**Generated File:**

```http
### ========================================
### Patient Module - API Tests
### ========================================
### Base URL: http://localhost:3000
### Module: patient
### Version: 1.0
### ========================================

@baseUrl = http://localhost:3000
@apiPath = /api/patients

### ========================================
### 1. HEALTH CHECK
### ========================================

### Health Check
GET {{baseUrl}}/health

### ========================================
### 2. CREATE PATIENT (POST)
### ========================================

### Create Patient - Success
POST {{baseUrl}}{{apiPath}}
Content-Type: application/json

{
  "name": "John Doe",
  "mrn": "PAT-20260127-001",
  "dob": "1990-01-15",
  "gender": "male",
  "nik": "3174012345678901",
  "phone": "+628123456789",
  "email": "john@example.com",
  "address": {
    "street": "Jl. Sudirman 123",
    "city": "Jakarta",
    "province": "DKI Jakarta",
    "postal_code": "12190"
  }
}

### Create Patient - Validation Error (Name too short)
POST {{baseUrl}}{{apiPath}}
Content-Type: application/json

{
  "name": "Jo",
  "mrn": "PAT-20260127-002",
  "dob": "1990-01-15",
  "gender": "male",
  "phone": "+628123456789"
}

### Create Patient - Invalid Email
POST {{baseUrl}}{{apiPath}}
Content-Type: application/json

{
  "name": "Jane Doe",
  "mrn": "PAT-20260127-003",
  "dob": "1995-05-20",
  "gender": "female",
  "phone": "+628123456789",
  "email": "invalid-email"
}

### Create Patient - Duplicate MRN
POST {{baseUrl}}{{apiPath}}
Content-Type: application/json

{
  "name": "Duplicate Patient",
  "mrn": "PAT-20260127-001",
  "dob": "1985-03-10",
  "gender": "male",
  "phone": "+628987654321"
}

### ========================================
### 3. GET PATIENT (READ)
### ========================================

### Get Patient by ID - Success
GET {{baseUrl}}{{apiPath}}/PAT-20260127-001

### Get Patient by ID - Not Found
GET {{baseUrl}}{{apiPath}}/nonexistent-id

### ========================================
### 4. UPDATE PATIENT (PUT)
### ========================================

### Update Patient - Success
PUT {{baseUrl}}{{apiPath}}/PAT-20260127-001
Content-Type: application/json

{
  "name": "John Doe Updated",
  "phone": "+628987654321",
  "email": "john.updated@example.com"
}

### Update Patient - Validation Error
PUT {{baseUrl}}{{apiPath}}/PAT-20260127-001
Content-Type: application/json

{
  "name": "Jo",
  "phone": "invalid-phone"
}

### ========================================
### 5. SEARCH PATIENTS
### ========================================

### Search by Name
GET {{baseUrl}}{{apiPath}}/search?name=john

### Search by MRN
GET {{baseUrl}}{{apiPath}}/search?mrn=PAT-20260127-001

### Search by NIK
GET {{baseUrl}}{{apiPath}}/search?nik=3174012345678901

### Search - No Results
GET {{baseUrl}}{{apiPath}}/search?name=nonexistent

### ========================================
### 6. DELETE PATIENT
### ========================================

### Delete Patient - Success (Soft Delete)
DELETE {{baseUrl}}{{apiPath}}/PAT-20260127-001

### Delete Patient - Not Found
DELETE {{baseUrl}}{{apiPath}}/nonexistent-id

### ========================================
### 7. LIST PATIENTS (with pagination)
### ========================================

### List All Patients
GET {{baseUrl}}{{apiPath}}

### List Patients - Page 1
GET {{baseUrl}}{{apiPath}}?page=1&limit=10

### List Patients - Filtered by Status
GET {{baseUrl}}{{apiPath}}?status=active

### ========================================
### 8. AUTHENTICATION TESTS (if required)
### ========================================

### Create Patient - No Auth Token (401)
POST {{baseUrl}}{{apiPath}}
Content-Type: application/json

{
  "name": "Unauthorized",
  "mrn": "PAT-20260127-999",
  "dob": "1990-01-01",
  "gender": "male",
  "phone": "+628123456789"
}

### ========================================
### NOTES:
### - Replace {{baseUrl}} with actual server URL
### - Add Authorization header if auth is required
### - Check response status codes
### - Validate response body structure
### ========================================
```

---

### Phase 4: Documentation Maintenance

---

#### SKILL 13: Sync Documentation with Code

**Purpose:** Update documentation based on code changes

**Input:**
```
User: "Sync patient module documentation"
```

**AI Actions:**
1. Scan `modules/patient/` for changes
2. Detect:
   - New @Route annotations
   - Changed validation tags
   - New struct fields
   - New repository methods
3. Update corresponding documentation
4. Create changelog entry

**Process:**

1. **Detect Code Changes:**
```go
// NEW: Added in patient_handler.go
// @Route "GET /history/{id}"
func (h *PatientHandler) GetHistory(id string) ([]*domain.Encounter, error) {
	// ...
}
```

2. **Update API_SPEC.md:**
```markdown
### 6. Get Patient History (NEW)
**Endpoint:** `GET /api/patients/history/{id}`
**Authorization:** Required

**Description:** Retrieve patient's encounter history

**Response:** `200 OK`
```json
{
  "patient_id": "PAT-001",
  "encounters": [
    {
      "id": "ENC-001",
      "date": "2026-01-15",
      "diagnosis": "Common cold"
    }
  ]
}
```
```

3. **Update CHANGELOG.md:**
```markdown
## [1.1.0] - 2026-01-27 - DRAFT

### Added
- Endpoint: `GET /api/patients/history/{id}` - Get patient encounter history

### Changed
- None

### Deprecated
- None
```

---

#### SKILL 14: Generate Documentation from Existing Code

**Purpose:** Create documentation for legacy code without docs

**Input:**
```
User: "Generate documentation for existing patient module"
```

**AI Actions:**
1. Scan all files in `modules/patient/`
2. Extract:
   - @Handler annotations → API endpoints
   - Struct tags → validation rules
   - Database queries → schema inference
3. Generate missing documentation

**Generated Documentation:**

**API_SPEC.md (reverse-engineered from code):**
```markdown
---
module: patient
version: 1.0
status: draft
note: Generated from existing code - Please review
---

# Patient Module - API Specification

**Note:** This documentation was auto-generated from code. Please review and update as needed.

## Detected Endpoints

### 1. Create Patient
**Endpoint:** `POST /api/patients`
**Source:** `patient_handler.go:15 - @Route "POST /"`

**Request Body:** (inferred from domain.CreatePatientRequest)
```json
{
  "name": "string (required, min=3, max=100)",
  "mrn": "string (required)",
  "dob": "string (required)",
  "gender": "string (required, one of: male, female, other)",
  "phone": "string (required, E.164 format)"
}
```

**TODO:**
- [ ] Add business rules description
- [ ] Add error response examples
- [ ] Add integration requirements
- [ ] Verify validation rules

[Continue for all detected endpoints...]
```

---

## Workflow Examples

### Example 1: Complete New Project from Scratch

```
Step 1: Generate BRD
User: "Buat BRD untuk sistem klinik dengan fitur patient, appointment, encounter"

AI: [SKILL 0] Generate docs/business/BRD.draft.md
User: Review BRD
User: "Approve BRD as v1.0"
AI: [SKILL 0.1] Create BRD.v1.0.md

---

Step 2: Generate Module Requirements
User: "Generate module requirements dari BRD"

AI: [SKILL 1] Generate:
    - docs/modules/patient/REQUIREMENTS.md
    - docs/modules/patient/API_SPEC.md
    - docs/modules/patient/SCHEMA.md
    - docs/modules/appointment/REQUIREMENTS.md
    - docs/modules/appointment/API_SPEC.md
    - docs/modules/appointment/SCHEMA.md
    - docs/modules/clinical/REQUIREMENTS.md
    - docs/modules/clinical/API_SPEC.md
    - docs/modules/clinical/SCHEMA.md

---

Step 3: Create Project
User: "Create Lokstra project: clinic-management"

AI: [SKILL 5] Generate complete project structure with example module

---

Step 4: Generate Modules from Specs
User: "Generate patient module from spec"
AI: [SKILL 6] Generate complete patient module (handler, repo, domain, tests, migrations)

User: "Generate appointment module from spec"
AI: [SKILL 6] Generate complete appointment module

User: "Generate clinical module from spec"
AI: [SKILL 6] Generate complete clinical module

---

Step 5: Validate
User: "Check consistency all modules"

AI: [SKILL 4] Run consistency check:
    ✅ Patient module: All OK
    ✅ Appointment module: All OK
    ❌ Clinical module: Missing diagnosis validation

---

Step 6: Run & Test
User: Run server: go run .
User: Test dengan .http files
User: Fix issues
User: Run unit tests: go test ./...

---

Result: Production-ready clinic management system in < 1 hour
```

---

### Example 2: Add Feature to Existing Project

```
Step 1: Update BRD
User: "Add e-prescription feature to BRD"

AI: [SKILL 0.2] Create BRD.draft.md (v1.1) with prescription feature
User: Review
User: "Approve as v1.1"
AI: [SKILL 0.1] Create BRD.v1.1.md

---

Step 2: Generate Module Design
User: "Generate prescription module requirements from BRD v1.1"

AI: [SKILL 1] Generate docs/modules/prescription/*

---

Step 3: Implement Module
User: "Generate prescription module from spec"

AI: [SKILL 6] Generate complete prescription module

---

Step 4: Integrate with Existing
User: "Update clinical handler to use prescription service"

AI: Add prescription repository injection to clinical handler
AI: Add createPrescription method

---

Step 5: Validate
User: "Check consistency prescription module"

AI: [SKILL 4] ✅ All OK

---

Result: New feature added with full documentation
```

---

### Example 3: Fix Inconsistency

```
Step 1: Detect Issue
User: "Check consistency patient module"

AI: [SKILL 4]
    ❌ Email validation: Spec says optional, code has required
    ❌ Missing endpoint: PUT /patients/{id}

---

Step 2: Fix Code
User: "Fix email validation based on spec"

AI: Update domain/patient.go:
    validate:"omitempty,email"  ← Fixed

---

Step 3: Add Missing Endpoint
User: "Add update patient endpoint from spec"

AI: [SKILL 7] Add PUT endpoint to handler + repository

---

Step 4: Re-validate
User: "Check consistency patient module"

AI: [SKILL 4] ✅ All OK

---

Result: Code now matches approved specification
```

---

## Best Practices

### For AI Agents

1. **Always Read Specs First**
   - Before generating any code, read approved documentation
   - Never guess - if spec unclear, ask user

2. **Generate Complete**
   - Don't generate partial code
   - Include: handler + repo + domain + tests + .http + migrations

3. **Validate Consistency**
   - After generation, auto-check against spec
   - Report any mismatches immediately

4. **Follow Lokstra Patterns**
   - Use @Handler and @Service annotations
   - Follow dependency injection patterns
   - Use proper validation tags

5. **Version Everything**
   - Track document versions
   - Update CHANGELOG.md
   - Never modify approved versions

6. **Test Coverage**
   - Generate unit tests for every handler method
   - Include success + error cases
   - Generate .http for manual testing

### For Developers

1. **Design First**
   - Write specs before coding
   - Get specs approved
   - Lock approved versions

2. **Use Consistency Checks**
   - Run regularly during development
   - Before PRs
   - Before deployment

3. **Maintain Documentation**
   - Update specs when requirements change
   - Sync code with specs
   - Keep CHANGELOG updated

4. **Follow Structure**
   - Use standard project structure
   - One module per bounded context
   - One file per entity/DTO

5. **Review AI Output**
   - Always review generated code
   - Check business logic
   - Verify security

---

## Lokstra-Specific Patterns

### 1. Handler Pattern

```go
// @Handler annotation registers handler
// @Handler name="<service-name>", prefix="<base-path>", middlewares=[...]
type MyHandler struct {
    // @Inject injects dependency
    // @Inject "<service-name>" - Direct service name
    // @Inject "@config.key" - Service name from config
    MyRepo repository.MyRepository
}

// @Route annotation defines endpoint
// @Route "HTTP_METHOD /path", ["middleware1", "middleware2"]
func (h *MyHandler) MyMethod(params *MyRequest) (*MyResponse, error) {
    // Implementation
}
```

### 2. Repository Pattern

```go
// @Service annotation registers repository implementation
// @Service "<service-name>"
type PostgresMyRepository struct {
    // @Inject "<dbpool-service-name>"
    dbPool serviceapi.DbPool
}

// Implement interface
var _ repository.MyRepository = (*PostgresMyRepository)(nil)
```

### 3. Configuration Pattern

```yaml
configs:
  repository:
    my-repository: postgres-my-repository  # Service name mapping

service-definitions:
  postgres-my-repository:
    type: postgres-my-repository
    depends-on: [db_main]
  
  db_main:
    type: dbpool_pg
    config:
      dsn: ${DB_DSN}

servers:
  api-server:
    addr: ":3000"
    published-services: [my-handler]
```

### 4. Validation Pattern

```go
// Use struct tags for validation
type CreateRequest struct {
    Name  string `json:"name" validate:"required,min=3,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,min=0,max=150"`
}
```

### 5. Error Handling Pattern

```go
func (h *MyHandler) Create(req *CreateRequest) (*MyEntity, error) {
    // Validate business rules
    if exists {
        return nil, fmt.Errorf("entity already exists: %s", req.Name)
    }
    
    // Call repository
    entity, err := h.MyRepo.Create(req)
    if err != nil {
        return nil, fmt.Errorf("failed to create entity: %w", err)
    }
    
    return entity, nil
}
```

---

## Summary

This SKILLS guide enables AI Agents to:

✅ **Design First** - Generate BRD, API specs, database schemas  
✅ **Implement Fast** - Generate complete modules from specifications  
✅ **Ensure Quality** - Auto-validate consistency between docs and code  
✅ **Maintain Easily** - Sync documentation with code changes  
✅ **Scale Confidently** - Follow Lokstra best practices  

**Result:** 10x productivity increase for Lokstra developers with AI assistance.

---

**Version History:**
- 2.0.0 (2026-01-27): Document-driven development workflow, 14 skills
- 1.0.0 (2025-11-12): Initial version, 9 basic skills

---

**For More Information:**
- Lokstra Documentation: https://primadi.github.io/lokstra/
- AI Agent Guide: https://primadi.github.io/lokstra/AI-AGENT-GUIDE
- Quick Reference: https://primadi.github.io/lokstra/QUICK-REFERENCE
