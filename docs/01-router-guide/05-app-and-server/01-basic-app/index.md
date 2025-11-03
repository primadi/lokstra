# Basic App Example

Demonstrates combining multiple routers into one app.

## What You'll Learn

- Create multiple routers
- Combine routers into one app
- Run app with graceful shutdown

## Running

```bash
# Navigate to example directory
cd docs/01-essentials/05-app-and-server/01-basic-app

# Run directly (go.mod already exists in project root)
go run main.go
```

## Testing

Use the included `test.http` file with VS Code REST Client extension, or use curl:

```bash
# Test endpoints
curl http://localhost:8080/users
curl http://localhost:8080/products
curl http://localhost:8080/stats
curl http://localhost:8080/logs
```

> 💡 **Tip**: Open `test.http` in VS Code for interactive testing with the REST Client extension!

## Output

```
🚀 Server starting on :8080
📋 Endpoints:
  GET /users
  GET /products
  GET /stats
  GET /logs

🛑 Press Ctrl+C to stop
```

Press `Ctrl+C` to see graceful shutdown in action.
