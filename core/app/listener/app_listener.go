package listener

import (
	"time"
)

type AppListener interface {
	// Listen port and serve HTTP requests.
	ListenAndServe() error
	// gracefully shutdown the listener within the given timeout duration.
	Shutdown(timeout time.Duration) error
	// get the number of active requests being handled by the listener.
	ActiveRequests() int
}
