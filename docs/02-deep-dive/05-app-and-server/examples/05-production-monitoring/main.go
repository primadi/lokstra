package main

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/primadi/lokstra"
)

// Production monitoring example

var (
	startTime       time.Time
	requestCount    atomic.Uint64
	errorCount      atomic.Uint64
	totalLatency    atomic.Uint64
	activeRequests  atomic.Int32
	requestDuration = make(map[string]time.Duration)
	requestMu       sync.RWMutex
)

// Metrics middleware
func MetricsMiddleware(next func() map[string]any) func() map[string]any {
	return func() map[string]any {
		start := time.Now()
		activeRequests.Add(1)
		requestCount.Add(1)

		// Call handler
		result := next()

		// Record metrics
		duration := time.Since(start)
		totalLatency.Add(uint64(duration.Milliseconds()))
		activeRequests.Add(-1)

		// Track endpoint latency
		requestMu.Lock()
		requestDuration["endpoint"] = duration
		requestMu.Unlock()

		return result
	}
}

// Metrics handler - Prometheus format
func MetricsHandler() string {
	uptime := time.Since(startTime).Seconds()
	requests := requestCount.Load()
	errors := errorCount.Load()
	avgLatency := float64(0)
	if requests > 0 {
		avgLatency = float64(totalLatency.Load()) / float64(requests)
	}

	return formatPrometheusMetrics(map[string]any{
		"http_requests_total":       requests,
		"http_errors_total":         errors,
		"http_request_duration_avg": avgLatency,
		"http_requests_active":      activeRequests.Load(),
		"process_uptime_seconds":    uptime,
	})
}

func formatPrometheusMetrics(metrics map[string]any) string {
	result := "# HELP http_requests_total Total number of HTTP requests\n"
	result += "# TYPE http_requests_total counter\n"
	result += formatMetric("http_requests_total", metrics["http_requests_total"])

	result += "\n# HELP http_errors_total Total number of HTTP errors\n"
	result += "# TYPE http_errors_total counter\n"
	result += formatMetric("http_errors_total", metrics["http_errors_total"])

	result += "\n# HELP http_request_duration_avg Average request duration in ms\n"
	result += "# TYPE http_request_duration_avg gauge\n"
	result += formatMetric("http_request_duration_avg", metrics["http_request_duration_avg"])

	result += "\n# HELP http_requests_active Currently active requests\n"
	result += "# TYPE http_requests_active gauge\n"
	result += formatMetric("http_requests_active", metrics["http_requests_active"])

	result += "\n# HELP process_uptime_seconds Process uptime in seconds\n"
	result += "# TYPE process_uptime_seconds counter\n"
	result += formatMetric("process_uptime_seconds", metrics["process_uptime_seconds"])

	return result
}

func formatMetric(name string, value any) string {
	return name + " " + formatValue(value) + "\n"
}

func formatValue(value any) string {
	switch v := value.(type) {
	case int:
		return formatInt(int64(v))
	case int32:
		return formatInt(int64(v))
	case int64:
		return formatInt(v)
	case uint64:
		return formatUint(v)
	case float64:
		return formatFloat(v)
	default:
		return "0"
	}
}

func formatInt(v int64) string {
	return formatFloat(float64(v))
}

func formatUint(v uint64) string {
	return formatFloat(float64(v))
}

func formatFloat(v float64) string {
	// Simple float formatting
	if v == float64(int64(v)) {
		return formatInt(int64(v))
	}
	// Format with 2 decimal places
	str := ""
	negative := v < 0
	if negative {
		v = -v
	}

	intPart := int64(v)
	fracPart := int64((v - float64(intPart)) * 100)

	if negative {
		str = "-"
	}
	str += formatInt(intPart) + "."
	if fracPart < 10 {
		str += "0"
	}
	str += formatInt(fracPart)

	return str
}

// Stats handler - JSON format
func StatsHandler() map[string]any {
	return map[string]any{
		"uptime_seconds":  time.Since(startTime).Seconds(),
		"requests_total":  requestCount.Load(),
		"errors_total":    errorCount.Load(),
		"requests_active": activeRequests.Load(),
		"avg_latency_ms":  getAvgLatency(),
		"memory_mb":       128, // Simulated
		"goroutines":      10,  // Simulated
	}
}

func getAvgLatency() float64 {
	requests := requestCount.Load()
	if requests == 0 {
		return 0
	}
	return float64(totalLatency.Load()) / float64(requests)
}

// Health handler
func HealthHandler() map[string]any {
	return map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}
}

// API handler with metrics
func APIHandler() map[string]any {
	// Simulate processing
	time.Sleep(10 * time.Millisecond)

	return map[string]any{
		"message": "API response",
		"data":    []int{1, 2, 3, 4, 5},
	}
}

// Error handler (simulates errors)
func ErrorHandler() map[string]any {
	errorCount.Add(1)
	return map[string]any{
		"error":   "Simulated error",
		"message": "This endpoint always returns an error",
	}
}

// Home handler
func HomeHandler() string {
	return `
	<html>
	<body>
		<h1>Production Monitoring Example</h1>
		<p>Demonstrates metrics and monitoring patterns</p>
		<h2>Monitoring Endpoints</h2>
		<ul>
			<li><a href="/metrics">Metrics</a> - Prometheus format</li>
			<li><a href="/stats">Stats</a> - JSON format</li>
			<li><a href="/health">Health</a> - Health check</li>
		</ul>
		<h2>Test Endpoints</h2>
		<ul>
			<li><a href="/api">API</a> - Normal endpoint</li>
			<li><a href="/error">Error</a> - Simulated error</li>
		</ul>
		<h2>Metrics Tracked</h2>
		<ul>
			<li>Request count</li>
			<li>Error count</li>
			<li>Active requests</li>
			<li>Average latency</li>
			<li>Uptime</li>
		</ul>
	</body>
	</html>
	`
}

func main() {
	startTime = time.Now()

	// Create router
	router := lokstra.NewRouter("main")
	router.GET("/", HomeHandler)
	router.GET("/metrics", MetricsHandler)
	router.GET("/stats", StatsHandler)
	router.GET("/health", HealthHandler)
	router.GET("/api", APIHandler)
	router.GET("/error", ErrorHandler)

	// Create app
	app := lokstra.NewApp("production-monitoring", ":3110", router)

	log.Println("Server starting on :3110")
	log.Println("Monitoring endpoints:")
	log.Println("  /metrics - Prometheus format")
	log.Println("  /stats - JSON format")
	log.Println("  /health - Health check")

	if err := app.Run(0); err != nil {
		log.Fatal(err)
	}
}
