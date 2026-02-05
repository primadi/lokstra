# Module Dependency Diagram Templates

Visualize module dependencies and interactions for multi-tenant systems.

---

## 1. Module Dependency Graph

Shows which modules depend on which (for dependency injection planning).

```mermaid
flowchart TD
    subgraph Core["🔐 Core Modules"]
        TENANT[tenant<br/>Tenant Management]
        AUTH[auth<br/>Authentication]
    end

    subgraph Domain["📦 Domain Modules"]
        PATIENT[patient<br/>Patient Records]
        VISIT[visit<br/>Visit Management]
        PRESCRIPTION[prescription<br/>Prescriptions]
        LAB[lab<br/>Lab Results]
    end

    subgraph Support["🛠️ Support Modules"]
        NOTIF[notification<br/>Notifications]
        AUDIT[audit<br/>Audit Logs]
        REPORT[report<br/>Reports]
    end

    subgraph Integration["🔗 External Integration"]
        SATUSEHAT[satusehat<br/>Health Registry]
        PAYMENT[payment<br/>Payment Gateway]
    end

    %% Core dependencies
    AUTH --> TENANT
    
    %% Domain dependencies
    PATIENT --> AUTH
    PATIENT --> TENANT
    
    VISIT --> PATIENT
    VISIT --> AUTH
    
    PRESCRIPTION --> VISIT
    PRESCRIPTION --> PATIENT
    
    LAB --> VISIT
    LAB --> PATIENT
    
    %% Support dependencies
    NOTIF --> AUTH
    NOTIF --> TENANT
    
    AUDIT --> AUTH
    AUDIT --> TENANT
    
    REPORT --> PATIENT
    REPORT --> VISIT
    REPORT --> AUTH
    
    %% External integrations
    SATUSEHAT --> PATIENT
    SATUSEHAT --> AUTH
    
    PAYMENT --> AUTH
    PAYMENT --> TENANT

    style TENANT fill:#e74c3c,color:#fff
    style AUTH fill:#e74c3c,color:#fff
    style PATIENT fill:#3498db,color:#fff
    style VISIT fill:#3498db,color:#fff
```

**How to Use:**
1. Identify Core modules (no dependencies or minimal)
2. Add Domain modules (business logic)
3. Add Support modules (cross-cutting concerns)
4. Show dependency arrows (A → B means "A depends on B")
5. Verify no circular dependencies

---

## 2. Tenant Context Flow

Shows how tenant context propagates through the system.

```mermaid
sequenceDiagram
    participant C as Client
    participant GW as API Gateway
    participant AUTH as Auth Service
    participant DOM as Domain Service<br/>(Patient/Visit/etc)
    participant DB as Database

    C->>GW: Request + X-Tenant-ID header
    GW->>AUTH: Validate Token + Tenant
    AUTH->>AUTH: Extract tenant_id from JWT
    AUTH->>DB: Verify user active in tenant
    DB-->>AUTH: User context
    AUTH-->>GW: Valid + User context
    
    GW->>DOM: Forward request<br/>+ tenant_id in context
    DOM->>DOM: Apply tenant filter
    DOM->>DB: Query with tenant_id<br/>WHERE tenant_id = ?
    DB-->>DOM: Tenant-scoped data
    DOM-->>GW: Response
    GW-->>C: Response

    Note over DOM,DB: ALL queries must include<br/>tenant_id filter
```

---

## 3. Multi-Tenant Architecture Layers

Shows system layers with tenant isolation.

```mermaid
flowchart TB
    subgraph Clients["👥 Clients (Multi-Tenant)"]
        C1[Tenant A - Clinic 1]
        C2[Tenant B - Clinic 2]
        C3[Tenant C - Clinic 3]
    end

    subgraph Gateway["🚪 API Gateway"]
        GW[Tenant-Aware Gateway<br/>Validates tenant_id]
    end

    subgraph Application["⚙️ Application Layer"]
        AUTH[Auth Service<br/>Multi-Tenant]
        PATIENT[Patient Service<br/>Multi-Tenant]
        VISIT[Visit Service<br/>Multi-Tenant]
    end

    subgraph Data["💾 Data Layer - Tenant Isolation"]
        DB[(PostgreSQL<br/>Tenant Partitioning)]
        CACHE[(Redis<br/>Tenant-keyed cache)]
    end

    C1 --> GW
    C2 --> GW
    C3 --> GW
    
    GW --> AUTH
    GW --> PATIENT
    GW --> VISIT
    
    AUTH --> DB
    AUTH --> CACHE
    PATIENT --> DB
    PATIENT --> CACHE
    VISIT --> DB
    VISIT --> CACHE

    style GW fill:#f39c12,color:#fff
    style DB fill:#27ae60,color:#fff
    style CACHE fill:#27ae60,color:#fff
```

