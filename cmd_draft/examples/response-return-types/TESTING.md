# Testing with test.http

This directory contains a comprehensive `test.http` file for testing all response return types functionality using VS Code REST Client extension.

## Prerequisites

### 1. Install REST Client Extension
Install the REST Client extension in VS Code:
- Extension ID: `humao.rest-client`
- Or search "REST Client" in VS Code Extensions marketplace
- [Extension Page](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)

### 2. Start the Server
```bash
go run main.go
```

Server will start at `http://localhost:8080`

## Using test.http

### Method 1: Click "Send Request"
1. Open `test.http` in VS Code
2. You'll see "Send Request" links above each HTTP request
3. Click "Send Request" to execute the request
4. Results appear in a new panel on the right

### Method 2: Keyboard Shortcut
1. Place cursor on any request
2. Press:
   - **Windows/Linux**: `Ctrl+Alt+R`
   - **macOS**: `Cmd+Alt+R`

### Method 3: Command Palette
1. Place cursor on any request
2. Open Command Palette (`Ctrl+Shift+P` or `Cmd+Shift+P`)
3. Type "Rest Client: Send Request"
4. Press Enter

## Test Categories

### 1. Regular Data Returns (Lines 8-16)
Test standard API responses that are auto-wrapped with `Api.Ok()`:
- `GET /regular/user` - Single user object
- `GET /regular/users` - Array of users

### 2. Response Pointer Returns (Lines 22-49)
Test full-control responses with custom status, headers, content-type:
- `POST /response/created` - 201 status code
- `GET /response/text` - Plain text
- `GET /response/html` - HTML response
- `GET /response/stream` - Server-sent events (streaming)
- `GET /response/custom-headers` - Custom HTTP headers
- `GET /response/nil` - Nil pointer handling

### 3. ApiHelper Returns (Lines 55-86)
Test API-formatted responses with standardized structure:
- `GET /api-helper/success` - Standard success
- `POST /api-helper/create` - Created (201)
- `GET /api-helper/list` - Paginated list with metadata
- `GET /api-helper/not-found` - 404 error
- `GET /api-helper/unauthorized` - 401 error
- `GET /api-helper/validation-error` - Validation errors
- `GET /api-helper/custom-headers` - Custom headers

### 4. Error Handling Priority (Lines 92-104)
Test that errors always take precedence:
- `GET /error-priority/response-error` - Error wins over Response
- `GET /error-priority/api-error` - Error wins over ApiHelper

### 5. Without Context Parameter (Lines 110-127)
Test handlers without `*request.Context`:
- `GET /no-context/greet?name=Alice` - Struct-only parameter
- `GET /no-context/greet` - Default value handling
- `GET /no-context/ping` - No parameters

### 6. Mixed Examples (Lines 133-159)
Test advanced patterns:
- `GET /mixed/response-value` - Response value (not pointer)
- `GET /mixed/api-value` - ApiHelper value (not pointer)
- `GET /mixed/conditional?format=html` - Conditional HTML
- `GET /mixed/conditional?format=text` - Conditional text
- `GET /mixed/conditional?format=json` - Conditional JSON

### Advanced Testing (Lines 165+)
- Header inspection tests
- Error scenarios
- Content-type variations
- Pagination metadata
- Parameter binding
- Edge cases
- Performance checks
- Quick smoke test

## Testing Tips

### Inspect Response Headers
After sending a request, check the response panel for:
- Status code (top line)
- Headers section (middle)
- Body section (bottom)

Example - Custom Headers Test:
```http
GET http://localhost:8080/response/custom-headers
```
Expected headers:
- `X-Custom-Header: custom-value`
- `X-Request-ID: req-123456`
- `X-Rate-Limit: 1000`
- `X-Rate-Remaining: 999`

### Test Streaming Responses
The streaming endpoint will output events over a few seconds:
```http
GET http://localhost:8080/response/stream
```
Watch the response panel as data streams in real-time.

### Compare Response Structures

**Regular return** (auto-wrapped):
```json
{
  "status": "success",
  "data": { "id": 1, "name": "John Doe" }
}
```

**ApiHelper return** (explicit):
```json
{
  "status": "success",
  "message": "operation completed",
  "data": { ... }
}
```

### Test Error Priority
These requests should return errors (NOT success responses):
```http
GET http://localhost:8080/error-priority/response-error
GET http://localhost:8080/error-priority/api-error
```

Expected: Error response, even though Response/ApiHelper had success status.

## Common Scenarios

### Scenario 1: Test New Feature
1. Start server: `go run main.go`
2. Open `test.http`
3. Run smoke test (lines 321-349)
4. Verify all 6 groups work correctly

### Scenario 2: Debug Specific Handler
1. Find the relevant request in `test.http`
2. Send the request
3. Inspect response status, headers, body
4. Modify handler code if needed
5. Restart server and re-test

### Scenario 3: Compare Behaviors
1. Send regular handler request
2. Send Response handler request
3. Send ApiHelper handler request
4. Compare the response structures

### Scenario 4: Test Error Handling
1. Run error priority tests (lines 92-104)
2. Run error response tests (lines 210-234)
3. Verify error responses match expectations

## Response Validation

### Success Response (200 OK)
```json
{
  "status": "success",
  "data": { ... }
}
```

### Created Response (201)
```json
{
  "status": "success",
  "message": "Resource created successfully",
  "data": { ... }
}
```

### List Response with Pagination
```json
{
  "status": "success",
  "data": [ ... ],
  "meta": {
    "page": 1,
    "page_size": 10,
    "total": 50,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  }
}
```

### Error Response (4xx/5xx)
```json
{
  "status": "error",
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found"
  }
}
```

### Validation Error (400)
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": [
      { "field": "email", "message": "Invalid email format" },
      { "field": "password", "message": "Password too short" }
    ]
  }
}
```

## Troubleshooting

### "Connection Refused"
- Make sure server is running: `go run main.go`
- Check server is listening on port 8080
- Verify no firewall blocking

### "Send Request" Link Not Showing
- Install REST Client extension
- Restart VS Code
- Make sure file is saved as `.http`

### Response Not Updating
- Stop and restart the server
- Clear VS Code cache
- Check for code errors in server logs

## Keyboard Shortcuts

| Action | Windows/Linux | macOS |
|--------|--------------|-------|
| Send Request | `Ctrl+Alt+R` | `Cmd+Alt+R` |
| Cancel Request | `Ctrl+Alt+K` | `Cmd+Alt+K` |
| Re-run Last Request | `Ctrl+Alt+L` | `Cmd+Alt+L` |

## VS Code Settings (Optional)

Add to your `.vscode/settings.json`:

```json
{
  "rest-client.defaultHeaders": {
    "User-Agent": "vscode-restclient"
  },
  "rest-client.timeoutinmilliseconds": 10000,
  "rest-client.followredirect": true,
  "rest-client.previewOption": "full"
}
```

## Alternative: Using cURL

If you prefer command line, you can also test with cURL:

```bash
# Regular data
curl http://localhost:8080/regular/user

# Plain text
curl http://localhost:8080/response/text

# With headers inspection
curl -v http://localhost:8080/response/custom-headers

# Streaming
curl -N http://localhost:8080/response/stream

# Query parameters
curl "http://localhost:8080/no-context/greet?name=Alice"
```

## See Also

- Main README: `README.md`
- Full documentation: `../../docs_draft/response-return-types.md`
- Quick reference: `../../docs_draft/RESPONSE-RETURN-TYPES-QUICKREF.md`

---

**Happy Testing!** 🚀
