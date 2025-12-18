package dbpool_pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/primadi/lokstra/serviceapi"

	"github.com/jackc/pgx/v5"
)

type pgxTxWrapper struct {
	tx    pgx.Tx
	txCtx *serviceapi.TxContext // Reference to context for state updates
}

// Begin implements Tx.
func (p *pgxTxWrapper) Begin(ctx context.Context) (serviceapi.DbTx, error) {
	return nil, errors.New("nested transactions not supported")
}

// Commit implements Tx.
func (p *pgxTxWrapper) Commit(ctx context.Context) error {
	err := p.tx.Commit(ctx)
	if err == nil && p.txCtx != nil {
		p.txCtx.SetCommitted()
	}
	return err
}

// Exec implements Tx.
func (p *pgxTxWrapper) Exec(ctx context.Context, query string, args ...any) (serviceapi.CommandResult, error) {
	tag, err := p.tx.Exec(ctx, query, args...)
	return serviceapi.NewCommandResult(tag.RowsAffected), err
}

// IsExists implements Tx.
func (p *pgxTxWrapper) IsExists(ctx context.Context, query string, args ...any) (bool, error) {
	var exists bool
	err := p.tx.QueryRow(ctx, fmt.Sprintf("SELECT EXISTS(%s)", query), args...).Scan(&exists)
	return exists, err
}

// Query implements Tx.
func (p *pgxTxWrapper) Query(ctx context.Context, query string, args ...any) (serviceapi.Rows, error) {
	rows, err := p.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &pgxRowIterator{rows: rows}, nil
}

// QueryRow implements Tx.
func (p *pgxTxWrapper) QueryRow(ctx context.Context, query string, args ...any) serviceapi.Row {
	return p.tx.QueryRow(ctx, query, args...)
}

// Release implements Tx.
func (p *pgxTxWrapper) Release() error {
	return errors.New("transaction does not support Release")
}

// IsErrorNoRows implements Tx.
func (p *pgxTxWrapper) IsErrorNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// Rollback implements Tx.
func (p *pgxTxWrapper) Rollback(ctx context.Context) error {
	err := p.tx.Rollback(ctx)
	if err == nil && p.txCtx != nil {
		p.txCtx.SetRolledBack()
	}
	return err
}

func (p *pgxTxWrapper) SelectOneRowMap(ctx context.Context, query string,
	args ...any) (serviceapi.RowMap, error) {
	rows, err := p.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectOneRow(rows, pgx.RowToMap)
}

// SelectMany implements Tx.
func (p *pgxTxWrapper) SelectManyRowMap(ctx context.Context, query string,
	args ...any) ([]map[string]any, error) {
	return pgxSelectMany(ctx, p.tx, query, args...)
}

// SelectManyWithMapper implements Tx.
func (p *pgxTxWrapper) SelectManyWithMapper(ctx context.Context,
	fnScan func(serviceapi.Row) (any, error), query string,
	args ...any) (any, error) {
	return pgxSelectManyWithMapper(ctx, p.tx, fnScan, query, args...)

}

// SelectMustOne implements Tx.
func (p *pgxTxWrapper) SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error {
	return pgxSelectMustOne(ctx, p.tx, query, args, dest...)
}

// SelectOne implements Tx.
func (p *pgxTxWrapper) SelectOne(ctx context.Context, query string, args []any, dest ...any) error {
	return p.tx.QueryRow(ctx, query, args...).Scan(dest...)
}

// Transaction implements Tx.
func (p *pgxTxWrapper) Transaction(ctx context.Context, fn func(tx serviceapi.DbConn) error) error {
	return errors.New("nested transactions not supported")
}

var _ serviceapi.DbTx = (*pgxTxWrapper)(nil)
