package main

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/syncmap"
)

func main() {
	// 1. Bootstrap Lokstra framework
	lokstra_init.Bootstrap()

	// 2. Load application config
	lokstra_registry.LoadConfig("config.yaml")

	// 3. Register routers
	registerRouters()

	// 4. Run the server
	if err := lokstra_registry.RunConfiguredServer(); err != nil {
		panic(err)
	}
}

func registerRouters() {
	r := lokstra.NewRouter("main-router")
	r.GET("/ping", func() string { return "pong" })

	maps := syncmap.NewSyncMap[string]("test")

	type getValueParams struct {
		Key string `query:"key" validate:"required"`
	}

	r.GET("/get-value", func(ctx *lokstra.RequestContext, p *getValueParams) string {
		ret, err := maps.Get(ctx, p.Key)
		if err != nil {
			return ""
		}

		return ret
	})

	type setValueParams struct {
		Key   string `json:"key" validate:"required"`
		Value string `json:"value" validate:"required"`
	}

	r.POST("/set-value", func(ctx *lokstra.RequestContext,
		p *setValueParams) error {
		return maps.Set(ctx, p.Key, p.Value)
	})

	r.GET("/all-keys", func(ctx *lokstra.RequestContext) ([]string, error) {
		return maps.Keys(ctx)
	})

	lokstra_registry.RegisterRouter(r.Name(), r)
}
