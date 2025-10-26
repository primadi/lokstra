# Access Control Middleware

> Role-based access control (RBAC)

## Overview

Access Control middleware provides role-based authorization for protected endpoints. It checks if the authenticated user has one of the required roles before allowing access.

## Import Path

```go
import "github.com/primadi/lokstra/middleware/accesscontrol"
```

---

## Configuration

### Config Type

```go
type Config struct {
    AllowedRoles []string // List of roles that can access
    RoleField    string   // Field name in user metadata where role is stored
    ErrorMessage string   // Custom error message
}
```

**Fields:**
- `AllowedRoles` - List of roles allowed to access the endpoint
- `RoleField` - Field name in user metadata containing role (default: "role")
- `ErrorMessage` - Custom error message for unauthorized access

---

## Usage

### Basic Usage

```go
router.Use(accesscontrol.Middleware(&accesscontrol.Config{
    AllowedRoles: []string{"admin", "manager"},
    RoleField:    "role",
}))
```

---

### Admin Only

```go
adminRouter := router.Group("/admin")
adminRouter.Use(accesscontrol.Middleware(&accesscontrol.Config{
    AllowedRoles: []string{"admin"},
}))
```

---

### Multiple Roles

```go
router.Use(accesscontrol.Middleware(&accesscontrol.Config{
    AllowedRoles: []string{"admin", "manager", "supervisor"},
}))
```

---

### Custom Error Message

```go
router.Use(accesscontrol.Middleware(&accesscontrol.Config{
    AllowedRoles: []string{"admin"},
    ErrorMessage: "Administrator access required",
}))
```

---

## YAML Configuration

```yaml
middlewares:
  - type: accesscontrol
    params:
      allowed_roles: ["admin", "manager"]
      role_field: role
      error_message: "Insufficient permissions"
```

---

## Convenience Functions

### RequireAdmin

Requires admin role only.

```go
func RequireAdmin() request.HandlerFunc
```

**Example:**
```go
adminRouter := router.Group("/admin")
adminRouter.Use(accesscontrol.RequireAdmin())
```

---

### RequireAdminOrManager

Requires admin or manager role.

```go
func RequireAdminOrManager() request.HandlerFunc
```

**Example:**
```go
managementRouter := router.Group("/management")
managementRouter.Use(accesscontrol.RequireAdminOrManager())
```

---

### RequireRoles

Requires one of the specified roles.

```go
func RequireRoles(roles ...string) request.HandlerFunc
```

**Example:**
```go
router.Use(accesscontrol.RequireRoles("admin", "manager", "supervisor"))
```

---

## Prerequisites

### JWT Auth Required

Access Control middleware **requires** JWT Auth middleware to be applied first:

```go
// ✅ Correct order
router.Use(
    jwtauth.Middleware(&jwtauth.Config{
        ValidatorServiceName: "auth_validator",
    }), // First - authenticate
    accesscontrol.RequireAdmin(), // Then - authorize
)
```

---

### User Info Structure

User info from JWT Auth must have metadata with role field:

```go
type UserInfo struct {
    UserID   string
    Username string
    Email    string
    Metadata map[string]interface{}{
        "role": "admin", // Role stored here
    }
}
```

---

## Examples

### Admin Dashboard

```go
// Public routes
router.POST("/auth/login", loginHandler)

// Protected routes
apiRouter := router.Group("/api")
apiRouter.Use(jwtauth.Middleware(&jwtauth.Config{
    ValidatorServiceName: "auth_validator",
}))

// Admin-only routes
adminRouter := apiRouter.Group("/admin")
adminRouter.Use(accesscontrol.RequireAdmin())

adminRouter.GET("/users", listUsersHandler)
adminRouter.DELETE("/users/:id", deleteUserHandler)
adminRouter.GET("/settings", getSettingsHandler)
```

---

### Multi-Level Access

```go
// Authenticated users
apiRouter := router.Group("/api")
apiRouter.Use(jwtauth.Middleware(&jwtauth.Config{
    ValidatorServiceName: "auth_validator",
}))

apiRouter.GET("/profile", getProfileHandler) // Any authenticated user

// Manager-level access
managerRouter := apiRouter.Group("/management")
managerRouter.Use(accesscontrol.RequireAdminOrManager())

managerRouter.GET("/reports", getReportsHandler)
managerRouter.GET("/team", getTeamHandler)

// Admin-only access
adminRouter := apiRouter.Group("/admin")
adminRouter.Use(accesscontrol.RequireAdmin())

adminRouter.POST("/users", createUserHandler)
adminRouter.DELETE("/users/:id", deleteUserHandler)
```

