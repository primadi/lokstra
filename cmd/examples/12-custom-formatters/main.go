package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/response/api_formatter"
	"github.com/primadi/lokstra/core/router"
)

// CustomCorporateFormatter implements company-specific response format
type CustomCorporateFormatter struct{}

func NewCustomCorporateFormatter() api_formatter.ResponseFormatter {
	return &CustomCorporateFormatter{}
}

func (f *CustomCorporateFormatter) Success(data any, message ...string) any {
	result := map[string]any{
		"responseCode":   "00",
		"responseStatus": "SUCCESS",
		"payload":        data,
		"timestamp":      "2024-01-01T12:00:00Z",
	}
	if len(message) > 0 {
		result["description"] = message[0]
	}
	return result
}

func (f *CustomCorporateFormatter) Created(data any, message ...string) any {
	result := map[string]any{
		"responseCode":   "01",
		"responseStatus": "CREATED",
		"payload":        data,
		"timestamp":      "2024-01-01T12:00:00Z",
	}
	if len(message) > 0 {
		result["description"] = message[0]
	}
	return result
}

func (f *CustomCorporateFormatter) Error(code string, message string, details ...map[string]any) any {
	result := map[string]any{
		"responseCode":   "99",
		"responseStatus": "ERROR",
		"errorCode":      code,
		"errorMessage":   message,
		"timestamp":      "2024-01-01T12:00:00Z",
	}
	if len(details) > 0 {
		result["errorDetails"] = details[0]
	}
	return result
}

func (f *CustomCorporateFormatter) ValidationError(message string, fields []api_formatter.FieldError) any {
	return map[string]any{
		"responseCode":     "98",
		"responseStatus":   "VALIDATION_ERROR",
		"errorMessage":     message,
		"validationErrors": fields,
		"timestamp":        "2024-01-01T12:00:00Z",
	}
}

func (f *CustomCorporateFormatter) NotFound(message string) any {
	return f.Error("NOT_FOUND", message)
}

func (f *CustomCorporateFormatter) List(data any, meta *api_formatter.ListMeta) any {
	result := map[string]any{
		"responseCode":   "00",
		"responseStatus": "SUCCESS",
		"payload":        data,
		"timestamp":      "2024-01-01T12:00:00Z",
	}
	if meta != nil {
		result["paginationInfo"] = meta
	}
	return result
}

func (f *CustomCorporateFormatter) ParseClientResponse(resp *http.Response, cr *api_formatter.ClientResponse) error {
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Store raw body and status code
	cr.RawBody = body
	cr.StatusCode = resp.StatusCode

	// Parse headers
	cr.Headers = make(map[string]any)
	for key, values := range resp.Header {
		if len(values) == 1 {
			cr.Headers[key] = values[0]
		} else {
			cr.Headers[key] = values
		}
	}

	// Try to parse as Corporate format
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		// If not valid JSON, treat as plain text
		cr.Status = "unknown"
		cr.Data = string(body)
		return nil
	}

	// Parse Corporate format
	if responseStatus, ok := result["responseStatus"].(string); ok {
		switch responseStatus {
		case "SUCCESS", "CREATED":
			cr.Status = "success"
			// Extract payload
			if payload, hasPayload := result["payload"]; hasPayload {
				cr.Data = payload
			}
			// Extract description as message
			if desc, hasDesc := result["description"]; hasDesc {
				cr.Message = fmt.Sprint(desc)
			}
			// Extract pagination info
			if paginationInfo, hasPagination := result["paginationInfo"]; hasPagination {
				if paginationMap, ok := paginationInfo.(map[string]any); ok {
					paginationBytes, _ := json.Marshal(paginationMap)
					var listMeta api_formatter.ListMeta
					if json.Unmarshal(paginationBytes, &listMeta) == nil {
						cr.Meta = &api_formatter.Meta{ListMeta: &listMeta}
					}
				}
			}
		case "ERROR", "VALIDATION_ERROR":
			cr.Status = "error"
			// Extract error details
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
		// Unknown format, store entire result as data
		cr.Status = "unknown"
		cr.Data = result
	}

	return nil
}

// MobileApiFormatter implements mobile-optimized response format
type MobileApiFormatter struct{}

