# Go Slice Aliasing Bug - Lokstra Framework

## 🐛 Critical Bug Discovery (Oct 2025)

We discovered a **subtle slice aliasing bug** in `core/request/handler.go` that caused all routes to execute the wrong handler when exactly 5 global middleware were registered.

### Root Cause

```go
// ❌ BUGGY CODE (before fix)
func NewHandler(h HandlerFunc, mw ...HandlerFunc) *Handler {
    return &Handler{
        handlers: append(mw, h), // BUG: Can alias when len(mw) < cap(mw)
    }
}
```

When `len(mw)=5` and `cap(mw)=8`, Go's `append()` **reuses the underlying array** instead of allocating a new one. This causes multiple routes to share the same array, and the last registered handler overwrites previous handlers.

### Why Only at Specific Counts?

The bug only appeared with **exactly 5, 6, or 7 middleware** because of Go's slice capacity growth algorithm:

```
Middleware Count → Capacity → Behavior
1-4              → 4         → append() allocates new array (SAFE at boundary 4)
5-7              → 8         → append() reuses array (DANGEROUS!)
8                → 8         → append() allocates new array (SAFE at boundary 8)
9-15             → 16        → append() reuses array (DANGEROUS!)
```

**The bug is non-deterministic** based on slice growth - a "magic number" that changes based on capacity!

### The Fix

```go
// ✅ FIXED CODE
func NewHandler(h HandlerFunc, mw ...HandlerFunc) *Handler {
    // Force NEW allocation to prevent aliasing
    handlers := make([]HandlerFunc, len(mw)+1)
    copy(handlers, mw)
    handlers[len(mw)] = h
    
    return &Handler{
        handlers: handlers,
    }
}
```

## 🚨 Dangerous Patterns to Avoid

### ❌ DANGEROUS: Different Variables

```go
source := []T{...}           // len=5, cap=8
dest1 := append(source, x)   // Reuses underlying array!
dest2 := append(source, y)   // Overwrites dest1[5]!
// Result: dest1[5] = y (BUG!)
```

**Why dangerous:**
- You don't know the capacity of `source`
- When `len(source) < cap(source)`, append reuses the array
- Multiple slices share the same underlying array
- Last append overwrites previous data

### ✅ SAFE: Append to Self

```go
slice := []T{...}
slice = append(slice, x)  // SAFE: Same variable
slice = append(slice, y)  // SAFE: Same variable
```

**Why safe:**
- Source and destination are the same variable
- No aliasing because no other variable references the old array
- `slice` always points to the latest version

### ✅ SAFE: Prepend with Literal

```go
slice := []T{...}
slice = append([]T{x}, slice...)  // SAFE: Literal always allocates new
```

**Why safe:**
- `[]T{x}` is a literal slice with `len=1, cap=1`
- append() **must** allocate new array (capacity insufficient)
- No possibility of aliasing

### ✅ SAFE: Explicit Copy

```go
source := []T{...}
dest := make([]T, len(source)+1)
copy(dest, source)
dest[len(source)] = x
```

**Why safe:**
- Explicitly allocates new array with `make()`
- No shared underlying array

## 🛡️ Prevention Strategies

### 1. Use Linters

Enable **staticcheck** and **golangci-lint** in your CI/CD:

```bash
# Install
go install honnef.co/go/tools/cmd/staticcheck@latest

# Run
staticcheck ./...
```

Relevant rules:
- **SA4000**: warns about slice append to different variable
- **gocritic**: appendAssign rule

### 2. Code Review Checklist

When reviewing code with `append()`:
- [ ] Is destination a different variable than source?
- [ ] Could this be called multiple times with same source?
- [ ] Is the source slice capacity unknown/variable?

If YES to any: 🚨 **DANGEROUS! Use explicit copy pattern.**

### 3. Golden Rules

```go
// ✅ ALWAYS SAFE
x = append(x, ...)              // Append to self
x = append([]T{...}, x...)      // Prepend with literal
x = make([]T, n); copy(x, y)    // Explicit copy

// ⚠️  CONTEXT DEPENDENT (usually dangerous)
y := append(x, ...)             // Different variables - CHECK CAPACITY!

// ❌ NEVER DO THIS
func process(slice []T) []T {
    return append(slice, x)     // Caller's slice might be modified!
}
```

## 📚 Why Go Doesn't Warn About This

This is a **known controversial design decision** in Go:

1. **Performance**: Slice reuse is intentional optimization
2. **Backward Compatibility**: Millions of codebases would break
3. **"Read The Manual"**: Go team expects developers to understand slice internals
4. **Community Solutions**: Linters like staticcheck detect these patterns

## 🔗 References

- [Go Blog: Slices internals](https://go.dev/blog/slices-intro)
- [Go Slices: usage and internals](https://go.dev/blog/slices)
- [Staticcheck SA4000](https://staticcheck.io/docs/checks#SA4000)
- Our bug investigation: See `cmd/slice-capacity/` for demonstrations

## 💡 Key Takeaways

1. **`append()` can reuse arrays** when `len < cap` - this is expected Go behavior
2. **Our bug** was assuming `append()` always allocates new arrays
3. **Magic numbers** (like 5) appear because of slice growth algorithm
4. **Prevention**: Use linters + explicit `make()+copy()` when creating independent slices
5. **This is NOT a Go bug** - it's a footgun by design 😅

---

**Last Updated**: October 24, 2025  
**Bug Discovered By**: Prima (during middleware testing)  
**Time to Debug**: ~2 hours (because of non-obvious symptoms)  
**Fixed In**: `core/request/handler.go`
