package websocketserver

// VersionNumber - Build-time variable
var VersionNumber string

// BuildDate - Build-time variable
var BuildDate string

var WSServerHost string = "localhost"
var WSServerPort string = "8081"
var WSServerAdapterPath string = "/api/v1/adapter"
var WSServerClientPath string = "/api/v1/riden"

var WSSServerAllClientsConnName string = "allClients"

// AdapterMessage holds a riden API message in the form of
// a []byte and the name of the client connection that sent or will
// receive the message. The name string is the unique detail that
// is needed to identify the client when using the WebSocket protocol,
// so that is why it is defined here, rather than in the Adapter
// module.
type AdapterMessage struct {
	ClientConnName string
	MessageBytes   []byte
}

func NewAdapterMessage(name string, msg []byte) AdapterMessage {
	return AdapterMessage{
		ClientConnName: name,
		MessageBytes:   msg,
	}
}
