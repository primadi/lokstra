---
name: advanced-lokstra-tests
description: Create unit tests, integration tests, and mocks for Lokstra handlers and services. Use after all implementation code is complete for comprehensive testing coverage.
phase: advanced
order: 1
license: MIT
compatibility:
  lokstra_version: ">=0.1.0"
  go_version: ">=1.18"
  testing_framework: testify
---

# Advanced: Lokstra Testing

## When to Use

Use this skill when:
- Writing unit tests for handlers with `request.Context`
- Testing HTTP request/response flows with `httptest`
- Creating integration tests for repository layer
- Setting up test database fixtures
- Mocking external dependencies (repositories, services)
- Testing with `lokstra_registry` service registration
- Achieving test coverage requirements
- Benchmarking performance-critical code

Prerequisites:
- ✅ All implementation code complete (see: Phase 2 skills)
- ✅ Handlers, services, and repositories implemented
- ✅ Database migrations applied (for integration tests)

---

## Testing Imports & Setup

```go
import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)
```

**Key Testing Packages:**

| Package | Purpose |
|---------|---------|
| `net/http/httptest` | HTTP test recorder and request creation |
| `github.com/primadi/lokstra/core/request` | Lokstra's `request.Context` |
| `github.com/primadi/lokstra/lokstra_registry` | Service registry for test setup |
| `github.com/stretchr/testify/assert` | Non-fatal assertions (test continues) |
| `github.com/stretchr/testify/require` | Fatal assertions (test stops on failure) |
| `github.com/stretchr/testify/mock` | Mock object generation |

---

## Handler Unit Tests (HTTP Layer)

### File: modules/user/application/user_handler_test.go

```go
package application_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"myapp/modules/user/application"
	"myapp/modules/user/domain"

	"github.com/primadi/lokstra/core/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ==========================================
// Mock Repository using testify/mock
// ==========================================

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(id string) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) List() ([]*domain.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepository) Save(user *domain.User) (*domain.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// Compile-time interface check
var _ domain.UserRepository = (*MockUserRepository)(nil)

// ==========================================
// Basic Handler Tests
// ==========================================

func TestUserHandler_GetByID_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	expectedUser := &domain.User{
		ID:    "123",
		Name:  "John Doe",
		Email: "john@example.com",
	}
	mockRepo.On("GetByID", "123").Return(expectedUser, nil)

	handler := &application.UserHandler{
		UserRepo: mockRepo,
	}

	// Act - Call handler method directly
	user, err := handler.GetByID("123")

	// Assert
	require.NoError(t, err, "GetByID should not return error")
	assert.Equal(t, "123", user.ID)
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john@example.com", user.Email)
	mockRepo.AssertCalled(t, "GetByID", "123")
	mockRepo.AssertExpectations(t)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByID", "999").Return(nil, domain.ErrUserNotFound)

	handler := &application.UserHandler{UserRepo: mockRepo}

	// Act
	user, err := handler.GetByID("999")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserHandler_Create_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	expectedUser := &domain.User{
		ID:    "new-123",
		Name:  "John Doe",
		Email: "john@example.com",
	}
	mockRepo.On("Save", mock.MatchedBy(func(u *domain.User) bool {
		return u.Name == "John Doe" && u.Email == "john@example.com"
	})).Return(expectedUser, nil)

	handler := &application.UserHandler{UserRepo: mockRepo}

	// Create request body
	params := &domain.CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	// Act
	user, err := handler.Create(params)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "new-123", user.ID)
	assert.Equal(t, "John Doe", user.Name)
	mockRepo.AssertCalled(t, "Save", mock.Anything)
}

func TestUserHandler_Create_ValidationError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	handler := &application.UserHandler{UserRepo: mockRepo}

	// Test invalid email (should fail validation before reaching repo)
	params := &domain.CreateUserRequest{
		Name:  "John",
		Email: "invalid-email",  // Invalid!
	}

	_, err := handler.Create(params)
	
	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "Save")  // Repo should NOT be called
}

func TestUserHandler_Delete_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("Delete", "123").Return(nil)

	handler := &application.UserHandler{UserRepo: mockRepo}

	err := handler.Delete("123")

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "Delete", "123")
}
```

