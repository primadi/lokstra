package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/primadi/lokstra"
)

// Graceful shutdown example

var (
	activeRequests int
	shutdownSignal chan os.Signal
)

// Slow handler simulates long-running request
func SlowHandler() map[string]any {
	activeRequests++
	defer func() { activeRequests-- }()

	log.Printf("Request started (active: %d)", activeRequests)

	// Simulate slow processing
	time.Sleep(5 * time.Second)

	log.Printf("Request completed (active: %d)", activeRequests)

	return map[string]any{
		"message": "Slow request completed",
		"took":    "5 seconds",
	}
}

// Fast handler
func FastHandler() map[string]any {
	activeRequests++
	defer func() { activeRequests-- }()

	return map[string]any{
		"message": "Fast response",
	}
}

// Status handler
func StatusHandler() map[string]any {
	return map[string]any{
		"status":          "running",
		"active_requests": activeRequests,
	}
}

// Home handler
func HomeHandler() string {
	return `
	<html>
	<body>
		<h1>Graceful Shutdown Example</h1>
		<p>Demonstrates proper shutdown handling</p>
		<h2>Endpoints</h2>
		<ul>
			<li><a href="/status">Status</a> - View active requests</li>
			<li><a href="/fast">Fast</a> - Quick response</li>
			<li><a href="/slow">Slow</a> - 5-second delay</li>
		</ul>
		<h2>Testing Graceful Shutdown</h2>
		<ol>
			<li>Make a request to /slow (takes 5 seconds)</li>
			<li>Press Ctrl+C to trigger shutdown</li>
			<li>Server waits for request to complete</li>
			<li>Server shuts down gracefully</li>
		</ol>
		<p>Watch server logs to see shutdown behavior</p>
	</body>
	</html>
	`
}

func main() {
	// Setup signal handling
	shutdownSignal = make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	// Create router
	router := lokstra.NewRouter("main")
	router.GET("/", HomeHandler)
	router.GET("/status", StatusHandler)
	router.GET("/fast", FastHandler)
	router.GET("/slow", SlowHandler)

	// Create app
	app := lokstra.NewApp("graceful-shutdown", ":3090", router)

	log.Println("Server starting on :3090")
	log.Println("Try making a /slow request, then press Ctrl+C")

	// Start server in goroutine
	go func() {
		if err := app.Run(30 * time.Second); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdownSignal

	log.Println("🛑 Shutdown signal received!")
	log.Printf("⏳ Waiting for %d active requests to complete...", activeRequests)

	// Give time for active requests to finish
	shutdownTimeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("⚠️  Shutdown timeout reached!")
			return
		case <-ticker.C:
			if activeRequests == 0 {
				log.Println("✅ All requests completed")
				log.Println("👋 Server shutdown complete")
				return
			}
			log.Printf("⏳ Still waiting... (active: %d)", activeRequests)
		}
	}
}
