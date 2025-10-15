package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/primadi/lokstra/cmd_draft/examples/23-service-local-remote/user_service"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

var local_service = service.LazyLoad[user_service.UserService]("user-service")

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("🧪 Service Local/Remote Example")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// ========================================================================
	// STEP 1: Register Service Factories (Local + Remote)
	// ========================================================================
	fmt.Println("📋 Step 1: Registering service factories...")

	// Register LOCAL factory - creates real implementation
	lokstra_registry.RegisterServiceFactoryLocal(
		"user",
		user_service.CreateLocalUserService,
	)

	// Register REMOTE factory - creates HTTP client
	lokstra_registry.RegisterServiceFactoryRemote(
		"user",
		user_service.CreateRemoteUserService,
	)

	fmt.Println("   ✅ Local factory registered")
	fmt.Println("   ✅ Remote factory registered")
	fmt.Println()

	// ========================================================================
	// STEP 2: Register Lazy Service (Framework will choose factory)
	// ========================================================================
	fmt.Println("📋 Step 2: Registering lazy service...")

	lokstra_registry.RegisterLazyService("user-service", "user", map[string]any{
		"router": "UserService", // Router name for remote calls
	})

	fmt.Println("   ✅ Lazy service registered")
	fmt.Println()

	// ========================================================================
	// STEP 3: Create Router from Service (for server side)
	// ========================================================================
	fmt.Println("📋 Step 3: Creating router from service...")

	// Create router from service methods
	userRouter := router.NewFromService(local_service.MustGet(), router.DefaultServiceRouterOptions())

	fmt.Println("   ✅ Router created with auto-generated routes:")
	userRouter.Walk(func(rt *route.Route) {
		fmt.Printf("      %s %s\n", rt.Method, rt.FullPath)
	})
	fmt.Println()

	// ========================================================================
	// STEP 4: Register Router (makes it accessible)
	// ========================================================================
	fmt.Println("📋 Step 4: Registering router...")

	lokstra_registry.RegisterRouter("UserService", userRouter)

	fmt.Println("   ✅ Router registered")
	fmt.Println()

	// ========================================================================
	// STEP 5: Simulate Framework Decision
	// ========================================================================
	fmt.Println("📋 Step 5: Framework auto-detection...")
	fmt.Println()

	// Set current server name
	lokstra_registry.SetCurrentServerName("server-a")

	// Register ClientRouter for SAME server (will be LOCAL)
	lokstra_registry.RegisterClientRouter(
		"UserService",           // routerName
		"server-a",              // serverName (SAME as current)
		"http://localhost:3001", // baseURL
		"",                      // addr
		30*time.Second,          // timeout
	)

	fmt.Println("   🔍 Current server: server-a")
	fmt.Println("   🔍 UserService router: server-a")
	fmt.Println("   ✅ Decision: LOCAL (same server)")
	fmt.Println()

	// ========================================================================
	// STEP 6: Test Service Usage - LOCAL
	// ========================================================================
	fmt.Println("📋 Step 6: Testing LOCAL service...")
	fmt.Println()

	// Get service - framework chooses LOCAL factory automatically!
	testLocalService()

	fmt.Println()

	// ========================================================================
	// STEP 7: Simulate Remote Scenario
	// ========================================================================
	fmt.Println("📋 Step 7: Simulating REMOTE scenario...")
	fmt.Println()

	// Clear service instance to force re-creation
	// (In real app, this happens automatically per request)

	// Register ClientRouter for DIFFERENT server (will be REMOTE)
	lokstra_registry.RegisterClientRouter(
		"UserService",              // routerName
		"server-b",                 // serverName (DIFFERENT from current)
		"http://user-service:3002", // baseURL
		"",                         // addr
		30*time.Second,             // timeout
	)

	fmt.Println("   🔍 Current server: server-a")
	fmt.Println("   🔍 UserService router: server-b")
	fmt.Println("   ✅ Decision: REMOTE (different server)")
	fmt.Println()

	// ========================================================================
	// Summary
	// ========================================================================
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("📝 Summary:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("✅ Developer Benefits:")
	fmt.Println("   1. Register factories once (local + remote)")
	fmt.Println("   2. Framework auto-detects based on deployment")
	fmt.Println("   3. Same interface for local and remote")
	fmt.Println("   4. No manual if/else logic needed")
	fmt.Println("   5. Error if factory not registered (fail-fast)")
	fmt.Println()
	fmt.Println("✅ Framework Decision Logic:")
	fmt.Println("   - ClientRouter.IsLocal = true  → Use LOCAL factory")
	fmt.Println("   - ClientRouter.IsLocal = false → Use REMOTE factory")
	fmt.Println("   - No ClientRouter found        → Default to LOCAL")
	fmt.Println()
	fmt.Println("✅ Error Handling:")
	fmt.Println("   - Need local but only remote registered → PANIC")
	fmt.Println("   - Need remote but only local registered → PANIC")
	fmt.Println()
}

var userService = service.LazyLoad[user_service.UserService]("user-service")

func testLocalService() {
	// Create mock context
	ctx := &request.Context{}

	// Call method - looks the same whether local or remote!
	user, err := userService.MustGet().GetUser(ctx, &user_service.GetUserRequest{
		UserID: "123",
	})

	if err != nil {
		log.Printf("   ❌ Error: %v", err)
		return
	}

	fmt.Printf("   ✅ Result: ID=%s, Name=%s, Email=%s\n", user.ID, user.Name, user.Email)
}
