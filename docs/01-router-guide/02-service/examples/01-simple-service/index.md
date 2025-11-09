# Example 01: Simple Service

> **Learn service registration and basic access patterns**  
> **Time**: 10 minutes • **Concepts**: Service factory, registration, LazyLoad

---

## 🎯 What You'll Learn

- ✅ Define a service with business logic
- ✅ Create a service factory function
- ✅ Register service in the registry
- ✅ Access service with `service.LazyLoad()` 
- ✅ Use `MustGet()` for clear error messages

---

## 🚀 Run It

```bash
cd docs/01-router-guide/02-service/examples/01-simple-service
go run main.go
```

**Server starts on**: `http://localhost:3000`

---

## 📝 Code Walkthrough

### Step 1: Define Service

```go
type UserService struct {
    users  []User
    nextID int
}

func (s *UserService) GetAll() ([]User, error) {
    return s.users, nil
}

func (s *UserService) GetByID(id int) (*User, error) {
    for _, user := range s.users {
        if user.ID == id {
            return &user, nil
        }
    }
    return nil, fmt.Errorf("user not found")
}

func (s *UserService) Create(name, email, role string) (*User, error) {
    user := User{
        ID:    s.nextID,
        Name:  name,
        Email: email,
        Role:  role,
    }
    s.users = append(s.users, user)
    s.nextID++
    return &user, nil
}
```

**💭 Key Points**:
- Service contains business logic
- Methods return `(data, error)` for proper error handling
- In-memory storage for simplicity (use DB in production!)

---

### Step 2: Create Factory Function

```go
func NewUserService() *UserService {
    return &UserService{
        users: []User{
            {ID: 1, Name: "Alice", Email: "alice@example.com", Role: "admin"},
            {ID: 2, Name: "Bob", Email: "bob@example.com", Role: "user"},
        },
        nextID: 3,
    }
}
```

**💭 Key Points**:
- Factory initializes service with dependencies
- Can inject DB, cache, other services, etc.
- Called once during app initialization

---

### Step 3: Register Service

```go
func main() {
    // Create and register service instance
    userSvc := NewUserService()
    lokstra_registry.RegisterService("users", userSvc)
    
    // ... create router and handlers
}
```

**💭 Key Points**:
- Register **before** creating app
- Service available to all routers and handlers
- Can be accessed by name "users"

---

### Step 4: Access Service with LazyLoad

```go
// Package-level: Cached after first access!
var userService = service.LazyLoad[*UserService]("users")

r.GET("/users", func() (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    
    // Access service - only 1-5ns overhead after first call!
    users, err := userService.MustGet().GetAll()
    if err != nil {
        api.InternalError(err.Error())
        return api, nil
    }
    
    api.Ok(users)
    return api, nil
})
```

**💭 Key Points**:
- `LazyLoad` at **package-level** (not function-level!)
- `MustGet()` panics with clear error if service not found
- Cached after first access (20-100x faster than registry lookup!)

---

## 🧪 Test Endpoints

### List All Users
```bash
curl http://localhost:3000/users
```

**Response**:
```json
{
  "status": "success",
  "data": [
    {"id": 1, "name": "Alice", "email": "alice@example.com", "role": "admin"},
    {"id": 2, "name": "Bob", "email": "bob@example.com", "role": "user"},
    {"id": 3, "name": "Charlie", "email": "charlie@example.com", "role": "user"}
  ]
}
```

---

### Get User by ID
```bash
curl http://localhost:3000/users/1
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com",
    "role": "admin"
  }
}
```

---

### Create New User
```bash
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Dave",
    "email": "dave@example.com",
    "role": "user"
  }'
```

**Response**:
```json
{
  "status": "success",
  "message": "User created successfully",
  "data": {
    "id": 4,
    "name": "Dave",
    "email": "dave@example.com",
    "role": "user"
  }
}
```

---

### Update User
```bash
curl -X PUT http://localhost:3000/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Updated",
    "email": "alice.new@example.com",
    "role": "admin"
  }'
```

---

### Delete User
```bash
curl -X DELETE http://localhost:3000/users/3
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "message": "User deleted successfully"
  }
}
```

---

## 💡 Key Concepts

### 1. Service Pattern
Services encapsulate business logic:
- ✅ Reusable across handlers
- ✅ Testable independently
- ✅ Clear responsibility separation

### 2. Factory Pattern
Factory functions initialize services:
- ✅ Dependency injection
- ✅ Error handling during setup
- ✅ Configuration

### 3. LazyLoad Pattern
`service.LazyLoad()` provides cached access:
- ✅ **Fast**: 1-5ns after first access
- ✅ **Clear errors**: MustGet() panics with service name
- ✅ **Simple**: No manual caching needed

### 4. Registry Pattern
Global service registry:
- ✅ Centralized service management
- ✅ Name-based access
- ✅ Dependency resolution

---

## 🎯 Best Practices Demonstrated

### ✅ Package-Level LazyLoad
```go
// ✅ GOOD: Package-level, cached forever
var userService = service.LazyLoad[*UserService]("users")

func handler() {
    users := userService.MustGet().GetAll()
}
```

### ❌ Function-Level LazyLoad
```go
// ❌ BAD: Created every request, cache useless!
func handler() {
    userService := service.LazyLoad[*UserService]("users")
    users := userService.MustGet().GetAll()
}
```

---

### ✅ Use MustGet() for Clear Errors
```go
// ✅ GOOD: Clear error message
users := userService.MustGet().GetAll()
// If service not found: "service 'users' not found or not initialized"
```

### ❌ Using Get() Without Nil Check
```go
// ❌ BAD: Confusing nil pointer error
users := userService.Get().GetAll()
// If service not found: "runtime error: invalid memory address"
```

---

## 🔗 What's Next?

**Continue to**:
- [Example 02 - LazyLoad vs GetService](../02-lazyload-vs-getservice/) - Performance comparison
- [Example 03 - Service Dependencies](../03-service-dependencies/) - DI pattern
- [Example 04 - Service as Router](../04-service-as-router/) - Auto-generate endpoints

**Related**:
- [Main Service Guide](../..)
- [Router Guide](../../../01-router)

---

**Back**: [Service Guide](../..)  
**Next**: [02 - Performance Comparison](../02-lazyload-vs-getservice/)
