# Plural Resource Name Feature - Implementation Summary

## Problem Identified

User discovered an **inconsistency** in the resource name configuration:

1. ✅ **`ResourceName`** - Could be configured via:
   - YAML: `Service.ResourceName` (legacy) or `AutoRouter.ResourceName` (new)
   - Code: `ServiceRouterOptions.ResourceName`
   
2. ❌ **`PluralResourceName`** - Could ONLY be configured via:
   - Code: `ServiceRouterOptions.WithPluralResourceName()` ✅
   - YAML: **NOT AVAILABLE** ❌

This meant users couldn't override irregular plurals (e.g., "person" → "people" instead of "persons") from YAML config.

## Solution Implemented

### 1. Added `PluralResourceName` field to `AutoRouter` struct

**File:** `core/config/config.go`

```go
type AutoRouter struct {
    Convention         string           
    PathPrefix         string           
    ResourceName       string           
    PluralResourceName string           `yaml:"plural-resource-name,omitempty"` // ✨ NEW
    Routes             []*RouteOverride 
}
```

### 2. Added helper method `GetPluralResourceName()`

**File:** `core/config/config.go`

```go
func (s *Service) GetPluralResourceName() string {
    // Only check AutoRouter (no legacy field for plural)
    if s.AutoRouter != nil && s.AutoRouter.PluralResourceName != "" {
        return s.AutoRouter.PluralResourceName
    }
    return "" // Empty means will be auto-pluralized from ResourceName
}
```

### 3. Updated router generation to use plural resource name

**File:** `lokstra_registry/config.go`

```go
func generateRouterFromService(svc *config.Service) error {
    // ... existing code ...
    
    resourceName := svc.GetResourceName()
    pluralResourceName := svc.GetPluralResourceName()  // ✨ NEW
    
    options := router.DefaultServiceRouterOptions().
        WithConvention(convention).
        WithPrefix(servicePrefix).
        WithResourceName(resourceName)
    
    // Apply plural resource name if specified
    if pluralResourceName != "" {
        options = options.WithPluralResourceName(pluralResourceName)  // ✨ NEW
    }
    
    // ... rest of code ...
}
```

## Usage Examples

### YAML Configuration

```yaml
services:
  # Irregular plural override
  - name: person-service
    type: person_service
    auto-router:
      convention: rest
      resource-name: person
      plural-resource-name: people  # ✨ NEW: Override "persons" → "people"
  
  # Auto-pluralization (no override needed)
  - name: user-service
    type: user_service
    auto-router:
      convention: rest
      resource-name: user
      # Will auto-pluralize to "users" ✅
```

### Generated Routes

**Person Service** (with override):
```
GET    /people        → ListPersons()
GET    /people/{id}   → GetPerson()
POST   /people        → CreatePerson()
```

**User Service** (auto-pluralized):
```
GET    /users         → ListUsers()
GET    /users/{id}    → GetUser()
POST   /users         → CreateUser()
```

## Priority Logic

```
PluralResourceName resolution:
1. AutoRouter.PluralResourceName (YAML override) ← ✨ NEW
2. Auto-pluralize from ResourceName (fallback)
```

## Benefits

✅ **Consistency**: Both `ResourceName` and `PluralResourceName` can be configured via YAML  
✅ **Flexibility**: Users can override irregular plurals from config  
✅ **Backward Compatible**: Empty string means auto-pluralization (existing behavior)  
✅ **i18n Support**: Works for non-English resources with different plural rules  

## Use Cases Supported

1. **Irregular Plurals**: person → people, child → children, tooth → teeth
2. **Uncountable Nouns**: data → data, information → information
3. **Non-English**: Indonesian "produk" (same singular/plural)
4. **Default Behavior**: Regular plurals still auto-generated (user → users)

## Files Modified

1. ✅ `core/config/config.go` - Added `PluralResourceName` field and helper method
2. ✅ `lokstra_registry/config.go` - Updated router generation to use plural override
3. ✅ `docs/plural-resource-name-config.md` - Documentation
4. ✅ `cmd/examples/.../config-plural-example.yaml` - Usage examples

## Testing

No errors detected during compilation:
```bash
✅ No errors found.
```

## Related

- Resource Name: Singular form used as base (e.g., "user", "person")
- Plural Resource Name: Used in REST URLs (e.g., "/users", "/people")
- Auto-pluralization: Simple rules (add 's', 'es', or 'ies')
- REST Convention: Uses plural names for collection endpoints

## Technical Notes

### Why ResourceName wasn't used in `generateRouteFromMethodName()`

The confusion arose because:

1. `resourceName` parameter is passed to `generateRouteFromMethodName()`
2. But it's **not used** inside the function (marked with `_ = resourceName`)
3. Only `pluralName` is actually used for route generation

**Reason**: The conversion from `resourceName` → `pluralName` happens **before** calling `generateRouteFromMethodName()`, so the function only needs the final plural form.

```go
// In GenerateRoutes():
resourceName := options.ResourceName        // "person"
pluralName := options.PluralResourceName    // "" (empty)

if pluralName == "" {
    pluralName = pluralize(resourceName)    // "people" (if correct logic)
}

// Pass pluralName to route generation
routeMeta := c.generateRouteFromMethodName(methodName, resourceName, pluralName)
//                                                      ^^^^^^^^^^^  ^^^^^^^^^^
//                                                      not used     actually used
```

This is why the setup path was long but the actual usage was hidden in the pluralization step!

## Conclusion

The feature is now **complete and consistent**:
- ✅ Both singular and plural resource names configurable via YAML
- ✅ Proper priority and fallback logic
- ✅ Backward compatible
- ✅ Well documented with examples
