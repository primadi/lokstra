package response_iface

type TemplateFunc = func(resp Response) any

// DefaultTemplateFunc formats the response object before write
func DefaultTemplateFunc(resp Response) any {
	return resp
}
