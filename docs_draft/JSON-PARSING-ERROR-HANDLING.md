# JSON Parsing Error Handling

## Overview

Improved error handling untuk JSON parsing errors di request body, mengubah error internal menjadi user-friendly validation errors.

## Problem

**Before Fix:**

Ketika client mengirim request dengan format JSON yang salah (misalnya array of strings padahal seharusnya array of objects), error yang dikembalikan adalah:

```json
{
  "status": "error",
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "services.CreateOrderRequest.Items: []*services.ItemOrder: readObjectStart: expect { or n, but found \", error found in #10 byte of ...|items\": [\"item_lapto|..., bigger context ..."
  }
}
```

**Issues:**
1. Error code adalah `INTERNAL_ERROR` (500) padahal seharusnya `VALIDATION_ERROR` (400)
2. Error message terlalu teknis dan sulit dipahami user
3. Tidak memberikan petunjuk yang jelas tentang masalahnya

## Solution

**After Fix:**

```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": [
      {
        "field": "body",
        "code": "INVALID_JSON",
        "message": "Invalid data type in request body. Expected an object but received a different type."
      }
    ]
  }
}
```

**Improvements:**
1. ✅ Error code berubah menjadi `VALIDATION_ERROR` dengan HTTP status 400
2. ✅ Error message user-friendly dan deskriptif
3. ✅ Structured field errors yang konsisten dengan validation errors lainnya

## Implementation

### Modified File: `core/request/request_helper.go`

```go
func unmarshalBody(data []byte, v any) error {
	err := jsonBodyDecoder.Unmarshal(data, v)
	if err == nil {
		return jsonDecoder.Unmarshal(data, v)
	}
	
	// Create a more user-friendly error message for JSON parsing errors
	errMsg := err.Error()
	
	// Try to detect common JSON parsing errors and provide better messages
	userFriendlyMsg := "Invalid JSON format"
	if strings.Contains(errMsg, "expect { or n, but found") {
		userFriendlyMsg = "Invalid data type in request body. Expected an object but received a different type."
	} else if strings.Contains(errMsg, "expects \" or n, but found") {
		userFriendlyMsg = "Invalid data type in request body. Expected a string but received a different type."
	} else if strings.Contains(errMsg, "readObjectStart") {
		userFriendlyMsg = "Invalid array element format. Expected object notation but received a different type."
	}
	
	// Wrap JSON parsing error as validation error for better error handling
	return &ValidationError{
		FieldErrors: []api_formatter.FieldError{
			{
				Field:   "body",
				Code:    "INVALID_JSON",
				Message: userFriendlyMsg,
			},
		},
	}
}
```

## Common Error Messages

| Original jsoniter Error | User-Friendly Message |
|------------------------|----------------------|
| `expect { or n, but found "` | Invalid data type in request body. Expected an object but received a different type. |
| `expects " or n, but found` | Invalid data type in request body. Expected a string but received a different type. |
| `readObjectStart` | Invalid array element format. Expected object notation but received a different type. |
| Other JSON errors | Invalid JSON format |

## Test Cases

### Test 1: Array of Strings Instead of Objects

**Request:**
```http
POST /api/v1/orders
Content-Type: application/json

{
  "user_id": "user_1",
  "items": ["item_laptop", "item_mouse"]  ← Wrong! Should be array of objects
}
```

**Expected Struct:**
```go
type CreateOrderRequest struct {
    UserID string       `json:"user_id"`
    Items  []*ItemOrder `json:"items"`  // Array of objects, not strings!
}

type ItemOrder struct {
    ProductID string  `json:"product_id"`
    Quantity  int     `json:"quantity"`
    Price     float64 `json:"price"`
}
```

**Response:**
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": [
      {
        "field": "body",
        "code": "INVALID_JSON",
        "message": "Invalid data type in request body. Expected an object but received a different type."
      }
    ]
  }
}
```

**HTTP Status:** 400 Bad Request

### Test 2: Correct Format (Should Work)

**Request:**
```http
POST /api/v1/orders
Content-Type: application/json

{
  "user_id": "user_1",
  "items": [
    {
      "product_id": "item_laptop",
      "quantity": 1,
      "price": 999.99
    },
    {
      "product_id": "item_mouse",
      "quantity": 2,
      "price": 29.99
    }
  ]
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "order_id": "order_1",
    "user_id": "user_1",
    "items": [...],
    "status": "pending"
  }
}
```

**HTTP Status:** 200 OK

## Benefits

1. **Better Developer Experience:** Clear error messages membantu client developers memahami masalah dengan cepat
2. **Consistent Error Format:** Menggunakan format yang sama dengan validation errors lainnya
3. **Proper HTTP Status Codes:** 400 Bad Request untuk user errors, bukan 500 Internal Server Error
4. **Easier Debugging:** Error messages yang deskriptif mengurangi waktu debugging

## Related Files

- `core/request/request_helper.go` - Core implementation
- `core/request/context.go` - Error handling in FinalizeResponse
- `core/response/api_formatter/types.go` - FieldError struct definition

## See Also

- [API Standards](./api-standard.md) - Error response format standards
- [Request Handling](./request-handling.md) - Request binding and validation
- [Validation](../common/validator/README.md) - Struct validation
