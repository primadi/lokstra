package main

import (
	"context"
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// ============================================
// 1. Simple Factory Pattern
// ============================================

type SimpleService struct {
	name string
}

func NewSimpleService(name string) *SimpleService {
	fmt.Printf("✓ Creating SimpleService: %s\n", name)
	return &SimpleService{name: name}
}

func (s *SimpleService) GetInfo() string {
	return fmt.Sprintf("SimpleService: %s", s.name)
}

func SimpleServiceFactory(deps map[string]any, config map[string]any) any {
	name := "simple-service"
	if nameVal, ok := config["name"].(string); ok {
		name = nameVal
	}
	return NewSimpleService(name)
}

// ============================================
// 2. Factory with Configuration
// ============================================

type ConfigurableService struct {
	apiKey     string
	maxRetries int
	timeout    time.Duration
}

func NewConfigurableService(apiKey string, maxRetries int, timeout time.Duration) *ConfigurableService {
	fmt.Printf("✓ Creating ConfigurableService (retries: %d, timeout: %v)\n", maxRetries, timeout)
	return &ConfigurableService{
		apiKey:     apiKey,
		maxRetries: maxRetries,
		timeout:    timeout,
	}
}

func (s *ConfigurableService) GetConfig() map[string]any {
	return map[string]any{
		"has_api_key": s.apiKey != "",
		"max_retries": s.maxRetries,
		"timeout_sec": s.timeout.Seconds(),
	}
}

func ConfigurableServiceFactory(deps map[string]any, config map[string]any) any {
	// Read configuration with defaults
	apiKey := ""
	if key, ok := config["api_key"].(string); ok {
		apiKey = key
	}

	maxRetries := 3
	if retries, ok := config["max_retries"].(int); ok {
		maxRetries = retries
	}

	timeout := 30 * time.Second
	if timeoutVal, ok := config["timeout_seconds"].(int); ok {
		timeout = time.Duration(timeoutVal) * time.Second
	}

	return NewConfigurableService(apiKey, maxRetries, timeout)
}

// ============================================
// 3. Factory with Dependencies
// ============================================

type DependentService struct {
	cache  *CacheService
	logger *LoggerService
}

func NewDependentService(cache *CacheService, logger *LoggerService) *DependentService {
	logger.Log("Creating DependentService")
	return &DependentService{
		cache:  cache,
		logger: logger,
	}
}

func (s *DependentService) ProcessData(key string, value string) string {
	s.logger.Log(fmt.Sprintf("Processing: %s", key))
	s.cache.Set(key, value)
	return s.cache.Get(key)
}

func DependentServiceFactory(deps map[string]any, config map[string]any) any {
	cache := deps["cache-service"].(*service.Cached[*CacheService])
	logger := deps["logger-service"].(*service.Cached[*LoggerService])

	return NewDependentService(cache.Get(), logger.Get())
}

// ============================================
// 4. Lifecycle Management
// ============================================

type LifecycleService struct {
	name       string
	startTime  time.Time
	isRunning  bool
	background context.CancelFunc
}

func NewLifecycleService(name string) *LifecycleService {
	return &LifecycleService{
		name:      name,
		startTime: time.Now(),
		isRunning: false,
	}
}

func (s *LifecycleService) Start(ctx context.Context) error {
	if s.isRunning {
		return fmt.Errorf("service already running")
	}

	bgCtx, cancel := context.WithCancel(ctx)
	s.background = cancel
	s.isRunning = true

	// Simulate background task
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-bgCtx.Done():
				fmt.Printf("✓ Background task stopped for %s\n", s.name)
				return
			case <-ticker.C:
				fmt.Printf("⏰ Background task running for %s (uptime: %v)\n",
					s.name, time.Since(s.startTime).Round(time.Second))
			}
		}
	}()

	fmt.Printf("✓ LifecycleService started: %s\n", s.name)
	return nil
}

func (s *LifecycleService) Stop() error {
	if !s.isRunning {
		return nil
	}

	if s.background != nil {
		s.background()
	}
	s.isRunning = false
	fmt.Printf("✓ LifecycleService stopped: %s\n", s.name)
	return nil
}

func (s *LifecycleService) GetStatus() map[string]any {
	return map[string]any{
		"name":    s.name,
		"running": s.isRunning,
		"uptime":  time.Since(s.startTime).String(),
	}
}

func LifecycleServiceFactory(deps map[string]any, config map[string]any) any {
	name := "lifecycle-service"
	if nameVal, ok := config["name"].(string); ok {
		name = nameVal
	}

	svc := NewLifecycleService(name)

	// Start service immediately
	if err := svc.Start(context.Background()); err != nil {
		fmt.Printf("⚠ Failed to start service: %v\n", err)
	}

	return svc
}

// ============================================
// Supporting Services
// ============================================

type CacheService struct {
	data map[string]string
}

func NewCacheService() *CacheService {
	fmt.Println("✓ Creating CacheService")
	return &CacheService{
		data: make(map[string]string),
	}
}

func (s *CacheService) Set(key, value string) {
	s.data[key] = value
}

func (s *CacheService) Get(key string) string {
	return s.data[key]
}

func CacheServiceFactory(deps map[string]any, config map[string]any) any {
	return NewCacheService()
}

type LoggerService struct {
	prefix string
}

func NewLoggerService(prefix string) *LoggerService {
	fmt.Printf("✓ Creating LoggerService with prefix: %s\n", prefix)
	return &LoggerService{prefix: prefix}
}

func (s *LoggerService) Log(message string) {
	fmt.Printf("[%s] %s\n", s.prefix, message)
}

