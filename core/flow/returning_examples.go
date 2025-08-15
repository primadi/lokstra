package flow

import (
	"fmt"

	"github.com/primadi/lokstra/core/iface"
)

// Example 1: INSERT with RETURNING using ExecReturning
func ExampleInsertWithReturning(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "insert-with-returning").
		BeginTx().
		ExecReturning("INSERT INTO users (name, email) VALUES (?, ?) RETURNING id, created_at",
			"John Doe", "john@example.com").
		WithName("user.create").
		SaveAs("new_user"). // Saves: {"id": 123, "created_at": "2024-..."}

		// Use the returned data in next step
		ExecSql("INSERT INTO user_profiles (user_id, status) VALUES (?, 'active')", "{{.new_user.id}}").
		WithName("profile.create").
		Done().
		CommitOrRollback()
}

// Example 2: UPDATE with RETURNING using WithReturning
func ExampleUpdateWithReturning(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "update-with-returning").
		BeginTx().
		ExecSql("UPDATE users SET name = ?, updated_at = NOW() WHERE id = ? RETURNING id, name, updated_at",
					"Jane Smith", 123).
		WithReturning(). // Mark as RETURNING query
		WithName("user.update").
		AffectOne().            // Ensure exactly one row updated
		SaveAs("updated_user"). // Saves: {"id": 123, "name": "Jane Smith", "updated_at": "..."}

		// Log the update
		DoNamed("audit.log_update", func(ctx *Context) error {
			userData, _ := ctx.Get("updated_user")
			user := userData.(map[string]any)
			// Log: User 123 updated to name "Jane Smith" at timestamp
			_ = user // Use the user data for logging
			ctx.Set("audit_logged", true)
			return nil
		}).
		CommitOrRollback()
}

// Example 3: DELETE with RETURNING for audit trail
func ExampleDeleteWithReturning(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "delete-with-audit").
		BeginTx().

		// Soft delete with RETURNING to capture what was deleted
		ExecReturning(`UPDATE users 
			SET deleted_at = NOW(), updated_at = NOW() 
			WHERE id = ? AND deleted_at IS NULL 
			RETURNING id, name, email, deleted_at`, 456).
		WithName("user.soft_delete").
		AffectOne().
		SaveAs("deleted_user"). // Saves deleted user data for audit

		// Create audit log entry
		ExecSql(`INSERT INTO audit_logs (table_name, action, record_id, old_data, created_at) 
			VALUES ('users', 'DELETE', ?, ?, NOW())`,
			"{{.deleted_user.id}}", "{{.deleted_user}}").
		WithName("audit.log_deletion").
		Done().
		CommitOrRollback()
}

// Example 4: Complex business operation with multiple RETURNING
func ExampleComplexBusinessWithReturning(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "complex-business-returning").
		BeginTx().

		// Create order and get generated ID and timestamp
		ExecReturning(`INSERT INTO orders (customer_id, total_amount, status, created_at) 
			VALUES (?, ?, 'pending', NOW()) 
			RETURNING id, created_at, status`, 789, 99.99).
		WithName("order.create").
		SaveAs("new_order").

		// Create order items using the order ID
		ExecSql(`INSERT INTO order_items (order_id, product_id, quantity, price) 
			VALUES (?, ?, ?, ?)`,
			"{{.new_order.id}}", 101, 2, 49.99).
		WithName("order_items.create").
		Done().

		// Update inventory and return new stock levels
		ExecReturning(`UPDATE inventory 
			SET quantity = quantity - ?, updated_at = NOW() 
			WHERE product_id = ? 
			RETURNING product_id, quantity, updated_at`, 2, 101).
		WithName("inventory.decrement").
		AffectOne().
		SaveAs("inventory_update").

		// Check if inventory is low and needs reordering
		DoNamed("inventory.check_reorder", func(ctx *Context) error {
			inventoryData, _ := ctx.Get("inventory_update")
			inventory := inventoryData.(map[string]any)

			quantity := inventory["quantity"].(int64)
			if quantity < 10 {
				ctx.Set("needs_reorder", true)
				ctx.Set("reorder_product_id", inventory["product_id"])
			}
			return nil
		}).

		// Create reorder notification if needed
		ExecSql(`INSERT INTO reorder_notifications (product_id, current_stock, created_at)
			SELECT ?, ?, NOW() 
			WHERE ? = true`,
			"{{.reorder_product_id}}", "{{.inventory_update.quantity}}", "{{.needs_reorder}}").
		WithName("reorder.notify").
		Done().
		CommitOrRollback()
}

// Example 5: UPSERT with RETURNING (PostgreSQL specific)
func ExampleUpsertWithReturning(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "upsert-with-returning").
		ExecReturning(`INSERT INTO user_settings (user_id, setting_key, setting_value) 
			VALUES (?, ?, ?) 
			ON CONFLICT (user_id, setting_key) 
			DO UPDATE SET 
				setting_value = EXCLUDED.setting_value,
				updated_at = NOW()
			RETURNING id, user_id, setting_key, setting_value, 
				CASE WHEN xmax = 0 THEN 'inserted' ELSE 'updated' END as action`,
			123, "theme", "dark").
		WithName("user_settings.upsert").
		SaveAs("upsert_result"). // Saves: {"id": 456, "action": "updated", ...}

		DoNamed("settings.log_action", func(ctx *Context) error {
			result, _ := ctx.Get("upsert_result")
			resultMap := result.(map[string]any)
			action := resultMap["action"].(string)

			// Log whether this was an insert or update
			ctx.Set("action_type", action)
			return nil
		})
}

// Example 6: Batch operation with RETURNING aggregated data
func ExampleBatchWithReturning(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "batch-with-returning").
		BeginTx().

		// Update multiple records and get summary
		ExecReturning(`UPDATE products 
			SET price = price * 1.1, updated_at = NOW() 
			WHERE category_id = ? 
			RETURNING 
				COUNT(*) as updated_count,
				AVG(price) as avg_new_price,
				MIN(price) as min_price,
				MAX(price) as max_price`, 5).
		WithName("products.bulk_price_update").
		SaveAs("price_update_summary").

		// Log the bulk update summary
		DoNamed("pricing.log_bulk_update", func(ctx *Context) error {
			summary, _ := ctx.Get("price_update_summary")
			summaryMap := summary.(map[string]any)

			count := summaryMap["updated_count"].(int64)
			avgPrice := summaryMap["avg_new_price"].(float64)

			ctx.Set("update_message",
				fmt.Sprintf("Updated %d products, new avg price: %.2f", count, avgPrice))
			return nil
		}).

		// Create pricing audit entry
		ExecSql(`INSERT INTO pricing_audits (category_id, action, summary, created_at) 
			VALUES (?, 'BULK_INCREASE_10PCT', ?, NOW())`,
			5, "{{.update_message}}").
		WithName("audit.log_pricing").
		Done().
		CommitOrRollback()
}

/*
RETURNING Usage Patterns:

1. INSERT RETURNING:
   - Get auto-generated IDs
   - Capture timestamps
   - Get computed values

2. UPDATE RETURNING:
   - Capture old and new values
   - Get update timestamps
   - Verify what was actually changed

3. DELETE RETURNING:
   - Audit trail of deleted data
   - Soft delete confirmations
   - Cascade operation data

4. UPSERT RETURNING:
   - Know if record was inserted or updated
   - Get final state after conflict resolution

Best Practices:
- Always use SaveAs() with RETURNING queries
- Use meaningful variable names for returned data
- Consider guards (AffectOne, etc.) for data integrity
- Use WithName() for better debugging
- Chain operations using returned data with template syntax {{.variable.field}}
*/
