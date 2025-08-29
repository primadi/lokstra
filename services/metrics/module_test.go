package metrics

import (
	"testing"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

// mockRegistrationContext implements registration.RegistrationContext for testing
type mockRegistrationContext struct {
	factories map[string]func(config any) (service.Service, error)
}

// GetRawHandler implements registration.Context.
func (m *mockRegistrationContext) GetRawHandler(name string) *registration.RawHandlerRegister {
	panic("unimplemented")
}

// RegisterRawHandler implements registration.Context.
func (m *mockRegistrationContext) RegisterRawHandler(name string, handler request.RawHandlerFunc) {
	panic("unimplemented")
}

// Implement all required methods for registration.RegistrationContext
func (m *mockRegistrationContext) RegisterService(serviceName string, service service.Service) error {
	return nil
}

func (m *mockRegistrationContext) GetService(serviceName string) (service.Service, error) {
	return nil, nil
}

func (m *mockRegistrationContext) CreateService(factoryName, serviceName string, config ...any) (service.Service, error) {
	return nil, nil
}

func (m *mockRegistrationContext) GetOrCreateService(factoryName, serviceName string, config ...any) (service.Service, error) {
	return nil, nil
}

func (m *mockRegistrationContext) RegisterServiceFactory(name string, factory func(config any) (service.Service, error)) {
	if m.factories == nil {
		m.factories = make(map[string]func(config any) (service.Service, error))
	}
	m.factories[name] = factory
}

func (m *mockRegistrationContext) GetServiceFactory(factoryName string) (service.ServiceFactory, bool) {
	return nil, false
}

func (m *mockRegistrationContext) GetServiceFactories(pattern string) []service.ServiceFactory {
	return nil
}

func (m *mockRegistrationContext) GetHandler(name string) *registration.HandlerRegister {
	return nil
}

func (m *mockRegistrationContext) RegisterHandler(name string, handler any) {
}

func (m *mockRegistrationContext) RegisterMiddlewareFactory(name string, middlewareFactory midware.Factory) error {
	return nil
}

func (m *mockRegistrationContext) RegisterMiddlewareFactoryWithPriority(name string, middlewareFactory midware.Factory, priority int) error {
	return nil
}

func (m *mockRegistrationContext) RegisterMiddlewareFunc(name string, middlewareFunc midware.Func) error {
	return nil
}

func (m *mockRegistrationContext) RegisterMiddlewareFuncWithPriority(name string, middlewareFunc midware.Func, priority int) error {
	return nil
}

func (m *mockRegistrationContext) GetMiddlewareFactory(name string) (midware.Factory, int, bool) {
	return nil, 0, false
}

func (m *mockRegistrationContext) GetValue(key string) (any, bool) {
	return nil, false
}

func (m *mockRegistrationContext) SetValue(key string, value any) {
}

func (m *mockRegistrationContext) RegisterCompiledModule(pluginPath string) error {
	return nil
}

func (m *mockRegistrationContext) RegisterCompiledModuleWithFuncName(pluginPath string, getModuleFuncName string) error {
	return nil
}

func (m *mockRegistrationContext) RegisterModule(getModuleFunc func() registration.Module) error {
	return nil
}

func (m *mockRegistrationContext) NewPermissionContextFromConfig(settings map[string]any, permission map[string]any) registration.Context {
	return m
}

var _ registration.Context = (*mockRegistrationContext)(nil)

func TestModule_Register(t *testing.T) {
	module := GetModule()
	regCtx := &mockRegistrationContext{}

	err := module.Register(regCtx)
	if err != nil {
		t.Fatalf("Failed to register module: %v", err)
	}

	if len(regCtx.factories) != 1 {
		t.Fatalf("Expected 1 factory, got %d", len(regCtx.factories))
	}

	factory, exists := regCtx.factories[FACTORY_NAME]
	if !exists {
		t.Fatalf("Factory %s not registered", FACTORY_NAME)
	}

	// Test factory with default config
	svc, err := factory(nil)
	if err != nil {
		t.Fatalf("Failed to create service with nil config: %v", err)
	}

	if svc == nil {
		t.Fatal("Service should not be nil")
	}

	metricsService, ok := svc.(*MetricsService)
	if !ok {
		t.Fatal("Service should be *MetricsService")
	}

	if !metricsService.config.Enabled {
		t.Error("Default service should be enabled")
	}
}

func TestModule_RegisterWithMapConfig(t *testing.T) {
	module := GetModule()
	regCtx := &mockRegistrationContext{}

	err := module.Register(regCtx)
	if err != nil {
		t.Fatalf("Failed to register module: %v", err)
	}

	factory := regCtx.factories[FACTORY_NAME]

	// Test with map configuration
	config := map[string]any{
		"enabled":          true,
		"endpoint":         "/custom-metrics",
		"namespace":        "test_app",
		"subsystem":        "api",
		"collect_interval": "30s",
		"timeout":          "5s",
		"host":             "0.0.0.0",
		"port":             9090,
		"buckets":          []interface{}{0.1, 1.0, 10.0},
		"labels": map[string]interface{}{
			"service": "test",
			"version": "1.0",
		},
		"include_go_metrics":      false,
		"include_process_metrics": false,
	}

	svc, err := factory(config)
	if err != nil {
		t.Fatalf("Failed to create service with map config: %v", err)
	}

	metricsService := svc.(*MetricsService)

	if metricsService.config.Endpoint != "/custom-metrics" {
		t.Errorf("Expected endpoint '/custom-metrics', got '%s'", metricsService.config.Endpoint)
	}

	if metricsService.config.Namespace != "test_app" {
		t.Errorf("Expected namespace 'test_app', got '%s'", metricsService.config.Namespace)
	}

	if metricsService.config.Subsystem != "api" {
		t.Errorf("Expected subsystem 'api', got '%s'", metricsService.config.Subsystem)
	}

	if metricsService.config.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", metricsService.config.Port)
	}

	if len(metricsService.config.Buckets) != 3 {
		t.Errorf("Expected 3 buckets, got %d", len(metricsService.config.Buckets))
	}

	if len(metricsService.config.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(metricsService.config.Labels))
	}

	if metricsService.config.IncludeGoMetrics {
		t.Error("Expected include_go_metrics to be false")
	}

	if metricsService.config.IncludeProcessMetrics {
		t.Error("Expected include_process_metrics to be false")
	}
}

