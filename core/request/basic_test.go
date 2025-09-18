package request_test

import (
	"testing"

	"github.com/primadi/lokstra/core/request"
)

func TestPackageCompilation(t *testing.T) {
	// Test that the request package compiles successfully
	t.Log("Request package compiled successfully")
}

// func TestHandlerFuncType(t *testing.T) {
// 	// Test that HandlerFunc type exists and can be used
// 	var handler request.HandlerFunc

// 	// Test that we can assign a function to it
// 	handler = func(ctx *request.Context) error {
// 		return nil
// 	}

// 	if handler == nil {
// 		t.Error("HandlerFunc should not be nil after assignment")
// 	}
// }

func TestHandlerRegisterStructure(t *testing.T) {
	// Test HandlerRegister struct
	hr := request.HandlerRegister{
		Name: "test-handler",
		HandlerFunc: func(ctx *request.Context) error {
			return nil
		},
	}

	if hr.Name != "test-handler" {
		t.Errorf("Expected handler name 'test-handler', got %s", hr.Name)
	}

	if hr.HandlerFunc == nil {
		t.Error("Expected HandlerFunc to be set")
	}
}

func TestRequestTypes(t *testing.T) {
	// Verify that important types are properly exported
	type typeTest struct {
		name     string
		typeFunc func() any
	}

	tests := []typeTest{
		{
			name: "Context",
			typeFunc: func() any {
				return &request.Context{}
			},
		},
		{
			name: "HandlerRegister",
			typeFunc: func() any {
				return &request.HandlerRegister{}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.typeFunc()
			if result == nil {
				t.Errorf("Type %s should not be nil", test.name)
			}
		})
	}
}
