package main

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/primadi/lokstra"
)

// This example demonstrates graceful shutdown handling in Lokstra applications.
// It shows how to handle shutdown signals, cleanup resources, and ensure data integrity.
//
// Learning Objectives:
// - Understand graceful shutdown patterns
// - Learn signal handling for clean termination
// - See resource cleanup and connection handling
// - Understand graceful request completion
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/getting-started.md#graceful-shutdown

// RequestCounter tracks active requests for graceful shutdown
var activeRequests int64

// simulateDBConnection represents a database connection that needs cleanup
type simulateDBConnection struct {
	connected bool
}

func (db *simulateDBConnection) Close() error {
	lokstra.Logger.Infof("🔌 Closing database connection...")
	db.connected = false
	time.Sleep(100 * time.Millisecond) // Simulate cleanup time
	lokstra.Logger.Infof("✅ Database connection closed")
	return nil
}

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()

	// Create application
	app := lokstra.NewApp(regCtx, "graceful-shutdown-app", ":8080")

	// Simulate external resources that need cleanup
	dbConnection := &simulateDBConnection{connected: true}

	// Middleware to track active requests
	app.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		// Increment active request counter
		atomic.AddInt64(&activeRequests, 1)
		lokstra.Logger.Debugf("📈 Active requests: %d", atomic.LoadInt64(&activeRequests))

		// Process request
		err := next(ctx)

		// Decrement active request counter
		atomic.AddInt64(&activeRequests, -1)
		lokstra.Logger.Debugf("📉 Active requests: %d", atomic.LoadInt64(&activeRequests))

		return err
	})

	// Routes that simulate various processing times
	app.GET("/quick", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message":      "Quick response",
			"process_time": "fast",
		})
	})

	app.GET("/slow", func(ctx *lokstra.Context) error {
		lokstra.Logger.Infof("🐌 Processing slow request...")
		time.Sleep(5 * time.Second) // Simulate slow processing
		return ctx.Ok(map[string]any{
			"message":      "Slow response completed",
			"process_time": "5 seconds",
		})
	})

	app.GET("/database", func(ctx *lokstra.Context) error {
		if !dbConnection.connected {
			return ctx.ErrorInternal("Database connection is closed")
		}

		// Simulate database operation
		time.Sleep(2 * time.Second)
		return ctx.Ok(map[string]any{
			"message": "Data retrieved from database",
			"records": []string{"record1", "record2", "record3"},
		})
	})

	app.GET("/health", func(ctx *lokstra.Context) error {
		status := "healthy"
		if !dbConnection.connected {
			status = "degraded"
		}

		return ctx.Ok(map[string]any{
			"status":          status,
			"active_requests": atomic.LoadInt64(&activeRequests),
			"database":        dbConnection.connected,
			"timestamp":       time.Now().Unix(),
		})
	})

	// Setup graceful shutdown
	setupGracefulShutdown(dbConnection)

	lokstra.Logger.Infof("🚀 Graceful Shutdown Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Try these endpoints:")
	lokstra.Logger.Infof("  GET /quick      - Fast response")
	lokstra.Logger.Infof("  GET /slow       - 5-second response (good for testing shutdown)")
	lokstra.Logger.Infof("  GET /database   - Database operation")
	lokstra.Logger.Infof("  GET /health     - Health check with active request count")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("To test graceful shutdown:")
	lokstra.Logger.Infof("  1. Start a slow request: curl http://localhost:8080/slow")
	lokstra.Logger.Infof("  2. While it's running, press Ctrl+C or send SIGTERM")
	lokstra.Logger.Infof("  3. Observe how the server waits for the request to complete")

	// Start the application
	if err := app.Start(); err != nil {
		lokstra.Logger.Errorf("❌ Application failed to start: %v", err)
	}
}

func setupGracefulShutdown(dbConnection *simulateDBConnection) {
	// Channel to receive OS signals
	sigChan := make(chan os.Signal, 1)

	// Register channel to receive specific signals
	signal.Notify(sigChan,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // Termination signal
		syscall.SIGQUIT, // Quit signal
	)

	// Start a goroutine to handle shutdown signals
	go func() {
		// Wait for shutdown signal
		sig := <-sigChan
		lokstra.Logger.Infof("🛑 Received shutdown signal: %v", sig)
		lokstra.Logger.Infof("🔄 Starting graceful shutdown...")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Start shutdown process
		go func() {
			// Stop accepting new requests
			lokstra.Logger.Infof("🚫 Stopping acceptance of new requests...")

			// Wait for active requests to complete
			lokstra.Logger.Infof("⏳ Waiting for active requests to complete...")
			for {
				active := atomic.LoadInt64(&activeRequests)
				if active == 0 {
					lokstra.Logger.Infof("✅ All requests completed")
					break
				}
				lokstra.Logger.Infof("⏱️  Waiting for %d active requests...", active)
				time.Sleep(500 * time.Millisecond)

				// Check if shutdown timeout reached
				select {
				case <-shutdownCtx.Done():
					lokstra.Logger.Warnf("⚠️  Shutdown timeout reached, forcing shutdown")
					return
				default:
					continue
				}
			}

			// Cleanup resources
			lokstra.Logger.Infof("🧹 Cleaning up resources...")
			if err := dbConnection.Close(); err != nil {
				lokstra.Logger.Errorf("❌ Error closing database: %v", err)
			}

			// Perform additional cleanup here
			lokstra.Logger.Infof("🔧 Performing final cleanup...")
			time.Sleep(200 * time.Millisecond) // Simulate cleanup time

			lokstra.Logger.Infof("✅ Graceful shutdown completed")
			os.Exit(0)
		}()

		// Wait for shutdown context to complete or timeout
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			lokstra.Logger.Errorf("❌ Graceful shutdown timeout, forcing exit")
			os.Exit(1)
		}
	}()
}

// Graceful Shutdown Best Practices:
//
// 1. Signal Handling:
//    - Listen for SIGINT (Ctrl+C), SIGTERM (systemd/docker), SIGQUIT
//    - Use signal.Notify() to receive OS signals
//    - Handle shutdown in separate goroutine
//
// 2. Request Completion:
//    - Track active requests with counters
//    - Stop accepting new requests first
//    - Wait for existing requests to complete
//    - Set reasonable timeout (30-60 seconds)
//
// 3. Resource Cleanup:
//    - Close database connections
//    - Flush logs and metrics
//    - Close file handles
//    - Cancel background processes
//
// 4. Timeout Handling:
//    - Use context.WithTimeout() for shutdown process
//    - Force shutdown if timeout exceeded
//    - Log timeout warnings
//
// 5. Production Considerations:
//    - Health check endpoints should reflect shutdown state
//    - Load balancers should stop sending traffic
//    - Container orchestrators (Docker, Kubernetes) respect signals
//    - Monitor shutdown duration and success rates

// Testing Graceful Shutdown:
//
// Terminal 1 (Start slow request):
//   curl http://localhost:8080/slow
//
// Terminal 2 (While slow request is running):
//   curl http://localhost:8080/health  # Check active requests
//   pkill -TERM graceful-shutdown-app  # Send termination signal
//
// Observe:
//   - Server stops accepting new requests
//   - Existing slow request completes
//   - Resources are cleaned up properly
//   - Clean exit with proper logging
