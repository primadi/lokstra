# Example 17: Complete Auth System

This example demonstrates a production-ready authentication and authorization system built with Lokstra, featuring:

- **JWT Authentication** with access and refresh tokens
- **Multiple Auth Flows**: Password, OTP (One-Time Password)
- **Role-Based Access Control (RBAC)** with middleware
- **Multi-Tenant Support**
- **Session Management** with Redis
- **Protected Routes** with different permission levels

## Architecture

### Services
- **Auth Service**: Handles login, token generation, and refresh
- **Auth Validator**: Validates JWT tokens
- **User Repository**: Manages user CRUD operations
- **Refresh Token Repository**: Stores and validates refresh tokens
- **KvStore**: Key-value storage backed by Redis

### Middleware
- **JwtAuth**: Validates JWT tokens and extracts user information
- **AccessControl**: Enforces role-based access control
- **CORS**: Cross-Origin Resource Sharing configuration
- **Request Logger**: Logs all HTTP requests
- **Recovery**: Handles panics gracefully

### API Routes

#### Public Routes (No Authentication)
- `GET /api/health` - Health check
- `GET /api/info` - System information

#### Auth Routes (No Authentication Required)
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login with username/password
- `POST /api/auth/refresh` - Refresh access token
- `POST /api/auth/otp/generate` - Generate OTP
- `POST /api/auth/otp/verify` - Verify OTP and login

#### Auth Routes (Authentication Required)
- `POST /api/auth/logout` - Logout user
- `GET /api/auth/me` - Get current user info

#### User Routes (Requires Authentication)
- `GET /api/user/profile` - Get user profile
- `PUT /api/user/profile` - Update user profile
- `POST /api/user/password/change` - Change password
- `GET /api/user/orders` - Get user orders
- `POST /api/user/orders` - Create new order

#### Admin Routes (Requires Admin Role)
- `GET /api/admin/users` - List all users
- `GET /api/admin/users/:id` - Get user by ID
- `POST /api/admin/users/:id/activate` - Activate user
- `POST /api/admin/users/:id/deactivate` - Deactivate user
- `DELETE /api/admin/users/:id` - Delete user
- `GET /api/admin/stats` - Get system statistics

## Prerequisites

### 1. PostgreSQL Database

Create a database and users table:

```sql
-- Create database
CREATE DATABASE lokstra_auth_demo;

-- Connect to the database
\c lokstra_auth_demo;

-- Create users table
CREATE TABLE users (
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
    UNIQUE(tenant_id, username)
);

-- Create index for faster lookups
CREATE INDEX idx_users_tenant_username ON users(tenant_id, username);
CREATE INDEX idx_users_email ON users(email);

-- Insert a test admin user (password: admin123)
INSERT INTO users (id, tenant_id, username, email, full_name, password_hash, is_active, metadata)
VALUES (
    'admin-001',
    'tenant1',
    'admin',
    'admin@example.com',
    'System Administrator',
    '$2a$10$X8z9J1YvN6qQ8Y1YvN6qQ8Y1YvN6qQ8Y1YvN6qQ8Y1YvN6qQ8Y1Yv',
    true,
    '{"role": "admin"}'
);

-- Insert a test regular user (password: user123)
INSERT INTO users (id, tenant_id, username, email, full_name, password_hash, is_active, metadata)
VALUES (
    'user-001',
    'tenant1',
    'john',
    'john@example.com',
    'John Doe',
    '$2a$10$Y9z8J2YvN7qQ9Y2YvN7qQ9Y2YvN7qQ9Y2YvN7qQ9Y2YvN7qQ9Y2Yv',
    true,
    '{"role": "user"}'
);
```

### 2. Redis

Make sure Redis is running:

```bash
# Using Docker
docker run -d -p 6379:6379 redis:alpine

# Or install locally
# Windows: Download from https://github.com/microsoftarchive/redis/releases
# Linux: sudo apt-get install redis-server
# macOS: brew install redis
```

