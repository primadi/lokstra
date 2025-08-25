package coreservice

import (
	"github.com/primadi/lokstra/core/registration"
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
func (c *CoreServiceModule) Register(regCtx registration.Context) error {
	// skip register, because this module is registered using defaults package

	return nil
}

var _ registration.Module = (*CoreServiceModule)(nil)

func GetModule() registration.Module {
	return &CoreServiceModule{}
}
