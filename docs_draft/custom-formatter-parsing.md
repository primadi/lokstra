# Custom Formatter Response Parsing

## Overview

Lokstra framework mendukung custom formatters yang dapat di-extend dengan kemampuan parsing HTTP response. Setiap custom formatter harus mengimplementasikan method `ParseClientResponse` dari interface `ResponseFormatter`.

## Custom Formatter Examples

### 1. CustomCorporateFormatter

Format response enterprise/corporate dengan struktur yang rigid dan informatif.

#### Response Format

**Success Response:**
```json
{
  "responseCode": "00",
  "responseStatus": "SUCCESS",
  "payload": {
    "id": 123,
    "name": "John Doe"
  },
  "description": "Operation successful",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Error Response:**
```json
{
  "responseCode": "99",
  "responseStatus": "ERROR",
  "errorCode": "USER_NOT_FOUND",
  "errorMessage": "User not found",
  "errorDetails": {
    "userId": 999
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Validation Error Response:**
```json
{
  "responseCode": "98",
  "responseStatus": "VALIDATION_ERROR",
  "errorMessage": "Invalid input",
  "validationErrors": [
    {
      "field": "email",
      "code": "INVALID_FORMAT",
      "message": "Email format is invalid"
    }
  ],
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### Implementation

```go
type CustomCorporateFormatter struct{}

func (f *CustomCorporateFormatter) ParseClientResponse(resp *http.Response, cr *api_formatter.ClientResponse) error {
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed to read response body: %w", err)
    }

    cr.RawBody = body
    cr.StatusCode = resp.StatusCode
    cr.Headers = make(map[string]any)
    for key, values := range resp.Header {
        if len(values) == 1 {
            cr.Headers[key] = values[0]
        } else {
            cr.Headers[key] = values
        }
    }

    var result map[string]any
    if err := json.Unmarshal(body, &result); err != nil {
        cr.Status = "unknown"
        cr.Data = string(body)
        return nil
    }

    if responseStatus, ok := result["responseStatus"].(string); ok {
        if responseStatus == "SUCCESS" || responseStatus == "CREATED" {
            cr.Status = "success"
            if payload, hasPayload := result["payload"]; hasPayload {
                cr.Data = payload
            }
            if desc, hasDesc := result["description"]; hasDesc {
                cr.Message = fmt.Sprint(desc)
            }
            // Extract pagination if present
            if paginationInfo, hasPagination := result["paginationInfo"]; hasPagination {
                if paginationMap, ok := paginationInfo.(map[string]any); ok {
                    paginationBytes, _ := json.Marshal(paginationMap)
                    var listMeta api_formatter.ListMeta
                    if json.Unmarshal(paginationBytes, &listMeta) == nil {
                        cr.Meta = &api_formatter.Meta{ListMeta: &listMeta}
                    }
                }
            }
        } else if responseStatus == "ERROR" || responseStatus == "VALIDATION_ERROR" {
            cr.Status = "error"
            errorObj := &api_formatter.Error{}
            if errorCode, hasCode := result["errorCode"]; hasCode {
                errorObj.Code = fmt.Sprint(errorCode)
            }
            if errorMessage, hasMessage := result["errorMessage"]; hasMessage {
                errorObj.Message = fmt.Sprint(errorMessage)
            }
            if errorDetails, hasDetails := result["errorDetails"]; hasDetails {
                if detailsMap, ok := errorDetails.(map[string]any); ok {
                    errorObj.Details = detailsMap
                }
            }
            if validationErrors, hasValidation := result["validationErrors"]; hasValidation {
                if fieldsSlice, ok := validationErrors.([]any); ok {
                    errorObj.Fields = make([]api_formatter.FieldError, 0, len(fieldsSlice))
                    for _, field := range fieldsSlice {
                        if fieldMap, ok := field.(map[string]any); ok {
                            fe := api_formatter.FieldError{}
                            if f, ok := fieldMap["field"].(string); ok {
                                fe.Field = f
                            }
                            if c, ok := fieldMap["code"].(string); ok {
                                fe.Code = c
                            }
                            if m, ok := fieldMap["message"].(string); ok {
                                fe.Message = m
                            }
                            fe.Value = fieldMap["value"]
                            errorObj.Fields = append(errorObj.Fields, fe)
                        }
                    }
                }
            }
            cr.Error = errorObj
        }
    } else {
        cr.Status = "unknown"
        cr.Data = result
    }

    return nil
}
```

#### Parsing Logic

1. **Read Response Body**: Baca seluruh HTTP response body
2. **Store Metadata**: Simpan raw body, status code, dan headers
3. **Parse JSON**: Attempt to parse body sebagai JSON
4. **Detect Status**: Check field `responseStatus`:
   - `"SUCCESS"` atau `"CREATED"` → Success response
   - `"ERROR"` atau `"VALIDATION_ERROR"` → Error response
5. **Extract Data**:
   - Success: Extract `payload`, `description`, `paginationInfo`
   - Error: Extract `errorCode`, `errorMessage`, `errorDetails`, `validationErrors`
6. **Fallback**: Jika format tidak dikenali, store sebagai unknown

### 2. MobileApiFormatter

Format response yang dioptimalkan untuk mobile apps dengan payload minimal.

#### Response Format

**Success Response:**
```json
{
  "ok": true,
  "data": {
    "id": 456,
    "title": "Item"
  },
  "msg": "Success"
}
```

**Error Response:**
```json
{
  "ok": false,
  "error": "Something went wrong",
  "code": "ERROR_CODE",
  "info": {
    "detail": "Additional info"
  }
}
```

**Validation Error Response:**
```json
{
  "ok": false,
  "error": "Validation failed",
  "fields": [
    {
      "field": "email",
      "code": "REQUIRED",
      "message": "Email is required"
    }
  ]
}
```

#### Implementation

```go
type MobileApiFormatter struct{}

func (f *MobileApiFormatter) ParseClientResponse(resp *http.Response, cr *api_formatter.ClientResponse) error {
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed to read response body: %w", err)
    }

    cr.RawBody = body
    cr.StatusCode = resp.StatusCode
    cr.Headers = make(map[string]any)
    for key, values := range resp.Header {
        if len(values) == 1 {
            cr.Headers[key] = values[0]
        } else {
            cr.Headers[key] = values
        }
    }

    var result map[string]any
    if err := json.Unmarshal(body, &result); err != nil {
        cr.Status = "unknown"
        cr.Data = string(body)
        return nil
    }

    if ok, hasOk := result["ok"].(bool); hasOk {
        if ok {
            cr.Status = "success"
            if data, hasData := result["data"]; hasData {
                cr.Data = data
            }
            if msg, hasMsg := result["msg"]; hasMsg {
                cr.Message = fmt.Sprint(msg)
            }
            // Extract pagination if present
            if page, hasPage := result["page"]; hasPage {
                if pageMap, ok := page.(map[string]any); ok {
                    pageBytes, _ := json.Marshal(pageMap)
                    var listMeta api_formatter.ListMeta
                    if json.Unmarshal(pageBytes, &listMeta) == nil {
                        cr.Meta = &api_formatter.Meta{ListMeta: &listMeta}
                    }
                }
            }
        } else {
            cr.Status = "error"
            errorObj := &api_formatter.Error{}
            if errorMsg, hasError := result["error"]; hasError {
                errorObj.Message = fmt.Sprint(errorMsg)
            }
            if code, hasCode := result["code"]; hasCode {
                errorObj.Code = fmt.Sprint(code)
            }
            if info, hasInfo := result["info"]; hasInfo {
                if infoMap, ok := info.(map[string]any); ok {
                    errorObj.Details = infoMap
                }
            }
            if fields, hasFields := result["fields"]; hasFields {
                if fieldsSlice, ok := fields.([]any); ok {
                    errorObj.Fields = make([]api_formatter.FieldError, 0, len(fieldsSlice))
                    for _, field := range fieldsSlice {
                        if fieldMap, ok := field.(map[string]any); ok {
                            fe := api_formatter.FieldError{}
                            if f, ok := fieldMap["field"].(string); ok {
                                fe.Field = f
                            }
                            if c, ok := fieldMap["code"].(string); ok {
                                fe.Code = c
                            }
                            if m, ok := fieldMap["message"].(string); ok {
                                fe.Message = m
                            }
                            fe.Value = fieldMap["value"]
                            errorObj.Fields = append(errorObj.Fields, fe)
                        }
                    }
                }
            }
            cr.Error = errorObj
        }
    } else {
        cr.Status = "unknown"
        cr.Data = result
    }

    return nil
}
```

#### Parsing Logic

1. **Read Response Body**: Baca seluruh HTTP response body
2. **Store Metadata**: Simpan raw body, status code, dan headers
3. **Parse JSON**: Attempt to parse body sebagai JSON
4. **Detect Status**: Check field `ok`:
   - `true` → Success response
   - `false` → Error response
5. **Extract Data**:
   - Success: Extract `data`, `msg`, `page`
   - Error: Extract `error`, `code`, `info`, `fields`
6. **Fallback**: Jika format tidak dikenali, store sebagai unknown

## Usage Examples

### Example 1: Parse Corporate Response

```go
package main

