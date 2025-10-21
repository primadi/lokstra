# Router Schema Refactoring Summary

## Overview

Refactored router configuration schema to eliminate redundancy and unify router definitions across YAML config and runtime API.

## Changes Made

### ✅ 1. Removed Redundant Types

**Removed:**
- `schema.RouterDefSimple` - Limited functionality (no path prefix, custom routes)
- Duplicate `RouterOverrideDef` and `RouteDef` definitions

**Result:** Single source of truth for router configuration

### ✅ 2. Unified Router Definition

**New `schema.RouterDef`:**
```go
type RouterDef struct {
    Service        string // Service name to generate router from
    Convention     string // Convention type (rest, rpc, graphql)
    Resource       string // Singular form, e.g., "user"
    ResourcePlural string // Plural form, e.g., "users" (optional)
    Overrides      string // Reference to RouterOverrideDef name (optional)
}
```

**Features:**
- Full convention support (REST, RPC, GraphQL)
- Resource naming control
- Reference-based overrides

### ✅ 3. Aligned RouterOverrideDef with autogen.RouteOverride

**New Structure:**
```go
type RouterOverrideDef struct {
    PathPrefix  string     // e.g., "/api/v1"
    Middlewares []string   // Router-level middleware names
    Hidden      []string   // Methods to hide
    Custom      []RouteDef // Custom route definitions
}

type RouteDef struct {
    Name        string   // Method name
    Method      string   // HTTP method override
    Path        string   // Path override
    Middlewares []string // Route-level middleware names
}
```

**Key Differences from Runtime API:**
- `[]string` for middlewares (names) vs `[]any` (instances)
- `Custom []RouteDef` (array) vs `Custom map[string]Route` (map)

### ✅ 4. Converter Functions

**Added `schema/converter.go`:**
```go
// Convert YAML router def to runtime conversion rule
func (r *RouterDef) ToConversionRule() autogen.ConversionRule

// Convert YAML overrides to runtime overrides
func (r *RouterOverrideDef) ToRouteOverride(registry MiddlewareRegistry) autogen.RouteOverride
```

**Middleware Resolution:**
```go
type DefaultMiddlewareRegistry struct{}

func (d *DefaultMiddlewareRegistry) GetMiddleware(name string) (any, bool) {
    if mw := lokstra_registry.CreateMiddleware(name); mw != nil {
        return mw, true
    }
    return nil, false
}
```

### ✅ 5. Router Registration

**Already Exists:**
```go
// Register router programmatically
lokstra_registry.RegisterRouter("my-router", router)

// Retrieve router
router := lokstra_registry.GetRouter("my-router")
```

---

## YAML Configuration Example

### Before (RouterDefSimple)

```yaml
routers:
  user-router:
    service: user-service
    overrides:
      GetSpecial:
        hide: false
        middleware:
          - auth
          - rate-limit
```

**Limitations:**
- ❌ No path prefix
- ❌ No custom paths/methods
- ❌ No global middlewares
- ❌ No convention control

### After (Unified RouterDef)

```yaml
# Define reusable overrides
router-overrides:
  api-overrides:
    path-prefix: /api/v1
    middlewares:
      - auth
      - logging
    hidden:
      - InternalMethod
    custom:
      - name: GetSpecial
        method: GET
        path: /special-users/{id}
        middlewares:
          - rate-limit

# Define routers
routers:
  user-router:
    service: user-service
    convention: rest
    resource: user
    resource-plural: users
    overrides: api-overrides

  order-router:
    service: order-service
    convention: rest
    resource: order
    overrides: api-overrides  # Reuse same overrides
```

**Benefits:**
- ✅ Full control: path prefix, conventions, resource naming
- ✅ Reusable overrides across multiple routers
- ✅ Global + route-specific middlewares
- ✅ Custom routes with method/path overrides
- ✅ Hide unwanted methods

---

## Code Usage Example

### 1. Register Middlewares and Services

```go
// Register middleware factories
lokstra_registry.RegisterMiddlewareFactory("auth", func(cfg map[string]any) request.HandlerFunc {
    return func(ctx *request.Context) error {
        // Auth logic
        return nil
    }
})

// Register middleware instances by name
lokstra_registry.RegisterMiddlewareName("auth", "auth", map[string]any{
    "required_role": "admin",
})
lokstra_registry.RegisterMiddlewareName("logging", "logging", nil)
lokstra_registry.RegisterMiddlewareName("rate-limit", "rate-limit", map[string]any{
    "requests_per_minute": 60,
})

// Register service
userService := &UserService{}
lokstra_registry.RegisterService("user-service", userService)
```

### 2. Create Router from YAML Config

