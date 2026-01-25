package serviceapi

import (
	"context"
	"fmt"
)

// --------------------
// Transaction Context Support
// --------------------

type contextKey string

const (
	txContextKeyPrefix   contextKey = "db_tx_"
	ignoreTransactionKey contextKey = "db_ignore_tx"
)

// TxContext holds transaction state for lazy transaction management
type TxContext struct {
	PoolName   string // Name of the pool (e.g., "db_auth")
	Tx         DbTx
	Conn       DbConn
	Counter    int
	committed  bool
	rolledBack bool
}

// IncrementCounter increments the transaction counter for nested transaction tracking
func (t *TxContext) IncrementCounter() {
	t.Counter++
}

// DecrementCounter decrements the transaction counter
func (t *TxContext) DecrementCounter() {
	t.Counter--
}

// SetCommitted marks the transaction as committed
func (t *TxContext) SetCommitted() {
	t.committed = true
}

// SetRolledBack marks the transaction as rolled back
func (t *TxContext) SetRolledBack() {
	t.rolledBack = true
}

// IsCommitted checks if transaction has been committed
func (t *TxContext) IsCommitted() bool {
	return t.committed
}

// IsRolledBack checks if transaction has been rolled back
func (t *TxContext) IsRolledBack() bool {
	return t.rolledBack
}

// BeginTransaction marks the context as needing a transaction for the specified pool name.
// The transaction is created lazily on first database operation.
// Returns a new context and a finalize function that should be deferred.
//
// Usage:
//
//	func (s *Service) DoWork(ctx context.Context) (err error) {
//		ctx, finish := serviceapi.BeginTransaction(ctx, "db_auth")
//		defer finish(&err)
//
//		// All operations using ctx will join the same transaction
//		s.Repository1.Create(ctx, ...)
//		s.Repository2.Update(ctx, ...)
//
//		return nil // Auto-commit on success, rollback on error
//	}
func BeginTransaction(ctx context.Context, poolName string) (context.Context, func(*error)) {
	txKey := contextKey(fmt.Sprintf("%s%s", txContextKeyPrefix, poolName))

	// Check if transaction already exists for this pool name
	if txCtx, ok := ctx.Value(txKey).(*TxContext); ok {
		// Nested call - increment counter (pseudo-nested transaction)
		txCtx.Counter++
		return ctx, func(err *error) {
			txCtx.Counter--
			if txCtx.Counter == 0 {
				finalizeTx(ctx, txCtx, err)
			}
		}
	}

	// Create new transaction context (lazy - not created yet)
	txCtx := &TxContext{
		PoolName: poolName,
		Counter:  1,
	}

	newCtx := context.WithValue(ctx, txKey, txCtx)

	return newCtx, func(err *error) {
		txCtx.Counter--
		if txCtx.Counter == 0 {
			finalizeTx(newCtx, txCtx, err)
		}
	}
}

// finalizeTx commits or rolls back the transaction based on error state
func finalizeTx(ctx context.Context, txCtx *TxContext, err *error) {
	if txCtx.Tx == nil {
		// No transaction was ever created (no database operations were executed)
		return
	}

	// Rollback if error occurred or already rolled back
	if txCtx.rolledBack || (err != nil && *err != nil) {
		if !txCtx.rolledBack {
			_ = txCtx.Tx.Rollback(ctx)
			txCtx.rolledBack = true
		}
	} else if !txCtx.committed {
		// Commit on success
		_ = txCtx.Tx.Commit(ctx)
		txCtx.committed = true
	}

	// Release connection
	if txCtx.Conn != nil {
		_ = txCtx.Conn.Release()
	}

	// Reset state to prevent reuse issues
	txCtx.Tx = nil
	txCtx.Conn = nil
	txCtx.committed = false
	txCtx.rolledBack = false
}

// WithoutTransaction creates a child context that explicitly ignores any parent transactions.
// Database operations using this context will use a new connection from the pool.
//
// Usage:
//
//	ctx, finish := BeginTransaction(ctx, dbPool)
//	defer finish(&err)
//
//	// This joins the transaction
//	repository.Create(ctx, tenant)
//
//	// This uses a separate connection (e.g., for audit logs that must commit)
//	isolatedCtx := serviceapi.WithoutTransaction(ctx)
//	auditRepository.Log(isolatedCtx, "tenant_created")
func WithoutTransaction(ctx context.Context) context.Context {
	return context.WithValue(ctx, ignoreTransactionKey, true)
}

// GetTransaction retrieves the transaction context for a specific pool name.
// Returns nil if no transaction exists for this pool.
func GetTransaction(ctx context.Context, poolName string) *TxContext {
	// Check if explicitly ignoring transactions
	if ignore, _ := ctx.Value(ignoreTransactionKey).(bool); ignore {
		return nil
	}

	txKey := contextKey(fmt.Sprintf("%s%s", txContextKeyPrefix, poolName))

	txCtx, _ := ctx.Value(txKey).(*TxContext)
	return txCtx
}
