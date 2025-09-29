
# Lokstra API Standards

This document defines the **opinionated standard for API Request and Response** in Lokstra.  
The goal is to ensure **consistency**, **developer experience (DX)**, and **auto-UI readiness** across all services.

---

## 📌 API Request Standard

### Struct Definition

```go
type PagingRequest struct {
    Page       int      `query:"page"`        // default: 1
    PageSize   int      `query:"page_size"`   // default: 20, max: 100
    OrderBy    []string `query:"order_by"`    // e.g. order_by=id,-name
    QueryAll   bool     `query:"all"`         // true → ignore paging
    Fields     []string `query:"fields"`      // e.g. fields=id,name,email
    Search     string   `query:"search"`      // global keyword search
    Filters    []string `query:"filter"`      // e.g. filter=status:active
    DataType   string   `query:"data_type"`   // "list" | "table", default "list"
    DataFormat string   `query:"data_format"` // "json", "json_download", "csv", "xlsx"
    Download   bool     `query:"download"`    // true = force download, false = inline
}
```

### Rules

- **page/page_size** → standard pagination, default and max limit applied.  
- **order_by** → multi-field allowed, prefix `-` means descending.  
- **all** → if true, paging is ignored. Backend may limit with `MaxData`.  
- **fields** → subset of columns to return, default is all fields.  
- **filters** → `field:value` format, multi = AND, comma = IN.  
- **data_type** → `"list"` returns JSON object array, `"table"` returns 2D array + headers.  
- **data_format** →  
  - `"json"` → wrapped `ApiResponse`.  
  - `"json_download"` → raw JSON data, served as file.  
  - `"csv"` → CSV file.  
  - `"xlsx"` → Excel file.  
- **download** → if true, response is forced as `attachment`. If false, served inline.  

---

## 📌 API Response Standard

### Base Struct

```go
type ApiResponse struct {
    Success  bool        `json:"success"`
    Message  string      `json:"message"`
    Data     any         `json:"data,omitempty"`
    Metadata *ListMeta   `json:"metadata,omitempty"`
    Error    any         `json:"error,omitempty"`
}
```

### List Metadata

```go
type ListMeta struct {
    Page       int               `json:"page,omitempty"`
    PageSize   int               `json:"pageSize,omitempty"`
    Total      int               `json:"total,omitempty"`
    MaxData    int               `json:"maxData,omitempty"`

    Headers    []string          `json:"headers,omitempty"`
    DataType   string            `json:"dataType"`
    DataFormat string            `json:"dataFormat"`
    Formats    map[string]string `json:"formats,omitempty"`

    OrderBy    []string          `json:"orderBy,omitempty"`
    Filters    map[string]any    `json:"filters,omitempty"`
    Search     string            `json:"search,omitempty"`
}
```

### Rules

- **Success/Error** → mutually exclusive.  
- **Metadata** is always included for list responses.  
- **Headers** and **Formats** are always returned, regardless of `dataType`.  
- **OrderBy, Filters, Search** are echoed from request.  
- **DataFormat** is echoed back to reflect the chosen output format.  

---

## 📌 Response Helper (ApiHelper)

Lokstra provides an `ApiHelper` to simplify response building.

```go
type ApiHelper struct {
    ctx *RequestContext
}

// ✅ Success Responses
func (a *ApiHelper) Ok(message string) error
func (a *ApiHelper) OkData(message string, data any) error
func (a *ApiHelper) OkDeleted(message string) error

// ✅ List Responses (Paginated)
func (a *ApiHelper) OkList(message string, items any, headers []string, pg Pagination) error
func (a *ApiHelper) OkListTable(message string, rows [][]any, headers []string, pg Pagination) error

// ✅ List Responses (All Data, No Pagination)
func (a *ApiHelper) OkListAll(message string, items any, headers []string) error
func (a *ApiHelper) OkListTableAll(message string, rows [][]any, headers []string) error

// ✅ List with Formats (for Auto-UI)
func (a *ApiHelper) OkListWithFormat(message string, items any, headers []string, formats map[string]string, pg Pagination) error
func (a *ApiHelper) OkListTableWithFormat(message string, rows [][]any, headers []string, formats map[string]string, pg Pagination) error

// ✅ Error Responses
func (a *ApiHelper) Error(code, details string) error
func (a *ApiHelper) ErrorInternal(message string) error
func (a *ApiHelper) ErrorNotFound(message string) error
func (a *ApiHelper) ErrorUnauthorized(message string) error
func (a *ApiHelper) ErrorForbidden(message string) error
func (a *ApiHelper) ErrorValidation(fields map[string]string) error
func (a *ApiHelper) ErrorRule(code, message string) error
```

---

## 📌 Example

### Request
```
GET /users?page=1&page_size=20&fields=id,name,email&order_by=-created_at&data_type=table&data_format=csv&download=true
```

### Response (CSV)
```
Content-Type: text/csv
Content-Disposition: attachment; filename="users.csv"

id,name,email
1,John,john@mail.com
2,Ana,ana@mail.com
```

---

## ✅ Summary

- **Request**: standardized query for list operations, including paging, ordering, filtering, field selection, data type, and data format.  
- **Response**: standardized JSON structure with optional raw outputs (JSON/CSV/XLSX).  
- **ApiHelper**: convenience wrapper for building responses consistently.  

This standard ensures **predictable APIs**, **easy integration**, and **auto-UI rendering** support.
