package service

type ServiceFactory = func(serviceName string, config any) (Service, error)

type Service interface {
	GetServiceUri() string
}
