package registration

import (
	"fmt"

	"github.com/primadi/lokstra/core/service"
)

func ErrUnsupportedConfig(config any) error {
	return fmt.Errorf("unsupported config type: %T", config)
}

func ErrInvalidServiceType(serviceName, expectedType string) error {
	return fmt.Errorf("invalid service type for %s, expected %s", serviceName, expectedType)
}

func ErrServiceNotFound(serviceName string) error {
	return fmt.Errorf("service %s not found", serviceName)
}

func ErrServiceAlreadyExists(serviceName string) error {
	return fmt.Errorf("service %s already exists", serviceName)
}

func ErrServiceIsNotAllowed(serviceName string) error {
	return fmt.Errorf("service %s is not allowed", serviceName)
}

func ErrServiceFactoryNotFound(factoryName string) error {
	return fmt.Errorf("service factory %s not found", factoryName)
}

// Registers a service with the given name into service registry.
//
// The service must be initialized before calling this function.
//
// Returns error:
//   - nil on success
//   - ErrServiceAlreadyExists if a service with the same name already exists, and allowReplace is false
//   - ErrServiceIsNotAllowed if registering the service is not allowed
func RegisterService(regCtx Context, serviceName string, service service.Service,
	allowReplace bool) error {
	return regCtx.RegisterService(serviceName, service, allowReplace)
}

// Retrieves a service by name from service registry.
//
// Returns error:
//   - nil on success
//   - ErrServiceNotAllowed if accessing the service is not allowed.
//   - ErrServiceNotFound if the service does not exist.
//   - ErrServiceTypeInvalid if the service is not of the expected type.
func GetService[T any](regCtx Context, name string) (T, error) {
	var zero T

	service, err := regCtx.GetService(name)
	if err != nil {
		return zero, err
	}
	svc, ok := service.(T)
	if !ok {
		return zero, ErrInvalidServiceType(name, fmt.Sprintf("%T", zero))
	}
	return svc, nil
}

// Creates a service using the specified factory and configuration,
// and insert into service registry.
//
// Returns error:
//   - nil on success
//   - ErrServiceAlreadyExists if a service with the same name already exists, and allowReplace is false
//   - ErrServiceIsNotAllowed if registering the service is not allowed
//   - ErrServiceFactoryNotFound if the specified factory does not exist
//   - ErrInvalidServiceType if the created service is not of the expected type
func CreateService[T any](regCtx Context, factoryName, serviceName string,
	allowReplace bool, config ...any) (T, error) {
	var zero T

	service, err := regCtx.CreateService(factoryName, serviceName, allowReplace, config...)
	if err != nil {
		return zero, fmt.Errorf("failed to create service '%s' using factory '%s': %w", serviceName, factoryName, err)
	}
	svc, ok := service.(T)
	if !ok {
		return zero, fmt.Errorf("service '%s' is not of type %T", serviceName, zero)
	}
	return svc, nil
}

// Retrieves a service by name if it exists, otherwise creates it using the specified factory
// and configuration, and insert into service registry.
//
// Returns error:
//   - nil on success
//   - ErrServiceNotAllowed if accessing the service is not allowed.
//   - ErrServiceFactoryNotFound if the specified factory does not exist
func GetOrCreateService[T any](regCtx Context, factoryName, serviceName string, config ...any) (T, error) {
	var zero T

	service, err := regCtx.GetOrCreateService(factoryName, serviceName, config...)
	if err != nil {
		return zero, fmt.Errorf("failed to get or create service '%s' using factory '%s': %w", serviceName, factoryName, err)
	}
	svc, ok := service.(T)
	if !ok {
		return zero, fmt.Errorf("service '%s' is not of type %T", serviceName, zero)
	}
	return svc, nil
}
