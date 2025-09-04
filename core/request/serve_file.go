package request

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/common/file_system"
)

func ServeFile(path any) HandlerFunc {
	var fileSystem http.FileSystem

	switch v := path.(type) {
	case string:
		fileSystem = http.Dir(v)
	case http.Dir:
		fileSystem = v
	case file_system.FSWithSubFS:
		subFs, err := fs.Sub(v.FS, v.SubFS)
		if err != nil {
			panic(err)
		}
		fileSystem = http.FS(subFs)
	case embed.FS:
		sub, err := file_system.SubFirstCommonFolder(v)
		if err != nil {
			panic(err)
		}
		fileSystem = http.FS(sub)
	case fs.FS:
		fileSystem = http.FS(v)
	default:
		panic("ServeFile: unsupported path type")
	}

	return func(ctx *Context) error {
		p := ctx.Request.URL.Path
		idx := strings.Index(p[1:], "/")
		if idx == -1 {
			http.FileServer(fileSystem).ServeHTTP(ctx.Writer, ctx.Request)
		} else {
			prefix := p[:idx+2]
			http.StripPrefix(prefix, http.FileServer(fileSystem)).
				ServeHTTP(ctx.Writer, ctx.Request)
		}
		return nil
	}
}
