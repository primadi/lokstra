package main

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/serviceapi"
)

// This example demonstrates Lokstra's type-safe service container system.
// It shows how to create, register, and use services with compile-time type safety.
//
// Learning Objectives:
// - Understand type-safe service resolution
// - Learn service interface design patterns
// - Explore service composition and dependencies
// - See service lifecycle and error handling
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/core-concepts.md#type-safe-services

// ===== Service Interfaces =====

// Cache service interface
type CacheService interface {
	Set(key string, value any) error
	Get(key string) (any, error)
	Delete(key string) error
	Clear() error
	GetStats() map[string]any
}

// Database service interface
type DatabaseService interface {
	FindUser(id int) (User, error)
	CreateUser(user User) (User, error)
	UpdateUser(user User) error
	DeleteUser(id int) error
	GetConnection() string
}

// Metrics service interface
type MetricsService interface {
	IncrementCounter(name string)
	RecordDuration(name string, duration float64)
	GetMetrics() map[string]any
}

// ===== Domain Models =====

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ===== Service Implementations =====

// In-memory cache implementation
type InMemoryCache struct {
	data  map[string]any
	stats map[string]int
}

func NewInMemoryCache() CacheService {
	return &InMemoryCache{
		data:  make(map[string]any),
		stats: make(map[string]int),
	}
}

func (c *InMemoryCache) Set(key string, value any) error {
	c.data[key] = value
	c.stats["sets"]++
	lokstra.Logger.Debugf("Cache SET: %s", key)
	return nil
}

func (c *InMemoryCache) Get(key string) (any, error) {
	value, exists := c.data[key]
	if !exists {
		c.stats["misses"]++
		return nil, fmt.Errorf("key not found: %s", key)
	}
	c.stats["hits"]++
	lokstra.Logger.Debugf("Cache GET: %s", key)
	return value, nil
}

func (c *InMemoryCache) Delete(key string) error {
	delete(c.data, key)
	c.stats["deletes"]++
	lokstra.Logger.Debugf("Cache DELETE: %s", key)
	return nil
}

func (c *InMemoryCache) Clear() error {
	c.data = make(map[string]any)
	c.stats["clears"]++
	lokstra.Logger.Debugf("Cache CLEAR")
	return nil
}

func (c *InMemoryCache) GetStats() map[string]any {
	return map[string]any{
		"type":    "in-memory",
		"size":    len(c.data),
		"hits":    c.stats["hits"],
		"misses":  c.stats["misses"],
		"sets":    c.stats["sets"],
		"deletes": c.stats["deletes"],
		"clears":  c.stats["clears"],
	}
}

// Mock database implementation
type MockDatabase struct {
	users  map[int]User
	nextID int
}

func NewMockDatabase() DatabaseService {
	db := &MockDatabase{
		users:  make(map[int]User),
		nextID: 1,
	}

	// Pre-populate with some data
	db.users[1] = User{ID: 1, Name: "John Doe", Email: "john@example.com"}
	db.users[2] = User{ID: 2, Name: "Jane Smith", Email: "jane@example.com"}
	db.nextID = 3

	return db
}

func (db *MockDatabase) FindUser(id int) (User, error) {
	user, exists := db.users[id]
	if !exists {
		return User{}, fmt.Errorf("user not found: %d", id)
	}
	lokstra.Logger.Debugf("Database FIND USER: %d", id)
	return user, nil
}

func (db *MockDatabase) CreateUser(user User) (User, error) {
	user.ID = db.nextID
	db.nextID++
	db.users[user.ID] = user
	lokstra.Logger.Debugf("Database CREATE USER: %d", user.ID)
	return user, nil
}

func (db *MockDatabase) UpdateUser(user User) error {
	if _, exists := db.users[user.ID]; !exists {
		return fmt.Errorf("user not found: %d", user.ID)
	}
	db.users[user.ID] = user
	lokstra.Logger.Debugf("Database UPDATE USER: %d", user.ID)
	return nil
}

