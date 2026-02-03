# Auth Module - Database Schema (Multi-Tenant)

**Purpose**: Complete PostgreSQL database schema for multi-tenant authentication module with RBAC, audit logging, and security best practices.

**Context**: This schema supports the Auth API documented in design-lokstra-api-specification. Demonstrates proper multi-tenant isolation at database level.

---

## Schema Overview

**Module**: auth  
**Database**: PostgreSQL 14+  
**Schema Name**: auth  
**Multi-Tenant**: Yes (tenant_id in all tables)  
**Character Set**: UTF-8  
**Collation**: en_US.UTF-8

**Tables**:
1. `tenants` - Tenant/clinic organizations
2. `users` - User accounts with multi-tenant isolation
3. `roles` - Role definitions per tenant
4. `permissions` - Permission definitions (global)
5. `role_permissions` - Role-permission mappings
6. `user_roles` - User-role assignments
7. `refresh_tokens` - JWT refresh token storage
8. `password_reset_tokens` - Password reset tokens
9. `login_attempts` - Failed login tracking
10. `audit_logs` - Authentication audit trail

**Total Size Estimate**: ~500 MB for 10K users (Year 1)

---

## 1. Tenants Table

**Purpose**: Store clinic/organization information (top-level tenant entity)

```sql
CREATE TABLE auth.tenants (
    -- Primary Key
    id TEXT PRIMARY KEY,  -- Format: tenant_uuid
    
    -- Business Fields
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,  -- URL-safe identifier
    domain VARCHAR(255),  -- Custom domain (optional)
    
    -- Contact Info
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    address TEXT,
    
    -- Subscription
    plan VARCHAR(50) NOT NULL DEFAULT 'free',  -- free, basic, premium, enterprise
    max_users INTEGER NOT NULL DEFAULT 5,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- active, suspended, inactive
    
    -- Settings
    settings JSONB DEFAULT '{}',  -- Tenant-specific configuration
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,  -- Soft delete
    
    -- Constraints
    CONSTRAINT chk_tenants_status CHECK (status IN ('active', 'suspended', 'inactive')),
    CONSTRAINT chk_tenants_plan CHECK (plan IN ('free', 'basic', 'premium', 'enterprise')),
    CONSTRAINT chk_tenants_max_users CHECK (max_users > 0),
    CONSTRAINT chk_tenants_slug_format CHECK (slug ~ '^[a-z0-9_-]+$')
);

-- Indexes
CREATE UNIQUE INDEX idx_tenants_slug ON auth.tenants(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_status ON auth.tenants(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_created_at ON auth.tenants(created_at DESC);
CREATE INDEX idx_tenants_domain ON auth.tenants(domain) WHERE domain IS NOT NULL AND deleted_at IS NULL;

-- GIN index for JSONB settings
CREATE INDEX idx_tenants_settings ON auth.tenants USING GIN(settings);

-- Comments
COMMENT ON TABLE auth.tenants IS 'Tenant organizations (clinics)';
COMMENT ON COLUMN auth.tenants.id IS 'Unique tenant identifier';
COMMENT ON COLUMN auth.tenants.slug IS 'URL-safe unique identifier for subdomain';
COMMENT ON COLUMN auth.tenants.settings IS 'JSON configuration: {logo, theme, features, etc}';
COMMENT ON COLUMN auth.tenants.deleted_at IS 'Soft delete timestamp (NULL = active)';
```

**Sample Data**:
```sql
INSERT INTO auth.tenants (id, name, slug, email, phone, plan, max_users, status)
VALUES 
  ('tenant_klinik_sehat_001', 'Klinik Sehat Sentosa', 'klinik-sehat', 'admin@kliniksehat.com', '+62-21-1234567', 'premium', 50, 'active'),
  ('tenant_rsia_bunda_002', 'RSIA Bunda Kasih', 'rsia-bunda', 'info@rsiabunda.com', '+62-21-7654321', 'enterprise', 200, 'active');
```

---

## 2. Users Table

**Purpose**: User accounts with multi-tenant isolation

