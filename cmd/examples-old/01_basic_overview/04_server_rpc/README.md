# RPC Server Example - Various Return Types

Contoh ini mendemonstrasikan berbagai jenis return types yang didukung oleh Lokstra RPC framework.

## 📋 Return Types yang Didukung

### 1. **String, Error**
```go
func Hello(name string) (string, error)
```
- Return: Primitive string type
- Use Case: Simple text responses

### 2. **Interface, Error** 
```go
func GetUser(id int) (UserIface, error)
```
- Return: Interface implementation (concrete struct)
- Use Case: Object-oriented responses dengan polymorphism

### 3. **Slice of Interface, Error**
```go
func GetUsers(limit int) ([]UserIface, error)
```
- Return: Array/slice dari interface implementations
- Use Case: List/collection responses

### 4. **Map, Error**
```go
func GetUserStats(id int) (map[string]any, error)
```
- Return: Key-value map dengan mixed types
- Use Case: Dynamic data structures, statistics

### 5. **Struct, Error**
```go
func GetSystemInfo() (SystemInfo, error)
```
- Return: Concrete struct type
- Use Case: Fixed data structures dengan typed fields

### 6. **Primitive Types, Error**
```go
func GetUserCount() (int, error)
func GetUserActive(id int) (bool, error) 
func GetServerTime() (time.Time, error)
```
- Return: Built-in Go types (int, bool, time.Time, etc.)
- Use Case: Simple data values

### 7. **any, Error**
```go
func GetDynamicData(dataType string) (any, error)
```
- Return: Any type (runtime-determined)
- Use Case: Dynamic responses, API flexibility

### 8. **Error Only (Void Operations)**
```go
func DeleteUser(id int) error
func ClearCache() error
func Ping() error
```
- Return: No data, only success/error status
- Use Case: Commands, operations tanpa return value

## 🚀 Cara Menjalankan

### 1. Start Server
```bash
cd cmd/examples/01_basic_overview/04_server_rpc
go run main.go
```

Server akan berjalan di `http://localhost:8080`

### 2. Test dengan Curl
```bash
# Test Hello method
curl -X POST http://localhost:8080/rpc \
  -H 'Content-Type: application/json' \
  -d '{"method": "Hello", "params": ["World"]}'

# Test GetUser method  
curl -X POST http://localhost:8080/rpc \
  -H 'Content-Type: application/json' \
  -d '{"method": "GetUser", "params": [123]}'

# Test void operation
curl -X POST http://localhost:8080/rpc \
  -H 'Content-Type: application/json' \
  -d '{"method": "Ping", "params": []}'
```

### 3. Test dengan Go Client
```go
// Import client
import "github.com/primadi/lokstra/modules/rpc_service"

// Create client
client := rpc_service.NewRpcClient("http://localhost:8080/rpc")

// Call methods
result, err := client.Call("Hello", []any{"World"})
```

### 4. Lihat Dokumentasi API
Buka browser: `http://localhost:8080/`

## 📖 Endpoints

| Method | URL | Return Type | Description |
|--------|-----|-------------|-------------|
| `GET /` | Documentation | JSON | API documentation |
| `GET /health` | Health check | JSON | Server status |
| `POST /rpc` | RPC endpoint | Various | All RPC methods |

## 🔗 RPC Methods

| Method | Parameters | Return Type | Example |
|--------|------------|-------------|---------|
| `Hello` | `[name: string]` | `string, error` | Basic greeting |
| `GetUser` | `[id: int]` | `UserIface, error` | User object |
| `GetUsers` | `[limit: int]` | `[]UserIface, error` | User list |
| `GetUserStats` | `[id: int]` | `map[string]any, error` | User statistics |
| `GetSystemInfo` | `[]` | `SystemInfo, error` | System information |
| `GetUserCount` | `[]` | `int, error` | Total users |
| `GetUserActive` | `[id: int]` | `bool, error` | User status |
| `GetServerTime` | `[]` | `time.Time, error` | Current time |
| `GetDynamicData` | `[type: string]` | `any, error` | Dynamic content |
| `DeleteUser` | `[id: int]` | `error` | Delete operation |
| `ClearCache` | `[]` | `error` | Cache clear |
| `Ping` | `[]` | `error` | Health ping |

## 📝 Protocol Details

- **Transport**: HTTP POST
- **Encoding**: MessagePack (msgpack)
- **URL Pattern**: `/rpc`
- **Method Mapping**: CamelCase → lowercase
- **Error Handling**: JSON error responses untuk HTTP errors

## 🎯 Key Learning Points

1. **Interface Support**: Server dapat return interface, client receive concrete struct
2. **Type Safety**: Client harus handle type conversion dari msgpack
3. **Error Handling**: Semua methods return error sebagai parameter terakhir
4. **Void Operations**: Methods bisa return hanya error (tanpa data)
5. **Flexible Returns**: any allows runtime type determination
6. **Protocol Compatibility**: Full compatibility dengan existing Lokstra RPC infrastructure

## 🛠️ Development Notes

- Service interface dan implementation terpisah untuk clean architecture
- Client examples menunjukkan best practices untuk handling berbagai return types
- Error handling comprehensive untuk production readiness
- Documentation lengkap untuk easy adoption
