package listener

import (
	"lokstra/core/router"
	"time"
)

const READ_TIMEOUT_KEY = "read_timeout"
const WRITE_TIMEOUT_KEY = "write_timeout"
const IDLE_TIMEOUT_LEY = "idle_timeout"

const DEFAULT_READ_TIMEOUT = 5 * time.Minute
const DEFAULT_WRITE_TIMEOUT = 5 * time.Minute
const DEFAULT_IDLE_TIMEOUT = 10 * time.Minute

func dumpRoutes(rt router.Router) {
	meta := rt.GetMeta()
	meta.DumpRoutes()
}
