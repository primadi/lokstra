package iface

// App interface defines the methods required for an application to start and stop.
type App interface {
	Start() error
	Stop() error
}
