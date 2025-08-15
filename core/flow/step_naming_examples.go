package flow

import (
	"errors"

	"github.com/primadi/lokstra/core/iface"
)

// Example 1: Default step names
func ExampleDefaultStepNaming(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "default-naming").
		BeginTx().                                              // Step name: "tx.begin"
		ExecSql("INSERT INTO users (name) VALUES (?)", "John"). // Step name: "sql.exec"
		Done().
		QueryRowSql("SELECT id FROM users WHERE name = ?", "John"). // Step name: "sql.query_row"
		SaveAs("user_id").
		CommitOrRollback() // Step name: "tx.end"
}

// Example 2: Custom meaningful step names
func ExampleMeaningfulStepNaming(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "meaningful-naming").
		BeginTx().
		ExecSql("INSERT INTO users (name) VALUES (?)", "John").
		WithName("user.create"). // Custom name: "user.create"
		AffectOne().
		Done().
		QueryRowSql("SELECT id FROM users WHERE name = ?", "John").
		WithName("user.find_by_name"). // Custom name: "user.find_by_name"
		SaveAs("user_id").
		DoNamed("user.send_welcome_email", func(ctx *Context) error {
			// Custom step name: "user.send_welcome_email"
			userID, _ := ctx.Get("user_id")
			// Send welcome email logic here
			_ = userID // Use the userID for email sending
			ctx.Set("email_sent", true)
			return nil
		}).
		CommitOrRollback()
}

// Example 3: Business domain naming
func ExampleBusinessDomainNaming(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "business-domain-naming").
		BeginTx().

		// Order validation
		QueryRowSql("SELECT status, amount FROM orders WHERE id = ?", 123).
		WithName("order.validate").
		SaveAs("order_data").
		DoNamed("order.check_status", func(ctx *Context) error {
			orderData, _ := ctx.Get("order_data")
			order := orderData.(map[string]any)
			if order["status"] != "pending" {
				return errors.New("order is not pending")
			}
			return nil
		}).

		// Payment processing
		DoNamed("payment.process", func(ctx *Context) error {
			// External payment service call
			ctx.Set("payment_id", "pay_12345")
			ctx.Set("payment_status", "completed")
			return nil
		}).

		// Order fulfillment
		ExecSql("UPDATE orders SET status = ?, payment_id = ? WHERE id = ?",
			"completed", "{{.payment_id}}", 123).
		WithName("order.fulfill").
		AffectOne().
		Done().

		// Inventory update
		ExecSql("UPDATE inventory SET quantity = quantity - 1 WHERE product_id = ?", 456).
		WithName("inventory.decrement").
		AffectAtLeast(1).
		Done().
		CommitOrRollback()
}

// Example 4: Debugging and monitoring
func ExampleDebugFriendlyNaming(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "debug-friendly-naming").

		// Data preparation
		QuerySql("SELECT id, balance FROM accounts WHERE status = 'active'").
		WithName("account.load_active").
		SaveAs("active_accounts").
		DoNamed("account.calculate_total_balance", func(ctx *Context) error {
			accounts, _ := ctx.Get("active_accounts")
			accountList := accounts.([]map[string]any)

			var total float64
			for _, account := range accountList {
				total += account["balance"].(float64)
			}

			ctx.Set("total_balance", total)
			return nil
		}).

		// Report generation
		ExecSql("INSERT INTO reports (type, total_amount, generated_at) VALUES (?, ?, NOW())",
			"balance_summary", "{{.total_balance}}").
		WithName("report.create_balance_summary").
		SaveAs("report_creation").
		DoNamed("report.send_notification", func(ctx *Context) error {
			// Send notification to admin
			ctx.Set("notification_sent", true)
			return nil
		})
}

// Example 5: Error handling with named steps
func ExampleErrorHandlingWithNaming(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "error-handling-naming").
		BeginTx().

		// Critical business logic with clear step names
		QueryRowSql("SELECT balance FROM accounts WHERE id = ?", 123).
		WithName("account.check_balance").
		SaveAs("account_balance").
		DoNamed("account.validate_withdrawal", func(ctx *Context) error {
			balance, _ := ctx.Get("account_balance")
			balanceMap := balance.(map[string]any)

			if balanceMap["balance"].(float64) < 100.0 {
				// Error will show: "step account.validate_withdrawal failed: insufficient funds"
				return errors.New("insufficient funds")
			}
			return nil
		}).
		ExecSql("UPDATE accounts SET balance = balance - ? WHERE id = ?", 50.0, 123).
		WithName("account.debit").
		AffectOne().
		Done().
		ExecSql("INSERT INTO transactions (account_id, amount, type) VALUES (?, ?, ?)",
			123, 50.0, "debit").
		WithName("transaction.record_debit").
		SaveAs("transaction_record").
		CommitOrRollback()
}

// Example 6: Step naming conventions
func ExampleStepNamingConventions(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "naming-conventions").

		// Convention: domain.action
		ExecSql("INSERT INTO products (name, price) VALUES (?, ?)", "Widget", 29.99).
		WithName("product.create").
		Done().

		// Convention: domain.action_target
		QueryRowSql("SELECT * FROM products WHERE name = ?", "Widget").
		WithName("product.find_by_name").
		SaveAs("product_data").

		// Convention: domain.business_operation
		DoNamed("inventory.reserve_stock", func(ctx *Context) error {
			// Reserve stock logic
			return nil
		}).

		// Convention: external_service.operation
		DoNamed("payment_gateway.charge_card", func(ctx *Context) error {
			// Payment gateway integration
			return nil
		}).

		// Convention: notification.channel_action
		DoNamed("notification.email_send", func(ctx *Context) error {
			// Email notification
			return nil
		})
}

/*
Naming Convention Guidelines:

1. Domain-based naming:
   - user.create, user.update, user.find_by_email
   - order.validate, order.fulfill, order.cancel
   - payment.process, payment.verify, payment.refund

2. Action-based naming:
   - database.backup, database.cleanup
   - cache.invalidate, cache.warm_up
   - file.upload, file.process, file.cleanup

3. Integration naming:
   - stripe.charge_card, stripe.create_customer
   - sendgrid.send_email, sendgrid.send_bulk
   - slack.post_message, slack.create_channel

4. Business process naming:
   - checkout.validate_cart, checkout.apply_discount
   - shipping.calculate_rate, shipping.create_label
   - inventory.check_availability, inventory.reserve

5. System operation naming:
   - audit.log_action, audit.create_trail
   - security.validate_token, security.check_permissions
   - monitoring.record_metric, monitoring.alert
*/
