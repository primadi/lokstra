package request

type HandlerFunc = func(ctx *Context) error
type HandlerFuncWithParam[T any] = func(ctx *Context, params *T) error

type HandlerRegister struct {
	Name        string
	HandlerFunc HandlerFunc
}
