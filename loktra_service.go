package lokstra

import (
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/service"
)

// Retrieves a service by name from service registry.
//
// Returns error:
//   - nil on success
//   - ErrServiceNotAllowed if accessing the service is not allowed.
//   - ErrServiceNotFound if the service does not exist.
//   - ErrServiceTypeInvalid if the service is not of the expected type.
func GetService[T service.Service](regCtx RegistrationContext, serviceName string) (T, error) {
	return registration.GetService[T](regCtx, serviceName)
}

// Retrieves a service by name if it exists, otherwise creates it using the specified factory
// and configuration, and insert into service registry.
//
// Returns error:
//   - nil on success
//   - ErrServiceNotAllowed if accessing the service is not allowed.
//   - ErrServiceFactoryNotFound if the specified factory does not exist
func GetOrCreateService[T any](regCtx RegistrationContext,
	serviceName string, factoryName string, config ...any) (T, error) {
	return registration.GetOrCreateService[T](regCtx, serviceName, factoryName, config...)
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
func CreateService[T any](regCtx RegistrationContext, factoryName, serviceName string,
	allowReplace bool, config ...any) (T, error) {
	return registration.CreateService[T](regCtx, factoryName, serviceName, allowReplace, config...)
}
