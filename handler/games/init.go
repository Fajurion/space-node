package games_actions

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
	conn, valid := caching.GetConnection(message.Client.ID)
	if !valid {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	// Close current session
	currentSession, valid := caching.GetSession(conn.CurrentSession)
	if valid && len(currentSession.ClientIds) == 1 {
		caching.CloseSession(currentSession.Id)
		valid := sendSessionClose(message.Client.Session, currentSession.Id)
		if !valid {
			wshandler.ErrorResponse(message, "server.error")
			return
		}
	}

	session, valid := caching.OpenGameSession(message.Client.ID, conn.ClientID, message.Client.Session, gameId)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	caching.JoinSession(message.Client.ID, session.Id)

	// Send new session to all clients
	clients, valid := caching.GetAllConnections(message.Client.Session)
	if !valid {
		wshandler.ErrorResponse(message, "server.error")
		return
	}
	adapters := make([]string, len(clients))
	i := 0
	for _, client := range clients {
		adapters[i] = client.Adapter
		i++
	}

	err := sendUpdateSession(adapters, session)
	if err != nil {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.NormalResponse(message, map[string]interface{}{
		"success": true,
		"session": session.Id,
		"min":     caching.GamesMap[session.Game].MinPlayers,
		"max":     caching.GamesMap[session.Game].MaxPlayers,
	})
}
