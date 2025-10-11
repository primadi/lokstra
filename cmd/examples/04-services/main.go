package main

import (
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

type CounterService struct {
	count int
}

func (s *CounterService) Increment() int {
	s.count++
	return s.count
}

func (s *CounterService) GetCount() int {
	return s.count
}

func CounterFactory(cfg map[string]any) any {
	seed := utils.GetValueFromMap(cfg, "seed", 0)

	return &CounterService{
		count: seed,
	}
}

func createAdminRouter() lokstra.Router {
	r := lokstra.NewRouter("main-router")

	myCounter := service.LazyLoad[*CounterService]("my-counter")

	r.GET("/count", func(c *lokstra.RequestContext) error {
		return c.Api.Ok(myCounter.Get().Increment())
	})
	return r
}

func main() {
	// Register the service factory and lazy service
	lokstra_registry.RegisterServiceFactory("counter", CounterFactory)
	lokstra_registry.RegisterLazyService("my-counter", "counter", map[string]any{"seed": 100})

	// Create app with main router
	mainRouter := createAdminRouter()

	// Create app
	app := lokstra.NewApp("admin-app", ":8080", mainRouter)
	// Create server with the app
	svr := lokstra.NewServer("service-di-server", app)

	// Print server and app info
	svr.PrintStartInfo()
	// Run server with 5 seconds graceful shutdown timeout
	svr.Run(5 * time.Second)
}
