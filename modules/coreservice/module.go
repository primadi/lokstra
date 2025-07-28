package coreservice

import (
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/modules/coreservice/listener"
	"github.com/primadi/lokstra/modules/coreservice/router_engine"
)

const DEFAULT_LISTENER_NAME = "coreservice.nethttp"
const DEFAULT_ROUTER_ENGINE_NAME = "coreservice.httprouter"

type CoreServiceModule struct{}

// Description implements registration.Module.
func (c *CoreServiceModule) Description() string {
	return "Core Service Module for Lokstra"
}

// Name implements registration.Module.
func (c *CoreServiceModule) Name() string {
	return "coreservice_module"
}

// Register implements registration.Module.
func (c *CoreServiceModule) Register(regCtx registration.Context) error {
	regCtx.RegisterServiceFactory(listener.NETHTTP_LISTENER_NAME,
		listener.NewNetHttpListener)
	regCtx.RegisterServiceFactory(listener.FASTHTTP_LISTENER_NAME,
		listener.NewFastHttpListener)
	regCtx.RegisterServiceFactory(listener.SECURE_NETHTTP_LISTENER_NAME,
		listener.NewSecureNetHttpListener)
	regCtx.RegisterServiceFactory(listener.HTTP3_LISTENER_NAME,
		listener.NewHttp3Listener)

	// Register Router Engine as Service Factory
	regCtx.RegisterServiceFactory(router_engine.HTTPROUTER_ROUTER_ENGINE_NAME,
		router_engine.NewHttpRouterEngine)
	regCtx.RegisterServiceFactory(router_engine.SERVEMUX_ROUTER_ENGINE_NAME,
		router_engine.NewServeMuxEngine)

	return nil
}

var _ registration.Module = (*CoreServiceModule)(nil)

func GetModule() registration.Module {
	return &CoreServiceModule{}
}
