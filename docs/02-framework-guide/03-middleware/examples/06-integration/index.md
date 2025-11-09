# Third-Party Middleware Integration

Learn how to integrate external middleware libraries with Lokstra.

## Running

```bash
go run main.go
```

Server starts on `http://localhost:3005`

## Adapter Pattern

```go
func AdaptMiddleware(thirdPartyMw func(string, string) bool) func(*request.Context) error {
    return func(c *request.Context) error {
        if !thirdPartyMw(c.R.Method, c.R.URL.Path) {
            return fmt.Errorf("rejected")
        }
        return c.Next()
    }
}
```

## Use Cases

- Integrating existing middleware libraries
- Wrapping third-party auth systems
- Adapting legacy middleware
- Creating middleware bridges
