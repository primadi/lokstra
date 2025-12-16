-- Tenant DB: Create tenants table
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    tenant_code VARCHAR(50) NOT NULL UNIQUE,
    tenant_name VARCHAR(200) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
