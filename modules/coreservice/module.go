package coreservice

import (
	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/core/registration"
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
func (c *CoreServiceModule) Register(regCtx iface.RegistrationContext) error {
	// skip register, because this module is registered on standardservices package

	return nil
}

var _ registration.Module = (*CoreServiceModule)(nil)

func GetModule() registration.Module {
	return &CoreServiceModule{}
}
