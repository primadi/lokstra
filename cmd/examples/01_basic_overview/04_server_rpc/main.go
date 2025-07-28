package main

import (
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/cmd/examples/01_basic_overview/04_server_rpc/hello_service"
)

func main() {
	// Create the greeting service with various return types
	svc := hello_service.NewGreetingService()

	regCtx := lokstra.NewGlobalRegistrationContext()

	// Create new application
	app := lokstra.NewApp(regCtx, "hello-service-app", ":8080")

	// Mount RPC service at /rpc endpoint
	app.MountRpcService("/rpc", svc, false)

	// Create server and app
	svr := lokstra.NewServer(regCtx, "hello-service-server")
	svr.AddApp(app)

	// Start the server
	svr.StartAndWait(10 * time.Second)
}
