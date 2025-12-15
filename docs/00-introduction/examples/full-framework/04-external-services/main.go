package main

import (
	"strings"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/deploy"
	svc "github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/04-external-services/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Register service factories
	lokstra_registry.RegisterRouterServiceType("order-service-factory",
		svc.OrderServiceFactory, nil,
		&deploy.ServiceTypeConfig{
			RouteOverrides: map[string]deploy.RouteConfig{
				"Refund": {Method: "POST", Path: "/orders/{id}/refund"},
			},
		},
	)

	lokstra_registry.RegisterRouterServiceType("payment-service-remote-factory",
		nil, svc.PaymentServiceRemoteFactory,
		&deploy.ServiceTypeConfig{
			RouteOverrides: map[string]deploy.RouteConfig{
				"CreatePayment": {Method: "POST", Path: "/payments"},
				"GetPayment":    {Method: "GET", Path: "/payments/{id}"},
				"Refund":        {Method: "POST", Path: "/payments/{id}/refund"},
			},
		},
	)

	printStartInfo()

	if err := lokstra_registry.LoadConfig(); err != nil {
		logger.LogPanic("❌ Failed to load config:", err)
	}

	if err := lokstra_registry.RunConfiguredServer(); err != nil {
		logger.LogPanic("❌ Failed to run server:", err)
	}

}

func printStartInfo() {
	logger.LogInfo("")
	logger.LogInfo(strings.Repeat("=", 70))
	logger.LogInfo("🌐 Example 06 - External Services Integration")
	logger.LogInfo(strings.Repeat("=", 70))
	logger.LogInfo("")
	logger.LogInfo("This example demonstrates:")
	logger.LogInfo("  ✅ External service integration (mock payment gateway)")
	logger.LogInfo("  ✅ proxy.Service for remote calls")
	logger.LogInfo("  ✅ Route override for non-standard endpoints")
	logger.LogInfo("  ✅ external-service-definitions in config")
	logger.LogInfo("")
	logger.LogInfo(strings.Repeat("=", 70))
	logger.LogInfo("")
	logger.LogInfo("📋 Prerequisites:")
	logger.LogInfo("  1. Start mock payment gateway first:")
	logger.LogInfo("     cd mock-payment-gateway && go run main.go")
	logger.LogInfo("     (Runs on http://localhost:9000)")
	logger.LogInfo("")
	logger.LogInfo("  2. Then start this server:")
	logger.LogInfo("     go run main.go")
	logger.LogInfo("")
	logger.LogInfo(strings.Repeat("=", 70))
	logger.LogInfo("")
	logger.LogInfo("🔗 API Endpoints:")
	logger.LogInfo("")
	logger.LogInfo("  Order Management:")
	logger.LogInfo("    POST   http://localhost:3000/orders        - Create order (processes payment)")
	logger.LogInfo("    GET    http://localhost:3000/orders/{id}   - Get order details")
	logger.LogInfo("    POST   http://localhost:3000/orders/{id}/refund - Refund order")
	logger.LogInfo("")
	logger.LogInfo("💡 How it works:")
	logger.LogInfo("  1. CreateOrder calls external payment gateway via proxy.Service")
	logger.LogInfo("  2. Payment gateway returns payment ID")
	logger.LogInfo("  3. Order is marked as 'paid' with payment ID")
	logger.LogInfo("  4. Refund also goes through external gateway")
	logger.LogInfo("")
	logger.LogInfo("📝 Test:")
	logger.LogInfo("  Use test.http file or:")
	logger.LogInfo("")
	logger.LogInfo("  # Create order (processes payment)")
	logger.LogInfo(`  curl -X POST http://localhost:3000/orders \`)
	logger.LogInfo(`    -H "Content-Type: application/json" \`)
	logger.LogInfo(`    -d '{"user_id": 1, "items": ["Book", "Pen"], "total_amount": 25.50, "currency": "USD"}'`)
	logger.LogInfo("")
	logger.LogInfo("  # Get order")
	logger.LogInfo(`  curl http://localhost:3000/orders/order_1`)
	logger.LogInfo("")
	logger.LogInfo("  # Refund order")
	logger.LogInfo(`  curl -X POST http://localhost:3000/orders/order_1/refund`)
	logger.LogInfo("")
	logger.LogInfo(strings.Repeat("=", 70))
	logger.LogInfo("")
}
