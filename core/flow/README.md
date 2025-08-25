# Lokstra Flow Pattern - Developer Experience Guide

## Overview

Lokstra Flow menyediakan pattern yang konsisten dan nyaman untuk membuat API endpoints dengan:
- **Type Safety**: Menggunakan generics untuk type-safe data transfer
- **Error Handling**: Konsisten error response format
- **Separation of Concerns**: Memisahkan parsing, business logic, dan response
- **Reusability**: Helper functions untuk pattern yang umum digunakan

## Basic Pattern

### Simple Handler (Current Standard)

```go
func CreateNewUserHandler() request.HandlerFunc {
	return flow.NewFlow("CreateNewUser").
		AddStep("parse_request", func(fctx *flow.Context, reqCtx *request.Context) error {
			var user auth.User
			if err := reqCtx.BindBody(&user); err != nil {
				return flow.BadRequestError(reqCtx, "Invalid request body: "+err.Error())
			}
			
			// Store user in flow context for next steps
			fctx.Set("user", user)
			return nil
		}).
		AddStep("create_user", func(fctx *flow.Context, reqCtx *request.Context) error {
			// Get user from previous step
			user := fctx.MustGet("user").(auth.User)

			// Create user through repository
			repo := repository.NewUserRepository(fctx.GetDbExecutor())
			if err := repo.CreateUser(reqCtx.Context, &user); err != nil {
				return flow.BadRequestError(reqCtx, "Failed to create user: "+err.Error())
			}

			// Return success response
			return reqCtx.OkCreated(map[string]interface{}{
				"message": "User created successfully",
				"user":    user,
			})
		}).
		AsHandler()
}
```

### Advanced Pattern with Helpers

```go
func CreateNewUserHandlerV2() request.HandlerFunc {
	return flow.NewFlow("CreateNewUserV2").
		AddStep("parse_request", func(fctx *flow.Context, reqCtx *request.Context) error {
			return flow.ParseJSONBody[auth.User](fctx, reqCtx, "user")
		}).
		AddStep("create_user", func(fctx *flow.Context, reqCtx *request.Context) error {
			user := flow.GetTyped[auth.User](fctx, "user")
			repo := repository.NewUserRepository(fctx.GetDbExecutor())
			
			if err := repo.CreateUser(reqCtx.Context, &user); err != nil {
				return flow.BadRequestError(reqCtx, "Failed to create user: "+err.Error())
			}
			
			fctx.Set("created_user", user)
			return nil
		}).
		AddStep("response", func(fctx *flow.Context, reqCtx *request.Context) error {
			user := flow.GetTyped[auth.User](fctx, "created_user")
			return reqCtx.OkCreated(map[string]interface{}{
				"message": "User created successfully",
				"user":    user,
			})
		}).
		AsHandler()
}
```

## Flow Context API

### Data Storage
```go
// Store data untuk step berikutnya
fctx.Set("key", value)

// Get data dengan type assertion manual
user := fctx.MustGet("user").(auth.User)

// Get data dengan type safety (recommended)
user := flow.GetTyped[auth.User](fctx, "user")

// Get dengan default value
count := flow.GetTypedOrDefault[int](fctx, "count", 0)
```

### Database Access
```go
// Get database executor (auto connection management)
db := fctx.GetDbExecutor()

// Manual transaction control (jika diperlukan)
err := fctx.CommitTx()
err := fctx.RollbackTx()
```

## Request Context API

### Input Parsing
```go
// JSON body binding
var user auth.User
err := reqCtx.BindBody(&user)

// Path parameters
userID := reqCtx.GetPathParam("id")

// Query parameters  
page := reqCtx.GetQueryParam("page")

// Headers
authToken := reqCtx.GetHeader("Authorization")
```

### Error Handling Pattern

**❌ WRONG - Don't do this:**
```go
// This returns nil, not an error - flow will continue!
return reqCtx.ErrorBadRequest("error message")
```

**✅ CORRECT - Do this:**
```go
// Method 1: Use helper functions (recommended)
return flow.BadRequestError(reqCtx, "error message")

// Method 2: Set response + return error
reqCtx.ErrorBadRequest("error message")
return errors.New("error message")

// Method 3: Return original error if available
if err := someOperation(); err != nil {
    reqCtx.ErrorBadRequest("Operation failed: " + err.Error())
    return err
}
```

**Available Error Helper Functions:**
```go
flow.BadRequestError(reqCtx, "message")     // 400 Bad Request + return HandledError
flow.NotFoundError(reqCtx, "message")       // 404 Not Found + return HandledError  
flow.InternalError(reqCtx, "message")       // 500 Internal Server Error + return HandledError
flow.ValidationError(reqCtx, "message", fieldErrors) // 400 with field errors + return HandledError
```

