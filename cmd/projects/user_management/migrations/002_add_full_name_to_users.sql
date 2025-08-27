-- Add full_name field to users table
-- Migration: 002_add_full_name_to_users

SET SEARCH_PATH TO user_management;

-- Add full_name column to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS full_name VARCHAR(255);

-- Create index for better search performance
CREATE INDEX IF NOT EXISTS idx_users_full_name ON users(full_name);

-- Update existing users with default full_name based on username
UPDATE users 
SET full_name = CASE 
    WHEN username = 'admin' THEN 'Administrator'
    ELSE INITCAP(REPLACE(username, '_', ' '))
END
WHERE full_name IS NULL;

-- Update the active_users view to include full_name
DROP VIEW active_users;
CREATE OR REPLACE VIEW active_users AS
SELECT 
    u.id, u.tenant_id, u.username, u.email, u.full_name, u.is_active,
    u.created_at, u.updated_at, u.last_login, u.metadata,
    t.name as tenant_name, t.display_name as tenant_display_name
FROM users u
LEFT JOIN tenants t ON u.tenant_id = t.id
WHERE u.deleted_at IS NULL 
ORDER BY u.created_at DESC;

-- Update default admin user with proper full_name
UPDATE users 
SET full_name = 'Default Administrator'
WHERE id = 'admin-001' AND full_name IS NULL;

-- Comment for documentation
COMMENT ON COLUMN users.full_name IS 'Full display name of the user';