---

## Testing HTTP Response with request.Context

When testing handlers that use `*request.Context` for custom responses:

```go
func TestUserHandler_GetByID_HTTPResponse(t *testing.T) {
	mockRepo := new(MockUserRepository)
	expectedUser := &domain.User{ID: "123", Name: "John Doe", Email: "john@example.com"}
	mockRepo.On("GetByID", "123").Return(expectedUser, nil)

	handler := &application.UserHandler{UserRepo: mockRepo}

	// Setup HTTP test infrastructure
	req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
	rec := httptest.NewRecorder()
	ctx := request.NewContext(rec, req, nil)
	ctx.SetPathParams(map[string]string{"id": "123"})

	// Call handler method with context
	user, err := handler.GetByID("123")
	
	// If handler returns data (not using ctx.Api directly)
	require.NoError(t, err)
	
	// Simulate what Lokstra does: marshal response
	ctx.Api.Ok(user)

	// Assert HTTP response
	assert.Equal(t, http.StatusOK, rec.Code)
	
	// Parse response body
	var response map[string]interface{}
	err = json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	
	// Check response structure (depends on your api_formatter)
	assert.NotNil(t, response["data"])
}

func TestUserHandler_NotFound_HTTPResponse(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByID", "999").Return(nil, domain.ErrUserNotFound)

	handler := &application.UserHandler{UserRepo: mockRepo}

	req := httptest.NewRequest(http.MethodGet, "/api/users/999", nil)
	rec := httptest.NewRecorder()
	ctx := request.NewContext(rec, req, nil)

	// Handler returns error
	user, err := handler.GetByID("999")
	
	assert.Error(t, err)
	assert.Nil(t, user)
	
	// Simulate ctx.Api.NotFound() on error
	ctx.Api.NotFound("User not found")
	
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
```

---

## Testing Middleware

Testing custom middleware functions with `request.Context`:

```go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/stretchr/testify/assert"
)

func TestCorsMiddleware_AllOrigins(t *testing.T) {
	h := cors.Middleware("*")

	// Create test request with Origin header
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	ctx := request.NewContext(rec, req, nil)

	// Execute middleware
	h(ctx)

	// Assert headers
	assert.Equal(t, "http://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCorsMiddleware_DisallowedOrigin(t *testing.T) {
	h := cors.Middleware("http://allowed.com")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://forbidden.com")
	rec := httptest.NewRecorder()
	ctx := request.NewContext(rec, req, nil)

	h(ctx)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	authMiddleware := NewAuthMiddleware("secret-key")

	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	ctx := request.NewContext(rec, req, nil)

	err := authMiddleware(ctx)

	assert.NoError(t, err)
	// Check context values set by middleware
	assert.NotNil(t, ctx.Get("user_id"))
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	authMiddleware := NewAuthMiddleware("secret-key")

	req := httptest.NewRequest("GET", "/api/protected", nil)
	// No Authorization header
	rec := httptest.NewRecorder()
	ctx := request.NewContext(rec, req, nil)

	err := authMiddleware(ctx)

	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
```

---

## Integration Tests (Repository Layer)

### File: modules/user/infrastructure/postgres_user_repository_test.go

