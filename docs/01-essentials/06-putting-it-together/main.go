package main

import (
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/router/autogen"
)

func main() {
	// Create service
	todoService := NewTodoService()

	// Create router with auto-generation
	rule := autogen.ConversionRule{
		Convention:     "rest",
		Resource:       "todo",
		ResourcePlural: "todos",
	}

	override := autogen.RouteOverride{}

	autoRouter := autogen.NewFromService(todoService, rule, override)

	// Create app
	app := lokstra.NewApp("todo-api", ":3000", autoRouter)

	log.Println("🚀 Todo API starting on :3000")
	log.Println("📋 API endpoints:")
	log.Println("  POST   /todos")
	log.Println("  GET    /todos")
	log.Println("  GET    /todos/{id}")
	log.Println("  PUT    /todos/{id}")
	log.Println("  DELETE /todos/{id}")
	log.Println()
	log.Println("🛑 Press Ctrl+C to stop")

	// Start server
	if err := app.Run(30 * time.Second); err != nil {
		log.Fatal(err)
	}
}
