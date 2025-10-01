package main

import (
	"fmt"

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
