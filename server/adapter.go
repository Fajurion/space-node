package server

import (
	"fajurion.com/voice-node/caching"
)

func SendToRoom(room string, bytes []byte) error {

	// TODO: Maybe add some sort of verification?

	for _, connection := range *caching.GetAllConnections(room) {
		if connection.Connected {
			_, err := udpServ.WriteToUDP(bytes, connection.Connection)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
