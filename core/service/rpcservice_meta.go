package service

type RpcServiceMeta struct {
	MethodParam string // default "method"
	ServiceName string
	ServiceInst Service
}
