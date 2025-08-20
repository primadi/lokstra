-- User Management Database Schema
-- This script creates the necessary tables for user management

-- Users table - stores user information
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(512) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE UNIQUE INDEX uniq_active_username 
ON users(username)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_tenant_email ON users(tenant_id, email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Create unique constraints for active users (not deleted)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_tenant_username_active 
ON users(tenant_id, username) 
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_tenant_email_active 
ON users(tenant_id, email) 
WHERE deleted_at IS NULL;

-- User sessions table (optional for session management)
CREATE TABLE IF NOT EXISTS user_sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(512) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true
);

-- Create indexes for sessions
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_is_active ON user_sessions(is_active);

-- User permissions table (optional for role-based access)
CREATE TABLE IF NOT EXISTS user_permissions (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission VARCHAR(255) NOT NULL,
    resource VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for permissions
CREATE INDEX IF NOT EXISTS idx_user_permissions_user_id ON user_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_permissions_permission ON user_permissions(permission);
CREATE INDEX IF NOT EXISTS idx_user_permissions_resource ON user_permissions(resource);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_permissions_unique 
ON user_permissions(user_id, permission, resource);

-- Audit log table for tracking user changes
CREATE TABLE IF NOT EXISTS user_audit_log (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) REFERENCES users(id) ON DELETE SET NULL,
    tenant_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL, -- CREATE, UPDATE, DELETE, LOGIN, LOGOUT
    old_values JSONB,
    new_values JSONB,
    changed_by VARCHAR(255), -- ID of user who made the change
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for audit log
CREATE INDEX IF NOT EXISTS idx_user_audit_log_user_id ON user_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_user_audit_log_tenant_id ON user_audit_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_audit_log_action ON user_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_user_audit_log_created_at ON user_audit_log(created_at);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to automatically update updated_at
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_sessions_updated_at 
    BEFORE UPDATE ON user_sessions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Sample tenant and default admin user (for testing)
-- Note: In production, these should be created through the application
INSERT INTO users (
    id, tenant_id, username, email, password_hash, is_active, 
    created_at, updated_at, metadata
) VALUES (
    'admin-001',
    'default',
    'admin',
    'admin@example.com',
    -- This is SHA256 hash of 'password123' (change in production!)
    'ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f',
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    '{"role": "admin", "created_by": "system"}'
)

-- Add initial permissions for admin user
INSERT INTO user_permissions (user_id, permission, resource) VALUES
    ('admin-001', 'user.create', '*'),
    ('admin-001', 'user.read', '*'),
    ('admin-001', 'user.update', '*'),
    ('admin-001', 'user.delete', '*'),
    ('admin-001', 'admin.*', '*')
ON CONFLICT (user_id, permission, resource) DO NOTHING;

-- Create view for active users (commonly used query)
CREATE OR REPLACE VIEW active_users AS
SELECT 
    id, tenant_id, username, email, is_active,
    created_at, updated_at, last_login, metadata
FROM users 
WHERE deleted_at IS NULL 
ORDER BY created_at DESC;

-- Create view for user stats
CREATE OR REPLACE VIEW user_stats AS
SELECT 
    tenant_id,
    COUNT(*) as total_users,
    COUNT(*) FILTER (WHERE is_active = true) as active_users,
    COUNT(*) FILTER (WHERE is_active = false) as inactive_users,
    COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE) as new_users_today,
    COUNT(*) FILTER (WHERE created_at >= date_trunc('week', CURRENT_DATE)) as new_users_this_week,
    COUNT(*) FILTER (WHERE created_at >= date_trunc('month', CURRENT_DATE)) as new_users_this_month
FROM users 
WHERE deleted_at IS NULL 
GROUP BY tenant_id;

COMMENT ON TABLE users IS 'Main user table storing user account information';
COMMENT ON TABLE user_sessions IS 'Active user sessions for authentication tracking';
COMMENT ON TABLE user_permissions IS 'User permissions for role-based access control';
COMMENT ON TABLE user_audit_log IS 'Audit trail for all user-related changes';
COMMENT ON VIEW active_users IS 'View of all active (non-deleted) users';
COMMENT ON VIEW user_stats IS 'View providing user statistics per tenant';
