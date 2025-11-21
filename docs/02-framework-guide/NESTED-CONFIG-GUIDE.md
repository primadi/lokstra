# Nested Config Access Guide

Lokstra mendukung **nested configuration** dengan tiga cara akses:

## 1. Flat Access (Dot Notation)

Akses langsung ke leaf value menggunakan dot notation:

```go
dsn := lokstra_registry.GetConfig("global-db.dsn", "")
schema := lokstra_registry.GetConfig("global-db.schema", "public")
```

**YAML:**
```yaml
configs:
  global-db:
    dsn: "postgres://user:password@localhost:5432/mydb"
    schema: "public"
```

## 2. Nested Access (Map)

Akses ke seluruh nested object sebagai map:

```go
// Returns: map[string]any{"dsn": "...", "schema": "..."}
dbConfig := lokstra_registry.GetConfig[map[string]any]("global-db", nil)

if dbConfig != nil {
    dsn := dbConfig["dsn"].(string)
    schema := dbConfig["schema"].(string)
}
```

## 3. Struct Binding (Automatic) ⭐ NEW!

Konversi otomatis dari map ke struct menggunakan `cast.ToStruct`:

```go
type DBConfig struct {
    DSN    string `json:"dsn"`
    Schema string `json:"schema"`
}

// Automatic conversion!
dbConfig := lokstra_registry.GetConfig[DBConfig]("global-db", DBConfig{})
fmt.Printf("DSN: %s, Schema: %s\n", dbConfig.DSN, dbConfig.Schema)

// With pointer
dbConfigPtr := lokstra_registry.GetConfig[*DBConfig]("global-db", nil)
if dbConfigPtr != nil {
    fmt.Printf("DSN: %s\n", dbConfigPtr.DSN)
}
```

## Use Cases

### 1. Simple Values (Flat Access)

Gunakan untuk akses cepat ke single value:

```go
apiKey := lokstra_registry.GetConfig("api.key", "")
timeout := lokstra_registry.GetConfig("server.timeout", 30)
```

### 2. Complex Objects (Nested Access)

Gunakan untuk konfigurasi yang kompleks:

```go
// YAML:
// configs:
//   database:
//     host: "localhost"
//     port: 5432
//     credentials:
//       user: "admin"
//       password: "secret"

dbConfig := lokstra_registry.GetConfig[map[string]any]("database", nil)
credentials := dbConfig["credentials"].(map[string]any)
user := credentials["user"].(string)
```

### 3. Service Factory Configuration

Dalam service factory, Anda bisa akses nested config:

```go
func UserServiceFactory(deps map[string]any, config map[string]any) any {
    // Flat access
    dsn := lokstra_registry.GetConfig("global-db.dsn", "")
    
    // Nested access
    dbConfig := lokstra_registry.GetConfig[map[string]any]("global-db", nil)
    
    return &UserServiceImpl{
        DSN: dsn,
        // ...
    }
}
```

## Deep Nesting

Mendukung nested config dengan kedalaman arbitrary:

```yaml
configs:
  app:
    server:
      http:
        host: "localhost"
        port: 8080
      grpc:
        host: "localhost"
        port: 9090
```

**Flat access:**
```go
httpPort := lokstra_registry.GetConfig("app.server.http.port", 8080)
grpcPort := lokstra_registry.GetConfig("app.server.grpc.port", 9090)
```

**Nested access:**
```go
// Get entire server config
serverConfig := lokstra_registry.GetConfig[map[string]any]("app.server", nil)

// Get just HTTP config
httpConfig := lokstra_registry.GetConfig[map[string]any]("app.server.http", nil)
```

## Best Practices

1. **Use flat access untuk simple values** - Lebih performant dan type-safe
2. **Use nested access untuk complex objects** - Lebih fleksibel untuk iterasi
3. **Always provide default values** - Hindari panic saat config tidak ditemukan
4. **Type assertion dengan check** - Gunakan comma-ok pattern untuk safety

```go
// ❌ Bad: Panic jika type tidak match
dsn := dbConfig["dsn"].(string)

// ✅ Good: Safe type assertion
if dsn, ok := dbConfig["dsn"].(string); ok {
    // Use dsn
}
```

## Implementation Details

Configs disimpan secara **flattened** di internal registry:

```
"global-db.dsn" -> "postgres://..."
"global-db.schema" -> "public"
```

Saat akses nested (e.g., `GetConfig("global-db", ...)`), sistem akan:
1. Cari semua keys dengan prefix `"global-db."`
2. Reconstruct nested map structure
3. Return map hasil reconstruction

Ini memberikan **best of both worlds**:
- ✅ Efficient flat access
- ✅ Flexible nested access
- ✅ No data duplication
