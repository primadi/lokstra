package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

// Test custom validator registration
func TestRegisterValidator_CustomValidator(t *testing.T) {
	// Register a custom UUID validator
	RegisterValidator("uuid", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
		if fieldValue.Kind() != reflect.String {
			return nil
		}

		value := fieldValue.String()
		if value == "" {
			return nil // Use required tag for empty check
		}

		// Simple UUID v4 regex
		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
		if !uuidRegex.MatchString(value) {
			return fmt.Errorf("%s must be a valid UUID", fieldName)
		}

		return nil
	})

	type TestStruct struct {
		ID string `json:"id" validate:"required,uuid"`
	}

	// Test valid UUID
	valid := TestStruct{ID: "550e8400-e29b-41d4-a716-446655440000"}
	errors, err := ValidateStruct(&valid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}

	// Test invalid UUID
	invalid := TestStruct{ID: "not-a-uuid"}
	errors, err = ValidateStruct(&invalid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) == 0 {
		t.Errorf("expected error for invalid UUID, got none")
	}
	if len(errors) > 0 && errors[0].Field != "id" {
		t.Errorf("expected error on 'id' field, got '%s'", errors[0].Field)
	}
}

func TestRegisterValidator_URLValidator(t *testing.T) {
	// Register a custom URL validator
	RegisterValidator("url", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
		if fieldValue.Kind() != reflect.String {
			return nil
		}

		value := fieldValue.String()
		if value == "" {
			return nil
		}

		// Simple URL validation
		if !regexp.MustCompile(`^https?://`).MatchString(value) {
			return fmt.Errorf("%s must be a valid URL (http:// or https://)", fieldName)
		}

		return nil
	})

	type TestStruct struct {
		Website string `json:"website" validate:"required,url"`
	}

	// Test valid URL
	valid := TestStruct{Website: "https://example.com"}
	errors, err := ValidateStruct(&valid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errors), errors)
	}

	// Test another valid URL
	valid2 := TestStruct{Website: "http://example.com"}
	errors, err = ValidateStruct(&valid2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}

	// Test invalid URL
	invalid := TestStruct{Website: "not-a-url"}
	errors, err = ValidateStruct(&invalid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) == 0 {
		t.Errorf("expected error for invalid URL, got none")
	}
}

func TestRegisterValidator_AlphanumericValidator(t *testing.T) {
	// Register a custom alphanumeric validator
	RegisterValidator("alphanum", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
		if fieldValue.Kind() != reflect.String {
			return nil
		}

		value := fieldValue.String()
		if value == "" {
			return nil
		}

		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(value) {
			return fmt.Errorf("%s must contain only alphanumeric characters", fieldName)
		}

		return nil
	})

	type TestStruct struct {
		Username string `json:"username" validate:"required,alphanum,min=3"`
	}

	// Test valid
	valid := TestStruct{Username: "user123"}
	errors, err := ValidateStruct(&valid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}

	// Test invalid (contains special char)
	invalid := TestStruct{Username: "user_123"}
	errors, err = ValidateStruct(&invalid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) == 0 {
		t.Errorf("expected error for invalid username, got none")
	}
}

func TestRegisterValidator_WithParameter(t *testing.T) {
	// Register a custom validator that uses ruleValue parameter
	RegisterValidator("startswith", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
		if fieldValue.Kind() != reflect.String {
			return nil
		}

		value := fieldValue.String()
		if value == "" {
			return nil
		}

		if !regexp.MustCompile(`^` + regexp.QuoteMeta(ruleValue)).MatchString(value) {
			return fmt.Errorf("%s must start with '%s'", fieldName, ruleValue)
		}

		return nil
	})

	type TestStruct struct {
		Code string `json:"code" validate:"required,startswith=PRD-"`
	}

	// Test valid
	valid := TestStruct{Code: "PRD-12345"}
	errors, err := ValidateStruct(&valid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errors), errors)
	}

	// Test invalid
	invalid := TestStruct{Code: "DEV-12345"}
	errors, err = ValidateStruct(&invalid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) == 0 {
		t.Errorf("expected error, got none")
	}
}

func TestRegisterValidator_Override(t *testing.T) {
	// Test that we can override an existing validator
	originalValidatorCount := len(validatorRegistry)

	// Register a custom email validator
	RegisterValidator("email", func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
		if fieldValue.Kind() != reflect.String {
			return nil
		}
		// Very strict: must end with @example.com
		value := fieldValue.String()
		if !regexp.MustCompile(`@example\.com$`).MatchString(value) {
			return fmt.Errorf("%s must be an @example.com email", fieldName)
		}
		return nil
	})

	type TestStruct struct {
		Email string `json:"email" validate:"email"`
	}

	// Now only @example.com emails should pass
	valid := TestStruct{Email: "test@example.com"}
	errors, err := ValidateStruct(&valid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}

	// Other emails should fail
	invalid := TestStruct{Email: "test@gmail.com"}
	errors, err = ValidateStruct(&invalid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) == 0 {
		t.Errorf("expected error for non-example.com email, got none")
	}

	// Registry size should not have increased
	if len(validatorRegistry) != originalValidatorCount {
		t.Errorf("expected registry size to remain %d, got %d", originalValidatorCount, len(validatorRegistry))
	}
}

func TestRegisterValidator_ThreadSafe(t *testing.T) {
	// Test that registration is thread-safe
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			validatorName := fmt.Sprintf("custom_%d", id)
			RegisterValidator(validatorName, func(fieldName string, fieldValue reflect.Value, ruleValue string) error {
				return nil
			})
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all validators were registered
	for i := 0; i < 10; i++ {
		validatorName := fmt.Sprintf("custom_%d", i)
		if _, ok := getValidator(validatorName); !ok {
			t.Errorf("validator %s was not registered", validatorName)
		}
	}
}
