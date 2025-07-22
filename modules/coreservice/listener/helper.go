package listener

import (
	"time"

	"github.com/primadi/lokstra/core/router"
)

const NETHTTP_LISTENER_NAME = "nethttp"
const FASTHTTP_LISTENER_NAME = "fasthttp"
const SECURE_NETHTTP_LISTENER_NAME = "secure_nethttp"
const HTTP3_LISTENER_NAME = "http3"

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