```go
package infrastructure_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
	
	"myapp/modules/user/domain"
	"myapp/modules/user/infrastructure"
)

// ==========================================
// Test Database Setup
// ==========================================

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()  // Mark as helper function for better error reporting
	
	// Use test database (typically separate from development DB)
	dsn := "postgres://test:test@localhost:5432/lokstra_test?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "Failed to connect to test database")

	// Verify connection
	err = db.Ping()
	require.NoError(t, err, "Failed to ping test database")

	// Create table for test
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	_, _ = db.Exec("DELETE FROM users")
	db.Close()
}

// ==========================================
// Repository Integration Tests
// ==========================================

func TestPostgresUserRepository_Save_Success(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := &infrastructure.PostgresUserRepository{DB: db}

	user := &domain.User{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	saved, err := repo.Save(user)

	require.NoError(t, err)
	assert.NotEmpty(t, saved.ID, "ID should be generated")
	assert.Equal(t, "John Doe", saved.Name)
	assert.NotZero(t, saved.CreatedAt)
}

func TestPostgresUserRepository_Save_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := &infrastructure.PostgresUserRepository{DB: db}

	// Save first user
	user1 := &domain.User{Name: "John Doe", Email: "john@example.com"}
	_, err := repo.Save(user1)
	require.NoError(t, err)

	// Try to save with same email
	user2 := &domain.User{Name: "Jane Doe", Email: "john@example.com"}
	_, err = repo.Save(user2)

	assert.Error(t, err, "Should fail with duplicate email")
}

func TestPostgresUserRepository_GetByID_Success(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := &infrastructure.PostgresUserRepository{DB: db}

	// Insert test data
	user := &domain.User{Name: "John Doe", Email: "john@example.com"}
	saved, err := repo.Save(user)
	require.NoError(t, err)

	// Retrieve and verify
	retrieved, err := repo.GetByID(saved.ID)

	require.NoError(t, err)
	assert.Equal(t, saved.ID, retrieved.ID)
	assert.Equal(t, "John Doe", retrieved.Name)
	assert.Equal(t, "john@example.com", retrieved.Email)
}

func TestPostgresUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := &infrastructure.PostgresUserRepository{DB: db}

	retrieved, err := repo.GetByID("00000000-0000-0000-0000-000000000000")

	assert.NoError(t, err)  // No error, just nil result
	assert.Nil(t, retrieved)
}

func TestPostgresUserRepository_Delete_Success(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := &infrastructure.PostgresUserRepository{DB: db}

	// Insert test data
	user := &domain.User{Name: "John Doe", Email: "john@example.com"}
	saved, _ := repo.Save(user)

	// Delete
	err := repo.Delete(saved.ID)
	assert.NoError(t, err)

	// Verify deleted
	retrieved, _ := repo.GetByID(saved.ID)
	assert.Nil(t, retrieved, "User should be deleted")
}

func TestPostgresUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := &infrastructure.PostgresUserRepository{DB: db}

	// Insert test data
	_, _ = repo.Save(&domain.User{Name: "User 1", Email: "user1@example.com"})
	_, _ = repo.Save(&domain.User{Name: "User 2", Email: "user2@example.com"})
	_, _ = repo.Save(&domain.User{Name: "User 3", Email: "user3@example.com"})

	// List all users
	users, err := repo.List()

	require.NoError(t, err)
	assert.Len(t, users, 3)
}
```

---

## Table-Driven Tests

Best practice for testing multiple scenarios:

