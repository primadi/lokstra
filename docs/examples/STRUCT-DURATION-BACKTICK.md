# Example: Struct with Duration Field - Auto Convert from YAML

## 1. Define Struct with Duration

```go
package application

import "time"

type ServerConfig struct {
    Host         string        `json:"host"`
    Port         int           `json:"port"`
    ReadTimeout  time.Duration `json:"read_timeout"`
    WriteTimeout time.Duration `json:"write_timeout"`
    IdleTimeout  time.Duration `json:"idle_timeout"`
}

type RetryConfig struct {
    MaxRetries int           `json:"max_retries"`
    Delay      time.Duration `json:"delay"`
    MaxDelay   time.Duration `json:"max_delay"`
}

// @RouterService name="app-service", prefix="/api"
type AppService struct {
    // @InjectCfgValue "server"
    Server ServerConfig
    
    // @InjectCfgValue "retry"
    Retry RetryConfig
}
```

## 2. YAML Config - Duration as String

```yaml
configs:
  server:
    host: "localhost"
    port: 8080
    read_timeout: "30s"      # ✅ String format - auto convert!
    write_timeout: "1m"      # ✅ String format
    idle_timeout: "2h"       # ✅ String format

  retry:
    max_retries: 3
    delay: "5s"              # ✅ String format
    max_delay: "1m30s"       # ✅ Complex duration string
```

## 3. How It Works

### At Code Generation Time:
```go
// Generated code uses cast.ToStruct
Server: func() ServerConfig {
    if v, ok := cfg["server"]; ok {
        var result ServerConfig
        if err := cast.ToStruct(v, &result); err == nil {
            return result
        }
    }
    return ServerConfig{}
}(),
```

### At Runtime:
- `cast.ToStruct()` reads YAML config as `map[string]any`
- Detects `time.Duration` field type
- If value is `string`, calls `time.ParseDuration()`
- If value is `int64`, treats as nanoseconds
- Converts automatically!

## 4. Default Values with Backtick

### Option A: Using Backtick (Recommended - No Escaping!)
```go
// @InjectCfgValue key="server", default=`ServerConfig{Host: "localhost", Port: 8080, ReadTimeout: 30*time.Second}`
Server ServerConfig
```

### Option B: Using Double Quotes (Need Escaping)
```go
// @InjectCfgValue key="server", default="ServerConfig{Host: \"localhost\", Port: 8080, ReadTimeout: 30*time.Second}"
Server ServerConfig
```

### Option C: Duration String Format
```go
// @InjectCfgValue key="timeout", default="15m"
Timeout time.Duration
```

## 5. Complete Example

```go
package application

import "time"

type DatabaseConfig struct {
    Host            string        `json:"host"`
    Port            int           `json:"port"`
    MaxConnections  int           `json:"max_connections"`
    ConnectTimeout  time.Duration `json:"connect_timeout"`
    QueryTimeout    time.Duration `json:"query_timeout"`
    IdleTimeout     time.Duration `json:"idle_timeout"`
}

// @RouterService name="user-service", prefix="/api/users"
type UserService struct {
    // @InjectCfgValue key="database", default=`DatabaseConfig{Host: "localhost", Port: 5432, MaxConnections: 10, ConnectTimeout: 5*time.Second, QueryTimeout: 30*time.Second, IdleTimeout: 10*time.Minute}`
    DB DatabaseConfig
}

// @Route "GET /{id}"
func (s *UserService) GetByID(id string) (string, error) {
    // s.DB.QueryTimeout will be properly set from YAML config
    // or use default value if not configured
    return "user", nil
}
```

```yaml
# config.yaml
configs:
  database:
    host: "prod-db.example.com"
    port: 5432
    max_connections: 50
    connect_timeout: "10s"    # ✅ Auto converts to time.Duration
    query_timeout: "1m"       # ✅ Auto converts
    idle_timeout: "30m"       # ✅ Auto converts

deployments:
  production:
    servers:
      api:
        addr: ":8080"
        published-services: [user-service]
```

## 6. Supported Duration Formats

All Go standard duration formats work:

- `"300ms"` → 300 milliseconds
- `"1.5s"` → 1.5 seconds
- `"30s"` → 30 seconds
- `"5m"` → 5 minutes
- `"2h"` → 2 hours
- `"1h30m"` → 1 hour 30 minutes
- `"2h45m30s"` → 2 hours 45 minutes 30 seconds

## 7. Benefits

✅ **Type-safe** - Compiler catches type mismatches  
✅ **Auto-conversion** - No manual parsing needed  
✅ **Clean syntax** - Use backtick to avoid escaping  
✅ **Flexible** - Support both string and numeric duration  
✅ **Nested structs** - Works with nested configuration  
✅ **Default values** - Full Go syntax support
