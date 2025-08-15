package flow

import (
	"fmt"
)

// txBeginStep implements Step interface for beginning a transaction
type txBeginStep struct{}

func (s *txBeginStep) Run(ctx *Context) error {
	// Check for existing transaction (fail fast on nested transactions)
	if ctx.dbTx != nil {
		return fmt.Errorf("nested transactions not supported: transaction already active")
	}

	// Get current connection
	conn, err := ctx.CurrentExecutor()
	if err != nil {
		return fmt.Errorf("failed to get DB connection for transaction: %w", err)
	}

	// Begin transaction
	tx, err := conn.Begin(ctx.StdContext())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Set transaction in context
	ctx.dbTx = tx
	return nil
}

func (s *txBeginStep) Meta() StepMeta {
	return StepMeta{
		Name: "tx.begin",
		Kind: StepTxBegin,
	}
}

// txEndStep implements Step interface for ending a transaction (commit or rollback)
type txEndStep struct {
	forceRollback bool
}

func (s *txEndStep) Run(ctx *Context) error {
	if ctx.dbTx == nil {
		return fmt.Errorf("no active transaction to commit or rollback")
	}

	var err error
	if s.forceRollback {
		err = ctx.dbTx.Rollback(ctx.StdContext())
		if err != nil {
			return fmt.Errorf("failed to rollback transaction: %w", err)
		}
	} else {
		err = ctx.dbTx.Commit(ctx.StdContext())
		if err != nil {
			// If commit fails, try to rollback
			if rollbackErr := ctx.dbTx.Rollback(ctx.StdContext()); rollbackErr != nil {
				return fmt.Errorf("failed to commit transaction: %w, and failed to rollback: %w", err, rollbackErr)
			}
			return fmt.Errorf("failed to commit transaction (rollback successful): %w", err)
		}
	}

	// Clear transaction from context
	ctx.dbTx = nil
	return nil
}

func (s *txEndStep) Meta() StepMeta {
	return StepMeta{
		Name: "tx.end",
		Kind: StepTxEnd,
	}
}

// ForceRollback sets the step to always rollback instead of commit
func (s *txEndStep) ForceRollback() *txEndStep {
	s.forceRollback = true
	return s
}
