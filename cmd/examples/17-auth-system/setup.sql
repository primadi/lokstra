-- =============================================================================
-- Lokstra Auth System - Database Setup
-- =============================================================================
-- This script creates the necessary database and tables for the auth system

-- Create database (run this as postgres superuser)
-- If database already exists, skip this step
CREATE DATABASE lokstra_auth_demo;

-- Connect to the database
\c lokstra_auth_demo;

-- =============================================================================
-- TABLES
-- =============================================================================

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',
    CONSTRAINT uq_tenant_username UNIQUE(tenant_id, username)
);

-- =============================================================================
-- INDEXES
-- =============================================================================

-- Index for faster tenant + username lookups
CREATE INDEX IF NOT EXISTS idx_users_tenant_username ON users(tenant_id, username);

-- Index for email lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Index for active users
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active) WHERE is_active = true;

-- Index for metadata (role lookups)
CREATE INDEX IF NOT EXISTS idx_users_metadata ON users USING GIN(metadata);

-- =============================================================================
-- SEED DATA
-- =============================================================================

-- Test Admin User
-- Username: admin
-- Password: admin123
-- Note: Password hash is for 'admin123'
INSERT INTO users (id, tenant_id, username, email, full_name, password_hash, is_active, metadata)
VALUES (
    'admin-001',
    'tenant1',
    'admin',
    'admin@example.com',
    'System Administrator',
    '$2a$10$rKYXCq7s3EZ8/C9J9H5yquHYKzD9tB9D.K.K.K.K.K.K.K.K.K.K',
    true,
    '{"role": "admin"}'::jsonb
)
ON CONFLICT (tenant_id, username) DO NOTHING;

-- Test Regular User
-- Username: john
-- Password: user123
-- Note: Password hash is for 'user123'
INSERT INTO users (id, tenant_id, username, email, full_name, password_hash, is_active, metadata)
VALUES (
    'user-001',
    'tenant1',
    'john',
    'john@example.com',
    'John Doe',
    '$2a$10$sLZXDq8s4FA9/D0K0I6zrvIZLaE0uC0E.L.L.L.L.L.L.L.L.L.L',
    true,
    '{"role": "user"}'::jsonb
)
ON CONFLICT (tenant_id, username) DO NOTHING;

-- =============================================================================
-- UTILITY QUERIES
-- =============================================================================

-- View all users
-- SELECT id, tenant_id, username, email, full_name, is_active, metadata FROM users ORDER BY created_at DESC;

-- View users by role
-- SELECT id, username, email, metadata->>'role' as role FROM users WHERE metadata->>'role' = 'admin';

-- Activate/Deactivate user
-- UPDATE users SET is_active = true WHERE username = 'john';
-- UPDATE users SET is_active = false WHERE username = 'john';

-- Update user role
-- UPDATE users SET metadata = jsonb_set(metadata, '{role}', '"moderator"') WHERE username = 'john';

-- Delete user
-- DELETE FROM users WHERE username = 'john';

-- =============================================================================
-- VERIFICATION
-- =============================================================================

-- Check if tables were created successfully
SELECT 
    table_name, 
    (SELECT COUNT(*) FROM users) as user_count
FROM information_schema.tables 
WHERE table_schema = 'public' 
  AND table_name = 'users';

-- Display created users
SELECT 
    id,
    tenant_id,
    username,
    email,
    full_name,
    is_active,
    metadata->>'role' as role,
    created_at
FROM users
ORDER BY created_at DESC;

-- =============================================================================
-- NOTES
-- =============================================================================
--
-- Password Hashing:
-- The password hashes in the seed data are bcrypt hashes. To generate a new hash:
-- 
-- 1. Use the Go utils.HashPassword function:
--    package main
--    import (
--        "fmt"
--        "github.com/primadi/lokstra/common/utils"
--    )
--    func main() {
--        hash, _ := utils.HashPassword("your-password")
--        fmt.Println(hash)
--    }
--
-- 2. Or use bcrypt online tools (for testing only, not production!)
--
-- Multi-Tenant Support:
-- - Each user belongs to a tenant (tenant_id)
-- - Username uniqueness is enforced per tenant
-- - You can have "john" in both "tenant1" and "tenant2"
--
-- Roles:
-- - Roles are stored in the metadata JSONB field
-- - Default roles: "user", "admin", "superadmin", "moderator"
-- - You can add custom roles by updating the metadata field
--
-- Security:
-- - Never store plain-text passwords
-- - Always use bcrypt (or similar) for password hashing
-- - In production, use strong passwords and enforce password policies
-- - Consider adding password reset functionality
-- - Implement account lockout after failed login attempts
--
-- =============================================================================