func (db *MockDatabase) DeleteUser(id int) error {
	if _, exists := db.users[id]; !exists {
		return fmt.Errorf("user not found: %d", id)
	}
	delete(db.users, id)
	lokstra.Logger.Debugf("Database DELETE USER: %d", id)
	return nil
}

func (db *MockDatabase) GetConnection() string {
	return "mock-database-connection"
}

// Simple metrics implementation
type SimpleMetrics struct {
	counters  map[string]int
	durations map[string][]float64
}

func NewSimpleMetrics() MetricsService {
	return &SimpleMetrics{
		counters:  make(map[string]int),
		durations: make(map[string][]float64),
	}
}

func (m *SimpleMetrics) IncrementCounter(name string) {
	m.counters[name]++
	lokstra.Logger.Debugf("Metrics COUNTER: %s = %d", name, m.counters[name])
}

func (m *SimpleMetrics) RecordDuration(name string, duration float64) {
	m.durations[name] = append(m.durations[name], duration)
	lokstra.Logger.Debugf("Metrics DURATION: %s = %.2fms", name, duration)
}

func (m *SimpleMetrics) GetMetrics() map[string]any {
	return map[string]any{
		"counters":  m.counters,
		"durations": m.durations,
	}
}

// ===== Service-Dependent Business Logic =====

// UserService depends on multiple services
type UserService struct {
	db      DatabaseService
	cache   CacheService
	metrics MetricsService
	logger  serviceapi.Logger
}

func NewUserService(regCtx lokstra.RegistrationContext) (*UserService, error) {
	// Resolve dependencies with type safety
	db, err := lokstra.GetService[DatabaseService](regCtx, "database")
	if err != nil {
		return nil, fmt.Errorf("failed to get database service: %w", err)
	}

	cache, err := lokstra.GetService[CacheService](regCtx, "cache")
	if err != nil {
		return nil, fmt.Errorf("failed to get cache service: %w", err)
	}

	metrics, err := lokstra.GetService[MetricsService](regCtx, "metrics")
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics service: %w", err)
	}

	logger, err := lokstra.GetService[serviceapi.Logger](regCtx, "logger")
	if err != nil {
		return nil, fmt.Errorf("failed to get logger service: %w", err)
	}

	lokstra.Logger.Infof("✅ UserService created with all dependencies")

	return &UserService{
		db:      db,
		cache:   cache,
		metrics: metrics,
		logger:  logger,
	}, nil
}

func (us *UserService) GetUser(id int) (User, error) {
	us.metrics.IncrementCounter("user_requests")

	// Try cache first
	cacheKey := fmt.Sprintf("user:%d", id)
	if cached, err := us.cache.Get(cacheKey); err == nil {
		us.metrics.IncrementCounter("cache_hits")
		return cached.(User), nil
	}

	// Cache miss - get from database
	us.metrics.IncrementCounter("cache_misses")
	user, err := us.db.FindUser(id)
	if err != nil {
		us.metrics.IncrementCounter("database_errors")
		return User{}, err
	}

	// Cache the result
	us.cache.Set(cacheKey, user)
	us.metrics.IncrementCounter("database_hits")

	return user, nil
}

func (us *UserService) CreateUser(name, email string) (User, error) {
	us.metrics.IncrementCounter("user_creations")

	user := User{Name: name, Email: email}
	createdUser, err := us.db.CreateUser(user)
	if err != nil {
		us.metrics.IncrementCounter("creation_errors")
		return User{}, err
	}

	// Cache the new user
	cacheKey := fmt.Sprintf("user:%d", createdUser.ID)
	us.cache.Set(cacheKey, createdUser)

	return createdUser, nil
}

func (us *UserService) UpdateUser(id int, name, email string) (User, error) {
	us.metrics.IncrementCounter("user_updates")

	user := User{ID: id, Name: name, Email: email}
	err := us.db.UpdateUser(user)
	if err != nil {
		us.metrics.IncrementCounter("update_errors")
		return User{}, err
	}

	// Update cache
	cacheKey := fmt.Sprintf("user:%d", id)
	us.cache.Set(cacheKey, user)

	return user, nil
}

