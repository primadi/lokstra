package ifacehook

import (
	"lokstra/common/response"
	"lokstra/core"
)

type OnResponseHook interface {
	OnResponse(resp *response.Response, ctx *core.RequestContext)
}
