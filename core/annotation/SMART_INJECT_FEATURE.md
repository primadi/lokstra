# Smart @Inject Implementation - Feature Summary

## Overview

Unified `@Inject` annotation that supports service injection, config value injection, and indirection for both patterns.

## Changes Made

### 1. Updated `ConfigInfo` struct (complex_processor.go)
Added fields for indirection support:
- `IsIndirect bool` - Indicates if config key is resolved from another config
- `IndirectKey string` - The config key that contains the actual key

### 2. Enhanced `@Inject` annotation (codegen.go)

**NEW: `cfg:` prefix for config value injection**

```go
type MyService struct {
    // @Inject "cfg:app.timeout"
    Timeout time.Duration  // Config value injection
    
    // @Inject "cfg:@jwt.key-path"
    JWTSecret string  // Indirect config resolution
}
```

### 3. Added helper functions (codegen.go)
- `parseIndirectConfigValue()` - Handles indirect config resolution
- `generateIndirectConfigCode()` - Generates type-safe Go code for indirection

## Syntax Reference

### @Inject Patterns (Complete)

| Syntax | Purpose | Example |
|--------|---------|---------|
| `@Inject "service-name"` | Direct service | `@Inject "user-repo"` |
| `@Inject "@config.key"` | Service from config | `@Inject "@store.impl"` |
| `@Inject "cfg:config.key"` | Config value | `@Inject "cfg:app.timeout"` |
| `@Inject "cfg:@config.key"` | Indirect config | `@Inject "cfg:@jwt.path"` |

## Use Cases

### 1. Environment-specific Configuration

**Before:**
```go
type EmailService struct {
    // Hard-coded config key
    // @Inject "cfg:email.api-key"
    APIKey string
}
```

**After:**
```go
type EmailService struct {
    // Dynamic config key based on environment
    // @Inject "cfg:@email.api-key-path"
    APIKey string
}

// config.yaml
configs:
  email:
    api-key-path: "secrets.email.prod-key"  # Changes per environment
  
  secrets:
    email:
      dev-key: "dev-123"
      prod-key: "prod-xyz"
```

### 2. Unified Injection Pattern

```go
type UserService struct {
    // @Inject "user-repo"
    UserRepo UserRepository
    
    // @Inject "cfg:app.timeout"
    Timeout time.Duration
    
    // @Inject "cfg:@cache.ttl-key"
    CacheTTL time.Duration
}
```

### 3. Feature Flags with Dynamic Config

```go
type FeatureService struct {
    // @Inject "cfg:@feature.flags-key"
    EnabledFeatures []string
}

// config.yaml
configs:
  feature:
    flags-key: "features.production"  # Switch between feature sets
  
  features:
    production:
      - "feature-a"
      - "feature-b"
    
    beta:
      - "feature-a"
      - "feature-b"
      - "feature-c"
```

## Benefits

1. **Simplicity**: Single `@Inject` annotation for everything
2. **Flexibility**: Config keys can be dynamically resolved
3. **Type-safe**: Generated code handles all type conversions
4. **Environment-agnostic**: Switch configs without code changes
5. **Consistent**: Same pattern for services and config values

## Implementation Details

### Code Generation

For indirect config (`@Inject "cfg:@jwt.key-path"`):

```go
func() string {
    if actualKey, ok := cfg["jwt.key-path"].(string); ok && actualKey != "" {
        if v, ok := cfg[actualKey].(string); ok { return v }
    }
    return ""
}()
```

Supports all types:
- Primitives: `string`, `int`, `bool`, `float64`
- Time: `time.Duration`
- Bytes: `[]byte`
- Slices: `[]string`, `[]int`, `[]struct`
- Structs: Custom types with `cast.ToStruct`

### Type Safety

All generated code includes:
- Type assertions with fallbacks
- Default value handling
- Error-safe parsing (time.Duration, numeric conversions)
- Nil safety for slices and pointers

## Migration Guide

### Using cfg: prefix

```go
// Direct config value
// @Inject "cfg:app.timeout"
Timeout time.Duration

// With indirection
// @Inject "cfg:@jwt.key-path"
JWTSecret string

// config.yaml
configs:
  jwt:
    key-path: "security.production-jwt-secret"  # Change this key per environment
  
  security:
    production-jwt-secret: "prod-secret"
    development-jwt-secret: "dev-secret"
```

## Testing

See [codegen_inject_test_example.go](./codegen_inject_test_example.go) for comprehensive examples.

## Documentation Updated

- ✅ [.github/copilot-instructions.md](.github/copilot-instructions.md) - Updated injection patterns table
- ✅ Template comments in generated code - Shows all supported patterns
- ✅ Example file - Comprehensive use cases

## Notes

- Single `@Inject` annotation for all injection types
- Use `cfg:` prefix for config value injection
- Use `cfg:@` prefix for indirect config resolution
- Indirect resolution (`@` prefix) resolves actual config key from another config value
- All config value types are fully supported with type conversion
