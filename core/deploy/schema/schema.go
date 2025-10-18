package schema

// DeployConfig is the root configuration structure for YAML files
// This matches the JSON schema and supports multi-file merging
type DeployConfig struct {
	Configs                  map[string]any                  `yaml:"configs" json:"configs"`
	ServiceDefinitions       map[string]*ServiceDef          `yaml:"service-definitions" json:"service-definitions"`
	Routers                  map[string]*RouterDefSimple     `yaml:"routers" json:"routers"`
	RemoteServiceDefinitions map[string]*RemoteServiceSimple `yaml:"remote-service-definitions" json:"remote-service-definitions"`
	Deployments              map[string]*DeploymentDefMap    `yaml:"deployments" json:"deployments"`
}

// RouterDefSimple is a simplified router definition for YAML
type RouterDefSimple struct {
	Service   string                  `yaml:"service" json:"service"`
	Overrides map[string]*RouteConfig `yaml:"overrides,omitempty" json:"overrides,omitempty"`
}

// RouteConfig defines route-specific configuration
type RouteConfig struct {
	Hide       bool     `yaml:"hide,omitempty" json:"hide,omitempty"`
	Middleware []string `yaml:"middleware,omitempty" json:"middleware,omitempty"`
}

// DeploymentDefMap is a deployment using map structure
type DeploymentDefMap struct {
	ConfigOverrides map[string]any           `yaml:"config-overrides,omitempty" json:"config-overrides,omitempty"`
	Servers         map[string]*ServerDefMap `yaml:"servers" json:"servers"`
}

// ServerDefMap is a server using map structure
type ServerDefMap struct {
	BaseURL string       `yaml:"base-url" json:"base-url"`
	Apps    []*AppDefMap `yaml:"apps" json:"apps"`
}

// AppDefMap is an app using map structure
type AppDefMap struct {
	Addr           string   `yaml:"addr" json:"addr"` // e.g., ":8080", "127.0.0.1:8080", "unix:/tmp/app.sock"
	Services       []string `yaml:"required-services,omitempty" json:"required-services,omitempty"`
	Routers        []string `yaml:"routers,omitempty" json:"routers,omitempty"`
	RemoteServices []string `yaml:"required-remote-services,omitempty" json:"required-remote-services,omitempty"`
}

// RemoteServiceSimple defines a remote service (simple YAML structure)
type RemoteServiceSimple struct {
	URL            string `yaml:"url" json:"url"`
	Resource       string `yaml:"resource" json:"resource"`
	ResourcePlural string `yaml:"resource-plural,omitempty" json:"resource-plural,omitempty"`
}

// DeploymentConfig is the root YAML configuration
type DeploymentConfig struct {
	// Global definitions (available to all deployments)
	Configs         []ConfigDef         `yaml:"configs"`
	Middlewares     []MiddlewareDef     `yaml:"middlewares"`
	Services        []ServiceDef        `yaml:"services"`
	Routers         []RouterDef         `yaml:"routers"`
	RouterOverrides []RouterOverrideDef `yaml:"router-overrides"`
	ServiceRouters  []ServiceRouterDef  `yaml:"service-routers"`

	// Deployments (select from global definitions)
	Deployments []DeploymentDef `yaml:"deployments"`
}

// ConfigDef defines a configuration value
type ConfigDef struct {
	Name  string `yaml:"name"`
	Value any    `yaml:"value"` // Can be string or ${...} reference
}

// MiddlewareDef defines a middleware instance
type MiddlewareDef struct {
	Name   string         `yaml:"name"`
	Type   string         `yaml:"type"`   // Factory type
	Config map[string]any `yaml:"config"` // Optional config
}

// ServiceDef defines a service instance
type ServiceDef struct {
	Name      string         `yaml:"name"`
	Type      string         `yaml:"type"`       // Factory type
	DependsOn []string       `yaml:"depends-on"` // Dependencies (can be "paramName:serviceName")
	Config    map[string]any `yaml:"config"`     // Optional config
}

// RouterDef defines a manual router (created in code)
type RouterDef struct {
	Name string `yaml:"name"`
	// Manual routers are just referenced by name
	// Actual router is created in code via router.New()
}

// RouterOverrideDef defines route overrides for a service router
type RouterOverrideDef struct {
	Name        string     `yaml:"name"`
	PathPrefix  string     `yaml:"path-prefix"` // Optional path prefix
	Middlewares []string   `yaml:"middlewares"` // Router-level middlewares
	Hidden      []string   `yaml:"hidden"`      // Methods to hide
	Routes      []RouteDef `yaml:"routes"`      // Individual route overrides
}

// RouteDef defines a single route override
type RouteDef struct {
	Name        string   `yaml:"name"`        // Method name
	Path        string   `yaml:"path"`        // Optional path override
	Method      string   `yaml:"method"`      // Optional HTTP method override
	Middlewares []string `yaml:"middlewares"` // Route-level middlewares
	Enabled     *bool    `yaml:"enabled"`     // nil = use default, true/false = override
}

// ServiceRouterDef defines a router auto-generated from a service
type ServiceRouterDef struct {
	Name       string `yaml:"name"`       // Router name
	Service    string `yaml:"service"`    // Service name to generate router from
	Convention string `yaml:"convention"` // Convention type (rest, rpc, custom)
	Overrides  string `yaml:"overrides"`  // Reference to RouterOverrideDef name
}

// DeploymentDef defines a deployment configuration
type DeploymentDef struct {
	Name            string         `yaml:"name"`
	ConfigOverrides map[string]any `yaml:"config-overrides"` // Override global configs
	Servers         []ServerDef    `yaml:"servers"`
}

// ServerDef defines a server in a deployment
type ServerDef struct {
	Name    string   `yaml:"name"`
	BaseURL string   `yaml:"base-url"`
	Apps    []AppDef `yaml:"apps"`
}

// AppDef defines an application running on a server
type AppDef struct {
	Addr           string             `yaml:"addr"`                     // e.g., ":8080", "127.0.0.1:8080", "unix:/tmp/app.sock"
	Services       []string           `yaml:"required-services"`        // Service names to instantiate
	Routers        []string           `yaml:"routers"`                  // Manual router names
	ServiceRouters []ServiceRouterRef `yaml:"service-routers"`          // Service routers (can be name or config)
	RemoteServices []RemoteServiceDef `yaml:"required-remote-services"` // Remote service proxies
}

// ServiceRouterRef can be either:
// - Just a string name (reference to global service-router)
// - Inline configuration
type ServiceRouterRef struct {
	// If Name is set, this is a reference to a global service-router
	Name string `yaml:"name"`

	// If Service is set, this is an inline service-router configuration
	Service    string `yaml:"service"`
	Convention string `yaml:"convention"`
	Overrides  string `yaml:"overrides"`
}

// RemoteServiceDef defines a remote service proxy
type RemoteServiceDef struct {
	Name          string `yaml:"name"`           // Service name
	URL           string `yaml:"url"`            // Remote service URL
	ServiceRouter string `yaml:"service-router"` // Optional: reference to service-router for convention
	Convention    string `yaml:"convention"`     // Optional: inline convention
	Overrides     string `yaml:"overrides"`      // Optional: router overrides
}
