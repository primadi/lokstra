package proxy

type RemoteService struct {
	Name    string
	BaseUrl string
	Timeout int // in seconds
}

func NewRemoteService(name string) *RemoteService {
	return &RemoteService{
		Name:    name,
		Timeout: 10, // default timeout
	}
}

func CallService[T any](rs *RemoteService, method string, params any) (T, error) {
	var zaro T
	// Implement the logic to call the remote service here.
	// This is a placeholder implementation.
	return zaro, nil
}
