package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/core/response/api_formatter"
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
	return f.Success(data, message...)
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
		switch responseStatus {
		case "SUCCESS", "CREATED":
			cr.Status = "success"
			if payload, hasPayload := result["payload"]; hasPayload {
				cr.Data = payload
			}
			if desc, hasDesc := result["description"]; hasDesc {
				cr.Message = fmt.Sprint(desc)
			}
		case "ERROR", "VALIDATION_ERROR":
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
			cr.Error = errorObj
		}
	} else {
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
	return f.Success(data, message...)
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
			cr.Error = errorObj
		}
	} else {
		cr.Status = "unknown"
		cr.Data = result
	}

	return nil
}

func main() {
	fmt.Println("🎨 Custom Formatter Response Parsing Examples")
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Println()

	// Example 1: Parse Corporate Success Response
	fmt.Println("📋 Example 1: Corporate Format - Success Response")
	fmt.Println(strings.Repeat("-", 70))

	corporateSuccessBody := `{
		"responseCode": "00",
		"responseStatus": "SUCCESS",
		"payload": {
			"id": 123,
			"name": "John Doe",
			"email": "john@example.com"
		},
		"description": "User data retrieved successfully",
		"timestamp": "2024-01-01T12:00:00Z"
	}`

	httpResp1 := createMockResponse(200, corporateSuccessBody)
	corporateFormatter := NewCustomCorporateFormatter()
	cr1 := &api_formatter.ClientResponse{}

	if err := corporateFormatter.ParseClientResponse(httpResp1, cr1); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", cr1.Status)
	fmt.Printf("Message: %s\n", cr1.Message)
	fmt.Printf("Data: %+v\n", cr1.Data)
	fmt.Printf("HTTP Status Code: %d\n", cr1.StatusCode)
	fmt.Println()

	// Example 2: Parse Corporate Error Response
	fmt.Println("📋 Example 2: Corporate Format - Error Response")
	fmt.Println(strings.Repeat("-", 70))

	corporateErrorBody := `{
		"responseCode": "99",
		"responseStatus": "ERROR",
		"errorCode": "USER_NOT_FOUND",
		"errorMessage": "User with ID 999 not found",
		"errorDetails": {
			"userId": 999,
			"timestamp": "2024-01-01T12:00:00Z"
		},
		"timestamp": "2024-01-01T12:00:00Z"
	}`

	httpResp2 := createMockResponse(404, corporateErrorBody)
	cr2 := &api_formatter.ClientResponse{}

	if err := corporateFormatter.ParseClientResponse(httpResp2, cr2); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", cr2.Status)
	if cr2.Error != nil {
		fmt.Printf("Error Code: %s\n", cr2.Error.Code)
		fmt.Printf("Error Message: %s\n", cr2.Error.Message)
		fmt.Printf("Error Details: %+v\n", cr2.Error.Details)
	}
	fmt.Printf("HTTP Status Code: %d\n", cr2.StatusCode)
	fmt.Println()

	// Example 3: Parse Mobile Success Response
	fmt.Println("📱 Example 3: Mobile Format - Success Response")
	fmt.Println(strings.Repeat("-", 70))

	mobileSuccessBody := `{
		"ok": true,
		"data": {
			"id": 456,
			"title": "Sample Item",
			"status": "active"
		},
		"msg": "Data fetched successfully"
	}`

	httpResp3 := createMockResponse(200, mobileSuccessBody)
	mobileFormatter := NewMobileApiFormatter()
	cr3 := &api_formatter.ClientResponse{}

	if err := mobileFormatter.ParseClientResponse(httpResp3, cr3); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", cr3.Status)
	fmt.Printf("Message: %s\n", cr3.Message)
	fmt.Printf("Data: %+v\n", cr3.Data)
	fmt.Printf("HTTP Status Code: %d\n", cr3.StatusCode)
	fmt.Println()

	// Example 4: Parse Mobile Error Response
	fmt.Println("📱 Example 4: Mobile Format - Error Response")
	fmt.Println(strings.Repeat("-", 70))

	mobileErrorBody := `{
		"ok": false,
		"error": "Invalid request parameters",
		"code": "VALIDATION_ERROR",
		"info": {
			"field": "email",
			"reason": "invalid format"
		}
	}`

	httpResp4 := createMockResponse(400, mobileErrorBody)
	cr4 := &api_formatter.ClientResponse{}

	if err := mobileFormatter.ParseClientResponse(httpResp4, cr4); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", cr4.Status)
	if cr4.Error != nil {
		fmt.Printf("Error Code: %s\n", cr4.Error.Code)
		fmt.Printf("Error Message: %s\n", cr4.Error.Message)
		fmt.Printf("Error Details: %+v\n", cr4.Error.Details)
	}
	fmt.Printf("HTTP Status Code: %d\n", cr4.StatusCode)
	fmt.Println()

	// Example 5: Register and Use Custom Formatters
	fmt.Println("🔧 Example 5: Registering Custom Formatters")
	fmt.Println(strings.Repeat("-", 70))

	api_formatter.RegisterFormatter("corporate", NewCustomCorporateFormatter)
	api_formatter.RegisterFormatter("mobile", NewMobileApiFormatter)

	fmt.Println("✅ Registered 'corporate' formatter")
	fmt.Println("✅ Registered 'mobile' formatter")
	fmt.Println()

	// Use registered formatter
	registeredFormatter := api_formatter.CreateFormatter("corporate")
	httpResp5 := createMockResponse(200, corporateSuccessBody)
	cr5 := &api_formatter.ClientResponse{}

	if err := registeredFormatter.ParseClientResponse(httpResp5, cr5); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("Using registered 'corporate' formatter:\n")
	fmt.Printf("Status: %s\n", cr5.Status)
	fmt.Printf("Data: %+v\n", cr5.Data)
	fmt.Println()

	// Summary
	printSummary()
}

// Helper functions
func createMockResponse(statusCode int, body string) *http.Response {
	resp := &http.Response{
		StatusCode: statusCode,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}
	resp.Body = &mockReadCloser{strings.NewReader(body)}
	resp.Header.Set("Content-Type", "application/json")
	return resp
}

type mockReadCloser struct {
	*strings.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}

func printSummary() {
	fmt.Println("📊 SUMMARY")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("✨ Custom Formatter Features:")
	fmt.Println("  • CustomCorporateFormatter - Company-specific format")
	fmt.Println("    - responseCode, responseStatus, payload, timestamp")
	fmt.Println("    - Structured error with errorCode, errorMessage, errorDetails")
	fmt.Println()
	fmt.Println("  • MobileApiFormatter - Mobile-optimized format")
	fmt.Println("    - Simple ok/error flag")
	fmt.Println("    - Minimal payload with data/msg")
	fmt.Println("    - Compact error format")
	fmt.Println()
	fmt.Println("🔑 Key Benefits:")
	fmt.Println("  ✅ Consistent parsing across different response formats")
	fmt.Println("  ✅ Easy to register and use custom formatters")
	fmt.Println("  ✅ Automatic error detection and extraction")
	fmt.Println("  ✅ Support for metadata and pagination")
	fmt.Println("  ✅ Fallback handling for unknown formats")
	fmt.Println()
	fmt.Println("💡 Usage Pattern:")
	fmt.Println("  1. Define custom formatter implementing ResponseFormatter interface")
	fmt.Println("  2. Implement ParseClientResponse() method")
	fmt.Println("  3. Register formatter: RegisterFormatter(\"name\", constructor)")
	fmt.Println("  4. Use formatter to parse HTTP responses")
	fmt.Println()
}
