# Mermaid Diagram Templates for BRD

Reusable diagram templates to clarify BRD documents. Copy and customize for your project.

---

## 1. System Context Diagram (C4 Level 1)

Use to show system and external actors/integrations.

```mermaid
flowchart TB
    subgraph Users["👥 Users"]
        FO[Front Office]
        DOC[Doctor]
        NURSE[Nurse]
        PHARM[Pharmacist]
        ADMIN[Admin]
    end

    subgraph System["🏥 [System Name]"]
        APP[Main Application]
    end

    subgraph External["🔗 External Systems"]
        SATUSEHAT[SatuSehat API]
        PAYMENT[Payment Gateway]
        EMAIL[Email Service]
    end

    FO --> APP
    DOC --> APP
    NURSE --> APP
    PHARM --> APP
    ADMIN --> APP

    APP <--> SATUSEHAT
    APP <--> PAYMENT
    APP --> EMAIL
```

**How to Use:**
1. Replace `[System Name]` with your system name
2. Adjust Users with user roles from BRD
3. Adjust External Systems with integrations

---

## 2. High-Level User Flow

Use to show the main user journey through the system.

```mermaid
flowchart LR
    subgraph Registration["📋 Registration"]
        A1[Patient Arrives] --> A2[Search/Register Patient]
        A2 --> A3[Create Queue Number]
    end

    subgraph Consultation["🩺 Consultation"]
        B1[Call Patient] --> B2[Examination & SOAP]
        B2 --> B3[Create Prescription]
    end

    subgraph Pharmacy["💊 Pharmacy"]
        C1[Receive Prescription] --> C2[Verify & Dispensing]
        C2 --> C3[Ready for Pickup]
    end

    subgraph Payment["💰 Payment"]
        D1[Generate Invoice] --> D2[Receive Payment]
        D2 --> D3[Complete]
    end

    Registration --> Consultation --> Pharmacy --> Payment
```

**How to Use:**
1. Adjust stages with your business process
2. Add/remove steps according to scope

---

## 3. User Role & Permission Matrix

Use to visualize RBAC (Role-Based Access Control).

```mermaid
flowchart TB
    subgraph Roles["🔐 User Roles"]
        ADMIN[Admin<br/>Full Access]
        MANAGER[Manager<br/>Reports + Settings]
        STAFF[Staff<br/>Operations]
        VIEWER[Viewer<br/>Read Only]
    end

    subgraph Modules["📦 Modules"]
        M1[User Management]
        M2[Patient Records]
        M3[Prescriptions]
        M4[Billing]
        M5[Reports]
    end

    ADMIN --> M1
    ADMIN --> M2
    ADMIN --> M3
    ADMIN --> M4
    ADMIN --> M5

    MANAGER --> M2
    MANAGER --> M4
    MANAGER --> M5

    STAFF --> M2
    STAFF --> M3
    STAFF --> M4

    VIEWER --> M5

    style ADMIN fill:#e74c3c,color:#fff
    style MANAGER fill:#f39c12,color:#fff
    style STAFF fill:#3498db,color:#fff
    style VIEWER fill:#95a5a6,color:#fff
```

---

## 4. State Diagram (Entity Lifecycle)

Use to show status transitions (e.g., Order, Appointment).

```mermaid
stateDiagram-v2
    [*] --> Draft: Create
    Draft --> Pending: Submit
    Pending --> Approved: Approve
    Pending --> Rejected: Reject
    Rejected --> Draft: Revise
    Approved --> InProgress: Start Work
    InProgress --> Completed: Finish
    InProgress --> OnHold: Pause
    OnHold --> InProgress: Resume
    Completed --> [*]
    
    note right of Pending
        Waiting for approval
        from authorized user
    end note
```

**Common Use Cases:**
- Order status (Draft → Pending → Processing → Shipped → Delivered)
- Appointment status (Scheduled → Confirmed → In Progress → Completed)
- Document status (Draft → Review → Approved → Published)

---

## 5. Entity Relationship Diagram (ERD Simplified)

Use for data model overview (not full schema).

```mermaid
erDiagram
    PATIENT ||--o{ VISIT : has
    PATIENT {
        uuid id PK
        string name
        string nik
        date dob
    }
    
    VISIT ||--o{ PRESCRIPTION : generates
    VISIT {
        uuid id PK
        uuid patient_id FK
        datetime visit_date
        string status
    }
    
    PRESCRIPTION ||--o{ PRESCRIPTION_ITEM : contains
    PRESCRIPTION {
        uuid id PK
        uuid visit_id FK
        uuid doctor_id FK
        datetime created_at
    }
    
    PRESCRIPTION_ITEM {
        uuid id PK
        uuid prescription_id FK
        uuid drug_id FK
        int quantity
        string dosage
    }
    
    DRUG ||--o{ PRESCRIPTION_ITEM : "used in"
    DRUG {
        uuid id PK
        string name
        string unit
        decimal price
    }
```

