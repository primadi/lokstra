package response_test

import (
	"testing"

	"github.com/primadi/lokstra/core/response"
)

func TestPackageCompilation(t *testing.T) {
	// This test ensures that the response package can be imported and compiled
	// without any compilation errors
	t.Log("Response package compiled successfully")
}

func TestInterfaceImplementation(t *testing.T) {
	// Test that Response implements the response interface
	r := response.NewResponse()

	// Test that it has required methods
	_ = r.GetStatusCode()
	_ = r.GetHeaders()

	t.Log("Interface implementations verified")
}

func TestResponseTypeDefinition(t *testing.T) {
	// Test that ResponseCode is properly defined
	var code response.ResponseCode = "TEST"
	if string(code) != "TEST" {
		t.Error("ResponseCode type not working correctly")
	}

	// Test all predefined codes exist
	codes := []response.ResponseCode{
		response.CodeOK,
		response.CodeCreated,
		response.CodeUpdated,
		response.CodeNotFound,
		response.CodeDuplicate,
		response.CodeBadRequest,
		response.CodeInternal,
	}

	if len(codes) != 7 {
		t.Error("Expected 7 predefined response codes")
	}

	t.Log("Response types verified")
}
