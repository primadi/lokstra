package old_registry

import (
	"fmt"
)

type Shutdownable interface {
	Shutdown() error
}

// Gracefully shutdown all services that implement the Shutdownable interface.
func ShutdownServices() {
	// Create a snapshot to avoid issues during shutdown
	snapshot := make(map[string]any)
	serviceRegistry.Range(func(key, value any) bool {
		snapshot[key.(string)] = value
		return true
	})

	for name, svc := range snapshot {
		if shutdownable, ok := svc.(Shutdownable); ok {
			if err := shutdownable.Shutdown(); err != nil {
				fmt.Printf("[ShutdownServices] Failed to shutdown service %s: %v\n", name, err)
			}
		}
	}
	fmt.Println("[ShutdownServices] gracefully shutdown all services.")
}