---

### Custom Roles

```go
// Content moderators
moderatorRouter := apiRouter.Group("/moderate")
moderatorRouter.Use(accesscontrol.RequireRoles("admin", "moderator"))

moderatorRouter.PUT("/posts/:id/approve", approvePostHandler)
moderatorRouter.DELETE("/posts/:id", deletePostHandler)

// Premium users
premiumRouter := apiRouter.Group("/premium")
premiumRouter.Use(accesscontrol.RequireRoles("premium", "admin"))

premiumRouter.GET("/exclusive", getPremiumContentHandler)
```

---

### Per-Endpoint Authorization

```go
// Different roles for different endpoints
router.GET("/data/read", 
    accesscontrol.RequireRoles("viewer", "editor", "admin"),
    readDataHandler)

router.POST("/data/create",
    accesscontrol.RequireRoles("editor", "admin"),
    createDataHandler)

router.DELETE("/data/:id",
    accesscontrol.RequireAdmin(),
    deleteDataHandler)
```

---

### Multi-Tenant with Roles

```go
router.GET("/tenant/data", func(c *request.Context) error {
    // Get tenant and role from JWT
    tenantID := jwtauth.GetCurrentTenantID(c.R.Context())
    userInfo, _ := jwtauth.GetUserInfo(c.R.Context())
    role := userInfo.Metadata["role"].(string)
    
    // Check role
    if role != "admin" && role != "manager" {
        return c.Api.Forbidden("Manager or admin role required")
    }
    
    // Get data for tenant
    data := dataService.GetByTenant(tenantID)
    return c.Api.Ok(data)
})
```

---

### Custom Role Field

```go
// Role stored in different metadata field
router.Use(accesscontrol.Middleware(&accesscontrol.Config{
    AllowedRoles: []string{"super_admin"},
    RoleField:    "permission_level", // Custom field name
}))
```

---

## Error Responses

### Unauthenticated

If user is not authenticated (no JWT):

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "error": "user not authenticated"
}
```

---

### Insufficient Permissions

If user doesn't have required role:

```http
HTTP/1.1 403 Forbidden
Content-Type: application/json

{
  "error": "insufficient permissions"
}
```

---

### Custom Error Message

```http
HTTP/1.1 403 Forbidden
Content-Type: application/json

{
  "error": "Administrator access required"
}
```

---

## Best Practices

### 1. Always Use After JWT Auth

```go
// ✅ Good - JWT auth first
router.Use(
    jwtauth.Middleware(&jwtauth.Config{}),
    accesscontrol.RequireAdmin(),
)

// 🚫 Bad - access control without auth
router.Use(
    accesscontrol.RequireAdmin(), // No user info available!
)
```

---

### 2. Use Groups for Role-Based Routes

```go
// ✅ Good - organized by role
adminRouter := apiRouter.Group("/admin")
adminRouter.Use(accesscontrol.RequireAdmin())

managerRouter := apiRouter.Group("/management")
managerRouter.Use(accesscontrol.RequireAdminOrManager())

// 🚫 Bad - repeated middleware on each route
router.GET("/admin/users", accesscontrol.RequireAdmin(), handler1)
router.GET("/admin/settings", accesscontrol.RequireAdmin(), handler2)
router.GET("/admin/logs", accesscontrol.RequireAdmin(), handler3)
```

---

### 3. Provide Clear Error Messages

```go
// ✅ Good - helpful error message
accesscontrol.Middleware(&accesscontrol.Config{
    AllowedRoles: []string{"admin"},
    ErrorMessage: "Administrator access required",
})

// 🚫 Bad - generic error
accesscontrol.Middleware(&accesscontrol.Config{
    AllowedRoles: []string{"admin"},
    ErrorMessage: "Access denied",
})
```

---

### 4. Use Least Privilege

```go
// ✅ Good - minimum required roles
accesscontrol.RequireRoles("editor", "admin")

