package handler

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/send"
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

	if !SendRoomData(message.Client.Session) {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	wshandler.SuccessResponse(message)
}

func SendRoomData(id string) bool {
	room, validRoom := caching.GetRoom(id)
	members, valid := caching.GetAllConnections(id)
	if !valid || !validRoom {
		return false
	}

	adapters := make([]string, len(members))
	returnableMembers := make([]string, len(members))
	i := 0
	for _, member := range members {
		returnableMembers[i] = member.Data + ":" + member.ClientID
		adapters[i] = member.Adapter
		i++
	}

	// Send to all
	err := send.Pipe(send.ProtocolWS, pipes.Message{
		Channel: pipes.BroadcastChannel(adapters),
		Local:   true,
		Event: pipes.Event{
			Name: "room_data",
			Data: map[string]interface{}{
				"start":   room.Start,
				"room":    room.Data,
				"members": returnableMembers,
			},
		},
	})
	return err == nil
}
