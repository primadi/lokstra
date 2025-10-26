package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Handler yang panic
func panicHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handler called - about to panic!")
	panic("BOOM! Something went wrong")
}

// Handler normal
func normalHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Normal handler called")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok", "message": "Server is still working!"}`))
}

func main() {
	fmt.Println("=== Go HTTP Server Panic Behavior Demo ===")
	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("")
	fmt.Println("Test endpoints:")
	fmt.Println("  GET http://localhost:8080/panic   - Will panic (no response)")
	fmt.Println("  GET http://localhost:8080/normal  - Normal response")
	fmt.Println("")
	fmt.Println("Try this:")
	fmt.Println("  1. curl http://localhost:8080/panic")
	fmt.Println("     → No response, connection closed")
	fmt.Println("  2. curl http://localhost:8080/normal")
	fmt.Println("     → Server still works!")
	fmt.Println("")

	http.HandleFunc("/panic", panicHandler)
	http.HandleFunc("/normal", normalHandler)

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Starting server...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
