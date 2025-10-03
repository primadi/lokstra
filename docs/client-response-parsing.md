# Client Response Parsing

## Overview

Lokstra framework menyediakan mekanisme standar untuk parsing HTTP response dari external/internal services menggunakan `ResponseFormatter` interface. Setiap formatter dapat mengimplementasikan cara parsing yang berbeda sesuai dengan format response yang diharapkan.

## ResponseFormatter Interface

Interface `ResponseFormatter` telah ditambahkan method baru:

```go
type ResponseFormatter interface {
    // ... existing methods ...
    
    // ParseClientResponse parses HTTP response into ClientResponse
    ParseClientResponse(resp *http.Response, cr *ClientResponse) error
}
```

## ClientResponse Struct

`ClientResponse` adalah struktur data yang menampung hasil parsing dari HTTP response:

```go
type ClientResponse struct {
    Status     string         `json:"status,omitempty"`      // "success" or "error"
    Message    string         `json:"message,omitempty"`     // Response message
    Data       any            `json:"data,omitempty"`        // Response data
    Error      *Error         `json:"error,omitempty"`       // Error details
    Meta       *Meta          `json:"meta,omitempty"`        // Metadata/pagination
    StatusCode int            `json:"status_code,omitempty"` // HTTP status code
    RawBody    []byte         `json:"-"`                     // Raw response body
    Headers    map[string]any `json:"headers,omitempty"`     // Response headers
}
```

## Built-in Formatters

Framework menyediakan 2 built-in formatters yang sudah diregister secara otomatis:

### 1. Default Formatter (ApiResponseFormatter)

Format response terstruktur dengan envelope:

**Success Response:**
```json
{
    "status": "success",
    "message": "Operation successful",
    "data": {
        "id": 123,
        "name": "John Doe"
    },
    "meta": {
        "page": 1,
        "page_size": 10,
        "total": 100
    }
}
```

**Error Response:**
```json
{
    "status": "error",
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Invalid input",
        "fields": [
            {
                "field": "email",
                "code": "INVALID_FORMAT",
                "message": "Email format is invalid"
            }
        ]
    }
}
```

### 2. Simple Formatter (SimpleResponseFormatter)

Format response sederhana tanpa envelope yang rigid:

**Success Response:**
```json
{
    "id": 123,
    "name": "John Doe"
}
```

atau dengan wrapper opsional:
```json
{
    "data": {"id": 123},
    "message": "Success"
}
```

**Error Response:**
```json
{
    "error": "Something went wrong",
    "code": "ERROR_CODE"
}
```

## Registrasi Formatters

Built-in formatters diregister otomatis saat package di-load:

```go
func init() {
    // Register built-in formatters
    RegisterFormatter("default", NewApiResponseFormatter)
    RegisterFormatter("simple", NewSimpleResponseFormatter)
}
```

## Custom Formatter

Programmer dapat menambahkan custom formatter dengan memanggil `RegisterFormatter`:

```go
// Define custom formatter
type MyCustomFormatter struct{}

func NewMyCustomFormatter() ResponseFormatter {
    return &MyCustomFormatter{}
}

func (f *MyCustomFormatter) ParseClientResponse(resp *http.Response, cr *ClientResponse) error {
    // Custom parsing logic
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    
    // Your custom parsing here
    // ...
    
    return nil
}

// ... implement other ResponseFormatter methods ...

// Register custom formatter
func init() {
    RegisterFormatter("my-custom", NewMyCustomFormatter)
}
```

## Usage Examples

### Example 1: Using Default Formatter

```go
package main

import (
    "net/http"
    "github.com/primadi/lokstra/core/response/api_formatter"
)

func main() {
    // Make HTTP request
    resp, err := http.Get("https://api.example.com/users/123")
    if err != nil {
        panic(err)
    }
    
    // Parse response using default formatter
    formatter := api_formatter.NewApiResponseFormatter()
    clientResp := &api_formatter.ClientResponse{}
    
    if err := formatter.ParseClientResponse(resp, clientResp); err != nil {
        panic(err)
    }
    
    // Access parsed data
    if clientResp.Status == "success" {
        fmt.Printf("Data: %+v\n", clientResp.Data)
        fmt.Printf("Message: %s\n", clientResp.Message)
    } else if clientResp.Error != nil {
        fmt.Printf("Error: %s - %s\n", clientResp.Error.Code, clientResp.Error.Message)
    }
}
```

### Example 2: Using Simple Formatter

