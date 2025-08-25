package serviceapi

import (
	"fmt"

	"github.com/primadi/lokstra/core/registration"
)

func GetService[T any](regCtx registration.Context, name string) (T, error) {
	var zero T

	service, err := regCtx.GetService(name)
	if err != nil {
		return zero, fmt.Errorf("service '%s' not found: %w", name, err)
	}
	svc, ok := service.(T)
	if !ok {
		return zero, fmt.Errorf("service '%s' is not of type %T", name, zero)
	}
	return svc, nil
}
