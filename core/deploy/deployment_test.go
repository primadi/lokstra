package deploy

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/primadi/lokstra/core/service"
)

// Mock service types for testing
type MockDB struct {
	DSN      string
	MaxConns int
}

type MockLogger struct {
	Level string
}

type MockUserService struct {
	DB     *service.Cached[*MockDB]     // Lazy-loaded
	Logger *service.Cached[*MockLogger] // Lazy-loaded
}

type MockOrderService struct {
	DB          *service.Cached[*MockDB]          // Lazy-loaded
	UserService *service.Cached[*MockUserService] // Lazy-loaded
	Logger      *service.Cached[*MockLogger]      // Lazy-loaded
}

// Mock factories
func mockDBFactory(deps map[string]any, config map[string]any) any {
	return &MockDB{
		DSN:      config["dsn"].(string),
		MaxConns: config["max-conns"].(int),
	}
}

func mockLoggerFactory(deps map[string]any, config map[string]any) any {
	return &MockLogger{
		Level: config["level"].(string),
	}
}

func mockUserServiceFactory(deps map[string]any, config map[string]any) any {
	return &MockUserService{
		DB:     service.Cast[*MockDB](deps["db"]),
		Logger: service.Cast[*MockLogger](deps["logger"]),
	}
}

func mockOrderServiceFactory(deps map[string]any, config map[string]any) any {
	return &MockOrderService{
		DB:          service.Cast[*MockDB](deps["dbOrder"]),
		UserService: service.Cast[*MockUserService](deps["userSvc"]),
		Logger:      service.Cast[*MockLogger](deps["logger"]),
	}
}

func setupTestRegistry() *GlobalRegistry {
	reg := NewGlobalRegistry()

	// Register factories
	reg.RegisterServiceType("mock-db", mockDBFactory, nil)
	reg.RegisterServiceType("mock-logger", mockLoggerFactory, nil)
	reg.RegisterServiceType("mock-user-service", mockUserServiceFactory, nil)
	reg.RegisterServiceType("mock-order-service", mockOrderServiceFactory, nil)

	// Define configs
	reg.DefineConfig(&schema.ConfigDef{
		Name:  "DB_MAX_CONNS",
		Value: 20,
	})

	reg.DefineConfig(&schema.ConfigDef{
		Name:  "LOG_LEVEL",
		Value: "info",
	})

	reg.DefineConfig(&schema.ConfigDef{
		Name:  "DB_DSN",
		Value: "postgres://localhost/testdb",
	})

	// Resolve configs
	if err := reg.ResolveConfigs(); err != nil {
		panic(err)
	}

	// Define services
	reg.DefineService(&schema.ServiceDef{
		Name: "db",
		Type: "mock-db",
		Config: map[string]any{
			"dsn":       "${@cfg:DB_DSN}",
			"max-conns": "${@cfg:DB_MAX_CONNS}",
		},
	})

	reg.DefineService(&schema.ServiceDef{
		Name: "logger",
		Type: "mock-logger",
		Config: map[string]any{
			"level": "${@cfg:LOG_LEVEL}",
		},
	})

	reg.DefineService(&schema.ServiceDef{
		Name:      "user-service",
		Type:      "mock-user-service",
		DependsOn: []string{"db", "logger"},
	})

	reg.DefineService(&schema.ServiceDef{
		Name:      "order-service",
		Type:      "mock-order-service",
		DependsOn: []string{"dbOrder:db", "userSvc:user-service", "logger"},
	})

	return reg
}

func TestDeployment_Creation(t *testing.T) {
	dep := New("test-deployment")

	if dep.Name() != "test-deployment" {
		t.Errorf("expected name 'test-deployment', got '%s'", dep.Name())
	}

	if dep.Registry() != Global() {
		t.Error("expected deployment to use global registry")
	}
}

func TestDeployment_ConfigOverrides(t *testing.T) {
	reg := setupTestRegistry()
	dep := NewWithRegistry("test", reg)

	// Set override
	dep.SetConfigOverride("LOG_LEVEL", "debug")

	// Get config (should return override)
	value, ok := dep.GetConfig("LOG_LEVEL")
	if !ok {
		t.Fatal("LOG_LEVEL config not found")
	}

	if value != "debug" {
		t.Errorf("expected 'debug', got '%v'", value)
	}

	// Get non-overridden config (should return global value)
	value, ok = dep.GetConfig("DB_MAX_CONNS")
	if !ok {
		t.Fatal("DB_MAX_CONNS config not found")
	}

	if value != 20 {
		t.Errorf("expected 20, got %v", value)
	}
}

