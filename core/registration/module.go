package registration

import "github.com/primadi/lokstra/core/iface"

// Module defines the interface for a module in Lokstra.
type Module interface {
	Name() string
	Description() string
	Register(regCtx iface.RegistrationContext) error
}
