package handler

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: update
func update(message wshandler.Message) {

	if message.ValidateForm("muted", "deafened") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	connectionId, valid := caching.ClientIDToConnectionID(message.Client.ID)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	connections, valid := caching.GetAllConnections(message.Client.Session)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	client := connections[connectionId]
	client.Muted = message.Data["muted"].(bool)
	client.Deafened = message.Data["deafened"].(bool)
	caching.SaveConnections(message.Client.Session, connections)

	wshandler.SuccessResponse(message)
}