```sql
CREATE TABLE auth.users (
    -- Primary Key
    id TEXT PRIMARY KEY,  -- Format: usr_ulid
    
    -- Multi-Tenant
    tenant_id TEXT NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    
    -- Authentication
    email VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,  -- Bcrypt hash
    email_verified BOOLEAN NOT NULL DEFAULT false,
    email_verified_at TIMESTAMPTZ,
    
    -- Profile
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    avatar_url TEXT,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- active, inactive, suspended, locked
    
    -- Security
    two_factor_enabled BOOLEAN NOT NULL DEFAULT false,
    two_factor_secret TEXT,  -- TOTP secret (encrypted)
    
    -- Metadata
    metadata JSONB DEFAULT '{}',  -- Additional user data (role-specific)
    preferences JSONB DEFAULT '{}',  -- User preferences
    
    -- Account Management
    password_changed_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    last_login_ip INET,
    login_count INTEGER NOT NULL DEFAULT 0,
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,  -- Soft delete
    
    -- Constraints
    CONSTRAINT chk_users_status CHECK (status IN ('active', 'inactive', 'suspended', 'locked')),
    CONSTRAINT chk_users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT chk_users_failed_attempts CHECK (failed_login_attempts >= 0),
    
    -- Composite unique constraint (tenant + email)
    CONSTRAINT uq_users_tenant_email UNIQUE(tenant_id, email)
);

-- Indexes
CREATE INDEX idx_users_tenant_id ON auth.users(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_tenant_email ON auth.users(tenant_id, email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email ON auth.users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON auth.users(tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON auth.users(tenant_id, created_at DESC);
CREATE INDEX idx_users_last_login ON auth.users(tenant_id, last_login_at DESC NULLS LAST);

-- GIN indexes for JSONB
CREATE INDEX idx_users_metadata ON auth.users USING GIN(metadata);
CREATE INDEX idx_users_preferences ON auth.users USING GIN(preferences);

-- Partial index for locked accounts
CREATE INDEX idx_users_locked ON auth.users(tenant_id, locked_until) 
    WHERE locked_until IS NOT NULL AND locked_until > NOW();

-- Comments
COMMENT ON TABLE auth.users IS 'User accounts with multi-tenant isolation';
COMMENT ON COLUMN auth.users.tenant_id IS 'Tenant identifier (composite FK with all relationships)';
COMMENT ON COLUMN auth.users.password_hash IS 'Bcrypt hash (cost factor 12)';
COMMENT ON COLUMN auth.users.metadata IS 'Role-specific data: {licenseNumber, specialization, etc}';
COMMENT ON COLUMN auth.users.failed_login_attempts IS 'Counter for account lockout (reset on successful login)';
COMMENT ON COLUMN auth.users.locked_until IS 'Account locked until this timestamp (NULL = not locked)';
```

**Sample Data**:
```sql
INSERT INTO auth.users (id, tenant_id, email, password_hash, full_name, phone, status, email_verified)
VALUES 
  ('usr_01HQSQXE9K8F2VJWX3QGH4YZ1A', 'tenant_klinik_sehat_001', 'dr.john@kliniksehat.com', 
   '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIeWCrm.Ga',  -- password: SecurePass123!
   'Dr. John Doe', '+62-812-3456-7890', 'active', true),
  ('usr_01HQSQXE9K8F2VJWX3QGH4YZ2B', 'tenant_klinik_sehat_001', 'nurse.jane@kliniksehat.com',
   '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIeWCrm.Ga',
   'Jane Smith', '+62-812-9876-5432', 'active', true);
```

---

## 3. Roles Table

**Purpose**: Role definitions per tenant (RBAC)

```sql
CREATE TABLE auth.roles (
    -- Primary Key
    id TEXT PRIMARY KEY,  -- Format: role_ulid
    
    -- Multi-Tenant
    tenant_id TEXT NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    
    -- Role Definition
    name VARCHAR(100) NOT NULL,  -- admin, doctor, nurse, receptionist
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Hierarchy
    level INTEGER NOT NULL DEFAULT 0,  -- Higher = more privileges
    
    -- Status
    is_system BOOLEAN NOT NULL DEFAULT false,  -- System roles (cannot be deleted)
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT chk_roles_level CHECK (level >= 0 AND level <= 100),
    
    -- Composite unique (tenant + name)
    CONSTRAINT uq_roles_tenant_name UNIQUE(tenant_id, name)
);

-- Indexes
CREATE INDEX idx_roles_tenant_id ON auth.roles(tenant_id);
CREATE INDEX idx_roles_tenant_name ON auth.roles(tenant_id, name);
CREATE INDEX idx_roles_level ON auth.roles(tenant_id, level DESC);
CREATE INDEX idx_roles_active ON auth.roles(tenant_id, is_active) WHERE is_active = true;

-- Comments
COMMENT ON TABLE auth.roles IS 'Role definitions per tenant (RBAC)';
COMMENT ON COLUMN auth.roles.level IS 'Hierarchy level (0-100): admin=100, doctor=80, nurse=60, receptionist=40';
COMMENT ON COLUMN auth.roles.is_system IS 'System-defined roles (cannot be deleted by users)';
```

