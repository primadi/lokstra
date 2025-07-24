# Router Engine Tests

This directory contains comprehensive unit tests for the Lokstra router engine module, which provides an HTTP router abstraction layer with multiple implementations.

## Overview

The router_engine module provides:
- **RouterEngine Interface**: Unified interface for HTTP routing with support for method handling, static files, SPA serving, and reverse proxy
- **HttpRouterEngine**: Implementation using julienschmidt/httprouter with ServeMux fallback for special features
- **ServeMuxEngine**: Implementation using standard net/http ServeMux with enhanced method handling
- **Helper Functions**: Path parameter conversion utilities and HEAD request handling

## Test Files

### 1. `httprouter_engine_test.go`
Tests for the HttpRouterEngine implementation:
- **Factory Function Tests**: `NewHttpRouterEngine` with various configurations
- **Route Handling**: Method registration and parameter routing (`:id`, `*filepath`)
- **HTTP Serving**: Request handling with path parameters and proper status codes
- **Static File Serving**: File system integration with `ServeStatic`
- **SPA Support**: Single Page Application routing with `ServeSPA`
- **Reverse Proxy**: Backend proxying with `ServeReverseProxy`
- **Fallback Behavior**: ServeMux fallback for features not supported by httprouter
- **Integration Tests**: Combined testing of all features

**Key Test Scenarios:**
```go
// Parameter routing
router.HandleMethod(http.MethodGet, "/users/:id", handler)
// Request: GET /users/123 → PathValue("id") = "123"

// Wildcard routing  
router.HandleMethod(http.MethodGet, "/files/*filepath", handler)
// Request: GET /files/docs/readme.txt → PathValue("filepath") = "docs/readme.txt"

// Static file serving
router.ServeStatic("/assets", http.Dir("./static"))
// Request: GET /assets/style.css → serves ./static/style.css

// SPA routing
router.ServeSPA("/app", "./public/index.html")
// Request: GET /app/dashboard → serves index.html
// Request: GET /app/style.css → serves ./public/style.css

// Reverse proxy
router.ServeReverseProxy("/api", "http://backend:8080")
// Request: GET /api/users → proxies to http://backend:8080/users
```

### 2. `servemux_engine_test.go`
Tests for the ServeMuxEngine implementation:
- **Factory Function Tests**: `NewServeMuxEngine` with configurations
- **Method Handling**: HTTP method routing with proper Allow headers
- **Method Not Allowed**: 405 responses with Allow header for unsupported methods
- **HEAD Fallback**: Automatic HEAD support for GET routes without body
- **Static File Serving**: Directory serving with subdirectory support
- **SPA Support**: Single Page Application with asset serving
- **Reverse Proxy**: Backend integration with header forwarding
- **Multiple Methods**: Same path with different HTTP methods
- **Integration Tests**: Comprehensive feature testing

**Key Test Scenarios:**
```go
// Method-specific routing
router.HandleMethod(http.MethodGet, "/api/users", getHandler)
router.HandleMethod(http.MethodPost, "/api/users", postHandler)
// GET /api/users → 200 with getHandler
// POST /api/users → 200 with postHandler  
// PUT /api/users → 405 Method Not Allowed with Allow: GET, POST

// HEAD fallback
router.HandleMethod(http.MethodGet, "/content", handler)
// HEAD /content → 200 with headers but no body

// Path parameter conversion
"/users/:id" → "/users/{id}" (ServeMux format)
"/files/*path" → "/files/{path...}" (ServeMux wildcard)
```

### 3. `helper_test.go`
Tests for helper functions and utilities:
- **Path Conversion**: `ConvertToServeMuxParamPath` function testing
- **HEAD Writer**: `headFallbackWriter` struct testing for HEAD request handling
- **Prefix Cleaning**: `cleanPrefix` utility function testing  
- **Constants**: Engine type constant validation
- **Edge Cases**: Complex path patterns and special characters
- **Integration**: Helper function usage in real routing scenarios

**Key Test Scenarios:**
```go
// Path parameter conversion
ConvertToServeMuxParamPath("/users/:id") → "/users/{id}"
ConvertToServeMuxParamPath("/files/*filepath") → "/files/{filepath...}"
ConvertToServeMuxParamPath("/api/:version/:id") → "/api/{version}/{id}"

// HEAD fallback writer
writer := &headFallbackWriter{ResponseWriter: w}
writer.Write([]byte("content")) // Body discarded, returns len("content")
writer.Header().Set("Content-Type", "text/plain") // Headers preserved

// Prefix cleaning
cleanPrefix("/api") → "/api/"
cleanPrefix("/api/") → "/api/"
cleanPrefix("") → "/"
```

