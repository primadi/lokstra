# Response Return Types Example

This example demonstrates the new flexibility in handler return types introduced in Lokstra framework.

## What's New?

Handlers can now return:
- `*response.Response` - For full control over response (status, headers, body, content-type)
- `*response.ApiHelper` - For API-formatted responses with custom headers
- `response.Response` - Value type (not pointer)
- `response.ApiHelper` - Value type (not pointer)

## Running the Example

```bash
cd cmd_draft/examples/response-return-types
go run main.go
```

Server will start at `http://localhost:8080`

## Testing

### Option 1: Using VS Code REST Client (Recommended)
We provide a comprehensive `test.http` file for easy testing in VS Code:

1. Install [REST Client extension](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)
2. Open `test.http` in VS Code
3. Click "Send Request" above any HTTP request
4. View results in the response panel

**See `TESTING.md` for detailed instructions.**

The `test.http` file includes:
- ✅ All 6 endpoint groups (20+ requests)
- ✅ Header inspection tests
- ✅ Error scenario tests
- ✅ Edge case tests
- ✅ Quick smoke test suite

### Option 2: Using cURL
Test from command line with cURL (see examples below).

## Endpoints

### Group 1: Regular Data Returns
Standard API responses (existing behavior):
- `GET /regular/user` - Returns user object
- `GET /regular/users` - Returns array of users

### Group 2: Response Pointer Returns
Full control over HTTP response:
- `POST /response/created` - Custom 201 status code
- `GET /response/text` - Plain text response
- `GET /response/html` - HTML response
- `GET /response/stream` - Server-sent events (streaming)
- `GET /response/custom-headers` - Response with custom headers
- `GET /response/nil` - Nil response (default success)

### Group 3: ApiHelper Returns
API-formatted responses:
- `GET /api-helper/success` - Standard success response
- `POST /api-helper/create` - Created with message (201)
- `GET /api-helper/list` - List with pagination metadata
- `GET /api-helper/not-found` - 404 error
- `GET /api-helper/unauthorized` - 401 error
- `GET /api-helper/validation-error` - Validation error with field details
- `GET /api-helper/custom-headers` - API response with custom headers

### Group 4: Error Handling Priority
Demonstrates error precedence:
- `GET /error-priority/response-error` - Error takes precedence over Response
- `GET /error-priority/api-error` - Error takes precedence over ApiHelper

### Group 5: Without Context Parameter
Handlers without `*request.Context`:
- `GET /no-context/greet?name=John` - Struct-only parameter
- `GET /no-context/ping` - No parameters at all

### Group 6: Mixed Examples
Advanced patterns:
- `GET /mixed/response-value` - Response value (not pointer)
- `GET /mixed/api-value` - ApiHelper value (not pointer)
- `GET /mixed/conditional?format=html` - Conditional response type (HTML)
- `GET /mixed/conditional?format=text` - Conditional response type (text)
- `GET /mixed/conditional?format=json` - Conditional response type (JSON)

## Testing with cURL

```bash
# Regular data
curl http://localhost:8080/regular/user

# Plain text response
curl http://localhost:8080/response/text

# HTML response
curl http://localhost:8080/response/html

# Streaming (watch events)
curl -N http://localhost:8080/response/stream

# Custom headers (check headers)
curl -v http://localhost:8080/response/custom-headers

# API list with pagination
curl http://localhost:8080/api-helper/list

# Validation error
curl http://localhost:8080/api-helper/validation-error

# Error priority (error wins)
curl http://localhost:8080/error-priority/response-error

# Struct-only handler
curl "http://localhost:8080/no-context/greet?name=Alice"

# Conditional response types
curl http://localhost:8080/mixed/conditional?format=html
curl http://localhost:8080/mixed/conditional?format=text
curl http://localhost:8080/mixed/conditional?format=json
```

## Key Concepts

### 1. Error Takes Precedence
```go
func Handler() (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(200).Json(data)  // IGNORED!
    return resp, errors.New("failed") // ERROR returned
}
```

### 2. Nil Response Handling
```go
func Handler() (*response.Response, error) {
    return nil, nil  // Sends default success: Api.Ok(nil)
}
```

### 3. Full Control with Response
```go
func Handler() (*response.Response, error) {
    resp := response.NewResponse()
    resp.RespHeaders = map[string][]string{
        "X-Custom": {"value"},
    }
    resp.WithStatus(201).Json(data)
    return resp, nil
}
```

### 4. API Format with ApiHelper
```go
func Handler() (*response.ApiHelper, error) {
    api := response.NewApiHelper()
    api.Resp().RespHeaders = map[string][]string{
        "X-API-Version": {"v1.0"},
    }
    api.Ok(data)
    return api, nil
}
```

## Benefits

1. **Flexibility** - Choose control level based on needs
2. **Type Safety** - Compile-time type checking
3. **Backward Compatible** - Existing handlers still work
4. **Performance** - Type detection once at registration
5. **Consistent** - Return Response/ApiHelper == accessing c.Resp/c.Api

## Documentation

- **Testing Guide**: `TESTING.md` - How to use test.http file
- Full documentation: `docs_draft/response-return-types.md`
- Quick reference: `docs_draft/RESPONSE-RETURN-TYPES-QUICKREF.md`
- Summary: `docs_draft/RESPONSE-RETURN-TYPES-SUMMARY.md`

## Files in This Example

- `main.go` - Full working example with all handler patterns
- `test.http` - VS Code REST Client test suite (20+ requests)
- `TESTING.md` - Detailed testing guide
- `README.md` - This file

## See Also

- `core/router/helper.go` - Implementation
- `core/router/helper_response_test.go` - Tests
- `core/response/response.go` - Response struct
- `core/response/api_helper.go` - ApiHelper

---

**Enjoy the flexibility!** 🚀