**Sample Data**:
```sql
INSERT INTO auth.roles (id, tenant_id, name, display_name, description, level, is_system)
VALUES 
  ('role_01HQSQXE9K8F2VJWX3QGH5YZ1A', 'tenant_klinik_sehat_001', 'admin', 'Administrator', 'Full system access', 100, true),
  ('role_01HQSQXE9K8F2VJWX3QGH5YZ2B', 'tenant_klinik_sehat_001', 'doctor', 'Doctor', 'Medical practitioner', 80, true),
  ('role_01HQSQXE9K8F2VJWX3QGH5YZ3C', 'tenant_klinik_sehat_001', 'nurse', 'Nurse', 'Nursing staff', 60, true),
  ('role_01HQSQXE9K8F2VJWX3QGH5YZ4D', 'tenant_klinik_sehat_001', 'receptionist', 'Receptionist', 'Front desk', 40, true);
```

---

## 4. Permissions Table

**Purpose**: Permission definitions (global, not tenant-specific)

```sql
CREATE TABLE auth.permissions (
    -- Primary Key
    id TEXT PRIMARY KEY,  -- Format: perm_ulid
    
    -- Permission Definition
    code VARCHAR(100) NOT NULL UNIQUE,  -- patient:read, patient:write, etc.
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Grouping
    resource VARCHAR(50) NOT NULL,  -- patient, appointment, prescription, etc.
    action VARCHAR(50) NOT NULL,  -- read, write, delete, manage
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT chk_permissions_code_format CHECK (code ~ '^[a-z_]+:[a-z_]+$')
);

-- Indexes
CREATE UNIQUE INDEX idx_permissions_code ON auth.permissions(code);
CREATE INDEX idx_permissions_resource ON auth.permissions(resource);
CREATE INDEX idx_permissions_action ON auth.permissions(action);

-- Comments
COMMENT ON TABLE auth.permissions IS 'Global permission definitions (not tenant-specific)';
COMMENT ON COLUMN auth.permissions.code IS 'Permission code format: resource:action (e.g., patient:read)';
```

**Sample Data**:
```sql
INSERT INTO auth.permissions (id, code, display_name, description, resource, action)
VALUES 
  ('perm_01HQSQXE9K8F2VJWX3QGH6YZ1A', 'patient:read', 'View Patients', 'Can view patient records', 'patient', 'read'),
  ('perm_01HQSQXE9K8F2VJWX3QGH6YZ2B', 'patient:write', 'Create/Update Patients', 'Can create and update patient records', 'patient', 'write'),
  ('perm_01HQSQXE9K8F2VJWX3QGH6YZ3C', 'patient:delete', 'Delete Patients', 'Can delete patient records', 'patient', 'delete'),
  ('perm_01HQSQXE9K8F2VJWX3QGH6YZ4D', 'appointment:manage', 'Manage Appointments', 'Can create, update, cancel appointments', 'appointment', 'manage'),
  ('perm_01HQSQXE9K8F2VJWX3QGH6YZ5E', 'prescription:create', 'Create Prescriptions', 'Can create prescriptions', 'prescription', 'create'),
  ('perm_01HQSQXE9K8F2VJWX3QGH6YZ6F', 'report:export', 'Export Reports', 'Can export data reports', 'report', 'export'),
  ('perm_01HQSQXE9K8F2VJWX3QGH6YZ7G', 'user:manage', 'Manage Users', 'Can create, update, delete users', 'user', 'manage'),
  ('perm_01HQSQXE9K8F2VJWX3QGH6YZ8H', 'settings:manage', 'Manage Settings', 'Can modify system settings', 'settings', 'manage');
```

---

## 5. Role Permissions Table

**Purpose**: Many-to-many mapping between roles and permissions