// 🚫 Bad - too permissive
accesscontrol.RequireRoles("viewer", "editor", "admin", "user")
```

---

### 5. Validate Role in User Service

```go
// ✅ Good - role is validated when JWT is issued
func (s *AuthService) Login(username, password string) (string, error) {
    user := s.userRepo.FindByUsername(username)
    
    // Validate role
    if !isValidRole(user.Role) {
        return "", errors.New("invalid role")
    }
    
    // Issue JWT with role in metadata
    token := issueJWT(user)
    return token, nil
}
```

---

## Performance

**Overhead:** ~100ns per request (role string comparison only)

**Impact:** Negligible

---

## Testing

### Test Admin Access

```go
func TestAdminAccess(t *testing.T) {
    // Mock JWT auth with admin role
    mockValidator := &MockValidator{
        GetUserInfoFunc: func(ctx context.Context, claims *auth.TokenClaims) (*auth.UserInfo, error) {
            return &auth.UserInfo{
                UserID: "user123",
                Metadata: map[string]interface{}{
                    "role": "admin",
                },
            }, nil
        },
    }
    
    lokstra_registry.RegisterService("auth_validator", mockValidator)
    
    router := lokstra.NewRouter()
    router.Use(
        jwtauth.Middleware(&jwtauth.Config{}),
        accesscontrol.RequireAdmin(),
    )
    
    router.GET("/admin", func(c *request.Context) error {
        return c.Api.Ok("admin page")
    })
    
    req := httptest.NewRequest("GET", "/admin", nil)
    req.Header.Set("Authorization", "Bearer token")
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, 200, rec.Code)
}
```

---

### Test Insufficient Permissions

```go
func TestInsufficientPermissions(t *testing.T) {
    // Mock JWT auth with user role (not admin)
    mockValidator := &MockValidator{
        GetUserInfoFunc: func(ctx context.Context, claims *auth.TokenClaims) (*auth.UserInfo, error) {
            return &auth.UserInfo{
                UserID: "user123",
                Metadata: map[string]interface{}{
                    "role": "user", // Not admin
                },
            }, nil
        },
    }
    
    lokstra_registry.RegisterService("auth_validator", mockValidator)
    
    router := lokstra.NewRouter()
    router.Use(
        jwtauth.Middleware(&jwtauth.Config{}),
        accesscontrol.RequireAdmin(),
    )
    
    req := httptest.NewRequest("GET", "/admin", nil)
    req.Header.Set("Authorization", "Bearer token")
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, 403, rec.Code)
}
```

---

## Common Patterns

### Hierarchical Roles

```go
// Define role hierarchy
var roleHierarchy = map[string]int{
    "user":       1,
    "moderator":  2,
    "manager":    3,
    "admin":      4,
    "super_admin": 5,
}

func hasPermission(userRole string, requiredRole string) bool {
    return roleHierarchy[userRole] >= roleHierarchy[requiredRole]
}

// Use in middleware
router.Use(func(c *request.Context) error {
    userInfo, _ := jwtauth.GetUserInfo(c.R.Context())
    userRole := userInfo.Metadata["role"].(string)
    
    if !hasPermission(userRole, "manager") {
        return c.Api.Forbidden("Manager role required")
    }
    
    return c.Next()
})
```

---

### Resource-Based Authorization

```go
// Check if user owns resource
router.DELETE("/posts/:id", func(c *request.Context) error {
    postID := c.Params.Get("id")
    userID := jwtauth.GetCurrentUserID(c.R.Context())
    
    post := postService.GetByID(postID)
    
    // Owner or admin can delete
    userInfo, _ := jwtauth.GetUserInfo(c.R.Context())
    role := userInfo.Metadata["role"].(string)
    
    if post.AuthorID != userID && role != "admin" {
        return c.Api.Forbidden("Cannot delete others' posts")
    }
    
    postService.Delete(postID)
    return c.Api.NoContent()
})
```

---

### Permission-Based Authorization

```go
// Check specific permissions instead of roles
type UserInfo struct {
    UserID      string
    Permissions []string // ["read:users", "write:users", "delete:users"]
}

func requirePermission(permission string) request.HandlerFunc {
    return func(c *request.Context) error {
        userInfo, _ := jwtauth.GetUserInfo(c.R.Context())
        permissions := userInfo.Metadata["permissions"].([]string)
        
        if !slices.Contains(permissions, permission) {
            return c.Api.Forbidden("Permission denied")
        }
        
        return c.Next()
    }
}

router.DELETE("/users/:id",
    requirePermission("delete:users"),
    deleteUserHandler)
```

---

## See Also

- **[JWT Auth](./jwt-auth.md)** - Authentication middleware
- **[CORS](./cors.md)** - Cross-origin handling
- **[Recovery](./recovery.md)** - Panic recovery

---

## Related Guides

- **[Authorization](../../04-guides/authorization/)** - Authorization patterns
- **[Security Best Practices](../../04-guides/security/)** - Security tips
- **[Multi-Tenancy](../../04-guides/multi-tenancy/)** - Multi-tenant setup
