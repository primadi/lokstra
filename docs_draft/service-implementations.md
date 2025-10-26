# Service Implementations Summary

This document provides an overview of all service implementations under `/services` that implement the contracts defined in `/serviceapi`.

## Service Implementation Pattern

All service implementations follow a standard pattern with three key functions:

1. **Service()** - Creates a service instance with the given configuration
2. **ServiceFactory()** - Factory function that creates a service from a map of parameters
3. **Register()** - Registers the ServiceFactory with the lokstra_registry

## Implemented Services

### 1. Redis Service (`/services/redis`)
**Service Type:** `redis`  
**Contract:** `serviceapi.Redis`  
**Description:** Redis client wrapper service

**Configuration:**
- `addr` - Redis server address (default: "localhost:6379")
- `password` - Redis password (default: "")
- `db` - Redis database number (default: 0)
- `pool_size` - Connection pool size (default: 10)

**Usage:**
```go
redis.Register()
```

---

### 2. KvStore Redis Service (`/services/kvstore_redis`)
**Service Type:** `kvstore_redis`  
**Contract:** `serviceapi.KvStore`  
**Description:** Key-value store implementation using Redis

**Configuration:**
- `addr` - Redis server address (default: "localhost:6379")
- `password` - Redis password (default: "")
- `db` - Redis database number (default: 0)
- `pool_size` - Connection pool size (default: 10)
- `prefix` - Key prefix for namespacing (default: "kv")

**Features:**
- Automatic key prefixing for namespace isolation
- JSON serialization for values
- TTL support for keys
- Pattern-based key search

**Usage:**
```go
kvstore_redis.Register()
```

---

### 3. Metrics Prometheus Service (`/services/metrics_prometheus`)
**Service Type:** `metrics_prometheus`  
**Contract:** `serviceapi.Metrics`  
**Description:** Metrics collection using Prometheus

**Configuration:**
- `namespace` - Namespace for all metrics (default: "app")
- `subsystem` - Subsystem for all metrics (default: "")

**Features:**
- Counter metrics
- Histogram metrics with default buckets
- Gauge metrics
- Thread-safe metric registration
- Label support for all metric types

**Usage:**
```go
metrics_prometheus.Register()
```

---

### 4. Auth Session Redis Service (`/services/auth_session_redis`)
**Service Type:** `auth_session_redis`  
**Contract:** `serviceapi/auth.Session`  
**Description:** Session management using Redis

**Configuration:**
- `addr` - Redis server address (default: "localhost:6379")
- `password` - Redis password (default: "")
- `db` - Redis database number (default: 0)
- `pool_size` - Connection pool size (default: 10)
- `prefix` - Key prefix for namespacing (default: "auth")

**Features:**
- Session storage with TTL
- User-based session tracking
- Bulk session deletion by user
- Multi-tenant support

**Usage:**
```go
auth_session_redis.Register()
```

---

### 5. Auth Token JWT Service (`/services/auth_token_jwt`)
**Service Type:** `auth_token_jwt`  
**Contract:** `serviceapi/auth.TokenIssuer`  
**Description:** JWT-based token issuer and verifier

**Configuration:**
- `secret_key` - Secret key for signing tokens (default: "change-me-in-production")
- `issuer` - Token issuer name (default: "lokstra")
- `access_ttl` - Default access token TTL (default: 15 minutes)
- `refresh_ttl` - Default refresh token TTL (default: 7 days)

**Features:**
- Access token generation
- Refresh token generation
- Token verification with type checking
- HMAC-SHA256 signing

**Usage:**
```go
auth_token_jwt.Register()
```

---

### 6. Auth User Repository PostgreSQL Service (`/services/auth_user_repo_pg`)
**Service Type:** `auth_user_repo_pg`  
**Contract:** `serviceapi/auth.UserRepository`  
**Description:** User management using PostgreSQL

