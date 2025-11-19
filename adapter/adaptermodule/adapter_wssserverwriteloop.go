package main

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

// WSServerWriteLoop waits for a message to appear on the WebSocketChannel channel, then sends that
// message to the WebSocket server to be forwarded to the API clients. If a signal consisting of
// any integer value is received on the Close channel, the loop will log a message and return.
func WSServerWriteLoop(wsServer *WebSocketServerConnnection) error {
	var err error

	Logger.Info().Msgf("Entered WSServerWriteLoop for remote address: %s",
		wsServer.RemoteConnString())
	ping := time.Tick(60 * time.Second)
	for {
		select {
		case <-ping:
			Logger.Info().Msg("Pinging WebSocketServer")
			wsServer.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(WriteControlDeadline))
			wsServer.KeepAliveTimer = time.AfterFunc(time.Duration(59*time.Second),
				InitializeWebSocketServerConn)
		case closeSignal := <-wsServer.Close:
			Logger.Info().Msgf("WebSocket server write loop has received a close signal, %d",
				closeSignal)
			// handle close
			return err
		case adapterMsg := <-wsServer.Write:
			// %q used to escape untrusted user input
			Logger.Info().Msgf("Writing message, %q, to WebSocket server at %s",
				string(adapterMsg.MessageBytes), wsServer.RemoteConnString())
			msg, err := json.Marshal(adapterMsg)
			if err != nil {
				Logger.Error().Msgf("Error marshaling delivery failure data slice: %s", err.Error())
				break
			}
			err = wsServer.Conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				Logger.Error().Msgf("Error %s when writing message to client", err)
			}
		}
	}
}
