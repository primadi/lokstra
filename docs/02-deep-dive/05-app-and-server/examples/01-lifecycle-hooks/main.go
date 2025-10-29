package main

import (
	"log"
	"time"

	"github.com/primadi/lokstra"
)

// Lifecycle hooks example

var (
	startTime time.Time
	requests  int
)

// Startup hook - called when app starts
func OnStartup() {
	log.Println("🚀 Application starting...")
	startTime = time.Now()
	requests = 0
	log.Println("✅ Initialization complete")
}

// Shutdown hook - called when app stops
func OnShutdown() {
	uptime := time.Since(startTime)
	log.Printf("⏱️  Application uptime: %v", uptime)
	log.Printf("📊 Total requests handled: %d", requests)
	log.Println("👋 Application shutting down...")
}

// Status handler
func StatusHandler() map[string]any {
	requests++
	return map[string]any{
		"status":         "running",
		"uptime_seconds": time.Since(startTime).Seconds(),
		"requests":       requests,
		"started_at":     startTime.Format(time.RFC3339),
	}
}

// Home handler
func HomeHandler() string {
	requests++
	return `
	<html>
	<body>
		<h1>Lifecycle Hooks Example</h1>
		<p>Demonstrates startup and shutdown hooks</p>
		<h2>Endpoints</h2>
		<ul>
			<li><a href="/status">Status</a> - View application status</li>
		</ul>
		<h2>Lifecycle Events</h2>
		<ul>
			<li>Startup - Initialize resources, log start time</li>
			<li>Shutdown - Cleanup, log statistics</li>
		</ul>
		<p>Check server logs to see lifecycle events</p>
		<p>Press Ctrl+C to trigger shutdown hook</p>
	</body>
	</html>
	`
}

func main() {
	// Call startup hook
	OnStartup()

	// Defer shutdown hook
	defer OnShutdown()

	// Create router
	router := lokstra.NewRouter("main")
	router.GET("/", HomeHandler)
	router.GET("/status", StatusHandler)

	// Create app
	app := lokstra.NewApp("lifecycle-hooks", ":3070", router)

	log.Println("Server starting on :3070")
	log.Println("Press Ctrl+C to see shutdown hook in action")

	// Run with graceful shutdown
	if err := app.Run(30 * time.Second); err != nil {
		log.Printf("Server error: %v", err)
	}
}
