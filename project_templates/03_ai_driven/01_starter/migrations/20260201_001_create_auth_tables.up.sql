-- Auth Module Migration: Create users table
-- Multi-tenant authentication system

-- Enable UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create user status enum
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');

-- =====================================
-- Users Table
-- =====================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    status user_status NOT NULL DEFAULT 'active',
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    password_changed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Email unique within tenant (same email allowed across tenants)
    CONSTRAINT users_tenant_email_unique UNIQUE (tenant_id, email),
    -- Username globally unique
    CONSTRAINT users_username_unique UNIQUE (username)
);

-- Indexes for users table
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- =====================================
-- Sessions Table
-- =====================================
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    access_token_hash VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for sessions table
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_tenant_id ON sessions(tenant_id);
CREATE INDEX idx_sessions_access_token_hash ON sessions(access_token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- =====================================
-- Refresh Tokens Table
-- =====================================
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT refresh_tokens_token_hash_unique UNIQUE (token_hash)
);

-- Indexes for refresh_tokens table
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- =====================================
-- Roles Table
-- =====================================
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID, -- NULL for global/system roles
    name VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    permissions JSONB NOT NULL DEFAULT '[]',
    is_system_role BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Role name unique per tenant (or globally if tenant_id is NULL)
    CONSTRAINT roles_tenant_name_unique UNIQUE (tenant_id, name)
);

-- Index for roles table
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_roles_name ON roles(name);

-- =====================================
-- Password Reset Tokens Table
-- =====================================
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT password_reset_tokens_token_hash_unique UNIQUE (token_hash)
);

-- Indexes for password_reset_tokens table
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

-- =====================================
-- Insert System Roles
-- =====================================
INSERT INTO roles (name, description, permissions, is_system_role, tenant_id) VALUES
    ('super_admin', 'System administrator with full access to all tenants', '["*:*"]', TRUE, NULL),
    ('tenant_admin', 'Tenant administrator with full access within their tenant', '["tenant:*"]', TRUE, NULL),
    ('manager', 'Manager with read access and team management', '["*:read", "team:*"]', TRUE, NULL),
    ('member', 'Regular member with read access and own data management', '["*:read", "own:*"]', TRUE, NULL),
    ('viewer', 'Read-only access to all resources', '["*:read"]', TRUE, NULL);

-- =====================================
-- Updated At Trigger
-- =====================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================
-- Comments for Documentation
-- =====================================
COMMENT ON TABLE users IS 'Multi-tenant user accounts for authentication';
COMMENT ON COLUMN users.tenant_id IS 'Tenant/organization this user belongs to';
COMMENT ON COLUMN users.email IS 'Email address, unique within tenant';
COMMENT ON COLUMN users.username IS 'Username, globally unique across all tenants';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password (cost 12)';
COMMENT ON COLUMN users.role IS 'User role name (references roles table)';
COMMENT ON COLUMN users.failed_login_attempts IS 'Counter for failed login attempts (lockout after 5)';
COMMENT ON COLUMN users.locked_until IS 'Account locked until this timestamp (15 min lockout)';

COMMENT ON TABLE sessions IS 'Active user sessions for token management';
COMMENT ON TABLE refresh_tokens IS 'Refresh tokens for token rotation';
COMMENT ON TABLE roles IS 'Role definitions with permissions (RBAC)';
COMMENT ON COLUMN roles.tenant_id IS 'NULL for system roles, tenant UUID for custom roles';
COMMENT ON TABLE password_reset_tokens IS 'Password reset tokens (single-use, 1 hour expiry)';
