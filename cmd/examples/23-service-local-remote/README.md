# Example 23: Service Local/Remote Auto-Detection

Contoh yang mendemonstrasikan **automatic local/remote service resolution** berdasarkan deployment configuration.

## 🎯 Problem yang Diselesaikan

### **Sebelum (Manual):**
```go
// ❌ Developer harus tahu apakah service local atau remote
var userService UserService
if os.Getenv("MODE") == "local" {
    userService = &LocalUserService{db: db}
} else {
    client := api_client.New(os.Getenv("USER_SERVICE_URL"))
    userService = &RemoteUserService{client: client}
}
```

### **Sesudah (Automatic):**
```go
// ✅ Framework auto-detect berdasarkan deployment!
userService := GetService[UserService]("user-service", nil)
user, err := userService.GetUser(ctx, req) // Magic! 🪄
```

---

## 📋 Konsep

### **1. Register 2 Factories (Local + Remote)**

```go
// Factory untuk LOCAL - real implementation
func CreateLocalUserService(config map[string]any) any {
    return &LocalUserService{db: getDB()}
}

// Factory untuk REMOTE - HTTP client wrapper
func CreateRemoteUserService(config map[string]any) any {
    client := GetClientRouter("UserService", nil)
    return &RemoteUserService{client: client}
}

// Register BOTH
RegisterServiceFactoryLocal("user", CreateLocalUserService)
RegisterServiceFactoryRemote("user", CreateRemoteUserService)
```

### **2. Framework Auto-Detects**

```go
// Framework decision logic in GetServiceFactory():
client := GetClientRouter("UserService", nil)

if client != nil && client.IsLocal {
    // Same server → use LOCAL factory
    return entry.localFactory
} else {
    // Different server → use REMOTE factory
    return entry.remoteFactory
}
```

### **3. Developer Gets Transparent Service**

```go
// Works sama baik local maupun remote!
userService := GetService[UserService]("user-service", nil)
user, err := userService.GetUser(ctx, req)
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    REGISTRATION PHASE                        │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  RegisterServiceFactoryLocal("user", CreateLocal)            │
│  RegisterServiceFactoryRemote("user", CreateRemote)          │
│                                                               │
│  RegisterLazyService("user-service", "user", config)         │
│                                                               │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ Framework Decision
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    RUNTIME RESOLUTION                         │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  GetService() → GetServiceFactory() → Check ClientRouter     │
│                                                               │
│  if ClientRouter.IsLocal:                                    │
│    ✅ Call LocalFactory  → &LocalUserService{}               │
│  else:                                                        │
│    ✅ Call RemoteFactory → &RemoteUserService{client}        │
│                                                               │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ Transparent Usage
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    USAGE PHASE                               │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  userService.GetUser(ctx, req)                               │
│                                                               │
│  Local:  → Query database directly                           │
│  Remote: → HTTP POST /users/{id}                             │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## 🚀 Cara Menjalankan

```bash
# Run example
go run main.go
```

### Expected Output:

```
======================================================================
🧪 Service Local/Remote Example
======================================================================

📋 Step 1: Registering service factories...
   ✅ Local factory registered
   ✅ Remote factory registered

📋 Step 2: Registering lazy service...
   ✅ Lazy service registered

📋 Step 3: Creating router from service...
   ✅ Router created with auto-generated routes:
      GET /users/{id}
      GET /users
      POST /users

📋 Step 4: Registering router...
   ✅ Router registered

📋 Step 5: Framework auto-detection...

   🔍 Current server: server-a
   🔍 UserService router: server-a
   ✅ Decision: LOCAL (same server)

📋 Step 6: Testing LOCAL service...

🏭 Factory: Creating LOCAL UserService
📍 LOCAL: GetUser called for ID=123
   ✅ Result: ID=123, Name=Local User 123, Email=user123@local.com

📋 Step 7: Simulating REMOTE scenario...

   🔍 Current server: server-a
   🔍 UserService router: server-b
   ✅ Decision: REMOTE (different server)

======================================================================
📝 Summary:
======================================================================

