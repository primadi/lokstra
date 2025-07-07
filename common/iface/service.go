package iface

type Service any

type ServiceFactory = func(config any) (Service, error)
