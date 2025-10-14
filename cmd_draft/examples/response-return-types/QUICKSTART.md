# VS Code REST Client - Quick Start

## 1. Install Extension

Open VS Code and install **REST Client**:
- Press `Ctrl+Shift+X` (Extensions)
- Search: "REST Client"
- Install: `humao.rest-client`
- Restart VS Code

## 2. Start Server

```bash
go run main.go
```

Server listening on `http://localhost:8080`

## 3. Open test.http

Open `test.http` in VS Code. You'll see something like:

```http
### Get single user (regular data return)
GET http://localhost:8080/regular/user
Accept: application/json
```

## 4. Send Request

Three ways to send:

### Method A: Click Link
Click **"Send Request"** text above the request

### Method B: Keyboard
- Place cursor on request
- Press `Ctrl+Alt+R` (Windows/Linux) or `Cmd+Alt+R` (Mac)

### Method C: Command Palette
- `Ctrl+Shift+P` → "Rest Client: Send Request"

## 5. View Results

Response appears in a new panel showing:
- **Status Code**: `200 OK`, `201 Created`, etc.
- **Headers**: All HTTP headers
- **Body**: Response content (JSON, HTML, text, etc.)

## Quick Test Flow

### Test #1: Regular Handler
```http
GET http://localhost:8080/regular/user
```
Click "Send Request" → Should see JSON with user data

### Test #2: Custom Headers
```http
GET http://localhost:8080/response/custom-headers
```
Click "Send Request" → Check response headers for `X-Custom-Header`

### Test #3: Error Priority
```http
GET http://localhost:8080/error-priority/response-error
```
Click "Send Request" → Should see error (NOT success)

### Test #4: Pagination
```http
GET http://localhost:8080/api-helper/list
```
Click "Send Request" → Check `meta` field for pagination

## Common Use Cases

### Check Headers
After sending request, scroll to **Response Headers** section:
```
HTTP/1.1 200 OK
Content-Type: application/json
X-Custom-Header: custom-value
X-Request-ID: req-123456
```

### Test Different Formats
```http
# HTML
GET http://localhost:8080/mixed/conditional?format=html

# Text
GET http://localhost:8080/mixed/conditional?format=text

# JSON
GET http://localhost:8080/mixed/conditional?format=json
```

### Test Streaming
```http
GET http://localhost:8080/response/stream
```
Watch data stream in real-time!

### Test Query Parameters
```http
GET http://localhost:8080/no-context/greet?name=Alice
```
Change `Alice` to any name and re-send

## Tips

1. **Multiple Requests**: Use `###` to separate requests
2. **Variables**: Define once, reuse everywhere:
   ```http
   @baseUrl = http://localhost:8080
   
   GET {{baseUrl}}/regular/user
   ```
3. **Request History**: Recent requests saved automatically
4. **Cancel Request**: `Ctrl+Alt+K` (or `Cmd+Alt+K`)

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "Send Request" not showing | Install REST Client extension |
| Connection refused | Start server: `go run main.go` |
| Response not updating | Restart server |

## File Structure

```
response-return-types/
├── main.go           # Server code
├── test.http         # ← Test file (open this)
├── README.md         # Overview
└── TESTING.md        # Detailed guide
```

## Next Steps

1. ✅ Install REST Client extension
2. ✅ Start server: `go run main.go`
3. ✅ Open `test.http`
4. ✅ Click "Send Request"
5. 🎉 View results!

For more details, see `TESTING.md`

---

**That's it! Start testing!** 🚀
