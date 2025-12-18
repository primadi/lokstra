package dbpool_pg

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/primadi/lokstra/serviceapi"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// errorRow implements serviceapi.Row for error cases
type errorRow struct {
	err error
}

func (e *errorRow) Scan(dest ...any) error {
	return e.err
}

// rowsWrapper wraps serviceapi.Rows and auto-releases connection on Close
type rowsWrapper struct {
	rows serviceapi.Rows
	conn serviceapi.DbConn
}

var _ serviceapi.Rows = (*rowsWrapper)(nil)

func (r *rowsWrapper) Next() bool {
	return r.rows.Next()
}

func (r *rowsWrapper) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *rowsWrapper) Close() error {
	defer r.conn.Release()
	return r.rows.Close()
}

func (r *rowsWrapper) Err() error {
	return r.rows.Err()
}

// rowWrapper wraps serviceapi.Row and auto-releases connection on Scan
type rowWrapper struct {
	row  serviceapi.Row
	conn serviceapi.DbConn
}

var _ serviceapi.Row = (*rowWrapper)(nil)

func (r *rowWrapper) Scan(dest ...any) error {
	defer r.conn.Release()
	return r.row.Scan(dest...)
}

type pgxPostgresPool struct {
	pool       *pgxpool.Pool
	poolName   string // Pool name for transaction tracking
	dsn        string
	schema     string
	rlsContext map[string]string
}

// Begin implements serviceapi.DbPool.
func (p *pgxPostgresPool) Begin(ctx context.Context) (serviceapi.DbTx, error) {
	panic("Begin() should not be called directly on DbPool. Use Acquire() first to get a DbConn, then call Begin() on it.")
}

// Exec implements serviceapi.DbPool.
func (p *pgxPostgresPool) Exec(ctx context.Context, query string, args ...any) (serviceapi.CommandResult, error) {
	cn, err := p.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer cn.Release()
	return cn.Exec(ctx, query, args...)
}

// IsErrorNoRows implements serviceapi.DbPool.
func (p *pgxPostgresPool) IsErrorNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// IsExists implements serviceapi.DbPool.
func (p *pgxPostgresPool) IsExists(ctx context.Context, query string, args ...any) (bool, error) {
	cn, err := p.Acquire(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer cn.Release()
	return cn.IsExists(ctx, query, args...)
}

// Ping implements serviceapi.DbPool.
func (p *pgxPostgresPool) Ping(ctx context.Context) error {
	cn, err := p.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer cn.Release()
	return cn.Ping(ctx)
}

// Query implements serviceapi.DbPool.
func (p *pgxPostgresPool) Query(ctx context.Context, query string, args ...any) (serviceapi.Rows, error) {
	cn, err := p.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	rows, err := cn.Query(ctx, query, args...)
	if err != nil {
		cn.Release()
		return nil, err
	}
	// Wrap rows to auto-release connection on Close()
	return &rowsWrapper{rows: rows, conn: cn}, nil
}

// QueryRow implements serviceapi.DbPool.
func (p *pgxPostgresPool) QueryRow(ctx context.Context, query string, args ...any) serviceapi.Row {
	cn, err := p.Acquire(ctx)
	if err != nil {
		return &errorRow{err: err}
	}
	row := cn.QueryRow(ctx, query, args...)
	// Wrap row to auto-release connection on Scan()
	return &rowWrapper{row: row, conn: cn}
}

// Release implements serviceapi.DbPool.
func (p *pgxPostgresPool) Release() error {
	panic("Release() should not be called directly on DbPool. Use Acquire() first to get a DbConn, then call Release() on it.")
}

// SelectManyRowMap implements serviceapi.DbPool.
func (p *pgxPostgresPool) SelectManyRowMap(ctx context.Context, query string, args ...any) ([]serviceapi.RowMap, error) {
	cn, err := p.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer cn.Release()
	return cn.SelectManyRowMap(ctx, query, args...)
}

// SelectManyWithMapper implements serviceapi.DbPool.
func (p *pgxPostgresPool) SelectManyWithMapper(ctx context.Context, fnScan func(serviceapi.Row) (any, error), query string, args ...any) (any, error) {
	cn, err := p.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer cn.Release()
	return cn.SelectManyWithMapper(ctx, fnScan, query, args...)
}

// SelectMustOne implements serviceapi.DbPool.
func (p *pgxPostgresPool) SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error {
	cn, err := p.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer cn.Release()
	return cn.SelectMustOne(ctx, query, args, dest...)
}

// SelectOne implements serviceapi.DbPool.
func (p *pgxPostgresPool) SelectOne(ctx context.Context, query string, args []any, dest ...any) error {
	cn, err := p.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer cn.Release()
	return cn.SelectOne(ctx, query, args, dest...)
}

// SelectOneRowMap implements serviceapi.DbPool.
func (p *pgxPostgresPool) SelectOneRowMap(ctx context.Context, query string, args ...any) (serviceapi.RowMap, error) {
	cn, err := p.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer cn.Release()
	return cn.SelectOneRowMap(ctx, query, args...)
}

// Transaction implements serviceapi.DbPool.
func (p *pgxPostgresPool) Transaction(ctx context.Context, fn func(tx serviceapi.DbExecutor) error) error {
	cn, err := p.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer cn.Release()
	return cn.Transaction(ctx, fn)
}

// SetSchemaRls implements serviceapi.DbPoolSchemaRls.
func (p *pgxPostgresPool) SetSchemaRls(schema string, rlsContext map[string]string) {
	p.schema = schema
	p.rlsContext = rlsContext
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

	if len(p.rlsContext) > 0 {
		// Build all SET LOCAL statements in one query
		var stmts []string
		for key, value := range p.rlsContext {
			// Key is sanitized as identifier, value is quoted as literal string
			stmt := fmt.Sprintf("SET LOCAL %s = '%s'", pgx.Identifier{key}.Sanitize(), value)
			stmts = append(stmts, stmt)
		}
		// Execute all statements in a single call
		combinedStmt := strings.Join(stmts, "; ")
		if _, err := conn.Exec(ctx, combinedStmt); err != nil {
			conn.Release()
			return nil, fmt.Errorf("failed to set RLS context: %w", err)
		}
	}
	return &pgxConnWrapper{
		conn:     conn,
		poolName: p.poolName,
	}, nil
}

var _ serviceapi.DbPool = (*pgxPostgresPool)(nil)
var _ serviceapi.DbPoolSchemaRls = (*pgxPostgresPool)(nil)

func NewPgxPostgresPool(poolName string, dsn string, schema string, rlsContext map[string]string) (*pgxPostgresPool, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &pgxPostgresPool{
		pool:       pool,
		poolName:   poolName,
		dsn:        dsn,
		schema:     schema,
		rlsContext: rlsContext,
	}, nil
}
