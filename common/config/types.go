package config

type LokstraConfig struct {
	Server   *ServerConfig    `yaml:"server,omitempty"`
	Apps     []*AppConfig     `yaml:"apps,omitempty"`
	Services []*ServiceConfig `yaml:"services,omitempty"`
	Modules  []*ModuleConfig  `yaml:"modules,omitempty"`
}

type ServerConfig struct {
	Name     string         `yaml:"name"`
	Settings map[string]any `yaml:"global_setting,omitempty"`
}

type MountStaticConfig struct {
	Prefix string `yaml:"prefix"`
	Folder string `yaml:"folder"`
}

type MountSPAConfig struct {
	Prefix       string `yaml:"prefix"`
	FallbackFile string `yaml:"fallback_file"`
}

type MountReverseProxyConfig struct {
	Prefix string `yaml:"prefix"`
	Target string `yaml:"target"`
}

type GroupConfig struct {
	Prefix string        `yaml:"prefix"`
	Groups []GroupConfig `yaml:"groups,omitempty"`
	Routes []RouteConfig `yaml:"routes,omitempty"`

	MiddlewareRaw      any                `yaml:"middleware"`
	Middleware         []MiddlewareConfig `yaml:"-"`
	OverrideMiddleware bool               `yaml:"override_middleware,omitempty"`

	MountStatic       []MountStaticConfig       `yaml:"mount_static,omitempty"`
	MountSPA          []MountSPAConfig          `yaml:"mount_spa,omitempty"`
	MountReverseProxy []MountReverseProxyConfig `yaml:"mount_reverse_proxy,omitempty"`
}

type AppConfig struct {
	Name             string `yaml:"name"`
	Address          string `yaml:"address"`
	ListenerType     string `yaml:"listener_type"`
	RouterEngineType string `yaml:"router_engine_type"`

	Groups        []GroupConfig      `yaml:"groups,omitempty"`
	Routes        []RouteConfig      `yaml:"routes,omitempty"`
	MiddlewareRaw any                `yaml:"middleware"`
	Middleware    []MiddlewareConfig `yaml:"-"`

	Settings          map[string]any            `yaml:"setting,omitempty"`
	MountStatic       []MountStaticConfig       `yaml:"mount_static,omitempty"`
	MountSPA          []MountSPAConfig          `yaml:"mount_spa,omitempty"`
	MountReverseProxy []MountReverseProxyConfig `yaml:"mount_reverse_proxy,omitempty"`
}

type MiddlewareConfig struct {
	Name    string         `yaml:"name"`
	Enabled bool           `yaml:"enabled,omitempty"`
	Config  map[string]any `yaml:"config,omitempty"`
}

type ServiceConfig struct {
	Name   string         `yaml:"name"`
	Type   string         `yaml:"type"`
	Config map[string]any `yaml:"config"`
}

type ModuleConfig struct {
	Name        string         `yaml:"name"`
	Path        string         `yaml:"path"`
	Entry       string         `yaml:"entry, omitempty"`
	Settings    map[string]any `yaml:"settings,omitempty"`
	Permissions map[string]any `yaml:"permissions,omitempty"`
}

type RouteConfig struct {
	Method             string             `yaml:"method"`
	Path               string             `yaml:"path"`
	Handler            string             `yaml:"handler"`
	OverrideMiddleware bool               `yaml:"override_middleware,omitempty"`
	MiddlewareRaw      any                `yaml:"middleware"`
	Middleware         []MiddlewareConfig `yaml:"-"`
}
