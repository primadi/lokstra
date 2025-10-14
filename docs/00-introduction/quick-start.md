# Quick Start Guide

> **Build your first Lokstra API in 5 minutes**

---

## 📋 Prerequisites

Before starting, make sure you have:

```bash
# Go 1.21 or higher
go version
# go version go1.21.0 or later
```

---

## 🚀 Step 1: Create Project

```bash
# Create project directory
mkdir my-first-api
cd my-first-api

# Initialize Go module
go mod init my-first-api

# Install Lokstra
go get github.com/primadi/lokstra@latest
```

---

## 📝 Step 2: Hello World (Minimal)

Create `main.go`:

```go
package main

import (
    "github.com/primadi/lokstra"
    "time"
)

func main() {
    // 1. Create router
    r := lokstra.NewRouter("api")
    
    // 2. Add route
    r.GET("/ping", func() string {
        return "pong"
    })
    
    // 3. Create app and run
    app := lokstra.NewApp("demo", ":3000", r)
    app.Run(30 * time.Second)
}
```

**Run it:**
```bash
go run main.go
# 🚀 Server starting on http://localhost:3000
```

**Test it:**
```bash
curl http://localhost:3000/ping
# "pong"
```

**✅ Congratulations!** You've built your first Lokstra API!

---

## 🎯 Step 3: Add More Routes

Let's build a simple User API:

```go
package main

import (
    "fmt"
    "github.com/primadi/lokstra"
    "time"
)

// Data models
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// In-memory storage
var users = []User{
    {ID: 1, Name: "Alice", Email: "alice@example.com"},
    {ID: 2, Name: "Bob", Email: "bob@example.com"},
}
var nextID = 3

func main() {
    r := lokstra.NewRouter("api")
    
    // GET /users - List all users
    r.GET("/users", func() ([]User, error) {
        return users, nil
    })
    
    // GET /users/{id} - Get one user
    r.GET("/users/{id}", getUser)
    
    // POST /users - Create user
    r.POST("/users", createUser)
    
    // PUT /users/{id} - Update user
    r.PUT("/users/{id}", updateUser)
    
    // DELETE /users/{id} - Delete user
    r.DELETE("/users/{id}", deleteUser)
    
    app := lokstra.NewApp("user-api", ":3000", r)
    fmt.Println("🚀 User API running on http://localhost:3000")
    app.Run(30 * time.Second)
}

// Request/Response types
type GetUserRequest struct {
    ID int `path:"id"`
}

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

type UpdateUserRequest struct {
    ID    int    `path:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type DeleteUserRequest struct {
    ID int `path:"id"`
}

// Handlers
func getUser(req *GetUserRequest) (*User, error) {
    for _, u := range users {
        if u.ID == req.ID {
            return &u, nil
        }
    }
    return nil, fmt.Errorf("user not found")
}

func createUser(req *CreateUserRequest) (*User, error) {
    user := User{
        ID:    nextID,
        Name:  req.Name,
        Email: req.Email,
    }
    nextID++
    users = append(users, user)
    return &user, nil
}

func updateUser(req *UpdateUserRequest) (*User, error) {
    for i, u := range users {
        if u.ID == req.ID {
            if req.Name != "" {
                users[i].Name = req.Name
            }
            if req.Email != "" {
                users[i].Email = req.Email
            }
            return &users[i], nil
        }
    }
    return nil, fmt.Errorf("user not found")
}

func deleteUser(req *DeleteUserRequest) error {
    for i, u := range users {
        if u.ID == req.ID {
            users = append(users[:i], users[i+1:]...)
            return nil
        }
    }
    return fmt.Errorf("user not found")
}
```

---

## 🧪 Step 4: Test Your API

**List all users:**
```bash
curl http://localhost:3000/users
```
```json
[
  {"id":1,"name":"Alice","email":"alice@example.com"},
  {"id":2,"name":"Bob","email":"bob@example.com"}
]
```

**Get one user:**
```bash
curl http://localhost:3000/users/1
```
```json
{"id":1,"name":"Alice","email":"alice@example.com"}
```

**Create new user:**
```bash
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'
```
```json
{"id":3,"name":"Charlie","email":"charlie@example.com"}
```

**Update user:**
```bash
curl -X PUT http://localhost:3000/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Smith"}'
```
```json
{"id":1,"name":"Alice Smith","email":"alice@example.com"}
```

**Delete user:**
```bash
curl -X DELETE http://localhost:3000/users/2
```

---

## 📚 What Just Happened?

Let's break down the magic:

### 1. **Flexible Handler Signatures**
```go
// Simple return
r.GET("/users", func() ([]User, error) {
    return users, nil
})

