package core

import "lokstra/core/router"

type App struct {
	router.Router

	name string
	port int
}

func NewApp(name string, port int) *App {
	return &App{
		name: name,
		port: port,
		// Router: router.NewRouter(),
	}
}
