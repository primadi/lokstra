package validator

import (
	"testing"

	"github.com/primadi/lokstra/core/response/api_formatter"
)

func TestValidateStruct_Required(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required"`
	}

	// Test valid
	valid := TestStruct{Name: "John", Email: "john@example.com"}
	errors, err := ValidateStruct(&valid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}

	// Test invalid - missing name
	invalid := TestStruct{Email: "john@example.com"}
	errors, err = ValidateStruct(&invalid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
	if len(errors) > 0 && errors[0].Field != "name" {
		t.Errorf("expected error on 'name' field, got '%s'", errors[0].Field)
	}
}

func TestValidateStruct_Email(t *testing.T) {
	type TestStruct struct {
		Email string `json:"email" validate:"required,email"`
	}

	// Test valid
	valid := TestStruct{Email: "test@example.com"}
	errors, err := ValidateStruct(&valid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}

	// Test invalid email
	invalid := TestStruct{Email: "not-an-email"}
	errors, err = ValidateStruct(&invalid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
}

func TestValidateStruct_MinMax(t *testing.T) {
	type TestStruct struct {
		Age      int    `json:"age" validate:"min=1,max=120"`
		Name     string `json:"name" validate:"min=3,max=50"`
		PageSize int    `json:"page_size" validate:"min=1,max=100"`
	}

	// Test valid
	valid := TestStruct{Age: 25, Name: "John Doe", PageSize: 10}
	errors, err := ValidateStruct(&valid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errors), errors)
	}

	// Test age too low
	invalid := TestStruct{Age: 0, Name: "John", PageSize: 10}
	errors, err = ValidateStruct(&invalid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) == 0 {
		t.Errorf("expected error for age, got none")
	}

	// Test age too high
	invalid2 := TestStruct{Age: 150, Name: "John", PageSize: 10}
	errors, err = ValidateStruct(&invalid2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) == 0 {
		t.Errorf("expected error for age, got none")
	}

	// Test name too short
	invalid3 := TestStruct{Age: 25, Name: "AB", PageSize: 10}
	errors, err = ValidateStruct(&invalid3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) == 0 {
		t.Errorf("expected error for name, got none")
	}
}

func TestValidateStruct_GtLt(t *testing.T) {
	type TestStruct struct {
		Price    float64 `json:"price" validate:"gt=0"`
		Discount float64 `json:"discount" validate:"gte=0,lte=100"`
	}

	// Test valid
	valid := TestStruct{Price: 99.99, Discount: 10}
	errors, err := ValidateStruct(&valid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}

	// Test invalid price (must be > 0, not >= 0)
	invalid := TestStruct{Price: 0, Discount: 10}
	errors, err = ValidateStruct(&invalid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) == 0 {
		t.Errorf("expected error for price, got none")
	}

	// Test invalid discount (must be <= 100)
	invalid2 := TestStruct{Price: 100, Discount: 150}
	errors, err = ValidateStruct(&invalid2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) == 0 {
		t.Errorf("expected error for discount, got none")
	}
}

func TestValidateStruct_OneOf(t *testing.T) {
	type TestStruct struct {
		Role   string `json:"role" validate:"required,oneof=admin user guest"`
		Status string `json:"status" validate:"oneof=active inactive pending"`
	}

	// Test valid
	valid := TestStruct{Role: "admin", Status: "active"}
	errors, err := ValidateStruct(&valid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}

	// Test invalid role
	invalid := TestStruct{Role: "superuser", Status: "active"}
	errors, err = ValidateStruct(&invalid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) == 0 {
		t.Errorf("expected error for role, got none")
	}
}

func TestValidateStruct_OptionalFields(t *testing.T) {
	type TestStruct struct {
		Name  string  `json:"name" validate:"required"`
		Email *string `json:"email" validate:"omitempty,email"`
		Phone *string `json:"phone"`
	}

	// Test with nil optional fields
	valid := TestStruct{Name: "John"}
	errors, err := ValidateStruct(&valid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}

	// Test with valid optional field
	email := "john@example.com"
	valid2 := TestStruct{Name: "John", Email: &email}
	errors, err = ValidateStruct(&valid2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d", len(errors))
	}

	// Test with invalid optional field
	invalidEmail := "not-an-email"
	invalid := TestStruct{Name: "John", Email: &invalidEmail}
	errors, err = ValidateStruct(&invalid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(errors) == 0 {
		t.Errorf("expected error for email, got none")
	}
}

func TestValidateStruct_ReturnType(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name" validate:"required"`
		Age   int    `json:"age" validate:"min=1,max=120"`
		Email string `json:"email" validate:"required,email"`
	}

	// Test that it returns []api_formatter.FieldError
	invalid := TestStruct{}
	errors, err := ValidateStruct(&invalid)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should have errors for name and email (both required)
	if len(errors) < 2 {
		t.Errorf("expected at least 2 errors, got %d", len(errors))
	}

	// Check that errors are of correct type
	for _, fieldError := range errors {
		var _ api_formatter.FieldError = fieldError
		if fieldError.Field == "" {
			t.Errorf("field error should have Field set")
		}
		if fieldError.Message == "" {
			t.Errorf("field error should have Message set")
		}
	}
}
