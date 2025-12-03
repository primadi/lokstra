package dbpool_pg

import (
	"context"

	"github.com/primadi/lokstra/serviceapi"
)

type PgxDbPoolWithSchema struct {
	pool   serviceapi.DbPool
	schema string
}

func NewDbPoolWithSchema(pool serviceapi.DbPool, schema string) serviceapi.DbPoolWithSchema {
	return &PgxDbPoolWithSchema{
		pool:   pool,
		schema: schema,
	}
}

var _ serviceapi.DbPoolWithSchema = (*PgxDbPoolWithSchema)(nil)

func (p *PgxDbPoolWithSchema) Acquire(ctx context.Context) (serviceapi.DbConn, error) {
	return p.pool.Acquire(ctx, p.schema)
}

func (p *PgxDbPoolWithSchema) Shutdown() error {
	return p.pool.Shutdown()
}