```go
// Load YAML config
var config schema.DeployConfig
yaml.Unmarshal(yamlData, &config)

// Get router definition
routerDef := config.Routers["user-router"]
overrideDef := config.RouterOverrides[routerDef.Overrides]

// Get service instance
service := lokstra_registry.GetService[*UserService](routerDef.Service)

// Convert to runtime types
conversionRule := routerDef.ToConversionRule()
middlewareRegistry := &schema.DefaultMiddlewareRegistry{}
routeOverride := overrideDef.ToRouteOverride(middlewareRegistry)

// Create router using autogen
router := autogen.NewFromService(service, conversionRule, routeOverride)

// Register router for later use
lokstra_registry.RegisterRouter("user-router", router)
```

### 3. Register Router Programmatically (Without YAML)

```go
// Create service
userService := &UserService{}

// Define conversion rule
conversionRule := autogen.ConversionRule{
    Convention:     convention.REST,
    Resource:       "user",
    ResourcePlural: "users",
}

// Define overrides
routeOverride := autogen.RouteOverride{
    PathPrefix: "/api/v1",
    Middlewares: []any{
        authMiddleware,
        loggingMiddleware,
    },
    Hidden: []string{"InternalMethod"},
    Custom: map[string]autogen.Route{
        "GetSpecial": {
            Method: "GET",
            Path:   "/special-users/{id}",
            Middlewares: []any{rateLimitMiddleware},
        },
    },
}

// Create router
router := autogen.NewFromService(userService, conversionRule, routeOverride)

// Register for use in other parts of application
lokstra_registry.RegisterRouter("user-router", router)
```

---

## Migration Guide

### For Existing YAML Configs

**Old Format (RouterDefSimple):**
```yaml
routers:
  my-router:
    service: my-service
    overrides:
      MyMethod:
        hide: true
```

**New Format:**
```yaml
# Optional: Define shared overrides
router-overrides:
  my-overrides:
    hidden:
      - MyMethod

# Updated router definition
routers:
  my-router:
    service: my-service
    convention: rest  # REQUIRED: specify convention
    resource: myresource  # OPTIONAL: defaults to service name
    overrides: my-overrides  # OPTIONAL: reference overrides
```

**Required Changes:**
1. Add `convention` field (e.g., `rest`, `rpc`, `graphql`)
2. Move `overrides.*.hide: true` → `router-overrides.*.hidden: [...]`
3. Move `overrides.*.middleware` → `router-overrides.*.custom[].middlewares`

### For Code Registration

**No changes needed!** `lokstra_registry.RegisterRouter()` already exists and works as before.

---

## Benefits

| Aspect | Before | After |
|--------|--------|-------|
| **Router Types** | 3 (RouterDefSimple, RouterOverrideDef x2, RouteDef x2) | 2 (RouterDef, RouterOverrideDef) |
| **Code Duplication** | 2 sets of override definitions | 1 unified definition |
| **YAML Flexibility** | Limited (no prefix, no custom routes) | Full control |
| **Reusability** | No (inline only) | Yes (reference-based) |
| **Convention Support** | Implicit | Explicit (REST/RPC/GraphQL) |
| **Runtime Registration** | Manual only | YAML + Manual |

---

## Files Changed

1. **core/deploy/schema/schema.go**
   - Removed `RouterDefSimple`, `RouteConfig`
   - Updated `RouterDef` with full configuration
   - Unified `RouterOverrideDef` and `RouteDef`

2. **core/deploy/schema/converter.go** (NEW)
   - `ToConversionRule()` - Convert RouterDef → autogen.ConversionRule
   - `ToRouteOverride()` - Convert RouterOverrideDef → autogen.RouteOverride
   - `DefaultMiddlewareRegistry` - Middleware name resolver

3. **core/router/autogen/autogen.go**
   - Added middleware support to `Route` struct
   - Fixed path joining with `path.Join()`
   - Pass middlewares to router methods

---

## Testing Recommendations

1. **Unit Tests**
   - Test `ToConversionRule()` conversion
   - Test `ToRouteOverride()` with middleware resolution
   - Test path joining edge cases

2. **Integration Tests**
   - Load YAML config and create routers
   - Verify middleware chain execution order
   - Test hidden routes are not registered
   - Test custom route overrides work

3. **Migration Tests**
   - Convert old YAML format to new format
   - Verify functional equivalence

---

## Future Enhancements

1. **Validation**
   - Add JSON Schema validation for YAML config
   - Validate middleware names exist before resolution
   - Validate convention types

2. **Error Handling**
   - Better error messages for missing middlewares
   - Validation errors with line numbers (YAML parser)

3. **Additional Conventions**
   - GraphQL convention implementation
   - gRPC convention implementation
   - Custom convention support

---

## Summary

This refactoring achieves:
- ✅ **DRY Principle** - Single definition for router config
- ✅ **Flexibility** - Full control over routes, paths, middlewares
- ✅ **Reusability** - Reference-based overrides
- ✅ **Extensibility** - Easy to add new conventions
- ✅ **Backward Compatibility** - Manual registration still works
- ✅ **Clean API** - Clear separation between YAML (schema) and Runtime (autogen)
