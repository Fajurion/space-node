package games_actions

import (
	"errors"

	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/caching/games"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/send"
	"github.com/Fajurion/pipesfiber/wshandler"
)

func SetupActions() {
	wshandler.Routes["game_init"] = initGame
	wshandler.Routes["game_event"] = gameEvent
	wshandler.Routes["game_start"] = startGame
}

func sendUpdateSession(adapters []string, session games.GameSession) error {
	return send.Pipe(send.ProtocolWS, pipes.Message{
		Channel: pipes.BroadcastChannel(adapters),
		Local:   true,
		Event: pipes.Event{
			Name: "session_update",
			Data: map[string]interface{}{
				"session": session.Id,
				"game":    session.Game,
				"state":   session.GameState,
				"min":     caching.GamesMap[session.Game].MinPlayers,
				"max":     caching.GamesMap[session.Game].MaxPlayers,
				"members": session.ClientIds,
			},
		},
	})
}

func sendSessionClose(room string, session string) bool {
	clients, valid := caching.GetAllConnections(room)
	if !valid {
		return false
	}
	adapters := make([]string, len(clients))
	i := 0
	for _, client := range clients {
		adapters[i] = client.Adapter
		i++
	}

	err := send.Pipe(send.ProtocolWS, pipes.Message{
		Channel: pipes.BroadcastChannel(adapters),
		Local:   true,
		Event: pipes.Event{
			Name: "session_close",
			Data: map[string]interface{}{
				"session": session,
			},
		},
	})
	return err == nil
}

func SendSessionUpdate(session games.GameSession) error {

	err := sendUpdateSession(session.ConnectionIds, session)
	if err != nil {
		return errors.New("server.error")
	}

	return err
}