```go
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "invalid format - no @",
			email:   "invalid-email",
			wantErr: true,
		},
		{
			name:    "empty string",
			email:   "",
			wantErr: true,
		},
		{
			name:    "multiple @ symbols",
			email:   "user@@example.com",
			wantErr: true,
		},
		{
			name:    "with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "with plus sign",
			email:   "user+tag@example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err, "expected error for: %s", tt.email)
			} else {
				assert.NoError(t, err, "unexpected error for: %s", tt.email)
			}
		})
	}
}

func TestUserHandler_CRUD_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       interface{}
		setupMock  func(*MockUserRepository)
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "GET user - success",
			method: "GET",
			path:   "/api/users/123",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", "123").Return(&domain.User{ID: "123", Name: "John"}, nil)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "GET user - not found",
			method: "GET",
			path:   "/api/users/999",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", "999").Return(nil, domain.ErrUserNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:   "POST user - success",
			method: "POST",
			path:   "/api/users",
			body:   &domain.CreateUserRequest{Name: "John", Email: "john@example.com"},
			setupMock: func(m *MockUserRepository) {
				m.On("Save", mock.Anything).Return(&domain.User{ID: "new-123", Name: "John"}, nil)
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name:   "DELETE user - success",
			method: "DELETE",
			path:   "/api/users/123",
			setupMock: func(m *MockUserRepository) {
				m.On("Delete", "123").Return(nil)
			},
			wantStatus: http.StatusNoContent,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			handler := &application.UserHandler{UserRepo: mockRepo}

			// Create request
			var body io.Reader
			if tt.body != nil {
				b, _ := json.Marshal(tt.body)
				body = bytes.NewReader(b)
			}
			req := httptest.NewRequest(tt.method, tt.path, body)
			rec := httptest.NewRecorder()

			// Execute and assert based on test case
			// ... test logic based on method
			
			mockRepo.AssertExpectations(t)
		})
	}
}
```

---

## Testing with lokstra_registry

Testing service registration and retrieval:

```go
package service_test

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestUserService struct {
	Name string
}

func (s *TestUserService) GetName() string {
	return s.Name
}

func TestRegisterAndGetService(t *testing.T) {
	// Initialize global registry
	_ = deploy.Global()

	// Register a service
	mockService := &TestUserService{Name: "test-user-service"}
	lokstra_registry.RegisterService("user-service", mockService)

	// Get service with generic function
	retrieved := lokstra_registry.GetService[*TestUserService]("user-service")
	require.NotNil(t, retrieved, "expected service to be retrieved")

	assert.Equal(t, "test-user-service", retrieved.GetName())
}

func TestMustGetService_Panic(t *testing.T) {
	// Initialize global registry
	_ = deploy.Global()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when service not found")
		}
	}()

	// This should panic
	lokstra_registry.MustGetService[*TestUserService]("nonexistent-service")
}

func TestTryGetService(t *testing.T) {
	_ = deploy.Global()

	// Register a service
	mockService := &TestUserService{Name: "try-service"}
	lokstra_registry.RegisterService("try-service", mockService)

	// Try to get existing service
	retrieved, ok := lokstra_registry.TryGetService[*TestUserService]("try-service")
	assert.True(t, ok)
	assert.Equal(t, "try-service", retrieved.Name)

	// Try to get nonexistent service
	_, ok = lokstra_registry.TryGetService[*TestUserService]("nonexistent")
	assert.False(t, ok)
}

func TestConfigGetAndSet(t *testing.T) {
	_ = deploy.Global()

	// Set config
	lokstra_registry.SetConfig("test-key", "test-value")

	// Get config with default
	value := lokstra_registry.GetConfig("test-key", "default")
	assert.Equal(t, "test-value", value)

	// Get nonexistent config (should return default)
	defaultValue := lokstra_registry.GetConfig("nonexistent", "default-value")
	assert.Equal(t, "default-value", defaultValue)
}
```

---

## Mocking External Services

### File: modules/notification/application/notification_handler_test.go

