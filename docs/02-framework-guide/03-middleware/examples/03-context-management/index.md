# Context Management Example

Learn how to store and retrieve request-scoped data using the context system.

## Running

```bash
go run main.go
```

Server starts on `http://localhost:3002`

## Key Patterns

### Storing Data

```go
c.Set("key", value)
```

### Retrieving Data

```go
value := c.Get("key")
```

## Use Cases

- User authentication data
- Request metadata
- Cross-middleware communication
- Request tracing
