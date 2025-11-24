# Research & Proof-of-Concept Tests

This folder contains **exploration tests** that prove concepts and behaviors during framework development. These are **NOT** production unit tests.

## Purpose

These tests are used to:
- **Prove** design decisions (e.g., lazy loading behavior)
- **Explore** edge cases (e.g., circular dependency handling)
- **Document** framework behavior for developers
- **Validate** assumptions during refactoring

## Tests in This Folder

### `lazy_loading_proof_test.go`
Proves that lazy loading works correctly in the framework:
- Services created **only when accessed** (not at registration)
- Dependencies resolved **only when service is created**
- Unused services are **never instantiated**

**Key Insights:**
- Service-level lazy: ✅ Works (via registry)
- Dependency-level lazy with `service.Cached`: ❌ Doesn't work (deps created during service creation)

### `circular_dependency_test.go`
Proves circular dependency behavior:
- Circular dependencies **always crash** (stack overflow)
- `service.Cached` **cannot prevent** circular dependency crashes
- Eager injection works fine for normal dependencies

## Why Separate Package?

These tests are in a separate `research` package (not `deploy`) because:
1. **Not production tests** - they prove concepts, not test functionality
2. **Documentation** - they serve as living documentation for design decisions
3. **Cleaner codebase** - keeps main package focused on actual unit tests
4. **Coverage isolation** - excluded from production test coverage

## Running These Tests

```bash
# Run all research tests
go test ./core/deploy/_research/

# Run specific test
go test ./core/deploy/_research/ -run TestLazyLoadingProof
```

## When to Add Tests Here

Add tests to `_research/` when:
- ✅ Proving a design decision
- ✅ Exploring edge cases during development
- ✅ Documenting non-obvious behavior
- ✅ Validating refactoring assumptions

**Do NOT add here:**
- ❌ Production unit tests (those go in main package)
- ❌ Integration tests (those go in dedicated test folders)
- ❌ Benchmarks (those go in main package with `_test.go`)

## Maintenance

These tests are **living documentation**. They should be:
- Updated when framework behavior changes
- Removed if no longer relevant
- Kept simple and focused on one concept each