func LoggerServiceFactory(deps map[string]any, config map[string]any) any {
	prefix := "APP"
	if prefixVal, ok := config["prefix"].(string); ok {
		prefix = prefixVal
	}
	return NewLoggerService(prefix)
}

// ============================================
// HTTP Handlers
// ============================================

func GetSimpleInfo() *response.ApiHelper {
	svc := lokstra_registry.GetService[*SimpleService]("simple-service")
	return response.NewApiOk(map[string]any{
		"service": "simple",
		"info":    svc.GetInfo(),
	})
}

func GetConfigurableInfo() *response.ApiHelper {
	svc := lokstra_registry.GetService[*ConfigurableService]("configurable-service")
	return response.NewApiOk(map[string]any{
		"service": "configurable",
		"config":  svc.GetConfig(),
	})
}

type ProcessParams struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}

func ProcessData(params *ProcessParams) *response.ApiHelper {
	svc := lokstra_registry.GetService[*DependentService]("dependent-service")
	result := svc.ProcessData(params.Key, params.Value)
	return response.NewApiOk(map[string]any{
		"service": "dependent",
		"key":     params.Key,
		"result":  result,
	})
}

func GetLifecycleStatus() *response.ApiHelper {
	svc := lokstra_registry.GetService[*LifecycleService]("lifecycle-service")
	return response.NewApiOk(map[string]any{
		"service": "lifecycle",
		"status":  svc.GetStatus(),
	})
}

func Home() *response.Response {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Service Factories Example</title>
    <style>
        body { font-family: Arial; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .endpoint { background: #f5f5f5; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .method { display: inline-block; padding: 3px 8px; border-radius: 3px; font-weight: bold; color: white; }
        .get { background: #61affe; }
        .post { background: #49cc90; }
        code { background: #eee; padding: 2px 6px; border-radius: 3px; }
        a { color: #0066cc; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>🏭 Service Factories Example</h1>
    
    <p>This example demonstrates various service factory patterns:</p>
    <ul>
        <li><strong>Simple Factory</strong> - Basic service instantiation</li>
        <li><strong>Configurable Factory</strong> - Services with configuration</li>
        <li><strong>Dependent Factory</strong> - Services with dependencies</li>
        <li><strong>Lifecycle Management</strong> - Services with start/stop</li>
    </ul>

    <h2>Test Endpoints</h2>

    <div class="endpoint">
        <span class="method get">GET</span>
        <a href="/simple">/simple</a> - Simple factory info
    </div>

    <div class="endpoint">
        <span class="method get">GET</span>
        <a href="/configurable">/configurable</a> - Configurable factory info
    </div>

    <div class="endpoint">
        <span class="method post">POST</span>
        <code>/process</code> - Process data with dependencies
        <br><small>Body: {"key": "test", "value": "data"}</small>
    </div>

    <div class="endpoint">
        <span class="method get">GET</span>
        <a href="/lifecycle">/lifecycle</a> - Lifecycle service status
    </div>

    <h2>📖 Documentation</h2>
    <p>See <code>index</code> for detailed explanation</p>
    <p>Use <code>test.http</code> for API testing</p>
</body>
</html>`

	return response.NewHtmlResponse(html)
}

// ============================================
// Main
// ============================================

func main() {
	// Register all services
	lokstra_registry.RegisterServiceType("cache-service", CacheServiceFactory, nil)
	lokstra_registry.RegisterServiceType("logger-service", LoggerServiceFactory, nil)
	lokstra_registry.RegisterServiceType("simple-service", SimpleServiceFactory, nil)
	lokstra_registry.RegisterServiceType("configurable-service", ConfigurableServiceFactory, nil)
	lokstra_registry.RegisterServiceType("dependent-service", DependentServiceFactory, nil)
	lokstra_registry.RegisterServiceType("lifecycle-service", LifecycleServiceFactory, nil)

	// Define services
	lokstra_registry.RegisterLazyService("cache-service", CacheServiceFactory, nil)
	lokstra_registry.RegisterLazyService("logger-service", LoggerServiceFactory, map[string]any{
		"prefix": "FACTORY",
	})
	lokstra_registry.RegisterLazyService("simple-service", SimpleServiceFactory, map[string]any{
		"name": "My Simple Service",
	})
	lokstra_registry.RegisterLazyService("configurable-service", ConfigurableServiceFactory, map[string]any{
		"api_key":         "sk-test-12345",
		"max_retries":     5,
		"timeout_seconds": 60,
	})
	lokstra_registry.RegisterLazyServiceWithDeps("dependent-service",
		DependentServiceFactory,
		map[string]string{
			"cache-service":  "cache-service",
			"logger-service": "logger-service",
		},
		nil, nil,
	)
	lokstra_registry.RegisterLazyService("lifecycle-service", LifecycleServiceFactory, map[string]any{
		"name": "Background Worker",
	})

	// Setup router
	router := lokstra.NewRouter("service-factories")
	router.GET("/", Home)
	router.GET("/simple", GetSimpleInfo)
	router.GET("/configurable", GetConfigurableInfo)
	router.POST("/process", ProcessData)
	router.GET("/lifecycle", GetLifecycleStatus)

	// Create and run app
	app := lokstra.NewApp("service-factories", ":3000", router)

	fmt.Println("🚀 Service Factories Example")
	fmt.Println("📍 http://localhost:3000")
	fmt.Println("\n📋 Available endpoints:")
	fmt.Println("   GET  /simple         - Simple factory")
	fmt.Println("   GET  /configurable   - Configurable factory")
	fmt.Println("   POST /process        - Dependent factory")
	fmt.Println("   GET  /lifecycle      - Lifecycle management")
	fmt.Println("\n🧪 Open test.http for API testing")

	if err := app.Run(0); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
