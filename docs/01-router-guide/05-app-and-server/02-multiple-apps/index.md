# Multiple Apps Example

Demonstrates running multiple apps on different ports with a server.

## What You'll Learn

- Create multiple apps on different ports
- Use Server to manage all apps
- Graceful shutdown for all apps together

## Running

```bash
# Navigate to example directory
cd docs/01-essentials/05-app-and-server/02-multiple-apps

# Run directly (go.mod already exists in project root)
go run main.go
```

## Testing

Use the included `test.http` file with VS Code REST Client extension, or use curl:

Open 3 terminal windows:

**Terminal 1 - API:**
```bash
curl http://localhost:8081/health
curl http://localhost:8081/users
```

**Terminal 2 - Admin:**
```bash
curl http://localhost:8082/dashboard
curl http://localhost:8082/users
```

**Terminal 3 - Metrics:**
```bash
curl http://localhost:8083/metrics
```

> 💡 **Tip**: Open `test.http` in VS Code for interactive testing with the REST Client extension!

## Output

```
🚀 Server starting all apps...
  API:     http://localhost:8081
  Admin:   http://localhost:8082
  Metrics: http://localhost:8083

🛑 Press Ctrl+C to stop all
```

Press `Ctrl+C` to shutdown all apps gracefully.
