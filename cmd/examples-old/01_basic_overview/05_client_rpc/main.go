package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/primadi/lokstra/cmd/examples-old/01_basic_overview/05_client_rpc/hello_service"
)

// ============================
// MAIN APPLICATION
// ============================

func main() {
	fmt.Println("🚀 RPC Client Example - Testing Various Return Types")
	fmt.Println("🔗 Connecting to server at http://localhost:8080/rpc")
	fmt.Println()

	// Create client
	client := hello_service.NewGreetingServiceClient("http://localhost:8080/rpc")

	// Test server connectivity first
	fmt.Println("🔍 Testing server connectivity...")
	if err := client.Ping(); err != nil {
		log.Fatalf("❌ Cannot connect to server: %v\n\nMake sure the server is running:\n  cd cmd/examples/01_basic_overview/04_server_rpc\n  go run main.go", err)
	}
	printSuccess("Server is responsive!")

	// Run all demos
	demoStringReturn(client)
	demoInterfaceReturn(client)
	demoSliceInterfaceReturn(client)
	demoMapReturn(client)
	demoStructReturn(client)
	demoPrimitiveReturns(client)
	demoDynamicReturn(client)
	demoVoidOperations(client)

	// Final summary
	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Println("🎉 All return type examples completed successfully!")
	fmt.Println("📖 This demonstrates the full range of return types supported by Lokstra RPC:")
	fmt.Println("   • string, error")
	fmt.Println("   • interface, error (UserIface → *User)")
	fmt.Println("   • []interface, error ([]UserIface → []*User)")
	fmt.Println("   • map[string]any, error")
	fmt.Println("   • struct, error (SystemInfo)")
	fmt.Println("   • primitive types, error (int, bool, time.Time)")
	fmt.Println("   • any, error (dynamic types)")
	fmt.Println("   • error only (void operations)")
	fmt.Printf("%s\n", strings.Repeat("=", 70))
}
