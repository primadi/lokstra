# Circular Dependency Detection

## Implementation

The Lokstra framework now includes **circular dependency detection** to prevent infinite loops during service resolution.

### How It Works

1. **Resolution Stack Tracking**: Each service resolution maintains a stack of service names being resolved
2. **Cycle Detection**: Before resolving a dependency, the framework checks if it's already in the resolution stack
3. **Early Panic**: If a cycle is detected, the framework immediately panics with a clear error message

### Example Error Message

```
circular dependency detected: service-a → service-b → service-a
```

This shows the exact chain of dependencies that caused the circular reference.

## Code Location

- **File**: `core/deploy/registry.go`
- **Method**: `getServiceAnyWithStack(name string, resolutionStack []string)`
- **Public API**: `GetServiceAny(name string) (any, bool)`

## Implementation Details

```go
// Public API - starts with empty stack
func (g *GlobalRegistry) GetServiceAny(name string) (any, bool) {
    return g.getServiceAnyWithStack(name, []string{})
}

// Internal implementation with circular detection
func (g *GlobalRegistry) getServiceAnyWithStack(name string, resolutionStack []string) (any, bool) {
    // Check for circular dependency
    for _, svcName := range resolutionStack {
        if svcName == name {
            chain := append(resolutionStack, name)
            panic(fmt.Sprintf("circular dependency detected: %s", strings.Join(chain, " → ")))
        }
    }
    
    // ... rest of implementation
    
    // Add to stack when resolving dependencies
    newStack := append(resolutionStack, name)
    depSvc, ok := g.getServiceAnyWithStack(serviceName, newStack)
}
```

## Test Coverage

See `circular_dependency_test.go` for test cases:

- ✅ `TestCircularDependency_StillCrashes` - Verifies circular dependency is detected and panics immediately
- ✅ Panic message includes clear dependency chain
- ✅ No infinite loops or hangs

## Benefits

1. **Immediate Feedback**: Developer sees error instantly instead of waiting for stack overflow
2. **Clear Error Message**: Shows exact circular dependency chain
3. **Prevent Resource Exhaustion**: No infinite loops consuming CPU/memory
4. **Better Developer Experience**: Easy to identify and fix circular dependencies

## Comparison: Before vs After

### Before (Infinite Loop)

```
// Service A depends on B
// Service B depends on A
reg.GetServiceAny("service-a")

// Result: Hangs forever until stack overflow
// - Consumes CPU spinning
// - Eventually crashes with obscure stack overflow error
```

### After (Immediate Detection)

```
// Service A depends on B
// Service B depends on A  
reg.GetServiceAny("service-a")

// Result: Immediate panic with clear message
// panic: circular dependency detected: service-a → service-b → service-a
```

## Related Changes

This was implemented as part of the **framework simplification** initiative:

1. Removed `service.Cached` wrapper complexity
2. Simplified dependency injection to direct type assertions
3. Added circular dependency detection for safety

See `_research/README.md` for more details on the research that led to these changes.
