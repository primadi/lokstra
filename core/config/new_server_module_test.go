package config

import (
	"testing"

	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartModulesFromConfig(t *testing.T) {
	regCtx := registration.NewGlobalContext()

	// Test case 1: Module with required service that doesn't exist
	t.Run("RequiredServices_NotFound", func(t *testing.T) {
		modules := []*ModuleConfig{
			{
				Name:             "test-module",
				Path:             "", // no plugin path
				RequiredServices: []string{"non-existent-service"},
			},
		}

		err := startModulesFromConfig(regCtx, modules)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires service non-existent-service")
	})

	// Test case 2: Create services from module config
	t.Run("CreateServices", func(t *testing.T) {
		// Add a service factory for testing
		regCtx.RegisterServiceFactory("test-factory", func(config any) (service.Service, error) {
			return &mockService{name: "test-service-instance"}, nil
		})

		modules := []*ModuleConfig{
			{
				Name: "test-module",
				Path: "", // no plugin path
				CreateServices: []ServiceConfig{
					{
						Name:   "test-service",
						Type:   "test-factory",
						Config: map[string]any{"key": "value"},
					},
				},
			},
		}

		err := startModulesFromConfig(regCtx, modules)
		require.NoError(t, err)

		// Verify service was created
		svc, err := regCtx.GetService("test-service")
		assert.NoError(t, err)
		assert.NotNil(t, svc)
	})

	// Test case 3: Empty module should succeed
	t.Run("EmptyModule", func(t *testing.T) {
		modules := []*ModuleConfig{
			{
				Name: "empty-module",
				Path: "", // no plugin path, no other configurations
			},
		}

		err := startModulesFromConfig(regCtx, modules)
		assert.NoError(t, err) // Should succeed with empty module
	})
}

func TestCreateServiceFromConfig(t *testing.T) {
	regCtx := registration.NewGlobalContext()

	// Register a test service factory
	regCtx.RegisterServiceFactory("test-factory-2", func(config any) (service.Service, error) {
		return &mockService{name: "created-service"}, nil
	})

	t.Run("ValidService", func(t *testing.T) {
		serviceConfig := &ServiceConfig{
			Name:   "test-service-2",
			Type:   "test-factory-2",
			Config: map[string]any{"setting": "value"},
		}

		err := createServiceFromConfig(regCtx, serviceConfig)
		require.NoError(t, err)

		// Verify service was created
		svc, err := regCtx.GetService("test-service-2")
		assert.NoError(t, err)
		assert.NotNil(t, svc)
	})

	t.Run("InvalidFactory", func(t *testing.T) {
		serviceConfig := &ServiceConfig{
			Name:   "invalid-service",
			Type:   "non-existent-factory",
			Config: map[string]any{},
		}

		err := createServiceFromConfig(regCtx, serviceConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create service")
	})
}

// Mock service for testing
type mockService struct {
	name string
}

func (m *mockService) InstanceName() string {
	return m.name
}

func (m *mockService) GetConfig(key string) any {
	return nil
}