---

## 4. Module Interaction (Cross-Module Communication)

Shows how modules communicate while maintaining tenant context.

```mermaid
flowchart LR
    subgraph Visit["Visit Module"]
        V1[Create Visit]
    end

    subgraph Patient["Patient Module"]
        P1[Get Patient Info]
        P2[Update Last Visit]
    end

    subgraph Prescription["Prescription Module"]
        RX1[Create Prescription]
    end

    subgraph Notification["Notification Module"]
        N1[Send SMS]
    end

    subgraph Auth["Auth Module"]
        A1[Validate Permission]
    end

    V1 -->|1. Check permission| A1
    A1 -->|tenant_id + user_id| V1
    
    V1 -->|2. Get patient<br/>+ tenant_id| P1
    P1 -->|Patient data| V1
    
    V1 -->|3. Create visit<br/>+ tenant_id| V1
    
    V1 -->|4. Update patient<br/>+ tenant_id| P2
    
    V1 -->|5. Create prescription<br/>+ tenant_id| RX1
    
    V1 -->|6. Notify patient<br/>+ tenant_id| N1

    style A1 fill:#e74c3c,color:#fff
    style V1 fill:#3498db,color:#fff
```

**Key Principle:** Every cross-module call includes tenant_id

---

## 5. Bounded Context Map (DDD)

Shows module boundaries and relationships.

```mermaid
flowchart TB
    subgraph TENANT["🏢 Tenant Context"]
        T1[Tenant Management<br/>Organization, Plans, Settings]
    end

    subgraph AUTH["🔐 Auth Context"]
        A1[Authentication<br/>Users, Roles, Sessions]
    end

    subgraph PATIENT["👤 Patient Context"]
        P1[Patient Management<br/>Patients, Contacts, Medical History]
    end

    subgraph CLINICAL["🏥 Clinical Context"]
        CL1[Visits & Encounters<br/>Appointments, SOAP, Diagnosis]
        CL2[Prescriptions<br/>Medications, Dosage, Dispensing]
        CL3[Lab & Diagnostic<br/>Orders, Results, Imaging]
    end

    subgraph BILLING["💰 Billing Context"]
        B1[Invoices & Payments<br/>Charges, Receipts, Balances]
    end

    subgraph SUPPORT["🛠️ Support Context"]
        S1[Notifications<br/>Email, SMS, Push]
        S2[Audit & Logging<br/>Activity Tracking]
    end

    TENANT -.->|Conformist| AUTH
    AUTH -.->|Published Language| PATIENT
    AUTH -.->|Published Language| CLINICAL
    
    PATIENT -->|Customer-Supplier| CLINICAL
    CLINICAL -->|Customer-Supplier| BILLING
    
    CLINICAL -.->|Open Host Service| SUPPORT
    BILLING -.->|Open Host Service| SUPPORT

    style TENANT fill:#e74c3c,color:#fff
    style AUTH fill:#e74c3c,color:#fff
    style PATIENT fill:#3498db,color:#fff
    style CLINICAL fill:#3498db,color:#fff
    style BILLING fill:#9b59b6,color:#fff
    style SUPPORT fill:#95a5a6,color:#fff
```

**Relationship Types:**
- **Conformist (dotted)**: Downstream conforms to upstream
- **Customer-Supplier (solid)**: Upstream serves downstream
- **Open Host Service (dotted)**: Shared service for multiple contexts

---

## 6. Circular Dependency Detection

Example of BAD design (circular dependencies).

```mermaid
flowchart LR
    A[User Module] -->|depends on| B[Order Module]
    B -->|depends on| C[Product Module]
    C -->|depends on| A

    style A fill:#e74c3c,color:#fff
    style B fill:#e74c3c,color:#fff
    style C fill:#e74c3c,color:#fff

    X[❌ CIRCULAR DEPENDENCY<br/>DETECTED]
    style X fill:#c0392b,color:#fff
```

**Solution:** Break cycle with dependency inversion or event-driven approach.

```mermaid
flowchart LR
    A[User Module] -->|depends on| AUTH[Auth Module]
    B[Order Module] -->|depends on| AUTH
    C[Product Module] -->|depends on| AUTH

    B -.->|publishes events| E[Event Bus]
    E -.->|subscribes| A
    E -.->|subscribes| C

    style AUTH fill:#27ae60,color:#fff
    style E fill:#f39c12,color:#fff

    OK[✅ NO CIRCULAR DEPENDENCIES]
    style OK fill:#27ae60,color:#fff
```

