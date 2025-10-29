# Service Router Quick Start

## What is Service Router?

Service Router automatically generates HTTP routes from your service methods using naming conventions. No more manual route registration!

## Quick Example

### Step 1: Define Your Service

```go
type UserService struct {
    users map[string]*User
}

// Convention: GetUser -> GET /users/{id}
func (s *UserService) GetUser(ctx *request.Context, id string) (*User, error) {
    user, exists := s.users[id]
    if !exists {
        return nil, errors.New("user not found")
    }
    return user, nil
}

// Convention: ListUsers -> GET /users
func (s *UserService) ListUsers(ctx *request.Context) ([]*User, error) {
    var users []*User
    for _, user := range s.users {
        users = append(users, user)
    }
    return users, nil
}

// Convention: CreateUser -> POST /users
func (s *UserService) CreateUser(ctx *request.Context, req *CreateUserRequest) (*User, error) {
    user := &User{
        ID:    generateID(),
        Name:  req.Name,
        Email: req.Email,
    }
    s.users[user.ID] = user
    return user, nil
}
```

### Step 2: Create Router

```go
// That's it! No manual route registration needed.
router := router.NewFromService(
    &UserService{},
    router.DefaultServiceRouterOptions().
        WithPrefix("/api/v1"),
)
```

### Step 3: Run Server

```go
app := app.New("my-app", ":8080", router)
if err := app.Run(30 * time.Second); err != nil {
    fmt.Println("Error starting server:", err)
}
```

## Convention Rules

| Method Name | Maps To | Example |
|-------------|---------|---------|
| `GetUser` | `GET /users/{id}` | Get single user |
| `ListUsers` | `GET /users` | Get all users |
| `CreateUser` | `POST /users` | Create new user |
| `UpdateUser` | `PUT /users/{id}` | Update user |
| `DeleteUser` | `DELETE /users/{id}` | Delete user |
| `SearchUsers` | `GET /users/search` | Search users |

## Method Signatures

### List/Search (no parameters)
```go
func (s *Service) ListUsers(ctx *request.Context) ([]*User, error)
```

### Get/Update/Delete (with ID)
```go
func (s *Service) GetUser(ctx *request.Context, id string) (*User, error)
```

### Create (with request body)
```go
func (s *Service) CreateUser(ctx *request.Context, req *CreateUserRequest) (*User, error)
```

### Update (with ID and body)
```go
func (s *Service) UpdateUser(ctx *request.Context, id string, req *UpdateUserRequest) (*User, error)
```

## Options

```go
opts := router.DefaultServiceRouterOptions().
    WithPrefix("/api/v1").                    // Add prefix to all routes
    WithResourceName("person").               // Override resource name
    WithPluralResourceName("people").         // Custom pluralization
    WithRouteOverride("CustomMethod", router.RouteMeta{
        HTTPMethod: "POST",
        Path: "/custom/path",
    })

router := router.NewFromService(service, opts)
```

## Response Format

### Success Response
```json
{
  "success": true,
  "data": { ... }
}
```

### Error Response
```json
{
  "success": false,
  "error": "error message"
}
```

## Run the Example

```bash
cd cmd/examples/18-service-router
go run main.go
```

Then test:
```bash
curl http://localhost:8080/api/v1/users
```

## Benefits

- ✅ **95% less code** - No manual route registration
- ✅ **Type-safe** - Service methods define API
- ✅ **Consistent** - Conventions ensure uniform APIs
- ✅ **Testable** - Test service logic directly
- ✅ **Maintainable** - Changes reflect automatically

## When NOT to Use

- Routes don't follow REST conventions
- Need complex custom routing logic
- Migrating existing non-standard APIs

In these cases, use manual router registration as before.
