package migration_runner_test

import (
	"context"
	"testing"

	"github.com/primadi/lokstra/serviceapi"
)

// MockDbPoolWithSchema implements serviceapi.DbPoolWithSchema for testing
type MockDbPoolWithSchema struct {
	executions []string // Track executed SQL
}

var _ serviceapi.DbPool = (*MockDbPoolWithSchema)(nil)

func (m *MockDbPoolWithSchema) Acquire(ctx context.Context) (serviceapi.DbConn, error) {
	return &MockDbConn{pool: m}, nil
}

func (m *MockDbPoolWithSchema) Shutdown() error {
	return nil
}

type MockDbConn struct {
	pool *MockDbPoolWithSchema
}

var _ serviceapi.DbConn = (*MockDbConn)(nil)

// Begin implements serviceapi.DbConn.
func (m *MockDbConn) Begin(ctx context.Context) (serviceapi.DbTx, error) {
	panic("unimplemented")
}

// Exec implements serviceapi.DbConn.
func (m *MockDbConn) Exec(ctx context.Context, query string, args ...any) (serviceapi.CommandResult, error) {
	m.pool.executions = append(m.pool.executions, query)
	return &MockCommandResult{rowsAffected: 1}, nil
}

// IsErrorNoRows implements serviceapi.DbConn.
func (m *MockDbConn) IsErrorNoRows(err error) bool {
	panic("unimplemented")
}

// IsExists implements serviceapi.DbConn.
func (m *MockDbConn) IsExists(ctx context.Context, query string, args ...any) (bool, error) {
	panic("unimplemented")
}

// Ping implements serviceapi.DbConn.
func (m *MockDbConn) Ping(context context.Context) error {
	panic("unimplemented")
}

// Query implements serviceapi.DbConn.
func (m *MockDbConn) Query(ctx context.Context, query string, args ...any) (serviceapi.Rows, error) {
	return MockRows{}, nil
}

// QueryRow implements serviceapi.DbConn.
func (m *MockDbConn) QueryRow(ctx context.Context, query string, args ...any) serviceapi.Row {
	return MockRow{}
}

// Release implements serviceapi.DbConn.
func (m *MockDbConn) Release() error {
	return nil
}

// SelectManyRowMap implements serviceapi.DbConn.
func (m *MockDbConn) SelectManyRowMap(ctx context.Context, query string, args ...any) ([]serviceapi.RowMap, error) {
	panic("unimplemented")
}

// SelectManyWithMapper implements serviceapi.DbConn.
func (m *MockDbConn) SelectManyWithMapper(ctx context.Context, fnScan func(serviceapi.Row) (any, error), query string, args ...any) (any, error) {
	panic("unimplemented")
}

// SelectMustOne implements serviceapi.DbConn.
func (m *MockDbConn) SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error {
	panic("unimplemented")
}

// SelectOne implements serviceapi.DbConn.
func (m *MockDbConn) SelectOne(ctx context.Context, query string, args []any, dest ...any) error {
	panic("unimplemented")
}

// SelectOneRowMap implements serviceapi.DbConn.
func (m *MockDbConn) SelectOneRowMap(ctx context.Context, query string, args ...any) (serviceapi.RowMap, error) {
	panic("unimplemented")
}

// Shutdown implements serviceapi.DbConn.
func (m *MockDbConn) Shutdown() error {
	panic("unimplemented")
}

// Transaction implements serviceapi.DbConn.
func (m *MockDbConn) Transaction(ctx context.Context, fn func(tx serviceapi.DbExecutor) error) error {
	return fn(m)
}

var _ serviceapi.DbConn = (*MockDbConn)(nil)

type MockCommandResult struct {
	rowsAffected int64
}

var _ serviceapi.CommandResult = (*MockCommandResult)(nil)

// RowsAffected implements serviceapi.CommandResult.
func (m *MockCommandResult) RowsAffected() int64 {
	return m.rowsAffected
}

type MockRows struct{}

func (r MockRows) Next() bool        { return false }
func (r MockRows) Scan(...any) error { return nil }
func (r MockRows) Close() error      { return nil }
func (r MockRows) Err() error        { return nil }

type MockRow struct{}

func (r MockRow) Scan(...any) error { return nil }

func TestMigrationFileNaming(t *testing.T) {
	tests := []struct {
		filename string
		valid    bool
	}{
		{"001_create_users.up.sql", true},
		{"001_create_users.down.sql", true},
		{"999_complex_name_with_numbers_123.up.sql", true},
		{"invalid.sql", false},
		{"001_test.sql", false},
		{"abc_test.up.sql", false},
	}

	// This would test the regex pattern
	// Implementation depends on exposing the pattern or having a validation method
	_ = tests
}
