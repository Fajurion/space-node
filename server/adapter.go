package server

import (
	"errors"

	"fajurion.com/voice-node/caching"
)

func SendToRoom(room string, prefix []byte, bytes []byte) error {

	connections, valid := caching.GetAllConnections(room)
	if !valid {
		return errors.New("room not found")
	}

	for _, connection := range connections {
		if connection.Connected {
			_, err := udpServ.WriteToUDP(append(prefix, bytes...), connection.Connection)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
