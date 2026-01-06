package eventbus

import (
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
)

const SERVICE_TYPE = "map-event-bus"

func Service() serviceapi.EventBus {
	return NewBus()
}

func ServiceFactory(_ map[string]any) any {
	return NewBus()
}

func Register() {
	lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory)
}
