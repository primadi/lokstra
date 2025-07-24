package router_test

import (
	"testing"

	"github.com/primadi/lokstra/core/router"
)

func TestPackageCompilation(t *testing.T) {
	// This test ensures that the router package can be imported and compiled
	// without any compilation errors
	t.Log("Router package compiled successfully")
}

func TestInterfaceImplementation(t *testing.T) {
	// Test that RouterImpl implements the Router interface
	var _ router.Router = (*router.RouterImpl)(nil)

	// Test that GroupImpl implements the Router interface
	var _ router.Router = (*router.GroupImpl)(nil)

	t.Log("Interface implementations verified")
}