import (
    "net/http"
    "github.com/primadi/lokstra/core/response/api_formatter"
)

func main() {
    // Register custom formatter
    api_formatter.RegisterFormatter("corporate", NewCustomCorporateFormatter)
    
    // Make HTTP request
    resp, err := http.Get("https://corporate-api.example.com/users/123")
    if err != nil {
        panic(err)
    }
    
    // Parse with corporate formatter
    formatter := api_formatter.CreateFormatter("corporate")
    clientResp := &api_formatter.ClientResponse{}
    
    if err := formatter.ParseClientResponse(resp, clientResp); err != nil {
        panic(err)
    }
    
    // Use parsed data
    if clientResp.Status == "success" {
        fmt.Printf("Payload: %+v\n", clientResp.Data)
        fmt.Printf("Description: %s\n", clientResp.Message)
    } else if clientResp.Error != nil {
        fmt.Printf("Error %s: %s\n", clientResp.Error.Code, clientResp.Error.Message)
    }
}
```

### Example 2: Parse Mobile Response

```go
func main() {
    // Register mobile formatter
    api_formatter.RegisterFormatter("mobile", NewMobileApiFormatter)
    
    // Make HTTP request
    resp, err := http.Get("https://mobile-api.example.com/items")
    if err != nil {
        panic(err)
    }
    
    // Parse with mobile formatter
    formatter := api_formatter.CreateFormatter("mobile")
    clientResp := &api_formatter.ClientResponse{}
    
    if err := formatter.ParseClientResponse(resp, clientResp); err != nil {
        panic(err)
    }
    
    // Use parsed data
    if clientResp.Status == "success" {
        items := clientResp.Data.([]map[string]any)
        fmt.Printf("Items: %+v\n", items)
    }
}
```

### Example 3: Handle Different Response Formats

```go
func parseResponse(resp *http.Response, formatterType string) (*api_formatter.ClientResponse, error) {
    formatter := api_formatter.CreateFormatter(formatterType)
    clientResp := &api_formatter.ClientResponse{}
    
    if err := formatter.ParseClientResponse(resp, clientResp); err != nil {
        return nil, err
    }
    
    return clientResp, nil
}

