package registration

// Module defines the interface for a module in Lokstra.
type Module interface {
	Name() string
	Description() string
	Register(regCtx Context) error
}
