package response_test

import (
	"testing"

	"github.com/primadi/lokstra/core/response"
)

func TestResponseCodeConstants(t *testing.T) {
	tests := []struct {
		name     string
		code     response.ResponseCode
		expected string
	}{
		{"OK code", response.CodeOK, "OK"},
		{"Created code", response.CodeCreated, "CREATED"},
		{"Updated code", response.CodeUpdated, "UPDATED"},
		{"Not found code", response.CodeNotFound, "NOT_FOUND"},
		{"Duplicate code", response.CodeDuplicate, "DUPLICATE"},
		{"Bad request code", response.CodeBadRequest, "BAD_REQUEST"},
		{"Internal error code", response.CodeInternal, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.code) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.code))
			}
		})
	}
}

func TestResponseCodeExtensibility(t *testing.T) {
	// Test that developers can create custom response codes
	const CustomCode response.ResponseCode = "CUSTOM_ERROR"

	if string(CustomCode) != "CUSTOM_ERROR" {
		t.Errorf("Expected CUSTOM_ERROR, got %s", string(CustomCode))
	}
}

func TestResponseCodeType(t *testing.T) {
	// Test that ResponseCode is based on string
	var code response.ResponseCode = "TEST_CODE"

	// Should be able to use as string
	str := string(code)
	if str != "TEST_CODE" {
		t.Errorf("Expected TEST_CODE, got %s", str)
	}
}
