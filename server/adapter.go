package server

import (
	"errors"

	"fajurion.com/voice-node/caching"
)

func SendToRoom(room string, bytes []byte) error {

	// TODO: Maybe add some sort of verification?

	connections, valid := caching.GetAllConnections(room)
	if !valid {
		return errors.New("room not found")
	}
	for _, connection := range connections {
		if connection.Connected {
			_, err := udpServ.WriteToUDP(bytes, connection.Connection)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
