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
	NamedDbPools               map[string]*DbPoolConfig        `yaml:"named-db-pools,omitempty" json:"named-db-pools,omitempty"`
	MiddlewareDefinitions      map[string]*MiddlewareDef       `yaml:"middleware-definitions,omitempty" json:"middleware-definitions,omitempty"`
	ServiceDefinitions         map[string]*ServiceDef          `yaml:"service-definitions" json:"service-definitions"`
	RouterDefinitions          map[string]*RouterDef           `yaml:"router-definitions,omitempty" json:"router-definitions,omitempty"` // Renamed from Routers
	ExternalServiceDefinitions map[string]*RemoteServiceSimple `yaml:"external-service-definitions,omitempty" json:"external-service-definitions,omitempty"`
	Deployments                map[string]*DeploymentDefMap    `yaml:"deployments" json:"deployments"`
}

// DbPoolConfig defines configuration for a named database pool
type DbPoolConfig struct {
	// Connection via DSN (highest priority)
	DSN string `yaml:"dsn,omitempty" json:"dsn,omitempty"`

	// Connection via components (used if DSN not provided)
	Host     string `yaml:"host,omitempty" json:"host,omitempty"`
	Port     int    `yaml:"port,omitempty" json:"port,omitempty"`
	Database string `yaml:"database,omitempty" json:"database,omitempty"`
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// Schema configuration
	Schema string `yaml:"schema,omitempty" json:"schema,omitempty"` // Default: "public"

	// Pool configuration (optional)
	MinConns    int    `yaml:"min-conns,omitempty" json:"min-conns,omitempty"`
	MaxConns    int    `yaml:"max-conns,omitempty" json:"max-conns,omitempty"`
	MaxIdleTime string `yaml:"max-idle-time,omitempty" json:"max-idle-time,omitempty"` // Duration string (e.g., "30m")
	MaxLifetime string `yaml:"max-lifetime,omitempty" json:"max-lifetime,omitempty"`   // Duration string (e.g., "1h")
	SSLMode     string `yaml:"sslmode,omitempty" json:"sslmode,omitempty"`             // SSL mode (disable, require, etc.)
}

// RouterDef defines a router auto-generated from a service
// Service name is derived from router name by removing "-router" suffix
// Example: "user-service-router" → service is "user-service"
type RouterDef struct {
	// Basic configuration
	Convention     string `yaml:"convention,omitempty" json:"convention,omitempty"`           // Convention type (rest, rpc, graphql) - optional if set in RegisterServiceType
	Resource       string `yaml:"resource,omitempty" json:"resource,omitempty"`               // Singular form, e.g., "user" - optional if set in RegisterServiceType
	ResourcePlural string `yaml:"resource-plural,omitempty" json:"resource-plural,omitempty"` // Plural form, e.g., "users" - optional if set in RegisterServiceType

	// Override configuration (inline - no more references)
	PathPrefix   string           `yaml:"path-prefix,omitempty" json:"path-prefix,omitempty"`     // e.g., "/api/v1"
	PathRewrites []PathRewriteDef `yaml:"path-rewrites,omitempty" json:"path-rewrites,omitempty"` // Regex-based path rewrites
	Middlewares  []string         `yaml:"middlewares,omitempty" json:"middlewares,omitempty"`     // Router-level middleware names
	Hidden       []string         `yaml:"hidden,omitempty" json:"hidden,omitempty"`               // Methods to hide
	Custom       []RouteDef       `yaml:"custom,omitempty" json:"custom,omitempty"`               // Custom route definitions (array in YAML)
}

// PathRewriteDef defines a regex-based path rewrite rule
type PathRewriteDef struct {
	Pattern     string `yaml:"pattern" json:"pattern"`         // Regex pattern to match (e.g., "^/api/v1/(.*)$")
	Replacement string `yaml:"replacement" json:"replacement"` // Replacement string (e.g., "/api/v2/$1")
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
	ConfigOverrides map[string]any `yaml:"config-overrides,omitempty" json:"config-overrides,omitempty"`

	// Inline definitions at deployment level (will be normalized to {deployment}.{name})
	InlineMiddlewares      map[string]*MiddlewareDef       `yaml:"middleware-definitions,omitempty" json:"middleware-definitions,omitempty"`
	InlineServices         map[string]*ServiceDef          `yaml:"service-definitions,omitempty" json:"service-definitions,omitempty"`
	InlineRouters          map[string]*RouterDef           `yaml:"router-definitions,omitempty" json:"router-definitions,omitempty"`
	InlineExternalServices map[string]*RemoteServiceSimple `yaml:"external-service-definitions,omitempty" json:"external-service-definitions,omitempty"`

	Servers map[string]*ServerDefMap `yaml:"servers" json:"servers"`
}

