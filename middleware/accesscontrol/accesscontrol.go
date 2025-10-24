package accesscontrol

import (
	"net/http"
	"slices"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/jwtauth"
)

const MIDDLEWARE_TYPE = "accesscontrol"

// Config represents the configuration for access control middleware
type Config struct {
	AllowedRoles []string `json:"allowed_roles" yaml:"allowed_roles"` // List of roles that can access
	RoleField    string   `json:"role_field" yaml:"role_field"`       // Field name in metadata where role is stored
	ErrorMessage string   `json:"error_message" yaml:"error_message"` // Custom error message
}

// Middleware creates a role-based access control middleware
// It checks if the user has one of the allowed roles
// This middleware should be used after JwtAuth middleware
func Middleware(cfg *Config) request.HandlerFunc {
	return request.HandlerFunc(func(c *request.Context) error {
		// Get user info from context (set by jwtauth middleware)
		userInfo, ok := jwtauth.GetUserInfo(c.R.Context())
		if !ok {
			c.W.WriteHeader(http.StatusUnauthorized)
			c.W.Header().Set("Content-Type", "application/json")
			c.W.Write([]byte(`{"error":"user not authenticated"}`))
			return nil
		}

		// Extract role from user metadata
		var userRole string
		if userInfo.Metadata != nil {
			if role, ok := userInfo.Metadata[cfg.RoleField]; ok {
				if roleStr, ok := role.(string); ok {
					userRole = roleStr
				}
			}
		}

		// Check if user has one of the allowed roles
		if !slices.Contains(cfg.AllowedRoles, userRole) {
			c.W.WriteHeader(http.StatusForbidden)
			c.W.Header().Set("Content-Type", "application/json")
			errMsg := cfg.ErrorMessage
			if errMsg == "" {
				errMsg = "insufficient permissions"
			}
			c.W.Write([]byte(`{"error":"` + errMsg + `"}`))
			return nil
		}

		return c.Next()
	})
}

func MiddlewareFactory(params map[string]any) request.HandlerFunc {
	cfg := &Config{
		AllowedRoles: utils.GetValueFromMap(params, "allowed_roles", []string{}),
		RoleField:    utils.GetValueFromMap(params, "role_field", "role"),
		ErrorMessage: utils.GetValueFromMap(params, "error_message", ""),
	}

	return Middleware(cfg)
}

func Register() {
	lokstra_registry.RegisterMiddlewareFactory(MIDDLEWARE_TYPE, MiddlewareFactory,
		lokstra_registry.AllowOverride(true))
}

// Convenience functions for common role checks

// RequireAdmin creates middleware that requires admin role
func RequireAdmin() request.HandlerFunc {
	return Middleware(&Config{
		AllowedRoles: []string{"admin"},
		RoleField:    "role",
		ErrorMessage: "admin access required",
	})
}

// RequireAdminOrManager creates middleware that requires admin or manager role
func RequireAdminOrManager() request.HandlerFunc {
	return Middleware(&Config{
		AllowedRoles: []string{"admin", "manager"},
		RoleField:    "role",
		ErrorMessage: "admin or manager access required",
	})
}

// RequireRoles creates middleware that requires one of the specified roles
func RequireRoles(roles ...string) request.HandlerFunc {
	return Middleware(&Config{
		AllowedRoles: roles,
		RoleField:    "role",
	})
}