// With request binding
r.POST("/users", func(req *CreateUserRequest) (*User, error) {
    return createUser(req)
})
```

Lokstra automatically:
- ✅ Binds request data to structs
- ✅ Validates with struct tags
- ✅ Encodes response to JSON
- ✅ Handles errors properly

### 2. **Auto Request Binding**
```go
type GetUserRequest struct {
    ID int `path:"id"`  // From URL path
}

type CreateUserRequest struct {
    Name  string `json:"name"`   // From JSON body
    Email string `json:"email"`
}
```

Tags:
- `path:"id"` - Extract from URL path
- `query:"page"` - Extract from query string
- `json:"name"` - Extract from JSON body
- `validate:"required"` - Validate field

### 3. **Auto JSON Response**
```go
func getUser(req *GetUserRequest) (*User, error) {
    return &user, nil  // Automatically becomes JSON
}
```

Lokstra converts:
- `*User` → JSON response with 200 OK
- `error` → JSON error with 500 (or custom code)
- `nil` error → Success response

---

## 🎨 Step 5: Add Middleware

Let's add logging:

```go
import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/middleware/request_logger"
)

func main() {
    r := lokstra.NewRouter("api")
    
    // Add logging middleware
    r.Use(request_logger.Middleware(nil))
    
    // Your routes...
    r.GET("/users", func() ([]User, error) {
        return users, nil
    })
    
    app := lokstra.NewApp("user-api", ":3000", r)
    app.Run(30 * time.Second)
}
```

**Now every request is logged:**
```
[INFO] GET /users 200 OK (5ms)
[INFO] POST /users 200 OK (12ms)
[INFO] GET /users/1 200 OK (3ms)
```

---

## 🔧 Step 6: Add CORS (Optional)

For frontend access:

```go
import (
    "github.com/primadi/lokstra/middleware/cors"
)

func main() {
    r := lokstra.NewRouter("api")
    
    // Add CORS
    corsConfig := map[string]any{
        "allow_origins": []string{"*"},
        "allow_methods": []string{"GET", "POST", "PUT", "DELETE"},
    }
    r.Use(
        request_logger.Middleware(nil),
        cors.Middleware(corsConfig),
    )
    
    // Routes...
    
    app := lokstra.NewApp("user-api", ":3000", r)
    app.Run(30 * time.Second)
}
```

---

## 📊 Complete Example

Here's the full working code:

<details>
<summary>Click to expand complete code</summary>

```go
package main

