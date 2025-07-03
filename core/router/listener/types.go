package listener

type ListenerType = string

const (
	NetHttpListenerType       ListenerType = "net/http"
	FastHttpListenerType      ListenerType = "fasthttp"
	SecureNetHttpListenerType ListenerType = "secure_net/http"
)
