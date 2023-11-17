package games

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: game_init
func initGame(message wshandler.Message) {

	if message.ValidateForm("game") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	gameId := message.Data["game"].(string)

	sessionId, valid := caching.OpenGameSession(message.Client.Session, gameId)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	caching.JoinSession(sessionId, message.Client.ID)

	wshandler.NormalResponse(message, map[string]interface{}{
		"success": true,
		"session": sessionId,
	})
}
