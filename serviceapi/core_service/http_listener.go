package core_service

import (
	"net/http"
	"time"
)

const NETHTTP_LISTENER_NAME = "nethttp"
const FASTHTTP_LISTENER_NAME = "fasthttp"
const SECURE_NETHTTP_LISTENER_NAME = "secure_nethttp"
const HTTP3_LISTENER_NAME = "http3"

const DEFAULT_LISTENER_NAME = NETHTTP_LISTENER_NAME

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
	// ListenerType returns the type of the HTTP listener.
	ListenerType() string
}
