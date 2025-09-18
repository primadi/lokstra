package midware

type Execution struct {
	Name   string
	Config any // Configuration for the middleware

	MiddlewareFn   Func // Function to create the middleware
	Priority       int  // Lower number means higher priority (1-100)
	ExecutionOrder int  // for internal use. Order of execution, lower number means earlier execution
}

func NewExecution(fn Func) *Execution {
	return &Execution{MiddlewareFn: fn, Priority: 5000}
}
