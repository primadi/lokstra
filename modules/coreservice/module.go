package coreservice

import (
	"github.com/primadi/lokstra/core/iface"
)

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
	// skip register, because this module is registered using defaults package

	return nil
}

var _ iface.Module = (*CoreServiceModule)(nil)

func GetModule() iface.Module {
	return &CoreServiceModule{}
}
