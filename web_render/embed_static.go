package web_render

import (
	"embed"
)

//go:embed static/*
var StaticEmbedFS embed.FS