```sql
CREATE TABLE auth.role_permissions (
    -- Primary Key
    id TEXT PRIMARY KEY,  -- Format: rp_ulid
    
    -- Foreign Keys (composite with tenant_id for data integrity)
    tenant_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    permission_id TEXT NOT NULL REFERENCES auth.permissions(id) ON DELETE CASCADE,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT fk_role_permissions_role 
        FOREIGN KEY (tenant_id, role_id) 
        REFERENCES auth.roles(tenant_id, id) 
        ON DELETE CASCADE,
    
    -- Unique constraint (one permission per role)
    CONSTRAINT uq_role_permissions UNIQUE(tenant_id, role_id, permission_id)
);

-- Indexes
CREATE INDEX idx_role_perms_tenant_id ON auth.role_permissions(tenant_id);
CREATE INDEX idx_role_perms_role_id ON auth.role_permissions(tenant_id, role_id);
CREATE INDEX idx_role_perms_permission_id ON auth.role_permissions(permission_id);

-- Comments
COMMENT ON TABLE auth.role_permissions IS 'Many-to-many mapping: roles <-> permissions';
COMMENT ON COLUMN auth.role_permissions.tenant_id IS 'Tenant ID (part of composite FK to roles table)';
```

**Sample Data**:
```sql
-- Admin role: all permissions
INSERT INTO auth.role_permissions (id, tenant_id, role_id, permission_id)
SELECT 
    'rp_' || gen_random_uuid()::text,
    'tenant_klinik_sehat_001',
    'role_01HQSQXE9K8F2VJWX3QGH5YZ1A',  -- admin role
    id
FROM auth.permissions;

-- Doctor role: patient read/write, appointment manage, prescription create
INSERT INTO auth.role_permissions (id, tenant_id, role_id, permission_id)
VALUES 
  ('rp_01HQSQXE9K8F2VJWX3QGH7YZ1A', 'tenant_klinik_sehat_001', 'role_01HQSQXE9K8F2VJWX3QGH5YZ2B', 'perm_01HQSQXE9K8F2VJWX3QGH6YZ1A'),
  ('rp_01HQSQXE9K8F2VJWX3QGH7YZ2B', 'tenant_klinik_sehat_001', 'role_01HQSQXE9K8F2VJWX3QGH5YZ2B', 'perm_01HQSQXE9K8F2VJWX3QGH6YZ2B'),
  ('rp_01HQSQXE9K8F2VJWX3QGH7YZ3C', 'tenant_klinik_sehat_001', 'role_01HQSQXE9K8F2VJWX3QGH5YZ2B', 'perm_01HQSQXE9K8F2VJWX3QGH6YZ4D'),
  ('rp_01HQSQXE9K8F2VJWX3QGH7YZ4D', 'tenant_klinik_sehat_001', 'role_01HQSQXE9K8F2VJWX3QGH5YZ2B', 'perm_01HQSQXE9K8F2VJWX3QGH6YZ5E');
```

---

## 6. User Roles Table

**Purpose**: Many-to-many mapping between users and roles

```sql
CREATE TABLE auth.user_roles (
    -- Primary Key
    id TEXT PRIMARY KEY,  -- Format: ur_ulid
    
    -- Foreign Keys (composite with tenant_id)
    tenant_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    
    -- Timestamps
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by TEXT,  -- User ID who assigned this role
    
    -- Constraints
    CONSTRAINT fk_user_roles_user 
        FOREIGN KEY (tenant_id, user_id) 
        REFERENCES auth.users(tenant_id, id) 
        ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role 
        FOREIGN KEY (tenant_id, role_id) 
        REFERENCES auth.roles(tenant_id, id) 
        ON DELETE CASCADE,
    
    -- Unique constraint (one role assignment per user)
    CONSTRAINT uq_user_roles UNIQUE(tenant_id, user_id, role_id)
);

-- Indexes
CREATE INDEX idx_user_roles_tenant_id ON auth.user_roles(tenant_id);
CREATE INDEX idx_user_roles_user_id ON auth.user_roles(tenant_id, user_id);
CREATE INDEX idx_user_roles_role_id ON auth.user_roles(tenant_id, role_id);

-- Comments
COMMENT ON TABLE auth.user_roles IS 'Many-to-many mapping: users <-> roles';
COMMENT ON COLUMN auth.user_roles.assigned_by IS 'User ID who assigned this role (audit trail)';
```

**Sample Data**:
```sql
INSERT INTO auth.user_roles (id, tenant_id, user_id, role_id, assigned_by)
VALUES 
  ('ur_01HQSQXE9K8F2VJWX3QGH8YZ1A', 'tenant_klinik_sehat_001', 'usr_01HQSQXE9K8F2VJWX3QGH4YZ1A', 'role_01HQSQXE9K8F2VJWX3QGH5YZ2B', NULL),  -- Dr. John = doctor
  ('ur_01HQSQXE9K8F2VJWX3QGH8YZ2B', 'tenant_klinik_sehat_001', 'usr_01HQSQXE9K8F2VJWX3QGH4YZ2B', 'role_01HQSQXE9K8F2VJWX3QGH5YZ3C', NULL);  -- Jane = nurse
```

