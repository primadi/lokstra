package service

type ServiceFactory = func(config any) (Service, error)

type Service = any

// Shutdownable is an optional interface that a service can implement to perform
// graceful shutdown tasks, such as releasing resources or saving state.
type Shutdownable interface {
	Shutdown() error
}