func NewMobileApiFormatter() api_formatter.ResponseFormatter {
	return &MobileApiFormatter{}
}

func (f *MobileApiFormatter) Success(data any, message ...string) any {
	result := map[string]any{
		"ok":   true,
		"data": data,
	}
	if len(message) > 0 {
		result["msg"] = message[0]
	}
	return result
}

func (f *MobileApiFormatter) Created(data any, message ...string) any {
	result := map[string]any{
		"ok":   true,
		"data": data,
		"new":  true,
	}
	if len(message) > 0 {
		result["msg"] = message[0]
	}
	return result
}

func (f *MobileApiFormatter) Error(code string, message string, details ...map[string]any) any {
	result := map[string]any{
		"ok":    false,
		"error": message,
		"code":  code,
	}
	if len(details) > 0 {
		result["info"] = details[0]
	}
	return result
}

func (f *MobileApiFormatter) ValidationError(message string, fields []api_formatter.FieldError) any {
	return map[string]any{
		"ok":     false,
		"error":  message,
		"fields": fields,
	}
}

func (f *MobileApiFormatter) NotFound(message string) any {
	return f.Error("NOT_FOUND", message)
}

func (f *MobileApiFormatter) List(data any, meta *api_formatter.ListMeta) any {
	result := map[string]any{
		"ok":   true,
		"data": data,
	}
	if meta != nil {
		result["page"] = meta
	}
	return result
}

func (f *MobileApiFormatter) ParseClientResponse(resp *http.Response, cr *api_formatter.ClientResponse) error {
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Store raw body and status code
	cr.RawBody = body
	cr.StatusCode = resp.StatusCode

	// Parse headers
	cr.Headers = make(map[string]any)
	for key, values := range resp.Header {
		if len(values) == 1 {
			cr.Headers[key] = values[0]
		} else {
			cr.Headers[key] = values
		}
	}

	// Try to parse as Mobile format
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		// If not valid JSON, treat as plain text
		cr.Status = "unknown"
		cr.Data = string(body)
		return nil
	}

	// Parse Mobile format - check "ok" field
	if ok, hasOk := result["ok"].(bool); hasOk {
		if ok {
			cr.Status = "success"
			// Extract data
			if data, hasData := result["data"]; hasData {
				cr.Data = data
			}
			// Extract message
			if msg, hasMsg := result["msg"]; hasMsg {
				cr.Message = fmt.Sprint(msg)
			}
			// Extract pagination
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
			// Error response
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
		// Unknown format, store entire result as data
		cr.Status = "unknown"
		cr.Data = result
	}

	return nil
}

// Example data
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users = []User{
	{1, "John Doe", "john@example.com"},
	{2, "Jane Smith", "jane@example.com"},
}

func GetUsers(c *request.Context) error {
	return c.Api.Ok(users)
}

func CreateUser(c *request.Context) error {
	var user User
	if err := c.Req.BindBody(&user); err != nil {
		fields := []api_formatter.FieldError{
			{Field: "body", Code: "INVALID_JSON", Message: err.Error()},
		}
		return c.Api.ValidationError("Invalid request body", fields)
	}

	user.ID = len(users) + 1
	users = append(users, user)
	return c.Api.Created(user, "User created successfully")
}

func GetUserNotFound(c *request.Context) error {
	return c.Api.NotFound("User not found")
}