func TestDeployment_ServerCreation(t *testing.T) {
	dep := New("test")

	server := dep.NewServer("main-server", "http://localhost")

	if server.Name() != "main-server" {
		t.Errorf("expected name 'main-server', got '%s'", server.Name())
	}

	if server.BaseURL() != "http://localhost" {
		t.Errorf("expected base URL 'http://localhost', got '%s'", server.BaseURL())
	}

	// Verify server is in deployment
	retrieved, ok := dep.GetServer("main-server")
	if !ok {
		t.Fatal("server not found in deployment")
	}

	if retrieved != server {
		t.Error("retrieved server is not the same instance")
	}
}

func TestDeployment_AppCreation(t *testing.T) {
	dep := New("test")
	server := dep.NewServer("main-server", "http://localhost")

	app := server.NewApp(3000)

	if app.Port() != 3000 {
		t.Errorf("expected port 3000, got %d", app.Port())
	}

	// Verify app is in server
	apps := server.Apps()
	if len(apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(apps))
	}

	if apps[0] != app {
		t.Error("app in server is not the same instance")
	}
}

func TestApp_AddService(t *testing.T) {
	reg := setupTestRegistry()
	dep := NewWithRegistry("test", reg)
	server := dep.NewServer("main-server", "http://localhost")
	app := server.NewApp(3000)

	// Add service
	app.AddService("db")

	// Verify service is added
	services := app.Services()
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}

	svc, ok := services["db"]
	if !ok {
		t.Fatal("db service not found")
	}

	if svc.name != "db" {
		t.Errorf("expected service name 'db', got '%s'", svc.name)
	}

	if svc.resolved {
		t.Error("service should not be resolved yet")
	}
}

func TestApp_GetService_Simple(t *testing.T) {
	reg := setupTestRegistry()
	dep := NewWithRegistry("test", reg)
	server := dep.NewServer("main-server", "http://localhost")
	app := server.NewApp(3000)

	// Add services
	app.AddServices("db", "logger")

	// Get DB service (should instantiate)
	dbInstance, err := app.GetService("db")
	if err != nil {
		t.Fatalf("failed to get db service: %v", err)
	}

	db, ok := dbInstance.(*MockDB)
	if !ok {
		t.Fatalf("expected *MockDB, got %T", dbInstance)
	}

	// Verify config resolution
	if db.DSN != "postgres://localhost/testdb" {
		t.Errorf("expected DSN 'postgres://localhost/testdb', got '%s'", db.DSN)
	}

	if db.MaxConns != 20 {
		t.Errorf("expected MaxConns 20, got %d", db.MaxConns)
	}

	// Get logger service
	loggerInstance, err := app.GetService("logger")
	if err != nil {
		t.Fatalf("failed to get logger service: %v", err)
	}

	logger, ok := loggerInstance.(*MockLogger)
	if !ok {
		t.Fatalf("expected *MockLogger, got %T", loggerInstance)
	}

	if logger.Level != "info" {
		t.Errorf("expected level 'info', got '%s'", logger.Level)
	}

	// Get same service again (should return cached)
	dbInstance2, err := app.GetService("db")
	if err != nil {
		t.Fatalf("failed to get db service second time: %v", err)
	}

	if dbInstance != dbInstance2 {
		t.Error("expected same instance (cached)")
	}
}

func TestApp_GetService_WithDependencies(t *testing.T) {
	reg := setupTestRegistry()
	dep := NewWithRegistry("test", reg)
	server := dep.NewServer("main-server", "http://localhost")
	app := server.NewApp(3000)

	// Add all services
	app.AddServices("db", "logger", "user-service")

	// Get user service (should instantiate with dependencies)
	userSvcInstance, err := app.GetService("user-service")
	if err != nil {
		t.Fatalf("failed to get user-service: %v", err)
	}

	userSvc, ok := userSvcInstance.(*MockUserService)
	if !ok {
		t.Fatalf("expected *MockUserService, got %T", userSvcInstance)
	}

	// Verify lazy dependencies were injected
	if userSvc.DB == nil {
		t.Fatal("DB lazy dependency not injected")
	}

	if userSvc.Logger == nil {
		t.Fatal("Logger lazy dependency not injected")
	}

	// Resolve lazy dependencies and verify (typed, no cast needed!)
	db := userSvc.DB.Get()
	if db.DSN != "postgres://localhost/testdb" {
		t.Errorf("expected DB DSN 'postgres://localhost/testdb', got '%s'", db.DSN)
	}

	logger := userSvc.Logger.Get()
	if logger.Level != "info" {
		t.Errorf("expected logger level 'info', got '%s'", logger.Level)
	}
}