✅ Developer Benefits:
   1. Register factories once (local + remote)
   2. Framework auto-detects based on deployment
   3. Same interface for local and remote
   4. No manual if/else logic needed
   5. Error if factory not registered (fail-fast)

✅ Framework Decision Logic:
   - ClientRouter.IsLocal = true  → Use LOCAL factory
   - ClientRouter.IsLocal = false → Use REMOTE factory
   - No ClientRouter found        → Default to LOCAL

✅ Error Handling:
   - Need local but only remote registered → PANIC
   - Need remote but only local registered → PANIC
```

---

## 💡 Implementation Details

### **Service Interface**

```go
type UserService interface {
    GetUser(ctx *request.Context, req *GetUserRequest) (*User, error)
    ListUsers(ctx *request.Context) ([]*User, error)
    CreateUser(ctx *request.Context, req *CreateUserRequest) (*User, error)
}
```

### **Local Implementation**

```go
type LocalUserService struct {
    db *sql.DB
}

func (s *LocalUserService) GetUser(ctx *request.Context, req *GetUserRequest) (*User, error) {
    // Real database logic
    var user User
    err := s.db.QueryRow("SELECT * FROM users WHERE id = $1", req.UserID).Scan(&user)
    return &user, err
}
```

### **Remote Implementation**

```go
type RemoteUserService struct {
    client *api_client.ClientRouter
}

func (s *RemoteUserService) GetUser(ctx *request.Context, req *GetUserRequest) (*User, error) {
    // HTTP call to remote service
    path := "/users/" + req.UserID
    return api_client.FetchAndCast[User](ctx, s.client, path)
}
```

---

## 🎓 Key Takeaways

### **1. Separation of Concerns**
- ✅ Local implementation: business logic only
- ✅ Remote implementation: HTTP client only
- ✅ Framework: routing decision only

### **2. Fail-Fast Philosophy**
```go
if entry.localFactory == nil {
    panic("service requires local factory but only remote is registered")
}
```
Jika factory tidak tersedia, **panic saat startup** (bukan runtime error).

### **3. Configuration-Driven**
```go
// Deployment config menentukan routing
RegisterClientRouter("UserService", "server-a", ...)  // IsLocal = true
RegisterClientRouter("UserService", "server-b", ...)  // IsLocal = false
```

### **4. Type-Safe**
```go
// Generic type checking
userService := GetService[UserService]("user-service", nil)
// Compile error jika type mismatch!
```

---

## 📚 Related Examples

- Example 18: Service Router basics
- Example 20: Service Router with struct-based parameters
- Example 22: All handler forms test

---

## 🔧 Advanced Usage

### **Multiple Service Modes**

```go
// Register di init.go
func init() {
    RegisterServiceFactoryLocal("user", CreateLocalUserService)
    RegisterServiceFactoryRemote("user", CreateRemoteUserService)
    
    RegisterServiceFactoryLocal("product", CreateLocalProductService)
    RegisterServiceFactoryRemote("product", CreateRemoteProductService)
}

// Config via YAML
services:
  - name: user-service
    type: user
    # Framework auto-detects local/remote based on router registration
    
  - name: product-service
    type: product
    # Framework auto-detects local/remote based on router registration
```

### **Testing with Mock**

```go
// In tests, override with mock
type MockUserService struct{}
func (s *MockUserService) GetUser(...) (*User, error) {
    return &User{ID: "mock"}, nil
}

// Register mock factory
RegisterServiceFactoryLocal("user", func(cfg map[string]any) any {
    return &MockUserService{}
}, AllowOverride(true))

// Tests use same GetService() call!
userService := GetService[UserService]("user-service", nil)
```

---

## ✅ Checklist untuk Production

- [x] Register local factory untuk setiap service
- [x] Register remote factory untuk setiap service
- [x] Service interface consistent antara local dan remote
- [x] Router registered dengan correct server name
- [x] ClientRouter registered untuk semua routers
- [x] Error handling untuk missing factories
- [x] Tests untuk both local and remote modes

---

**Happy Coding! 🚀**