func main() {
	// Register custom formatters
	api_formatter.RegisterFormatter("corporate", NewCustomCorporateFormatter)
	api_formatter.RegisterFormatter("mobile", NewMobileApiFormatter)

	r := router.New("custom-formatter-example")

	// Routes with default formatter
	r.GET("/default/users", GetUsers)
	r.POST("/default/users", CreateUser)
	r.GET("/default/users/404", GetUserNotFound)

	// Routes that demonstrate formatter switching
	r.GET("/corporate/users", func(c *request.Context) error {
		response.SetApiResponseFormatterByName("corporate")
		return GetUsers(c)
	})

	r.POST("/corporate/users", func(c *request.Context) error {
		response.SetApiResponseFormatterByName("corporate")
		return CreateUser(c)
	})

	r.GET("/corporate/users/404", func(c *request.Context) error {
		response.SetApiResponseFormatterByName("corporate")
		return GetUserNotFound(c)
	})

	r.GET("/mobile/users", func(c *request.Context) error {
		response.SetApiResponseFormatterByName("mobile")
		return GetUsers(c)
	})

	r.POST("/mobile/users", func(c *request.Context) error {
		response.SetApiResponseFormatterByName("mobile")
		return CreateUser(c)
	})

	r.GET("/mobile/users/404", func(c *request.Context) error {
		response.SetApiResponseFormatterByName("mobile")
		return GetUserNotFound(c)
	})

	// Route to switch back to default
	r.GET("/reset", func(c *request.Context) error {
		response.SetApiResponseFormatterByName("api")
		return c.Api.Ok(map[string]string{"formatter": "reset to default"})
	})

	fmt.Println("🎯 Custom Response Formatter Example")
	fmt.Println()
	fmt.Println("📍 Default Formatter (structured API response):")
	fmt.Println("GET  /default/users     - Default structured format")
	fmt.Println("POST /default/users     - Default creation response")
	fmt.Println("GET  /default/users/404 - Default error response")
	fmt.Println()
	fmt.Println("🏢 Corporate Formatter (company-specific format):")
	fmt.Println("GET  /corporate/users     - Corporate response format")
	fmt.Println("POST /corporate/users     - Corporate creation format")
	fmt.Println("GET  /corporate/users/404 - Corporate error format")
	fmt.Println()
	fmt.Println("📱 Mobile Formatter (mobile-optimized format):")
	fmt.Println("GET  /mobile/users     - Mobile response format")
	fmt.Println("POST /mobile/users     - Mobile creation format")
	fmt.Println("GET  /mobile/users/404 - Mobile error format")
	fmt.Println()
	fmt.Println("🔄 Formatter Control:")
	fmt.Println("GET  /reset - Reset to default formatter")

	printFormatterComparison()
}

func printFormatterComparison() {
	fmt.Println("\n📄 FORMATTER OUTPUT COMPARISON:")

	fmt.Println("\n🔹 Default Formatter (api):")
	fmt.Println(`{
  "status": "success",
  "data": [
    {"id": 1, "name": "John Doe", "email": "john@example.com"}
  ]
}`)

	fmt.Println("\n🔹 Corporate Formatter:")
	fmt.Println(`{
  "responseCode": "00",
  "responseStatus": "SUCCESS",
  "payload": [
    {"id": 1, "name": "John Doe", "email": "john@example.com"}
  ],
  "timestamp": "2024-01-01T12:00:00Z"
}`)

	fmt.Println("\n🔹 Mobile Formatter:")
	fmt.Println(`{
  "ok": true,
  "data": [
    {"id": 1, "name": "John Doe", "email": "john@example.com"}
  ]
}`)

	fmt.Println("\n❌ ERROR HANDLING COMPARISON:")

	fmt.Println("\n🔹 Default Error:")
	fmt.Println(`{
  "status": "error",
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found"
  }
}`)

	fmt.Println("\n🔹 Corporate Error:")
	fmt.Println(`{
  "responseCode": "99",
  "responseStatus": "ERROR",
  "errorCode": "NOT_FOUND",
  "errorMessage": "User not found",
  "timestamp": "2024-01-01T12:00:00Z"
}`)

	fmt.Println("\n🔹 Mobile Error:")
	fmt.Println(`{
  "ok": false,
  "error": "User not found",
  "code": "NOT_FOUND"
}`)

	fmt.Println("\n✨ CUSTOM FORMATTER BENEFITS:")
	fmt.Println("• Registry Pattern: Register formatters at startup")
	fmt.Println("• Runtime Switching: Change formats per request/route")
	fmt.Println("• Legacy Integration: Maintain existing API contracts")
	fmt.Println("• Team Consistency: Enforce company-wide response standards")
	fmt.Println("• Environment-Specific: Different formats for different clients")
	fmt.Println()
	fmt.Println("🎚️ IMPLEMENTATION STEPS:")
	fmt.Println("1. Implement ResponseFormatter interface")
	fmt.Println("2. Register formatter: response.RegisterFormatter(\"name\", constructor)")
	fmt.Println("3. Switch at startup: response.SetApiResponseFormatterByName(\"name\")")
	fmt.Println("4. Or switch per route: middleware or route handler")
}
