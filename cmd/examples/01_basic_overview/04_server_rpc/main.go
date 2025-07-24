package main

import (
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/cmd/examples/01_basic_overview/04_server_rpc/hello_service"
)

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()

	svc := hello_service.NewGreetingService("greeting_service")
	svr := lokstra.NewServer(regCtx, "hello-service-server")
	app := svr.NewApp("hello-service-app", ":8080")

	app.MountRpcService("/rpc", svc, false)

	svr.StartAndWait(10 * time.Second)
}
