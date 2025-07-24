package rpc_service

import "github.com/primadi/lokstra/core/registration"

func RegisterModule(ctx registration.Context) error {
	svc, err := NewRpcServer("default", nil)
	if err != nil {
		return err
	}
	ctx.RegisterService(svc)

	return nil
}
