package main

import (
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/service"
	lokstra_registry "github.com/primadi/lokstra/lokstra_registry"
)

// UserService - simple service for demo
type UserService struct {
	callCount int
}

func (s *UserService) GetUsers() []string {
	s.callCount++
	return []string{"Alice", "Bob", "Charlie"}
}

func (s *UserService) GetCallCount() int {
	return s.callCount
}

// Package-level LazyLoad (cached!)
var userServiceLazy = service.LazyLoad[*UserService]("user-service")

func main() {
	// Register service
	lokstra_registry.RegisterServiceType("user-service", func() any {
		log.Println("🏭 UserService factory called (creating instance)")
		return &UserService{}
	})

	// Create router with 3 different access methods
	router := lokstra.NewRouter("demo")

	// Method 1: GetService() - Lookup every request
	router.GET("/method1-getservice", func() map[string]any {
		start := time.Now()

		// ⚠️ Registry lookup EVERY request
		userService := lokstra_registry.GetService[*UserService]("user-service")
		if userService == nil {
			return map[string]any{"error": "service not found"}
		}

		users := userService.GetUsers()
		elapsed := time.Since(start)

		return map[string]any{
			"method":      "GetService()",
			"users":       users,
			"elapsed_ns":  elapsed.Nanoseconds(),
			"performance": "SLOW (~100-200ns overhead)",
			"note":        "Registry map lookup every request",
		}
	})

	// Method 2: MustGetService() - Lookup every request, panics if not found
	router.GET("/method2-mustgetservice", func() map[string]any {
		start := time.Now()

		// ⚠️ Registry lookup EVERY request, panics if not found
		userService := lokstra_registry.MustGetService[*UserService]("user-service")
		users := userService.GetUsers()
		elapsed := time.Since(start)

		return map[string]any{
			"method":      "MustGetService()",
			"users":       users,
			"elapsed_ns":  elapsed.Nanoseconds(),
			"performance": "SLOW (~100-200ns overhead)",
			"note":        "Registry map lookup every request, panics if not found",
		}
	})

	// Method 3: LazyLoad() - Cached! ⭐ RECOMMENDED
	router.GET("/method3-lazyload", func() map[string]any {
		start := time.Now()

		// ✅ Cached! Only 1-5ns overhead after first access
		userService := userServiceLazy.MustGet()
		users := userService.GetUsers()
		elapsed := time.Since(start)

		return map[string]any{
			"method":      "LazyLoad.MustGet()",
			"users":       users,
			"elapsed_ns":  elapsed.Nanoseconds(),
			"performance": "FAST (~1-5ns overhead after first access)",
			"note":        "Cached after first access - 20-100x faster!",
		}
	})

	// Benchmark endpoint - call each method 1000 times
	router.GET("/benchmark", func() map[string]any {
		iterations := 1000

		// Warmup
		_ = lokstra_registry.GetService[*UserService]("user-service")
		_ = userServiceLazy.MustGet()

		// Benchmark Method 1: GetService()
		start1 := time.Now()
		for i := 0; i < iterations; i++ {
			userService := lokstra_registry.GetService[*UserService]("user-service")
			if userService == nil {
				continue
			}
			_ = userService.GetUsers()
		}
		elapsed1 := time.Since(start1)

		// Benchmark Method 2: MustGetService()
		start2 := time.Now()
		for i := 0; i < iterations; i++ {
			userService := lokstra_registry.MustGetService[*UserService]("user-service")
			_ = userService.GetUsers()
		}
		elapsed2 := time.Since(start2)

		// Benchmark Method 3: LazyLoad()
		start3 := time.Now()
		for i := 0; i < iterations; i++ {
			userService := userServiceLazy.MustGet()
			_ = userService.GetUsers()
		}
		elapsed3 := time.Since(start3)

		avgNs1 := elapsed1.Nanoseconds() / int64(iterations)
		avgNs2 := elapsed2.Nanoseconds() / int64(iterations)
		avgNs3 := elapsed3.Nanoseconds() / int64(iterations)

		speedup := float64(avgNs1) / float64(avgNs3)

		return map[string]any{
			"iterations": iterations,
			"results": map[string]any{
				"GetService": map[string]any{
					"total_ns": elapsed1.Nanoseconds(),
					"avg_ns":   avgNs1,
					"note":     "Map lookup every call",
				},
				"MustGetService": map[string]any{
					"total_ns": elapsed2.Nanoseconds(),
					"avg_ns":   avgNs2,
					"note":     "Map lookup + panic check every call",
				},
				"LazyLoad": map[string]any{
					"total_ns": elapsed3.Nanoseconds(),
					"avg_ns":   avgNs3,
					"note":     "Cached after first access",
				},
			},
			"comparison": map[string]any{
				"speedup":        fmt.Sprintf("%.1fx faster", speedup),
				"winner":         "LazyLoad",
				"recommendation": "Use LazyLoad for production code!",
			},
		}
	})

	// Stats endpoint
	router.GET("/stats", func() map[string]any {
		userService := userServiceLazy.MustGet()
		return map[string]any{
			"total_calls": userService.GetCallCount(),
			"note":        "Total times GetUsers() was called across all methods",
		}
	})

	// Create app
	app := lokstra.NewApp("lazyload-demo", ":3000", router)

	fmt.Println("🚀 LazyLoad vs GetService Demo")
	fmt.Println("📊 Compare performance of different service access methods:")
	fmt.Println()
	fmt.Println("   Method 1 (SLOW):  GET /method1-getservice")
	fmt.Println("   Method 2 (SLOW):  GET /method2-mustgetservice")
	fmt.Println("   Method 3 (FAST):  GET /method3-lazyload ⭐")
	fmt.Println()
	fmt.Println("   Benchmark:        GET /benchmark")
	fmt.Println("   Stats:            GET /stats")
	fmt.Println()
	fmt.Println("🔬 Expected results:")
	fmt.Println("   - GetService/MustGetService: ~100-200ns per call")
	fmt.Println("   - LazyLoad: ~1-5ns per call (20-100x faster!)")
	fmt.Println()
	fmt.Println("Server: http://localhost:3000")

	// Run
	if err := app.Run(30 * time.Second); err != nil {
		log.Fatal(err)
	}
}
