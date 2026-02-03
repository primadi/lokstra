# Auth Module - Dependency Diagram

This document shows the dependency relationships for the auth module in a multi-tenant SaaS application.

## Module Dependency Graph

```mermaid
flowchart TB
    subgraph External["External Systems"]
        EMAIL[Email Service]
    end

    subgraph Core["Core Modules"]
        TENANT[Tenant Module]
        AUTH[Auth Module]
        AUDIT[Audit Module]
    end

    subgraph Application["Application Modules"]
        USER_PROFILE[User Profile Module]
        API_GATEWAY[API Gateway]
        OTHER[Other Modules...]
    end

    %% Auth dependencies
    AUTH -->|validates tenant| TENANT
    AUTH -.->|sends emails| EMAIL

    %% Modules depending on Auth
    API_GATEWAY -->|validates tokens| AUTH
    USER_PROFILE -->|requires auth| AUTH
    AUDIT -->|logs events| AUTH
    OTHER -->|requires auth| AUTH
```

## Tenant Context Flow

```mermaid
sequenceDiagram
    participant Client
    participant Gateway as API Gateway
    participant Auth as Auth Module
    participant DB as Database

    Client->>Gateway: Request with JWT
    Gateway->>Auth: Validate Token
    Auth->>Auth: Verify Signature
    Auth->>Auth: Check Expiration
    Auth->>Auth: Extract tenant_id
    Auth->>DB: Verify User Active
    Auth->>DB: Verify Tenant Active
    Auth-->>Gateway: User Context (user_id, tenant_id, permissions)
    Gateway->>Gateway: Attach Context to Request
    Gateway-->>Client: Response
```

## Authentication Flow

```mermaid
sequenceDiagram
    participant User
    participant Client
    participant Auth as Auth Module
    participant DB as Database

    User->>Client: Enter Credentials + Tenant
    Client->>Auth: POST /login
    Auth->>DB: Find User by Email in Tenant
    Auth->>Auth: Check Account Lock
    Auth->>Auth: Verify Password Hash
    
    alt Valid Credentials
        Auth->>DB: Reset Failed Attempts
        Auth->>Auth: Generate JWT (user_id, tenant_id, role)
        Auth->>Auth: Generate Refresh Token
        Auth->>DB: Create Session
        Auth-->>Client: {access_token, refresh_token, user}
    else Invalid Credentials
        Auth->>DB: Increment Failed Attempts
        alt 5+ Failed Attempts
            Auth->>DB: Lock Account (15 min)
        end
        Auth-->>Client: 401 Invalid Credentials
    end
```

## Multi-Tenant Data Flow

```mermaid
flowchart LR
    subgraph Tenant_A["Tenant A"]
        UA1[User A1]
        UA2[User A2]
    end

    subgraph Tenant_B["Tenant B"]
        UB1[User B1]
        UB2[User B2]
    end

    subgraph Auth_Layer["Auth Layer"]
        JWT[JWT with tenant_id]
        MIDDLEWARE[Auth Middleware]
    end

    subgraph Database["Database"]
        USERS[(Users Table)]
        note[("All queries filter by tenant_id")]
    end

    UA1 --> JWT
    UA2 --> JWT
    UB1 --> JWT
    UB2 --> JWT
    
    JWT --> MIDDLEWARE
    MIDDLEWARE --> USERS
```

## RBAC Permission Model

```mermaid
flowchart TB
    subgraph Roles["Predefined Roles"]
        SUPER[super_admin]
        TADMIN[tenant_admin]
        MGR[manager]
        MEMBER[member]
        VIEWER[viewer]
    end

    subgraph Permissions["Permissions"]
        ALL["*:*"]
        TENANT_ALL["tenant:*"]
        READ_ALL["*:read"]
        TEAM["team:*"]
        OWN["own:*"]
    end

    subgraph Scope["Scope"]
        GLOBAL[Global - All Tenants]
        SINGLE[Single Tenant]
    end

    SUPER --> ALL --> GLOBAL
    TADMIN --> TENANT_ALL --> SINGLE
    MGR --> READ_ALL --> SINGLE
    MGR --> TEAM --> SINGLE
    MEMBER --> READ_ALL --> SINGLE
    MEMBER --> OWN --> SINGLE
    VIEWER --> READ_ALL --> SINGLE
```

## Token Refresh Flow

```mermaid
sequenceDiagram
    participant Client
    participant Auth as Auth Module
    participant DB as Database

    Client->>Auth: POST /refresh {refresh_token}
    Auth->>DB: Find Refresh Token
    
    alt Token Valid & Not Used
        Auth->>DB: Mark Token as Used
        Auth->>DB: Get User
        Auth->>Auth: Check User & Tenant Active
        Auth->>Auth: Generate New Access Token
        Auth->>Auth: Generate New Refresh Token
        Auth->>DB: Store New Refresh Token
        Auth-->>Client: {new_access_token, new_refresh_token}
    else Token Invalid or Used
        Auth-->>Client: 401 Invalid Token
    end
```

## Password Reset Flow

```mermaid
sequenceDiagram
    participant User
    participant Client
    participant Auth as Auth Module
    participant Email as Email Service
    participant DB as Database

    User->>Client: Forgot Password
    Client->>Auth: POST /password/forgot {email, tenant_id}
    Auth->>DB: Find User by Email in Tenant
    
    alt User Exists
        Auth->>Auth: Generate Reset Token (1hr expiry)
        Auth->>DB: Store Reset Token
        Auth->>Email: Send Reset Email
    end
    
    Auth-->>Client: "If email exists, reset link sent"
    
    Note over User,Client: User clicks email link
    
    User->>Client: Enter New Password
    Client->>Auth: POST /password/reset {token, password}
    Auth->>DB: Validate Reset Token
    
    alt Token Valid
        Auth->>DB: Update Password Hash
        Auth->>DB: Mark Token as Used
        Auth->>DB: Revoke All Sessions
        Auth-->>Client: "Password reset successfully"
    else Token Invalid
        Auth-->>Client: 400 Invalid Token
    end
```

## Integration Points Summary

| Module | Direction | Purpose | Data Exchanged |
|--------|-----------|---------|----------------|
| Tenant | Auth → Tenant | Validate tenant status | tenant_id → tenant info |
| Notification | Auth → Notification | Send emails | email data → delivery status |
| API Gateway | Gateway → Auth | Validate requests | token → user context |
| All Modules | Module → Auth | Authorization | user context → access decision |
| Audit | Auth → Audit | Log events | auth events → audit log |