---

## 7. Refresh Tokens Table

**Purpose**: Store JWT refresh tokens for session management

```sql
CREATE TABLE auth.refresh_tokens (
    -- Primary Key
    id TEXT PRIMARY KEY,  -- Format: rt_ulid
    
    -- Foreign Keys
    tenant_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    
    -- Token Data
    token_hash TEXT NOT NULL UNIQUE,  -- SHA-256 hash of token
    expires_at TIMESTAMPTZ NOT NULL,
    
    -- Device Info
    user_agent TEXT,
    ip_address INET,
    device_id TEXT,
    
    -- Status
    revoked BOOLEAN NOT NULL DEFAULT false,
    revoked_at TIMESTAMPTZ,
    revoked_reason TEXT,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT fk_refresh_tokens_user 
        FOREIGN KEY (tenant_id, user_id) 
        REFERENCES auth.users(tenant_id, id) 
        ON DELETE CASCADE
);

-- Indexes
CREATE UNIQUE INDEX idx_refresh_tokens_hash ON auth.refresh_tokens(token_hash) WHERE NOT revoked;
CREATE INDEX idx_refresh_tokens_user ON auth.refresh_tokens(tenant_id, user_id);
CREATE INDEX idx_refresh_tokens_expires ON auth.refresh_tokens(expires_at) WHERE NOT revoked;
CREATE INDEX idx_refresh_tokens_revoked ON auth.refresh_tokens(tenant_id, user_id, revoked);

-- Partial index for active tokens
CREATE INDEX idx_refresh_tokens_active ON auth.refresh_tokens(tenant_id, user_id) 
    WHERE NOT revoked AND expires_at > NOW();

-- Comments
COMMENT ON TABLE auth.refresh_tokens IS 'JWT refresh tokens for session management';
COMMENT ON COLUMN auth.refresh_tokens.token_hash IS 'SHA-256 hash of refresh token (never store plain tokens)';
COMMENT ON COLUMN auth.refresh_tokens.revoked IS 'True if token manually revoked (logout)';
```

---

## 8. Password Reset Tokens Table

**Purpose**: Store temporary password reset tokens

```sql
CREATE TABLE auth.password_reset_tokens (
    -- Primary Key
    id TEXT PRIMARY KEY,  -- Format: prt_ulid
    
    -- Foreign Keys
    tenant_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    
    -- Token Data
    token_hash TEXT NOT NULL UNIQUE,  -- SHA-256 hash
    expires_at TIMESTAMPTZ NOT NULL,
    
    -- Status
    used BOOLEAN NOT NULL DEFAULT false,
    used_at TIMESTAMPTZ,
    
    -- Metadata
    ip_address INET,
    user_agent TEXT,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT fk_password_reset_user 
        FOREIGN KEY (tenant_id, user_id) 
        REFERENCES auth.users(tenant_id, id) 
        ON DELETE CASCADE,
    CONSTRAINT chk_expires_future CHECK (expires_at > created_at)
);

-- Indexes
CREATE UNIQUE INDEX idx_password_reset_hash ON auth.password_reset_tokens(token_hash) WHERE NOT used;
CREATE INDEX idx_password_reset_user ON auth.password_reset_tokens(tenant_id, user_id);
CREATE INDEX idx_password_reset_expires ON auth.password_reset_tokens(expires_at) WHERE NOT used;

-- Comments
COMMENT ON TABLE auth.password_reset_tokens IS 'Temporary password reset tokens (15 min TTL)';
COMMENT ON COLUMN auth.password_reset_tokens.token_hash IS 'SHA-256 hash of reset token';
```

---

## 9. Login Attempts Table

**Purpose**: Track failed login attempts for security (account lockout)

