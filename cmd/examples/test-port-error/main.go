package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/primadi/lokstra/core/app/listener"
)

func main() {
	// Create first listener on port 8090
	config1 := map[string]any{
		"addr": ":8090",
	}
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("App 1"))
	})

	listener1 := listener.NewNetHttp(config1, handler1)

	// Start first listener in background
	go func() {
		fmt.Println("Starting first listener on :8090...")
		if err := listener1.ListenAndServe(); err != nil {
			fmt.Printf("Listener 1 error: %v\n", err)
		}
	}()

	// Wait for first listener to start
	fmt.Println("Waiting for first listener to start...")
	time.Sleep(1 * time.Second)

	// Now try to create second listener on the same port
	fmt.Println("\nTrying to start second listener on same port :8090...")
	config2 := map[string]any{
		"addr": ":8090",
	}
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("App 2"))
	})

	listener2 := listener.NewNetHttp(config2, handler2)

	// This should fail with our nice error message
	fmt.Println("\n=== Attempting to bind to already-used port ===")
	if err := listener2.ListenAndServe(); err != nil {
		fmt.Printf("\n%v\n", err)
		fmt.Println("==========================================")
	}
}
