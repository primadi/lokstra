# Router Deep Dive - Examples

This folder contains advanced routing examples demonstrating all handler forms and patterns.

## Examples

### 01 - All Handler Forms *(Coming Soon)*
Demonstrates all 29 handler signatures with working code.

**Topics**: Handler forms, signatures, when to use each

### 02 - Parameter Binding *(Coming Soon)*
Advanced parameter extraction, validation, and custom binding.

**Topics**: Path params, query params, headers, body binding, validation

### 03 - Lifecycle Hooks *(Coming Soon)*
Before/after hooks and middleware integration patterns.

**Topics**: Before hooks, after hooks, middleware chain

### 04 - Route Priorities *(Coming Soon)*
Understanding route matching, conflicts, and resolution.

**Topics**: Route matching, priorities, wildcards

### 05 - Error Handling *(Coming Soon)*
Structured error handling and recovery patterns.

**Topics**: Error types, middleware, recovery, responses

### 06 - Performance *(Coming Soon)*
Benchmarks and optimization techniques.

**Topics**: Handler overhead, benchmarking, optimization

### 07 - Testing *(Coming Soon)*
Unit and integration testing strategies.

**Topics**: Unit tests, integration tests, mocking

---

## Running Examples

Each example follows this structure:
```
01-all-handler-forms/
├── main.go              # Working code
├── index            # Detailed explanation
├── test.http            # HTTP test requests
└── benchmarks_test.go   # Performance tests (if applicable)
```

To run an example:
```bash
cd 01-all-handler-forms
go run main.go

# Test with curl or test.http file
curl http://localhost:3000/endpoint
```

---

**Status**: Examples are being prepared and will be available soon.
