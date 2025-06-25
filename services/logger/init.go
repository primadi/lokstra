package logger

import "lokstra/core"

func init() {
	core.RegisterNamedService("logger", LoggerFactory)
}