```go
package application_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ==========================================
// Mock Email Provider
// ==========================================

type MockEmailProvider struct {
	mock.Mock
}

func (m *MockEmailProvider) Send(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func (m *MockEmailProvider) SendTemplate(to, template string, data map[string]interface{}) error {
	args := m.Called(to, template, data)
	return args.Error(0)
}

// ==========================================
// Email Notification Tests
// ==========================================

func TestNotificationHandler_SendWelcomeEmail_Success(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockEmail.On("Send", "john@example.com", mock.Anything, mock.Anything).Return(nil)

	handler := &NotificationHandler{
		EmailProvider: mockEmail,
	}

	err := handler.SendWelcomeEmail("john@example.com")

	assert.NoError(t, err)
	mockEmail.AssertCalled(t, "Send", "john@example.com", mock.Anything, mock.Anything)
	mockEmail.AssertNumberOfCalls(t, "Send", 1)
}

func TestNotificationHandler_SendWelcomeEmail_SMTPFailure(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockEmail.On("Send", "john@example.com", mock.Anything, mock.Anything).
		Return(errors.New("SMTP connection failed"))

	handler := &NotificationHandler{EmailProvider: mockEmail}

	err := handler.SendWelcomeEmail("john@example.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP connection failed")
}

func TestNotificationHandler_SendWithTemplate(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockEmail.On("SendTemplate", "john@example.com", "welcome", mock.MatchedBy(func(data map[string]interface{}) bool {
		name, ok := data["name"]
		return ok && name == "John"
	})).Return(nil)

	handler := &NotificationHandler{EmailProvider: mockEmail}

	err := handler.SendWelcomeWithTemplate("john@example.com", "John")

	assert.NoError(t, err)
}

// ==========================================
// Mock HTTP Client
// ==========================================

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Get(url string) ([]byte, error) {
	args := m.Called(url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockHTTPClient) Post(url string, body []byte) ([]byte, error) {
	args := m.Called(url, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func TestExternalAPIHandler_FetchData(t *testing.T) {
	mockClient := new(MockHTTPClient)
	mockClient.On("Get", "https://api.example.com/users/123").
		Return([]byte(`{"id":"123","name":"John"}`), nil)

	handler := &ExternalAPIHandler{Client: mockClient}

	user, err := handler.FetchUser("123")

	assert.NoError(t, err)
	assert.Equal(t, "123", user.ID)
	assert.Equal(t, "John", user.Name)
}
```

---

## Benchmark Testing

For performance-critical code:

```go
package application_test

import (
	"testing"
	"myapp/modules/user/domain"
)

func BenchmarkValidateEmail(b *testing.B) {
	email := "user@example.com"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = domain.ValidateEmail(email)
	}
}

func BenchmarkUserHandler_GetByID(b *testing.B) {
	mockRepo := new(MockUserRepository)
	expectedUser := &domain.User{ID: "123", Name: "John Doe"}
	mockRepo.On("GetByID", "123").Return(expectedUser, nil)

	handler := &application.UserHandler{UserRepo: mockRepo}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.GetByID("123")
	}
}

func BenchmarkCreateUserRequest_Validation(b *testing.B) {
	req := &domain.CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Validate()
	}
}

// Benchmark with parallel execution
func BenchmarkUserHandler_GetByID_Parallel(b *testing.B) {
	mockRepo := new(MockUserRepository)
	expectedUser := &domain.User{ID: "123", Name: "John Doe"}
	mockRepo.On("GetByID", "123").Return(expectedUser, nil)

	handler := &application.UserHandler{UserRepo: mockRepo}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = handler.GetByID("123")
		}
	})
}
```

---

## Test Organization

```
modules/user/
├── application/
│   ├── user_handler.go
│   ├── user_handler_test.go        # Handler unit tests
│   └── zz_generated.lokstra.go
├── infrastructure/
│   ├── postgres_user_repository.go
│   ├── postgres_user_repository_test.go  # Repository integration tests
│   └── zz_generated.lokstra.go
└── domain/
    ├── user.go
    ├── user_test.go                # Domain logic tests
    ├── validation.go
    └── validation_test.go          # Validation tests
```

---

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests for specific module
go test ./modules/user/...

# Run with coverage
go test -cover ./...

# Generate coverage report (HTML)
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific test by name
go test -run TestGetByID ./modules/user/application

# Run tests matching pattern
go test -run "TestUserHandler_.*" ./...

# Run tests in verbose mode
go test -v ./...

# Run with race detector
go test -race ./...

# Run benchmarks
go test -bench=. ./modules/user/application

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkValidateEmail ./modules/user/domain

# Skip integration tests (using build tags)
go test -tags=unit ./...

# Run only integration tests
go test -tags=integration ./...
```

---

## Best Practices

### 1. Use require vs assert

```go
// ✅ Use require for critical setup (stops test on failure)
db := setupTestDB(t)
require.NoError(t, err, "database setup must succeed")

