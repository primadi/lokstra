package ifacehook

import "lokstra/core"

type OnContextHook interface {
	OnContext(ctx *core.RequestContext)
}
