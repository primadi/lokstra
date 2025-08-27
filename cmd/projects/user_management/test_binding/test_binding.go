package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	baseURL := "http://localhost:8081"

	fmt.Println("🧪 Testing Smart Content-Type Binding...")
	fmt.Println("=========================================")

	// Test 1: Form-encoded submission to API endpoint
	fmt.Println("\n1. Testing form-encoded submission to API endpoint:")
	testFormToAPI(baseURL)

	// Test 2: JSON submission to API endpoint (existing functionality)
	fmt.Println("\n2. Testing JSON submission to API endpoint:")
	testJSONToAPI(baseURL)

	// Test 3: Form submission to UI endpoint
	fmt.Println("\n3. Testing form submission to UI endpoint:")
	testFormToUI(baseURL)
}

func testFormToAPI(baseURL string) {
	// Create form data
	formData := url.Values{
		"username":  {"formtest"},
		"email":     {"formtest@example.com"},
		"password":  {"password123"},
		"is_active": {"true"},
	}

	// Make request
	resp, err := http.Post(
		baseURL+"/api/v1/users",
		"application/x-www-form-urlencoded",
		strings.NewReader(formData.Encode()),
	)

	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("   Status: %d\n", resp.StatusCode)
	fmt.Printf("   Response: %s\n", string(body))

	if resp.StatusCode == 201 {
		fmt.Printf("   ✅ Form-encoded binding to API endpoint successful!\n")
	} else {
		fmt.Printf("   ❌ Form-encoded binding failed\n")
	}
}

func testJSONToAPI(baseURL string) {
	// Create JSON data
	payload := map[string]interface{}{
		"username":  "jsontest",
		"email":     "jsontest@example.com",
		"password":  "password456",
		"is_active": true,
	}

	jsonData, _ := json.Marshal(payload)

	// Make request
	resp, err := http.Post(
		baseURL+"/api/v1/users",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("   Status: %d\n", resp.StatusCode)
	fmt.Printf("   Response: %s\n", string(body))

	if resp.StatusCode == 201 {
		fmt.Printf("   ✅ JSON binding to API endpoint successful!\n")
	} else {
		fmt.Printf("   ❌ JSON binding failed\n")
	}
}

func testFormToUI(baseURL string) {
	// Create form data
	formData := url.Values{
		"username":  {"uitest"},
		"email":     {"uitest@example.com"},
		"password":  {"password789"},
		"is_active": {"on"}, // HTML checkbox value
	}

	// Create request with HTMX headers
	req, err := http.NewRequest("POST", baseURL+"/users/create", strings.NewReader(formData.Encode()))
	if err != nil {
		fmt.Printf("   ❌ Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "#main-content")

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("   Status: %d\n", resp.StatusCode)

	// For UI endpoint, we expect HTML response
	if strings.Contains(string(body), "user") || resp.StatusCode == 200 {
		fmt.Printf("   ✅ Form submission to UI endpoint successful!\n")
		fmt.Printf("   Response contains: %s\n", string(body)[:min(200, len(body))]+"...")
	} else {
		fmt.Printf("   ❌ Form submission to UI endpoint failed\n")
		fmt.Printf("   Response: %s\n", string(body))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
