package coreservice

import (
	"github.com/primadi/lokstra/common/module"
	"github.com/primadi/lokstra/modules/coreservice/listener"
	"github.com/primadi/lokstra/modules/coreservice/router_engine"
)

func RegisterModule(ctx module.RegistrationContext) error {
	// Register Listener as Service Factories
	ctx.RegisterServiceFactory(listener.NETHTTP_LISTENER_NAME, listener.NewNetHttpListener)
	ctx.RegisterServiceFactory(listener.FASTHTTP_LISTENER_NAME, listener.NewFastHttpListener)
	ctx.RegisterServiceFactory(listener.SECURE_NETHTTP_LISTENER_NAME, listener.NewSecureNetHttpListener)
	ctx.RegisterServiceFactory(listener.HTTP3_LISTENER_NAME, listener.NewHttp3Listener)

	// Register Router Engine as Service Factory
	ctx.RegisterServiceFactory(router_engine.HTTPROUTER_ROUTER_ENGINE_NAME, router_engine.NewHttpRouterEngine)
	ctx.RegisterServiceFactory(router_engine.SERVEMUX_ROUTER_ENGINE_NAME, router_engine.NewServeMuxEngine)

	return nil
}
