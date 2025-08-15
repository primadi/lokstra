package flow

import (
	"context"
	"testing"

	"github.com/primadi/lokstra/serviceapi"
)

func TestTransactionSteps(t *testing.T) {
	handler := NewHandler(nil, "test-handler")

	// Test BeginTx adds a transaction begin step
	handler.BeginTx()

	if len(handler.steps) != 1 {
		t.Errorf("Expected 1 step after BeginTx, got %d", len(handler.steps))
	}

	// Verify it's a txBeginStep
	if step, ok := handler.steps[0].(*txBeginStep); ok {
		meta := step.Meta()
		if meta.Name != "tx.begin" {
			t.Errorf("Expected step name 'tx.begin', got '%s'", meta.Name)
		}
		if meta.Kind != StepTxBegin {
			t.Errorf("Expected step kind StepTxBegin, got %v", meta.Kind)
		}
	} else {
		t.Error("Expected first step to be txBeginStep")
	}

	// Test CommitOrRollback adds a transaction end step
	handler.CommitOrRollback()

	if len(handler.steps) != 2 {
		t.Errorf("Expected 2 steps after CommitOrRollback, got %d", len(handler.steps))
	}

	// Verify it's a txEndStep
	if step, ok := handler.steps[1].(*txEndStep); ok {
		meta := step.Meta()
		if meta.Name != "tx.end" {
			t.Errorf("Expected step name 'tx.end', got '%s'", meta.Name)
		}
		if meta.Kind != StepTxEnd {
			t.Errorf("Expected step kind StepTxEnd, got %v", meta.Kind)
		}
		if step.forceRollback {
			t.Error("Expected forceRollback to be false for CommitOrRollback")
		}
	} else {
		t.Error("Expected second step to be txEndStep")
	}
}

func TestRollbackStep(t *testing.T) {
	handler := NewHandler(nil, "test-handler")

	// Test Rollback adds a forced rollback step
	handler.Rollback()

	if len(handler.steps) != 1 {
		t.Errorf("Expected 1 step after Rollback, got %d", len(handler.steps))
	}

	// Verify it's a txEndStep with forceRollback = true
	if step, ok := handler.steps[0].(*txEndStep); ok {
		meta := step.Meta()
		if meta.Name != "tx.end" {
			t.Errorf("Expected step name 'tx.end', got '%s'", meta.Name)
		}
		if meta.Kind != StepTxEnd {
			t.Errorf("Expected step kind StepTxEnd, got %v", meta.Kind)
		}
		if !step.forceRollback {
			t.Error("Expected forceRollback to be true for Rollback")
		}
	} else {
		t.Error("Expected step to be txEndStep")
	}
}

func TestTransactionChaining(t *testing.T) {
	handler := NewHandler(nil, "test-handler")

	// Test fluent API chaining
	result := handler.
		BeginTx().
		ExecSql("INSERT INTO test VALUES (?)", 1).
		Done().
		QueryRowSql("SELECT COUNT(*) FROM test").
		SaveAs("count")

	if result != handler {
		t.Error("Expected chaining to return same handler")
	}

	if len(handler.steps) != 3 {
		t.Errorf("Expected 3 steps in chain, got %d", len(handler.steps))
	}

	// Verify step types
	expectedTypes := []string{"tx.begin", "sql.exec", "sql.query_row"}
	for i, expected := range expectedTypes {
		if i >= len(handler.steps) {
			t.Errorf("Missing step %d", i)
			continue
		}

		meta := handler.steps[i].Meta()
		if meta.Name != expected {
			t.Errorf("Step %d: expected name '%s', got '%s'", i, expected, meta.Name)
		}
	}
}

func TestTxStepMeta(t *testing.T) {
	// Test txBeginStep meta
	beginStep := &txBeginStep{}
	beginMeta := beginStep.Meta()

	if beginMeta.Name != "tx.begin" {
		t.Errorf("Expected begin step name 'tx.begin', got '%s'", beginMeta.Name)
	}

	if beginMeta.Kind != StepTxBegin {
		t.Errorf("Expected begin step kind StepTxBegin, got %v", beginMeta.Kind)
	}

	// Test txEndStep meta
	endStep := &txEndStep{}
	endMeta := endStep.Meta()

	if endMeta.Name != "tx.end" {
		t.Errorf("Expected end step name 'tx.end', got '%s'", endMeta.Name)
	}

	if endMeta.Kind != StepTxEnd {
		t.Errorf("Expected end step kind StepTxEnd, got %v", endMeta.Kind)
	}
}

func TestNestedTransactionPanic(t *testing.T) {
	handler := NewHandler(nil, "test-handler")

	// First BeginTx should work
	handler.BeginTx()

	// Second BeginTx should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nested BeginTx, but no panic occurred")
		} else {
			expectedMsg := "multiple BeginTx() calls detected: nested transactions not supported"
			if r != expectedMsg {
				t.Errorf("Expected panic message '%s', got '%v'", expectedMsg, r)
			}
		}
	}()

	handler.BeginTx() // This should panic
}

func TestNestedTransactionRuntimeError(t *testing.T) {
	// Test runtime detection of nested transactions
	ctx := NewContext(context.Background(), nil, "test")

	// Simulate already having a transaction
	ctx.dbTx = &mockTx{} // Mock transaction object

	step := &txBeginStep{}
	err := step.Run(ctx)

	if err == nil {
		t.Error("Expected error for nested transaction at runtime")
	}

	expectedMsg := "nested transactions not supported: transaction already active"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// mockTx is a simple mock for testing
type mockTx struct{}

func (m *mockTx) Commit(ctx context.Context) error   { return nil }
func (m *mockTx) Rollback(ctx context.Context) error { return nil }
func (m *mockTx) Exec(ctx context.Context, query string, args ...any) (serviceapi.CommandResult, error) {
	return nil, nil
}
func (m *mockTx) Query(ctx context.Context, query string, args ...any) (serviceapi.Rows, error) {
	return nil, nil
}
func (m *mockTx) QueryRow(ctx context.Context, query string, args ...any) serviceapi.Row {
	return nil
}
func (m *mockTx) SelectOne(ctx context.Context, query string, args []any, dest ...any) error {
	return nil
}
func (m *mockTx) SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error {
	return nil
}
func (m *mockTx) SelectOneRowMap(ctx context.Context, query string, args ...any) (serviceapi.RowMap, error) {
	return nil, nil
}
func (m *mockTx) SelectManyRowMap(ctx context.Context, query string, args ...any) ([]serviceapi.RowMap, error) {
	return nil, nil
}
func (m *mockTx) SelectManyWithMapper(ctx context.Context, fnScan func(serviceapi.Row) (any, error), query string, args ...any) (any, error) {
	return nil, nil
}
func (m *mockTx) IsExists(ctx context.Context, query string, args ...any) (bool, error) {
	return false, nil
}
func (m *mockTx) IsErrorNoRows(err error) bool                     { return false }
func (m *mockTx) Begin(ctx context.Context) (serviceapi.Tx, error) { return nil, nil }
func (m *mockTx) Transaction(ctx context.Context, fn func(tx serviceapi.DbConn) error) error {
	return nil
}
func (m *mockTx) Release() error { return nil }
