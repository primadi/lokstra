package schema

// DeployConfig is the root configuration structure for YAML files
// This matches the JSON schema and supports multi-file merging
type DeployConfig struct {
	Configs                  map[string]any                  `yaml:"configs" json:"configs"`
	ServiceDefinitions       map[string]*ServiceDef          `yaml:"service-definitions" json:"service-definitions"`
	Routers                  map[string]*RouterDef           `yaml:"routers" json:"routers"`
	RouterOverrides          map[string]*RouterOverrideDef   `yaml:"router-overrides,omitempty" json:"router-overrides,omitempty"`
	RemoteServiceDefinitions map[string]*RemoteServiceSimple `yaml:"remote-service-definitions" json:"remote-service-definitions"`
	Deployments              map[string]*DeploymentDefMap    `yaml:"deployments" json:"deployments"`
}

// RouterDef defines a router auto-generated from a service
type RouterDef struct {
	Service        string `yaml:"service" json:"service"`                                     // Service name to generate router from
	Convention     string `yaml:"convention" json:"convention"`                               // Convention type (rest, rpc, graphql)
	Resource       string `yaml:"resource,omitempty" json:"resource,omitempty"`               // Singular form, e.g., "user"
	ResourcePlural string `yaml:"resource-plural,omitempty" json:"resource-plural,omitempty"` // Plural form, e.g., "users"
	Overrides      string `yaml:"overrides,omitempty" json:"overrides,omitempty"`             // Reference to RouterOverrideDef name
}

// RouterOverrideDef defines route overrides for a service router
// This is the YAML representation of autogen.RouteOverride
type RouterOverrideDef struct {
	PathPrefix  string     `yaml:"path-prefix,omitempty" json:"path-prefix,omitempty"` // e.g., "/api/v1"
	Middlewares []string   `yaml:"middlewares,omitempty" json:"middlewares,omitempty"` // Router-level middleware names
	Hidden      []string   `yaml:"hidden,omitempty" json:"hidden,omitempty"`           // Methods to hide
	Custom      []RouteDef `yaml:"custom,omitempty" json:"custom,omitempty"`           // Custom route definitions (array in YAML, converted to map at runtime)
}

// RouteDef defines a single route override
// This is the YAML representation of autogen.Route
type RouteDef struct {
	Name        string   `yaml:"name" json:"name"`                                   // Method name
	Method      string   `yaml:"method,omitempty" json:"method,omitempty"`           // HTTP method override
	Path        string   `yaml:"path,omitempty" json:"path,omitempty"`               // Path override
	Middlewares []string `yaml:"middlewares,omitempty" json:"middlewares,omitempty"` // Route-level middleware names
}

// DeploymentDefMap is a deployment using map structure
type DeploymentDefMap struct {
	ConfigOverrides map[string]any           `yaml:"config-overrides,omitempty" json:"config-overrides,omitempty"`
	Servers         map[string]*ServerDefMap `yaml:"servers" json:"servers"`
}

// ServerDefMap is a server using map structure
type ServerDefMap struct {
	BaseURL        string       `yaml:"base-url" json:"base-url"`
	Services       []string     `yaml:"required-services,omitempty" json:"required-services,omitempty"`               // Server-level services (shared across apps)
	RemoteServices []string     `yaml:"required-remote-services,omitempty" json:"required-remote-services,omitempty"` // Server-level remote services (shared)
	Apps           []*AppDefMap `yaml:"apps" json:"apps"`
}

// AppDefMap is an app using map structure
type AppDefMap struct {
	Addr           string   `yaml:"addr" json:"addr"`                                                             // e.g., ":8080", "127.0.0.1:8080", "unix:/tmp/app.sock"
	Services       []string `yaml:"required-services,omitempty" json:"required-services,omitempty"`               // App-level services (app-specific)
	Routers        []string `yaml:"routers,omitempty" json:"routers,omitempty"`                                   // Routers to include in this app (auto-published for discovery)
	RemoteServices []string `yaml:"required-remote-services,omitempty" json:"required-remote-services,omitempty"` // App-level remote services (app-specific)
}

// RemoteServiceSimple defines a remote service (simple YAML structure)
type RemoteServiceSimple struct {
	URL            string `yaml:"url" json:"url"`
	Resource       string `yaml:"resource" json:"resource"`
	ResourcePlural string `yaml:"resource-plural,omitempty" json:"resource-plural,omitempty"`
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
