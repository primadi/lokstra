package iface

type Service any

type ServiceFactory = func(config any) (Service, error)

type WithStop interface {
	Stop() error
}
