# Annotation Parsing Example

This example demonstrates how the Lokstra annotation parser correctly:

1. **Detects valid annotations** - Annotations directly above declarations
2. **Ignores indented annotations** - TAB or multi-space indented (code examples in docs)
3. **Ignores annotations with too many empty lines** - Prevents matching stray annotations

## Files

- `annotation_example.go` - Sample Go file with both valid and invalid annotations
- `main.go` - Parser test program

## Run

```bash
cd core/annotation/examples/annotation_parsing
go run .
```

## Expected Output

The parser should find **4 valid annotations**:
1. `@Handler` on `UserService` struct
2. `@Inject` on `UserRepo` field
3. `@Route` on `GetByID` method
4. `@Route` on `Create` method

And should **IGNORE** these (indented in documentation):
- Line 8: `@Handler` in RegisterMiddleware doc (TAB-indented)
- Line 22: `@Route` in AnotherFunction doc (multi-space indented)

## Rules

### Valid Annotation Format
```go
// @Handler name="service-name"
type MyService struct {}
```

### Invalid (Ignored) Formats

**TAB-indented (Go doc code example):**
```go
// Example:
//
//	@Handler name="example"
//
// The above is ignored
```

**Multi-space indented:**
```go
// Example:
//
//    @Handler name="example"
//
// The above is ignored (3+ spaces)
```

**Too many empty lines:**
```go
// @Handler name="example"
//
//
//
//
// Too far from declaration - ignored
type MyService struct {}
```
