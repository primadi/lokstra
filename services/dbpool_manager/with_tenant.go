package dbpool_manager

import (
	"context"

	"github.com/primadi/lokstra/serviceapi"
)

type pgxDbPoolWithTenant struct {
	pool     serviceapi.DbPool
	schema   string
	tenantID string
}

var _ serviceapi.DbPoolWithTenant = (*pgxDbPoolWithTenant)(nil)

func (p *pgxDbPoolWithTenant) Acquire(ctx context.Context) (serviceapi.DbConn, error) {
	return p.pool.AcquireMultiTenant(ctx, p.schema, p.tenantID)
}

func (p *pgxDbPoolWithTenant) Shutdown() error {
	return p.pool.Shutdown()
}
