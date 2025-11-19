package main

import (
	"encoding/json"
	a "riden/adapter"
	wss "riden/websocketserver"
	"time"

	"github.com/gorilla/websocket"
)

// WSServerReadLoop reads messages from the given WebSocketServer, unmarshals
// the message to determine its message type, and calls the appropriate
// function to process the message
func WSServerReadLoop(wsServer *WebSocketServerConnnection) error {
	Logger.Info().Msgf("Entered WSServerReadLoop for remote address: %s",
		wsServer.RemoteConnString())

	for {
		msgType, message, err := wsServer.Conn.ReadMessage()
		if err != nil {
			Logger.Error().Msgf("Error %s when reading message from WebSocketServer", err.Error())
			return err
		}
		// Reject binary type messages and send close control message
		if msgType == websocket.BinaryMessage {
			Logger.Info().Msgf("Received binary message type from WebSocketServer: %s, Sending close message with code 1003",
				wsServer.RemoteConnString())
			msg := websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Binary data not supported")
			wsServer.Conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))
			// Continue here and wait for close control message reply before exiting the read loop
			// The peer should respond with a close message which will cause ReadMessage()
			// to return a CloseError and the read loop to exit.
			// If the peer does not respond with a close control message, but continues to send
			// binary type messages, we will continue to send close control messages.
			// If the WebSocketServer begins to send text type messages we will be able to receive them
			// and process them.
			continue
		}
		// %q used to escape untrusted user input
		Logger.Info().Msgf("Received message from WebSocketServer connection %s: %q", wsServer.RemoteConnString(), string(message))

		messageType, adapterMessage, err := GetMessageFromWebSocketMsg(message)
		if err != nil {
			continue
		}

		// Process this message
		mlMsg := NewMockLogicMessage(adapterMessage.ClientConnName,
			a.ConnectionTypeWebSocket, string(messageType), adapterMessage.MessageBytes)

		go ProcessMessageToMockLogic(&mlMsg)
	}
}

// GetMessageFromWebSocketMsg retrieves the message type string from a riden
// API message received through a WebSocket connection.
func GetMessageFromWebSocketMsg(wssMsg []byte) (MessageType, wss.AdapterMessage, error) {
	var err error
	var messageType MessageType

	// Unmarshal message to get message type
	var adapterMsg wss.AdapterMessage
	err = json.Unmarshal(wssMsg, &adapterMsg)
	if err != nil {
		Logger.Error().Msgf("Error unmarshalling JSON into AdapterMessage: %s", err.Error())
		return "", adapterMsg, err
	}

	var raw map[string]json.RawMessage
	err = json.Unmarshal(adapterMsg.MessageBytes, &raw)
	if err != nil {
		Logger.Error().Msgf("Error unmarshalling JSON into json.RawMessage: %s", err.Error())
		return "", adapterMsg, err
	}

	if _, ok := raw["MessageType"]; ok {
		err = json.Unmarshal(raw["MessageType"], &messageType)
		if err != nil {
			Logger.Error().Msgf("Error unmarshalling JSON into raw[\"MessageType\"]: %s", err.Error())
			return "", adapterMsg, err
		}
	}

	return messageType, adapterMsg, err
}
