package schema

import _ "embed"

//go:embed lokstra.schema.json
var schemaBytes []byte

// GetSchemaBytes returns the embedded JSON schema for validation
func GetSchemaBytes() []byte {
	return schemaBytes
}

// DeployConfig is the root configuration structure for YAML files
// This matches the JSON schema and supports multi-file merging
type DeployConfig struct {
	Configs                    map[string]any                  `yaml:"configs" json:"configs"`
	MiddlewareDefinitions      map[string]*MiddlewareDef       `yaml:"middleware-definitions,omitempty" json:"middleware-definitions,omitempty"`
	ServiceDefinitions         map[string]*ServiceDef          `yaml:"service-definitions" json:"service-definitions"`
	Routers                    map[string]*RouterDef           `yaml:"routers" json:"routers"`
	RouterOverrides            map[string]*RouterOverrideDef   `yaml:"router-overrides,omitempty" json:"router-overrides,omitempty"`
	ExternalServiceDefinitions map[string]*RemoteServiceSimple `yaml:"external-service-definitions,omitempty" json:"external-service-definitions,omitempty"`
	Deployments                map[string]*DeploymentDefMap    `yaml:"deployments" json:"deployments"`
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
	BaseURL string       `yaml:"base-url" json:"base-url"`
	Apps    []*AppDefMap `yaml:"apps,omitempty" json:"apps,omitempty"`

	// Helper fields (1 server = 1 app shorthand)
	// If these are present, a new app will be created and PREPENDED to Apps array
	// This allows mixing shorthand with additional apps
	HelperAddr              string   `yaml:"addr,omitempty" json:"addr,omitempty"`
	HelperRouters           []string `yaml:"routers,omitempty" json:"routers,omitempty"`
	HelperPublishedServices []string `yaml:"published-services,omitempty" json:"published-services,omitempty"`
}

// AppDefMap is an app using map structure
type AppDefMap struct {
	Addr              string   `yaml:"addr" json:"addr"`                                                 // e.g., ":8080", "127.0.0.1:8080", "unix:/tmp/app.sock"
	Routers           []string `yaml:"routers,omitempty" json:"routers,omitempty"`                       // Routers to include in this app
	PublishedServices []string `yaml:"published-services,omitempty" json:"published-services,omitempty"` // Services to auto-generate routers for
}

// RemoteServiceSimple defines an external service (outside this deployment)
// For external services, you typically need to override everything since their API structure may differ
type RemoteServiceSimple struct {
	URL            string         `yaml:"url" json:"url"`
	Type           string         `yaml:"type,omitempty" json:"type,omitempty"`                       // Factory type (auto-creates service wrapper)
	Resource       string         `yaml:"resource,omitempty" json:"resource,omitempty"`               // Resource name (singular)
	ResourcePlural string         `yaml:"resource-plural,omitempty" json:"resource-plural,omitempty"` // Resource name (plural)
	Convention     string         `yaml:"convention,omitempty" json:"convention,omitempty"`           // Convention type (rest, rpc, graphql)
	Overrides      string         `yaml:"overrides,omitempty" json:"overrides,omitempty"`             // Reference to RouterOverrideDef name for full customization
	Config         map[string]any `yaml:"config,omitempty" json:"config,omitempty"`                   // Additional config for factory
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
