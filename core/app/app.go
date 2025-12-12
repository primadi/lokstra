package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/app/listener"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/lokstra_handler"
)

type App struct {
	name           string
	mainRouter     router.Router
	listenerConfig map[string]any

	listener listener.AppListener
}

// Create a new App instance with default listener configuration
func New(name string, addr string, routers ...router.Router) *App {
	return NewWithConfig(name, addr, "default", nil, routers...)
}

// Create a new App instance with custom listener configuration
func NewWithConfig(name string, addr string, listenerType string,
	cfg map[string]any, routers ...router.Router) *App {
	if cfg == nil {
		cfg = make(map[string]any)
	}
	cfg["addr"] = addr
	cfg["listener-type"] = listenerType

	app := &App{
		name:           name,
		listenerConfig: cfg,
	}

	for _, rt := range routers {
		app.AddRouter(rt)
	}
	return app
}

// Get the app name
func (a *App) GetName() string {
	return a.name
}

// Get the app listening address
func (a *App) GetAddress() string {
	if addr, ok := a.listenerConfig["addr"].(string); ok {
		return addr
	}
	return ""
}

// Get the main router of the app
func (a *App) GetRouter() router.Router {
	return a.mainRouter
}

// Add a router to the app. If there's already a router, it will be chained.
func (a *App) AddRouter(rt router.Router) {
	a.AddRouterWithPrefix(rt, "")
}

// Add a router to the app with a prefix. If there's already a router, it will be chained.
func (a *App) AddRouterWithPrefix(rt router.Router, appPrefix string) {
	// each router is cloned to avoid side effects
	r := rt.Clone()
	if a.mainRouter == nil {
		a.mainRouter = r
	} else {
		a.mainRouter.SetNextChainWithPrefix(r, appPrefix)
	}
}

// ReverseProxyRewrite represents path rewrite rules for reverse proxy
type ReverseProxyRewrite struct {
	From string // Pattern to match in path (regex supported)
	To   string // Replacement pattern
}

// ReverseProxyConfig represents a single reverse proxy configuration
type ReverseProxyConfig struct {
	Prefix      string               // URL prefix to match (e.g., "/api")
	StripPrefix bool                 // Whether to strip the prefix before forwarding
	Target      string               // Target backend URL (e.g., "http://api-server:8080")
	Rewrite     *ReverseProxyRewrite // Path rewrite rules
}

// AddReverseProxies creates a router for reverse proxies and mounts them.
// This is typically called from the config loader when reverse-proxies are defined in YAML.
// Reverse proxy router is prepended (mounted first) before other routers.
func (a *App) AddReverseProxies(proxies []*ReverseProxyConfig) {
	if len(proxies) == 0 {
		return
	}

	logger.LogInfo("📦 [%s] Adding %d reverse proxy(ies)...\n", a.name, len(proxies))

	// Create a dedicated router for reverse proxies
	proxyRouter := router.New(a.name + "-reverse-proxy")

	for _, proxy := range proxies {
		stripPrefix := ""
		if proxy.StripPrefix {
			stripPrefix = proxy.Prefix
		}

		// Prepare rewrite config if specified
		var rewrite *lokstra_handler.ReverseProxyRewrite
		if proxy.Rewrite != nil && proxy.Rewrite.From != "" {
			rewrite = &lokstra_handler.ReverseProxyRewrite{
				From: proxy.Rewrite.From,
				To:   proxy.Rewrite.To,
			}
			logger.LogInfo("   🔄 %s -> %s (strip: %v, rewrite: %s -> %s)",
				proxy.Prefix, proxy.Target, proxy.StripPrefix, proxy.Rewrite.From, proxy.Rewrite.To)
		} else {
			logger.LogInfo("   🔄 %s -> %s (strip: %v)",
				proxy.Prefix, proxy.Target, proxy.StripPrefix)
		}

		handler := lokstra_handler.MountReverseProxy(stripPrefix, proxy.Target, rewrite)
		proxyRouter.ANYPrefix(proxy.Prefix, handler)
	}

	// Prepend proxy router (make it the first router)
	if a.mainRouter != nil {
		// Save existing router chain
		existingRouter := a.mainRouter
		// Set proxy router as main
		a.mainRouter = proxyRouter
		// Chain existing routers after proxy router
		a.mainRouter.SetNextChainWithPrefix(existingRouter, "")
	} else {
		// No existing router, just set proxy router as main
		a.mainRouter = proxyRouter
	}

	logger.LogInfo("✅ [%s] Reverse proxies added successfully\n", a.name)
}

func (a *App) numRouters() int {
	if a.mainRouter == nil {
		return 0
	}

	curRouter := a.mainRouter
	count := 0

	for curRouter != nil {
		count++
		curRouter = curRouter.GetNextChain()
	}
	return count
}

// Print app start information, including the number of routers and their routes
func (a *App) PrintStartInfo() {
	logger.LogInfo("Starting [%s] with %d router(s) on address %s",
		a.name, a.numRouters(), a.listenerConfig["addr"])

	if a.mainRouter != nil {
		a.mainRouter.PrintRoutes()
	}
}

// Start the app. It blocks until the app stops or returns an error.
// Shutdown must be called separately.
func (a *App) Start() error {
	a.listener = listener.CreateListener(a.listenerConfig, a.mainRouter)
	return a.listener.ListenAndServe()
}

// Shutdown gracefully shuts down the app with a timeout.
func (a *App) Shutdown(timeout time.Duration) error {
	if a.listener != nil {
		return a.listener.Shutdown(timeout)
	}
	return nil
}

// Starts the app and blocks until a termination signal is received.
// It shuts down gracefully with the given timeout.
func (a *App) Run(timeout time.Duration) error {
	// Run app in background
	errCh := make(chan error, 1)
	go func() {
		if err := a.Start(); err != nil {
			errCh <- err
		}
	}()

	// Wait for signal or app error
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.LogInfo("Received shutdown signal: %v", sig)
		if err := a.Shutdown(timeout); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		return nil
	case err := <-errCh:
		return fmt.Errorf("app error: %w", err)
	}
}
