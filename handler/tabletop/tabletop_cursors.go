package tabletop_handlers

import (
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: tc_move
func moveCursor(message wshandler.Message) {

	if message.ValidateForm("x", "y") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	x := message.Data["x"].(float64)
	y := message.Data["y"].(float64)

	// Notify other clients
	valid := SendEventToMembersNoSelf(message.Client.ID, message.Client.Session, pipes.Event{
		Name: "tc_moved",
		Data: map[string]interface{}{
			"c": message.Client.ID,
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
