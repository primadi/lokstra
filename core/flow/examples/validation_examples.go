package main

import (
	"github.com/primadi/lokstra/core/flow"
	"github.com/primadi/lokstra/core/request"
)

// Example DTOs
type CreateUserRequestDTO struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Age      int    `json:"age"`
}

type UpdateUserRequestDTO struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

type CreateProductRequestDTO struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	CategoryID  int     `json:"category_id"`
}

// Example 1: Simple required validation
func CreateUserHandler() request.HandlerFunc {
	return flow.NewFlow[CreateUserRequestDTO]("CreateUser").
		// Simple validation - just check required fields
		AddValidateRequired("Username", "Email", "Password").
		AddAction("create_user", func(fctx *flow.Context[CreateUserRequestDTO]) error {
			user := fctx.Params
			// Type-safe access to validated fields
			_ = user.Username // guaranteed to be non-empty
			_ = user.Email    // guaranteed to be non-empty
			_ = user.Password // guaranteed to be non-empty

			return fctx.OkCreated(map[string]any{
				"id":       123,
				"username": user.Username,
				"email":    user.Email,
			})
		}).AsHandler()
}

// Example 2: Advanced validation with rules
func CreateUserAdvancedHandler() request.HandlerFunc {
	return flow.NewFlow[CreateUserRequestDTO]("CreateUserAdvanced").
		// Advanced validation with custom rules
		AddValidateRequest([]flow.FieldValidator{
			{
				Field: "Username",
				Rules: []flow.ValidationRule{
					flow.Required(),
					flow.MinLength(3),
				},
			},
			{
				Field: "Email",
				Rules: []flow.ValidationRule{
					flow.Required(),
					flow.Email(),
				},
			},
			{
				Field: "Password",
				Rules: []flow.ValidationRule{
					flow.Required(),
					flow.MinLength(8),
				},
			},
		}).
		AddAction("check_username_unique", func(fctx *flow.Context[CreateUserRequestDTO]) error {
			// Business validation
			if usernameExists(fctx.Params.Username) {
				return fctx.ErrorBadRequest("Username already taken")
			}
			return nil
		}).
		AddAction("check_email_unique", func(fctx *flow.Context[CreateUserRequestDTO]) error {
			if emailExists(fctx.Params.Email) {
				return fctx.ErrorBadRequest("Email already registered")
			}
			return nil
		}).
		AddAction("create_user", func(fctx *flow.Context[CreateUserRequestDTO]) error {
			user := fctx.Params
			// Create user logic here
			return fctx.OkCreated(user)
		}).AsHandler()
}

// Example 3: Update with optional fields
func UpdateUserHandler() request.HandlerFunc {
	return flow.NewFlow[UpdateUserRequestDTO]("UpdateUser").
		// Validate only fields that are provided
		AddValidateRequest([]flow.FieldValidator{
			{
				Field: "Email",
				Rules: []flow.ValidationRule{
					flow.Email(), // No Required() - optional field
				},
			},
			{
				Field: "Password",
				Rules: []flow.ValidationRule{
					flow.MinLength(8), // No Required() - optional field
				},
			},
		}).
		AddAction("update_user", func(fctx *flow.Context[UpdateUserRequestDTO]) error {
			update := fctx.Params

			// Type-safe access to optional fields
			if update.Email != nil {
				// Email is validated if provided
				_ = *update.Email
			}

			if update.Password != nil {
				// Password is validated if provided
				_ = *update.Password
			}

			return fctx.Ok(map[string]any{
				"updated": true,
			})
		}).AsHandler()
}

// Example 4: Custom validation rule
func StrongPassword() flow.ValidationRule {
	return func(value any) (bool, string) {
		password, ok := value.(string)
		if !ok {
			return false, "must be a string"
		}

		if len(password) < 8 {
			return false, "must be at least 8 characters"
		}

		hasUpper := false
		hasLower := false
		hasDigit := false

		for _, r := range password {
			switch {
			case r >= 'A' && r <= 'Z':
				hasUpper = true
			case r >= 'a' && r <= 'z':
				hasLower = true
			case r >= '0' && r <= '9':
				hasDigit = true
			}
		}

		if !hasUpper {
			return false, "must contain uppercase letter"
		}
		if !hasLower {
			return false, "must contain lowercase letter"
		}
		if !hasDigit {
			return false, "must contain number"
		}

		return true, ""
	}
}

// Example 5: Using custom validation rule
func CreateUserStrongPasswordHandler() request.HandlerFunc {
	return flow.NewFlow[CreateUserRequestDTO]("CreateUserStrong").
		AddValidateRequest([]flow.FieldValidator{
			{
				Field: "Username",
				Rules: []flow.ValidationRule{
					flow.Required(),
					flow.MinLength(3),
				},
			},
			{
				Field: "Email",
				Rules: []flow.ValidationRule{
					flow.Required(),
					flow.Email(),
				},
			},
			{
				Field: "Password",
				Rules: []flow.ValidationRule{
					flow.Required(),
					StrongPassword(), // Custom rule
				},
			},
		}).
		AddAction("create_user", func(fctx *flow.Context[CreateUserRequestDTO]) error {
			return fctx.OkCreated(fctx.Params)
		}).AsHandler()
}

// Example 6: Product validation with numeric fields
func CreateProductHandler() request.HandlerFunc {
	return flow.NewFlow[CreateProductRequestDTO]("CreateProduct").
		AddValidateRequired("Name", "Price", "CategoryID").
		AddAction("validate_business_rules", func(fctx *flow.Context[CreateProductRequestDTO]) error {
			product := fctx.Params

			// Business validation
			if product.Price <= 0 {
				return fctx.ErrorBadRequest("Price must be greater than 0")
			}

			if product.CategoryID <= 0 {
				return fctx.ErrorBadRequest("Invalid category ID")
			}

			return nil
		}).
		AddAction("create_product", func(fctx *flow.Context[CreateProductRequestDTO]) error {
			return fctx.OkCreated(fctx.Params)
		}).AsHandler()
}

// Mock functions for examples
func usernameExists(username string) bool {
	return username == "admin" // Mock check
}

func emailExists(email string) bool {
	return email == "admin@example.com" // Mock check
}

func main() {
	// This file demonstrates validation patterns
	// In real app, these handlers would be registered with router

	_ = CreateUserHandler()
	_ = CreateUserAdvancedHandler()
	_ = UpdateUserHandler()
	_ = CreateUserStrongPasswordHandler()
	_ = CreateProductHandler()
}
