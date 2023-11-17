package games

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: game_event
func gameEvent(message wshandler.Message) {

	if message.ValidateForm("session", "name", "data") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	sessionId := message.Data["session"].(string)
	conn, valid := caching.GetConnection(message.Client.ID)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	if conn.CurrentSession != sessionId {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	valid = caching.ForwardGameEvent(sessionId, caching.EventContext{
		Client:  message.Client,
		Name:    message.Data["name"].(string),
		Session: sessionId,
		Data:    message.Data["data"].(map[string]interface{}),
	})
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	wshandler.SuccessResponse(message)
}
