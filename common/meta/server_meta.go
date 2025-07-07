// package meta contains the code-time setup structures that will be resolved
// into live runtime components when StartServer is called.
package meta

import (
	"lokstra/common/iface"
)

// ServerMeta is a code-defined structure representing a server setup.
type ServerMeta struct {
	name string

	apps     []*AppMeta
	settings map[string]any
	handlers map[string]iface.HandlerFunc
}

var InternalServer = &ServerMeta{
	apps:     []*AppMeta{},
	settings: map[string]any{},
	handlers: make(map[string]iface.HandlerFunc),
}

var serverCreated = false

func NewServer(name string) *ServerMeta {
	if serverCreated {
		panic("server already created, cannot create a new one")
	}
	serverCreated = true
	return InternalServer
}

// AddApp adds a new app to the server.
func (s *ServerMeta) AddApp(app *AppMeta) *ServerMeta {
	s.apps = append(s.apps, app)
	return s
}

// SetSetting sets a configuration setting for the server.
func (s *ServerMeta) SetSetting(key string, value any) *ServerMeta {
	if s.settings == nil {
		s.settings = make(map[string]any)
	}
	s.settings[key] = value
	return s
}

// GetSetting retrieves a configuration setting by key.
func (s *ServerMeta) GetSetting(key string) (any, bool) {
	if s.settings == nil {
		return nil, false
	}
	value, exists := s.settings[key]
	return value, exists
}

func (s *ServerMeta) GetSettings() map[string]any {
	return s.settings
}

func (s *ServerMeta) GetName() string {
	return s.name
}

func (s *ServerMeta) GetApps() []*AppMeta {
	if s.apps == nil {
		return []*AppMeta{}
	}
	return s.apps
}

func (s *ServerMeta) RegisterHandler(name string, handler iface.HandlerFunc) *ServerMeta {
	s.handlers[name] = handler

	return s
}