---

## 6. Timeline / Gantt Chart

Use for project milestones and timeline.

```mermaid
gantt
    title Project Timeline
    dateFormat YYYY-MM-DD
    
    section Phase 1: Foundation
    Project Setup           :done, p1, 2026-02-01, 7d
    Database Design         :done, p2, after p1, 5d
    Core API Development    :active, p3, after p2, 14d
    
    section Phase 2: Features
    User Management         :p4, after p3, 7d
    Core Module             :p5, after p3, 10d
    Secondary Module        :p6, after p5, 10d
    
    section Phase 3: Integration
    SatuSehat Integration   :p7, after p6, 14d
    Payment Gateway         :p8, after p6, 7d
    
    section Phase 4: Launch
    Testing & QA            :p9, after p7, 14d
    UAT                     :p10, after p9, 7d
    Go Live                 :milestone, m1, after p10, 0d
```

---

## 7. Decision Tree (Business Logic)

Use for complex business rules.

```mermaid
flowchart TD
    START[Customer Arrives] --> Q1{New Customer?}
    
    Q1 -->|Yes| NEW[Register New]
    Q1 -->|No| SEARCH[Search Existing]
    
    NEW --> Q2{ID Valid?}
    Q2 -->|Yes| EXTERNAL[Fetch from SatuSehat]
    Q2 -->|No| MANUAL[Manual Input]
    
    EXTERNAL --> REGISTER[Create Visit Record]
    MANUAL --> REGISTER
    SEARCH --> REGISTER
    
    REGISTER --> Q3{Service Type?}
    Q3 -->|General| GENERAL[General Queue]
    Q3 -->|Specialist| SPECIALIST[Specialist Queue]
    Q3 -->|Emergency| EMERGENCY[Direct to Emergency]
    
    GENERAL --> WAITING[Wait for Call]
    SPECIALIST --> WAITING
    EMERGENCY --> TRIAGE[Direct Triage]
```

---

## 8. Sequence Diagram (API Flow)

Use to explain interactions between components.

```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant API as Backend API
    participant DB as Database
    participant EXT as External API

    U->>FE: Submit Form
    FE->>API: POST /api/records
    API->>API: Validate Request
    
    alt Validation Failed
        API-->>FE: 400 Bad Request
        FE-->>U: Show Error
    else Validation OK
        API->>DB: INSERT record
        DB-->>API: Created
        API->>EXT: POST to SatuSehat
        EXT-->>API: 201 Created
        API-->>FE: 201 Created + Data
        FE-->>U: Success Message
    end
```

---

## 9. Architecture Overview (Deployment)

Use for deployment architecture.

```mermaid
flowchart TB
    subgraph Client["🖥️ Client Layer"]
        BROWSER[Web Browser]
        MOBILE[Mobile App]
    end

    subgraph Gateway["🚪 API Gateway"]
        NGINX[Nginx / Traefik]
    end

    subgraph App["⚙️ Application Layer"]
        API1[API Server 1]
        API2[API Server 2]
    end

    subgraph Data["💾 Data Layer"]
        POSTGRES[(PostgreSQL)]
        REDIS[(Redis Cache)]
    end

    subgraph External["🔗 External"]
        SATUSEHAT[SatuSehat API]
        S3[Object Storage]
    end

    BROWSER --> NGINX
    MOBILE --> NGINX
    NGINX --> API1
    NGINX --> API2
    API1 --> POSTGRES
    API2 --> POSTGRES
    API1 --> REDIS
    API2 --> REDIS
    API1 <--> SATUSEHAT
    API1 --> S3
```

---

## Usage Tips

### Embedding in BRD

```markdown
## 3. System Overview

### 3.1 Context Diagram

\`\`\`mermaid
flowchart TB
    ... (copy diagram here)
\`\`\`

### 3.2 High-Level Flow

\`\`\`mermaid
flowchart LR
    ... (copy diagram here)
\`\`\`
```

### Best Practices

1. **Keep it simple** - BRD diagrams should be high-level, not detailed
2. **Label everything** - Use clear, business-friendly labels
3. **Color coding** - Use colors to distinguish roles/states
4. **One diagram per concept** - Don't cram everything into one diagram
5. **Test rendering** - Preview in VS Code or GitHub before publishing

### Recommended Diagrams per BRD Section

| BRD Section | Recommended Diagram |
|-------------|---------------------|
| Executive Summary | System Context (C4) |
| Business Process | High-Level User Flow |
| User Roles | Role & Permission Matrix |
| Scope | Timeline / Gantt |
| Data Model | ERD Simplified |
| Integration | Sequence Diagram |
| Technical Architecture | Architecture Overview |
