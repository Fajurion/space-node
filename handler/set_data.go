package handler

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: set_data
func setData(message wshandler.Message) {

	if message.ValidateForm("data") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	// Set data
	valid := caching.SetRoomData(message.Client.Session, message.Data["data"].(string))
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	// TODO: Send change to all clients in room

	wshandler.SuccessResponse(message)
}
