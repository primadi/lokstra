# RPC Client Example - Testing Various Return Types

Client example yang mendemonstrasikan cara menggunakan semua return types dari RPC server di `04_server_rpc`.

## 🎯 **Overview**

Client ini terhubung ke RPC server dan menguji semua 12 methods dengan berbagai return types:

1. **String Return** - `Hello(name string) (string, error)`
2. **Interface Return** - `GetUser(id int) (UserIface, error)`
3. **Slice Interface Return** - `GetUsers(limit int) ([]UserIface, error)`
4. **Map Return** - `GetUserStats(id int) (map[string]interface{}, error)`
5. **Struct Return** - `GetSystemInfo() (SystemInfo, error)`
6. **Primitive Returns** - `GetUserCount() (int, error)`, `GetUserActive(id int) (bool, error)`, `GetServerTime() (time.Time, error)`
7. **Dynamic Return** - `GetDynamicData(dataType string) (interface{}, error)`
8. **Void Operations** - `DeleteUser(id int) error`, `ClearCache() error`, `Ping() error`

## 🚀 **Cara Menjalankan**

### 1. Start RPC Server (Terminal 1)
```bash
cd cmd/examples/01_basic_overview/04_server_rpc
go run main.go
```

Tunggu sampai server output:
```
🚀 RPC Server Example - Various Return Types
📋 Server starting on :8080
```

### 2. Run RPC Client (Terminal 2)
```bash
cd cmd/examples/01_basic_overview/05_client_rpc
go run main.go
```

## 📋 **Expected Output**

```
🚀 RPC Client Example - Testing Various Return Types
🔗 Connecting to server at http://localhost:8080/rpc

🔍 Testing server connectivity...
✅ Server is responsive!

==================== STRING RETURN TYPE ====================
✅ Hello("World") → Hello, World!
✅ Hello("Lokstra") → Hello, Lokstra!
✅ Hello("Go Developer") → Hello, Go Developer!
❌ Hello("") Error: remote error: name cannot be empty

==================== INTERFACE RETURN TYPE ====================
✅ GetUser(123) → User{ID:123, Name:User-123, Email:user123@example.com, Active:true}
✅ GetUser(456) → User{ID:456, Name:User-456, Email:user456@example.com, Active:true}
❌ GetUser(0) Error: remote error: invalid user ID: 0
❌ GetUser(-1) Error: remote error: invalid user ID: -1

==================== SLICE INTERFACE RETURN TYPE ====================
✅ GetUsers(3) → 3 users
   [0] ID:1, Name:User-1, Active:true
   [1] ID:2, Name:User-2, Active:false
   [2] ID:3, Name:User-3, Active:true
✅ GetUsers(5) → 5 users
   [0] ID:1, Name:User-1, Active:true
   [1] ID:2, Name:User-2, Active:false
   [2] ID:3, Name:User-3, Active:true
   ... and 2 more users
❌ GetUsers(0) Error: remote error: limit must be between 1 and 100
❌ GetUsers(101) Error: remote error: limit must be between 1 and 100

==================== MAP RETURN TYPE ====================
✅ GetUserStats(123) → 6 fields
   user_id: 123 (type: int8)
   login_count: 42 (type: int8)
   last_login: 1737889234 (type: int64)
   is_premium: true (type: bool)
   balance: 123.45 (type: float64)
   achievements: [first_login complete_profile power_user] (type: []interface {})

==================== STRUCT RETURN TYPE ====================
✅ GetSystemInfo → SystemInfo
   Version: 1.0.0
   Uptime: 5 days
   Memory: 512MB
   CPU Usage: 25.5%
   Connected: 42

==================== PRIMITIVE RETURN TYPES ====================
✅ GetUserCount → 1337
✅ GetUserActive(124) → true
✅ GetUserActive(125) → false
❌ GetUserActive(0) Error: remote error: invalid user ID: 0
✅ GetServerTime → 2025-01-25 14:23:54

==================== DYNAMIC INTERFACE{} RETURN TYPE ====================
✅ GetDynamicData(user) → map[string]interface {}
   Map with 5 keys
✅ GetDynamicData(stats) → map[string]interface {}
   Map with 4 keys
✅ GetDynamicData(message) → string
   String: "This is a dynamic string message"
✅ GetDynamicData(number) → int8
   Value: 42
✅ GetDynamicData(list) → []interface {}
   Slice with 3 items
❌ GetDynamicData(unknown) Error: remote error: unknown data type: unknown

==================== VOID OPERATIONS (ERROR ONLY) ====================
✅ Ping → Success
✅ ClearCache → Success
✅ DeleteUser(999) → Success
❌ DeleteUser(0) Error: remote error: invalid user ID: 0

======================================================================
🎉 All return type examples completed successfully!
📖 This demonstrates the full range of return types supported by Lokstra RPC:
   • string, error
   • interface, error (UserIface → *User)
   • []interface, error ([]UserIface → []*User)
   • map[string]interface{}, error
   • struct, error (SystemInfo)
   • primitive types, error (int, bool, time.Time)
   • interface{}, error (dynamic types)
   • error only (void operations)
======================================================================
```

