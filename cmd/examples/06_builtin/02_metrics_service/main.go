package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/logger"
	"github.com/primadi/lokstra/services/metrics"
)

// This example demonstrates comprehensive usage of Lokstra's built-in metrics service.
// It shows how to use Prometheus metrics for monitoring application performance,
// business metrics, and system health.
//
// Learning Objectives:
// - Understand metrics service configuration and setup
// - Learn to use counters, gauges, and histograms effectively
// - See practical monitoring patterns and best practices
// - Explore metrics collection and Prometheus integration
// - Master performance measurement and business metrics
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/services/metrics.md

func main() {
	fmt.Println("📊 Metrics Service Example - Comprehensive Monitoring")
	fmt.Println("")

	// Create registration context
	regCtx := lokstra.NewGlobalRegistrationContext()

	// Register service modules
	regCtx.RegisterModule(metrics.GetModule)
	regCtx.RegisterModule(logger.GetModule)

	// Configure metrics service
	metricsConfig := map[string]interface{}{
		"enabled":                 true,
		"endpoint":                "/metrics",
		"namespace":               "lokstra_demo",
		"subsystem":               "api",
		"collect_interval":        "10s",
		"include_go_metrics":      true,
		"include_process_metrics": true,
		"buckets":                 []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
	}

	// Create metrics service
	_, err := regCtx.CreateService("metrics", "app-metrics", metricsConfig)
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create metrics service: %v", err)
	}

	// Create logger service
	loggerConfig := map[string]interface{}{
		"level":  "info",
		"format": "json",
		"output": "stdout",
	}
	_, err = regCtx.CreateService("logger", "app-logger", loggerConfig)
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create logger service: %v", err)
	}

	// Register handlers with metrics instrumentation
	registerHandlers(regCtx)

	// Create and configure server
	server := lokstra.NewServer(regCtx, "metrics-server")

	app := lokstra.NewApp(regCtx, "main-app", ":8080")

	// Register routes and middleware

	// Simple request logging middleware
	app.Use(func(ctx *lokstra.Context, next lokstra.HandlerFunc) error {
		logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")
		logger.Infof("%s %s from %s", ctx.Request.Method, ctx.Request.URL.Path, ctx.Request.RemoteAddr)
		return next(ctx)
	})

	// Home endpoint
	app.GET("/", "home")

	// User operations endpoint
	app.GET("/api/users", "user_operations")

	// Order processing endpoint
	app.GET("/api/orders", "order_processing")

	// Analytics endpoint
	app.GET("/api/analytics", "analytics")

	// Health check endpoint
	app.GET("/health", "health")

	// Add metrics endpoint
	server.AddApp(app)

	// Start server
	err = server.Start()
	if err != nil {
		lokstra.Logger.Fatalf("Failed to load server: %v", err)
	}

	fmt.Println("🎯 Starting server with metrics monitoring...")
	fmt.Println("📊 Metrics endpoint: http://localhost:8080/metrics")
	fmt.Println("🔗 Test endpoints to generate metrics data")
	fmt.Println("")

	// Start background metrics generation
	go generateBusinessMetrics(regCtx)
	go monitorSystemHealth(regCtx)

	// Start the server
	if err := server.Start(); err != nil {
		lokstra.Logger.Fatalf("Failed to start server: %v", err)
	}
}

