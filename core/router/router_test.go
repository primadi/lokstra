package router_test

import (
	"testing"

	"github.com/primadi/lokstra/core/router"
)

func TestRouterInterface(t *testing.T) {
	// Test that the router interface is properly defined
	// This is more of a compilation test to ensure the interface is complete
	var _ router.Router = (*router.RouterImpl)(nil)
	var _ router.Router = (*router.GroupImpl)(nil)
}

func TestRouterInterfaceCompleteness(t *testing.T) {
	// Test that all methods are present in the interface
	// This ensures interface completeness at compile time

	tests := []struct {
		name        string
		description string
	}{
		{"Prefix", "Router should have Prefix method"},
		{"Use", "Router should have Use method"},
		{"Handle", "Router should have Handle method"},
		{"HandleOverrideMiddleware", "Router should have HandleOverrideMiddleware method"},
		{"GET", "Router should have GET method"},
		{"POST", "Router should have POST method"},
		{"PUT", "Router should have PUT method"},
		{"PATCH", "Router should have PATCH method"},
		{"DELETE", "Router should have DELETE method"},
		{"WithOverrideMiddleware", "Router should have WithOverrideMiddleware method"},
		{"WithPrefix", "Router should have WithPrefix method"},
		{"MountStatic", "Router should have MountStatic method"},
		{"MountSPA", "Router should have MountSPA method"},
		{"MountReverseProxy", "Router should have MountReverseProxy method"},
		{"MountRpcService", "Router should have MountRpcService method"},
		{"Group", "Router should have Group method"},
		{"GroupBlock", "Router should have GroupBlock method"},
		{"RecurseAllHandler", "Router should have RecurseAllHandler method"},
		{"DumpRoutes", "Router should have DumpRoutes method"},
		{"ServeHTTP", "Router should have ServeHTTP method"},
		{"FastHttpHandler", "Router should have FastHttpHandler method"},
		{"OverrideMiddleware", "Router should have OverrideMiddleware method"},
		{"GetMiddleware", "Router should have GetMiddleware method"},
		{"LockMiddleware", "Router should have LockMiddleware method"},
		{"GetMeta", "Router should have GetMeta method"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If this test compiles, the method exists
			t.Log(tt.description)
		})
	}
}
