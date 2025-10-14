package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/core/response/api_formatter"
)

func main() {
	// Example 1: Parse ApiResponse format (default formatter)
	fmt.Println("=== Example 1: ApiResponse Format ===")

	// Simulate HTTP response with ApiResponse format
	apiResponseBody := `{
		"status": "success",
		"message": "Data fetched successfully",
		"data": {
			"id": 123,
			"name": "John Doe",
			"email": "john@example.com"
		},
		"meta": {
			"page": 1,
			"page_size": 10,
			"total": 100
		}
	}`

	httpResp1 := &http.Response{
		StatusCode: 200,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}
	httpResp1.Body = &mockReadCloser{strings.NewReader(apiResponseBody)}
	httpResp1.Header.Set("Content-Type", "application/json")

	// Use default formatter
	defaultFormatter := api_formatter.NewApiResponseFormatter()
	cr1 := &api_formatter.ClientResponse{}

	if err := defaultFormatter.ParseClientResponse(httpResp1, cr1); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", cr1.Status)
	fmt.Printf("Message: %s\n", cr1.Message)
	fmt.Printf("Data: %+v\n", cr1.Data)
	fmt.Printf("Meta: %+v\n", cr1.Meta)
	fmt.Printf("HTTP Status Code: %d\n\n", cr1.StatusCode)

	// Example 2: Parse Simple format
	fmt.Println("=== Example 2: Simple Format ===")

	simpleResponseBody := `{
		"data": {
			"id": 456,
			"title": "Sample Item"
		},
		"message": "Success"
	}`

	httpResp2 := &http.Response{
		StatusCode: 200,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}
	httpResp2.Body = &mockReadCloser{strings.NewReader(simpleResponseBody)}
	httpResp2.Header.Set("Content-Type", "application/json")

	// Use simple formatter
	simpleFormatter := api_formatter.NewSimpleResponseFormatter()
	cr2 := &api_formatter.ClientResponse{}

	if err := simpleFormatter.ParseClientResponse(httpResp2, cr2); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", cr2.Status)
	fmt.Printf("Message: %s\n", cr2.Message)
	fmt.Printf("Data: %+v\n", cr2.Data)
	fmt.Printf("HTTP Status Code: %d\n\n", cr2.StatusCode)

	// Example 3: Parse Error Response
	fmt.Println("=== Example 3: Error Response ===")

	errorResponseBody := `{
		"status": "error",
		"error": {
			"code": "VALIDATION_ERROR",
			"message": "Invalid input data",
			"fields": [
				{
					"field": "email",
					"code": "INVALID_FORMAT",
					"message": "Email format is invalid"
				}
			]
		}
	}`

	httpResp3 := &http.Response{
		StatusCode: 400,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}
	httpResp3.Body = &mockReadCloser{strings.NewReader(errorResponseBody)}

	// Use default formatter
	cr3 := &api_formatter.ClientResponse{}

	if err := defaultFormatter.ParseClientResponse(httpResp3, cr3); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", cr3.Status)
	fmt.Printf("Error Code: %s\n", cr3.Error.Code)
	fmt.Printf("Error Message: %s\n", cr3.Error.Message)
	if len(cr3.Error.Fields) > 0 {
		fmt.Printf("Field Errors:\n")
		for _, fe := range cr3.Error.Fields {
			fmt.Printf("  - %s: %s (%s)\n", fe.Field, fe.Message, fe.Code)
		}
	}
	fmt.Printf("HTTP Status Code: %d\n\n", cr3.StatusCode)

	// Example 4: Using global formatter
	fmt.Println("=== Example 4: Using Global Formatter ===")

	// Set global formatter by name
	api_formatter.SetGlobalFormatterByName("simple")

	httpResp4 := &http.Response{
		StatusCode: 200,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}
	httpResp4.Body = &mockReadCloser{strings.NewReader(`{"id": 789, "status": "active"}`)}

	cr4 := &api_formatter.ClientResponse{}
	globalFormatter := api_formatter.GetGlobalFormatter()

	if err := globalFormatter.ParseClientResponse(httpResp4, cr4); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", cr4.Status)
	fmt.Printf("Data: %+v\n", cr4.Data)
	fmt.Printf("HTTP Status Code: %d\n", cr4.StatusCode)
}

// mockReadCloser is a helper to create io.ReadCloser from strings.Reader
type mockReadCloser struct {
	*strings.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}
