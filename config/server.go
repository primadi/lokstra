package config

type ServerConfig struct {
	Name     string
	Apps     []AppConfig
	Settings map[string]any
}

type AppConfig struct {
	Name   string
	Port   int
	Router RouterConfig
}

type RouterConfig struct {
	Middlewares []string      `json:"middlewares"`
	Routes      []RouteConfig `json:"routes"`
}

type RouteConfig struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Action string `json:"action"`
}