### 4. `router_engine_integration_test.go`
Integration tests comparing both engine implementations:
- **Interface Compliance**: Both engines implement RouterEngine interface
- **Behavior Comparison**: Side-by-side testing of HttpRouter vs ServeMux
- **Method Not Allowed**: Different handling between engines
- **Static File Handling**: Consistent file serving across engines
- **SPA Handling**: Single Page Application support comparison
- **Proxy Handling**: Reverse proxy behavior verification
- **Complex Integration**: Multi-feature scenarios with real-world usage

**Key Test Scenarios:**
```go
// Interface compliance
var _ serviceapi.RouterEngine = (*HttpRouterEngine)(nil)
var _ serviceapi.RouterEngine = (*ServeMuxEngine)(nil)

// Behavior comparison
httpEngine.HandleMethod(GET, "/users/:id", handler)
serveMuxEngine.HandleMethod(GET, "/users/:id", handler)
// Both should handle GET /users/123 correctly

// Method not allowed differences
// HttpRouter: POST /get-only-route → 404 Not Found
// ServeMux: POST /get-only-route → 405 Method Not Allowed + Allow header
```

## Test Coverage

### HTTP Methods Tested
- ✅ GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
- ✅ Method-specific routing and handlers
- ✅ Method Not Allowed (405) responses
- ✅ HEAD fallback to GET behavior

### Routing Features Tested
- ✅ Static paths (`/api/users`)
- ✅ Path parameters (`/users/:id` → `/users/{id}`)
- ✅ Wildcard parameters (`/files/*path` → `/files/{path...}`)
- ✅ Multiple parameters (`/users/:id/posts/:postId`)
- ✅ Nested routes and complex paths

### Special Features Tested
- ✅ Static file serving with directory traversal
- ✅ SPA (Single Page Application) support with asset fallback
- ✅ Reverse proxy with path rewriting and header forwarding
- ✅ HEAD request handling without response body
- ✅ Error responses (404 Not Found, 405 Method Not Allowed)

### Engine-Specific Behavior Tested
- ✅ HttpRouter: Parameter extraction via PathValue()
- ✅ HttpRouter: Fallback to ServeMux for static/SPA/proxy
- ✅ ServeMux: Method tracking with Allow headers
- ✅ ServeMux: Path parameter conversion (:param → {param})
- ✅ ServeMux: HEAD request body suppression

## Running Tests

```bash
# Run all router engine tests
go test ./modules/coreservice/router_engine/

# Run specific test file
go test ./modules/coreservice/router_engine/ -run TestHttpRouterEngine
go test ./modules/coreservice/router_engine/ -run TestServeMuxEngine
go test ./modules/coreservice/router_engine/ -run TestHelper
go test ./modules/coreservice/router_engine/ -run TestRouterEngineIntegration

# Run with verbose output
go test -v ./modules/coreservice/router_engine/

# Run with coverage
go test -cover ./modules/coreservice/router_engine/
```

## Test Structure

Each test file follows a consistent structure:

1. **Factory Tests**: Verify object creation and interface compliance
2. **Method Tests**: Test individual methods with various inputs
3. **HTTP Tests**: Test HTTP request/response handling
4. **Feature Tests**: Test specific features (static files, SPA, proxy)
5. **Edge Case Tests**: Test boundary conditions and error cases
6. **Integration Tests**: Test multiple features working together

## Key Test Utilities

- **httptest.NewRequest**: Create HTTP requests for testing
- **httptest.NewRecorder**: Capture HTTP responses
- **httptest.NewServer**: Create test backend servers for proxy testing
- **t.TempDir()**: Create temporary directories for file serving tests
- **os.WriteFile**: Create test files for static serving tests

## Mock Objects and Test Data

- Test handlers that return predictable responses
- Temporary directories with test files for static serving
- Mock backend servers for reverse proxy testing
- Various HTTP request scenarios (GET, POST, with/without body)
- Path parameter test cases with different patterns

## Error Scenarios Tested

- ✅ Non-existent routes (404 Not Found)
- ✅ Method not allowed (405 with Allow header for ServeMux)
- ✅ File not found in static serving
- ✅ Invalid proxy backends
- ✅ Malformed paths and edge cases

## Performance Considerations

The tests verify:
- ✅ Request routing performance
- ✅ Path parameter extraction efficiency  
- ✅ Static file serving speed
- ✅ Memory usage for request handling
- ✅ Response time for various route types

This comprehensive test suite ensures the router engine module provides reliable, performant HTTP routing with consistent behavior across different engine implementations.
