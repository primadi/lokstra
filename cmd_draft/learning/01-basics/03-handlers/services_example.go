package main

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/lokstra_registry"
)

// This file demonstrates how handlers use services via ServiceContainer pattern
// See: cmd/learning/01-basics/05-services for detailed service patterns

// === Mock Services (simplified) ===

type UserService struct {
	name string
}

func (s *UserService) GetUser(id string) map[string]any {
	return map[string]any{
		"id":      id,
		"name":    "User " + id,
		"service": s.name,
	}
}

func (s *UserService) CreateUser(name string) map[string]any {
	return map[string]any{
		"id":      "new-id",
		"name":    name,
		"service": s.name,
	}
}

type CacheService struct {
	store map[string]string
	name  string
}

func (s *CacheService) Get(key string) (string, bool) {
	val, ok := s.store[key]
	fmt.Printf("Cache GET: %s -> %v\n", key, ok)
	return val, ok
}

func (s *CacheService) Set(key, value string) {
	s.store[key] = value
	fmt.Printf("Cache SET: %s = %s\n", key, value)
}

// === Service Container Pattern ===

type ServiceContainer struct {
	userCache  *UserService
	cacheCache *CacheService
}

// Lazy-loaded getters with caching
func (sc *ServiceContainer) GetUser() *UserService {
	sc.userCache = lokstra_registry.GetServiceCached("user", sc.userCache)
	return sc.userCache
}

func (sc *ServiceContainer) GetCache() *CacheService {
	sc.cacheCache = lokstra_registry.GetServiceCached("cache", sc.cacheCache)
	return sc.cacheCache
}

// Global service container
var services = &ServiceContainer{}

// === Handler Examples Using Services ===

func setupServiceRoutes(r lokstra.Router) {
	// Register mock services
	lokstra_registry.RegisterService("user", &UserService{name: "user-service"})
	lokstra_registry.RegisterService("cache", &CacheService{
		name:  "cache-service",
		store: make(map[string]string),
	})

	// Handler using services via container
	r.GET("/services/users/:id", func(c *lokstra.RequestContext) error {
		id := c.Req.PathParam("id", "")

		// Get services via container (lazy + cached)
		userService := services.GetUser()
		cacheService := services.GetCache()

		// Try cache first
		if cached, ok := cacheService.Get("user:" + id); ok {
			return c.Api.Ok(map[string]any{
				"source": "cache",
				"data":   cached,
			})
		}

		// Get from service
		user := userService.GetUser(id)

		// Store in cache
		cacheService.Set("user:"+id, fmt.Sprintf("%v", user))

		return c.Api.Ok(map[string]any{
			"source": "service",
			"data":   user,
		})
	})

	// Handler with smart binding + services
	type CreateUserReq struct {
		Name  string `json:"name" validate:"required,min=3"`
		Email string `json:"email" validate:"required,email"`
	}

	r.POST("/services/users", func(c *lokstra.RequestContext, req *CreateUserReq) error {
		userService := services.GetUser()
		cacheService := services.GetCache()

		// Create user
		user := userService.CreateUser(req.Name)

		// Cache it
		userID := user["id"].(string)
		cacheService.Set("user:"+userID, fmt.Sprintf("%v", user))

		return c.Api.Created(user, "User created")
	})
}

// Key Points:
// 1. ServiceContainer caches services at struct field level
// 2. Getter methods use GetService("name", cache) pattern
// 3. Handlers call services.GetUser(), services.GetCache()
// 4. First call loads from registry, subsequent calls use cache
// 5. Clean separation: handlers → container → services
