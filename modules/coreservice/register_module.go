package coreservice

import (
	"github.com/primadi/lokstra/common/module"
	"github.com/primadi/lokstra/modules/coreservice/listener"
	"github.com/primadi/lokstra/modules/coreservice/router_engine"
	"github.com/primadi/lokstra/serviceapi"
)

func RegisterModule(ctx module.RegistrationContext) error {
	// Register Listener as Service Factories
	ctx.RegisterServiceFactory(serviceapi.NETHTTP_LISTENER_NAME, listener.NewNetHttpListener)
	ctx.RegisterServiceFactory(serviceapi.FASTHTTP_LISTENER_NAME, listener.NewFastHttpListener)
	ctx.RegisterServiceFactory(serviceapi.SECURE_NETHTTP_LISTENER_NAME, listener.NewSecureNetHttpListener)
	ctx.RegisterServiceFactory(serviceapi.HTTP3_LISTENER_NAME, listener.NewHttp3Listener)

	// Register Router Engine as Service Factory
	ctx.RegisterServiceFactory(serviceapi.HTTPROUTER_ROUTER_ENGINE_NAME, router_engine.NewHttpRouterEngine)
	ctx.RegisterServiceFactory(serviceapi.SERVEMUX_ROUTER_ENGINE_NAME, router_engine.NewServeMuxEngine)

	return nil
}
