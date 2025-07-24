# Request Package Tests

This directory contains comprehensive unit tests for the `core/request` package in the Lokstra framework.

## Test Files Overview

### 1. `basic_test.go`
Tests fundamental package compilation and type definitions:
- Package compilation verification
- HandlerFunc type testing
- HandlerRegister structure validation
- Basic type exports verification

### 2. `context_test.go`
Tests the Context struct and its core functionality:
- Context creation and initialization (`NewContext`)
- Context retrieval from request (`ContextFromRequest`)
- Path parameter extraction (`GetPathParam`)
- Query parameter extraction (`GetQueryParam`)
- Header extraction (`GetHeader`)
- Header value checking (`IsHeaderContainValue`)
- Raw body reading and caching (`GetRawBody`)
- Context cancellation behavior
- Embedded type functionality (context.Context and response.Response)

### 3. `binding_test.go`
Tests parameter binding functionality:
- Path parameter binding (`BindPath`)
- Query parameter binding (`BindQuery`)
- Header parameter binding (`BindHeader`)
- Request body binding (`BindBody`)
- Complete binding (`BindAll`)
- Map parameter binding
- Slice parameter binding
- Error handling for invalid data types
- Mixed parameter type binding

### 4. `binding_utils_test.go`
Tests the underlying utility functions for binding:
- Type conversion for different data types (string, int, bool, float)
- Slice handling with comma-separated values
- Multiple value handling for same parameter
- Error cases for unsupported types
- Invalid type conversion scenarios
- Empty value handling
- Various integer and float type support
- Boolean value parsing
- Whitespace trimming in comma-separated values

### 5. `binding_meta_test.go`
Tests binding metadata and tag parsing:
- Struct tag recognition (path, query, header)
- Binding metadata caching
- Slice field detection
- Map field detection
- Multiple tag type handling
- Untagged field ignoring
- Pointer type handling
- Empty tag value handling
- Nested struct behavior

### 6. `integration_test.go`
Comprehensive integration tests:
- Complete real-world request binding scenarios
- API endpoint simulation
- Handler function workflow testing
- Context cancellation with binding
- Multiple binding calls on same context
- Error handling scenarios
- Complex parameter combinations

## Test Coverage

The test suite covers:

### Core Functionality
- ✅ Context creation and lifecycle
- ✅ Parameter extraction (path, query, headers)
- ✅ Body reading and caching
- ✅ Context cancellation handling

### Binding System
- ✅ All binding methods (Path, Query, Header, Body, All)
- ✅ Type conversion (string, int, bool, float, slices)
- ✅ Complex types (maps, slices, custom unmarshaling)
- ✅ Error handling for invalid conversions
- ✅ Tag parsing and metadata caching

### Edge Cases
- ✅ Empty values and missing parameters
- ✅ Invalid JSON bodies
- ✅ Unsupported field types
- ✅ Multiple values for same parameter
- ✅ Whitespace handling
- ✅ Context cancellation scenarios

### Integration Scenarios
- ✅ Real-world API endpoint patterns
- ✅ Handler function workflows
- ✅ Complex parameter combinations
- ✅ Error propagation

## Test Structures

The tests use various mock structures to validate binding:

```go
// Path parameters
type PathParams struct {
    ID   string `path:"id"`
    Name string `path:"name"`
}

// Query parameters with different types
type QueryParams struct {
    Name   string   `query:"name"`
    Age    int      `query:"age"`
    Active bool     `query:"active"`
    Tags   []string `query:"tags"`
}

// Header parameters
type HeaderParams struct {
    ContentType   string   `header:"Content-Type"`
    Authorization string   `header:"Authorization"`
    CustomHeaders []string `header:"X-Custom-Header"`
}

// Body parameters
type BodyParams struct {
    Title       string `body:"title"`
    Description string `body:"description"`
    Count       int    `body:"count"`
}

// Map and indexed parameters
type MapParams struct {
    Metadata map[string]string `query:"meta"`
}
```

## Running Tests

To run all request package tests:
```bash
go test ./core/request
```

To run with verbose output:
```bash
go test -v ./core/request
```

To run specific test files:
```bash
go test -v ./core/request -run TestContext
go test -v ./core/request -run TestBinding
```

## Mock Dependencies

The tests use:
- `net/http/httptest` for HTTP request/response mocking
- Standard library `testing` package
- No external dependencies beyond the Lokstra framework

## Error Scenarios Tested

1. **Type Conversion Errors**: Invalid strings for numeric types
2. **JSON Parsing Errors**: Malformed JSON in request bodies
3. **Unsupported Types**: Complex types not supported by binding
4. **Empty Values**: Missing or empty parameter handling
5. **Context Cancellation**: Binding behavior with cancelled contexts

## Best Practices Demonstrated

1. **Comprehensive Coverage**: Tests cover all public methods and edge cases
2. **Realistic Scenarios**: Integration tests mirror real-world usage
3. **Error Handling**: Validates both success and failure paths
4. **Mock Structures**: Uses representative test data structures
5. **Clean Setup**: Each test is independent with proper setup/teardown

The test suite ensures the request package provides reliable parameter binding and context management for HTTP requests in the Lokstra framework.
