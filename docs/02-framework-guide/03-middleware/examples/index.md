# Middleware Deep Dive - Examples

This folder contains advanced middleware patterns and custom middleware creation examples.

## Examples

### ✅ 01 - Custom Middleware
Build production-ready custom middleware.

**Topics**: Middleware creation, context handling, error propagation

[View Example](./01-custom-middleware/) | [main.go](./01-custom-middleware/main.go) | [test.http](./01-custom-middleware/test.http)

### ✅ 02 - Composition
Advanced middleware composition and chaining patterns.

**Topics**: Chaining, conditional middleware, dynamic loading

[View Example](./02-composition/) | [main.go](./02-composition/main.go) | [test.http](./02-composition/test.http)

### ✅ 03 - Context Management
Store and retrieve request-scoped data safely.

**Topics**: Context storage, propagation, thread safety

[View Example](./03-context-management/) | [main.go](./03-context-management/main.go) | [test.http](./03-context-management/test.http)

### ✅ 04 - Error Recovery
Panic recovery and graceful error handling.

**Topics**: Panic recovery, error transformation, logging

[View Example](./04-error-recovery/) | [main.go](./04-error-recovery/main.go) | [test.http](./04-error-recovery/test.http)

### ✅ 05 - Performance
Benchmark middleware overhead and optimization.

**Topics**: Overhead analysis, benchmarking, optimization

[View Example](./05-performance/) | [main.go](./05-performance/main.go) | [test.http](./05-performance/test.http)

### ✅ 06 - Integration
Integrate third-party middleware libraries.

**Topics**: Adapters, compatibility, migration

[View Example](./06-integration/) | [main.go](./06-integration/main.go) | [test.http](./06-integration/test.http)

---

## Running Examples

Each example follows this structure:
```
01-custom-middleware/
├── main.go              # Working code
├── index             # Detailed explanation
└── test.http            # HTTP test requests
```

To run an example:
```bash
cd 01-custom-middleware
go run main.go

# Test with test.http file or curl
curl http://localhost:3000
```

---

## Middleware Signature

All middleware in Lokstra follow this signature:

```go
func(c *request.Context) error
```

### Key Methods

- `c.Next()` - Call next middleware/handler
- `c.R` - Access *http.Request
- `c.Set(key, value)` - Store request-scoped data
- `c.Get(key)` - Retrieve stored data

### Example

```go
func LoggingMiddleware(c *request.Context) error {
    log.Printf("%s %s", c.R.Method, c.R.URL.Path)
    return c.Next()
}
```

---

**Status**: ✅ All 6 middleware examples complete and ready to use!
