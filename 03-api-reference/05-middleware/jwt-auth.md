# JWT Auth Middleware

> JSON Web Token authentication

## Overview

JWT Auth middleware validates JSON Web Tokens and adds user information to the request context. It integrates with a validator service to verify tokens and extract user details.

## Import Path

```go
import "github.com/primadi/lokstra/middleware/jwtauth"
```

---

## Configuration

### Config Type

```go
type Config struct {
    ValidatorServiceName string   // Service name for token validation
    TokenHeader          string   // Header name for token (default: "Authorization")
    TokenPrefix          string   // Token prefix (default: "Bearer ")
    SkipPaths            []string // Paths to skip authentication
    ErrorMessage         string   // Custom error message
}
```

**Fields:**
- `ValidatorServiceName` - Name of validator service in registry (default: "auth_validator")
- `TokenHeader` - HTTP header containing token (default: "Authorization")
- `TokenPrefix` - Prefix to strip from token (default: "Bearer ")
- `SkipPaths` - List of paths to skip authentication
- `ErrorMessage` - Custom error message for invalid tokens

---

## Usage

### Basic Usage

```go
router.Use(jwtauth.Middleware(&jwtauth.Config{
    ValidatorServiceName: "auth_validator",
}))
```

---

### Skip Specific Paths

```go
router.Use(jwtauth.Middleware(&jwtauth.Config{
    ValidatorServiceName: "auth_validator",
    SkipPaths: []string{
        "/auth/login",
        "/auth/register",
        "/health",
        "/public/**",
    },
}))
```

---

### Custom Header

```go
router.Use(jwtauth.Middleware(&jwtauth.Config{
    ValidatorServiceName: "auth_validator",
    TokenHeader:          "X-Auth-Token",
    TokenPrefix:          "",
}))
```

---

### Custom Error Message

```go
router.Use(jwtauth.Middleware(&jwtauth.Config{
    ValidatorServiceName: "auth_validator",
    ErrorMessage:         "Invalid or expired session. Please login again",
}))
```

---

## YAML Configuration

```yaml
middlewares:
  - type: jwtauth
    params:
      validator_service_name: auth_validator
      token_header: Authorization
      token_prefix: "Bearer "
      skip_paths:
        - "/auth/login"
        - "/auth/register"
        - "/health"
        - "/public/**"
      error_message: "Authentication required"
```

---

## Validator Service

### Interface

JWT middleware requires a validator service implementing:

```go
type Validator interface {
    ValidateAccessToken(ctx context.Context, token string) (*TokenClaims, error)
    GetUserInfo(ctx context.Context, claims *TokenClaims) (*UserInfo, error)
}
```

---

### TokenClaims

```go
type TokenClaims struct {
    UserID    string                 // User ID
    TenantID  string                 // Tenant ID (for multi-tenancy)
    ExpiresAt int64                  // Expiration timestamp
    IssuedAt  int64                  // Issue timestamp
    Custom    map[string]interface{} // Custom claims
}
```

---

### UserInfo

```go
type UserInfo struct {
    UserID    string                 // User ID
    Username  string                 // Username
    Email     string                 // Email
    TenantID  string                 // Tenant ID
    Roles     []string               // User roles
    Metadata  map[string]interface{} // Additional metadata
}
```

---

## Context Keys

JWT middleware stores data in request context using these keys:

```go
const (
    UserInfoKey    ContextKey = "user_info"    // *auth.UserInfo
    TokenClaimsKey ContextKey = "token_claims" // *auth.TokenClaims
)
```

---

## Helper Functions

### GetUserInfo

Extracts user info from context.

```go
func GetUserInfo(ctx context.Context) (*auth.UserInfo, bool)
```

**Example:**
```go
userInfo, ok := jwtauth.GetUserInfo(c.R.Context())
if !ok {
    return c.Api.Unauthorized("User not authenticated")
}

log.Printf("User: %s", userInfo.Username)
```

---

### GetTokenClaims

Extracts token claims from context.

```go
func GetTokenClaims(ctx context.Context) (*auth.TokenClaims, bool)
```

**Example:**
```go
claims, ok := jwtauth.GetTokenClaims(c.R.Context())
if !ok {
    return c.Api.Unauthorized("No token claims")
}

log.Printf("Token expires at: %d", claims.ExpiresAt)
```

---

### GetCurrentUserID

Convenience function to get current user ID.

```go
func GetCurrentUserID(ctx context.Context) string
```

**Example:**
```go
userID := jwtauth.GetCurrentUserID(c.R.Context())
if userID == "" {
    return c.Api.Unauthorized("User not authenticated")
}
```

---

### GetCurrentTenantID

Convenience function to get current tenant ID.

```go
func GetCurrentTenantID(ctx context.Context) string
```

