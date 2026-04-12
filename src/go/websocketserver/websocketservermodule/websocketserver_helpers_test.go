package main

import (
	"net/http"
	wss "riden/websocketserver"

	"github.com/gorilla/websocket"
)

type mockAdapterReceiveHandler struct {
	upgrader websocket.Upgrader
	Message  chan wss.AdapterMessage
}

func (wsh mockAdapterReceiveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var c *websocket.Conn

	c, err = wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger.Error().Msgf("error when upgrading mock adapter connection to websocket: %s", err.Error())
		return
	}
	Logger.Info().Msgf("Mock adapter connected at remote address: %s", c.RemoteAddr().String())

	AdapterConn.Initialize(c)

ForLoop:
	for {
		select {
		case <-AdapterConn.Close:
			// handle close
			break ForLoop
		case adapterMsg := <-AdapterConn.Write:
			wsh.Message <- adapterMsg
		}
	}

	// Cleanup
	AdapterConn.ClearConnection()
}

type mockClientReceiveHandler struct {
	upgrader websocket.Upgrader
	Message  chan wss.AdapterMessage
}

func (wsh mockClientReceiveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var c *websocket.Conn

	c, err = wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger.Error().Msgf("error when upgrading mock client connection to websocket: %s", err.Error())
		return
	}
	Logger.Info().Msgf("Mock client connected from remote address: %s", c.RemoteAddr().String())

	clientConn := Client{
		WSConn: c,
	}
	safeClients.Store(clientConn.RemoteConnString(), &clientConn)

	clientConn.Initialize()

ForLoop:
	for {
		select {
		case <-clientConn.Close:
			// handle close
			break ForLoop
		case adapterMsg := <-clientConn.Write:
			wsh.Message <- adapterMsg
		}
	}

	// Clean up
	safeClients.Delete(clientConn.RemoteConnString())
}
