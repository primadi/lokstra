// Transaction Usage Examples for lokstra/core/flow

package flow

import (
	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/serviceapi"
)

// Example 1: Basic transaction usage
func ExampleBasicTransaction(regCtx iface.RegistrationContext, pool serviceapi.DbPool) *Handler {
	return NewHandler(regCtx, "basic-transaction").
		BeginTx().
		ExecSql("INSERT INTO users (name, email) VALUES (?, ?)", "John Doe", "john@example.com").
		AffectOne().
		Done().
		ExecSql("INSERT INTO user_profiles (user_id, bio) VALUES (LAST_INSERT_ID(), ?)", "Software Developer").
		AffectOne().
		Done().
		CommitOrRollback()
}

// Example 2: Transaction with conditional rollback
func ExampleConditionalTransaction(regCtx iface.RegistrationContext, pool serviceapi.DbPool) *Handler {
	return NewHandler(regCtx, "conditional-transaction").
		BeginTx().
		QueryRowSql("SELECT balance FROM accounts WHERE id = ?", 123).
		ScanTo(func(row serviceapi.Row) error {
			var balance float64
			if err := row.Scan(&balance); err != nil {
				return err
			}
			// Business logic can be added here
			return nil
		}).
		ExecSql("UPDATE accounts SET balance = balance - ? WHERE id = ?", 100.0, 123).
		AffectOne().
		Done().
		ExecSql("INSERT INTO transactions (account_id, amount, type) VALUES (?, ?, ?)", 123, -100.0, "withdrawal").
		AffectOne().
		Done().
		CommitOrRollback()
}

// Example 3: Forced rollback for testing or error scenarios
func ExampleForcedRollback(regCtx iface.RegistrationContext, pool serviceapi.DbPool) *Handler {
	return NewHandler(regCtx, "forced-rollback").
		BeginTx().
		ExecSql("INSERT INTO temp_data (value) VALUES (?)", "test").
		Done().
		Rollback() // This will always rollback, useful for testing
}

// Example 4: Complex transaction with multiple operations
func ExampleComplexTransaction(regCtx iface.RegistrationContext, pool serviceapi.DbPool) *Handler {
	return NewHandler(regCtx, "complex-transaction").
		BeginTx().

		// Create order
		ExecSql("INSERT INTO orders (customer_id, total) VALUES (?, ?)", 1, 99.99).
		AffectOne().
		SaveAs("order_id").

		// Add order items
		ExecSql("INSERT INTO order_items (order_id, product_id, quantity, price) VALUES (?, ?, ?, ?)",
			"{{.order_id}}", 101, 2, 49.995).
		AffectOne().
		Done().

		// Update inventory
		ExecSql("UPDATE products SET stock = stock - ? WHERE id = ?", 2, 101).
		AffectOne().
		EnsureExists(nil). // Ensure product exists
		Done().

		// Log the transaction
		ExecSql("INSERT INTO audit_log (action, table_name, record_id) VALUES (?, ?, ?)",
			"order_created", "orders", "{{.order_id}}").
		Done().
		CommitOrRollback()
}

// Context shows how to use transaction steps in flow execution
func ExampleTransactionContext() {
	// This demonstrates the runtime behavior:
	// 1. BeginTx() creates a database transaction and stores it in ctx.dbTx
	// 2. All subsequent SQL operations use the transaction
	// 3. CommitOrRollback() either commits (on success) or rolls back (on error)
	// 4. Rollback() always rolls back regardless of success/failure
}

// ExampleNestedTransactionError demonstrates how nested transactions are prevented
func ExampleNestedTransactionError() {
	// INVALID: This will panic at build time
	// handler := NewHandler(regCtx, "invalid").
	//     BeginTx().
	//     BeginTx() // PANIC: multiple BeginTx() calls detected

	// INVALID: This will error at runtime if somehow a transaction already exists
	// If ctx.dbTx is already set when BeginTx step runs:
	// Error: "nested transactions not supported: transaction already active"

	// CORRECT: Use separate handlers for separate transactions
	// handler1 := NewHandler(regCtx, "tx1").BeginTx().ExecSql(...).CommitOrRollback()
	// handler2 := NewHandler(regCtx, "tx2").BeginTx().ExecSql(...).CommitOrRollback()
}
