package flow

import (
	"fmt"
	"reflect"
	"strings"
)

// Validation helper types
type ValidationRule func(value any) (bool, string)
type FieldValidator struct {
	Field string
	Rules []ValidationRule
}

// Common validation rules
func Required() ValidationRule {
	return func(value any) (bool, string) {
		if value == nil {
			return false, "is required"
		}

		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return false, "is required"
			}
			return true, ""
		case *string:
			if v == nil || strings.TrimSpace(*v) == "" {
				return false, "is required"
			}
			return true, ""
		default:
			// Use reflection for other types
			rv := reflect.ValueOf(value)
			if rv.Kind() == reflect.Ptr {
				if rv.IsNil() {
					return false, "is required"
				}
				return true, ""
			}
			if rv.IsZero() {
				return false, "is required"
			}
			return true, ""
		}
	}
}

func MinLength(min int) ValidationRule {
	return func(value any) (bool, string) {
		str, ok := value.(string)
		if !ok {
			return false, "must be a string"
		}
		if len(strings.TrimSpace(str)) < min {
			return false, fmt.Sprintf("must be at least %d characters", min)
		}
		return true, ""
	}
}

func Email() ValidationRule {
	return func(value any) (bool, string) {
		str, ok := value.(string)
		if !ok {
			return false, "must be a string"
		}

		// Simple email validation
		str = strings.TrimSpace(str)
		if str == "" {
			return false, "must be a valid email address"
		}

		if !strings.Contains(str, "@") || !strings.Contains(str, ".") {
			return false, "must be a valid email address"
		}

		// More basic checks
		parts := strings.Split(str, "@")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return false, "must be a valid email address"
		}

		return true, ""
	}
}

// Validation helper for generic flow context
func ValidateStruct[T any](fctx *Context[T], validators []FieldValidator) error {
	fieldErrors := make(map[string]string)

	// Use reflection to get field values from fctx.Params
	rv := reflect.ValueOf(fctx.Params)
	rt := reflect.TypeOf(fctx.Params)

	// Handle pointer types
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rt.Elem()
	}

	for _, validator := range validators {
		// Find field in struct
		field, found := rt.FieldByName(validator.Field)
		if !found {
			continue
		}

		fieldValue := rv.FieldByName(validator.Field)
		if !fieldValue.IsValid() {
			continue
		}

		// Run validation rules
		for _, rule := range validator.Rules {
			valid, message := rule(fieldValue.Interface())
			if !valid {
				fieldName := getJSONFieldName(field)
				fieldErrors[fieldName] = message
				break // Stop at first validation error for this field
			}
		}
	}

	if len(fieldErrors) > 0 {
		return fctx.ErrorValidation("Validation failed", fieldErrors)
	}

	return nil
}

// Helper to get JSON field name from struct tag
func getJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return strings.ToLower(field.Name)
	}

	// Handle json:"field_name,omitempty"
	parts := strings.Split(jsonTag, ",")
	if parts[0] == "-" {
		return strings.ToLower(field.Name)
	}

	return parts[0]
}

// AddValidateRequest helper method for Flow
func (f *Flow[T]) AddValidateRequest(validators []FieldValidator) *Flow[T] {
	return f.AddAction("validate_request", func(fctx *Context[T]) error {
		return ValidateStruct(fctx, validators)
	})
}

// Convenience helper for simple required field validation
func (f *Flow[T]) AddValidateRequired(fields ...string) *Flow[T] {
	validators := make([]FieldValidator, len(fields))
	for i, field := range fields {
		validators[i] = FieldValidator{
			Field: field,
			Rules: []ValidationRule{Required()},
		}
	}
	return f.AddValidateRequest(validators)
}
