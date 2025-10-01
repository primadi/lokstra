package main

import (
	"fmt"
	"log"
	"os"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Sample app with realistic structure
func main() {
	fmt.Println("🚀 Lokstra Realistic YAML Config Demo")
	fmt.Println("=====================================")

	// Step 1: Setup factories (usually in separate package)
	setupFactories()

	// Step 2: Register components with code (main business logic)
	setupRouters()
	setupServers()

	// Step 3: Apply YAML configuration to modify behavior
	configPath := "./app-config"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	if err := applyConfiguration(configPath); err != nil {
		log.Fatalf("Failed to apply configuration: %v", err)
	}

	// Step 4: Start the configured server
	// Demo general configuration usage
	fmt.Println("\n🔧 General Configuration Demo...")
	serverURL := lokstra_registry.GetConfigString("server-url", "http://localhost")
	maxConn := lokstra_registry.GetConfigInt("max-connections", 100)
	debugEnabled := lokstra_registry.GetConfigBool("debug-enabled", false)

	fmt.Printf("📡 Server URL: %s\n", serverURL)
	fmt.Printf("⚡ Max Connections: %d\n", maxConn)
	fmt.Printf("🐛 Debug Mode: %t\n", debugEnabled)

	// Show all available configs
	fmt.Printf("📋 Available configs: %v\n", lokstra_registry.ListConfigNames())

	fmt.Println("\n🎯 Starting configured server...")
	srv := lokstra_registry.GetServer("main-server")
	if srv != nil {
		fmt.Println("✅ Server ready to start with configuration")
		// In real app: srv.Start()
	} else {
		log.Fatal("❌ Server not found in registry")
	}
}

// Setup middleware and service factories
func setupFactories() {
	fmt.Println("\n1. Setting up factories...")

	// Middleware factories
	lokstra_registry.RegisterMiddlewareFactory("logger", func(config map[string]any) request.HandlerFunc {
		level := "info"
		if l, ok := config["level"].(string); ok {
			level = l
		}
		return func(c *request.Context) error {
			// Use general config for debug level
			debugEnabled := lokstra_registry.GetConfigBool("debug-enabled", false)
			if debugEnabled {
				fmt.Printf("[%s-DEBUG] %s %s\n", level, c.R.Method, c.R.URL.Path)
			} else {
				fmt.Printf("[%s] %s %s\n", level, c.R.Method, c.R.URL.Path)
			}
			return nil
		}
	})

	lokstra_registry.RegisterMiddlewareFactory("auth", func(config map[string]any) request.HandlerFunc {
		// Use general config for JWT secret with fallback to middleware config
		secret := lokstra_registry.GetConfigString("jwt-secret", "default-secret")
		if s, ok := config["secret"].(string); ok && s != "" {
			secret = s // Override with middleware-specific config if provided
		}
		return func(c *request.Context) error {
			auth := c.R.Header.Get("Authorization")
			if auth == "" {
				return c.Resp.WithStatus(401).Json(map[string]string{"error": "unauthorized"})
			}
			fmt.Printf("[AUTH] Validated with secret: %s...\n", secret[:5])
			return nil
		}
	})

	lokstra_registry.RegisterMiddlewareFactory("cors", func(config map[string]any) request.HandlerFunc {
		return func(c *request.Context) error {
			c.W.Header().Set("Access-Control-Allow-Origin", "*")
			return nil
		}
	})

	lokstra_registry.RegisterMiddlewareFactory("rate-limit", func(config map[string]any) request.HandlerFunc {
		return func(c *request.Context) error {
			fmt.Printf("[RATE-LIMIT] Checking rate limit for %s\n", c.R.URL.Path)
			return nil
		}
	})

	lokstra_registry.RegisterMiddlewareFactory("cache", func(config map[string]any) request.HandlerFunc {
		return func(c *request.Context) error {
			fmt.Printf("[CACHE] Checking cache for %s\n", c.R.URL.Path)
			return nil
		}
	})

	lokstra_registry.RegisterMiddlewareFactory("monitoring", func(config map[string]any) request.HandlerFunc {
		return func(c *request.Context) error {
			fmt.Printf("[MONITORING] Monitoring request %s %s\n", c.R.Method, c.R.URL.Path)
			return nil
		}
	})

	lokstra_registry.RegisterMiddlewareFactory("audit-log", func(config map[string]any) request.HandlerFunc {
		return func(c *request.Context) error {
			fmt.Printf("[AUDIT] Admin action: %s %s\n", c.R.Method, c.R.URL.Path)
			return nil
		}
	})

	// Service factories
	lokstra_registry.RegisterServiceFactory("database", func(config map[string]any) any {
		host := config["host"].(string)
		db := config["database"].(string)
		// Use general config for pool size
		poolSize := lokstra_registry.GetConfigInt("database-pool-size", 10)
		fmt.Printf("📊 Connecting to database: %s/%s with pool size: %d\n", host, db, poolSize)
		return map[string]string{
			"status":    "connected",
			"host":      host,
			"db":        db,
			"pool_size": fmt.Sprintf("%d", poolSize),
		}
	})

	lokstra_registry.RegisterServiceFactory("cache", func(config map[string]any) any {
		addr := config["addr"].(string)
		fmt.Printf("🔄 Connecting to cache: %s\n", addr)
		return map[string]string{"status": "connected", "addr": addr}
	})

	fmt.Println("✅ Factories registered")
}

// Setup routers with business logic
func setupRouters() {
	fmt.Println("\n2. Setting up routers...")

	// API Router - register routes in code first
	apiRouter := router.New("api")

	// Register routes with names that match YAML config
	apiRouter.GET("/api/health", func(c *request.Context) error {
		return c.Resp.WithStatus(200).Json(map[string]string{
			"status":  "healthy",
			"service": "api",
		})
	}, route.WithNameOption("health-endpoint")) // Name matches config

	apiRouter.GET("/api/users", func(c *request.Context) error {
		return c.Resp.WithStatus(200).Json([]map[string]string{
			{"id": "1", "name": "John Doe"},
			{"id": "2", "name": "Jane Smith"},
		})
	}, route.WithNameOption("users-endpoint")) // Name matches config

	apiRouter.GET("/api/status", func(c *request.Context) error {
		return c.Resp.WithStatus(200).Json(map[string]string{
			"status": "operational",
		})
	}, route.WithNameOption("status-endpoint")) // Name matches config

	lokstra_registry.RegisterRouter("api", apiRouter)
	fmt.Println("✅ API router registered with named routes")

	// Admin Router - register routes in code first
	adminRouter := router.New("admin")

	adminRouter.GET("/admin", func(c *request.Context) error {
		return c.Resp.WithStatus(200).Json(map[string]string{
			"page": "admin dashboard",
		})
	}, route.WithNameOption("admin-dashboard")) // Name matches config

	adminRouter.GET("/admin/users", func(c *request.Context) error {
		return c.Resp.WithStatus(200).Json(map[string]string{
			"page": "user management",
		})
	}, route.WithNameOption("admin-users")) // Name matches config

	adminRouter.POST("/admin/reset-all", func(c *request.Context) error {
		return c.Resp.WithStatus(200).Json(map[string]string{
			"action": "reset completed",
		})
	}, route.WithNameOption("admin-reset")) // Name matches config (can be disabled via YAML)

	lokstra_registry.RegisterRouter("admin", adminRouter)
	fmt.Println("✅ Admin router registered with named routes")
}

// Setup servers
func setupServers() {
	fmt.Println("\n3. Setting up servers...")

	// Create apps (will be configured via YAML)
	apiApp := app.New("api-service", ":8080")
	adminApp := app.New("admin-panel", ":8081")

	// Create server
	mainServer := server.New("main-server", apiApp, adminApp)
	lokstra_registry.RegisterServer("main-server", mainServer)
	fmt.Println("✅ Main server registered")
}

// Apply YAML configuration
func applyConfiguration(configPath string) error {
	fmt.Printf("\n4. Loading configuration from: %s\n", configPath)

	var cfg config.Config

	// Check if config path exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("⚠️  Config path does not exist, using embedded config\n")
		return applyEmbeddedConfig()
	}

	// Determine if it's a file or directory
	info, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("failed to stat config path: %v", err)
	}

	if info.IsDir() {
		err = config.LoadConfigDir(configPath, &cfg)
	} else {
		err = config.LoadConfigFile(configPath, &cfg)
	}

	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %v", err)
	}

	// Configuration loaded successfully

	// Apply configuration
	return config.ApplyAllConfig(&cfg, "main-server")
}