func main() {
    // Corporate API
    corpResp, _ := http.Get("https://corporate.example.com/data")
    corpData, _ := parseResponse(corpResp, "corporate")
    
    // Mobile API
    mobileResp, _ := http.Get("https://mobile.example.com/data")
    mobileData, _ := parseResponse(mobileResp, "mobile")
    
    // Default API
    defaultResp, _ := http.Get("https://api.example.com/data")
    defaultData, _ := parseResponse(defaultResp, "default")
}
```

## Best Practices

### 1. Consistent Error Handling

```go
func (f *MyFormatter) ParseClientResponse(resp *http.Response, cr *ClientResponse) error {
    defer resp.Body.Close()
    
    // Always store these
    cr.StatusCode = resp.StatusCode
    cr.RawBody = body
    cr.Headers = extractHeaders(resp.Header)
    
    // Parse with fallback
    var result map[string]any
    if err := json.Unmarshal(body, &result); err != nil {
        cr.Status = "unknown"
        cr.Data = string(body)
        return nil // Don't fail on parse error
    }
    
    // Your parsing logic here
    
    return nil
}
```

### 2. Type Assertions with Safety

```go
// Always use comma-ok idiom
if value, ok := result["field"].(string); ok {
    cr.Message = value
}

// For nested structures
if nested, ok := result["data"].(map[string]any); ok {
    if id, ok := nested["id"].(float64); ok {
        // Use id
    }
}
```

### 3. Flexible Field Detection

```go
// Support multiple field names
message := ""
if msg, ok := result["message"].(string); ok {
    message = msg
} else if msg, ok := result["msg"].(string); ok {
    message = msg
} else if desc, ok := result["description"].(string); ok {
    message = desc
}
cr.Message = message
```

### 4. Validation Error Extraction

```go
// Extract validation errors safely
if fields, hasFields := result["validationErrors"]; hasFields {
    if fieldsSlice, ok := fields.([]any); ok {
        errorObj.Fields = make([]api_formatter.FieldError, 0, len(fieldsSlice))
        for _, field := range fieldsSlice {
            if fieldMap, ok := field.(map[string]any); ok {
                fe := api_formatter.FieldError{}
                // Extract field properties safely
                if f, ok := fieldMap["field"].(string); ok {
                    fe.Field = f
                }
                // ... more fields
                errorObj.Fields = append(errorObj.Fields, fe)
            }
        }
    }
}
```

## Testing Custom Formatters

```go
package main

