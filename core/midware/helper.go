package midware

import "github.com/primadi/lokstra/core/request"

type Func = func(next request.HandlerFunc) request.HandlerFunc
type Factory = func(config any) Func

func Named(name string, config ...any) *Execution {
	var cfg any
	switch len(config) {
	case 0:
		cfg = nil // No config provided, use nil
	case 1:
		cfg = config[0] // Single config provided
	default:
		cfg = config // Multiple configs provided, use as is
	}
	return &Execution{Name: name, Config: cfg, Priority: 5000}
}