```go
func main() {
    resp, err := http.Get("https://api.example.com/items")
    if err != nil {
        panic(err)
    }
    
    // Parse response using simple formatter
    formatter := api_formatter.NewSimpleResponseFormatter()
    clientResp := &api_formatter.ClientResponse{}
    
    if err := formatter.ParseClientResponse(resp, clientResp); err != nil {
        panic(err)
    }
    
    // Access parsed data
    fmt.Printf("Status: %s\n", clientResp.Status)
    fmt.Printf("Data: %+v\n", clientResp.Data)
}
```

### Example 3: Using Global Formatter

```go
func main() {
    // Set global formatter by name
    api_formatter.SetGlobalFormatterByName("simple")
    
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        panic(err)
    }
    
    // Use global formatter
    formatter := api_formatter.GetGlobalFormatter()
    clientResp := &api_formatter.ClientResponse{}
    
    if err := formatter.ParseClientResponse(resp, clientResp); err != nil {
        panic(err)
    }
    
    fmt.Printf("Data: %+v\n", clientResp.Data)
}
```

### Example 4: Using with ClientRouter

```go
func main() {
    // Assume you have a ClientRouter instance
    resp, err := clientRouter.GET("/api/users/123")
    if err != nil {
        panic(err)
    }
    
    // Parse with chosen formatter
    formatter := api_formatter.CreateFormatter("default")
    clientResp := &api_formatter.ClientResponse{}
    
    if err := formatter.ParseClientResponse(resp, clientResp); err != nil {
        panic(err)
    }
    
    // Handle response
    if clientResp.Status == "success" {
        user := clientResp.Data.(map[string]any)
        fmt.Printf("User: %s\n", user["name"])
    }
}
```

## Parsing Logic Details

### ApiResponseFormatter Parsing

1. Membaca seluruh response body
2. Menyimpan raw body dan HTTP status code
3. Parse headers ke dalam map
4. Mencoba parse JSON sebagai `ApiResponse`
5. Jika gagal, treat sebagai raw data
6. Map fields dari `ApiResponse` ke `ClientResponse`

### SimpleResponseFormatter Parsing

1. Membaca seluruh response body
2. Menyimpan raw body dan HTTP status code
3. Parse headers ke dalam map
4. Mencoba parse JSON sebagai generic `map[string]any`
5. Jika gagal, treat sebagai plain text
6. Detect error berdasarkan field "error"
7. Extract data, message, dan meta jika ada
8. Jika tidak ada struktur khusus, seluruh JSON menjadi data

## Error Handling

Kedua formatter menangani error dengan cara yang berbeda:

**Default Formatter:**
- Mengharapkan structure error yang lengkap dengan `code`, `message`, `details`, dan `fields`
- Jika response tidak sesuai format, data disimpan sebagai raw string

**Simple Formatter:**
- Lebih fleksibel dalam mendeteksi error
- Cek keberadaan field "error" untuk menentukan error response
- Support berbagai format error yang lebih sederhana

## Best Practices

1. **Pilih Formatter yang Sesuai**: Gunakan formatter yang sesuai dengan format API yang akan di-consume
2. **Handle Errors**: Selalu check error dari `ParseClientResponse` dan juga check `ClientResponse.Status`
3. **Check Status Code**: Gunakan `ClientResponse.StatusCode` untuk HTTP status code checks
4. **Type Assertion**: Lakukan type assertion saat mengakses `ClientResponse.Data` untuk tipe yang spesifik
5. **Custom Formatter**: Buat custom formatter jika API menggunakan format yang unique
6. **Global Formatter**: Set global formatter di aplikasi startup untuk consistency

## Integration dengan Existing Code

Method `ParseClientResponse` dapat diintegrasikan dengan existing HTTP clients:

```go
// Integration dengan ClientRouter
func (c *ClientRouter) GetParsed(path string, formatter ResponseFormatter) (*ClientResponse, error) {
    resp, err := c.GET(path)
    if err != nil {
        return nil, err
    }
    
    cr := &ClientResponse{}
    if err := formatter.ParseClientResponse(resp, cr); err != nil {
        return nil, err
    }
    
    return cr, nil
}

// Usage
clientResp, err := clientRouter.GetParsed("/api/users", api_formatter.NewApiResponseFormatter())
```

## Summary

Dengan menambahkan `ParseClientResponse` method ke `ResponseFormatter` interface:

1. ✅ Standarisasi cara parsing HTTP response
2. ✅ Support multiple response formats (default, simple)
3. ✅ Extensible - mudah tambah custom formatter
4. ✅ Consistent API untuk semua formatters
5. ✅ Built-in formatters sudah registered otomatis
6. ✅ Rich error handling dan metadata extraction
