package service

type ServiceFactory = func(config any) (Service, error)

type Service = any