**Example:**
```go
tenantID := jwtauth.GetCurrentTenantID(c.R.Context())
// Use for multi-tenant queries
```

---

## Examples

### Protected Endpoint

```go
router.GET("/profile", func(c *request.Context) error {
    // Get user info from context
    userInfo, ok := jwtauth.GetUserInfo(c.R.Context())
    if !ok {
        return c.Api.Unauthorized("User not authenticated")
    }
    
    return c.Api.Ok(map[string]any{
        "user_id":  userInfo.UserID,
        "username": userInfo.Username,
        "email":    userInfo.Email,
    })
})
```

---

### Public and Protected Routes

```go
// Public routes (no auth)
router.POST("/auth/login", loginHandler)
router.POST("/auth/register", registerHandler)

// Protected routes
apiGroup := router.Group("/api")
apiGroup.Use(jwtauth.Middleware(&jwtauth.Config{
    ValidatorServiceName: "auth_validator",
}))

apiGroup.GET("/profile", getProfileHandler)
apiGroup.PUT("/profile", updateProfileHandler)
```

---

### Multi-Tenant Access

```go
router.GET("/data", func(c *request.Context) error {
    tenantID := jwtauth.GetCurrentTenantID(c.R.Context())
    if tenantID == "" {
        return c.Api.Unauthorized("Tenant not identified")
    }
    
    // Query data for specific tenant
    data := dataService.GetByTenant(tenantID)
    return c.Api.Ok(data)
})
```

---

### Custom Claims Access

```go
router.GET("/premium", func(c *request.Context) error {
    claims, ok := jwtauth.GetTokenClaims(c.R.Context())
    if !ok {
        return c.Api.Unauthorized("No token")
    }
    
    // Check custom claim
    isPremium, _ := claims.Custom["premium"].(bool)
    if !isPremium {
        return c.Api.Forbidden("Premium membership required")
    }
    
    return c.Api.Ok("Premium content")
})
```

---

### Audit Logging

```go
func auditMiddleware() request.HandlerFunc {
    return func(c *request.Context) error {
        userID := jwtauth.GetCurrentUserID(c.R.Context())
        
        // Log request with user context
        log.Printf("[AUDIT] User: %s, Method: %s, Path: %s",
            userID, c.R.Method, c.R.URL.Path)
        
        return c.Next()
    }
}

router.Use(
    jwtauth.Middleware(&jwtauth.Config{
        ValidatorServiceName: "auth_validator",
    }),
    auditMiddleware(), // After JWT auth
)
```

---

### Refresh Token Endpoint

```go
// Skip JWT auth for refresh endpoint
router.POST("/auth/refresh", func(c *request.Context) error {
    refreshToken := c.Body.GetString("refresh_token")
    
    // Validate refresh token and issue new access token
    newToken, err := authService.RefreshAccessToken(refreshToken)
    if err != nil {
        return c.Api.Unauthorized("Invalid refresh token")
    }
    
    return c.Api.Ok(map[string]any{
        "access_token": newToken,
    })
})
```

---

## Validator Service Implementation

### Example Validator

```go
package authservice

import (
    "context"
    "errors"
    "github.com/golang-jwt/jwt/v5"
    "github.com/primadi/lokstra/serviceapi/auth"
)

type JWTValidator struct {
    secretKey []byte
    userRepo  UserRepository
}

func (v *JWTValidator) ValidateAccessToken(ctx context.Context, tokenStr string) (*auth.TokenClaims, error) {
    // Parse JWT token
    token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("invalid signing method")
        }
        return v.secretKey, nil
    })
    
    if err != nil || !token.Valid {
        return nil, errors.New("invalid token")
    }
    
    // Extract claims
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("invalid claims")
    }
    
    return &auth.TokenClaims{
        UserID:    claims["user_id"].(string),
        TenantID:  claims["tenant_id"].(string),
        ExpiresAt: int64(claims["exp"].(float64)),
        IssuedAt:  int64(claims["iat"].(float64)),
        Custom:    claims,
    }, nil
}

func (v *JWTValidator) GetUserInfo(ctx context.Context, claims *auth.TokenClaims) (*auth.UserInfo, error) {
    // Fetch user from database
    user, err := v.userRepo.FindByID(ctx, claims.UserID)
    if err != nil {
        return nil, err
    }
    
    return &auth.UserInfo{
        UserID:   user.ID,
        Username: user.Username,
        Email:    user.Email,
        TenantID: user.TenantID,
        Roles:    user.Roles,
        Metadata: map[string]interface{}{
            "role":      user.Role,
            "is_active": user.IsActive,
        },
    }, nil
}
```

---

### Register Validator

```go
func init() {
    validator := &JWTValidator{
        secretKey: []byte(os.Getenv("JWT_SECRET")),
        userRepo:  NewUserRepository(),
    }
    
    lokstra_registry.RegisterService("auth_validator", validator)
}
```

