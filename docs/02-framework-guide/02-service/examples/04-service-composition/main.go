package main

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Demonstrates layered services, decorators, and composition patterns

// Base Service
type DataService struct{}

func (*DataService) Get(id int) string {
	return fmt.Sprintf("Data-%d", id)
}

func DataServiceFactory(deps map[string]any, config map[string]any) any {
	return &DataService{}
}

// Decorator Service (adds caching)
type CachedDataService struct {
	base  *DataService
	cache map[int]string
}

func NewCachedDataService(base *DataService) *CachedDataService {
	return &CachedDataService{
		base:  base,
		cache: make(map[int]string),
	}
}

func (s *CachedDataService) Get(id int) string {
	if cached, ok := s.cache[id]; ok {
		fmt.Printf("✓ Cache hit for %d\n", id)
		return cached
	}
	result := s.base.Get(id)
	s.cache[id] = result
	return result
}

func CachedDataServiceFactory(deps map[string]any, config map[string]any) any {
	base := deps["data-service"].(*service.Cached[*DataService])
	return NewCachedDataService(base.Get())
}

// HTTP Handlers
type GetDataParams struct {
	ID int `path:"id"`
}

func GetData(params *GetDataParams) *response.ApiHelper {
	svc := lokstra_registry.GetService[*CachedDataService]("cached-data-service")
	result := svc.Get(params.ID)
	return response.NewApiOk(map[string]any{
		"id":   params.ID,
		"data": result,
	})
}

func Home() *response.Response {
	return response.NewHtmlResponse(`<!DOCTYPE html>
<html><head><title>Service Composition</title></head>
<body><h1>🎭 Service Composition Example</h1>
<p>Demonstrates layered services and decorators.</p>
<ul><li>GET /data/:id - Get data (with caching decorator)</li></ul>
</body></html>`)
}

func main() {
	lokstra_registry.RegisterServiceType("data-service", DataServiceFactory)
	lokstra_registry.RegisterServiceType("cached-data-service", CachedDataServiceFactory)

	lokstra_registry.RegisterLazyService("data-service", DataServiceFactory, nil)
	lokstra_registry.RegisterLazyServiceWithDeps("cached-data-service",
		CachedDataServiceFactory,
		map[string]string{"data-service": "data-service"},
		nil, nil,
	)

	router := lokstra.NewRouter("service-composition")
	router.GET("/", Home)
	router.GET("/data/:id", GetData)

	app := lokstra.NewApp("service-composition", ":3000", router)
	fmt.Println("🚀 Service Composition Example")
	fmt.Println("📍 http://localhost:3000")

	if err := app.Run(0); err != nil {
		fmt.Println("Error:", err)
	}
}