import (
    "net/http"
    "strings"
    "testing"
)

func TestCorporateFormatterParseSuccess(t *testing.T) {
    body := `{
        "responseCode": "00",
        "responseStatus": "SUCCESS",
        "payload": {"id": 123},
        "description": "Success"
    }`
    
    resp := &http.Response{
        StatusCode: 200,
        Body:       io.NopCloser(strings.NewReader(body)),
        Header:     make(http.Header),
    }
    
    formatter := NewCustomCorporateFormatter()
    cr := &api_formatter.ClientResponse{}
    
    err := formatter.ParseClientResponse(resp, cr)
    
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    if cr.Status != "success" {
        t.Errorf("Expected status 'success', got '%s'", cr.Status)
    }
    
    if cr.Message != "Success" {
        t.Errorf("Expected message 'Success', got '%s'", cr.Message)
    }
}

func TestCorporateFormatterParseError(t *testing.T) {
    body := `{
        "responseCode": "99",
        "responseStatus": "ERROR",
        "errorCode": "NOT_FOUND",
        "errorMessage": "Resource not found"
    }`
    
    resp := &http.Response{
        StatusCode: 404,
        Body:       io.NopCloser(strings.NewReader(body)),
        Header:     make(http.Header),
    }
    
    formatter := NewCustomCorporateFormatter()
    cr := &api_formatter.ClientResponse{}
    
    err := formatter.ParseClientResponse(resp, cr)
    
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    if cr.Status != "error" {
        t.Errorf("Expected status 'error', got '%s'", cr.Status)
    }
    
    if cr.Error == nil {
        t.Fatal("Expected error object, got nil")
    }
    
    if cr.Error.Code != "NOT_FOUND" {
        t.Errorf("Expected error code 'NOT_FOUND', got '%s'", cr.Error.Code)
    }
}
```

## Summary

### Key Points

1. **Custom Formatters**: Implement `ParseClientResponse` untuk support custom response formats
2. **Consistent Structure**: Semua parser harus populate `ClientResponse` dengan cara yang konsisten
3. **Error Handling**: Always handle JSON parse errors gracefully dengan fallback
4. **Type Safety**: Use comma-ok idiom untuk type assertions
5. **Flexible Parsing**: Support berbagai field names dan structures
6. **Testing**: Write tests untuk verify parsing logic

### Formatter Comparison

| Feature | Corporate | Mobile | Default | Simple |
|---------|-----------|--------|---------|--------|
| Status Field | `responseStatus` | `ok` | `status` | `error` or data |
| Data Field | `payload` | `data` | `data` | `data` or root |
| Error Field | `errorCode/Message` | `error/code` | `error.code/message` | `error` |
| Message Field | `description` | `msg` | `message` | `message` |
| Pagination | `paginationInfo` | `page` | `meta` | `meta` |
| Timestamp | ✅ | ❌ | ❌ | ❌ |
| Response Code | ✅ | ❌ | ❌ | ❌ |

### Implementation Checklist

- [x] Implement all ResponseFormatter interface methods
- [x] Add ParseClientResponse method
- [x] Handle success responses
- [x] Handle error responses
- [x] Handle validation errors
- [x] Extract pagination/metadata
- [x] Store raw body and headers
- [x] Fallback for unknown formats
- [x] Register formatter with meaningful name
- [x] Write unit tests
- [x] Document expected format
