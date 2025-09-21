package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/primadi/lokstra"
)

// This example demonstrates comprehensive error handling patterns in Lokstra.
// It shows different error types, validation errors, error middleware, and recovery patterns.
//
// Learning Objectives:
// - Understand Lokstra's structured error responses
// - Learn validation error handling
// - Explore error middleware and recovery
// - See custom error types and handling
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/core-concepts.md#error-handling

// ===== Custom Error Types =====

type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e BusinessError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Custom errors for different business scenarios
var (
	ErrInsufficientFunds = BusinessError{
		Code:    "INSUFFICIENT_FUNDS",
		Message: "Insufficient funds for this transaction",
	}
	ErrAccountLocked = BusinessError{
		Code:    "ACCOUNT_LOCKED",
		Message: "Account is temporarily locked",
	}
	ErrInvalidOperation = BusinessError{
		Code:    "INVALID_OPERATION",
		Message: "Operation not allowed in current state",
	}
)

// ===== Service Layer with Error Handling =====

type BankAccount struct {
	ID      int     `json:"id"`
	Balance float64 `json:"balance"`
	Status  string  `json:"status"`
}

type BankService struct {
	accounts map[int]*BankAccount
}

func NewBankService() *BankService {
	accounts := make(map[int]*BankAccount)
	accounts[1] = &BankAccount{ID: 1, Balance: 1000.0, Status: "active"}
	accounts[2] = &BankAccount{ID: 2, Balance: 50.0, Status: "active"}
	accounts[3] = &BankAccount{ID: 3, Balance: 2000.0, Status: "locked"}

	return &BankService{accounts: accounts}
}

func (bs *BankService) GetAccount(id int) (*BankAccount, error) {
	account, exists := bs.accounts[id]
	if !exists {
		return nil, fmt.Errorf("account not found: %d", id)
	}
	return account, nil
}

func (bs *BankService) Withdraw(accountID int, amount float64) (*BankAccount, error) {
	account, err := bs.GetAccount(accountID)
	if err != nil {
		return nil, err
	}

	if account.Status == "locked" {
		return nil, ErrAccountLocked
	}

	if account.Balance < amount {
		err := ErrInsufficientFunds
		err.Details = fmt.Sprintf("Available: %.2f, Requested: %.2f", account.Balance, amount)
		return nil, err
	}

	if amount <= 0 {
		return nil, ErrInvalidOperation
	}

	account.Balance -= amount
	return account, nil
}

func (bs *BankService) Deposit(accountID int, amount float64) (*BankAccount, error) {
	account, err := bs.GetAccount(accountID)
	if err != nil {
		return nil, err
	}

	if account.Status == "locked" {
		return nil, ErrAccountLocked
	}

	if amount <= 0 {
		return nil, ErrInvalidOperation
	}

	account.Balance += amount
	return account, nil
}

// ===== Request/Response Types with Validation =====

type WithdrawRequest struct {
	AccountID int     `json:"account_id" validate:"required,min=1"`
	Amount    float64 `json:"amount" validate:"required,min=0.01,max=10000"`
	PIN       string  `json:"pin" validate:"required,len=4"`
}

type DepositRequest struct {
	AccountID int     `json:"account_id" validate:"required,min=1"`
	Amount    float64 `json:"amount" validate:"required,min=0.01,max=50000"`
}