```sql
CREATE TABLE auth.login_attempts (
    -- Primary Key
    id TEXT PRIMARY KEY,  -- Format: la_ulid
    
    -- Identification
    tenant_id TEXT NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,  -- Email attempted (user may not exist)
    user_id TEXT,  -- NULL if user doesn't exist
    
    -- Attempt Details
    success BOOLEAN NOT NULL,
    failure_reason TEXT,  -- invalid_password, account_locked, etc.
    
    -- Context
    ip_address INET NOT NULL,
    user_agent TEXT,
    
    -- Timestamps
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_login_attempts_tenant ON auth.login_attempts(tenant_id, email, attempted_at DESC);
CREATE INDEX idx_login_attempts_ip ON auth.login_attempts(ip_address, attempted_at DESC);
CREATE INDEX idx_login_attempts_user ON auth.login_attempts(tenant_id, user_id, attempted_at DESC) WHERE user_id IS NOT NULL;
CREATE INDEX idx_login_attempts_failed ON auth.login_attempts(tenant_id, email, attempted_at DESC) WHERE NOT success;

-- Partial index for recent failed attempts (last 24 hours)
CREATE INDEX idx_login_attempts_recent_failed ON auth.login_attempts(tenant_id, email) 
    WHERE NOT success AND attempted_at > NOW() - INTERVAL '24 hours';

-- Comments
COMMENT ON TABLE auth.login_attempts IS 'Login attempt tracking for security and account lockout';
COMMENT ON COLUMN auth.login_attempts.failure_reason IS 'invalid_password, account_locked, account_disabled, invalid_credentials';
```

---

## 10. Audit Logs Table

**Purpose**: Comprehensive audit trail for authentication events

```sql
CREATE TABLE auth.audit_logs (
    -- Primary Key
    id TEXT PRIMARY KEY,  -- Format: audit_ulid
    
    -- Multi-Tenant
    tenant_id TEXT NOT NULL REFERENCES auth.tenants(id) ON DELETE CASCADE,
    
    -- Actor
    user_id TEXT,  -- NULL for system actions
    
    -- Event
    event_type VARCHAR(50) NOT NULL,  -- login, logout, password_change, role_assign, etc.
    event_category VARCHAR(50) NOT NULL,  -- authentication, authorization, account_management
    resource_type VARCHAR(50),  -- user, role, permission
    resource_id TEXT,
    
    -- Details
    description TEXT NOT NULL,
    metadata JSONB,  -- Additional context
    
    -- Result
    success BOOLEAN NOT NULL,
    error_message TEXT,
    
    -- Context
    ip_address INET,
    user_agent TEXT,
    request_id TEXT,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_audit_tenant_user ON auth.audit_logs(tenant_id, user_id, created_at DESC);
CREATE INDEX idx_audit_event_type ON auth.audit_logs(tenant_id, event_type, created_at DESC);
CREATE INDEX idx_audit_created_at ON auth.audit_logs(created_at DESC);
CREATE INDEX idx_audit_resource ON auth.audit_logs(tenant_id, resource_type, resource_id) WHERE resource_id IS NOT NULL;

-- GIN index for JSONB metadata
CREATE INDEX idx_audit_metadata ON auth.audit_logs USING GIN(metadata);

-- Partial index for failed events
CREATE INDEX idx_audit_failed ON auth.audit_logs(tenant_id, created_at DESC) WHERE NOT success;

-- Comments
COMMENT ON TABLE auth.audit_logs IS 'Comprehensive audit trail for authentication/authorization events';
COMMENT ON COLUMN auth.audit_logs.event_type IS 'login, logout, password_change, role_assign, permission_grant, etc.';
COMMENT ON COLUMN auth.audit_logs.metadata IS 'Additional event context: {old_value, new_value, changes, etc}';
```

---

## Triggers & Functions

### 1. Update Timestamp Trigger

```sql
-- Function to auto-update updated_at
CREATE OR REPLACE FUNCTION auth.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to tables
CREATE TRIGGER trg_tenants_updated_at
    BEFORE UPDATE ON auth.tenants
    FOR EACH ROW
    EXECUTE FUNCTION auth.update_updated_at_column();

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON auth.users
    FOR EACH ROW
    EXECUTE FUNCTION auth.update_updated_at_column();

CREATE TRIGGER trg_roles_updated_at
    BEFORE UPDATE ON auth.roles
    FOR EACH ROW
    EXECUTE FUNCTION auth.update_updated_at_column();
```

### 2. Audit Log Trigger