### How AsHandler Works with Errors

The `AsHandler` method intelligently handles different types of errors:

```go
func (f *Flow) AsHandler() request.HandlerFunc {
	return func(reqCtx *request.Context) error {
		// ... execute steps ...
		if err := step.action(flowCtx, reqCtx); err != nil {
			// Check if this is a handled error (response already set)
			if _, isHandled := err.(HandledError); isHandled {
				return nil // HTTP request successfully handled
			}
			
			// Unhandled error - bubble up as 500 Internal Server Error
			return err
		}
		return nil
	}
}
```

**Error Types:**

1. **HandledError** (from helper functions):
   - Response is already set (400, 404, 500, etc.)
   - `AsHandler` returns `nil` → HTTP request successfully handled
   - Client receives proper error response

2. **Regular errors**:
   - No response is set
   - `AsHandler` returns the error → HTTP 500 Internal Server Error
   - Used for unexpected/unhandled errors

**Example:**
```go
// This will result in HTTP 400 Bad Request response
return flow.BadRequestError(reqCtx, "Invalid input")

// This will result in HTTP 500 Internal Server Error  
return errors.New("database connection failed")
```

## Helper Functions

### Type-Safe Parsing
```go
// Parse JSON dengan automatic error handling
return flow.ParseJSONBody[auth.User](fctx, reqCtx, "user")

// Equivalent manual implementation:
var user auth.User
if err := reqCtx.BindBody(&user); err != nil {
    return reqCtx.ErrorBadRequest("Invalid request body: " + err.Error())
}
fctx.Set("user", user)
return nil
```

### Type-Safe Data Access
```go
// Type-safe get (panic jika tidak ada)
user := flow.GetTyped[auth.User](fctx, "user")

// Type-safe get dengan default
count := flow.GetTypedOrDefault[int](fctx, "count", 0)
```

## Complete CRUD Example

```go
// CREATE
func CreateUserHandler() request.HandlerFunc {
	return flow.NewFlow("CreateUser").
		AddStep("parse", func(fctx *flow.Context, reqCtx *request.Context) error {
			return flow.ParseJSONBody[auth.User](fctx, reqCtx, "user")
		}).
		AddStep("create", func(fctx *flow.Context, reqCtx *request.Context) error {
			user := flow.GetTyped[auth.User](fctx, "user")
			repo := repository.NewUserRepository(fctx.GetDbExecutor())
			
			if err := repo.CreateUser(reqCtx.Context, &user); err != nil {
				return reqCtx.ErrorBadRequest("Failed to create user: " + err.Error())
			}
			
			return reqCtx.OkCreated(map[string]interface{}{
				"message": "User created successfully",
				"user":    user,
			})
		}).
		AsHandler()
}

// READ
func GetUserHandler() request.HandlerFunc {
	return flow.NewFlow("GetUser").
		AddStep("parse_params", func(fctx *flow.Context, reqCtx *request.Context) error {
			userID := reqCtx.GetPathParam("id")
			if userID == "" {
				return reqCtx.ErrorBadRequest("User ID is required")
			}
			fctx.Set("user_id", userID)
			return nil
		}).
		AddStep("get_user", func(fctx *flow.Context, reqCtx *request.Context) error {
			userID := flow.GetTyped[string](fctx, "user_id")
			repo := repository.NewUserRepository(fctx.GetDbExecutor())
			
			user, err := repo.GetUserByID(reqCtx.Context, userID)
			if err != nil {
				return reqCtx.ErrorNotFound("User not found")
			}
			
			return reqCtx.Ok(map[string]interface{}{
				"message": "User retrieved successfully",
				"user":    user,
			})
		}).
		AsHandler()
}

// UPDATE  
func UpdateUserHandler() request.HandlerFunc {
	return flow.NewFlow("UpdateUser").
		AddStep("parse_request", func(fctx *flow.Context, reqCtx *request.Context) error {
			userID := reqCtx.GetPathParam("id")
			if userID == "" {
				return reqCtx.ErrorBadRequest("User ID is required")
			}
			
			var user auth.User
			if err := reqCtx.BindBody(&user); err != nil {
				return reqCtx.ErrorBadRequest("Invalid request body: " + err.Error())
			}
			
			user.ID = userID // Ensure ID matches path param
			fctx.Set("user", user)
			return nil
		}).
		AddStep("update_user", func(fctx *flow.Context, reqCtx *request.Context) error {
			user := flow.GetTyped[auth.User](fctx, "user")
			repo := repository.NewUserRepository(fctx.GetDbExecutor())
			
			if err := repo.UpdateUser(reqCtx.Context, &user); err != nil {
				return reqCtx.ErrorBadRequest("Failed to update user: " + err.Error())
			}
			
			return reqCtx.OkUpdated(map[string]interface{}{
				"message": "User updated successfully",
				"user":    user,
			})
		}).
		AsHandler()
}

// DELETE
func DeleteUserHandler() request.HandlerFunc {
	return flow.NewFlow("DeleteUser").
		AddStep("parse_params", func(fctx *flow.Context, reqCtx *request.Context) error {
			userID := reqCtx.GetPathParam("id")
			if userID == "" {
				return reqCtx.ErrorBadRequest("User ID is required")
			}
			fctx.Set("user_id", userID)
			return nil
		}).
		AddStep("delete_user", func(fctx *flow.Context, reqCtx *request.Context) error {
			userID := flow.GetTyped[string](fctx, "user_id")
			repo := repository.NewUserRepository(fctx.GetDbExecutor())
			
			if err := repo.DeleteUser(reqCtx.Context, userID); err != nil {
				return reqCtx.ErrorBadRequest("Failed to delete user: " + err.Error())
			}
			
			return reqCtx.Ok(map[string]interface{}{
				"message": "User deleted successfully",
			})
		}).
		AsHandler()
}
```

