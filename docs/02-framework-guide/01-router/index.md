# Router Deep Dive

> **Master all 29 handler forms and advanced routing patterns**  
> **Time**: 60-75 minutes • **Level**: Advanced • **Prerequisites**: [Essentials - Router](../../01-router-guide/01-router/)

---

## 🎯 What You'll Learn

- All 29 handler signatures and when to use each
- Handler lifecycle (before/after hooks)
- Advanced parameter binding
- Route priorities and conflict resolution
- Error handling strategies
- Performance optimizations

---

## 📚 Topics

### 1. All Handler Forms
Understanding all 29 handler signatures:
- Basic forms (void, error, response)
- Context forms (with *request.Context)
- Parameter forms (with path/query params)
- Combined forms (context + params + response)

### 2. Handler Selection Strategy
Learn when to use each handler form:
- Performance considerations
- Code clarity vs flexibility
- Error handling needs
- Response formatting requirements

### 3. Advanced Parameter Binding
Deep dive into parameter extraction:
- Path parameters
- Query parameters
- Header parameters
- Body binding
- Custom validators

### 4. Route Lifecycle
Master the handler lifecycle:
- Before hooks
- Main handler execution
- After hooks
- Error handling flow

### 5. Route Priorities
Understanding route matching:
- Exact matches vs patterns
- Parameter routes
- Wildcard routes
- Conflict resolution

### 6. Error Handling Patterns
Advanced error handling:
- Structured errors
- Error middleware
- Custom error responses
- Error recovery

### 7. Performance Optimization
Optimize your routes:
- Handler form selection impact
- Parameter binding overhead
- Response serialization
- Benchmarking techniques

### 8. Debugging and Testing
Debug complex routing:
- Route inspection
- Request tracing
- Unit testing handlers
- Integration testing

---

## 📂 Examples

All examples are in the `examples/` folder:

### [01 - All Handler Forms](examples/01-all-handler-forms/)
Demonstrates all 29 handler signatures with working examples.

### [02 - Parameter Binding](examples/02-parameter-binding/)
Advanced parameter extraction and validation.

### [03 - Lifecycle Hooks](examples/03-lifecycle-hooks/)
Before/after hooks and middleware integration.

### [04 - Route Priorities](examples/04-route-priorities/)
Understanding route matching and conflicts.

### [05 - Error Handling](examples/05-error-handling/)
Structured error handling patterns.

### [06 - Performance](examples/06-performance/)
Benchmarks and optimization techniques.

### [07 - Testing](examples/07-testing/)
Unit and integration testing strategies.

---

## 🚀 Quick Start

```bash
# Run any example
cd docs/02-deep-dive/01-router/examples/01-all-handler-forms
go run main.go

# Test with provided test.http
# (use VS Code REST Client extension)
```

---

## 📖 Prerequisites

Before diving in, make sure you understand:
- [Basic routing](../../01-router-guide/01-router/)
- [Handler basics](../../01-router-guide/01-router/#handlers)
- [Parameter binding](../../01-router-guide/01-router/#parameters)

---

## 🎯 Learning Path

1. **Study all handler forms** → Understand available options
2. **Learn selection strategy** → Choose the right form
3. **Master parameters** → Extract data efficiently
4. **Understand lifecycle** → Control execution flow
5. **Handle errors** → Build robust handlers
6. **Optimize** → Improve performance
7. **Test** → Ensure correctness

---

## 💡 Key Takeaways

After completing this section:
- ✅ You'll know all 29 handler forms
- ✅ You'll choose the right form for each use case
- ✅ You'll handle complex parameter scenarios
- ✅ You'll write testable, performant handlers
- ✅ You'll debug routing issues effectively

---

**Coming Soon** - Examples and detailed content are being prepared.

**Next**: [Service Deep Dive](../02-service/) →
