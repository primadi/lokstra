package rpc_service

import "github.com/primadi/lokstra/common/module"

func RegisterModule(ctx module.RegistrationContext) error {
	svc, err := NewRpcServer("default", nil)
	if err != nil {
		return err
	}
	ctx.RegisterService(svc)

	return nil
}