// registerHandlers registers HTTP handlers with comprehensive metrics instrumentation
func registerHandlers(regCtx lokstra.RegistrationContext) {
	// Home handler with basic metrics
	regCtx.RegisterHandler("home", func(ctx *lokstra.Context) error {
		metrics, _ := serviceapi.GetService[serviceapi.Metrics](regCtx, "app-metrics")
		logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")

		// Track API calls
		metrics.IncCounter("api_requests_total", serviceapi.Labels{
			"endpoint": "/",
			"method":   "GET",
			"handler":  "home",
		})

		// Measure response time
		start := time.Now()
		defer func() {
			duration := time.Since(start).Seconds()
			metrics.ObserveHistogram("api_request_duration_seconds", duration, serviceapi.Labels{
				"endpoint": "/",
				"method":   "GET",
			})
		}()

		logger.Infof("Home endpoint accessed")

		return ctx.Ok(map[string]interface{}{
			"message": "Metrics Service Demo",
			"features": []string{
				"Request counting",
				"Response time measurement",
				"Business metrics tracking",
				"System health monitoring",
				"Prometheus integration",
			},
			"endpoints": map[string]string{
				"metrics":   "/metrics",
				"user_ops":  "/api/users",
				"orders":    "/api/orders",
				"analytics": "/api/analytics",
				"health":    "/health",
			},
		})
	})

	// User operations handler with detailed metrics
	regCtx.RegisterHandler("user_operations", func(ctx *lokstra.Context) error {
		metrics, _ := serviceapi.GetService[serviceapi.Metrics](regCtx, "app-metrics")
		logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")

		// Track user operations
		operation := ctx.GetQueryParamWithDefault("operation", "view")
		userType := ctx.GetQueryParamWithDefault("type", "regular")

		metrics.IncCounter("user_operations_total", serviceapi.Labels{
			"operation": operation,
			"user_type": userType,
		})

		// Simulate processing time based on operation
		processingTime := getOperationTime(operation)
		time.Sleep(time.Duration(processingTime) * time.Millisecond)

		// Track operation duration
		metrics.ObserveHistogram("user_operation_duration_seconds",
			float64(processingTime)/1000.0, serviceapi.Labels{
				"operation": operation,
				"user_type": userType,
			})

		// Track active users (gauge)
		activeUsers := rand.Intn(100) + 50
		metrics.SetGauge("active_users_current", float64(activeUsers), serviceapi.Labels{
			"user_type": userType,
		})

		logger.Infof("User operation completed: %s for %s user", operation, userType)

		return ctx.Ok(map[string]interface{}{
			"operation":       operation,
			"user_type":       userType,
			"processing_time": processingTime,
			"active_users":    activeUsers,
			"timestamp":       time.Now(),
		})
	})

	// Order processing handler with business metrics
	regCtx.RegisterHandler("order_processing", func(ctx *lokstra.Context) error {
		metrics, _ := serviceapi.GetService[serviceapi.Metrics](regCtx, "app-metrics")
		logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")

		// Simulate order details
		orderType := ctx.GetQueryParamWithDefault("type", "product")
		region := ctx.GetQueryParamWithDefault("region", "us-east")
		amount := rand.Float64()*1000 + 10

		// Track order metrics
		metrics.IncCounter("orders_processed_total", serviceapi.Labels{
			"order_type": orderType,
			"region":     region,
		})

		// Track order values
		metrics.ObserveHistogram("order_value_dollars", amount, serviceapi.Labels{
			"order_type": orderType,
			"region":     region,
		})

		// Track current orders in processing
		ordersInProcess := rand.Intn(20) + 5
		metrics.SetGauge("orders_processing_current", float64(ordersInProcess), serviceapi.Labels{
			"region": region,
		})

		// Simulate processing time
		processingTime := time.Duration(rand.Intn(500)+100) * time.Millisecond
		time.Sleep(processingTime)

		logger.Infof("Order processed: type=%s, region=%s, amount=%.2f", orderType, region, amount)

		return ctx.Ok(map[string]interface{}{
			"order_id":        fmt.Sprintf("ord_%d", time.Now().Unix()),
			"type":            orderType,
			"region":          region,
			"amount":          amount,
			"processing_time": processingTime.String(),
			"status":          "completed",
		})
	})

	// Analytics handler with performance metrics
	regCtx.RegisterHandler("analytics", func(ctx *lokstra.Context) error {
		metrics, _ := serviceapi.GetService[serviceapi.Metrics](regCtx, "app-metrics")
		logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")

		// Track analytics queries
		queryType := ctx.GetQueryParamWithDefault("query", "dashboard")
		dataSource := ctx.GetQueryParamWithDefault("source", "database")

		metrics.IncCounter("analytics_queries_total", serviceapi.Labels{
			"query_type":  queryType,
			"data_source": dataSource,
		})

		// Simulate heavy computation
		start := time.Now()
		computationTime := time.Duration(rand.Intn(2000)+500) * time.Millisecond
		time.Sleep(computationTime)

		// Track computation performance
		metrics.ObserveHistogram("analytics_computation_seconds",
			time.Since(start).Seconds(), serviceapi.Labels{
				"query_type":  queryType,
				"data_source": dataSource,
			})

		// Track cache hit ratio
		cacheHitRate := rand.Float64()
		metrics.SetGauge("analytics_cache_hit_ratio", cacheHitRate, serviceapi.Labels{
			"data_source": dataSource,
		})

		logger.Infof("Analytics query completed: type=%s, source=%s", queryType, dataSource)

		return ctx.Ok(map[string]interface{}{
			"query_type":        queryType,
			"data_source":       dataSource,
			"computation_time":  computationTime.String(),
			"cache_hit_rate":    cacheHitRate,
			"records_processed": rand.Intn(10000) + 1000,
		})
	})

	// Health check with system metrics
	regCtx.RegisterHandler("health", func(ctx *lokstra.Context) error {
		metrics, _ := serviceapi.GetService[serviceapi.Metrics](regCtx, "app-metrics")

		// Track health checks
		metrics.IncCounter("health_checks_total", serviceapi.Labels{
			"status": "success",
		})

		// System health metrics
		cpuUsage := rand.Float64() * 100
		memoryUsage := rand.Float64() * 100
		diskUsage := rand.Float64() * 100

		metrics.SetGauge("system_cpu_usage_percent", cpuUsage, nil)
		metrics.SetGauge("system_memory_usage_percent", memoryUsage, nil)
		metrics.SetGauge("system_disk_usage_percent", diskUsage, nil)

		return ctx.Ok(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
			"system": map[string]float64{
				"cpu_usage":    cpuUsage,
				"memory_usage": memoryUsage,
				"disk_usage":   diskUsage,
			},
		})
	})
}

