package main

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/examples/application_architecture/modules/user_management/handlers"
	"github.com/primadi/lokstra/examples/application_architecture/modules/user_management/repository"
	"github.com/primadi/lokstra/examples/application_architecture/modules/user_management/services"
	"github.com/primadi/lokstra/serviceapi"
)

// ModuleConfig holds the module configuration
type ModuleConfig struct {
	TableName         string `yaml:"table_name"`
	EnableSoftDelete  bool   `yaml:"enable_soft_delete"`
	ValidationEnabled bool   `yaml:"validation_enabled"`
}

// UserManagementModule represents the user management module
type UserManagementModule struct {
	config  *ModuleConfig
	handler *handlers.UserHandler
}

// RegisterModule is the entry point for the user management module
// This function will be called by the Lokstra framework when loading the module
func RegisterModule(ctx iface.RegistrationContext) error {
	// Parse module settings
	config := &ModuleConfig{
		TableName:         "users",
		EnableSoftDelete:  true,
		ValidationEnabled: true,
	}

	if settings, _ := ctx.GetValue("module_settings"); settings != nil {
		if moduleSettings, ok := settings.(map[string]interface{}); ok {
			if tableName, ok := moduleSettings["table_name"].(string); ok {
				config.TableName = tableName
			}
			if enableSoftDelete, ok := moduleSettings["enable_soft_delete"].(bool); ok {
				config.EnableSoftDelete = enableSoftDelete
			}
			if validationEnabled, ok := moduleSettings["validation_enabled"].(bool); ok {
				config.ValidationEnabled = validationEnabled
			}
		}
	}

	// Get required services
	dbPoolService, err := ctx.GetService("db_pool")
	if err != nil {
		return fmt.Errorf("failed to get db_pool service: %w", err)
	}

	dbPool, ok := dbPoolService.(*pgxpool.Pool)
	if !ok {
		return fmt.Errorf("db_pool service is not a *pgxpool.Pool")
	}

	loggerService, err := ctx.GetService("logger")
	if err != nil {
		return fmt.Errorf("failed to get logger service: %w", err)
	}

	// Create repository, service, and handler
	userRepo := repository.NewPostgresUserRepository(dbPool, config.TableName)
	userService := services.NewUserService(userRepo, loggerService.(serviceapi.Logger), config.ValidationEnabled)
	userHandler := handlers.NewUserHandler(userService)

	// Register handlers
	ctx.RegisterHandler("user_management.list_users", userHandler.ListUsers)
	ctx.RegisterHandler("user_management.get_user", userHandler.GetUser)
	ctx.RegisterHandler("user_management.create_user", userHandler.CreateUser)
	ctx.RegisterHandler("user_management.update_user", userHandler.UpdateUser)
	ctx.RegisterHandler("user_management.delete_user", userHandler.DeleteUser)

	return nil
}

// Additional module functions for advanced module features

// RequiredServices returns the list of required services for this module
func RequiredServices() []string {
	return []string{"db_pool", "logger"}
}

// CreateServices returns services that this module wants to create
func CreateServices() map[string]interface{} {
	// This module doesn't create services, but could in more complex scenarios
	return nil
}

// RegisterServiceFactories registers service factories that other modules can use
func RegisterServiceFactories(ctx iface.RegistrationContext) error {
	// This module doesn't register service factories, but could for reusable components
	return nil
}

// RegisterHandlers registers additional handlers (already done in RegisterModule)
func RegisterHandlers(ctx iface.RegistrationContext) error {
	// Handlers are already registered in RegisterModule
	// This could be used for registering additional handlers or organizing registration
	return nil
}

// RegisterMiddleware registers middleware that this module provides
func RegisterMiddleware(ctx iface.RegistrationContext) error {
	// This module doesn't provide middleware, but could for user authentication, etc.
	return nil
}
