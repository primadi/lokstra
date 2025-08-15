package flow

import (
	"errors"
	"fmt"
	"log"

	"github.com/primadi/lokstra/core/iface"
)

// Example 1: Simple custom logic
func ExampleSimpleCustomLogic(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "simple-custom").
		ExecSql("INSERT INTO logs (message) VALUES (?)", "Starting process").
		Done().
		Do(func(ctx *Context) error {
			log.Println("Custom logic: Process started")
			ctx.Set("process_started", true)
			return nil
		}).
		ExecSql("UPDATE logs SET status = 'completed'").
		Done()
}

// Example 2: Conditional business logic
func ExampleConditionalLogic(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "conditional-business").
		QueryRowSql("SELECT balance FROM accounts WHERE id = ?", 123).
		SaveAs("account_balance").
		Do(func(ctx *Context) error {
			// Get the balance from previous query
			balanceData, exists := ctx.Get("account_balance")
			if !exists {
				return errors.New("account balance not found")
			}

			balanceMap := balanceData.(map[string]any)
			balance := balanceMap["balance"].(float64)

			// Business logic: Check minimum balance
			if balance < 100.0 {
				ctx.Set("low_balance_warning", true)
				ctx.Set("notification_message", "Warning: Low account balance")
			}

			return nil
		}).
		ExecSql("INSERT INTO notifications (message) VALUES (?)", "{{.notification_message}}").
		Done()
}

// Example 3: Complex data transformation
func ExampleDataTransformation(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "data-transformation").
		QuerySql("SELECT id, name, email FROM users WHERE status = 'active'").
		SaveAs("active_users").
		Do(func(ctx *Context) error {
			// Get the users data
			usersData, exists := ctx.Get("active_users")
			if !exists {
				return nil // No users to process
			}

			users := usersData.([]map[string]any)

			// Transform data: create email list
			var emailList []string
			for _, user := range users {
				if email, ok := user["email"].(string); ok && email != "" {
					emailList = append(emailList, email)
				}
			}

			// Save transformed data
			ctx.Set("email_list", emailList)
			ctx.Set("email_count", len(emailList))

			return nil
		}).
		ExecSql("INSERT INTO email_campaigns (recipient_count) VALUES (?)", "{{.email_count}}").
		Done()
}

// Example 4: Error handling with validation
func ExampleValidationWithDo(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "validation-example").
		BeginTx().
		QueryRowSql("SELECT id, status FROM orders WHERE id = ?", 456).
		SaveAs("order_data").
		Do(func(ctx *Context) error {
			// Validate order status
			orderData, exists := ctx.Get("order_data")
			if !exists {
				return errors.New("order not found")
			}

			order := orderData.(map[string]any)
			status := order["status"].(string)

			// Business rule: can only process pending orders
			if status != "pending" {
				return fmt.Errorf("cannot process order with status: %s", status)
			}

			// Set processing flag
			ctx.Set("can_process", true)
			return nil
		}).
		ExecSql("UPDATE orders SET status = 'processing' WHERE id = ?", 456).
		AffectOne().
		Done().
		CommitOrRollback()
}

// Example 5: Integration with external services
func ExampleExternalServiceIntegration(regCtx iface.RegistrationContext) *Handler {
	return NewHandler(regCtx, "external-integration").
		QueryRowSql("SELECT id, email, amount FROM invoices WHERE status = 'pending'").
		SaveAs("pending_invoice").
		Do(func(ctx *Context) error {
			// Get invoice data
			invoiceData, exists := ctx.Get("pending_invoice")
			if !exists {
				return nil // No pending invoices
			}

			invoice := invoiceData.(map[string]any)
			email := invoice["email"].(string)
			amount := invoice["amount"].(float64)

			// Simulate external service call (e.g., payment processor)
			paymentResult, err := processPayment(email, amount)
			if err != nil {
				return fmt.Errorf("payment processing failed: %w", err)
			}

			// Save payment result
			ctx.Set("payment_id", paymentResult.ID)
			ctx.Set("payment_status", paymentResult.Status)

			return nil
		}).
		ExecSql("UPDATE invoices SET payment_id = ?, status = ? WHERE id = ?",
			"{{.payment_id}}", "{{.payment_status}}", "{{.pending_invoice.id}}").
		Done()
}

// Mock external service function
type PaymentResult struct {
	ID     string
	Status string
}

func processPayment(email string, amount float64) (*PaymentResult, error) {
	_ = email
	_ = amount
	// Simulate external API call
	return &PaymentResult{
		ID:     "pay_12345",
		Status: "completed",
	}, nil
}
