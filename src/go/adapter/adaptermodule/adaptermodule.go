package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	a "riden/adapter"
	"riden/logger"
	pb "riden/proto"
	wss "riden/websocketserver"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

// Logger Handles all log writing for the Adapter
var Logger logger.Logger

var LogDirectory string

// WriteControlDeadline is a duration of time to be added to the time of the
// WebSocket control message write operation
var WriteControlDeadline time.Duration = 5 * time.Second

var WSServerURL string = "ws://" + wss.WSServerHost + ":" + wss.WSServerPort + wss.WSServerAdapterPath

const WSChannelBufferSize int = 32

type WebSocketServerConnnection struct {
	Conn  *websocket.Conn
	Close chan struct{}
	Write chan wss.AdapterMessage
	// RetryStatusCodes contains the list of status codes to retry,
	// use "x" as a wildcard for a single digit (default: [500])
	RetryStatusCodes []string
	KeepAliveTimer   *time.Timer
}

func (ws *WebSocketServerConnnection) RemoteConnString() string {
	return ws.Conn.RemoteAddr().String()
}

// Initialize sets up a client's channels and handlers
func (ws *WebSocketServerConnnection) Initialize() {
	ws.Close = make(chan struct{})
	ws.Write = make(chan wss.AdapterMessage, WSChannelBufferSize)

	ws.Conn.SetPongHandler(func(msg string) error {
		ws.StopKeepAliveTimer()
		return nil
	})

	go WSServerReadLoop(ws)
	go WSServerWriteLoop(ws)
}

func (ws *WebSocketServerConnnection) Reset() {
	Logger.Info().Msg("Resetting WebSocketServer connection")
	close(ws.Close)
	close(ws.Write)

	if ws.Conn != nil {
		// Send a close message with 1001 status code
		msg := websocket.FormatCloseMessage(websocket.CloseGoingAway, "")
		Logger.Info().Msgf("Attempting to close remote connection: %s", ws.RemoteConnString())
		ws.Conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))
		ws.Conn.Close()
		ws.Conn = nil
	}
}

func (ws *WebSocketServerConnnection) StopKeepAliveTimer() {
	if ws.KeepAliveTimer != nil {
		ws.KeepAliveTimer.Stop()
	}
}

var WebSocketServerConn WebSocketServerConnnection

var WebSocketDialTimer *time.Timer

var GRPCServer *grpc.Server

type MockLogicMessage struct {
	ConnName        string
	ConnType        string
	MessageType     string
	APIMessageBytes []byte
}

func NewMockLogicMessage(connName, connType string,
	msgType string, msgBytes []byte) MockLogicMessage {
	return MockLogicMessage{
		ConnName:        connName,
		ConnType:        connType,
		MessageType:     msgType,
		APIMessageBytes: msgBytes,
	}
}

// MessageType is used with json.Unmarshal() to decode the
// MessageType from a riden API defined message
type MessageType string

func InitializeConnections() {
	// TODO: Define handshake messages between the MockLogic and the Adapter
	// and only proceed to initiating the WebSocketServer connection if the
	// MockLogic has completed the handshake
	// Connect to MockLogic through gRPC first, then create WebSocketServer
	// connection
	go InitializeGRPCServer()

	// Initialize WebSocketServer connection
	InitializeWebSocketServerConn()
}

func InitializeWebSocketServerConn() {
	if WebSocketServerConn.Conn != nil {
		WebSocketServerConn.Reset()
	}
	WebSocketServerConn = WebSocketServerConnnection{
		RetryStatusCodes: []string{"500"},
	}
	DialWebsocketServer()
}

func InitializeGRPCServer() {
	// Start server
	listener, err := net.Listen("tcp", a.GRPCServerAddress)
	if err != nil {
		Logger.Error().Msgf("Error: failed to listen: %s", err.Error())
		return
	}
	s := grpc.NewServer()
	GRPCServer = s
	pb.RegisterAdapterServer(s, &adapterServer{})
	Logger.Info().Msgf("gRPC server listening at %v", listener.Addr())

	// Serve blocks until the process is killed or the server is stopped.
	if err := s.Serve(listener); err != nil {
		Logger.Error().Msgf("Error: Failed to serve: %s", err.Error())
	}
}

func DialWebsocketServer() {
	var err error
	ws, response, err := websocket.DefaultDialer.Dial(WSServerURL, nil)
	if err != nil {
		Logger.Error().Msgf("Error dialing WebSocketServer: %s", err.Error())
		WebSocketDialTimer = time.AfterFunc(time.Duration(2*time.Second), DialWebsocketServer)
		return
	}

	if response.StatusCode != http.StatusSwitchingProtocols {
		bodyContents, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			Logger.Error().Msgf("Error reading response body after receiving status, %d, : %s",
				response.StatusCode, err.Error())
		}
		Logger.Error().Msgf("Could not upgrade WebSocketServer connection to WebSocket; Received http status, %d, with body: %s",
			response.StatusCode, string(bodyContents))
		// Retry only if status matches RetryStatusCodes
		ok, err := IsRetryCodeMatch(response.StatusCode, WebSocketServerConn.RetryStatusCodes)
		if err != nil {
			Logger.Error().Msgf("Error matching retry status codes after receiving status, %d, : %s",
				response.StatusCode, err.Error())
		}
		if ok {
			WebSocketDialTimer = time.AfterFunc(time.Duration(2*time.Second), DialWebsocketServer)
			return
		}

		return
	}

	if WebSocketDialTimer != nil {
		WebSocketDialTimer.Stop()
	}
	Logger.Info().Msg("WebSocketServer connection created successfully")
	WebSocketServerConn.Conn = ws
	WebSocketServerConn.Initialize()
}

func IsRetryCodeMatch(returnedCode int, retryStatusCodes []string) (bool, error) {
	var matched bool
	var err error
	// Check if the status code matches any of the retry status codes
	returnedCodeStr := strconv.Itoa(returnedCode)
	for _, retryCode := range retryStatusCodes {
		if len(retryCode) != 3 {
			err = fmt.Errorf("invalid retry status code: %s", retryCode)
			return matched, err
		}
		matched := true
		for i := range retryCode {
			c := retryCode[i]
			if c == 'x' {
				continue
			}
			if c != returnedCodeStr[i] {
				matched = false
				break
			}
		}
		if matched {
			return matched, err
		}
	}

	return matched, err
}

func Usage() {
	fmt.Println("Usage:", os.Args[0], "log_dir log_level")
	os.Exit(1) // 1 - Non-zero exit code indicates an error
}

func main() {
	// Parse the arguments
	flag.Parse()

	if len(flag.Args()) != 2 {
		Usage()
	}

	LogDirectory = flag.Arg(0)

	fmt.Printf("Log Directory: %s\n", LogDirectory)

	logFile := path.Join(LogDirectory, "adapter.log")

	// Startup procedures
	var err error
	Logger, err = logger.InitializeLogger(logFile, flag.Arg(1))
	if err != nil {
		fmt.Println("Error opening log file:", logFile, ":", err.Error())
		os.Exit(1)
	}

	// Make channels to pass messages to the gRPC bi-directional streams
	GRPCChans.MakeReserveTrip()
	GRPCChans.MakeAtDock()
	GRPCChans.MakeOnBoat()
	GRPCChans.MakeOffBoat()

	// Initialize connections to the necessary servers
	InitializeConnections()

	// Block forever
	select {}
}
