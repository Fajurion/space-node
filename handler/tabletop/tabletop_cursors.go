package tabletop_handlers

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: tc_move
func moveCursor(message wshandler.Message) {

	if message.ValidateForm("x", "y") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	connection, valid := caching.GetConnection(message.Client.ID)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	x := message.Data["x"].(float64)
	y := message.Data["y"].(float64)

	// Notify other clients
	valid = SendEventToMembers(message.Client.Session, pipes.Event{
		Name: "tc_moved",
		Data: map[string]interface{}{
			"c": connection.ClientID,
			"x": x,
			"y": y,
		},
	})
	if !valid {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.SuccessResponse(message)
}