type TransferRequest struct {
	FromAccount int     `json:"from_account" validate:"required,min=1"`
	ToAccount   int     `json:"to_account" validate:"required,min=1"`
	Amount      float64 `json:"amount" validate:"required,min=0.01,max=10000"`
	PIN         string  `json:"pin" validate:"required,len=4"`
}

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()

	// Create bank service
	bankService := NewBankService()

	app := lokstra.NewApp(regCtx, "error-handling-app", ":8080")

	// ===== Error Handling Middleware =====

	// Request logging middleware with error capture
	app.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		start := time.Now()
		method := ctx.Request.Method
		path := ctx.Request.URL.Path

		lokstra.Logger.Infof("🔄 %s %s - Starting", method, path)

		err := next(ctx)

		duration := time.Since(start)
		status := ctx.StatusCode

		if err != nil {
			lokstra.Logger.Errorf("❌ %s %s - Error: %v (took %v)", method, path, err, duration)
		} else if status >= 400 {
			lokstra.Logger.Warnf("⚠️  %s %s - Status: %d (took %v)", method, path, status, duration)
		} else {
			lokstra.Logger.Infof("✅ %s %s - Status: %d (took %v)", method, path, status, duration)
		}

		return err
	})

	// Panic recovery middleware
	app.Use(func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
		defer func() {
			if r := recover(); r != nil {
				lokstra.Logger.Errorf("🚨 PANIC recovered: %v", r)
				ctx.ErrorInternal("An unexpected error occurred")
			}
		}()

		return next(ctx)
	})

	// ===== Error Handling Endpoints =====

	// 1. Basic error responses
	app.GET("/errors/demo", func(ctx *lokstra.Context) error {
		errorType := ctx.GetQueryParam("type")

		switch errorType {
		case "notfound":
			return ctx.ErrorNotFound("The requested resource was not found")

		case "badrequest":
			return ctx.ErrorBadRequest("The request parameters are invalid")

		case "duplicate":
			return ctx.ErrorDuplicate("This resource already exists")

		case "internal":
			return ctx.ErrorInternal("An internal server error occurred")

		case "validation":
			return ctx.ErrorValidation("Validation failed", map[string]string{
				"email":    "Must be a valid email address",
				"password": "Must be at least 8 characters",
				"age":      "Must be between 18 and 120",
			})

		default:
			return ctx.Ok(map[string]any{
				"message": "Error demo endpoint",
				"available_types": []string{
					"notfound", "badrequest", "duplicate", "internal", "validation",
				},
				"usage": "Add ?type=<error_type> to test different errors",
			})
		}
	})

	// 2. Account lookup with not found handling
	type AccountParams struct {
		ID int `path:"id"`
	}

	app.GET("/accounts/:id", func(ctx *lokstra.Context, params *AccountParams) error {
		account, err := bankService.GetAccount(params.ID)
		if err != nil {
			return ctx.ErrorNotFound(fmt.Sprintf("Account %d not found", params.ID))
		}

		return ctx.Ok(account)
	})

	// 3. Withdrawal with business error handling
	app.POST("/accounts/withdraw", func(ctx *lokstra.Context, req *WithdrawRequest) error {
		// Simulate PIN validation
		if req.PIN != "1234" {
			return ctx.ErrorBadRequest("Invalid PIN")
		}

		account, err := bankService.Withdraw(req.AccountID, req.Amount)
		if err != nil {
			// Handle different error types
			var businessErr BusinessError
			if errors.As(err, &businessErr) {
				switch businessErr.Code {
				case "INSUFFICIENT_FUNDS":
					return ctx.ErrorBadRequest(businessErr.Message + ". " + businessErr.Details)
				case "ACCOUNT_LOCKED":
					return ctx.ErrorBadRequest(businessErr.Message)
				case "INVALID_OPERATION":
					return ctx.ErrorBadRequest(businessErr.Message)
				default:
					return ctx.ErrorInternal("Unknown business error")
				}
			}

			// Handle generic errors
			if err.Error() == fmt.Sprintf("account not found: %d", req.AccountID) {
				return ctx.ErrorNotFound(err.Error())
			}

			return ctx.ErrorInternal("Withdrawal failed")
		}

		return ctx.Ok(map[string]any{
			"message": "Withdrawal successful",
			"account": account,
		})
	})

	// 4. Deposit with validation error handling
	app.POST("/accounts/deposit", func(ctx *lokstra.Context, req *DepositRequest) error {
		account, err := bankService.Deposit(req.AccountID, req.Amount)
		if err != nil {
			var businessErr BusinessError
			if errors.As(err, &businessErr) {
				return ctx.ErrorBadRequest(businessErr.Message)
			}

			if err.Error() == fmt.Sprintf("account not found: %d", req.AccountID) {
				return ctx.ErrorNotFound(err.Error())
			}

			return ctx.ErrorInternal("Deposit failed")
		}

		return ctx.OkCreated(map[string]any{
			"message": "Deposit successful",
			"account": account,
		})
	})

	// 5. Transfer with multiple error conditions
	app.POST("/accounts/transfer", func(ctx *lokstra.Context, req *TransferRequest) error {
		// Validate PIN
		if req.PIN != "1234" {
			return ctx.ErrorBadRequest("Invalid PIN")
		}

		// Validate accounts are different
		if req.FromAccount == req.ToAccount {
			return ctx.ErrorBadRequest("Cannot transfer to the same account")
		}

		// Check source account
		fromAccount, err := bankService.GetAccount(req.FromAccount)
		if err != nil {
			return ctx.ErrorNotFound(fmt.Sprintf("Source account %d not found", req.FromAccount))
		}

		// Check destination account
		toAccount, err := bankService.GetAccount(req.ToAccount)
		if err != nil {
			return ctx.ErrorNotFound(fmt.Sprintf("Destination account %d not found", req.ToAccount))
		}

		// Perform withdrawal
		_, err = bankService.Withdraw(req.FromAccount, req.Amount)
		if err != nil {
			var businessErr BusinessError
			if errors.As(err, &businessErr) {
				return ctx.ErrorBadRequest("Transfer failed: " + businessErr.Message)
			}
			return ctx.ErrorInternal("Transfer failed during withdrawal")
		}

		// Perform deposit
		_, err = bankService.Deposit(req.ToAccount, req.Amount)
		if err != nil {
			// Rollback withdrawal (in real app, use transactions)
			bankService.Deposit(req.FromAccount, req.Amount)

			var businessErr BusinessError
			if errors.As(err, &businessErr) {
				return ctx.ErrorBadRequest("Transfer failed: " + businessErr.Message)
			}
			return ctx.ErrorInternal("Transfer failed during deposit")
		}

		return ctx.Ok(map[string]any{
			"message":      "Transfer successful",
			"from_account": fromAccount,
			"to_account":   toAccount,
			"amount":       req.Amount,
		})
	})

	// 6. Endpoint that triggers panic for recovery testing
	app.GET("/panic", func(ctx *lokstra.Context) error {
		panic("This is a test panic!")
	})

	// 7. Validation error demonstration
	type ComplexValidationRequest struct {
		Email       string  `json:"email" validate:"required,email"`
		Password    string  `json:"password" validate:"required,min=8"`
		Age         int     `json:"age" validate:"required,min=18,max=120"`
		Amount      float64 `json:"amount" validate:"required,min=0.01"`
		PhoneNumber string  `json:"phone" validate:"required,e164"`
	}

	app.POST("/validate-complex", func(ctx *lokstra.Context, req *ComplexValidationRequest) error {
		// In a real Lokstra app, validation would happen automatically
		// This example shows manual validation for demonstration

		fieldErrors := make(map[string]string)

		if req.Email == "" {
			fieldErrors["email"] = "Email is required"
		}
		if req.Password == "" || len(req.Password) < 8 {
			fieldErrors["password"] = "Password must be at least 8 characters"
		}
		if req.Age < 18 || req.Age > 120 {
			fieldErrors["age"] = "Age must be between 18 and 120"
		}
		if req.Amount <= 0 {
			fieldErrors["amount"] = "Amount must be greater than 0"
		}

		if len(fieldErrors) > 0 {
			return ctx.ErrorValidation("Validation failed", fieldErrors)
		}

		return ctx.Ok(map[string]any{
			"message": "Validation passed",
			"data":    req,
		})
	})

	// 8. Error aggregation endpoint
	app.GET("/errors/health", func(ctx *lokstra.Context) error {
		errors := []string{}

		// Check various system components
		if bankService == nil {
			errors = append(errors, "Bank service unavailable")
		}

		// Simulate other health checks
		if time.Now().Second()%10 == 0 {
			errors = append(errors, "Database connection unstable")
		}

		if len(errors) > 0 {
			return ctx.ErrorInternal(fmt.Sprintf("Health check failed: %v", errors))
		}

		return ctx.Ok(map[string]any{
			"status": "healthy",
			"checks": "All systems operational",
		})
	})

	lokstra.Logger.Infof("🚀 Error Handling Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Error Handling Examples:")
	lokstra.Logger.Infof("  Basic Errors:")
	lokstra.Logger.Infof("    GET  /errors/demo?type=notfound    - 404 Not Found")
	lokstra.Logger.Infof("    GET  /errors/demo?type=validation  - 400 Validation Error")
	lokstra.Logger.Infof("    GET  /errors/demo?type=internal    - 500 Internal Error")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  Business Logic Errors:")
	lokstra.Logger.Infof("    GET  /accounts/999                 - Account not found")
	lokstra.Logger.Infof("    POST /accounts/withdraw            - Business rule violations")
	lokstra.Logger.Infof("    POST /accounts/transfer            - Complex error handling")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("  Advanced Error Handling:")
	lokstra.Logger.Infof("    GET  /panic                        - Panic recovery")
	lokstra.Logger.Infof("    POST /validate-complex             - Field validation")
	lokstra.Logger.Infof("    GET  /errors/health                - Health check errors")

	app.Start(true)
}