**Configuration:**
- `dbpool_service_name` - Name of the DbPool service to use (default: "dbpool_pg")
- `schema` - Database schema (default: "public")
- `table_name` - Table name for users (default: "users")

**Features:**
- CRUD operations for users
- Multi-tenant support
- Password hash storage
- User metadata support
- Last login tracking

**Dependencies:**
- Requires a registered `DbPool` service

**Usage:**
```go
auth_user_repo_pg.Register()
```

---

### 7. Auth Flow Password Service (`/services/auth_flow_password`)
**Service Type:** `auth_flow_password`  
**Flow Name:** `password`  
**Contract:** `serviceapi/auth.Flow`  
**Description:** Username/password authentication flow

**Configuration:**
- `user_repo_service_name` - Name of the UserRepository service (default: "auth_user_repo_pg")

**Features:**
- Username/password authentication
- Bcrypt password verification
- Active user check
- Multi-tenant support

**Payload Format:**
```json
{
  "tenant_id": "tenant-123",
  "username": "user@example.com",
  "password": "password123"
}
```

**Dependencies:**
- Requires a registered `UserRepository` service

**Usage:**
```go
auth_flow_password.Register()
```

---

### 8. Auth Flow OTP Service (`/services/auth_flow_otp`)
**Service Type:** `auth_flow_otp`  
**Flow Name:** `otp`  
**Contract:** `serviceapi/auth.Flow`  
**Description:** One-Time Password authentication flow

**Configuration:**
- `user_repo_service_name` - Name of the UserRepository service (default: "auth_user_repo_pg")
- `kvstore_service_name` - Name of the KvStore service (default: "kvstore_redis")
- `otp_length` - Length of OTP code (default: 6)
- `otp_ttl_seconds` - OTP validity period in seconds (default: 300 = 5 minutes)
- `max_attempts` - Maximum OTP generation attempts (default: 5)

**Features:**
- OTP generation with configurable length
- OTP verification with TTL
- Failed attempt tracking
- Active user check
- Multi-tenant support

**OTP Generation:**
```go
flow := lokstra_registry.GetService("auth_flow_otp", flow)
otp, err := flow.GenerateOTP(ctx, "tenant-123", "user@example.com")
```

**Payload Format:**
```json
{
  "tenant_id": "tenant-123",
  "username": "user@example.com",
  "otp": "123456"
}
```

**Dependencies:**
- Requires a registered `UserRepository` service
- Requires a registered `KvStore` service

**Usage:**
```go
auth_flow_otp.Register()
```

---

### 9. Auth Service (`/services/auth_service`)
**Service Type:** `auth_service`  
**Contract:** `serviceapi/auth.Service`  
**Description:** Main authentication service orchestrating flows, tokens, and sessions

**Configuration:**
- `token_issuer_service_name` - Name of the TokenIssuer service (default: "auth_token_jwt")
- `session_service_name` - Name of the Session service (default: "auth_session_redis")
- `flow_service_names` - Map of flow names to service names (default: {"password": "auth_flow_password", "otp": "auth_flow_otp"})
- `access_token_ttl` - Access token TTL (default: 15 minutes)
- `refresh_token_ttl` - Refresh token TTL (default: 7 days)

**Features:**
- Login with multiple authentication flows
- Token refresh with rotation
- Logout
- Session management
- Multi-tenant support

**Dependencies:**
- Requires a registered `TokenIssuer` service
- Requires a registered `Session` service
- Requires one or more registered `Flow` services

**Usage:**
```go
auth_service.Register()
```

---

### 10. Auth Validator Service (`/services/auth_validator`)
**Service Type:** `auth_validator`  
**Contract:** `serviceapi/auth.Validator`  
**Description:** Token validation and user info extraction for middleware

**Configuration:**
- `token_issuer_service_name` - Name of the TokenIssuer service (default: "auth_token_jwt")
- `user_repo_service_name` - Name of the UserRepository service (optional)

