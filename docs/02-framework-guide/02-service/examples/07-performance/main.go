package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Demonstrates performance patterns: LazyLoad vs Direct access

type DataService struct {
	name string
}

func NewDataService(name string) *DataService {
	return &DataService{name: name}
}

func (s *DataService) GetData() string {
	return fmt.Sprintf("Data from %s", s.name)
}

func DataServiceFactory(deps map[string]any, config map[string]any) any {
	return NewDataService("DataService")
}

// Service using LazyLoad
type LazyConsumer struct {
	data *service.Cached[*DataService]
}

func NewLazyConsumer(data *service.Cached[*DataService]) *LazyConsumer {
	return &LazyConsumer{data: data}
}

func (s *LazyConsumer) ProcessLazy() string {
	// Lazy evaluation - only gets service when needed
	return s.data.Get().GetData()
}

func LazyConsumerFactory(deps map[string]any, config map[string]any) any {
	return NewLazyConsumer(deps["data-service"].(*service.Cached[*DataService]))
}

// Service using Direct access
type DirectConsumer struct {
	data *DataService
}

func NewDirectConsumer(data *DataService) *DirectConsumer {
	return &DirectConsumer{data: data}
}

func (s *DirectConsumer) ProcessDirect() string {
	return s.data.GetData()
}

func DirectConsumerFactory(deps map[string]any, config map[string]any) any {
	data := deps["data-service"].(*service.Cached[*DataService])
	return NewDirectConsumer(data.Get())
}

// HTTP Handlers
func TestLazy() *response.ApiHelper {
	start := time.Now()
	svc := lokstra_registry.GetService[*LazyConsumer]("lazy-consumer")

	var results []string
	for i := 0; i < 1000; i++ {
		results = append(results, svc.ProcessLazy())
	}

	elapsed := time.Since(start)

	return response.NewApiOk(map[string]any{
		"method":      "LazyLoad",
		"iterations":  1000,
		"duration_ms": elapsed.Milliseconds(),
		"sample":      results[0],
	})
}

func TestDirect() *response.ApiHelper {
	start := time.Now()
	svc := lokstra_registry.GetService[*DirectConsumer]("direct-consumer")

	var results []string
	for i := 0; i < 1000; i++ {
		results = append(results, svc.ProcessDirect())
	}

	elapsed := time.Since(start)

	return response.NewApiOk(map[string]any{
		"method":      "Direct",
		"iterations":  1000,
		"duration_ms": elapsed.Milliseconds(),
		"sample":      results[0],
	})
}

func Home() *response.Response {
	return response.NewHtmlResponse(`<!DOCTYPE html>
<html><head><title>Performance</title></head>
<body><h1>⚡ Performance Example</h1>
<p>Benchmarks comparing LazyLoad vs Direct service access.</p>
<ul>
<li>GET /lazy - Test LazyLoad pattern (1000 iterations)</li>
<li>GET /direct - Test Direct access (1000 iterations)</li>
</ul>
</body></html>`)
}

func main() {
	lokstra_registry.RegisterServiceType("data-service", DataServiceFactory, nil)
	lokstra_registry.RegisterServiceType("lazy-consumer", LazyConsumerFactory, nil)
	lokstra_registry.RegisterServiceType("direct-consumer", DirectConsumerFactory, nil)

	lokstra_registry.RegisterLazyService("data-service", DataServiceFactory, nil)
	lokstra_registry.RegisterLazyServiceWithDeps("lazy-consumer",
		LazyConsumerFactory,
		map[string]string{"data-service": "data-service"},
		nil, nil,
	)
	lokstra_registry.RegisterLazyServiceWithDeps("direct-consumer",
		DirectConsumerFactory,
		map[string]string{"data-service": "data-service"},
		nil, nil,
	)

	router := lokstra.NewRouter("performance")
	router.GET("/", Home)
	router.GET("/lazy", TestLazy)
	router.GET("/direct", TestDirect)

	app := lokstra.NewApp("performance", ":3000", router)
	fmt.Println("🚀 Performance Example")
	fmt.Println("📍 http://localhost:3000")

	if err := app.Run(0); err != nil {
		fmt.Println("Error:", err)
	}
}