// Error Handling Best Practices:
//
// 1. Structured Error Responses:
//    - Use Lokstra's built-in error response methods
//    - Provide consistent error format across all endpoints
//    - Include appropriate HTTP status codes
//    - Add meaningful error messages for clients
//
// 2. Validation Errors:
//    - Use ErrorValidation for field-level validation errors
//    - Provide field-specific error messages
//    - Validate at handler entry point
//    - Use struct tags for automatic validation
//
// 3. Business Logic Errors:
//    - Create custom error types for business rules
//    - Map business errors to appropriate HTTP status codes
//    - Include relevant context in error messages
//    - Log business errors for monitoring
//
// 4. Error Recovery:
//    - Use middleware for panic recovery
//    - Log panics with full context
//    - Return user-friendly error messages
//    - Monitor panic frequency for stability
//
// 5. Error Monitoring:
//    - Log all errors with appropriate levels
//    - Include request context in error logs
//    - Use structured logging for analysis
//    - Implement health checks for error detection

// Test Commands:
//
// # Basic error types
// curl "http://localhost:8080/errors/demo?type=notfound"
// curl "http://localhost:8080/errors/demo?type=validation"
// curl "http://localhost:8080/errors/demo?type=internal"
//
// # Account operations
// curl http://localhost:8080/accounts/1
// curl http://localhost:8080/accounts/999
//
// # Withdrawal (insufficient funds)
// curl -X POST http://localhost:8080/accounts/withdraw \
//      -H "Content-Type: application/json" \
//      -d '{"account_id":2,"amount":100,"pin":"1234"}'
//
// # Withdrawal (locked account)
// curl -X POST http://localhost:8080/accounts/withdraw \
//      -H "Content-Type: application/json" \
//      -d '{"account_id":3,"amount":50,"pin":"1234"}'
//
// # Valid deposit
// curl -X POST http://localhost:8080/accounts/deposit \
//      -H "Content-Type: application/json" \
//      -d '{"account_id":1,"amount":100}'
//
// # Transfer
// curl -X POST http://localhost:8080/accounts/transfer \
//      -H "Content-Type: application/json" \
//      -d '{"from_account":1,"to_account":2,"amount":50,"pin":"1234"}'
//
// # Panic recovery
// curl http://localhost:8080/panic
//
// # Validation errors
// curl -X POST http://localhost:8080/validate-complex \
//      -H "Content-Type: application/json" \
//      -d '{"email":"invalid","password":"short","age":15,"amount":-10}'