---

## 7. Module Maturity Levels (Implementation Order)

Shows recommended implementation sequence.

```mermaid
flowchart TB
    subgraph L1["Level 1: Foundation<br/>Week 1-2"]
        T[tenant<br/>Tenant Management]
        A[auth<br/>Authentication]
    end

    subgraph L2["Level 2: Core Domain<br/>Week 3-4"]
        P[patient<br/>Patient Management]
        V[visit<br/>Visit/Encounter]
    end

    subgraph L3["Level 3: Extended Domain<br/>Week 5-6"]
        RX[prescription<br/>Prescriptions]
        LAB[lab<br/>Lab Results]
    end

    subgraph L4["Level 4: Support Services<br/>Week 7-8"]
        B[billing<br/>Billing]
        N[notification<br/>Notifications]
        AU[audit<br/>Audit Logs]
    end

    subgraph L5["Level 5: Integrations<br/>Week 9-10"]
        SS[satusehat<br/>Health Registry]
        PAY[payment<br/>Payment Gateway]
        RPT[report<br/>Reports]
    end

    L1 ==> L2
    L2 ==> L3
    L3 ==> L4
    L4 ==> L5

    style L1 fill:#e74c3c,color:#fff
    style L2 fill:#3498db,color:#fff
    style L3 fill:#9b59b6,color:#fff
    style L4 fill:#f39c12,color:#fff
    style L5 fill:#27ae60,color:#fff
```

---

## 8. Tenant Isolation Strategy

Shows data isolation at different layers.

```mermaid
flowchart TB
    subgraph Request["Request Flow"]
        R1[Client Request<br/>X-Tenant-ID: tenant-a]
    end

    subgraph Gateway["API Gateway"]
        G1{Validate<br/>Tenant ID}
    end

    subgraph Service["Service Layer"]
        S1[Apply Tenant Filter<br/>tenantID = 'tenant-a']
    end

    subgraph Database["Database Layer"]
        D1[Table: patients<br/>WHERE tenant_id = 'tenant-a']
        D2[Partition: tenant_a<br/>Physical Isolation]
        D3[Row Level Security<br/>Policy: tenant_id check]
    end

    R1 --> G1
    G1 -->|Valid| S1
    G1 -->|Invalid| REJECT[❌ 403 Forbidden]
    
    S1 --> D1
    S1 --> D2
    S1 --> D3

    style G1 fill:#f39c12,color:#fff
    style S1 fill:#3498db,color:#fff
    style D2 fill:#27ae60,color:#fff
```

**Isolation Strategies:**
1. **Application-Level**: Service layer applies tenant filter (most common)
2. **Partition-Level**: Physical table partitioning by tenant_id
3. **Row-Level Security**: PostgreSQL RLS policies (defense in depth)

---

## Usage Tips

### Embedding in Requirements

```markdown
## 9. Module Dependencies

### Dependency Graph

\`\`\`mermaid
flowchart TD
    ... (copy diagram here)
\`\`\`

### Tenant Context Flow

\`\`\`mermaid
sequenceDiagram
    ... (copy diagram here)
\`\`\`
```

### Best Practices

1. **Start simple** - Begin with Level 1 foundation modules
2. **Avoid cycles** - Use dependency detection diagram to validate
3. **Document context** - Show how tenant_id propagates
4. **Show layers** - Separate concerns (auth, domain, support)
5. **Implementation order** - Use maturity levels to plan sprints

### Recommended Diagrams per Phase

| Phase | Recommended Diagram |
|-------|---------------------|
| Module Planning | Module Dependency Graph, Bounded Context Map |
| Architecture Review | Multi-Tenant Architecture Layers |
| Implementation | Module Maturity Levels |
| Validation | Circular Dependency Detection |
| Documentation | Module Interaction, Tenant Context Flow |

---

## Multi-Tenant Checklist

Use this when designing module dependencies for multi-tenant systems:

- [ ] Every module has access to tenant_id context
- [ ] All database queries include tenant_id filter
- [ ] Cross-module calls include tenant_id parameter
- [ ] Auth module validates tenant_id in tokens
- [ ] No circular dependencies between modules
- [ ] Core modules (tenant, auth) have no domain dependencies
- [ ] Tenant isolation tested (cross-tenant access prevention)
