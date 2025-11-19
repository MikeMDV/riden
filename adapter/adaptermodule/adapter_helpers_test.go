package main

import (
	"encoding/json"
	"net/http"
	wss "riden/websocketserver"
	"time"

	"github.com/gorilla/websocket"
)

type mockWebSocketReceiveHandler struct {
	upgrader websocket.Upgrader
	Message  chan wss.AdapterMessage
}

func (wsh mockWebSocketReceiveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var c *websocket.Conn

	c, err = wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger.Error().Msgf("error when upgrading mock translation layer connection to websocket: %s", err.Error())
		return
	}
	Logger.Info().Msgf("Mock WebSocket Server connected at remote address: %s", c.RemoteAddr().String())

	_, message, err := c.ReadMessage()
	if err != nil {
		Logger.Error().Msgf("Error when reading message from adapter: %s", err.Error())
		// Cleanup
		c = nil
		return
	}

	var adapterMsg wss.AdapterMessage
	err = json.Unmarshal(message, &adapterMsg)
	if err != nil {
		Logger.Error().Msgf("Error unmarshalling JSON: %s", err.Error())
	}

	wsh.Message <- adapterMsg

	// Cleanup
	c = nil
}

// mockWebSocketReserveSendHandler receives any message and replies with a Reserve message
type mockWebSocketReserveSendHandler struct {
	upgrader websocket.Upgrader
}

func (wsh mockWebSocketReserveSendHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var c *websocket.Conn

	c, err = wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger.Error().Msgf("error when upgrading mock translation layer connection to websocket: %s", err.Error())
		return
	}
	Logger.Info().Msgf("Mock WebSocket Server connected at remote address: %s", c.RemoteAddr().String())

	msgType, _, err := c.ReadMessage()
	if err != nil {
		Logger.Error().Msgf("Error when reading message from adapter: %s", err.Error())
		// Cleanup
		c = nil
		return
	}

	Logger.Info().Msgf("Received msgType, %d, from address: %s", msgType, c.RemoteAddr().String())

	// Create a Reserve message to send to the adapter
	adapterMsg := wss.NewAdapterMessage(testClientConnectionName, testReserveTripAPIMessageBytes)
	adapterMsgBytes, err := json.Marshal(adapterMsg)
	if err != nil {
		Logger.Error().Msgf("Error marshaling delivery failure data slice: %s", err.Error())
	}
	err = c.WriteMessage(websocket.TextMessage, adapterMsgBytes)
	if err != nil {
		Logger.Error().Msgf("Error when writing message to client: %s", err.Error())
	}

	time.Sleep(1 * time.Second)

	// Cleanup
	c = nil
}
