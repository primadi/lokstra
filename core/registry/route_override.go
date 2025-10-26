package registry

type RouteOverride struct {
	Name              string   `json:"name"`
	Path              string   `json:"path"`
	Method            string   `json:"method"`
	Middlewares       []string `json:"middlewares"`
	OverrrideParentMw *bool    `json:"override_parent_mw"`
}

type RouterOverride struct {
	Name              string           `json:"name"`
	BasePath          *string          `json:"base_path"`
	Middlewares       *[]string        `json:"middlewares"`
	OverrrideParentMw *bool            `json:"override_parent_mw"`
	Routes            *[]RouteOverride `json:"routes"`
}

func NewRouterOverride(name string) *RouterOverride {
	return &RouterOverride{
		Name: name,
	}
}

func (ro *RouterOverride) SetBasePath(basePath string) *RouterOverride {
	ro.BasePath = &basePath
	return ro
}
func (ro *RouterOverride) SetMiddlewares(overrideParent bool, mws ...string) *RouterOverride {
	ro.Middlewares = &mws
	ro.OverrrideParentMw = &overrideParent
	return ro
}
func (ro *RouterOverride) AddRoute(routeName, path, method string, overrideParent bool, mws ...string) *RouterOverride {
	route := RouteOverride{
		Name:              routeName,
		Path:              path,
		Method:            method,
		Middlewares:       mws,
		OverrrideParentMw: &overrideParent,
	}
	*ro.Routes = append(*ro.Routes, route)
	return ro
}
