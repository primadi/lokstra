package main

import (
	"log"
	"net/http"
)

func main() {
	// Initialize the router with a name
	router := setupRouter()

	// Start the HTTP server
	log.Println("Starting server on :3000")

	router.PrintRoutes()
	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
