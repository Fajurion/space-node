package handler

import (
	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/util"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: setup
func setupUDP(message wshandler.Message) {

	if message.ValidateForm("data") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	// Insert data
	if !caching.SetMemberData(message.Client.Session, message.Client.ID, message.Data["data"].(string)) {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	// Generate new connection
	connection := caching.EmptyConnection(message.Client.ID, message.Client.Session)

	if !SendRoomData(message.Client.Session) {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.NormalResponse(message, map[string]interface{}{
		"success": true,
		"id":      connection.ClientID,
		"key":     connection.KeyBase64(),
		"port":    util.UDPPort,
	})
}