```sql
-- Function to auto-create audit logs on user changes
CREATE OR REPLACE FUNCTION auth.audit_user_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        INSERT INTO auth.audit_logs (id, tenant_id, user_id, event_type, event_category, 
                                      resource_type, resource_id, description, success, metadata)
        VALUES (
            'audit_' || gen_random_uuid()::text,
            NEW.tenant_id,
            NEW.id,
            'user_updated',
            'account_management',
            'user',
            NEW.id,
            'User profile updated',
            true,
            jsonb_build_object(
                'changed_fields', jsonb_build_object(
                    'email', CASE WHEN OLD.email <> NEW.email THEN jsonb_build_object('old', OLD.email, 'new', NEW.email) ELSE NULL END,
                    'full_name', CASE WHEN OLD.full_name <> NEW.full_name THEN jsonb_build_object('old', OLD.full_name, 'new', NEW.full_name) ELSE NULL END,
                    'status', CASE WHEN OLD.status <> NEW.status THEN jsonb_build_object('old', OLD.status, 'new', NEW.status) ELSE NULL END
                )
            )
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_audit
    AFTER UPDATE ON auth.users
    FOR EACH ROW
    EXECUTE FUNCTION auth.audit_user_changes();
```

---

## Row-Level Security (RLS)

### Enable RLS on All Tables

```sql
-- Tenants
ALTER TABLE auth.tenants ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON auth.tenants
    USING (id = current_setting('app.current_tenant_id', true)::TEXT);

-- Users
ALTER TABLE auth.users ENABLE ROW LEVEL SECURITY;
CREATE POLICY user_tenant_isolation ON auth.users
    USING (tenant_id = current_setting('app.current_tenant_id', true)::TEXT);

-- Roles
ALTER TABLE auth.roles ENABLE ROW LEVEL SECURITY;
CREATE POLICY role_tenant_isolation ON auth.roles
    USING (tenant_id = current_setting('app.current_tenant_id', true)::TEXT);

-- Role Permissions
ALTER TABLE auth.role_permissions ENABLE ROW LEVEL SECURITY;
CREATE POLICY role_perm_tenant_isolation ON auth.role_permissions
    USING (tenant_id = current_setting('app.current_tenant_id', true)::TEXT);

-- User Roles
ALTER TABLE auth.user_roles ENABLE ROW LEVEL SECURITY;
CREATE POLICY user_role_tenant_isolation ON auth.user_roles
    USING (tenant_id = current_setting('app.current_tenant_id', true)::TEXT);

-- Refresh Tokens
ALTER TABLE auth.refresh_tokens ENABLE ROW LEVEL SECURITY;
CREATE POLICY refresh_token_tenant_isolation ON auth.refresh_tokens
    USING (tenant_id = current_setting('app.current_tenant_id', true)::TEXT);

-- Password Reset Tokens
ALTER TABLE auth.password_reset_tokens ENABLE ROW LEVEL SECURITY;
CREATE POLICY password_reset_tenant_isolation ON auth.password_reset_tokens
    USING (tenant_id = current_setting('app.current_tenant_id', true)::TEXT);

-- Login Attempts
ALTER TABLE auth.login_attempts ENABLE ROW LEVEL SECURITY;
CREATE POLICY login_attempt_tenant_isolation ON auth.login_attempts
    USING (tenant_id = current_setting('app.current_tenant_id', true)::TEXT);

-- Audit Logs
ALTER TABLE auth.audit_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_tenant_isolation ON auth.audit_logs
    USING (tenant_id = current_setting('app.current_tenant_id', true)::TEXT);
```

**Usage in Application**:
```go
// Set tenant context at connection/transaction level
_, err := tx.Exec("SET LOCAL app.current_tenant_id = $1", tenantID)
```

---

## Indexes Summary

**Total Indexes**: 57

**By Type**:
- Primary Key: 10
- Unique: 8
- Regular: 25
- Partial: 10
- Composite: 12
- GIN (JSONB): 4

**By Purpose**:
- Tenant isolation: 12
- Foreign key lookups: 15
- Sorting/filtering: 10
- Full-text search: 0 (can add if needed)
- Security (active tokens, locked accounts): 6
- Audit trail: 5

---

## Query Examples

### 1. User Login Query

```sql
SELECT u.id, u.tenant_id, u.email, u.password_hash, u.full_name, u.status, 
       u.email_verified, u.failed_login_attempts, u.locked_until
FROM auth.users u
WHERE u.tenant_id = $1 
  AND u.email = $2 
  AND u.deleted_at IS NULL;
```

### 2. Get User Permissions

```sql
SELECT DISTINCT p.code
FROM auth.users u
JOIN auth.user_roles ur ON ur.tenant_id = u.tenant_id AND ur.user_id = u.id
JOIN auth.roles r ON r.tenant_id = ur.tenant_id AND r.id = ur.role_id
JOIN auth.role_permissions rp ON rp.tenant_id = r.tenant_id AND rp.role_id = r.id
JOIN auth.permissions p ON p.id = rp.permission_id
WHERE u.tenant_id = $1 
  AND u.id = $2 
  AND u.deleted_at IS NULL
  AND r.is_active = true;
```

