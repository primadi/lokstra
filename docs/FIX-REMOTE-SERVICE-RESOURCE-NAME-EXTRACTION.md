# Fix: Remote Service Resource Name Extraction

## Problem

Remote service client was generating incorrect paths when `resource-name` was not explicitly set in YAML config:

```yaml
services:
  - name: user-service
    type: user_service
    auto-router:
      convention: "rest"
      path-prefix: "/api/v1"
      # resource-name: "user"  ← NOT EXPLICITLY SET
```

**Result:**
```
methodName: "GetUser"
path: /api/v1/resources/{id}  ❌ WRONG!
```

**Expected:**
```
methodName: "GetUser"
path: /api/v1/users/{id}  ✅ CORRECT!
```

## Root Cause

In `lokstra_registry/config.go`, when registering remote services, the code was checking if `svc.AutoRouter.ResourceName` was explicitly set, but **NOT using the fallback logic** that extracts resource name from service type.

### Before (Broken)

```go
// REMOTE SERVICE registration
if svc.AutoRouter != nil {
    if svc.AutoRouter.ResourceName != "" {
        remoteConfig["resource-name"] = svc.AutoRouter.ResourceName
    }
    // ❌ If not set, resource-name is missing from config!
}
```

When `resource-name` was not in YAML, it was **not passed to remote service**, causing RemoteService to use the fallback "resources" in `buildPathWithParams()`.

### The Correct Logic

The `Service` struct already has a method `GetResourceName()` that implements the fallback logic:

```go
// From core/config/config.go
func (s *Service) GetResourceName() string {
    // Priority: AutoRouter > derived from type
    if s.AutoRouter != nil && s.AutoRouter.ResourceName != "" {
        return s.AutoRouter.ResourceName
    }
    // Extract resource name from service type if not explicitly set
    // E.g., "user_service" -> "user"
    if s.Type != "" {
        return extractResourceNameFromType(s.Type)
    }
    return s.Name
}
```

And `extractResourceNameFromType()`:

```go
// From core/config/helper.go
func extractResourceNameFromType(serviceType string) string {
    name := serviceType
    
    // 1. Take part before underscore: "user_service" -> "user"
    if idx := strings.Index(name, "_"); idx != -1 {
        name = name[:idx]
    }
    
    // 2. Remove common suffixes (Service, Impl, Handler, etc.)
    for _, suffix := range commonSuffixes {
        if strings.HasSuffix(name, suffix) {
            name = strings.TrimSuffix(name, suffix)
            break
        }
    }
    
    // 3. Remove versioning like `_v1`, `_V2`
    name = versionSuffixRe.ReplaceAllString(name, "")
    
    // 4. Normalize casing
    return strings.ToLower(strings.TrimSpace(name))
}
```

## Solution

Use the helper methods (`GetResourceName()`, `GetConvention()`, `GetPathPrefix()`, etc.) when building remote service config, instead of directly accessing `AutoRouter` fields.

### After (Fixed)

```go
// REMOTE SERVICE registration
remoteConfig := map[string]any{
    "base_url":     location.BaseURL,
    "service_name": svc.Name,
    "router":       svc.Name,
}

// ✅ Use helper methods with fallback logic
convention := svc.GetConvention("")
if convention != "" {
    remoteConfig["convention"] = convention
}

pathPrefix := svc.GetPathPrefix()
if pathPrefix != "" {
    remoteConfig["path-prefix"] = pathPrefix
}

resourceName := svc.GetResourceName()  // ✅ Has fallback!
if resourceName != "" {
    remoteConfig["resource-name"] = resourceName
}

pluralResourceName := svc.GetPluralResourceName()
if pluralResourceName != "" {
    remoteConfig["plural-resource-name"] = pluralResourceName
}

// Route overrides
routeOverrides := svc.GetRouteOverrides()
if len(routeOverrides) > 0 {
    // ... convert to config format
}
```

## Benefits

### 1. **Consistency with Server-Side**

Server-side router generation already uses these helper methods:

```go
// From lokstra_registry/config.go - generateRouterFromService()
convention := svc.GetConvention("")       // ✅ Uses helper
servicePrefix := svc.GetPathPrefix()      // ✅ Uses helper
resourceName := svc.GetResourceName()     // ✅ Uses helper
pluralResourceName := svc.GetPluralResourceName()  // ✅ Uses helper
```

Now remote service registration uses the **same helpers**, ensuring consistency.

### 2. **Automatic Resource Name Extraction**

```yaml
services:
  - name: user-service
    type: user_service          # ← Will extract "user"
    auto-router:
      convention: "rest"
      path-prefix: "/api/v1"
      # No need to specify resource-name! ✅
```

