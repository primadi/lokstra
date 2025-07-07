package coreservice_module

import (
	"lokstra/common/component"
	"lokstra/modules/coreservice_module/listener"
	"lokstra/modules/coreservice_module/router_engine"
	"lokstra/serviceapi/core_service"
)

func Register(ctx component.ComponentContext) error {
	// Register Listener as Service Factories
	ctx.RegisterServiceFactory(core_service.NETHTTP_LISTENER_NAME, listener.NewNetHttpListener)
	ctx.RegisterServiceFactory(core_service.FASTHTTP_LISTENER_NAME, listener.NewFastHttpListener)
	ctx.RegisterServiceFactory(core_service.SECURE_NETHTTP_LISTENER_NAME, listener.NewSecureNetHttpListener)

	// Register Router Engine as Service Factory
	ctx.RegisterServiceFactory(core_service.HTTPROUTER_ROUTER_ENGINE_NAME, router_engine.NewHttpRouterEngine)
	ctx.RegisterServiceFactory(core_service.SERVEMUX_ROUTER_ENGINE_NAME, router_engine.NewServeMuxEngine)

	return nil
}
