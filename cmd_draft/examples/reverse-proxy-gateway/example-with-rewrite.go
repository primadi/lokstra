package main

// Example 3: Reverse proxy with path rewrite
// func example3_WithRewrite() {
// 	r := lokstra.NewRouter("api-gateway")

// 	// Example 1: Rewrite /api/v1/* to /v2/*
// 	// /api/v1/users -> http://localhost:9000/v2/users
// 	rewrite1 := &lokstra_handler.ReverseProxyRewrite{
// 		From: "^/v1", // regex pattern
// 		To:   "/v2",  // replacement
// 	}
// 	r.ANYPrefix("/api", lokstra_handler.MountReverseProxy("/api", "http://localhost:9000", rewrite1))

// 	// Example 2: Rewrite /old/* to /new/*
// 	// /old/path -> http://localhost:9001/new/path
// 	rewrite2 := &lokstra_handler.ReverseProxyRewrite{
// 		From: "^/old",
// 		To:   "/new",
// 	}
// 	r.ANYPrefix("/old", lokstra_handler.MountReverseProxy("", "http://localhost:9001", rewrite2))

// 	// Example 3: Strip prefix and rewrite
// 	// /legacy/api/users -> http://localhost:9002/v2/users
// 	rewrite3 := &lokstra_handler.ReverseProxyRewrite{
// 		From: "^/api", // after stripping /legacy, path is /api/users
// 		To:   "/v2",   // becomes /v2/users
// 	}
// 	r.ANYPrefix("/legacy", lokstra_handler.MountReverseProxy("/legacy", "http://localhost:9002", rewrite3))

// 	app := lokstra.NewApp("gateway", ":8080", r)

// 	log.Println("Starting reverse proxy with path rewrite on :8080")
// 	log.Println("  /api/v1/* -> http://localhost:9000/v2/*")
// 	log.Println("  /old/*    -> http://localhost:9001/new/*")
// 	log.Println("  /legacy/api/* -> http://localhost:9002/v2/*")

// 	if err := app.Run(5 * time.Second); err != nil {
// 		log.Fatal(err)
// 	}
// }

// Uncomment to run this example
// func main() {
// 	example3_WithRewrite()
// }