func TestApp_GetService_WithAliases(t *testing.T) {
	reg := setupTestRegistry()
	dep := NewWithRegistry("test", reg)
	server := dep.NewServer("main-server", "http://localhost")
	app := server.NewApp(3000)

	// Add all services
	app.AddServices("db", "logger", "user-service", "order-service")

	// Get order service (has aliased dependencies)
	orderSvcInstance, err := app.GetService("order-service")
	if err != nil {
		t.Fatalf("failed to get order-service: %v", err)
	}

	orderSvc, ok := orderSvcInstance.(*MockOrderService)
	if !ok {
		t.Fatalf("expected *MockOrderService, got %T", orderSvcInstance)
	}

	// Verify lazy dependencies were injected with correct aliases
	if orderSvc.DB == nil {
		t.Fatal("DB lazy dependency (dbOrder) not injected")
	}

	if orderSvc.UserService == nil {
		t.Fatal("UserService lazy dependency (userSvc) not injected")
	}

	if orderSvc.Logger == nil {
		t.Fatal("Logger lazy dependency not injected")
	}

	// Verify it's the same DB instance (lazy cached, typed)
	dbInstance, _ := app.GetService("db")
	dbResolved := orderSvc.DB.Get()
	if dbResolved != dbInstance.(*MockDB) {
		t.Error("expected same DB instance (cached)")
	}

	// Verify it's the same UserService instance (lazy cached, MustGet for fail-fast)
	userSvcInstance, _ := app.GetService("user-service")
	userSvcResolved := orderSvc.UserService.MustGet()
	if userSvcResolved != userSvcInstance.(*MockUserService) {
		t.Error("expected same UserService instance (cached)")
	}
}

func TestApp_FluentAPI(t *testing.T) {
	reg := setupTestRegistry()
	dep := NewWithRegistry("test", reg)

	// Test fluent API chaining
	dep.SetConfigOverride("LOG_LEVEL", "debug").
		SetConfigOverride("DB_MAX_CONNS", 50)

	server := dep.NewServer("main-server", "http://localhost")

	app := server.NewApp(3000).
		AddServices("db", "logger", "user-service").
		AddRouter("health-router", nil)

	// Verify config overrides
	logLevel, _ := dep.GetConfig("LOG_LEVEL")
	if logLevel != "debug" {
		t.Errorf("expected LOG_LEVEL 'debug', got '%v'", logLevel)
	}

	maxConns, _ := dep.GetConfig("DB_MAX_CONNS")
	if maxConns != 50 {
		t.Errorf("expected DB_MAX_CONNS 50, got %v", maxConns)
	}

	// Verify services added
	if len(app.Services()) != 3 {
		t.Errorf("expected 3 services, got %d", len(app.Services()))
	}

	// Verify router added
	if len(app.Routers()) != 1 {
		t.Errorf("expected 1 router, got %d", len(app.Routers()))
	}
}

func TestParseDependency(t *testing.T) {
	tests := []struct {
		input     string
		wantParam string
		wantSvc   string
	}{
		{"db", "db", "db"},
		{"logger", "logger", "logger"},
		{"dbOrder:db-order", "dbOrder", "db-order"},
		{"userSvc:user-service", "userSvc", "user-service"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			param, svc := parseDependency(tt.input)
			if param != tt.wantParam {
				t.Errorf("param: expected '%s', got '%s'", tt.wantParam, param)
			}
			if svc != tt.wantSvc {
				t.Errorf("service: expected '%s', got '%s'", tt.wantSvc, svc)
			}
		})
	}
}

func TestApp_ServiceNotFound(t *testing.T) {
	reg := setupTestRegistry()
	dep := NewWithRegistry("test", reg)
	server := dep.NewServer("main-server", "http://localhost")
	app := server.NewApp(3000)

	// Try to get non-existent service
	_, err := app.GetService("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent service")
	}

	expectedMsg := "service nonexistent not found in app"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestApp_MissingDependency(t *testing.T) {
	// NOTE: With lazy loading, missing dependencies are only detected
	// when the dependency is actually accessed via .Get()
	// The service instantiation itself will succeed

	reg := setupTestRegistry()
	dep := NewWithRegistry("test", reg)
	server := dep.NewServer("main-server", "http://localhost")
	app := server.NewApp(3000)

	// Add user-service but not its dependencies
	app.AddService("user-service")

	// Service instantiation succeeds (lazy loading)
	userSvc, err := app.GetService("user-service")
	if err != nil {
		t.Fatalf("unexpected error during service instantiation: %v", err)
	}

	if userSvc == nil {
		t.Fatal("expected service instance")
	}

	// TODO: The panic will occur when the factory tries to access deps["db"].Get()
	// This is expected behavior with lazy loading - errors happen on access, not registration
}