import (
    "fmt"
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/middleware/request_logger"
    "github.com/primadi/lokstra/middleware/cors"
    "time"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

var users = []User{
    {ID: 1, Name: "Alice", Email: "alice@example.com"},
    {ID: 2, Name: "Bob", Email: "bob@example.com"},
}
var nextID = 3

func main() {
    r := lokstra.NewRouter("api")
    
    // Middleware
    corsConfig := map[string]any{
        "allow_origins": []string{"*"},
        "allow_methods": []string{"GET", "POST", "PUT", "DELETE"},
    }
    r.Use(
        request_logger.Middleware(nil),
        cors.Middleware(corsConfig),
    )
    
    // Routes
    r.GET("/users", func() ([]User, error) {
        return users, nil
    })
    r.GET("/users/{id}", getUser)
    r.POST("/users", createUser)
    r.PUT("/users/{id}", updateUser)
    r.DELETE("/users/{id}", deleteUser)
    
    // Start
    app := lokstra.NewApp("user-api", ":3000", r)
    fmt.Println("🚀 User API running on http://localhost:3000")
    fmt.Println("📖 Try: curl http://localhost:3000/users")
    app.Run(30 * time.Second)
}

// Request types
type GetUserRequest struct {
    ID int `path:"id"`
}

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

type UpdateUserRequest struct {
    ID    int    `path:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type DeleteUserRequest struct {
    ID int `path:"id"`
}

// Handlers
func getUser(req *GetUserRequest) (*User, error) {
    for _, u := range users {
        if u.ID == req.ID {
            return &u, nil
        }
    }
    return nil, fmt.Errorf("user not found")
}

func createUser(req *CreateUserRequest) (*User, error) {
    user := User{
        ID:    nextID,
        Name:  req.Name,
        Email: req.Email,
    }
    nextID++
    users = append(users, user)
    return &user, nil
}

func updateUser(req *UpdateUserRequest) (*User, error) {
    for i, u := range users {
        if u.ID == req.ID {
            if req.Name != "" {
                users[i].Name = req.Name
            }
            if req.Email != "" {
                users[i].Email = req.Email
            }
            return &users[i], nil
        }
    }
    return nil, fmt.Errorf("user not found")
}

func deleteUser(req *DeleteUserRequest) error {
    for i, u := range users {
        if u.ID == req.ID {
            users = append(users[:i], users[i+1:]...)
            return nil
        }
    }
    return fmt.Errorf("user not found")
}
```

</details>

---

## 🎯 What You've Learned

In just 5 minutes, you've learned:

- ✅ Creating routers and routes
- ✅ Flexible handler signatures
- ✅ Automatic request binding
- ✅ Automatic JSON responses
- ✅ Error handling
- ✅ Adding middleware
- ✅ Building REST APIs

**And you have a working CRUD API!** 🎉

---

## 🚀 Next Steps

### Want to Learn More?

**Systematic Learning (Recommended)**:
👉 [Essentials Tutorial](../01-essentials/README.md) - Complete step-by-step guide

**Specific Topics**:
- [Router Deep Dive](../01-essentials/01-router/README.md) - All routing features
- [Services](../01-essentials/02-service/README.md) - Organize business logic
- [Middleware](../01-essentials/03-middleware/README.md) - Request processing
- [Configuration](../01-essentials/04-configuration/README.md) - YAML configs

**Real Examples**:
👉 [Complete Examples](../05-examples/README.md) - Production-ready apps

---

## 💡 Tips for Beginners

### 1. **Start Simple**
```go
// Begin with this
r.GET("/users", func() ([]User, error) {
    return getUsers()
})

// Add complexity as needed
r.GET("/users", func(ctx *request.Context, req *GetUsersRequest) (*response.Response, error) {
    // Full control
})
```

### 2. **Use Struct Tags**
```go
type Request struct {
    ID    int    `path:"id"`       // From URL
    Page  int    `query:"page"`    // From query string
    Name  string `json:"name"`     // From body
    Token string `header:"X-Token"` // From header
}
```

### 3. **Let Lokstra Handle Errors**
```go
// Simple - Lokstra converts to JSON error
func handler() (User, error) {
    if err := validate(); err != nil {
        return User{}, err  // Auto 500 error
    }
    return user, nil
}
```

### 4. **Print Routes for Debugging**
```go
r := lokstra.NewRouter("api")
// ... add routes ...

r.PrintRoutes()  // See all registered routes
```

Output:
```
[api] GET /users -> api.GET[users]
[api] POST /users -> api.POST[users]
[api] GET /users/{id} -> api.GET[users_id]
```

---

## 🆘 Having Issues?

### Common Problems:

**"Cannot find package"**
```bash
go get github.com/primadi/lokstra@latest
go mod tidy
```

**"Port already in use"**
```go
// Change port
app := lokstra.NewApp("demo", ":3001", r)  // Use 3001
```

**"Handler signature error"**
Check supported forms in [Router Guide](../01-essentials/01-router/README.md)

---

## 📚 Resources

- **Documentation**: [Full Docs](../README.md)
- **Examples**: [Code Examples](../05-examples/README.md)
- **API Reference**: [API Docs](../03-api-reference/README.md)
- **GitHub**: [Report Issues](https://github.com/primadi/lokstra/issues)

---

**Happy coding!** 🚀

**Next**: [Learn Essentials](../01-essentials/README.md) or [Understand Architecture](architecture.md)
