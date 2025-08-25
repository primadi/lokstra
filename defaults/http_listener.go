package defaults

import (
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/modules/coreservice/listener"
	"github.com/primadi/lokstra/serviceapi"
)

const (
	HTTP_LISTENER_FASTHTTP       = serviceapi.HTTP_LISTENER_PREFIX + "fast_http"
	HTTP_LISTENER_NETHTTP        = serviceapi.HTTP_LISTENER_PREFIX + "net_http"
	HTTP_LISTENER_SECURE_NETHTTP = serviceapi.HTTP_LISTENER_PREFIX + "secure_net_http"
	HTTP_LISTENER_HTTP3          = serviceapi.HTTP_LISTENER_PREFIX + "http3"
)

func RegisterAllHTTPListeners(regCtx registration.Context) {
	regCtx.RegisterServiceFactory(HTTP_LISTENER_FASTHTTP,
		listener.NewFastHttpListener)
	regCtx.RegisterServiceFactory(HTTP_LISTENER_NETHTTP,
		listener.NewNetHttpListener)
	regCtx.RegisterServiceFactory(HTTP_LISTENER_SECURE_NETHTTP,
		listener.NewSecureNetHttpListener)
	regCtx.RegisterServiceFactory(HTTP_LISTENER_HTTP3,
		listener.NewHttp3Listener)

	// Register default listener
	regCtx.RegisterServiceFactory(serviceapi.HTTP_LISTENER_PREFIX+"default",
		listener.NewNetHttpListener)
}