---

## Error Responses

### Missing Token

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "error": "missing authentication token"
}
```

---

### Invalid Token

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "error": "invalid or expired token"
}
```

---

### User Not Found

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "error": "failed to get user info"
}
```

---

## Best Practices

### 1. Place After CORS, Before Handlers

```go
// ✅ Good order
router.Use(
    recovery.Middleware(&recovery.Config{}),
    cors.Middleware(allowedOrigins),          // First (preflight)
    jwtauth.Middleware(&jwtauth.Config{}),    // Then auth
)

// 🚫 Bad - blocks preflight requests
router.Use(
    recovery.Middleware(&recovery.Config{}),
    jwtauth.Middleware(&jwtauth.Config{}),    // Blocks OPTIONS
    cors.Middleware(allowedOrigins),
)
```

---

### 2. Skip Public Paths

```go
// ✅ Good - explicit skip list
jwtauth.Middleware(&jwtauth.Config{
    SkipPaths: []string{
        "/auth/login",
        "/auth/register",
        "/auth/forgot-password",
        "/health",
        "/public/**",
    },
})

// 🚫 Bad - protect public endpoints
jwtauth.Middleware(&jwtauth.Config{
    SkipPaths: []string{}, // Login fails!
})
```

---

### 3. Use Secure Secret Keys

```go
// ✅ Good - strong secret from environment
secretKey := []byte(os.Getenv("JWT_SECRET"))
if len(secretKey) < 32 {
    log.Fatal("JWT_SECRET must be at least 32 bytes")
}

// 🚫 Bad - hardcoded weak secret
secretKey := []byte("secret123")
```

---

### 4. Set Appropriate Token Expiry

```go
// ✅ Good - short-lived access tokens
accessTokenExpiry := 15 * time.Minute
refreshTokenExpiry := 7 * 24 * time.Hour

// 🚫 Bad - long-lived access tokens
accessTokenExpiry := 30 * 24 * time.Hour // Too long!
```

---

### 5. Validate Token Claims

```go
// ✅ Good - check expiry and other claims
func (v *JWTValidator) ValidateAccessToken(ctx context.Context, tokenStr string) (*auth.TokenClaims, error) {
    token, err := jwt.Parse(tokenStr, keyFunc)
    if err != nil {
        return nil, err
    }
    
    // Check expiration
    if claims.ExpiresAt < time.Now().Unix() {
        return nil, errors.New("token expired")
    }
    
    // Check issuer
    if claims.Issuer != "myapp" {
        return nil, errors.New("invalid issuer")
    }
    
    return claims, nil
}
```

---

## Testing

### Test with Valid Token

```go
func TestProtectedEndpoint(t *testing.T) {
    // Create mock validator
    mockValidator := &MockValidator{
        ValidateFunc: func(ctx context.Context, token string) (*auth.TokenClaims, error) {
            return &auth.TokenClaims{
                UserID:   "user123",
                TenantID: "tenant1",
            }, nil
        },
        GetUserInfoFunc: func(ctx context.Context, claims *auth.TokenClaims) (*auth.UserInfo, error) {
            return &auth.UserInfo{
                UserID:   "user123",
                Username: "testuser",
                Email:    "test@example.com",
            }, nil
        },
    }
    
    lokstra_registry.RegisterService("auth_validator", mockValidator)
    
    router := lokstra.NewRouter()
    router.Use(jwtauth.Middleware(&jwtauth.Config{
        ValidatorServiceName: "auth_validator",
    }))
    
    // Test with token
    req := httptest.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer test-token")
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, 200, rec.Code)
}
```

---

### Test Unauthorized Access

```go
func TestUnauthorized(t *testing.T) {
    router := lokstra.NewRouter()
    router.Use(jwtauth.Middleware(&jwtauth.Config{
        ValidatorServiceName: "auth_validator",
    }))
    
    // Test without token
    req := httptest.NewRequest("GET", "/protected", nil)
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, 401, rec.Code)
}
```

---

## Performance

**Overhead:** ~1-10ms per request

**Components:**
- Token parsing: ~500μs
- Signature verification: ~500μs
- Database user lookup: ~1-5ms

**Optimization:**
- Cache user info by token
- Use Redis for session storage
- Minimize database queries

---

## See Also

- **[Access Control](./access-control.md)** - Role-based access control
- **[CORS](./cors.md)** - Cross-origin handling
- **[Recovery](./recovery.md)** - Panic recovery

---

## Related Guides

- **[Authentication](../../04-guides/authentication/)** - Auth patterns
- **[Security Best Practices](../../04-guides/security/)** - Security tips
- **[Testing](../../04-guides/testing/)** - Testing strategies
