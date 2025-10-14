package app

import (
	"testing"

	"github.com/primadi/lokstra/core/router"
)

func TestAddReverseProxies(t *testing.T) {
	tests := []struct {
		name        string
		proxies     []*ReverseProxyConfig
		wantRouters int // Expected number of routers after adding proxies
	}{
		{
			name:        "No proxies",
			proxies:     nil,
			wantRouters: 0,
		},
		{
			name:        "Empty proxies",
			proxies:     []*ReverseProxyConfig{},
			wantRouters: 0,
		},
		{
			name: "Single proxy",
			proxies: []*ReverseProxyConfig{
				{
					Prefix:      "/api",
					StripPrefix: true,
					Target:      "http://backend:8080",
				},
			},
			wantRouters: 1,
		},
		{
			name: "Multiple proxies",
			proxies: []*ReverseProxyConfig{
				{
					Prefix:      "/api",
					StripPrefix: true,
					Target:      "http://api-server:8080",
				},
				{
					Prefix:      "/auth",
					StripPrefix: false,
					Target:      "http://auth-server:9000",
				},
			},
			wantRouters: 1, // All proxies in one router
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New("test-app", ":8080")

			app.AddReverseProxies(tt.proxies)

			got := app.numRouters()
			if got != tt.wantRouters {
				t.Errorf("numRouters() = %d, want %d", got, tt.wantRouters)
			}
		})
	}
}

func TestAddReverseProxies_WithExistingRouter(t *testing.T) {
	app := New("test-app", ":8080")

	// Add a regular router first
	regularRouter := router.New("regular-router")
	app.AddRouter(regularRouter)

	// Add reverse proxies
	proxies := []*ReverseProxyConfig{
		{
			Prefix:      "/api",
			StripPrefix: true,
			Target:      "http://backend:8080",
		},
	}
	app.AddReverseProxies(proxies)

	// Should have 2 routers: proxy router + regular router
	got := app.numRouters()
	want := 2
	if got != want {
		t.Errorf("numRouters() = %d, want %d", got, want)
	}

	// First router should be the proxy router
	mainRouter := app.GetRouter()
	if mainRouter == nil {
		t.Fatal("mainRouter is nil")
	}

	// Verify router name contains "reverse-proxy"
	if mainRouter.Name() != "test-app-reverse-proxy" {
		t.Errorf("First router name = %s, want test-app-reverse-proxy", mainRouter.Name())
	}
}
