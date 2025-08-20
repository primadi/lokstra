package dsl

import (
	"errors"
	"fmt"
)

// LocalizedError represents an error with I18n support
type LocalizedError struct {
	Code    string
	Params  map[string]any
	Message string
}

func (e *LocalizedError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// NewLocalizedError creates a new localized error
func NewLocalizedError(code string, params map[string]any) *LocalizedError {
	return &LocalizedError{
		Code:   code,
		Params: params,
	}
}

// Helper functions for common error scenarios

// ErrValidationFailed creates a validation error with I18n support
func ErrValidationFailed(field string, value any) *LocalizedError {
	return NewLocalizedError("validation.failed", map[string]any{
		"field": field,
		"value": value,
	})
}

// ErrDatabaseOperation creates a database operation error with I18n support
func ErrDatabaseOperation(operation string, err error) *LocalizedError {
	return NewLocalizedError("database.operation_failed", map[string]any{
		"operation": operation,
		"error":     err.Error(),
	})
}

// ErrRequiredField creates a required field error with I18n support
func ErrRequiredField(field string) *LocalizedError {
	return NewLocalizedError("validation.required_field", map[string]any{
		"field": field,
	})
}

// ErrUnauthorized creates an unauthorized error with I18n support
func ErrUnauthorized() *LocalizedError {
	return NewLocalizedError("auth.unauthorized", nil)
}

// ErrForbidden creates a forbidden error with I18n support
func ErrForbidden(resource string) *LocalizedError {
	return NewLocalizedError("auth.forbidden", map[string]any{
		"resource": resource,
	})
}

// ErrNotFound creates a not found error with I18n support
func ErrNotFound(resource string, id any) *LocalizedError {
	return NewLocalizedError("resource.not_found", map[string]any{
		"resource": resource,
		"id":       id,
	})
}

// LocalizeError takes an error and converts it to localized message if possible
func LocalizeError[TParam any](ctx *FlowContext[TParam], err error) error {
	if localizedErr, ok := err.(*LocalizedError); ok {
		message := ctx.GetLocalizedMessage(localizedErr.Code, localizedErr.Params)
		return errors.New(message)
	}
	return err
}

// LocalizeErrorString takes an error and returns localized string
func LocalizeErrorString[TParam any](ctx *FlowContext[TParam], err error) string {
	if localizedErr, ok := err.(*LocalizedError); ok {
		return ctx.GetLocalizedMessage(localizedErr.Code, localizedErr.Params)
	}
	return err.Error()
}

// ErrorStep creates a step that always returns an error (useful for validation)
func ErrorStep[TParam any](err error) Step[TParam] {
	return newStepCustom(func(ctx *FlowContext[TParam]) error {
		return err
	})
}

// ValidateStep creates a validation step
func ValidateStep[TParam any](validationFn func(*FlowContext[TParam]) error) Step[TParam] {
	return newStepCustom(func(ctx *FlowContext[TParam]) error {
		if err := validationFn(ctx); err != nil {
			// Add metrics for validation failure
			if ctx.serviceVar.Metrics != nil {
				ctx.serviceVar.Metrics.IncCounter("dsl_validation_failed", map[string]string{
					"error_type": fmt.Sprintf("%T", err),
				})
			}
			return err
		}

		// Add metrics for validation success
		if ctx.serviceVar.Metrics != nil {
			ctx.serviceVar.Metrics.IncCounter("dsl_validation_passed", nil)
		}

		return nil
	})
}
