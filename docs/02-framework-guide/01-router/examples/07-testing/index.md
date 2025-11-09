# Testing Deep Dive

> **Master testing strategies for Lokstra handlers**

This example demonstrates unit testing, integration testing, and testing best practices.

## Testing Approaches

### 1. Unit Testing Handlers

Test handlers in isolation without starting the server:

```go
func TestGetUser(t *testing.T) {
    params := UserIDParam{ID: "123"}
    result := GetUser(params)
    
    // Assert response
    if result == nil {
        t.Fatal("Expected non-nil response")
    }
}
```

---

### 2. Integration Testing

Test with full HTTP request/response cycle:

```go
func TestGetUserIntegration(t *testing.T) {
    router := setupRouter()
    app := lokstra.NewApp("test", ":0", router)
    
    // Make HTTP request
    resp := makeRequest(app, "GET", "/users/123")
    
    if resp.StatusCode != 200 {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }
}
```

---

### 3. Table-Driven Tests

Test multiple scenarios efficiently:

```go
func TestGetUser_TableDriven(t *testing.T) {
    tests := []struct {
        name       string
        userID     string
        wantStatus int
        wantError  bool
    }{
        {"valid user", "123", 200, false},
        {"not found", "999", 404, true},
        {"invalid id", "abc", 400, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

---

## Testing Patterns

### Pattern 1: Direct Handler Testing

```go
// Handler
func GetUser(params UserIDParam) *response.ApiHelper {
    if params.ID == "999" {
        return response.NewApiNotFound("User not found")
    }
    return response.NewApiOk(map[string]any{"id": params.ID})
}

// Test
func TestGetUser_Success(t *testing.T) {
    result := GetUser(UserIDParam{ID: "123"})
    // Verify result
}

func TestGetUser_NotFound(t *testing.T) {
    result := GetUser(UserIDParam{ID: "999"})
    // Verify 404 response
}
```

---

### Pattern 2: Testing with Context

```go
// Handler
func GetWithAuth(c *request.Context) *response.ApiHelper {
    token := c.Req.Header("Authorization")
    if token == "" {
        return response.NewApiUnauthorized("Missing token")
    }
    return response.NewApiOk(map[string]any{"authenticated": true})
}

// Test
func TestGetWithAuth(t *testing.T) {
    // Create mock context
    ctx := createMockContext()
    ctx.Req.SetHeader("Authorization", "Bearer token123")
    
    result := GetWithAuth(ctx)
    // Verify authenticated
}
```

---

### Pattern 3: Testing Validation

```go
func TestCreateUser_Validation(t *testing.T) {
    tests := []struct {
        name    string
        request CreateUserRequest
        wantErr bool
    }{
        {
            name: "valid",
            request: CreateUserRequest{
                Email: "test@example.com",
                Age:   25,
            },
            wantErr: false,
        },
        {
            name: "invalid email",
            request: CreateUserRequest{
                Email: "invalid",
                Age:   25,
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test validation
        })
    }
}
```

---

## Mocking Dependencies

### Database Mocking

```go
type UserRepository interface {
    GetUser(id string) (*User, error)
}

type MockUserRepo struct {
    users map[string]*User
}

func (m *MockUserRepo) GetUser(id string) (*User, error) {
    if user, ok := m.users[id]; ok {
        return user, nil
    }
    return nil, errors.New("not found")
}

// In tests
func TestGetUser_WithMock(t *testing.T) {
    mockRepo := &MockUserRepo{
        users: map[string]*User{
            "123": {ID: "123", Name: "John"},
        },
    }
    
    // Use mockRepo in handler
}
```

---

## Test Helpers

### Assert Helpers

```go
func assertEqual(t *testing.T, expected, actual any) {
    t.Helper()
    if expected != actual {
        t.Errorf("Expected %v, got %v", expected, actual)
    }
}

func assertNotNil(t *testing.T, value any) {
    t.Helper()
    if value == nil {
        t.Error("Expected non-nil value")
    }
}
```

### Setup Helpers

```go
func setupTestRouter() lokstra.Router {
    router := lokstra.NewRouter("test")
    router.GET("/users/:id", GetUser)
    router.POST("/users", CreateUser)
    return router
}

func setupTestApp() *app.App {
    router := setupTestRouter()
    return lokstra.NewApp("test", ":0", router)
}
```

---

## Integration Testing

### HTTP Test Helper

```go
func makeRequest(app *app.App, method, path string, body any) *http.Response {
    // Create test server
    ts := httptest.NewServer(app.Handler())
    defer ts.Close()
    
    // Make request
    var reqBody io.Reader
    if body != nil {
        jsonData, _ := json.Marshal(body)
        reqBody = bytes.NewReader(jsonData)
    }
    
    req, _ := http.NewRequest(method, ts.URL+path, reqBody)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, _ := client.Do(req)
    
    return resp
}
```

---

## Best Practices

### ✅ Do

```go
// Use table-driven tests
func TestHandler(t *testing.T) {
    tests := []struct{
        name string
        // ...
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}

// Use t.Helper() in helpers
func assertEqual(t *testing.T, expected, actual any) {
    t.Helper() // Shows correct line number in failures
    // ...
}

// Test error cases
func TestHandler_Error(t *testing.T) {
    // Test what happens when things go wrong
}
```

### ❌ Don't

```go
// Don't skip error checks in tests
result, _ := GetUser("123") // ❌ Check errors

// Don't test implementation details
// Test behavior, not internals

// Don't use time.Sleep in tests
time.Sleep(1 * time.Second) // ❌ Use proper synchronization
```

---

## Test Coverage

### Generate Coverage Report

```bash
# Run tests with coverage
go test -cover

# Generate HTML coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Target Coverage

- **Critical paths**: 90%+ coverage
- **Business logic**: 80%+ coverage
- **Handlers**: 70%+ coverage

---

## Running Tests

```bash
# Run all tests
go test ./...

# Run specific test
go test -run TestGetUser

# Run with verbose output
go test -v

# Run with coverage
go test -cover

# Run benchmarks
go test -bench=.
```

---

## Key Takeaways

✅ **Test handlers directly** for unit tests  
✅ **Use table-driven tests** for multiple scenarios  
✅ **Mock external dependencies** (database, APIs)  
✅ **Test error cases** thoroughly  
✅ **Use test helpers** to reduce duplication  
✅ **Aim for 70-90% coverage** on critical code  
✅ **Integration tests** for critical flows

---

## Related Examples

- [02-parameter-binding](../02-parameter-binding/) - Testing validation
- [05-error-handling](../05-error-handling/) - Testing error responses
- [06-performance](../06-performance/) - Benchmarking