## Best Practices

### 1. Step Naming
- Gunakan nama yang deskriptif: `"parse_request"`, `"validate_data"`, `"create_user"`, `"send_notification"`
- Konsisten dengan naming convention

### 2. Error Handling  
- Selalu return error yang meaningful
- Gunakan response method yang tepat (`ErrorBadRequest`, `ErrorNotFound`, dll)
- Include context dalam error message

### 3. Data Flow
- Store data di flow context untuk step berikutnya
- Gunakan type-safe helpers: `flow.GetTyped[T]()`
- Validasi data di step terpisah

### 4. Database Operations
- Gunakan `fctx.GetDbExecutor()` untuk auto connection management
- Manual transaction hanya jika diperlukan
- Repository pattern untuk database operations

### 5. Response Format
- Konsisten response structure
- Always include meaningful message
- Use appropriate HTTP status codes

## Migration dari Handler Lama

### Before (Traditional Handler)
```go
func OldCreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var user auth.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // business logic...
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

### After (Flow Pattern)
```go
func CreateUserHandler() request.HandlerFunc {
    return flow.NewFlow("CreateUser").
        AddStep("parse", func(fctx *flow.Context, reqCtx *request.Context) error {
            return flow.ParseJSONBody[auth.User](fctx, reqCtx, "user")
        }).
        AddStep("create", func(fctx *flow.Context, reqCtx *request.Context) error {
            user := flow.GetTyped[auth.User](fctx, "user")
            // business logic...
            return reqCtx.OkCreated(response)
        }).
        AsHandler()
}
```

## Advanced Features

### Custom Validation Step
```go
func ValidateUserStep() flow.Step {
    return flow.NewStep("validate_user", func(fctx *flow.Context, reqCtx *request.Context) error {
        user := flow.GetTyped[auth.User](fctx, "user")
        
        if user.Email == "" {
            return reqCtx.ErrorBadRequest("Email is required")
        }
        
        if !isValidEmail(user.Email) {
            return reqCtx.ErrorBadRequest("Invalid email format")
        }
        
        return nil
    })
}
```

### Reusable Business Logic
```go
func CreateUserWithNotificationHandler() request.HandlerFunc {
    return flow.NewFlow("CreateUserWithNotification").
        AddStep("parse", func(fctx *flow.Context, reqCtx *request.Context) error {
            return flow.ParseJSONBody[auth.User](fctx, reqCtx, "user")
        }).
        AddSteps(
            ValidateUserStep(),
            CreateUserStep(),
            SendWelcomeEmailStep(),
        ).
        AddStep("response", func(fctx *flow.Context, reqCtx *request.Context) error {
            user := flow.GetTyped[auth.User](fctx, "created_user")
            return reqCtx.OkCreated(map[string]interface{}{
                "message": "User created and welcome email sent",
                "user":    user,
            })
        }).
        AsHandler()
}
```

Pattern ini memberikan:
- ✅ **Consistent DX**: Same pattern untuk semua endpoints
- ✅ **Type Safety**: Compile-time type checking
- ✅ **Error Handling**: Unified error response format  
- ✅ **Testability**: Easy to unit test individual steps
- ✅ **Reusability**: Reusable steps dan helpers
- ✅ **Maintainability**: Clear separation of concerns