// generateBusinessMetrics simulates background business metrics generation
func generateBusinessMetrics(regCtx lokstra.RegistrationContext) {
	metrics, _ := serviceapi.GetService[serviceapi.Metrics](regCtx, "app-metrics")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Simulate business metrics
		revenue := rand.Float64() * 10000
		metrics.SetGauge("business_revenue_current", revenue, serviceapi.Labels{
			"currency": "USD",
		})

		// Customer satisfaction
		satisfaction := 4.0 + rand.Float64()
		metrics.SetGauge("customer_satisfaction_score", satisfaction, nil)

		// Transaction volume
		transactions := rand.Intn(1000) + 100
		metrics.SetGauge("transaction_volume_current", float64(transactions), nil)

		// Error rate
		errorRate := rand.Float64() * 0.1 // 0-10%
		metrics.SetGauge("application_error_rate", errorRate, nil)
	}
}

// monitorSystemHealth simulates system health monitoring
func monitorSystemHealth(regCtx lokstra.RegistrationContext) {
	metrics, _ := serviceapi.GetService[serviceapi.Metrics](regCtx, "app-metrics")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Database connections
		dbConnections := rand.Intn(50) + 10
		metrics.SetGauge("database_connections_active", float64(dbConnections), nil)

		// Cache hit ratio
		cacheHits := rand.Float64()*0.4 + 0.6 // 60-100%
		metrics.SetGauge("cache_hit_ratio_total", cacheHits, nil)

		// Queue depth
		queueDepth := rand.Intn(100)
		metrics.SetGauge("message_queue_depth", float64(queueDepth), serviceapi.Labels{
			"queue_type": "processing",
		})
	}
}

// getOperationTime returns simulated processing time based on operation type
func getOperationTime(operation string) int {
	switch operation {
	case "create":
		return rand.Intn(300) + 200 // 200-500ms
	case "update":
		return rand.Intn(200) + 100 // 100-300ms
	case "delete":
		return rand.Intn(150) + 50 // 50-200ms
	case "view":
		return rand.Intn(100) + 25 // 25-125ms
	default:
		return rand.Intn(100) + 50 // 50-150ms
	}
}

// Metrics Service Key Concepts:
//
// 1. Counter Metrics:
//    - Track events that only increase (requests, errors, orders)
//    - Use IncCounter() with descriptive labels
//    - Perfect for rate calculations and totals
//    - Examples: API calls, user registrations, transactions
//
// 2. Histogram Metrics:
//    - Measure distributions of values (response times, sizes)
//    - Use ObserveHistogram() with measurement values
//    - Provides percentiles, averages, and totals
//    - Examples: request duration, payload sizes, computation time
//
// 3. Gauge Metrics:
//    - Track values that can go up and down (current state)
//    - Use SetGauge() with current values
//    - Represents instantaneous measurements
//    - Examples: active users, queue depth, system resources
//
// 4. Labels and Dimensions:
//    - Add context to metrics with key-value labels
//    - Enable filtering and aggregation in monitoring tools
//    - Keep cardinality reasonable to avoid performance issues
//    - Examples: region, user_type, error_type, endpoint

// Best Practices:
//
// 1. Metric Naming:
//    - Use consistent naming conventions
//    - Include units in metric names (seconds, bytes, total)
//    - Use descriptive but concise names
//    - Follow Prometheus naming guidelines
//
// 2. Label Usage:
//    - Use labels for dimensions you want to query by
//    - Avoid high-cardinality labels (user IDs, timestamps)
//    - Keep label values consistent and predictable
//    - Use lowercase with underscores
//
// 3. Performance Considerations:
//    - Minimize metric creation in hot paths
//    - Pre-register metrics when possible
//    - Use appropriate metric types for use cases
//    - Monitor metrics cardinality
//
// 4. Business Value:
//    - Track business KPIs alongside technical metrics
//    - Measure user experience and satisfaction
//    - Monitor revenue and conversion metrics
//    - Create actionable alerts and dashboards

// Test Commands:
//
// # Start the application
// go run main.go
//
// # View Prometheus metrics
// curl http://localhost:8080/metrics
//
// # Test different endpoints to generate metrics
// curl http://localhost:8080/
// curl "http://localhost:8080/api/users?operation=create&type=premium"
// curl "http://localhost:8080/api/orders?type=subscription&region=eu-west"
// curl "http://localhost:8080/api/analytics?query=revenue&source=cache"
// curl http://localhost:8080/health
//
// # Generate load to see metrics in action
// for i in {1..10}; do curl "http://localhost:8080/api/users?operation=view"; done
// for i in {1..5}; do curl "http://localhost:8080/api/orders?type=product&region=us-east"; done
//
// # Monitor metrics over time
// watch -n 2 'curl -s http://localhost:8080/metrics | grep lokstra_demo'
