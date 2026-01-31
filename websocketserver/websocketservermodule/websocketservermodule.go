package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"riden/logger"
	wss "riden/websocketserver"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Logger Handles all log writing for the WebSocketServer
var Logger logger.Logger

var LogDirectory string

// Client holds the details of a client's WebSocket connection. Close
// is used to signal that the connection is closing soon and no new message
// operations should occur on the connection.
type Client struct {
	WSConn *websocket.Conn
	Close  chan struct{}
	Write  chan wss.AdapterMessage
}

func (c *Client) RemoteConnString() string {
	return c.WSConn.RemoteAddr().String()
}

// Initialize sets up a client's channels and handlers
func (c *Client) Initialize() {
	c.Close = make(chan struct{})
	c.Write = make(chan wss.AdapterMessage, ChannelBufferSize)
}

func (c *Client) CleanUpAfterReadLoop() {
	Logger.Info().Msgf("Closed connection to remote address: %s", c.RemoteConnString())
	// Send a close signal to write loop
	close(c.Close)
	close(c.Write)
	c.WSConn.Close()
}

const ChannelBufferSize int = 1024

// WriteControlDeadline is a duration of time to be added to the time of the
// control message write operation
var WriteControlDeadline time.Duration = 5 * time.Second

// safeClients stores the safe client clonnection details in
// a [string]*Client map
var safeClients sync.Map

// AdapterConnection holds a special instance of a Client for the adapter
// client to communicate with the WebSocket server. This is the only connection
// between the WebSocket server and the adapter.
type AdapterConnection struct {
	mux sync.RWMutex
	Client
}

func (tc *AdapterConnection) Initialize(wsConn *websocket.Conn) {
	tc.mux.Lock()
	defer tc.mux.Unlock()

	tc.Client = Client{
		WSConn: wsConn,
	}
	tc.Client.Initialize()
}

func (tc *AdapterConnection) IsConnectionSet() bool {
	tc.mux.RLock()
	defer tc.mux.RUnlock()
	return tc.WSConn != nil
}

func (tc *AdapterConnection) ClearConnection() {
	tc.mux.Lock()
	defer tc.mux.Unlock()

	tc.Client = Client{
		WSConn: nil,
	}
}

func (tc *AdapterConnection) CleanUpAfterReadLoop() {
	tc.mux.Lock()
	defer tc.mux.Unlock()

	tc.Client.CleanUpAfterReadLoop()
}

func (tc *AdapterConnection) RemoteConnString() string {
	tc.mux.RLock()
	defer tc.mux.RUnlock()

	return tc.WSConn.RemoteAddr().String()
}

var AdapterConn AdapterConnection

type clientWebSocketHandler struct {
	upgrader websocket.Upgrader
}

// ServeHTTP handles an incoming connection from a client and upgrades the connection
// to a WebSocket connection. ServeHTTP() is launched in its own goroutine by the
// listener and then the message reader loop is called here, so that it continues to
// run in the same goroutine. If the message reader loop encounters an error, it will
// return here and the connection will be closed. This function also launches a
// message writer loop in its own goroutine. The gorilla websocket documentation
// states that, "Connections support one concurrent reader and one concurrent writer.
// Applications are responsible for ensuring that no more than one goroutine
// calls the write methods (NextWriter, SetWriteDeadline, WriteMessage, WriteJSON,
// EnableWriteCompression, SetCompressionLevel) concurrently and that no more than one
// goroutine calls the read methods (NextReader, SetReadDeadline, ReadMessage, ReadJSON,
// SetPongHandler, SetPingHandler) concurrently." So no other goroutines besides these
// two should be reading or writing to the websocket.Conn that is returned from
// Upgrade().
func (wsh clientWebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var c *websocket.Conn

	if !AdapterConn.IsConnectionSet() {
		// The adapter is not connected, so we will not upgrade this connection
		Logger.Info().Msg("Adapter is not connected.")
		Logger.Info().Msgf("Not upgrading this connection attempt from remote address: %s",
			r.RemoteAddr)
		ReturnError(w, http.StatusInternalServerError)
		return
	}

	c, err = wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger.Error().Msgf("error %s when upgrading connection to websocket", err.Error())
		return
	}
	Logger.Info().Msgf("Client connected from remote address: %s", c.RemoteAddr().String())

	clientConn := Client{
		WSConn: c,
	}
	safeClients.Store(clientConn.RemoteConnString(), &clientConn)

	clientConn.Initialize()

	// Launch the message writer loop that will close when it receives a close signal
	go ClientWriteLoop(&clientConn)

	// Call the message reader loop that returns here if the loop is broken
	err = ClientReadLoop(&clientConn)
	Logger.Info().Msgf("ClientReadLoop returned for remote address: %s, with err: %v",
		clientConn.RemoteConnString(), err)
	clientConn.CleanUpAfterReadLoop()
	safeClients.Delete(clientConn.RemoteConnString())
}

