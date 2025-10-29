# Service Deep Dive - Examples

This folder contains advanced service patterns, DI, remote services, and architecture examples.

## Examples

### 01 - Service Factories ✅
Custom service initialization and factory patterns.

**Topics**: Factories, initialization, lifecycle

**Files**: main.go, index.md, test.http

### 02 - Remote Services ✅
HTTP-based service communication and configuration.

**Topics**: Remote services, HTTP client, ClientRouter

**Files**: main.go, index.md, test.http

### 03 - Auto-Router Advanced ✅
Complex auto-router scenarios and conventions.

**Topics**: Auto-router, conventions, parameter mapping

**Files**: main.go, index.md, test.http

### 04 - Service Composition ✅
Layered services, decorators, and composition patterns.

**Topics**: Layering, decorators, composition

**Files**: main.go, index.md, test.http

### 05 - Service Architecture ✅
Domain-driven design and clean architecture patterns.

**Topics**: DDD, clean architecture, boundaries

**Files**: main.go, index.md, test.http

### 06 - Testing ✅
Mock services, integration testing, contract testing.

**Topics**: Mocking, integration tests, contracts

**Files**: main.go, index.md, test.http

### 07 - Performance ✅
Benchmarks comparing service access methods.

**Topics**: LazyLoad, GetService, benchmarking

**Files**: main.go, index.md, test.http

### 08 - Migration Pattern ✅
Monolith to microservices migration strategy.

**Topics**: Migration, monolith, microservices

**Files**: main.go, index.md, test.http

---

## Running Examples

Each example follows this structure:
```
01-service-factories/
├── main.go              # Working code
├── index.md             # Detailed explanation
└── test.http            # HTTP test requests
```

To run an example:
```bash
cd 01-service-factories
go run main.go

# Test with curl or test.http file
curl http://localhost:3000/endpoint
```

---

**Status**: ✅ All 8 examples complete
