package main

import (
	"encoding/json"
	wss "riden/websocketserver"
	"time"

	"github.com/gorilla/websocket"
)

// AdapterReadLoop reads messages from the given adapter client, gets the client
// connection name from the message, and sends the message through a channel to the
// appropriate ClientWriteLoop to be sent to the client
func AdapterReadLoop(a *Client) error {
	Logger.Info().Msgf("Entered AdapterReadLoop for remote address: %s",
		a.RemoteConnString())

	for {
		msgType, message, err := a.WSConn.ReadMessage()
		if err != nil {
			Logger.Error().Msgf("Error when reading message from adapter: %s", err.Error())
			return err
		}
		// Reject binary type messages and send close control message
		if msgType == websocket.BinaryMessage {
			Logger.Warn().Msgf("Received binary message type from adapter: %s, Sending close message with code 1003", a.RemoteConnString())
			msg := websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Binary data not supported")
			a.WSConn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))
			// Continue here and wait for close control message reply before exiting the read loop
			// The peer should respond with a close message which will cause ReadMessage()
			// to return a CloseError and the read loop to exit.
			// If the peer does not respond with a close control message, but continues to send
			// binary type messages, we will continue to send close control messages.
			// If the adapter begins to send text type messages we will be able to receive them
			// and process them.
			continue
		}
		// %q used to escape untrusted user input
		Logger.Info().Msgf("Received message from adapter connection %s: %q", a.RemoteConnString(), string(message))

		// Decode message to get client connection name
		var adapterMsg wss.AdapterMessage
		err = json.Unmarshal(message, &adapterMsg)
		if err != nil {
			Logger.Error().Msgf("Error unmarshalling JSON: %s", err.Error())
		}

		if adapterMsg.ClientConnName == wss.WSSServerAllClientsConnName {
			safeClients.Range(func(key, clientVal interface{}) bool {
				client := clientVal.(*Client)

				select {
				case client.Write <- adapterMsg:
				default:
					Logger.Error().Msgf("could not place messaage on client.Write: %+v", adapterMsg)
				}

				return true
			})
		} else {
			// Write to the appropriate client connection
			var client *Client
			clientVal, ok := safeClients.Load(adapterMsg.ClientConnName)
			if !ok {
				Logger.Error().Msgf("No Client was found for connection name, %s. Unable to write message: %s",
					adapterMsg.ClientConnName, string(adapterMsg.MessageBytes))
				continue
			}
			client = clientVal.(*Client)

			select {
			case client.Write <- adapterMsg:
			default:
				Logger.Error().Msgf("could not place messaage on client.Write: %+v", adapterMsg)
			}
		}
	}
}