**Features:**
- Access token validation
- Refresh token validation
- User info extraction from token claims
- Type checking for tokens

**Dependencies:**
- Requires a registered `TokenIssuer` service
- Optionally uses a `UserRepository` service

**Usage:**
```go
auth_validator.Register()
```

---

## Contract Improvements

The following improvements were made to the service contracts:

### 1. TokenIssuer Interface Enhancement
**File:** `/serviceapi/auth/token_issuer.go`

Added `VerifyToken` method and `TokenClaims` struct:
```go
type TokenIssuer interface {
    IssueAccessToken(ctx context.Context, auth *Result, ttl time.Duration) (string, error)
    IssueRefreshToken(ctx context.Context, auth *Result, ttl time.Duration) (string, error)
    VerifyToken(ctx context.Context, token string) (*TokenClaims, error)
}

type TokenClaims struct {
    UserID    string
    TenantID  string
    Metadata  map[string]any
    TokenType string
    IssuedAt  time.Time
    ExpiresAt time.Time
}
```

### 2. New Validator Contract
**File:** `/serviceapi/auth/validator.go`

Added new `Validator` interface for middleware and auth validation:
```go
type Validator interface {
    ValidateAccessToken(ctx context.Context, token string) (*TokenClaims, error)
    ValidateRefreshToken(ctx context.Context, token string) (*TokenClaims, error)
    GetUserInfo(ctx context.Context, claims *TokenClaims) (*UserInfo, error)
}

type UserInfo struct {
    UserID   string
    TenantID string
    Username string
    Email    string
    Metadata map[string]any
}
```

### 3. Helper Constants Update
**File:** `/serviceapi/auth/helper.go`

Added `VALIDATOR_TYPE` constant.

---

## Dependencies Added

The following Go dependencies were added:

1. **Prometheus Client** - For metrics collection
   ```
   github.com/prometheus/client_golang/prometheus
   ```

2. **JWT Library** - For token generation and verification
   ```
   github.com/golang-jwt/jwt/v5
   ```

---

## Usage Example

```go
package main

import (
    "github.com/primadi/lokstra/services/redis"
    "github.com/primadi/lokstra/services/kvstore_redis"
    "github.com/primadi/lokstra/services/metrics_prometheus"
    "github.com/primadi/lokstra/services/dbpool_pg"
    "github.com/primadi/lokstra/services/auth_session_redis"
    "github.com/primadi/lokstra/services/auth_token_jwt"
    "github.com/primadi/lokstra/services/auth_user_repo_pg"
    "github.com/primadi/lokstra/services/auth_flow_password"
    "github.com/primadi/lokstra/services/auth_flow_otp"
    "github.com/primadi/lokstra/services/auth_service"
    "github.com/primadi/lokstra/services/auth_validator"
)

func main() {
    // Register all service factories
    redis.Register()
    kvstore_redis.Register()
    metrics_prometheus.Register()
    dbpool_pg.Register()
    auth_session_redis.Register()
    auth_token_jwt.Register()
    auth_user_repo_pg.Register()
    auth_flow_password.Register()
    auth_flow_otp.Register()
    auth_service.Register()
    auth_validator.Register()
    
    // Services are now available for use via lokstra_registry
}
```

---

## Database Schema for Users Table

For the `auth_user_repo_pg` service, you'll need a users table with the following schema:

```sql
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    password_hash TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login TIMESTAMP,
    metadata JSONB,
    UNIQUE(tenant_id, username)
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
```

---

## Next Steps

1. **Testing** - Create unit tests for each service implementation
2. **Integration Tests** - Test the complete auth flow end-to-end
3. **Middleware** - Create HTTP middleware using the auth_validator service
4. **Documentation** - Add inline documentation and examples
5. **Additional Flows** - Implement more authentication flows (OAuth2, magic link, etc.)
6. **Rate Limiting** - Add rate limiting for authentication attempts
