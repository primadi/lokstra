package lokstra

import "lokstra/common/meta"

type ServerMeta = meta.ServerMeta
type AppMeta = meta.AppMeta
type RouterMeta = meta.RouterMeta

func NewServerMeta(name string) *ServerMeta {
	return meta.NewServer(name)
}

func NewAppMeta(name string, port int) *AppMeta {
	return meta.NewApp(name, port)
}

func NewRouter() *RouterMeta {
	return meta.NewRouter()
}
