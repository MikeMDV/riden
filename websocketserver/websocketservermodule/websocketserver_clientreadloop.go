package main

import (
	wss "riden/websocketserver"
	"time"

	"github.com/gorilla/websocket"
)

// ClientReadLoop reads messages from the given client, adds the client connection name
// to the message, and sends the message through a channel to the AdapterWriteLoop to
// be sent to the adapter
func ClientReadLoop(c *Client) error {
	var adapterLost bool
	Logger.Info().Msgf("Entered ClientReadLoop for remote address: %s",
		c.RemoteConnString())

	for {
		msgType, message, err := c.WSConn.ReadMessage()
		if err != nil {
			Logger.Error().Msgf("Error when reading message from client: %s", err.Error())
			return err
		}
		// Reject binary type messages and send close control message
		if msgType == websocket.BinaryMessage {
			Logger.Warn().Msgf("Received binary message type from client: %s, Sending close message with code 1003", c.RemoteConnString())
			msg := websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Binary data not supported")
			c.WSConn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))
			// Continue here and wait for close control message reply before exiting the read loop
			// The peer should respond with a close message which will cause ReadMessage()
			// to return a CloseError and the read loop to exit.
			// If the peer does not respond with a close control message, but continues to send
			// binary type messages, we will continue to send close control messages.
			// If the client begins to send text type messages we will be able to receive them
			// and process them.
			continue
		}

		// %q used to escape untrusted user input
		Logger.Info().Msgf("Received message from client connection %s: %q", c.RemoteConnString(), string(message))

		adapterMsg := wss.NewAdapterMessage(c.RemoteConnString(), message)

		// Check if adapter channel is open
		if AdapterConn.Write != nil && !adapterLost {
			AdapterConn.Write <- adapterMsg
		} else {
			// The adapter connection and the write channel were lost. Reply with a close control message
			Logger.Error().Msgf("Adapter write channel was lost. Cannot process message from client: %s, Sending close message with code 1011",
				c.RemoteConnString())
			msg := websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Server encountered an unexpected error")
			c.WSConn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(WriteControlDeadline))

			// Remove this Client from safeClients so the adapter cannot write to this client when
			// it reconnects
			safeClients.Delete(c.RemoteConnString())

			// Set adapter lost flag and continue to respond with close messages if the
			// client does not respond with a close control message and keeps attempting to
			// send further messages on this client connection.
			adapterLost = true
			continue
		}
	}
}
