package main

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/primadi/lokstra/common/validator"
)

// init registers custom validators
func init() {
	// UUID validator - validates UUID v4 format
	validator.RegisterValidator("uuid", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
		if fieldValue.Kind() != reflect.String {
			return nil
		}

		value := fieldValue.String()
		if value == "" {
			return nil // Use 'required' tag separately for empty check
		}

		// UUID v4 regex pattern
		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
		if !uuidRegex.MatchString(value) {
			return fmt.Errorf("%s must be a valid UUID v4", fieldName)
		}

		return nil
	})

	// Product code validator - must start with specified prefix
	validator.RegisterValidator("startswith", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
		if fieldValue.Kind() != reflect.String {
			return nil
		}

		value := fieldValue.String()
		if value == "" {
			return nil
		}

		if len(value) < len(ruleValue) || value[:len(ruleValue)] != ruleValue {
			return fmt.Errorf("%s must start with '%s'", fieldName, ruleValue)
		}

		return nil
	})

	// Alphanumeric validator - only letters and numbers
	validator.RegisterValidator("alphanum", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
		if fieldValue.Kind() != reflect.String {
			return nil
		}

		value := fieldValue.String()
		if value == "" {
			return nil
		}

		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(value) {
			return fmt.Errorf("%s must contain only alphanumeric characters (no spaces or special chars)", fieldName)
		}

		return nil
	})

	// URL validator - must be valid http/https URL
	validator.RegisterValidator("url", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
		if fieldValue.Kind() != reflect.String {
			return nil
		}

		value := fieldValue.String()
		if value == "" {
			return nil
		}

		if !regexp.MustCompile(`^https?://[^\s]+\.[^\s]+`).MatchString(value) {
			return fmt.Errorf("%s must be a valid URL (http:// or https://)", fieldName)
		}

		return nil
	})

	fmt.Println("✅ Custom validators registered: uuid, startswith, alphanum, url")
}
