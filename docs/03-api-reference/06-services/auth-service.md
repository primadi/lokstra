# Auth Service

The `auth_service` is the main authentication orchestrator that coordinates login, token refresh, and logout operations by integrating multiple authentication components (flows, token issuer, session store).

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [Registration](#registration)
- [Authentication Flows](#authentication-flows)
- [Token Management](#token-management)
- [Session Management](#session-management)
- [Usage](#usage)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

**Service Type:** `auth_service`

**Interface:** `auth.Service`

**Key Features:**

```
✓ Multi-Flow Authentication  - Support multiple auth flows (password, OTP, OAuth)
✓ Token Management          - Access and refresh tokens
✓ Session Storage           - Persistent session management
✓ Token Rotation            - Automatic refresh token rotation
✓ Modular Design            - Pluggable flows and components
✓ Multi-Tenant Support      - Tenant-aware authentication
```

## Architecture

### Component Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                        Auth Service                          │
│                                                              │
│  ┌────────────┐    ┌──────────────┐    ┌───────────────┐  │
│  │   Flows    │    │ Token Issuer │    │    Session    │  │
│  │            │    │              │    │               │  │
│  │ • Password │───▶│   JWT        │───▶│    Redis      │  │
│  │ • OTP      │    │              │    │               │  │
│  │ • OAuth    │    │              │    │               │  │
│  └────────────┘    └──────────────┘    └───────────────┘  │
│       │                                                      │
│       ▼                                                      │
│  ┌────────────────────────────────────────────────────┐    │
│  │              User Repository (PostgreSQL)          │    │
│  └────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────┘
```

### Authentication Flow

```
1. Login Request
   ↓
2. Auth Service receives request
   ↓
3. Select authentication flow (password/OTP/etc)
   ↓
4. Flow authenticates user
   ↓
5. Generate access & refresh tokens
   ↓
6. Store session in Redis
   ↓
7. Return tokens to client
```

### Dependencies

```go
auth_service
├── auth_token_jwt         // Token generation & verification
├── auth_session_redis     // Session storage
│   └── redis
└── auth_flow_*            // Authentication flows
    └── auth_user_repo_pg  // User database access
        └── dbpool_pg
```

## Configuration

### Config Struct

```go
type Config struct {
    TokenIssuerServiceName string            `json:"token_issuer_service_name"`
    SessionServiceName     string            `json:"session_service_name"`
    FlowServiceNames       map[string]string `json:"flow_service_names"`
    AccessTokenTTL         time.Duration     `json:"access_token_ttl"`
    RefreshTokenTTL        time.Duration     `json:"refresh_token_ttl"`
}
```

### YAML Configuration

**Basic Configuration:**

```yaml
services:
  # Dependencies
  main_db:
    type: dbpool_pg
    config:
      host: localhost
      database: myapp
      
  my_redis:
    type: redis
    config:
      addr: localhost:6379
  
  # Auth components
  my_token_issuer:
    type: auth_token_jwt
    config:
      secret_key: ${JWT_SECRET}
      issuer: myapp
      
  my_session:
    type: auth_session_redis
    config:
      addr: localhost:6379
      prefix: auth
      
  my_user_repo:
    type: auth_user_repo_pg
    config:
      dbpool_service_name: main_db
      table_name: users
      
  my_password_flow:
    type: auth_flow_password
    config:
      user_repo_service_name: my_user_repo
      
  # Main auth service
  my_auth:
    type: auth_service
    config:
      token_issuer_service_name: my_token_issuer
      session_service_name: my_session
      flow_service_names:
        password: my_password_flow
      access_token_ttl: 15m
      refresh_token_ttl: 168h  # 7 days
```

**Multi-Flow Configuration:**

```yaml
services:
  # ... other services ...
  
  my_password_flow:
    type: auth_flow_password
    config:
      user_repo_service_name: my_user_repo
      
  my_otp_flow:
    type: auth_flow_otp
    config:
      user_repo_service_name: my_user_repo
      kvstore_service_name: my_kvstore
      otp_length: 6
      otp_ttl: 5m
      
  my_auth:
    type: auth_service
    config:
      token_issuer_service_name: my_token_issuer
      session_service_name: my_session
      flow_service_names:
        password: my_password_flow
        otp: my_otp_flow
      access_token_ttl: 15m
      refresh_token_ttl: 168h
```

### Programmatic Configuration

```go
import (
    "time"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi/auth"
    "github.com/primadi/lokstra/services"
)

// Register all auth services
services.RegisterAuthServices()

// Create dependencies
lokstra_registry.NewService[any]("main_db", "dbpool_pg", dbConfig)
lokstra_registry.NewService[any]("my_token_issuer", "auth_token_jwt", tokenConfig)
lokstra_registry.NewService[any]("my_session", "auth_session_redis", sessionConfig)
lokstra_registry.NewService[any]("my_user_repo", "auth_user_repo_pg", userRepoConfig)
lokstra_registry.NewService[any]("my_password_flow", "auth_flow_password", flowConfig)

// Create auth service
authSvc := lokstra_registry.NewService[auth.Service](
    "my_auth", "auth_service",
    map[string]any{
        "token_issuer_service_name": "my_token_issuer",
        "session_service_name":      "my_session",
        "flow_service_names": map[string]string{
            "password": "my_password_flow",
        },
        "access_token_ttl":  15 * time.Minute,
        "refresh_token_ttl": 7 * 24 * time.Hour,
    },
)
```

## Registration

### Basic Registration

```go
import "github.com/primadi/lokstra/services/auth_service"

func init() {
    auth_service.Register()
}
```

### Bulk Registration

```go
import "github.com/primadi/lokstra/services"

func main() {
    // Registers all auth services
    services.RegisterAuthServices()
    
    // Or register all services
    services.RegisterAllServices()
}
```

## Authentication Flows

### Interface Definition

```go
type Service interface {
    Login(ctx context.Context, input LoginRequest) (*LoginResponse, error)
    RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
    Logout(ctx context.Context, refreshToken string) error
}

type LoginRequest struct {
    Flow    string         // "password", "otp", "oauth", etc.
    Payload map[string]any // Flow-specific credentials
}

type LoginResponse struct {
    AccessToken  string
    RefreshToken string
    ExpiresIn    int64  // seconds
}
```

### Login

**Password Authentication:**

```go
ctx := context.Background()

response, err := authSvc.Login(ctx, auth.LoginRequest{
    Flow: "password",
    Payload: map[string]any{
        "tenant_id": "tenant-123",
        "username":  "user@example.com",
        "password":  "password123",
    },
})

if err != nil {
    if errors.Is(err, auth.ErrInvalidCredentials) {
        return nil, ErrBadCredentials
    }
    return nil, err
}

// response.AccessToken  - JWT access token (15 min TTL)
// response.RefreshToken - JWT refresh token (7 days TTL)
// response.ExpiresIn    - 900 (seconds)
```

**OTP Authentication:**

```go
response, err := authSvc.Login(ctx, auth.LoginRequest{
    Flow: "otp",
    Payload: map[string]any{
        "tenant_id": "tenant-123",
        "username":  "user@example.com",
        "otp_code":  "123456",
    },
})
```

### Token Refresh

```go
// Client's refresh token is expiring soon or access token expired
response, err := authSvc.RefreshToken(ctx, oldRefreshToken)

if err != nil {
    if errors.Is(err, auth.ErrTokenExpired) {
        return nil, ErrReauthenticationRequired
    }
    if errors.Is(err, auth.ErrTokenNotFound) {
        return nil, ErrSessionNotFound
    }
    return nil, err
}

// New tokens returned
// response.AccessToken  - New access token
// response.RefreshToken - New refresh token (rotated)
// response.ExpiresIn    - New expiration time
```

**Token Rotation:**
- Old refresh token is invalidated
- New refresh token is generated
- Old session is deleted
- New session is created

### Logout

```go
err := authSvc.Logout(ctx, refreshToken)
if err != nil {
    log.Printf("Logout failed: %v", err)
    // Continue anyway - client should discard tokens
}

// Session deleted from Redis
// Refresh token invalidated
// Access token remains valid until expiration (stateless)
```

## Token Management

### Token Types

**Access Token:**
- Short-lived (15 minutes default)
- Used for API authorization
- Contains user identity and metadata
- Stateless (no server-side validation needed)

**Refresh Token:**
- Long-lived (7 days default)
- Used to obtain new access tokens
- Stored in session storage
- Can be revoked server-side

### Token Payload

```go
type TokenClaims struct {
    UserID    string         `json:"user_id"`
    TenantID  string         `json:"tenant_id"`
    Metadata  map[string]any `json:"metadata,omitempty"`
    TokenType string         `json:"token_type"`  // "access" or "refresh"
    IssuedAt  time.Time      `json:"issued_at"`
    ExpiresAt time.Time      `json:"expires_at"`
}
```

### TTL Configuration

```yaml
config:
  access_token_ttl: 15m    # Short: Limits exposure
  refresh_token_ttl: 168h  # Long: Better UX
```

**Recommendations:**
- **Access Token:** 5-30 minutes (balance security vs. UX)
- **Refresh Token:** 1-14 days (depends on security requirements)
- **Remember Me:** Longer refresh token (30-90 days)

## Session Management

### Session Storage

Sessions are stored in Redis with the refresh token as the key:

```
Key: auth:session:{refresh_token}
Value: {
  "user_id": "user-123",
  "tenant_id": "tenant-456",
  "metadata": {
    "username": "user@example.com",
    "email": "user@example.com",
    "role": "admin"
  }
}
TTL: 7 days
```

### Session Operations

**Set Session:**
```go
// Automatically done during Login
// Stored with refresh token TTL
```

**Get Session:**
```go
// Automatically retrieved during RefreshToken
// Returns session data for token generation
```

**Delete Session:**
```go
// Automatically done during Logout
// Invalidates refresh token
```

## Usage

### HTTP Handler Example

```go
package handler

import (
    "encoding/json"
    "net/http"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi/auth"
)

type AuthHandler struct {
    authSvc auth.Service
}

func NewAuthHandler() *AuthHandler {
    return &AuthHandler{
        authSvc: lokstra_registry.GetService[auth.Service]("my_auth"),
    }
}

// POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Flow     string         `json:"flow"`
        Payload  map[string]any `json:"payload"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    response, err := h.authSvc.Login(r.Context(), auth.LoginRequest{
        Flow:    req.Flow,
        Payload: req.Payload,
    })
    
    if err != nil {
        if errors.Is(err, auth.ErrInvalidCredentials) {
            http.Error(w, "Invalid credentials", http.StatusUnauthorized)
            return
        }
        http.Error(w, "Login failed", http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(response)
}

// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    response, err := h.authSvc.RefreshToken(r.Context(), req.RefreshToken)
    if err != nil {
        http.Error(w, "Token refresh failed", http.StatusUnauthorized)
        return
    }
    
    json.NewEncoder(w).Encode(response)
}

// POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    err := h.authSvc.Logout(r.Context(), req.RefreshToken)
    if err != nil {
        // Log error but don't fail
        log.Printf("Logout error: %v", err)
    }
    
    w.WriteHeader(http.StatusOK)
}
```

### Middleware Integration

Use with the JWT middleware for protected routes:

```go
import (
    "github.com/primadi/lokstra/middleware/jwtauth"
)

// Protected route
protectedRouter.Use(jwtauth.New(&jwtauth.Config{
    ValidatorServiceName: "my_validator",
    SkipPaths: []string{
        "/api/auth/login",
        "/api/auth/refresh",
    },
}))
```

## Best Practices

### Token Storage (Client-Side)

```javascript
✓ DO: Store tokens securely

// Access token in memory (most secure)
let accessToken = response.access_token;

// Refresh token in httpOnly cookie (secure)
document.cookie = `refresh_token=${response.refresh_token}; HttpOnly; Secure; SameSite=Strict`;

// Or in sessionStorage (less secure but works)
sessionStorage.setItem('access_token', response.access_token);
localStorage.setItem('refresh_token', response.refresh_token);

✗ DON'T: Store tokens in localStorage for sensitive apps
// BAD: Vulnerable to XSS
localStorage.setItem('access_token', token);
```

### Error Handling

```go
✓ DO: Handle specific auth errors

response, err := authSvc.Login(ctx, request)
if err != nil {
    switch {
    case errors.Is(err, auth.ErrInvalidCredentials):
        return nil, ErrBadCredentials  // 401
    case errors.Is(err, auth.ErrFlowNotFound):
        return nil, ErrInvalidFlow     // 400
    case errors.Is(err, auth.ErrUserNotActive):
        return nil, ErrAccountDisabled // 403
    default:
        return nil, ErrInternalServer  // 500
    }
}

✗ DON'T: Return generic errors
if err != nil {
    return nil, err  // BAD: Leaks internal details
}
```

### Token Refresh Strategy

```go
✓ DO: Implement proactive token refresh

// Client-side: Refresh before expiration
if (tokenExpiresIn < 60) {  // Less than 1 minute
    await refreshToken();
}

✓ DO: Handle 401 responses
if (response.status === 401) {
    // Try to refresh token
    const refreshed = await refreshToken();
    if (refreshed) {
        // Retry original request
        return retryRequest(originalRequest);
    } else {
        // Redirect to login
        redirectToLogin();
    }
}

✗ DON'T: Wait for access token to expire
// BAD: User experiences interruption
// Refresh only after 401 error
```

### Flow Selection

```go
✓ DO: Validate flow names

allowedFlows := map[string]bool{
    "password": true,
    "otp":      true,
}

if !allowedFlows[request.Flow] {
    return nil, ErrInvalidFlow
}

response, err := authSvc.Login(ctx, request)

✗ DON'T: Accept arbitrary flow names
// BAD: Could try non-existent flows
response, err := authSvc.Login(ctx, request)
```

### Multi-Tenant Security

```go
✓ DO: Always include tenant_id

payload := map[string]any{
    "tenant_id": tenantIDFromRequest,  // From subdomain/header
    "username":  username,
    "password":  password,
}

✓ DO: Validate tenant access
if user.TenantID != requestedTenantID {
    return nil, ErrUnauthorized
}

✗ DON'T: Allow cross-tenant access
// BAD: User could access other tenants
payload := map[string]any{
    "username": username,
    "password": password,
}
```

## Examples

### Complete Authentication System

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/primadi/lokstra/lokstra_registry"
    "github.com/primadi/lokstra/serviceapi/auth"
    "github.com/primadi/lokstra/services"
)

func setupAuth() {
    // Register all auth services
    services.RegisterAuthServices()
    
    // Configure database
    lokstra_registry.NewService[any]("main_db", "dbpool_pg", map[string]any{
        "host":     "localhost",
        "database": "myapp",
        "username": "postgres",
        "password": "password",
    })
    
    // Configure token issuer
    lokstra_registry.NewService[any]("token_issuer", "auth_token_jwt", map[string]any{
        "secret_key":   "your-secret-key",
        "issuer":       "myapp",
        "access_ttl":   15 * time.Minute,
        "refresh_ttl":  7 * 24 * time.Hour,
    })
    
    // Configure session storage
    lokstra_registry.NewService[any]("session", "auth_session_redis", map[string]any{
        "addr":   "localhost:6379",
        "prefix": "auth",
    })
    
    // Configure user repository
    lokstra_registry.NewService[any]("user_repo", "auth_user_repo_pg", map[string]any{
        "dbpool_service_name": "main_db",
        "schema":              "public",
        "table_name":          "users",
    })
    
    // Configure password flow
    lokstra_registry.NewService[any]("password_flow", "auth_flow_password", map[string]any{
        "user_repo_service_name": "user_repo",
    })
    
    // Configure main auth service
    lokstra_registry.NewService[auth.Service]("auth", "auth_service", map[string]any{
        "token_issuer_service_name": "token_issuer",
        "session_service_name":      "session",
        "flow_service_names": map[string]string{
            "password": "password_flow",
        },
        "access_token_ttl":  15 * time.Minute,
        "refresh_token_ttl": 7 * 24 * time.Hour,
    })
}

func main() {
    setupAuth()
    
    authSvc := lokstra_registry.GetService[auth.Service]("auth")
    
    // Login
    response, err := authSvc.Login(context.Background(), auth.LoginRequest{
        Flow: "password",
        Payload: map[string]any{
            "tenant_id": "tenant-123",
            "username":  "admin@example.com",
            "password":  "password123",
        },
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Access Token: %s", response.AccessToken)
    log.Printf("Refresh Token: %s", response.RefreshToken)
    log.Printf("Expires In: %d seconds", response.ExpiresIn)
}
```

### Multi-Flow Authentication

```go
func handleLogin(w http.ResponseWriter, r *http.Request) {
    authSvc := lokstra_registry.GetService[auth.Service]("auth")
    
    var req struct {
        Flow     string         `json:"flow"`
        Payload  map[string]any `json:"payload"`
    }
    
    json.NewDecoder(r.Body).Decode(&req)
    
    var response *auth.LoginResponse
    var err error
    
    switch req.Flow {
    case "password":
        response, err = authSvc.Login(r.Context(), auth.LoginRequest{
            Flow: "password",
            Payload: map[string]any{
                "tenant_id": extractTenantID(r),
                "username":  req.Payload["username"],
                "password":  req.Payload["password"],
            },
        })
        
    case "otp":
        response, err = authSvc.Login(r.Context(), auth.LoginRequest{
            Flow: "otp",
            Payload: map[string]any{
                "tenant_id": extractTenantID(r),
                "username":  req.Payload["username"],
                "otp_code":  req.Payload["otp_code"],
            },
        })
        
    default:
        http.Error(w, "Invalid auth flow", http.StatusBadRequest)
        return
    }
    
    if err != nil {
        http.Error(w, "Authentication failed", http.StatusUnauthorized)
        return
    }
    
    json.NewEncoder(w).Encode(response)
}
```

## Related Documentation

- [Services Overview](index) - Service architecture
- [Auth Validator](auth-validator) - Token validation for middleware
- [Auth Token JWT](auth-token-jwt) - JWT token generation
- [Auth Flow Password](auth-flow-password) - Password authentication
- [Auth Session Redis](auth-session-redis) - Session management
- [JWT Middleware](../05-middleware/jwt-auth) - Request authentication

---

**Next:** [Auth Validator Service](auth-validator) - Token validation