func (us *UserService) DeleteUser(id int) error {
	us.metrics.IncrementCounter("user_deletions")

	err := us.db.DeleteUser(id)
	if err != nil {
		us.metrics.IncrementCounter("deletion_errors")
		return err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("user:%d", id)
	us.cache.Delete(cacheKey)

	return nil
}

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()

	// ===== Register Service Factories =====

	// Register cache service
	regCtx.RegisterServiceFactory("in-memory-cache", func(config any) (any, error) {
		lokstra.Logger.Infof("🔧 Creating In-Memory Cache Service")
		return NewInMemoryCache(), nil
	})

	// Register database service
	regCtx.RegisterServiceFactory("mock-database", func(config any) (any, error) {
		lokstra.Logger.Infof("🔧 Creating Mock Database Service")
		return NewMockDatabase(), nil
	})

	// Register metrics service
	regCtx.RegisterServiceFactory("simple-metrics", func(config any) (any, error) {
		lokstra.Logger.Infof("🔧 Creating Simple Metrics Service")
		return NewSimpleMetrics(), nil
	})

	// Register user service (depends on other services)
	regCtx.RegisterServiceFactory("user-service", func(config any) (any, error) {
		lokstra.Logger.Infof("🔧 Creating User Service with dependencies")
		return NewUserService(regCtx)
	})

	// ===== Create Service Instances =====

	// Create services in dependency order
	_, err := regCtx.CreateService("in-memory-cache", "cache", true)
	if err != nil {
		lokstra.Logger.Errorf("Failed to create cache service: %v", err)
		return
	}

	_, err = regCtx.CreateService("mock-database", "database", true)
	if err != nil {
		lokstra.Logger.Errorf("Failed to create database service: %v", err)
		return
	}

	_, err = regCtx.CreateService("simple-metrics", "metrics", true)
	if err != nil {
		lokstra.Logger.Errorf("Failed to create metrics service: %v", err)
		return
	}

	_, err = regCtx.CreateService("user-service", "user-service", true)
	if err != nil {
		lokstra.Logger.Errorf("Failed to create user service: %v", err)
		return
	}

	lokstra.Logger.Infof("✅ All services registered successfully")

	// ===== Application Routes =====

	app := lokstra.NewApp(regCtx, "type-safe-services-app", ":8080")

	// Get user endpoint
	type GetUserParams struct {
		ID int `path:"id"`
	}

	app.GET("/users/:id", func(ctx *lokstra.Context, params *GetUserParams) error {
		userService, err := lokstra.GetService[*UserService](regCtx, "user-service")
		if err != nil {
			return ctx.ErrorInternal("User service not available")
		}

		user, err := userService.GetUser(params.ID)
		if err != nil {
			return ctx.ErrorNotFound(err.Error())
		}

		return ctx.Ok(user)
	})

	// Create user endpoint
	type CreateUserRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	app.POST("/users", func(ctx *lokstra.Context, req *CreateUserRequest) error {
		userService, err := lokstra.GetService[*UserService](regCtx, "user-service")
		if err != nil {
			return ctx.ErrorInternal("User service not available")
		}

		user, err := userService.CreateUser(req.Name, req.Email)
		if err != nil {
			return ctx.ErrorInternal(err.Error())
		}

		return ctx.OkCreated(user)
	})

	// Update user endpoint
	type UpdateUserParams struct {
		ID int `path:"id"`
	}

	app.PUT("/users/:id", func(ctx *lokstra.Context, params *UpdateUserParams, req *CreateUserRequest) error {
		userService, err := lokstra.GetService[*UserService](regCtx, "user-service")
		if err != nil {
			return ctx.ErrorInternal("User service not available")
		}

		user, err := userService.UpdateUser(params.ID, req.Name, req.Email)
		if err != nil {
			if err.Error() == fmt.Sprintf("user not found: %d", params.ID) {
				return ctx.ErrorNotFound(err.Error())
			}
			return ctx.ErrorInternal(err.Error())
		}

		return ctx.OkUpdated(user)
	})

	// Delete user endpoint
	app.DELETE("/users/:id", func(ctx *lokstra.Context, params *GetUserParams) error {
		userService, err := lokstra.GetService[*UserService](regCtx, "user-service")
		if err != nil {
			return ctx.ErrorInternal("User service not available")
		}

		err = userService.DeleteUser(params.ID)
		if err != nil {
			if err.Error() == fmt.Sprintf("user not found: %d", params.ID) {
				return ctx.ErrorNotFound(err.Error())
			}
			return ctx.ErrorInternal(err.Error())
		}

		return ctx.Ok(map[string]any{
			"message": "User deleted successfully",
			"id":      params.ID,
		})
	})

	// Service status endpoint
	app.GET("/services/status", func(ctx *lokstra.Context) error {
		status := make(map[string]any)

		// Get cache stats
		if cache, err := lokstra.GetService[CacheService](regCtx, "cache"); err == nil {
			status["cache"] = cache.GetStats()
		}

		// Get metrics
		if metrics, err := lokstra.GetService[MetricsService](regCtx, "metrics"); err == nil {
			status["metrics"] = metrics.GetMetrics()
		}

		// Get database connection info
		if db, err := lokstra.GetService[DatabaseService](regCtx, "database"); err == nil {
			status["database"] = map[string]any{
				"connection": db.GetConnection(),
				"status":     "connected",
			}
		}

		return ctx.Ok(map[string]any{
			"message":  "Service status",
			"services": status,
		})
	})

	// Clear cache endpoint
	app.POST("/cache/clear", func(ctx *lokstra.Context) error {
		cache, err := lokstra.GetService[CacheService](regCtx, "cache")
		if err != nil {
			return ctx.ErrorInternal("Cache service not available")
		}

		if err := cache.Clear(); err != nil {
			return ctx.ErrorInternal("Failed to clear cache")
		}

		return ctx.Ok(map[string]any{
			"message": "Cache cleared successfully",
		})
	})

	lokstra.Logger.Infof("🚀 Type-Safe Services Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Try these endpoints:")
	lokstra.Logger.Infof("  GET    /users/1           - Get user (uses cache)")
	lokstra.Logger.Infof("  POST   /users             - Create user")
	lokstra.Logger.Infof("  PUT    /users/1           - Update user")
	lokstra.Logger.Infof("  DELETE /users/1           - Delete user")
	lokstra.Logger.Infof("  GET    /services/status   - Service status and metrics")
	lokstra.Logger.Infof("  POST   /cache/clear       - Clear cache")

	app.Start()
}