### 3. Check Recent Failed Login Attempts

```sql
SELECT COUNT(*)
FROM auth.login_attempts
WHERE tenant_id = $1 
  AND email = $2 
  AND NOT success 
  AND attempted_at > NOW() - INTERVAL '15 minutes';
```

### 4. List Active Users Per Tenant

```sql
SELECT u.id, u.email, u.full_name, u.status, u.last_login_at,
       array_agg(r.name) as roles
FROM auth.users u
LEFT JOIN auth.user_roles ur ON ur.tenant_id = u.tenant_id AND ur.user_id = u.id
LEFT JOIN auth.roles r ON r.tenant_id = ur.tenant_id AND r.id = ur.role_id
WHERE u.tenant_id = $1 
  AND u.deleted_at IS NULL 
  AND u.status = 'active'
GROUP BY u.id, u.email, u.full_name, u.status, u.last_login_at
ORDER BY u.created_at DESC;
```

---

## Data Volume Estimates

| Table | Initial | Year 1 | Year 5 | Notes |
|-------|---------|--------|--------|-------|
| tenants | 100 | 1,000 | 5,000 | Clinics/organizations |
| users | 1,000 | 10,000 | 50,000 | ~10 users per tenant avg |
| roles | 400 | 4,000 | 20,000 | ~4 roles per tenant |
| permissions | 50 | 100 | 150 | Global, grows slowly |
| role_permissions | 2,000 | 20,000 | 100,000 | ~5 perms per role avg |
| user_roles | 1,000 | 10,000 | 50,000 | 1 role per user avg |
| refresh_tokens | 2,000 | 20,000 | 100,000 | ~2 active tokens per user |
| password_reset_tokens | 100 | 1,000 | 5,000 | Short-lived, cleaned up |
| login_attempts | 10,000 | 500,000 | 5,000,000 | High volume, partitioned |
| audit_logs | 50,000 | 2,000,000 | 20,000,000 | High volume, partitioned |

**Total Size Estimate**: ~500 MB (Year 1), ~5 GB (Year 5)

---

## Performance Optimization

### 1. Partitioning for Large Tables

```sql
-- Partition login_attempts by month
CREATE TABLE auth.login_attempts (
    -- columns as defined above
) PARTITION BY RANGE (attempted_at);

CREATE TABLE auth.login_attempts_2026_01 PARTITION OF auth.login_attempts
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE auth.login_attempts_2026_02 PARTITION OF auth.login_attempts
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

-- Partition audit_logs by month
CREATE TABLE auth.audit_logs_partitioned (
    -- same structure
) PARTITION BY RANGE (created_at);
```

### 2. Cleanup Old Data

```sql
-- Delete expired tokens (run daily)
DELETE FROM auth.refresh_tokens 
WHERE expires_at < NOW() - INTERVAL '30 days';

DELETE FROM auth.password_reset_tokens 
WHERE expires_at < NOW() - INTERVAL '7 days';

-- Archive old login attempts (run monthly)
INSERT INTO auth.login_attempts_archive 
SELECT * FROM auth.login_attempts 
WHERE attempted_at < NOW() - INTERVAL '90 days';

DELETE FROM auth.login_attempts 
WHERE attempted_at < NOW() - INTERVAL '90 days';
```

---

## Security Best Practices

1. **Never store plain passwords** - Always use Bcrypt with cost 12
2. **Never store plain tokens** - Hash refresh tokens and reset tokens with SHA-256
3. **Enable RLS** - Defense in depth for multi-tenant isolation
4. **Use composite FKs** - Include tenant_id in all foreign key relationships
5. **Audit everything** - Log all authentication and authorization events
6. **Soft delete** - Use deleted_at instead of hard deletes for audit trail
7. **Lock accounts** - After 5 failed login attempts, lock for 15 minutes
8. **Expire tokens** - Access tokens 1 hour, refresh tokens 30 days
9. **Rotate secrets** - Change JWT signing keys periodically
10. **Monitor anomalies** - Track unusual login patterns, multiple IPs, etc.

---

**File Size**: 28 KB  
**Last Updated**: 2024-01-20  
**Related**: AUTH_API_EXAMPLE.md, MULTI_TENANT_SCHEMA_PATTERNS.md
