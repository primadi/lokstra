---
name: lokstra-code-advanced
description: Advanced Lokstra code generation including comprehensive tests, configuration management, middleware, and consistency validation. Use after basic implementation to add testing, advanced features, and production readiness. SKILL 8-13 from original guide.
license: MIT
metadata:
  author: lokstra-framework
  version: "1.0"
  framework: lokstra
  skill-level: advanced
compatibility: Designed for GitHub Copilot, Cursor, Claude Code
---

# Lokstra Code Implementation (Advanced)

## When to use this skill

Use this skill when:
- Basic implementation done (domain, repository, handler)
- Need comprehensive unit and integration tests
- Adding middleware (auth, rate limiting, etc.)
- Production readiness checks
- Configuration management
- Consistency validation

## SKILL 8: Handler Tests

```go
// modules/{module}/handler/{entity}_handler_test.go
package handler

func TestGetByID(t *testing.T) {
    // Setup mock repository
    mockRepo := &MockRepository{
        GetByIDFunc: func(ctx context.Context, id string) (*domain.{Entity}, error) {
            return &domain.{Entity}{
                ID:   id,
                Name: "Test",
            }, nil
        },
    }
    
    handler := &{Entity}Handler{Repo: mockRepo}
    
    // Test
    result, err := handler.GetByID("test-id")
    assert.NoError(t, err)
    assert.Equal(t, "test-id", result.ID)
}
```

## SKILL 9: Repository Tests

```go
// modules/{module}/repository/{entity}_repository_test.go
package repository

func TestCreate(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    repo := &{Entity}RepositoryImpl{DB: db}
    
    entity := &domain.{Entity}{
        ID:   uuid.New().String(),
        Name: "Test",
    }
    
    err := repo.Create(context.Background(), entity)
    assert.NoError(t, err)
    
    // Verify
    found, err := repo.GetByID(context.Background(), entity.ID)
    assert.NoError(t, err)
    assert.Equal(t, entity.Name, found.Name)
}
```

## SKILL 10: Middleware

### Per-Route Middleware

```go
// @Route "DELETE /{id}", middlewares=["auth", "admin"]
func (h *{Entity}Handler) Delete(id string) error {
    return h.Repo.Delete(context.Background(), id)
}
```

### Custom Middleware

```go
// modules/shared/middleware/rate_limit.go
func RateLimit(limit int) lokstra.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Rate limiting logic
            next.ServeHTTP(w, r)
        })
    }
}
```

## SKILL 11: Config Management

### Advanced config.yaml

```yaml
configs:
  # Environment-specific
  database:
    host: ${DB_HOST:localhost}        # From env or default
    port: ${DB_PORT:5432}
    pool-size: ${DB_POOL_SIZE:50}
  
  # Feature flags
  features:
    enable-caching: true
    enable-rate-limiting: true
  
  # Service selection
  repository:
    implementation: "${REPO_IMPL:postgres-repository}"

service-definitions:
  # With custom factory
  custom-service:
    type: custom-factory
    depends-on: [db-pool, cache]
    config:
      timeout: "30s"
      retry-count: 3
```

### Config Injection

```go
// @Handler name="handler"
type Handler struct {
    // @Inject "cfg:database.pool-size"
    PoolSize int
    
    // @Inject "cfg:features.enable-caching"
    CacheEnabled bool
    
    // @Inject "cfg:@repository.implementation"
    RepoName string  // Resolved from config
}
```

## SKILL 12: Integration Tests

```go
// tests/integration/{entity}_test.go
func TestEndToEnd(t *testing.T) {
    // Start test server
    server := startTestServer(t)
    defer server.Stop()
    
    // Create entity
    resp, err := http.Post(
        server.URL+"/api/{module}",
        "application/json",
        strings.NewReader(`{"name":"Test","email":"test@example.com"}`),
    )
    assert.NoError(t, err)
    assert.Equal(t, 201, resp.StatusCode)
    
    // Get entity
    var created struct {
        Data struct {
            ID string `json:"id"`
        } `json:"data"`
    }
    json.NewDecoder(resp.Body).Decode(&created)
    
    resp, err = http.Get(server.URL + "/api/{module}/" + created.Data.ID)
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

## SKILL 13: Consistency Checks

### Validation Checklist

- [ ] All @Handler services registered in config.yaml
- [ ] All @Inject dependencies available
- [ ] All routes have proper HTTP methods
- [ ] All DTOs have validation tags
- [ ] All error responses standardized
- [ ] All database migrations tested
- [ ] All tests passing (unit + integration)
- [ ] No circular dependencies
- [ ] Config values have defaults
- [ ] Middleware properly applied

### Automated Validation

```bash
# Run all checks
lokstra autogen . --verify

# Check config consistency
go run . --check-config

# Run all tests
go test ./...

# Check coverage
go test -cover ./...
```

## Production Readiness

### Logging

```go
logger.LogInfo("[{Module}] {Entity} created: %s", entity.ID)
logger.LogError("[{Module}] Failed to create {entity}: %v", err)
```

### Error Handling

```go
if err != nil {
    logger.LogError("[{Module}] Database error: %v", err)
    return ctx.Api.InternalServerError("Failed to process request")
}
```

### Health Checks

```go
// Add health endpoint
r.GET("/health", func() string {
    // Check database connection
    // Check dependencies
    return "OK"
})
```

## Performance Optimization

1. **Database Indexes** - Ensure all foreign keys indexed
2. **Connection Pooling** - Configure min/max connections
3. **Caching** - Add Redis for frequently accessed data
4. **Rate Limiting** - Protect endpoints
5. **Load Testing** - Use k6 or similar

## Resources

- **Testing Guide:** See Lokstra documentation
- **Middleware Examples:** [references/MIDDLEWARE_EXAMPLES.md](references/MIDDLEWARE_EXAMPLES.md)
- **Performance Tips:** Lokstra performance guide
