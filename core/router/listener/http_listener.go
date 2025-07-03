package listener

import (
	"errors"
	"net/http"
	"time"
)

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
}

// ListenerType defines the type of HTTP listener.
func NewHttpListener(listenerType ListenerType) (HttpListener, error) {
	switch listenerType {
	case NetHttpListenerType:
		return NewNetHttpListener(), nil
	case FastHttpListenerType:
		return NewFastHttpListener(), nil
	case SecureNetHttpListenerType:
		// TODO: Load Key
		// svr := core.GetServer()
		// keyfile, ok := svr.GetSetting(core.KEY_FILE)
		// if !ok {
		// 	return nil, errors.New("key file not set in server settings")
		// }
		// certfile, ok := svr.GetSetting(core.CERT_FILE)
		// if !ok {
		// 	return nil, errors.New("cert file not set in server settings")
		// }
		var certFile any = ""
		var keyFile any = ""
		return NewSecureNetHttpListener(certFile.(string), keyFile.(string)), nil
	default:
		return nil, errors.New("unsupported listener type")
	}
}
