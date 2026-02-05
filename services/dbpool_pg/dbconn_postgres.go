package dbpool_pg

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/primadi/lokstra/serviceapi"
)

type pgxConnWrapper struct {
	conn     *pgxpool.Conn
	poolName string // Pool name for transaction tracking
}

// getExecutor returns the appropriate executor based on transaction context.
// If a transaction is active, it returns the transaction.
// Otherwise, it returns the connection itself.
func (c *pgxConnWrapper) getExecutor(ctx context.Context) (dbExecutor, error) {
	// Check if there's an active transaction for this pool name
	if txCtx := serviceapi.GetTransaction(ctx, c.poolName); txCtx != nil {
		// Transaction already created? Reuse it
		if txCtx.Tx != nil {
			return txCtx.Tx.(*pgxTxWrapper).tx, nil
		}

		// Lazy create transaction on first use
		tx, err := c.conn.Begin(ctx)
		if err != nil {
			return nil, err
		}

		// Repository in context for reuse
		txCtx.Tx = &pgxTxWrapper{tx: tx, txCtx: txCtx}
		txCtx.Conn = c

		return tx, nil
	}

	// No transaction - use connection directly
	return c.conn, nil
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
	executor, err := c.getExecutor(ctx)
	if err != nil {
		return nil, err
	}
	tag, err := executor.Exec(ctx, query, args...)
	return serviceapi.NewCommandResult(tag.RowsAffected), err
}

func (c *pgxConnWrapper) Query(ctx context.Context, query string, args ...any) (serviceapi.Rows, error) {
	executor, err := c.getExecutor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := executor.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &pgxRowIterator{rows: rows}, nil
}

func (c *pgxConnWrapper) QueryRow(ctx context.Context, query string, args ...any) serviceapi.Row {
	executor, err := c.getExecutor(ctx)
	if err != nil {
		return &errorRowConn{err: err}
	}
	return executor.QueryRow(ctx, query, args...)
}

// errorRowConn implements serviceapi.Row for error cases in Conn
type errorRowConn struct {
	err error
}

func (e *errorRowConn) Scan(dest ...any) error {
	return e.err
}

func (c *pgxConnWrapper) IsErrorNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func (c *pgxConnWrapper) SelectOne(ctx context.Context, query string, args []any, dest ...any) error {
	return c.conn.QueryRow(ctx, query, args...).Scan(dest...)
}

func (c *pgxConnWrapper) SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error {
	executor, err := c.getExecutor(ctx)
	if err != nil {
		return err
	}
	return pgxSelectMustOne(ctx, executor, query, args, dest...)
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
	executor, err := c.getExecutor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := executor.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectOneRow(rows, pgx.RowToMap)
}

func (c *pgxConnWrapper) SelectManyRowMap(ctx context.Context, query string,
	args ...any) ([]serviceapi.RowMap, error) {
	executor, err := c.getExecutor(ctx)
	if err != nil {
		return nil, err
	}
	return pgxSelectMany(ctx, executor, query, args...)
}

func pgxSelectMany(ctx context.Context,
	conn dbExecutor, query string, args ...any) ([]serviceapi.RowMap, error) {
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToMap)
}

func (c *pgxConnWrapper) SelectManyWithMapper(ctx context.Context,
	fnScan func(serviceapi.Row) (any, error), query string, args ...any) (any, error) {

	executor, err := c.getExecutor(ctx)
	if err != nil {
		return nil, err
	}
	return pgxSelectManyWithMapper(ctx, executor, fnScan, query, args...)
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
	executor, err := c.getExecutor(ctx)
	if err != nil {
		return false, err
	}
	var exists bool
	err = executor.QueryRow(ctx, fmt.Sprintf("SELECT EXISTS(%s)", query), args...).Scan(&exists)
	return exists, err
}

func (c *pgxConnWrapper) Begin(ctx context.Context) (serviceapi.DbTx, error) {
	// Check if already in transaction context
	if txCtx := serviceapi.GetTransaction(ctx, c.poolName); txCtx != nil {
		if txCtx.Tx != nil {
			// Already in transaction, increment counter and return it
			txCtx.IncrementCounter()
			return txCtx.Tx, nil
		}
		// Lazy create transaction
		tx, err := c.conn.Begin(ctx)
		if err != nil {
			return nil, err
		}
		txCtx.Tx = &pgxTxWrapper{tx: tx, txCtx: txCtx}
		txCtx.Conn = c
		return txCtx.Tx, nil
	}

	// Create new transaction (no TxContext)
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &pgxTxWrapper{tx: tx}, nil
}

func (c *pgxConnWrapper) Transaction(ctx context.Context, fn func(tx serviceapi.DbExecutor) error) error {
	// Check if already in transaction context
	if txCtx := serviceapi.GetTransaction(ctx, c.poolName); txCtx != nil {
		if txCtx.Tx != nil {
			// Already in transaction, just execute the function
			return fn(txCtx.Tx)
		}
		// Lazy create transaction
		tx, err := c.conn.Begin(ctx)
		if err != nil {
			return err
		}
		txCtx.Tx = &pgxTxWrapper{tx: tx, txCtx: txCtx}
		txCtx.Conn = c
		if err := fn(txCtx.Tx); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		return tx.Commit(ctx)
	}

	// Create new transaction (no TxContext)
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

var _ serviceapi.DbConn = (*pgxConnWrapper)(nil)