## 🔧 **Client Implementation Details**

### **1. Type-Safe Client Methods**
```go
type GreetingServiceClient struct {
    client *rpc_service.RpcClient
}

// String return
func (c *GreetingServiceClient) Hello(name string) (string, error) {
    result, err := c.client.Call("Hello", []interface{}{name})
    // Handle type conversion...
}

// Interface return
func (c *GreetingServiceClient) GetUser(id int) (*User, error) {
    var user User
    err := c.client.CallAndUnmarshal("GetUser", &user, id)
    return &user, err
}
```

### **2. Type Conversion Handling**
- **MessagePack Type Mapping**: `int8`, `int16`, `int32`, `int64` → `int`
- **Interface Slicing**: `[]interface{}` → `[]*User`
- **Map Handling**: Direct `map[string]interface{}` usage
- **Time Parsing**: Multiple format support
- **Error Propagation**: Server errors properly forwarded

### **3. Error Handling**
- **Connectivity Check**: Ping server before running tests
- **Graceful Degradation**: Continue testing other methods if one fails
- **Detailed Error Messages**: Clear indication of what went wrong
- **Server Instructions**: Help user start server if not running

## 🎓 **Learning Points**

### **1. Interface Return Types**
Server returns `UserIface`, client receives concrete `*User`:
```go
// Server
func GetUser(id int) (UserIface, error) {
    return &User{...}, nil  // Concrete type
}

// Client  
func GetUser(id int) (*User, error) {
    var user User
    err := client.CallAndUnmarshal("GetUser", &user, id)
    return &user, nil  // Convert to concrete type
}
```

### **2. MessagePack Type Handling**
```go
// Handle msgpack int variations
switch v := result.(type) {
case int: return v, nil
case int8: return int(v), nil
case int16: return int(v), nil
case int32: return int(v), nil 
case int64: return int(v), nil
}
```

### **3. Slice Interface Conversion**
```go
// Convert []interface{} to []*User
if resultSlice, ok := result.([]interface{}); ok {
    var users []*User
    for _, item := range resultSlice {
        // Convert each item...
    }
}
```

## 🔗 **Integration with Server**

Client perfectly integrates dengan server di `04_server_rpc`:

| Server Method | Client Method | Return Type | Status |
|---------------|---------------|-------------|--------|
| `Hello` | `Hello` | `string, error` | ✅ |
| `GetUser` | `GetUser` | `UserIface → *User` | ✅ |
| `GetUsers` | `GetUsers` | `[]UserIface → []*User` | ✅ |
| `GetUserStats` | `GetUserStats` | `map[string]interface{}` | ✅ |
| `GetSystemInfo` | `GetSystemInfo` | `SystemInfo` | ✅ |
| `GetUserCount` | `GetUserCount` | `int` | ✅ |
| `GetUserActive` | `GetUserActive` | `bool` | ✅ |
| `GetServerTime` | `GetServerTime` | `time.Time` | ✅ |
| `GetDynamicData` | `GetDynamicData` | `interface{}` | ✅ |
| `DeleteUser` | `DeleteUser` | `error` only | ✅ |
| `ClearCache` | `ClearCache` | `error` only | ✅ |
| `Ping` | `Ping` | `error` only | ✅ |

## 🎯 **Best Practices Demonstrated**

1. **Type Safety** - Proper type conversion dan validation
2. **Error Handling** - Comprehensive error checking dan reporting
3. **Code Organization** - Clean separation of concerns
4. **Documentation** - Clear examples dan explanations
5. **User Experience** - Helpful error messages dan instructions
6. **Protocol Compliance** - Full compatibility dengan Lokstra RPC

Client ini adalah contoh production-ready untuk menggunakan Lokstra RPC framework! 🚀
