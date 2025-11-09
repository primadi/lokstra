# Lokstra Router-Only Template

This is a simple template demonstrating how to use Lokstra as a router-only solution, without the full framework features.

## What's Included

This template demonstrates:

- **Basic Router Setup**: Creating and configuring a Lokstra router
- **Route Groups**: Organizing routes with path prefixes (`/users` and `/roles`)
- **RESTful Endpoints**: All HTTP methods (GET, POST, PUT, PATCH, DELETE)
- **Standard Middleware**: Using built-in recovery and slow request logger middleware
- **Custom Middleware**: Creating your own middleware for logging or other purposes
- **Request Handling**: Parsing request bodies, path parameters, and query strings
- **Response Formatting**: Using the API helper for consistent JSON responses
- **File Organization**: Separating concerns into multiple files

## Project Structure

```
.
├── main.go         # Application entry point
├── router.go       # Router setup and route definitions
├── middleware.go   # Custom middleware implementations
├── handlers.go     # Request handler functions
├── test.http       # HTTP test file for API testing
├── go.mod          # Go module dependencies
└── README.md       # This file
```

## Quick Start

### 1. Install Dependencies

```bash
go mod download
```

### 2. Run the Application

```bash
go run .
```

The server will start on `http://localhost:3000`

### 3. Test the Endpoints

We provide a `test.http` file with all API endpoints ready to test.

#### Option 1: Using VS Code REST Client Extension (Recommended)

1. Install the [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) extension in VS Code
2. Open `test.http` file
3. Click "Send Request" above any request to test it
4. View the response in the right panel

#### Option 2: Using curl Commands

You can also test using curl from the terminal:

**User Endpoints:**

```bash
# Get all users
curl http://localhost:3000/users

# Get a specific user
curl http://localhost:3000/users/123

# Create a new user
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Johnson","email":"alice@example.com"}'

# Update a user (full update)
curl -X PUT http://localhost:3000/users/123 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Johnson","email":"alice.johnson@example.com"}'

# Partially update a user
curl -X PATCH http://localhost:3000/users/123 \
  -H "Content-Type: application/json" \
  -d '{"email":"newemail@example.com","name":"Updated Name"}'

# Delete a user
curl -X DELETE http://localhost:3000/users/123
```

**Role Endpoints:**

```bash
# Get all roles
curl http://localhost:3000/roles

# Get a specific role
curl http://localhost:3000/roles/1

# Create a new role
curl -X POST http://localhost:3000/roles \
  -H "Content-Type: application/json" \
  -d '{"name":"Moderator","description":"Can moderate content"}'

# Update a role
curl -X PUT http://localhost:3000/roles/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Admin","description":"Full system access"}'

# Partially update a role
curl -X PATCH http://localhost:3000/roles/1 \
  -H "Content-Type: application/json" \
  -d '{"description":"Updated administrator role"}'

# Delete a role
curl -X DELETE http://localhost:3000/roles/1

# Assign role to user (nested resource example)
curl -X POST http://localhost:3000/roles/1/users/123
```

## Key Concepts

### Router Creation

```go
r := lokstra.NewRouter("demo_router")
```

Creates a new router instance with a descriptive name.

### Middleware

Middleware functions run before your handlers, useful for:
- Logging
- Authentication
- Error recovery
- Request validation

Apply middleware globally:
```go
r.Use(recovery.Middleware(recovery.DefaultConfig()))
r.Use(customLoggingMiddleware())
```

### Route Groups

Organize related routes under a common prefix:
```go
users := r.AddGroup("/users")
users.GET("", handleGetUsers)        // GET /users
users.GET("/:id", handleGetUser)     // GET /users/:id
users.POST("", handleCreateUser)     // POST /users
```

### Handler Functions

Lokstra supports automatic parameter binding and validation. Handlers can use struct parameters:

```go
type getUserParams struct {
    ID string `path:"id" validate:"required"`
}

func handleGetUser(p *getUserParams) (*User, error) {
    // Path parameters are automatically bound and validated
    user := &User{
        ID:    p.ID,
        Name:  "John Doe",
        Email: "john@example.com",
    }
    return user, nil
}
```

The framework automatically:
- Binds path parameters, query strings, and request body
- Validates input using struct tags
- Handles errors and returns appropriate responses
- Marshals return values to JSON

### Request Binding

Lokstra automatically binds and validates request data using struct tags:

```go
type createUserParams struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

func handleCreateUser(p *createUserParams) (*User, error) {
    // Data is already bound and validated
    user := &User{
        ID:    "123",
        Name:  p.Name,
        Email: p.Email,
    }
    return user, nil
}
```

**Wildcard Binding for Flexible Data:**

For PATCH operations or when you need to accept arbitrary JSON, use the `json:"*"` wildcard tag:

```go
type patchUserParams struct {
    ID      string         `path:"id" validate:"required"`
    Updates map[string]any `json:"*"` // Captures entire request body
}

func handlePatchUser(p *patchUserParams) (map[string]any, error) {
    // p.Updates contains all fields from request body
    result := map[string]any{
        "id":      p.ID,
        "updated": p.Updates,
    }
    return result, nil
}
```

Supported binding sources:
- `path:"id"` - Path parameters
- `json:"name"` - Request body (JSON)
- `json:"*"` - Entire request body as map[string]any (wildcard)
- `query:"page"` - Query parameters
- `header:"Authorization"` - HTTP headers

### Response Helpers

Handlers can return data directly or errors:

```go
// Return data directly (200 OK)
func handleGetUsers() ([]User, error) {
    users := []User{...}
    return users, nil
}

// Return error for error cases
func handleGetUser(p *getUserParams) (*User, error) {
    if userNotFound {
        return nil, fmt.Errorf("user not found")
    }
    return user, nil
}
```

The framework automatically:
- Wraps successful responses in standardized JSON format
- Converts errors to appropriate HTTP status codes
- Handles validation errors with detailed field information

## Customization

### Adding New Routes

1. Define the route in `router.go`:
```go
users.GET("/active", handleGetActiveUsers)
```

2. Create the handler in `handlers.go`:
```go
func handleGetActiveUsers() ([]User, error) {
    activeUsers := []User{...}
    return activeUsers, nil
}
```

### Creating Custom Middleware

See `middleware.go` for examples. Basic structure:
```go
func myMiddleware() request.HandlerFunc {
    return request.HandlerFunc(func(c *request.Context) error {
        // Before handler
        
        err := c.Next() // Call next handler
        
        // After handler
        return err
    })
}
```

### Changing the Port

Edit `main.go`:
```go
http.ListenAndServe(":8080", router) // Use port 8080 instead
```

## Next Steps

- Add database integration
- Implement authentication/authorization
- Add input validation
- Create more complex route structures
- Add tests for handlers

## Learn More

- [Lokstra Documentation](https://primadi.github.io/lokstra)

## Notes

This template uses mock data for demonstration. In a real application:
- Use a database for persistence
- Add proper error handling
- Implement authentication
- Add input validation
- Write unit and integration tests
