package service

import (
	"fmt"
)

type ServiceFactory = func(config any) (Service, error)

type Service = any

func ErrUnsupportedConfig(config any) error {
	return fmt.Errorf("unsupported config type: %T", config)
}

func ErrInvalidServiceType(serviceName, expectedType string) error {
	return fmt.Errorf("invalid service type for %s, expected %s", serviceName, expectedType)
}

func ErrServiceNotFound(serviceName string) error {
	return fmt.Errorf("service %s not found", serviceName)
}
