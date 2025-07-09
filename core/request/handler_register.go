package request

type HandlerFunc = func(ctx *Context) error

type HandlerRegister struct {
	Name        string
	HandlerFunc HandlerFunc
}