func TestModule_RegisterWithInvalidConfig(t *testing.T) {
	module := GetModule()
	regCtx := &mockRegistrationContext{}

	err := module.Register(regCtx)
	if err != nil {
		t.Fatalf("Failed to register module: %v", err)
	}

	factory := regCtx.factories[FACTORY_NAME]

	// Test with invalid config type
	_, err = factory(123)
	if err == nil {
		t.Error("Expected error with invalid config type")
	}
}

func TestModule_Name(t *testing.T) {
	module := GetModule()
	if module.Name() != FACTORY_NAME {
		t.Errorf("Expected module name '%s', got '%s'", FACTORY_NAME, module.Name())
	}
}

func TestModule_Description(t *testing.T) {
	module := GetModule()
	description := module.Description()
	if description == "" {
		t.Error("Module description should not be empty")
	}
}

func TestModule_Interface(t *testing.T) {
	module := GetModule()

	// Test that module implements registration.Module
	var _ registration.Module = module
}

func TestParseBuckets(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []float64
		hasError bool
	}{
		{
			name:     "valid float slice",
			input:    []interface{}{0.1, 1.0, 10.0},
			expected: []float64{0.1, 1.0, 10.0},
			hasError: false,
		},
		{
			name:     "valid mixed numbers",
			input:    []interface{}{0.1, 1, int64(10)},
			expected: []float64{0.1, 1.0, 10.0},
			hasError: false,
		},
		{
			name:     "direct float64 slice",
			input:    []float64{0.1, 1.0, 10.0},
			expected: []float64{0.1, 1.0, 10.0},
			hasError: false,
		},
		{
			name:     "invalid type",
			input:    "not an array",
			expected: nil,
			hasError: true,
		},
		{
			name:     "invalid element",
			input:    []interface{}{0.1, "invalid", 10.0},
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseBuckets(tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d buckets, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected bucket[%d] = %f, got %f", i, expected, result[i])
				}
			}
		})
	}
}

func TestParseLabels(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected map[string]string
		hasError bool
	}{
		{
			name: "valid string map",
			input: map[string]interface{}{
				"service": "test",
				"version": "1.0",
			},
			expected: map[string]string{
				"service": "test",
				"version": "1.0",
			},
			hasError: false,
		},
		{
			name: "mixed value types",
			input: map[string]interface{}{
				"service": "test",
				"port":    8080,
				"enabled": true,
			},
			expected: map[string]string{
				"service": "test",
				"port":    "8080",
				"enabled": "true",
			},
			hasError: false,
		},
		{
			name: "direct string map",
			input: map[string]string{
				"service": "test",
				"version": "1.0",
			},
			expected: map[string]string{
				"service": "test",
				"version": "1.0",
			},
			hasError: false,
		},
		{
			name:     "invalid type",
			input:    "not a map",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLabels(tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d labels, got %d", len(tt.expected), len(result))
				return
			}

			for key, expected := range tt.expected {
				if result[key] != expected {
					t.Errorf("Expected label[%s] = %s, got %s", key, expected, result[key])
				}
			}
		})
	}
}
