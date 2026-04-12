package main

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

// AdapterWriteLoop waits for a message to appear on the Write channel, then sends that
// message to the given adapter client. If a signal consisting of any integer value is received
// on the Close channel, the loop will log a message and return.
func AdapterWriteLoop(a *Client) error {
	var err error

	remoteAddr := a.RemoteConnString()
	Logger.Info().Msgf("Entered AdapterWriteLoop for remote address: %s",
		remoteAddr)

	for {
		select {
		case <-a.Close:
			Logger.Info().Msgf("Client write loop for adapter connected at %s has received a close signal",
				remoteAddr)
			// handle close
			return err
		case adapterMsg := <-a.Write:
			// %q used to escape untrusted user input
			Logger.Info().Msgf("Writing message, %q, to adapter at %s",
				string(adapterMsg.MessageBytes), remoteAddr)
			msg, err := json.Marshal(adapterMsg)
			if err != nil {
				Logger.Error().Msgf("Error marshaling delivery failure data slice: %s", err.Error())
				break
			}
			err = a.WSConn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				Logger.Error().Msgf("Error when writing message to client: %s", err.Error())
			}
		}
	}
}
