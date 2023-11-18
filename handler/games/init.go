package games

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/send"
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

	session, valid := caching.OpenGameSession(message.Client.ID, conn.ClientID, message.Client.Session, gameId)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	caching.JoinSession(session.Id, message.Client.ID)

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
	})
}

func sendUpdateSession(adapters []string, session caching.GameSession) error {
	return send.Pipe(send.ProtocolWS, pipes.Message{
		Channel: pipes.BroadcastChannel(adapters),
		Local:   true,
		Event: pipes.Event{
			Name: "session_update",
			Data: map[string]interface{}{
				"session": session.Id,
				"game":    session.Game,
				"members": session.ClientIds,
			},
		},
	})
}
