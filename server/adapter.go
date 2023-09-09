package server

import (
	"fajurion.com/voice-node/caching"
)

func SendToRoom(room string, bytes []byte) error {

	// TODO: Maybe add some sort of verification?

	caching.GetAllConnections(room)
	for _, connection := range caching.GetAllConnections(room) {
		_, err := udpServ.WriteToUDP(bytes, connection)
		if err != nil {
			return err
		}
	}

	return nil
}
