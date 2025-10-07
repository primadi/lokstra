package lokstra_registry

import (
	"fmt"
)

type Shutdownable interface {
	Shutdown() error
}

// Gracefully shutdown all services that implement the Shutdownable interface.
func ShutdownServices() {
	serviceMutex.RLock()
	// Create a snapshot to avoid holding lock during shutdown
	snapshot := make(map[string]any, len(serviceRegistry))
	for k, v := range serviceRegistry {
		snapshot[k] = v
	}
	serviceMutex.RUnlock()

	for name, svc := range snapshot {
		if shutdownable, ok := svc.(Shutdownable); ok {
			if err := shutdownable.Shutdown(); err != nil {
				fmt.Printf("[ShutdownServices] Failed to shutdown service %s: %v\n", name, err)
			}
		}
	}
	fmt.Println("[ShutdownServices] gracefully shutdown all services.")
}
