package serviceapi

import (
	"net/http"
	"time"
)

const HTTP_LISTENER_PREFIX string = "lokstra.http_listener."

type HttpListener interface {
	// ListenAndServe starts the HTTP server on the specified address.
	// It returns an error if the server fails to start.
	ListenAndServe(addr string, handler http.Handler) error
	// Shutdown gracefully stops the HTTP server.
	// It waits for all active requests to finish before shutting down.
	Shutdown(shutdownTimeout time.Duration) error
	// IsRunning checks if the HTTP server is currently running.
	IsRunning() bool
	// ActiveRequest returns the number of currently active requests.
	ActiveRequest() int

	// GetStartMessage returns a message indicating where the server is listening.
	GetStartMessage(addr string) string
}
