# Dynamic Configuration

Configuration that can be updated at runtime without restart.

## Running

```bash
go run main.go
```

Server starts on `http://localhost:3040`

## Features

- Thread-safe configuration store
- Auto-refresh every 15 seconds
- Manual update endpoint
- Real-time config changes

## Endpoints

- `GET /status` - View current configuration
- `POST /update` - Trigger configuration update

## Use Cases

- Feature flags
- Rate limits
- A/B testing
- Emergency configuration changes

## Pattern

```go
type ConfigStore struct {
    mu     sync.RWMutex
    values map[string]string
}

func (cs *ConfigStore) Get(key string) string {
    cs.mu.RLock()
    defer cs.mu.RUnlock()
    return cs.values[key]
}
```
