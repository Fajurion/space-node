package server

import (
	"crypto/aes"
	"errors"

	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/util"
)

func SendToRoom(room string, clientId string, bytes []byte) error {

	connections, valid := caching.GetAllConnections(room)
	if !valid {
		return errors.New("room not found")
	}

	clientIdBytes := []byte(clientId)
	for _, connection := range connections {
		if connection.Connected && connection.ClientID != clientId {

			// Create new cipher
			cipher, err := aes.NewCipher(*connection.Key)
			if err != nil {
				util.Log.Println("[udp] Error: Could not create cipher for client id")
				return err
			}

			encryptedPrefix, err := util.EncryptAES(cipher, clientIdBytes)
			if err != nil {
				util.Log.Println("[udp] Error: Could not encrypt client id")
				return err
			}

			_, err = udpServ.WriteToUDP(append(encryptedPrefix, bytes...), connection.Connection)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
