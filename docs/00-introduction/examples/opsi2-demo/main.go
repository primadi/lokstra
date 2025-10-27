package main

import (
	"fmt"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// UserRepository represents a simple repository
type UserRepository struct {
	Name string
}

// UserService represents a service with dependencies
type UserService struct {
	Repository *service.Cached[*UserRepository]
	MaxUsers   int
}

func main() {
	fmt.Println("=== Demonstrating Opsi 2: Unified Service Registration ===")

	// Step 1: Register factory types (like in lokstra.go)
	fmt.Println("1. Registering factory types...")

	lokstra_registry.RegisterServiceType("user-repository-factory",
		func(deps, cfg map[string]any) any {
			return &UserRepository{
				Name: "User Repository",
			}
		},
		nil, // No remote factory
	)

	lokstra_registry.RegisterServiceType("user-service-factory",
		func(deps, cfg map[string]any) any {
			// IMPORTANT: When using string factory type with depends-on,
			// dependencies are wrapped in service.Cached
			// Use service.Cast to get the actual value
			repo := service.Cast[*UserRepository](deps["user-repository"])
			maxUsers := cfg["max-users"].(int)

			return &UserService{
				Repository: repo,
				MaxUsers:   maxUsers,
			}
		},
		nil, // No remote factory
	)

	// Step 2: Register services using CODE (like YAML) - NEW!
	fmt.Println("2. Registering services via CODE (equivalent to YAML)...")

	// Register dependency first
	lokstra_registry.RegisterLazyService("user-repository",
		"user-repository-factory", // ← String factory type (NEW!)
		nil,
	)

	// Register service with dependency - THIS IS THE USER'S ORIGINAL REQUEST
	lokstra_registry.RegisterLazyService("user-service",
		"user-service-factory", // ← String factory type (NEW!)
		map[string]any{
			"depends-on": []string{"user-repository"}, // ← Dependency specification
			"max-users":  1000,                        // ← Config parameter
		},
	)

	fmt.Println("   ✓ user-repository registered")
	fmt.Println("   ✓ user-service registered with dependency on user-repository")
	fmt.Println()

	// Step 3: Retrieve and use the service
	fmt.Println("3. Retrieving service (auto-instantiation with dependency injection)...")

	userSvc := lokstra_registry.MustGetService[*UserService]("user-service")

	fmt.Printf("   ✓ Service retrieved successfully!\n")
	fmt.Printf("   ✓ Repository: %s\n", userSvc.Repository.Get().Name)
	fmt.Printf("   ✓ Max Users: %d\n", userSvc.MaxUsers)
	fmt.Println()

	// Step 4: Demonstrate equivalence with YAML
	fmt.Println("4. This CODE registration is equivalent to this YAML:")
	fmt.Print(`
   service-definitions:
     user-repository:
       type: user-repository-factory

     user-service:
       type: user-service-factory
       depends-on:
         - user-repository
       config:
         max-users: 1000
`)

	// Step 5: Show internal state
	fmt.Println("5. Internal registry state:")
	fmt.Printf("   ✓ HasLazyService('user-service'): %v\n",
		deploy.Global().HasLazyService("user-service"))

	serviceDef := deploy.Global().GetDeferredServiceDef("user-service")
	if serviceDef != nil {
		fmt.Printf("   ✓ Service definition found:\n")
		fmt.Printf("     - Type: %s\n", serviceDef.Type)
		fmt.Printf("     - Dependencies: %v\n", serviceDef.DependsOn)
		fmt.Printf("     - Config: %v\n", serviceDef.Config)
	}

	fmt.Println("\n=== SUCCESS: Opsi 2 implementation working! ===")
}
