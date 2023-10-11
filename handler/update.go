package handler

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/send"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: update
func update(message wshandler.Message) {

	if message.ValidateForm("muted", "deafened") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	connection, valid := caching.GetConnection(message.Client.ID)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	connections, valid := caching.GetAllConnections(message.Client.Session)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	client := connections[connection.ClientID]
	client.Muted = message.Data["muted"].(bool)
	client.Deafened = message.Data["deafened"].(bool)
	caching.SaveConnections(message.Client.Session, connections)

	// Send to all
	if !SendStateUpdate(connection.ClientID, message.Client.Session, client.Muted, client.Deafened) {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.SuccessResponse(message)
}

func SendStateUpdate(member string, room string, muted bool, deafened bool) bool {

	// Get all adapters
	adapters, valid := caching.GetAllAdapters(room)
	if !valid {
		return false
	}

	// Send to all
	err := send.Pipe(send.ProtocolWS, pipes.Message{
		Local:   true,
		Channel: pipes.BroadcastChannel(adapters),
		Event: pipes.Event{
			Name: "member_update",
			Data: map[string]interface{}{
				"member":   member,
				"muted":    muted,
				"deafened": deafened,
			},
		},
	})
	return err == nil
}
