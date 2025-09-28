package lokstra_registry

import "fmt"

type Shutdownable interface {
	Shutdown() error
}

func ShutdownServices() {
	for name, svc := range serviceRegistry {
		if shutdownable, ok := svc.(Shutdownable); ok {
			if err := shutdownable.Shutdown(); err != nil {
				fmt.Printf("[ShutdownServices] Failed to shutdown service %s: %v\n", name, err)
			}
		}
	}
	fmt.Println("[ShutdownServices] gracefully shutdown all services.")
}
