package recovery

import "github.com/primadi/lokstra/core/request"

// Called when a panic is recovered
var OnRecover func(ctx *request.Context, err any, stack string)

func SetRecoverHook(hook func(ctx *request.Context, err any, stack string)) {
	OnRecover = hook
}
