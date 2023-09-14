package handler

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/send"
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

	message.Client.SendEvent(pipes.Event{
		Name:   "udp",
		Sender: send.SenderSystem,
		Data: map[string]interface{}{
			"id":  connection.ClientID,
			"key": connection.KeyBase64(),
		},
	})

	wshandler.SuccessResponse(message)
}
