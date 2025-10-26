package router_test

import (
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/core/router"
)

func TestNestedGroup3Level(t *testing.T) {
	r := router.New("root")
	level1 := r.AddGroup("/level1")
	level2 := level1.AddGroup("/level2")
	level3 := level2.AddGroup("/level3")
	level3.GET("/test", func(c *request.Context) error {
		return nil
	})

	var found bool
	r.Walk(func(rt *route.Route) {
		if rt.FullPath == "/level1/level2/level3/test" {
			found = true
		}
	})
	if !found {
		t.Errorf("Route in nested group not found")
	}
}

func TestChainRouter3Chain(t *testing.T) {
	r1 := router.New("r1")
	r2 := router.New("r2")
	r3 := router.New("r3")
	r1.SetNextChain(r2).SetNextChainWithPrefix(r3, "/r3")

	order := []string{}
	r1.GET("/a", func(c *request.Context) error {
		order = append(order, "r1")
		return nil
	})
	r2.GET("/b", func(c *request.Context) error {
		order = append(order, "r2")
		return nil
	})
	r3.GET("/c", func(c *request.Context) error {
		order = append(order, "r3")
		return nil
	})

	// Walk all routes in chain
	var r1Found, r2Found, r3Found bool
	r1.Walk(func(rt *route.Route) {
		switch rt.FullPath {
		case "/a":
			r1Found = true
		case "/b":
			r2Found = true
		case "/r3/c":
			r3Found = true
		}
	})
	if !r1Found || !r2Found || !r3Found {
		t.Errorf("Not all routes in chain found")
	}

	if r1.GetNextChain() != r2 || r2.GetNextChain() != r3 {
		t.Errorf("Chain not set correctly")
	}
}

func TestRouterWalk(t *testing.T) {
	r := router.New("root")
	r.GET("/a", func(c *request.Context) error { return nil })
	g := r.AddGroup("/g")
	g.GET("/b", func(c *request.Context) error { return nil })

	count := 0
	r.Walk(func(rt *route.Route) {
		count++
	})
	if count != 2 {
		t.Errorf("Expected 2 routes, got %d", count)
	}
}

func TestMiddlewareOrder(t *testing.T) {
	calls := []string{}
	r := router.New("root")

	r.Use(func(c *request.Context) error {
		calls = append(calls, "router-mw")
		return c.Next()
	})

	r.GET("/x", func(c *request.Context) error {
		calls = append(calls, "handler")
		return nil
	}, func(c *request.Context) error {
		calls = append(calls, "route-mw")
		return c.Next()
	})

	// Simulate request using httptest
	req := httptest.NewRequest("GET", "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check order
	if len(calls) != 3 || calls[0] != "router-mw" || calls[1] != "route-mw" || calls[2] != "handler" {
		t.Errorf("Middleware/handler order incorrect: %v", calls)
	}
}
