package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/loader"
	svc "github.com/primadi/lokstra/docs/00-introduction/examples/06-external-services/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Parse flags
	serverFlag := flag.String("server", "app.api-server", "Server to run (deployment.server format)")
	flag.Parse()

	// Register service factories
	lokstra_registry.RegisterServiceType("order-service-factory",
		svc.OrderServiceFactory, nil,
		deploy.WithResource("order", "orders"),
		deploy.WithConvention("rest"),
		deploy.WithRouteOverride("Refund", "POST /orders/{id}/refund"),
	)

	lokstra_registry.RegisterServiceType("payment-service-remote-factory",
		nil, svc.PaymentServiceRemoteFactory,
		deploy.WithResource("payment", "payments"),
		deploy.WithConvention("rest"),
		deploy.WithRouteOverride("CreatePayment", "POST /payments"),
		deploy.WithRouteOverride("GetPayment", "GET /payments/{id}"),
		deploy.WithRouteOverride("Refund", "POST /payments/{id}/refund"),
	)

	// Load config and build deployment topology
	if err := loader.LoadAndBuild([]string{"config.yaml"}); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Print info and run server
	printStartInfo()

	// Run server (compositeKey format: "deployment.server")
	if err := lokstra_registry.RunServer(*serverFlag, 30*time.Second); err != nil {
		log.Fatal(err)
	}
}

func printStartInfo() {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("🌐 Example 06 - External Services Integration")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("This example demonstrates:")
	fmt.Println("  ✅ External service integration (mock payment gateway)")
	fmt.Println("  ✅ proxy.Service for remote calls")
	fmt.Println("  ✅ Route override for non-standard endpoints")
	fmt.Println("  ✅ external-service-definitions in config")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("📋 Prerequisites:")
	fmt.Println("  1. Start mock payment gateway first:")
	fmt.Println("     cd mock-payment-gateway && go run main.go")
	fmt.Println("     (Runs on http://localhost:9000)")
	fmt.Println()
	fmt.Println("  2. Then start this server:")
	fmt.Println("     go run main.go")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("🔗 API Endpoints:")
	fmt.Println()
	fmt.Println("  Order Management:")
	fmt.Println("    POST   http://localhost:3000/orders        - Create order (processes payment)")
	fmt.Println("    GET    http://localhost:3000/orders/{id}   - Get order details")
	fmt.Println("    POST   http://localhost:3000/orders/{id}/refund - Refund order")
	fmt.Println()
	fmt.Println("💡 How it works:")
	fmt.Println("  1. CreateOrder calls external payment gateway via proxy.Service")
	fmt.Println("  2. Payment gateway returns payment ID")
	fmt.Println("  3. Order is marked as 'paid' with payment ID")
	fmt.Println("  4. Refund also goes through external gateway")
	fmt.Println()
	fmt.Println("📝 Test:")
	fmt.Println("  Use test.http file or:")
	fmt.Println()
	fmt.Println("  # Create order (processes payment)")
	fmt.Println(`  curl -X POST http://localhost:3000/orders \`)
	fmt.Println(`    -H "Content-Type: application/json" \`)
	fmt.Println(`    -d '{"user_id": 1, "items": ["Book", "Pen"], "total_amount": 25.50, "currency": "USD"}'`)
	fmt.Println()
	fmt.Println("  # Get order")
	fmt.Println(`  curl http://localhost:3000/orders/order_1`)
	fmt.Println()
	fmt.Println("  # Refund order")
	fmt.Println(`  curl -X POST http://localhost:3000/orders/order_1/refund`)
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
}
