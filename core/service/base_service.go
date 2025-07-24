package service

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
