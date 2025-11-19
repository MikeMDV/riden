package main

import (
	"github.com/gorilla/websocket"
)

// ClientWriteLoop waits for a message to appear on the Write channel, then sends that
// message to the given client. If a signal consisting of any integer value is received
// on the Close channel, the loop will log a message and return.
func ClientWriteLoop(c *Client) error {
	var err error

	Logger.Info().Msgf("Entered ClientWriteLoop for remote address: %s",
		c.RemoteConnString())

	for {
		select {
		case closeSignal := <-c.Close:
			Logger.Info().Msgf("Client write loop for client connected at %s has received a close signal, %d",
				c.RemoteConnString(), closeSignal)
			// handle close
			return err
		case adapterMsg := <-c.Write:
			// %q used to escape untrusted user input
			Logger.Info().Msgf("Writing message, %q, to client at %s",
				string(adapterMsg.MessageBytes), c.RemoteConnString())
			err = c.WSConn.WriteMessage(websocket.TextMessage, adapterMsg.MessageBytes)
			if err != nil {
				Logger.Error().Msgf("Error when writing message to client: %s", err.Error())
			}
		}
	}
}
