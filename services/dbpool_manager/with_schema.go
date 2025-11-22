package dbpool_manager

import (
	"context"

	"github.com/primadi/lokstra/serviceapi"
)

type pgxDbPoolWithSchema struct {
	pool   serviceapi.DbPool
	schema string
}

var _ serviceapi.DbPoolWithSchema = (*pgxDbPoolWithSchema)(nil)

func (p *pgxDbPoolWithSchema) Acquire(ctx context.Context) (serviceapi.DbConn, error) {
	return p.pool.Acquire(ctx, p.schema)
}

func (p *pgxDbPoolWithSchema) Shutdown() error {
	return p.pool.Shutdown()
}