// ServerDefMap is a server using map structure
type ServerDefMap struct {
	BaseURL string `yaml:"base-url" json:"base-url"`

	// Inline definitions at server level (will be normalized to {deployment}.{server}.{name})
	InlineMiddlewares      map[string]*MiddlewareDef       `yaml:"middleware-definitions,omitempty" json:"middleware-definitions,omitempty"`
	InlineServices         map[string]*ServiceDef          `yaml:"service-definitions,omitempty" json:"service-definitions,omitempty"`
	InlineRouters          map[string]*RouterDef           `yaml:"router-definitions,omitempty" json:"router-definitions,omitempty"`
	InlineExternalServices map[string]*RemoteServiceSimple `yaml:"external-service-definitions,omitempty" json:"external-service-definitions,omitempty"`

	Apps []*AppDefMap `yaml:"apps,omitempty" json:"apps,omitempty"`

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

	// Handler configurations (mount at app level)
	ReverseProxies []*ReverseProxyDef `yaml:"reverse-proxies,omitempty" json:"reverse-proxies,omitempty"` // Reverse proxy configurations
	MountSpa       []*MountSpaDef     `yaml:"mount-spa,omitempty" json:"mount-spa,omitempty"`             // SPA mount configurations
	MountStatic    []*MountStaticDef  `yaml:"mount-static,omitempty" json:"mount-static,omitempty"`       // Static file mount configurations
}

// RemoteServiceSimple defines an external service (outside this deployment)
// For external services, you typically need to override everything since their API structure may differ
type RemoteServiceSimple struct {
	URL    string         `yaml:"url" json:"url"`
	Type   string         `yaml:"type,omitempty" json:"type,omitempty"`     // Factory type (auto-creates service wrapper)
	Router *RouterDef     `yaml:"router,omitempty" json:"router,omitempty"` // Embedded router definition
	Config map[string]any `yaml:"config,omitempty" json:"config,omitempty"` // Additional config for factory
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
	Type      string         `yaml:"type"`             // Factory type
	DependsOn []string       `yaml:"depends-on"`       // Dependencies (can be "paramName:serviceName")
	Router    *RouterDef     `yaml:"router,omitempty"` // Embedded router definition (auto-generated router for this service)
	Config    map[string]any `yaml:"config"`           // Optional config
}

// ReverseProxyDef defines a reverse proxy configuration
type ReverseProxyDef struct {
	Prefix      string                  `yaml:"prefix" json:"prefix"`                                 // URL prefix to match (e.g., "/api")
	StripPrefix bool                    `yaml:"strip-prefix,omitempty" json:"strip-prefix,omitempty"` // Whether to strip the prefix before forwarding
	Target      string                  `yaml:"target" json:"target"`                                 // Target backend URL (e.g., "http://api-server:8080")
	Rewrite     *ReverseProxyRewriteDef `yaml:"rewrite,omitempty" json:"rewrite,omitempty"`           // Path rewrite rules
}

// ReverseProxyRewriteDef represents path rewrite rules for reverse proxy
type ReverseProxyRewriteDef struct {
	From string `yaml:"from" json:"from"` // Pattern to match in path (regex supported)
	To   string `yaml:"to" json:"to"`     // Replacement pattern
}

// MountSpaDef defines a Single Page Application mount configuration
type MountSpaDef struct {
	Prefix string `yaml:"prefix" json:"prefix"` // URL prefix (e.g., "/app", "/")
	Dir    string `yaml:"dir" json:"dir"`       // Directory path containing SPA files (e.g., "./dist", "./build")
}

// MountStaticDef defines a static file mount configuration
type MountStaticDef struct {
	Prefix string `yaml:"prefix" json:"prefix"` // URL prefix (e.g., "/static", "/assets")
	Dir    string `yaml:"dir" json:"dir"`       // Directory path containing static files (e.g., "./public", "./static")
}