**Works for all patterns:**
- `user_service` → `user`
- `order_service` → `order`
- `payment_service` → `payment`
- `UserService` → `user`
- `OrderServiceImpl` → `order`

### 3. **Less Configuration Required**

**Before:**
```yaml
# Had to explicitly set resource-name
- name: user-service
  type: user_service
  auto-router:
    resource-name: "user"  ← Required!
```

**After:**
```yaml
# Auto-extracted from type
- name: user-service
  type: user_service
  auto-router:
    # resource-name auto-extracted ✅
```

### 4. **Override Still Works**

If you want to override, you still can:

```yaml
- name: person-service
  type: user_service
  auto-router:
    resource-name: "person"     # ← Override
    plural-resource-name: "people"  # ← Override
```

## Testing

### Example Service Type Extraction

| Service Type | Extracted Resource Name |
|-------------|------------------------|
| `user_service` | `user` |
| `order_service` | `order` |
| `payment_service` | `payment` |
| `auth_service` | `auth` |
| `cart_service` | `cart` |
| `invoice_service` | `invoice` |
| `UserService` | `user` |
| `OrderServiceImpl` | `order` |

### Path Generation Test

**Config:**
```yaml
- name: user-service
  type: user_service
  auto-router:
    convention: "rest"
    path-prefix: "/api/v1"
```

**Method:** `GetUser`

**Generated Path:**
```
Before: /api/v1/resources/{id}  ❌
After:  /api/v1/users/{id}      ✅
```

**Server Route:**
```
[userServiceLocal] GET /api/v1/users/{id} -> userServiceLocal.GetUser
```

**Client Call:**
```go
httpMethod: GET
path: /api/v1/users/{id}
```

✅ **PERFECT MATCH!**

## Related Changes

### Modified Files
1. `lokstra_registry/config.go` - Use helper methods in remote service registration

### Helper Methods Used
- `svc.GetConvention(globalDefault)` - Get convention with fallback
- `svc.GetPathPrefix()` - Get path prefix
- `svc.GetResourceName()` - **Get resource name with auto-extraction** ✅
- `svc.GetPluralResourceName()` - Get plural name
- `svc.GetRouteOverrides()` - Get route overrides

### Extraction Logic
- `extractResourceNameFromType(serviceType)` - Extract resource from type name

## Examples

### Example 1: Minimal Config

```yaml
services:
  - name: user-service
    type: user_service
    auto-router:
      path-prefix: "/api/v1"
```

**Auto-extracted:**
- `convention`: "rest" (default)
- `resource-name`: "user" (from type)
- `plural-resource-name`: "users" (auto-pluralized)

**Generated paths:**
- `GetUser` → `GET /api/v1/users/{id}`
- `ListUsers` → `GET /api/v1/users`
- `CreateUser` → `POST /api/v1/users`

### Example 2: With Override

```yaml
services:
  - name: people-service
    type: user_service
    auto-router:
      path-prefix: "/api/v1"
      resource-name: "person"
      plural-resource-name: "people"
```

**Applied config:**
- `convention`: "rest" (default)
- `resource-name`: "person" (override)
- `plural-resource-name`: "people" (override)

**Generated paths:**
- `GetUser` → `GET /api/v1/people/{id}`
- `ListUsers` → `GET /api/v1/people`
- `CreateUser` → `POST /api/v1/people`

### Example 3: With Route Overrides

```yaml
services:
  - name: auth-service
    type: auth_service
    auto-router:
      path-prefix: "/api/v1"
      routes:
        - name: Login
          method: POST
          path: "/login"
        - name: ValidateToken
          method: POST
          path: "/validate-token"
```

**Auto-extracted:**
- `resource-name`: "auth" (from type)

**Generated paths:**
- `Login` → `POST /api/v1/login` (override)
- `ValidateToken` → `POST /api/v1/validate-token` (override)
- `GetSession` → `GET /api/v1/auths/{id}` (convention)

## Impact

### Before Fix

❌ Remote services with implicit resource names generated wrong paths
❌ Required explicit `resource-name` in YAML for every service
❌ Inconsistent with server-side router generation
❌ More verbose configuration

### After Fix

✅ Remote services auto-extract resource names from type
✅ Minimal YAML configuration required
✅ Consistent with server-side router generation
✅ Less configuration, same behavior

## Conclusion

This fix ensures that remote service clients use the **same resource name extraction logic** as server-side routers, providing:

1. **Consistency** - Client and server use same resource names
2. **Convention over Configuration** - Less YAML needed
3. **Maintainability** - Single source of truth for extraction logic
4. **Correctness** - Paths match server routes exactly

**Key principle:** Use the helper methods (`GetResourceName()`, etc.) instead of directly accessing `AutoRouter` fields to get the full benefit of fallback logic.
