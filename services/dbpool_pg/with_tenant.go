package dbpool_pg

import (
	"context"

	"github.com/primadi/lokstra/serviceapi"
)

type PgxDbPoolWithTenant struct {
	pool     serviceapi.DbPool
	schema   string
	tenantID string
}

func NewDbPoolWithTenant(pool serviceapi.DbPool, schema string, tenantID string) serviceapi.DbPoolWithTenant {
	return &PgxDbPoolWithTenant{
		pool:     pool,
		schema:   schema,
		tenantID: tenantID,
	}
}

var _ serviceapi.DbPoolWithTenant = (*PgxDbPoolWithTenant)(nil)

func (p *PgxDbPoolWithTenant) Acquire(ctx context.Context) (serviceapi.DbConn, error) {
	return p.pool.AcquireMultiTenant(ctx, p.schema, p.tenantID)
}

func (p *PgxDbPoolWithTenant) Shutdown() error {
	return p.pool.Shutdown()
}