type adapterWebSocketHandler struct {
	upgrader websocket.Upgrader
}

// ServeHTTP handles an incoming connection from the adapter and upgrades
// the connection to a WebSocket connection. ServeHTTP() is launched in its own
// goroutine by the listener and then the message reader loop is called here,
// so that it continues to run in the same goroutine. If the message reader loop
// encounters an error, it will return here and the connection will be closed. This
// function also launches a message writer loop in its own goroutine. The gorilla
// websocket documentation states that, "Connections support one concurrent reader
// and one concurrent writer. Applications are responsible for ensuring that no more
// than one goroutine calls the write methods (NextWriter, SetWriteDeadline,
// WriteMessage, WriteJSON, EnableWriteCompression, SetCompressionLevel) concurrently
// and that no more than one goroutine calls the read methods (NextReader,
// SetReadDeadline, ReadMessage, ReadJSON, SetPongHandler, SetPingHandler) concurrently."
// So no other goroutines besides these two should be reading or writing to the
// websocket.Conn that is returned from Upgrade().
func (wsh adapterWebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var c *websocket.Conn

	if AdapterConn.IsConnectionSet() {
		// There is already a connection to the adapter, so we will not upgrade this
		// connection attempt. Only one connection from the adapter is allowed.
		Logger.Info().Msgf("Adapter is already connected from remote address: %s",
			AdapterConn.RemoteConnString())
		Logger.Info().Msgf("Not upgrading this connection attempt from remote address: %s",
			r.RemoteAddr)
		ReturnError(w, http.StatusTooManyRequests)
		return
	}

	c, err = wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger.Error().Msgf("error %s when upgrading adapter connection to websocket", err.Error())
		return
	}
	Logger.Info().Msgf("Adapter connected from remote address: %s", c.RemoteAddr().String())

	AdapterConn.Initialize(c)

	// Launch the message writer loop that will close when it receives a close signal
	go AdapterWriteLoop(&AdapterConn.Client)

	// Call the message reader loop that returns here if the loop is broken
	err = AdapterReadLoop(&AdapterConn.Client)
	Logger.Info().Msgf("AdapterReadLoop returned for remote address: %s, with err: %s",
		AdapterConn.RemoteConnString(), err.Error())
	AdapterConn.CleanUpAfterReadLoop()

	// Clear client connection
	AdapterConn.ClearConnection()
}

// ReturnError returns the given status code and reason on the given ResponseWriter
func ReturnError(w http.ResponseWriter, status int) {
	w.Header().Set("Sec-Websocket-Version", "13")
	http.Error(w, http.StatusText(status), status)
}

func Usage() {
	fmt.Println("Usage:", os.Args[0], "log_dir log_level host port")
	os.Exit(1) // 1 - Non-zero exit code indicates an error
}

func main() {
	// Parse the arguments
	flag.Parse()

	if len(flag.Args()) != 4 {
		Usage()
	}

	LogDirectory = flag.Arg(0)

	fmt.Printf("Log Directory: %s\n", LogDirectory)

	logFile := path.Join(LogDirectory, "websocketserver.log")

	// Startup procedures
	var err error
	Logger, err = logger.InitializeLogger(logFile, flag.Arg(1))
	if err != nil {
		fmt.Println("Error opening log file:", logFile, ":", err)
		os.Exit(1)
	}

	clientWebSocketHandler := clientWebSocketHandler{
		upgrader: websocket.Upgrader{},
	}
	adapterWebSocketHandler := adapterWebSocketHandler{
		upgrader: websocket.Upgrader{},
	}
	http.Handle("/api/v1/riden", clientWebSocketHandler)
	http.Handle("/api/v1/adapter", adapterWebSocketHandler)
	Logger.Info().Msgf("WebSocketServer Version: %s Build Date: %s",
		wss.VersionNumber, wss.BuildDate)
	Logger.Info().Msg("Starting websocket server...")
	Logger.Info().Msgf("Listening at %s:%s", flag.Arg(2), flag.Arg(3))
	err = http.ListenAndServe(fmt.Sprintf("%s:%s", flag.Arg(2), flag.Arg(3)), nil)
	Logger.Info().Msgf("ListenAndServe returned: %s", err.Error())
}
