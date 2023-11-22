package games_actions

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/send"
	"github.com/Fajurion/pipesfiber/wshandler"
)

func SetupActions() {
	wshandler.Routes["game_init"] = initGame
	wshandler.Routes["game_event"] = gameEvent
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
