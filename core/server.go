package core

const CERT_FILE = "cert_file"
const KEY_FILE = "key_file"

type Server struct {
	name string

	apps     []*App
	settings map[string]any
}

var server = &Server{
	apps:     []*App{},
	settings: make(map[string]any),
}

// Server is a singleton instance of the server.
func GetServer() *Server {
	return server
}

func (s *Server) GetName() string {
	return s.name
}

func (s *Server) SetName(name string) {
	s.name = name
}

func (s *Server) AddApp(app *App) {
	s.apps = append(s.apps, app)
}

// SetSetting sets a configuration setting for the server.
func (s *Server) SetSetting(key string, value any) {
	if s.settings == nil {
		s.settings = make(map[string]any)
	}
	s.settings[key] = value
}

// GetSetting retrieves a configuration setting by key.
func (s *Server) GetSetting(key string) (any, bool) {
	if s.settings == nil {
		return nil, false
	}
	value, exists := s.settings[key]
	return value, exists
}

func (s *Server) Start() error {
	// var wg sync.WaitGroup
	// errCh := make(chan error, len(s.apps))

	// for _, app := range s.apps {
	// 	wg.Add(1)
	// 	go func(a *App) {
	// 		defer wg.Done()
	// 		if err := a.Start(); err != nil {
	// 			errCh <- fmt.Errorf("app '%s' failed: %w", app.GetName(), err)
	// 		}
	// 	}(app)
	// }

	// wg.Wait()
	// close(errCh)

	// if len(errCh) > 0 {
	// 	return <-errCh
	// }
	return nil
}
