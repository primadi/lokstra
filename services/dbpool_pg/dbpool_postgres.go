package dbpool_pg

import (
	"context"

	"github.com/primadi/lokstra/serviceapi"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxPostgresPool struct {
	pool   *pgxpool.Pool
	dsn    string
	schema string
	rlsID  string
}

// SetSchemaRls implements serviceapi.DbPoolSchemaRls.
func (p *pgxPostgresPool) SetSchemaRls(schema string, rlsID string) {
	p.schema = schema
	p.rlsID = rlsID
}

// Shutdown implements serviceapi.DbPool.
func (p *pgxPostgresPool) Shutdown() error {
	p.pool.Close()
	return nil
}

func (p *pgxPostgresPool) Acquire(ctx context.Context) (serviceapi.DbConn, error) {
	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	if len(p.schema) > 0 {
		stmt := "SET search_path TO " + pgx.Identifier{p.schema}.Sanitize()
		if _, err := conn.Exec(ctx, stmt); err != nil {
			conn.Release()
			return nil, err
		}
	}

	if len(p.rlsID) > 0 {
		// set RLS context
		stmt := "SET LOCAL app.current_rls = " + pgx.Identifier{p.rlsID}.Sanitize()
		if _, err := conn.Exec(ctx, stmt); err != nil {
			conn.Release()
			return nil, err
		}
	}
	return &pgxConnWrapper{conn: conn}, nil
}

var _ serviceapi.DbPool = (*pgxPostgresPool)(nil)
var _ serviceapi.DbPoolSchemaRls = (*pgxPostgresPool)(nil)

func NewPgxPostgresPool(dsn string, schema string, rlsID string) (*pgxPostgresPool, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &pgxPostgresPool{
		pool:   pool,
		dsn:    dsn,
		schema: schema,
		rlsID:  rlsID,
	}, nil
}