// Fallback embedded configuration
func applyEmbeddedConfig() error {
	cfg := config.Config{
		Middlewares: []config.Middleware{
			{Name: "api-logger", Type: "logger", Config: map[string]interface{}{"level": "info"}},
			{Name: "api-cors", Type: "cors", Config: map[string]interface{}{}},
			{Name: "admin-auth", Type: "auth", Config: map[string]interface{}{"secret": "admin-secret-key"}},
		},
		Services: []config.Service{
			{Name: "main-db", Type: "database", Config: map[string]interface{}{
				"host": "localhost", "database": "app_db",
			}},
			{Name: "session-cache", Type: "cache", Config: map[string]interface{}{
				"addr": "localhost:6379",
			}},
		},
		Routers: []config.Router{
			{
				Name: "api",
				Use:  []string{"api-logger", "api-cors"},
				Routes: []config.Route{
					{Name: "status-endpoint", Use: []string{"monitoring"}},
				},
			},
			{
				Name: "admin",
				Use:  []string{"api-logger", "admin-auth"},
				Routes: []config.Route{
					{Name: "admin-users", Use: []string{"audit-log"}},
				},
			},
		},
		Servers: []config.Server{
			{
				Name:     "main-server",
				Services: []string{"main-db", "session-cache"},
				Apps: []config.App{
					{Name: "api-service", Addr: ":8080", Routers: []string{"api"}},
					{Name: "admin-panel", Addr: ":8081", Routers: []string{"admin"}},
				},
			},
		},
	}

	fmt.Println("📝 Using embedded configuration")
	return config.ApplyAllConfig(&cfg, "main-server")
}
