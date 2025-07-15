package iface

type Service any

type ServiceFactory = func(config any) (Service, error)

type ServiceMeta struct {
	Description string
	Tags        []string // Tags for categorization
}

type ServiceModule interface {
	Name() string
	Factory(config any) (Service, error)
	Meta() *ServiceMeta
}