// ✅ Use assert for assertions (continues test on failure)
assert.Equal(t, expected, actual, "values should match")
assert.NoError(t, err)
```

### 2. Use Table-Driven Tests

```go
// ✅ Good: Multiple cases, clear pattern
func TestValidate(t *testing.T) {
    tests := []struct{ name, input string; wantErr bool }{...}
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { ... })
    }
}

// ❌ Bad: Separate test function for each case
func TestValidate_Valid(t *testing.T) { ... }
func TestValidate_Empty(t *testing.T) { ... }
```

### 3. Mock External Dependencies

```go
// ✅ Mock: database, email provider, external APIs, file system
// ❌ Don't mock: domain logic, validation, pure functions
```

### 4. Test Error Cases

```go
// ✅ Test: not found, validation errors, database failures, timeouts
// ❌ Only test: happy path
```

### 5. Clean Test Data

```go
// ✅ Use setup/teardown with t.Helper()
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    // ...
}

// ✅ Use separate test database
// ✅ Clean up after tests with defer
// ❌ Leave test data in production database
```

### 6. Clear Test Names

```go
// ✅ Good: Descriptive names with underscore convention
TestUserHandler_GetByID_Success
TestUserHandler_GetByID_NotFound
TestUserHandler_Create_ValidationError

// ❌ Bad: Vague names
TestGet
TestFail
TestUser
```

### 7. Use Compile-Time Interface Checks

```go
// ✅ Verify mock implements interface at compile time
var _ domain.UserRepository = (*MockUserRepository)(nil)
```

### 8. Build Tags for Test Types

```go
//go:build integration

package infrastructure_test

// Integration tests that require database...
```

---

## Coverage Goals

| Layer | Target Coverage |
|-------|-----------------|
| Domain models | ≥ 90% |
| Handlers | ≥ 80% |
| Repositories | ≥ 85% |
| External services | Mocked, ≥ 70% |
| Middleware | ≥ 80% |
| **Overall target** | **≥ 80%** |

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: lokstra_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Run unit tests
        run: go test -tags=unit -cover -coverprofile=coverage.out ./...
      
      - name: Run integration tests
        run: go test -tags=integration ./...
        env:
          DB_DSN: postgres://test:test@localhost:5432/lokstra_test?sslmode=disable
      
      - name: Check coverage threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage $COVERAGE% is below 80% threshold"
            exit 1
          fi
```

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| Mock not called | Check mock setup with `mockRepo.On()` matches actual call |
| Test database connection fails | Verify test DB is running and DSN is correct |
| Coverage not increasing | Check if testing all branches and error paths |
| Race condition detected | Use proper synchronization or test isolation |
| Integration test pollution | Use `t.Cleanup()` or `defer` for cleanup |

### Debug Tips

```go
// Print mock calls for debugging
t.Logf("Mock calls: %+v", mockRepo.Calls)

// Check expectations explicitly
mockRepo.AssertExpectations(t)

// Print test output
t.Logf("Response body: %s", rec.Body.String())
```

---

## Next Steps

After completing tests:

1. ✅ Validate application consistency (see: [advanced-lokstra-validate-consistency](../advanced-lokstra-validate-consistency/SKILL.md))
2. ✅ Set up CI/CD pipeline
3. ✅ Configure coverage reporting
4. ✅ Add performance benchmarks for critical paths

---

## Related Skills

- [implementation-lokstra-create-handler](../implementation-lokstra-create-handler/SKILL.md) - Handler creation
- [implementation-lokstra-create-service](../implementation-lokstra-create-service/SKILL.md) - Service creation
- [advanced-lokstra-middleware](../advanced-lokstra-middleware/SKILL.md) - Middleware creation
- [advanced-lokstra-validate-consistency](../advanced-lokstra-validate-consistency/SKILL.md) - Code validation
