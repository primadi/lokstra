# Flow Validation Helpers

## Overview

Flow validation helpers menyediakan cara yang clean dan type-safe untuk melakukan validasi input di API endpoints.

## Basic Usage

### Simple Required Validation
```go
func CreateUserHandler() request.HandlerFunc {
	return flow.NewFlow[CreateUserRequestDTO]("CreateUser").
		AddValidateRequired("Username", "Email", "Password").
		AddAction("create_user", func(fctx *flow.Context[CreateUserRequestDTO]) error {
			// All required fields validated automatically
			user := fctx.Params // Type-safe access
			return fctx.ReqCtx.OkCreated(user)
		}).AsHandler()
}
```

### Advanced Validation with Rules
```go
func CreateUserHandler() request.HandlerFunc {
	return flow.NewFlow[CreateUserRequestDTO]("CreateUser").
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
		AddAction("create_user", func(fctx *flow.Context[CreateUserRequestDTO]) error {
			return fctx.ReqCtx.OkCreated(fctx.Params)
		}).AsHandler()
}
```

## Available Validation Rules

### Required()
Validates that field is not empty/nil:
```go
flow.Required() // Field must have a value
```

### MinLength(min int)
Validates minimum string length:
```go
flow.MinLength(8) // String must be at least 8 characters
```

### Email()
Validates email format:
```go
flow.Email() // Must be valid email format
```

## Helper Methods

### AddValidateRequired(fields ...string)
Quick validation for required fields:
```go
AddValidateRequired("Username", "Email", "Password")
```

### AddValidateRequest(validators []FieldValidator)
Full validation with custom rules:
```go
AddValidateRequest([]flow.FieldValidator{
	{Field: "Email", Rules: []flow.ValidationRule{flow.Required(), flow.Email()}},
})
```

## Response Format

### Successful Validation
Validation passes → continues to next step

### Validation Errors
```json
{
  "success": false,
  "code": "BAD_REQUEST",
  "message": "Validation failed",
  "errors": {
    "username": "is required",
    "email": "must be a valid email address",
    "password": "must be at least 8 characters"
  }
}
```

## Examples

### 1. Simple CRUD Validation
```go
// CREATE - all fields required
func CreateUserHandler() request.HandlerFunc {
	return flow.NewFlow[CreateUserRequestDTO]("CreateUser").
		AddValidateRequired("Username", "Email", "Password").
		AddAction("create", createUserAction).AsHandler()
}

// UPDATE - partial validation  
func UpdateUserHandler() request.HandlerFunc {
	return flow.NewFlow[UpdateUserRequestDTO]("UpdateUser").
		AddValidateRequest([]flow.FieldValidator{
			{Field: "Email", Rules: []flow.ValidationRule{flow.Email()}},
			{Field: "Password", Rules: []flow.ValidationRule{flow.MinLength(8)}},
		}).
		AddAction("update", updateUserAction).AsHandler()
}
```

### 2. Business Rule Validation
```go
func CreateUserHandler() request.HandlerFunc {
	return flow.NewFlow[CreateUserRequestDTO]("CreateUser").
		// Input validation
		AddValidateRequired("Username", "Email", "Password").
		
		// Business validation
		AddAction("check_email_unique", func(fctx *flow.Context[CreateUserRequestDTO]) error {
			if emailExists(fctx.Params.Email) {
				return flow.BadRequestError(fctx.ReqCtx, "Email already exists")
			}
			return nil
		}).
		
		AddAction("create_user", createUserAction).AsHandler()
}
```

### 3. Custom Validation Rules
```go
// Custom validation rule
func StrongPassword() flow.ValidationRule {
	return func(value any) (bool, string) {
		password, ok := value.(string)
		if !ok {
			return false, "must be a string"
		}
		
		if len(password) < 8 {
			return false, "must be at least 8 characters"
		}
		
		if !containsUppercase(password) {
			return false, "must contain uppercase letter"
		}
		
		if !containsNumber(password) {
			return false, "must contain number"
		}
		
		return true, ""
	}
}

// Usage
AddValidateRequest([]flow.FieldValidator{
	{Field: "Password", Rules: []flow.ValidationRule{flow.Required(), StrongPassword()}},
})
```

## Benefits

1. **Type Safety**: Generic Context[T] provides compile-time type checking
2. **Reusability**: Validation rules can be reused across endpoints
3. **Declarative**: Clear and readable validation definitions
4. **Consistent**: Standard error response format
5. **Extensible**: Easy to add custom validation rules
6. **Performance**: Early validation before business logic

## Migration from Manual Validation

### Before
```go
AddAction("validate", func(fctx *flow.Context[CreateUserRequestDTO]) error {
	if fctx.Params.Username == "" {
		return flow.BadRequestError(fctx.ReqCtx, "Username is required")
	}
	if fctx.Params.Email == "" {
		return flow.BadRequestError(fctx.ReqCtx, "Email is required")
	}
	if fctx.Params.Password == "" {
		return flow.BadRequestError(fctx.ReqCtx, "Password is required")
	}
	return nil
})
```

### After
```go
AddValidateRequired("Username", "Email", "Password")
```

Much cleaner and less error-prone! 🎉
