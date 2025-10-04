package dbpool_pg

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/primadi/lokstra/serviceapi"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxPostgresPool struct {
	dsn  string
	pool *pgxpool.Pool
}

var _ serviceapi.DbPool = (*pgxPostgresPool)(nil)

func (p *pgxPostgresPool) GetSetting(key string) any {
	if key == "dsn" {
		return p.dsn
	}
	return nil
}

func NewPgxPostgresPool(ctx context.Context, dsn string) (*pgxPostgresPool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &pgxPostgresPool{
		dsn:  dsn,
		pool: pool,
	}, nil
}

func (p *pgxPostgresPool) Acquire(ctx context.Context, schema string) (serviceapi.DbConn, error) {
	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	stmt := "SET search_path TO " + pgx.Identifier{schema}.Sanitize()
	if _, err := conn.Exec(ctx, stmt); err != nil {
		conn.Release()
		return nil, err
	}

	return &pgxConnWrapper{conn: conn}, nil
}

func (p *pgxPostgresPool) AcquireMultiTenant(ctx context.Context, schema string, tenantID string) (serviceapi.DbConn, error) {
	conn, err := p.Acquire(ctx, schema)
	if err != nil {
		return nil, err
	}

	if tenantID != "" {
		// set RLS context
		stmt := "SET LOCAL app.current_tenant = " + pgx.Identifier{tenantID}.Sanitize()
		if _, err := conn.Exec(ctx, stmt); err != nil {
			conn.Release()
			return nil, err
		}
	}

	return conn, nil
}

type pgxConnWrapper struct {
	conn *pgxpool.Conn
}

// Shutdown implements serviceapi.DbConn.
func (c *pgxConnWrapper) Shutdown() error {
	c.conn.Release()
	return nil
}

// Ping implements serviceapi.DbConn.
func (c *pgxConnWrapper) Ping(context context.Context) error {
	return c.conn.Ping(context)
}

func (c *pgxConnWrapper) Exec(ctx context.Context, query string, args ...any) (serviceapi.CommandResult, error) {
	tag, err := c.conn.Exec(ctx, query, args...)
	return serviceapi.NewCommandResult(tag.RowsAffected), err
}

func (c *pgxConnWrapper) Query(ctx context.Context, query string, args ...any) (serviceapi.Rows, error) {
	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &pgxRowIterator{rows: rows}, nil
}

func (c *pgxConnWrapper) QueryRow(ctx context.Context, query string, args ...any) serviceapi.Row {
	return c.conn.QueryRow(ctx, query, args...)
}

func (c *pgxConnWrapper) IsErrorNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func (c *pgxConnWrapper) SelectOne(ctx context.Context, query string, args []any, dest ...any) error {
	return c.conn.QueryRow(ctx, query, args...).Scan(dest...)
}

func (c *pgxConnWrapper) SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error {
	return pgxSelectMustOne(ctx, c.conn, query, args, dest...)
}

func pgxSelectMustOne(ctx context.Context, conn dbExecutor, query string, args []any, dest ...any) error {
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return pgx.ErrNoRows
	}

	err = rows.Scan(dest...)
	if err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}
	if rows.Next() {
		return errors.New("selectMustOne: more than one row returned")
	}

	return nil
}

func (c *pgxConnWrapper) SelectOneRowMap(ctx context.Context, query string,
	args ...any) (serviceapi.RowMap, error) {
	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectOneRow(rows, pgx.RowToMap)
}

func (c *pgxConnWrapper) SelectManyRowMap(ctx context.Context, query string,
	args ...any) ([]serviceapi.RowMap, error) {
	return pgxSelectMany(ctx, c.conn, query, args...)
}

func pgxSelectMany(ctx context.Context,
	conn dbExecutor, query string, args ...any) ([]serviceapi.RowMap, error) {
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToMap)
	// var resultSlice []map[string]any
	// for rows.Next() {
	// 	columns := rows.FieldDescriptions()
	// 	values := make([]any, len(columns))
	// 	valuePtrs := make([]any, len(columns))

	// 	for i := range columns {
	// 		valuePtrs[i] = &values[i]
	// 	}

	// 	if err := rows.Scan(valuePtrs...); err != nil {
	// 		return nil, fmt.Errorf("failed to scan row: %w", err)
	// 	}

	// 	rowMap := make(map[string]any)
	// 	for i, col := range columns {
	// 		rowMap[string(col.Name)] = values[i]
	// 	}
	// 	resultSlice = append(resultSlice, rowMap)
	// }

	// if err := rows.Err(); err != nil {
	// 	return nil, fmt.Errorf("error iterating rows: %w", err)
	// }

	// return resultSlice, nil
}

func (c *pgxConnWrapper) SelectManyWithMapper(ctx context.Context,
	fnScan func(serviceapi.Row) (any, error), query string, args ...any) (any, error) {

	return pgxSelectManyWithMapper(ctx, c.conn, fnScan, query, args...)
}

func pgxSelectManyWithMapper(ctx context.Context,
	conn dbExecutor, fnScan func(serviceapi.Row) (any, error), query string, args ...any) (any, error) {
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	fnScanType := reflect.TypeOf(fnScan)
	returnType := fnScanType.Out(0)

	resultSlice := reflect.MakeSlice(reflect.SliceOf(returnType), 0, 10)

	for rows.Next() {
		item, err := fnScan(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		resultSlice = reflect.Append(resultSlice, reflect.ValueOf(item))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return resultSlice.Interface(), nil
}

func (c *pgxConnWrapper) IsExists(ctx context.Context, query string, args ...any) (bool, error) {
	var exists bool
	err := c.conn.QueryRow(ctx, fmt.Sprintf("SELECT EXISTS(%s)", query), args...).Scan(&exists)
	return exists, err
}

func (c *pgxConnWrapper) Begin(ctx context.Context) (serviceapi.DbTx, error) {
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &pgxTxWrapper{tx: tx}, nil
}

func (c *pgxConnWrapper) Transaction(ctx context.Context, fn func(tx serviceapi.DbExecutor) error) error {
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return err
	}
	wrapper := &pgxTxWrapper{tx: tx}
	if err := fn(wrapper); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (c *pgxConnWrapper) Release() error {
	c.conn.Release()
	return nil
}
