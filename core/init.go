package core

import (
	"lokstra/common/response"
	"lokstra/common/response/response_iface"
)

func init() {
	globalRuntime.responseFormatter = response.NewJSONFormatter()
	globalRuntime.responseTemplateFunc = response_iface.DefaultTemplateFunc
}
