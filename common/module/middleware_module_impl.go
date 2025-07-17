package module

import "github.com/primadi/lokstra/common/iface"

type MiddlewareModuleImpl struct {
	name    string
	factory iface.MiddlewareFactory
	meta    *iface.MiddlewareMeta
}

func NewMiddlewareModule(name string, factory iface.MiddlewareFactory,
	meta *iface.MiddlewareMeta) *MiddlewareModuleImpl {
	if meta == nil {
		meta = &iface.MiddlewareMeta{
			Priority:    50, // Default priority
			Description: "",
			Tags:        nil, // Default empty tags
		}
	}

	return &MiddlewareModuleImpl{
		name:    name,
		factory: factory,
		meta:    meta,
	}
}

func (m *MiddlewareModuleImpl) Name() string {
	return m.name
}

func (m *MiddlewareModuleImpl) Factory(config any) iface.MiddlewareFunc {
	return m.factory(config)
}

func (m *MiddlewareModuleImpl) Meta() *iface.MiddlewareMeta {
	return m.meta
}