### 3. Configuration

Update `config.yaml` with your database and Redis credentials.

**Important**: Change the `secret_key` in production!

## Running the Example

1. **Install dependencies**:
```bash
go mod tidy
```

2. **Set up the database** (see PostgreSQL section above)

3. **Start Redis** (see Redis section above)

4. **Run the application**:
```bash
cd cmd/examples/17-auth-system
go run .
```

The server will start on `http://localhost:8080`

## Testing the API

### 1. Register a New User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant1",
    "username": "alice",
    "email": "alice@example.com",
    "full_name": "Alice Smith",
    "password": "password123",
    "role": "user"
  }'
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant1",
    "username": "alice",
    "password": "password123"
  }'
```

Response:
```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "refresh_abc123...",
    "expires_in": 900,
    "token_type": "Bearer"
  }
}
```

### 3. Access Protected Route

```bash
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 4. Refresh Access Token

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### 5. Admin Operations (Requires Admin Role)

```bash
# List all users
curl -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"

# Get system stats
curl -X GET http://localhost:8080/api/admin/stats \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

## Password Hashing

To generate password hashes for testing, you can use this Go code:

```go
package main

import (
    "fmt"
    "github.com/primadi/lokstra/common/utils"
)

func main() {
    hash, _ := utils.HashPassword("your-password-here")
    fmt.Println(hash)
}
```

## Security Notes

1. **Secret Keys**: Change the `secret_key` in `config.yaml` for production use
2. **HTTPS**: Always use HTTPS in production
3. **Token Storage**: Store refresh tokens securely (HttpOnly cookies recommended)
4. **Password Policy**: Implement strong password requirements
5. **Rate Limiting**: Add rate limiting for auth endpoints
6. **Token Rotation**: Implement refresh token rotation for enhanced security

## Key Features Demonstrated

### 1. JWT Authentication Flow
- Token generation with custom claims
- Token validation and user extraction
- Refresh token mechanism
- Token expiration handling

### 2. Role-Based Access Control
- Middleware-based permission checking
- Role hierarchy support
- Custom role validation

### 3. Multi-Tenant Support
- Tenant isolation at the database level
- Tenant-aware authentication
- User scoping per tenant

### 4. Session Management
- Redis-backed session storage
- Session invalidation on logout
- Concurrent session handling

### 5. Error Handling
- Standardized API error responses
- Proper HTTP status codes
- User-friendly error messages

## Extending the Example

### Adding New Roles

1. Update the database metadata:
```sql
UPDATE users SET metadata = '{"role": "moderator"}' WHERE username = 'bob';
```

2. Add role check in middleware:
```go
moderatorOnly := accesscontrol.RequireRoles("admin", "moderator")
router.GET("/api/moderator/panel", moderatorPanelHandler, jwtMiddleware, moderatorOnly)
```

### Adding OAuth2 Support

To add OAuth2 (Google, GitHub, etc.):

1. Implement OAuth2 flow in auth service
2. Add OAuth2 configuration to `config.yaml`
3. Create OAuth2 callback handlers
4. Link OAuth2 accounts to local users

### Adding Email Verification

1. Generate verification tokens
2. Send verification emails
3. Add verification endpoint
4. Update user repository to track verification status

## Troubleshooting

### "connection refused" errors
- Check if PostgreSQL and Redis are running
- Verify connection settings in `config.yaml`

### "invalid token" errors
- Check if `secret_key` matches in auth service and validator
- Verify token hasn't expired
- Ensure Authorization header format: `Bearer <token>`

### "permission denied" errors
- Check user role in database metadata
- Verify middleware is applied correctly
- Check AccessControl configuration

## Related Examples

- Example 09: API Standard
- Example 11: Two-Layer Responses
- Example 13: Router Integration
- Example 14: Client Response Parsing

## License

MIT License - See LICENSE file in the root directory