// Type-Safe Services Key Benefits:
//
// 1. Compile-Time Safety:
//    - Generic functions ensure correct service types
//    - Interface-based contracts prevent implementation coupling
//    - Type errors caught during compilation, not runtime
//
// 2. Dependency Injection:
//    - Services automatically resolve their dependencies
//    - Clear dependency graphs and initialization order
//    - Easy to mock for testing
//
// 3. Service Lifecycle:
//    - Services are singletons by default
//    - Lazy initialization when first requested
//    - Proper error handling for missing dependencies
//
// 4. Interface Segregation:
//    - Services implement focused interfaces
//    - Easy to substitute implementations
//    - Clear service contracts
//
// 5. Testability:
//    - Easy to inject mock implementations
//    - Clear separation of concerns
//    - Isolated unit testing

// Test Commands:
//
// # Get existing user
// curl http://localhost:8080/users/1
//
// # Create new user
// curl -X POST http://localhost:8080/users \
//      -H "Content-Type: application/json" \
//      -d '{"name":"Alice Johnson","email":"alice@example.com"}'
//
// # Update user
// curl -X PUT http://localhost:8080/users/1 \
//      -H "Content-Type: application/json" \
//      -d '{"name":"John Updated","email":"john.updated@example.com"}'
//
// # Delete user
// curl -X DELETE http://localhost:8080/users/2
//
// # Check service status and metrics
// curl http://localhost:8080/services/status
//
// # Clear cache
// curl -X POST http://localhost:8080/cache/clear
