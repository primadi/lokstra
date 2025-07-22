package service

type ServiceFactory = func(serviceName string, config any) (Service, error)

type ServiceMeta struct {
	Description string
	Tags        []string // Tags for categorization
}

type ServiceModule interface {
	FactoryName() string
	Factory(serviceName string, config any) (Service, error)
	Meta() *ServiceMeta
}

type BaseService struct {
	name string
}

func NewBaseService(serviceName string) *BaseService {
	return &BaseService{
		name: serviceName,
	}
}

func (s *BaseService) GetServiceName() string {
	if s.name == "" {
		panic("Service name is not set")
	}
	return s.name
}

type Service interface {
	GetServiceUri() string
}
