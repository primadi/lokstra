package serviceapi

type Shutdownable interface {
	Shutdown() error
}
