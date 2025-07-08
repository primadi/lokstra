package listener

import "lokstra/core/router"

func dumpRoutes(rt router.Router) {
	meta := rt.GetMeta()
	meta.DumpRoutes()
}
