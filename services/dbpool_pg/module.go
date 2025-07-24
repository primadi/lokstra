package dbpool_pg

import (
	"github.com/primadi/lokstra/core/registration"
)

type dbPoolPgModule struct{}

// Name implements registration.Module.
func (d *dbPoolPgModule) Name() string {
	return "dbpool_pg"
}

// Register implements registration.Module.
func (d *dbPoolPgModule) Register(regCtx registration.Context) error {
	regCtx.RegisterServiceFactory(d.Name(), ServiceFactory)

	return nil
}

// Description implements service.Module.
func (d *dbPoolPgModule) Description() string {
	return "PostgreSQL Database Pool Service Module"
}

var _ registration.Module = (*dbPoolPgModule)(nil)

func GetModule() registration.Module {
	return &dbPoolPgModule{}
}
