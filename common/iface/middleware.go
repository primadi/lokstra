package iface

import "lokstra/core/request"

type MiddlewareFunc func(next request.HandlerFunc) request.HandlerFunc

type MiddlewareHandler interface {
	GetName() string
	GetMiddlewareFunc() MiddlewareFunc
}
