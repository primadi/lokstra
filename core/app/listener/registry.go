package listener

import (
	"net/http"

	"github.com/primadi/lokstra/common/utils"
)

type AppListenerFactory func(config map[string]any, handler http.Handler) AppListener

var listenerRegistry = make(map[string]AppListenerFactory)

// RegisterListener registers a new AppListener factory with a name.
func RegisterListener(name string, factory AppListenerFactory) {
	listenerRegistry[name] = factory
}

// CreateListener creates a new AppListener based on the listener type.
func CreateListener(config map[string]any, handler http.Handler) AppListener {
	listenerType := utils.GetValueFromMap(config, "listener-type", "default")
	if factory, exists := listenerRegistry[listenerType]; exists {
		return factory(config, handler)
	}
	panic("Unsupported listener type: " + listenerType)
}

func init() {
	RegisterListener("nethttp", NewNetHttp)
	RegisterListener("default", NewNetHttp)
	RegisterListener("http3", NewHttp3)
	RegisterListener("fasthttp", NewFastHttp)
}
