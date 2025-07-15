package pg_dbpool

import (
	"context"
	"errors"
	"fmt"
	"lokstra/serviceapi"

	"github.com/jackc/pgx/v5"
)

type pgxTxWrapper struct {
	tx pgx.Tx
}

// Begin implements Tx.
func (p *pgxTxWrapper) Begin(ctx context.Context) (serviceapi.Tx, error) {
	return nil, errors.New("nested transactions not supported")
}

// Commit implements Tx.
func (p *pgxTxWrapper) Commit(ctx context.Context) error {
	return p.tx.Commit(ctx)
}

// Exec implements Tx.
func (p *pgxTxWrapper) Exec(ctx context.Context, query string, args ...any) (serviceapi.CommandResult, error) {
	tag, err := p.tx.Exec(ctx, query, args...)
	return &pgxCommandResult{fnRowsAffected: tag.RowsAffected}, err
}

// IsExists implements Tx.
func (p *pgxTxWrapper) IsExists(ctx context.Context, query string, args ...any) (bool, error) {
	var exists bool
	err := p.tx.QueryRow(ctx, fmt.Sprintf("SELECT EXISTS(%s)", query), args...).Scan(&exists)
	return exists, err
}

// Query implements Tx.
func (p *pgxTxWrapper) Query(ctx context.Context, query string, args ...any) (serviceapi.RowIterator, error) {
	rows, err := p.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &pgxRowIterator{rows: rows}, nil
}

// QueryRow implements Tx.
func (p *pgxTxWrapper) QueryRow(ctx context.Context, query string, args ...any) serviceapi.RowScanner {
	return p.tx.QueryRow(ctx, query, args...)
}

// Release implements Tx.
func (p *pgxTxWrapper) Release() error {
	return errors.New("transaction does not support Release")
}

// Rollback implements Tx.
func (p *pgxTxWrapper) Rollback(ctx context.Context) error {
	return p.tx.Rollback(ctx)
}

// SelectMany implements Tx.
func (p *pgxTxWrapper) SelectMany(ctx context.Context, query string, args ...any) (any, error) {
	return pgxSelectMany(ctx, p.tx, query, args...)
}

// SelectManyWithMapper implements Tx.
func (p *pgxTxWrapper) SelectManyWithMapper(ctx context.Context,
	fnScan func(serviceapi.RowScanner) (any, error), query string, args ...any) (any, error) {
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
func (p *pgxTxWrapper) Transaction(ctx context.Context, fn func(tx serviceapi.DBConn) error) error {
	return errors.New("nested transactions not supported")
}

var _ serviceapi.Tx = (*pgxTxWrapper)(nil)
