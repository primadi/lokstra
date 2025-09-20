package main

import (
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/static_files"
)

func createHtmxAndStaticRoutes(app *lokstra.App) {
	sf := static_files.EmbedFS(htmxFS, "htmx_app")

	// Mount HTMX pages at root
	app.MountHtmx("/", nil, sf.Sources...)

	// Static files (CSS, JS, images, etc)
	sfStatic := static_files.EmbedFS(htmxFS, "htmx_app/static")
	// Mount static assets
	app.MountStatic("/static/", false, sfStatic.Sources...)
}
