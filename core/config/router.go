package config

type MiddlewareConfig struct {
	Name      string `yaml:"name"`
	Parameter []any  `yaml:"parameter,omitempty"` // optional, can be empty
}

type RouteConfig struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Method      string             `yaml:"method"`
	Path        string             `yaml:"path"`
	Middleware  []MiddlewareConfig `yaml:"middleware"`
}

type RouterConfig struct {
	Routes []RouteConfig `yaml:"routes"`
}
